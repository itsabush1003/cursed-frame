package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/util"
)

type ReconnectUsecase struct {
	secret []byte
	ur     IUserRepository
}

func (ru *ReconnectUsecase) Execute(reconnectKey string) (token string, err error) {
	userIDStr, err := util.Decrypt(reconnectKey, ru.secret)
	if err != nil {
		return "", err
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", err
	}

	user, err := ru.ur.FetchByUserID(uid)
	if err != nil {
		return "", err
	}

	if err := user.RefreshAccessToken(); err != nil {
		return "", errors.ErrUnsupported
	}
	ru.ur.Save(user)
	return user.GetAccessToken(), nil
}

func NewReconnectUsecase(secret []byte, ur IUserRepository) *ReconnectUsecase {
	return &ReconnectUsecase{
		secret: secret,
		ur:     ur,
	}
}
