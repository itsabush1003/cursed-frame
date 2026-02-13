package usecase

import "github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"

type NextQuizUsecase struct {
	gm *core.GameManager
}

func (nqu *NextQuizUsecase) Execute() error {
	return nqu.gm.NextQuiz()
}

func NewNextQuizUsecase(gm *core.GameManager) *NextQuizUsecase {
	return &NextQuizUsecase{
		gm: gm,
	}
}