package infra

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/repository"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const SQLitePlaceholder string = "?"

const (
	UserTable     string = "User"
	ImageTable    string = "UserImage"
	ProfileTable  string = "UserProfile"
	QuestionTable string = "ProfileQuestion"
)

type column struct {
	Name       string
	Type       string
	Constraint string
}

func (c column) Definition() string {
	if c.Name == "" || c.Type == "" {
		return ""
	}
	str := fmt.Sprintf("%s %s", c.Name, c.Type)
	if c.Constraint != "" {
		str = str + " " + c.Constraint
	}
	return str
}

type columns struct {
	Columns []column
}

func (cols columns) ColNames() []string {
	f := func(yield func(v string) bool) {
		for _, c := range cols.Columns {
			if !yield(c.Name) {
				return
			}
		}
	}
	return slices.Collect(f)
}

func (cols columns) toDDL() string {
	f := func(yield func(v string) bool) {
		for _, c := range cols.Columns {
			if !yield(c.Definition()) {
				return
			}
		}
	}
	return strings.Join(slices.Collect(f), ", ")
}

var columnMap map[string]columns = map[string]columns{
	UserTable: columns{Columns: []column{
		{Name: "user_id", Type: "TEXT", Constraint: "PRIMARY KEY"},
		{Name: "name", Type: "TEXT"},
		{Name: "access_token", Type: "TEXT", Constraint: "UNIQUE"},
		{Name: "team_id", Type: "INTEGER"},
		{Name: "is_ready", Type: "BOOLEAN"},
		{Name: "version", Type: "INTEGER"},
	}},
	ImageTable: columns{Columns: []column{
		{Name: "user_id", Type: "TEXT", Constraint: "PRIMARY KEY"},
		{Name: "image_id", Type: "TEXT", Constraint: "UNIQUE"},
	}},
	ProfileTable: columns{Columns: []column{
		{Name: "user_id", Type: "TEXT"},
		{Name: "profile_id", Type: "INTEGER"},
		{Name: "answer", Type: "TEXT"},
	}},
	QuestionTable: columns{Columns: []column{
		{Name: "question_id", Type: "INTEGER", Constraint: "PRIMARY KEY"},
		{Name: "question_text", Type: "TEXT"},
		{Name: "quiz_text", Type: "TEXT"},
		{Name: "sample_answer", Type: "TEXT"},
	}},
}

var migrations map[string]string = map[string]string{
	UserTable:     fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(%s);", UserTable, columnMap[UserTable].toDDL()),
	ImageTable:    fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(%s);", ImageTable, columnMap[ImageTable].toDDL()),
	ProfileTable:  fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(%s, PRIMARY KEY(user_id, profile_id));", ProfileTable, columnMap[ProfileTable].toDDL()),
	QuestionTable: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(%s);", QuestionTable, columnMap[QuestionTable].toDDL()),
}

var databases map[string][]string = map[string][]string{
	"User": []string{UserTable},
	"UserAttribute": []string{
		ImageTable,
		ProfileTable,
	},
	"Master": []string{QuestionTable},
}

var doBatchTables []string = []string{
	UserTable,
	ProfileTable,
}

type Mode string

const (
	Read  Mode = "READ"
	Write Mode = "WRITE"
)

const (
	ReadWriteDsnOption string = "_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&cache=shared&_pragma=mmap_size(1024)&_pragma=mutex_mode(0)&_pragma=locking_mode(EXCLUSIVE)&_pragma=synchronous(NORMAL)&_pragma=temp_store(MEMORY)&_pragma=cache_size(10000)"
	ReadOnlyDsnOption  string = "_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&mode=ro&cache=shared&immutable=1&_pragma=mmap_size(1024)&_pragma=mutex_mode(0)&_pragma=locking_mode(EXCLUSIVE)&_pragma=synchronous(OFF)&_pragma=temp_store(MEMORY)&_pragma=cache_size(10000)"
)

// なぜかsqlxにNamedExecを使えるinterfaceがなかったので
type Execerx interface {
	sqlx.Execer
	NamedExec(string, interface{}) (sql.Result, error)
}

type SQLiteDB struct {
	connections map[string]map[Mode]*sqlx.DB
	writeQueues map[string]chan<- repository.WriteRequest
	dbFileDir   string
}

func (db *SQLiteDB) Close() {
	for dbName, conn := range db.connections {
		db.closeQueue(dbName)
		if conn[Read] != nil {
			conn[Read].Close()
		}
		if conn[Write] != nil {
			conn[Write].Close()
		}
	}
	os.RemoveAll(db.dbFileDir)
}

func (db *SQLiteDB) closeQueue(dbName string) {
	if db.writeQueues[dbName] != nil {
		close(db.writeQueues[dbName])
		db.writeQueues[dbName] = nil
	}
}

