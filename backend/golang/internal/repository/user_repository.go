package repository

import (
	"database/sql"
	"errors"
	"maps"
	"reflect"
	"slices"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

func getDBColums(obj any) map[string]string {
	t := reflect.TypeOf(obj)
	c := make(map[string]string, t.NumField())
	for i := range t.NumField() {
		c[t.Field(i).Name] = t.Field(i).Tag.Get("db")
	}
	return c
}

type DBUserRow struct {
	UserID      string `db:"user_id"`
	Name        string `db:"name"`
	AccessToken string `db:"access_token"`
	TeamID      int    `db:"team_id"`
	IsReady     bool   `db:"is_ready"`
	Version     int    `db:"version"`
}

func (ur *DBUserRow) UpdateChangedColumns(user *model.User) []string {
	// UserIDとNameは変わらない想定なので、それ以外の4つフィールドでチェック
	changed := make([]string, 0, 4)
	if ur.AccessToken != user.GetAccessToken() {
		ur.AccessToken = user.GetAccessToken()
		changed = append(changed, "access_token")
	}
	if ur.TeamID != int(user.GetTeamID()) {
		ur.TeamID = int(user.GetTeamID())
		changed = append(changed, "team_id")
	}
	if ur.IsReady != user.GetIsReady() {
		ur.IsReady = user.GetIsReady()
		changed = append(changed, "is_changed")
	}
	if len(changed) > 0 {
		// DB上の値と何かしらの差分があった場合のみ、バージョンを上げてDBを更新する
		user.IncrementVersion()
		ur.Version = int(user.GetVersion())
		changed = append(changed, "version")
	}
	return changed
}

var userDBColumns map[string]string = getDBColums(DBUserRow{})

type UserRepository struct {
	c  *cache.Cache
	db IDatabase
}

func (ur *UserRepository) Save(user *model.User) error {
	method := Update
	var (
		targets []string
		params  DBUserRow
		conds   string
	)

	// 保存する前に、引数で渡されたユーザの情報が最新のものかどうか確認する
	err := ur.db.QueryRow("User", "SELECT * FROM User WHERE user_id = :user_id", DBUserRow{
		UserID: user.GetUserID().String(),
	}).StructScan(&params)
	if err != nil {
		// DBに存在しない＝初登録なので、その場合はエラーにせずINSERTする
		if errors.Is(err, sql.ErrNoRows) {
			method = Insert
			targets = slices.Collect(maps.Values(userDBColumns))
			params = DBUserRow{
				UserID:      user.GetUserID().String(),
				Name:        user.GetName(),
				AccessToken: user.GetAccessToken(),
				TeamID:      int(user.GetTeamID()),
				IsReady:     user.GetIsReady(),
				Version:     int(user.GetVersion()),
			}
			conds = ""
		} else {
			return err
		}
	} else {
		if params.Version != int(user.GetVersion()) {
			// userの情報がおかしいのでキャッシュから消してエラーを返す
			// ＝＞次回アクセス時にはDBから正しいデータが取れるはず
			ur.c.Delete(user.GetAccessToken())
			return errors.New("Internal data inconsistency")
		}
		targets = (&params).UpdateChangedColumns(user)
		// 更新するべき情報が無い場合はこのままリターン
		if len(targets) == 0 {
			return nil
		}
		conds = "user_id = :user_id"
	}
	resultCh := make(chan error, 1)
	ur.db.Command("User", WriteRequest{
		Table:    "User",
		Method:   method,
		Targets:  targets,
		Params:   params,
		Conds:    conds,
		ResultCh: resultCh,
	})
	if err = <-resultCh; err != nil {
		return err
	}
	ur.c.Set(user.GetAccessToken(), *user, cache.DefaultExpiration)
	return nil
}

func (ur *UserRepository) SaveBulk(users []model.User) error {
	for _, user := range users {
		params := DBUserRow{
			UserID:      user.GetUserID().String(),
			Name:        user.GetName(),
			AccessToken: user.GetAccessToken(),
			TeamID:      int(user.GetTeamID()),
			IsReady:     user.GetIsReady(),
			Version:     int(user.GetVersion()),
		}
		resultCh := make(chan error, 1)
		ur.db.Command("User", WriteRequest{
			Table:    "User",
			Method:   Update,
			Targets:  slices.Collect(maps.Values(userDBColumns)),
			Params:   params,
			Conds:    "user_id = :user_id",
			ResultCh: resultCh,
		})
		if err := <-resultCh; err != nil {
			return err
		}
		ur.c.Delete(user.GetAccessToken())
		ur.c.Set(user.GetAccessToken(), user, cache.DefaultExpiration)
	}
	return nil
}

func (ur *UserRepository) FetchByUserID(uid uuid.UUID) (*model.User, error) {
	dbUser := DBUserRow{}
	if err := ur.db.QueryRow("User", "SELECT * FROM User WHERE user_id = :user_id", DBUserRow{
		UserID: uid.String(),
	}).StructScan(&dbUser); err != nil {
		return nil, err
	}
	return model.ReconstructUser(
		dbUser.UserID,
		dbUser.Name,
		dbUser.AccessToken,
		dbUser.TeamID,
		dbUser.IsReady,
		dbUser.Version,
	)
}

func (ur *UserRepository) FetchByUserIDs(uids []uuid.UUID) ([]model.User, error) {
	users := make([]model.User, 0, len(uids))
	uidList := make([]string, 0, len(uids))
	for _, uid := range uids {
		uidList = append(uidList, uid.String())
	}
	rows, err := ur.db.QueryIn("User", "SELECT * FROM User WHERE user_id in (?)", uidList)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		dbUser := DBUserRow{}
		if err := rows.StructScan(&dbUser); err != nil {
			return nil, err
		}
		user, err := model.ReconstructUser(
			dbUser.UserID,
			dbUser.Name,
			dbUser.AccessToken,
			dbUser.TeamID,
			dbUser.IsReady,
			dbUser.Version,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	return users, nil
}

func (ur *UserRepository) FetchByTeamID(tid uint32) ([]model.User, error) {
	var n int
	if err := ur.db.QueryRow("User", "SELECT COUNT(*) FROM User WHERE team_id = :team_id", DBUserRow{
		TeamID: int(tid),
	}).Scan(&n); err != nil {
		return nil, err
	}
	users := make([]model.User, 0, n)
	rows, err := ur.db.Query("User", "SELECT * FROM User WHERE team_id = :team_id", DBUserRow{
		TeamID: int(tid),
	})
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		dbUser := DBUserRow{}
		if err := rows.StructScan(&dbUser); err != nil {
			return nil, err
		}
		user, err := model.ReconstructUser(
			dbUser.UserID,
			dbUser.Name,
			dbUser.AccessToken,
			dbUser.TeamID,
			dbUser.IsReady,
			dbUser.Version,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	return users, nil
}

func (ur *UserRepository) FetchByToken(token string) (*model.User, error) {
	if usr, found := ur.c.Get(token); found {
		user, ok := usr.(model.User)
		if ok {
			return &user, nil
		}
		// Userにキャストできない異常な何かが入ってることになるので、消しておく
		ur.c.Delete(token)
	}

	dbUser := DBUserRow{}
	if err := ur.db.QueryRow("User", "SELECT * FROM User WHERE access_token = :access_token", DBUserRow{
		AccessToken: token,
	}).StructScan(&dbUser); err != nil {
		return nil, err
	}

	user, err := model.ReconstructUser(
		dbUser.UserID,
		dbUser.Name,
		dbUser.AccessToken,
		dbUser.TeamID,
		dbUser.IsReady,
		dbUser.Version,
	)
	if err != nil {
		return nil, err
	}
	ur.c.Delete(token)
	ur.c.Set(token, *user, cache.DefaultExpiration)
	return user, nil
}

func (ur *UserRepository) RemoveUser(uid uuid.UUID) error {
	dbUser := DBUserRow{}
	if err := ur.db.QueryRow("User", "SELECT * FROM User WHERE user_id = :user_id", DBUserRow{
		UserID: uid.String(),
	}).StructScan(&dbUser); err != nil {
		return err
	}
	resCh := make(chan error)
	ur.db.Command("User", WriteRequest{
		Table:    "User",
		Method:   Delete,
		Targets:  []string{"user_id"},
		Params:   dbUser,
		Conds:    "user_id = :user_id",
		ResultCh: resCh,
	})
	if err := <-resCh; err != nil {
		return err
	}
	ur.c.Delete(dbUser.AccessToken)
	return nil
}

func NewUserRepository(c *cache.Cache) *UserRepository {
	return &UserRepository{
		c: c,
	}
}
