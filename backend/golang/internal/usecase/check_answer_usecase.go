package usecase

import "github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"

type CheckAnswersUsecase struct {
	gm *core.GameManager
}

func (cau *CheckAnswersUsecase) Execute() (map[core.TeamID]core.Result, core.Choice, error) {
	teamResult, answerMap, err := cau.gm.CollectAnswer()
	if err != nil {
		return nil, core.Choice{}, err
	}
	cau.gm.DistributeAnswer(teamResult, answerMap)
	return teamResult, cau.gm.GetCurrentAnswer(), nil
}

func NewCheckAnswersUsecase(gm *core.GameManager) *CheckAnswersUsecase {
	return &CheckAnswersUsecase{
		gm: gm,
	}
}