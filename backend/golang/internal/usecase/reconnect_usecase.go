package usecase

import (
	"errors"

	"github.com/google/uuid"
)

type ReconnectUsecase struct {
	secret string
	ur     IUserRepository
}

func (ru *ReconnectUsecase) Execute(userIDStr string, secret string) (token string, err error) {
	if secret != ru.secret {
		return "", errors.New("Invalid Acccess")
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

func NewReconnectUsecase(secret string, ur IUserRepository) *ReconnectUsecase {
	return &ReconnectUsecase{
		secret: secret,
		ur:     ur,
	}
}
