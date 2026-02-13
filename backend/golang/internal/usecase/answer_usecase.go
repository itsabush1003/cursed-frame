package usecase

import (
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

type AnswerDTO struct {
	ChoiceID uint
	ChoiceText string
}

type AnswerUsecase struct {
	gm *core.GameManager
}

func (au *AnswerUsecase) Execute(user *model.User, answer AnswerDTO) (core.Result, map[uint]int, error) {
	teamAnswer, correctAnswer, err := au.gm.Answer(user.GetUserID(), core.TeamID(user.GetTeamID()), core.Choice{
		ChoiceID: answer.ChoiceID,
		ChoiceText: answer.ChoiceText,
	})
	if err != nil {
		return core.Result{}, nil, err
	}
	return core.Result{
		Answer: teamAnswer.TeamAnswer,
		IsCorrect: teamAnswer.TeamAnswer == correctAnswer,
	}, teamAnswer.AnswerMap, nil
}

func NewAnswerUsecase(gm *core.GameManager) *AnswerUsecase {
	return &AnswerUsecase{
		gm: gm,
	}
}