package usecase

import (
	"errors"

	"github.com/google/uuid"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

type ChangeTeamUsecase struct {
	ur IUserRepository
}

func (ctu *ChangeTeamUsecase) Execute(userIDStr string, newTeamID uint32) error {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}

	user, err := ctu.ur.FetchByUserID(uid)
	if err != nil {
		return err
	}
	if user.GetTeamID() == model.UNDEFINED.Raw() {
		return errors.New("Teams have not been splitted yet")
	}

	teammates, err := ctu.ur.FetchByTeamID(user.GetTeamID())
	if err != nil {
		return err
	}
	if len(teammates)-1 < model.MinTeamUser {
		return errors.New("Cannot change team because a team must have at least 3 users")
	}

	user.SetTeamID(newTeamID)
	if err = ctu.ur.Save(user); err != nil {
		return err
	}

	return nil
}

func NewChangeTeamUsecase(ur IUserRepository) *ChangeTeamUsecase {
	return &ChangeTeamUsecase{
		ur: ur,
	}
}
