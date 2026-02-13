package usecase

import (
	"github.com/google/uuid"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
)

type IUserRepositoryForAdmin interface {
	IUserRepository
	RemoveUser(uuid.UUID) error
}

type RejectUserUsecase struct {
	gm *core.GameManager
	ur IUserRepositoryForAdmin
}

func (ruu *RejectUserUsecase) Execute(userIDStr string) error {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}

	// Userを最初に消してこれ以上のアクセスを防ぐ
	if err = ruu.ur.RemoveUser(uid); err != nil {
		return err
	}

	if err = ruu.gm.DisconnectLobby(uid); err != nil {
		return nil
	}

	return nil
}

func NewRejectUserUsecase(gm *core.GameManager, ur IUserRepositoryForAdmin) *RejectUserUsecase {
	return &RejectUserUsecase{
		gm: gm,
		ur: ur,
	}
}
