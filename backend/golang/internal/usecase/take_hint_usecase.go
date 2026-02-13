package usecase

import (
	"github.com/google/uuid"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
)

type TakeHintUsecase struct {
	gm *core.GameManager
}

func (thu *TakeHintUsecase) Execute(uid uuid.UUID, hint string) error {
	return thu.gm.TakeHint(uid, hint)
}

func NewTakeHintUsecase(gm *core.GameManager) *TakeHintUsecase {
	return &TakeHintUsecase{
		gm: gm,
	}
}