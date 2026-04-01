package usecase

import "github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"

type ReadyQuizUsecase struct {
	gm *core.GameManager
}

func (nqu *ReadyQuizUsecase) Execute() error {
	return nqu.gm.StartCount()
}

func NewReadyQuizUsecase(gm *core.GameManager) *ReadyQuizUsecase {
	return &ReadyQuizUsecase{
		gm: gm,
	}
}
