package repository

import (
	"maps"
	"slices"

	"github.com/google/uuid"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

type DBProfileRow struct {
	UserID    string `db:"user_id"`
	ProfileID int    `db:"profile_id"`
	Answer    string `db:"answer"`
}

var profileDBColumns map[string]string = getDBColums(DBProfileRow{})

type UserProfileRepository struct {
	db IDatabase
}

func (upr *UserProfileRepository) Save(profile *model.UserProfile) error {
	resultCh := make(chan error, 1)
	upr.db.Command("UserAttribute", WriteRequest{
		Table:   "UserProfile",
		Method:  Insert,
		Targets: slices.Collect(maps.Values(profileDBColumns)),
		Params: DBProfileRow{
			UserID:    profile.GetUserID().String(),
			ProfileID: int(profile.GetProfileID()),
			Answer:    profile.GetAnswer(),
		},
		Conds:    "",
		ResultCh: resultCh,
	})
	if err := <-resultCh; err != nil {
		return err
	}
	return nil
}

func (upr *UserProfileRepository) FetchByProfileIDWithUserGroup(pid uint, users []uuid.UUID) ([]model.UserProfile, error) {
	profiles := make([]model.UserProfile, 0, len(users))
	uidList := make([]string, 0, len(users))
	for _, uid := range users {
		uidList = append(uidList, uid.String())
	}
	rows, err := upr.db.QueryIn("UserAttribute", "SELECT * FROM UserProfile WHERE user_id in (?) AND profile_id = :profile_id", uidList, DBProfileRow{ProfileID: int(pid)})
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		dbProfile := DBProfileRow{}
		if err := rows.StructScan(&dbProfile); err != nil {
			return nil, err
		}
		uid, err := uuid.Parse(dbProfile.UserID)
		if err != nil {
		}
		profile, err := model.NewUserProfile(
			uid,
			uint(dbProfile.ProfileID),
			dbProfile.Answer,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, *profile)
	}

	return profiles, nil
}

func NewUserProfileRepository(db IDatabase) *UserProfileRepository {
	return &UserProfileRepository{
		db: db,
	}
}
