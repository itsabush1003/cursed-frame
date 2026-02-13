package infra

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"slices"
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
	return db.connections[dbName][Read].Queryx(sql, params)
}

func (db *SQLiteDB) QueryRow(dbName string, sql string, params ...any) *sqlx.Row {
	return db.connections[dbName][Read].QueryRowx(sql, params)
}

func (db *SQLiteDB) QueryIn(dbName string, sql string, params ...any) (*sqlx.Rows, error) {
	query, newParams, err := sqlx.In(sql, params)
	if err != nil {
		return nil, err
	}
	return db.connections[dbName][Read].Queryx(query, newParams)
}

func NewSQLiteDB(dbFileDir string) (*SQLiteDB, error) {
	connections := make(map[string]map[Mode]*sqlx.DB, len(databases))
	for dbName := range databases {
		connections[dbName] = make(map[Mode]*sqlx.DB, 2)
		reader, err := sqlx.Open("sqlite", fmt.Sprintf("file:%s/%s.db?%s", dbFileDir, dbName, ReadOnlyDsnOption))
		if err != nil {
			// 途中で失敗した場合に過去に生成済みのものもcloseする
			for _, conn := range connections {
				if conn != nil {
					conn[Read].Close()
					conn[Write].Close()
				}
			}
			os.RemoveAll(dbFileDir)
			return nil, err
		}
		writer, err := sqlx.Open("sqlite", fmt.Sprintf("file:%s/%s.db?%s", dbFileDir, dbName, ReadWriteDsnOption))
		if err != nil {
			// 途中で失敗した場合に過去に生成済みのものもcloseする
			for _, conn := range connections {
				if conn != nil {
					conn[Read].Close()
					conn[Write].Close()
				}
			}
			reader.Close()
			os.RemoveAll(dbFileDir)
			return nil, err
		}
		writer.SetMaxOpenConns(1)
		connections[dbName][Read] = reader
		connections[dbName][Write] = writer
	}

	// 書き込みキューをデータベースにつき１つに限定したいのでOnceValueで作る
	queues := sync.OnceValue(func() map[string]chan<- repository.WriteRequest {
		batchSize := 10
		queueMap := make(map[string]chan<- repository.WriteRequest, len(databases)-1)
		for db, _ := range databases {
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
