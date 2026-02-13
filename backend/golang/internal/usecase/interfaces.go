package usecase

import (
	"github.com/google/uuid"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

type IUserRepository interface {
	Save(*model.User) error
	SaveBulk([]model.User) error
	FetchByUserID(uuid.UUID) (*model.User, error)
	FetchByUserIDs([]uuid.UUID) ([]model.User, error)
	FetchByTeamID(uint32) ([]model.User, error)
}

type IUserImageRepository interface {
	Save(uuid.UUID, string) error
	FetchByUserID(uuid.UUID) (string, error)
}

type IUserProfileRepository interface {
	Save(*model.UserProfile) error
	FetchByProfileIDWithUserGroup(uint, []uuid.UUID) ([]model.UserProfile, error)
}

type IProfileQuestionRepository interface {
	FetchByQuestionID(uint) (*model.ProfileQuestion, error)
	FetchAllQuestions() ([]model.ProfileQuestion, error)
}