func (db *SQLiteDB) Command(dbName string, req repository.WriteRequest) {
	if dbName == "Master" {
		req.ResultCh <- errors.New("Master database is read only")
		return
	}
	q, ok := db.writeQueues[dbName]
	if !ok {
		req.ResultCh <- errors.New(dbName + " is not exist")
	}
	select {
	case q <- req:
	case <-time.After(time.Second):
		req.ResultCh <- errors.New("Write queue is full")
	}
}

func (db *SQLiteDB) Query(dbName string, sql string, params ...any) (*sqlx.Rows, error) {
	// paramsが空の場合はそのままSQLを実行
	if len(params) == 0 {
		return db.connections[dbName][Read].Queryx(sql)
	}

	// SQLに名前付きパラメータ(:)が含まれている場合はNamedQueryを使用
	// sqlxではPostgreSQL用に::の対応があるが、今はSQLite用なので対応しない
	if strings.Contains(sql, ":") {
		return db.connections[dbName][Read].NamedQuery(sql, params)
	}

	// 通常のプレースホルダ(?)の場合は通常のQueryxを使用
	return db.connections[dbName][Read].Queryx(sql, params...)
}

func (db *SQLiteDB) QueryRow(dbName string, sql string, params ...any) *sqlx.Row {
	// paramsが空の場合はsqlx.Namedがエラーを返してしまうのでそのままSQLを実行
	if len(params) == 0 {
		return db.connections[dbName][Read].QueryRowx(sql)
	}

	// SQLに名前付きパラメータ(:)が含まれている場合はNamedQueryを使用
	// sqlxではPostgreSQL用に::の対応があるが、今はSQLite用なので対応しない
	if strings.Contains(sql, ":") {
		query, newParams, err := sqlx.Named(sql, params)
		if err != nil {
			// 空のRowを返してしまうとErrNoRowsが出て処理上都合の悪い場合があるのと、エラーの内容が完全に消えてしまうので、エラーを詰めたRowを返す
			return db.connections[dbName][Read].QueryRowx("SELECT ?", err.Error())
		}
		sql, params = query, newParams
	}
	return db.connections[dbName][Read].QueryRowx(sql, params...)
}

func (db *SQLiteDB) QueryIn(dbName string, sql string, params ...any) (*sqlx.Rows, error) {
	query, newParams, err := sqlx.In(sql, params...)
	if err != nil {
		return nil, err
	}
	return db.connections[dbName][Read].Queryx(query, newParams...)
}

func readXSVFile(file fs.File) (header []string, body [][]string, err error) {
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	header = rows[0]
	body = rows[1:]
	return header, body, nil
}

func createValueMap(header []string, body [][]string, columns columns) []map[string]any {
	values := make([]map[string]any, 0, len(body))
	colNames := columns.ColNames()
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, row := range body {
		// 過剰な最適化説もあるけどやっておく
		wg.Go(func() {
			valueMap := make(map[string]any, len(colNames))
			for i, col := range colNames {
				idx := slices.Index(header, col)
				switch columns.Columns[i].Type {
				case "INTEGER":
					if idx < 0 {
						valueMap[col] = 0
						continue
					}
					n, err := strconv.Atoi(row[idx])
					if err != nil {
						n = 0
					}
					valueMap[col] = n
				case "BOOLEAN":
					// BOOLEANってSQLiteには無いらしいけど動くし分かりやすいからこのままで
					if idx < 0 {
						valueMap[col] = false
						continue
					}
					b, err := strconv.ParseBool(row[idx])
					if err != nil {
						b = false
					}
					valueMap[col] = b
				case "TEXT":
					if idx < 0 {
						valueMap[col] = ""
						continue
					}
					valueMap[col] = row[idx]
				default:
					valueMap[col] = nil
				}
			}
			mu.Lock()
			defer mu.Unlock()
			values = append(values, valueMap)
		})
	}
	wg.Wait()
	return values
}

