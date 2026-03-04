package usecase

import "github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"

type SetReadyUsecase struct {
	ur IUserRepository
}

func (sru *SetReadyUsecase) Execute(user *model.User) error {
	user.SetReady()
	return sru.ur.Save(user)
}

func NewSetReadyUsecase(ur IUserRepository) *SetReadyUsecase {
	return &SetReadyUsecase{
		ur: ur,
	}
}
