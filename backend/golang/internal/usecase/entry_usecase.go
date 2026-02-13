package usecase

import (
	"github.com/google/uuid"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

type UserDTO struct {
	UserID      uuid.UUID
	AccessToken string
}

type EntryUsecase struct {
	ur IUserRepository
}

func (ueu *EntryUsecase) Execute(name string) (UserDTO, error) {
	user, err := model.NewUser(name)
	if err != nil {
		return UserDTO{}, err
	}
	if err = ueu.ur.Save(user); err != nil {
		return UserDTO{}, err
	}
	return UserDTO{
		UserID:      user.GetUserID(),
		AccessToken: user.GetAccessToken(),
	}, nil
}

func NewEntryUsecase(ur IUserRepository) *EntryUsecase {
	return &EntryUsecase{
		ur: ur,
	}
}