func NewSQLiteDB(dbFileDir string, dbSources fs.FS) (*SQLiteDB, error) {
	connections := make(map[string]map[Mode]*sqlx.DB, len(databases))
	clearConnections := func() {
		// 途中で失敗した場合に過去に生成済みのものをcloseする
		for _, conns := range connections {
			for _, con := range conns {
				con.Close()
			}
		}
	}
	readSourceFile := func(fileName string) ([]string, [][]string, error) {
		file, err := dbSources.Open(fileName)
		if err != nil {
			return nil, nil, err
		}
		defer file.Close()
		return readXSVFile(file)
	}
	for dbName := range databases {
		connections[dbName] = make(map[Mode]*sqlx.DB, 2)
		reader, err := sqlx.Open("sqlite", fmt.Sprintf("file:%s/%s.db?%s", dbFileDir, dbName, ReadOnlyDsnOption))
		if err != nil {
			clearConnections()
			os.RemoveAll(dbFileDir)
			return nil, err
		}
		writer, err := sqlx.Open("sqlite", fmt.Sprintf("file:%s/%s.db?%s", dbFileDir, dbName, ReadWriteDsnOption))
		if err != nil {
			clearConnections()
			reader.Close()
			os.RemoveAll(dbFileDir)
			return nil, err
		}
		writer.SetMaxOpenConns(1)
		connections[dbName][Read] = reader
		connections[dbName][Write] = writer

		// DB初期化処理
		for _, table := range databases[dbName] {
			if _, err = connections[dbName][Write].Exec(migrations[table]); err != nil {
				break
			}
			filePath := fmt.Sprintf("%s/%s.csv", dbName, table)
			if match, err := fs.Glob(dbSources, filePath); err == nil && match != nil {
				header, body, err := readSourceFile(filePath)
				if err != nil {
					break
				}
				colNames := columnMap[table].ColNames()
				query := fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s);", table, strings.Join(colNames, ", "), ":"+strings.Join(colNames, ", :"))
				values := createValueMap(header, body, columnMap[table])
				if _, err := connections[dbName][Write].NamedExec(query, values); err != nil {
					break
				}
			}
		}
		if err != nil {
			clearConnections()
			os.RemoveAll(dbFileDir)
			return nil, err
		}
	}

	// 書き込みキューをデータベースにつき１つに限定したいのでOnceValueで作る
	queues := sync.OnceValue(func() map[string]chan<- repository.WriteRequest {
		batchSize := 10
		queueMap := make(map[string]chan<- repository.WriteRequest, len(databases)-1)
		for db := range databases {
			// Masterは起動時以外書き込む必要が無いのでスキップ
			if db == "Master" {
				continue
			}
			q := make(chan repository.WriteRequest, 100)

			go func(db string) {
				batch := make([]repository.WriteRequest, 0, batchSize)
				sendErr := func(ch chan<- error, err error) {
					select {
					case ch <- err:
					case <-time.After(10 * time.Millisecond):
						// ここでブロックされてしまうと全部止まってしまうので早めにタイムアウト
					}
				}
				doInsert := func(ext *sqlx.DB, req repository.WriteRequest) {
					stmt := fmt.Sprintf(
						"INSERT INTO %s(%s) VALUES (%s);",
						req.Table,
						strings.Join(req.Targets, ", "),
						":"+strings.Join(req.Targets, ", :"),
					)
					_, err := ext.NamedExec(stmt, req.Params)
					sendErr(req.ResultCh, err)
				}
				flush := func(ext *sqlx.DB, requests []repository.WriteRequest) {
					params := make([]any, 0, len(requests))
					for _, req := range requests {
						params = append(params, req.Params)
					}
					resCh := make(chan error, 1)
					// batchに保存されてるWriteRequestはParams以外同じになるはずなのでindex=0のものを利用
					// ResultChだけはここで一回受けてから元のWriteRequestに同じものを渡したいので新規に作成
					newRequest := repository.WriteRequest{
						Table:    requests[0].Table,
						Targets:  requests[0].Targets,
						Params:   params,
						Conds:    requests[0].Conds,
						ResultCh: resCh,
					}
					doInsert(ext, newRequest)
					err := <-resCh
					for _, req := range requests {
						sendErr(req.ResultCh, err)
					}
				}
				timer := time.NewTimer(time.Second)
				for {
					select {
					case req, ok := <-q:
						if !ok {
							return
						}
						if !slices.Contains(databases[db], req.Table) {
							sendErr(req.ResultCh, errors.New("Table is not exist"))
							continue
						}
						switch req.Method {
						case repository.Insert:
							if slices.Contains(doBatchTables, req.Table) {
								batch = append(batch, req)
								if len(batch) >= batchSize {
									flush(connections[db][Write], batch)
									batch = batch[:0]
								}
							} else {
								doInsert(connections[db][Write], req)
							}
						case repository.Update:
							values := make([]string, 0, len(req.Targets))
							for _, t := range req.Targets {
								values = append(values, fmt.Sprintf("%s = :%s", t, t))
							}
							stmt := fmt.Sprintf(
								"UPDATE %s SET %s WHERE %s;",
								req.Table,
								strings.Join(values, ", "),
								req.Conds,
							)
							_, err := connections[db][Write].NamedExec(stmt, req.Params)
							sendErr(req.ResultCh, err)
						case repository.Delete:
							stmt := fmt.Sprintf("DELETE FROM %s WHERE %s;", req.Table, req.Conds)
							_, err := connections[db][Write].NamedExec(stmt, req.Params)
							sendErr(req.ResultCh, err)
						default:
							req.ResultCh <- errors.ErrUnsupported
						}
					case <-timer.C:
						if len(batch) > 0 {
							flush(connections[db][Write], batch)
							batch = batch[:0]
						}
						timer.Reset(time.Second)
					}
				}
			}(db)

			queueMap[db] = q
		}

		return queueMap
	})()

	return &SQLiteDB{
		connections: connections,
		writeQueues: queues,
		dbFileDir:   dbFileDir,
	}, nil
}
