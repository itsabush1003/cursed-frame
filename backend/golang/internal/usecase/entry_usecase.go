package usecase

import (
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/util"
)

type EntryDTO struct {
	AccessToken  string
	ReconnectKey string
}

type EntryUsecase struct {
	ur IUserRepository
	secret []byte
}

func (ueu *EntryUsecase) Execute(name string) (EntryDTO, error) {
	user, err := model.NewUser(name)
	if err != nil {
		return EntryDTO{}, err
	}
	key, err := util.Encrypt(user.GetUserID().String(), ueu.secret)
	if err != nil {
		return EntryDTO{}, err
	}
	if err = ueu.ur.Save(user); err != nil {
		return EntryDTO{}, err
	}
	return EntryDTO{
		AccessToken: user.GetAccessToken(),
		ReconnectKey: key,
	}, nil
}

func NewEntryUsecase(ur IUserRepository, secret []byte) *EntryUsecase {
	return &EntryUsecase{
		ur: ur,
		secret: secret,
	}
}
