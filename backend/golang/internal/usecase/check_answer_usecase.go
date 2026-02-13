package usecase

import "github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"

type CheckAnswersUsecase struct {
	gm *core.GameManager
}

func (cau *CheckAnswersUsecase) Execute() (map[core.TeamID]core.Result, error) {
	teamResult, answerMap, err := cau.gm.CollectAnswer()
	if err != nil {
		return nil, err
	}
	cau.gm.DistributeAnswer(teamResult, answerMap)
	return teamResult, nil
}

func NewCheckAnswersUsecase(gm *core.GameManager) *CheckAnswersUsecase {
	return &CheckAnswersUsecase{
		gm: gm,
	}
}