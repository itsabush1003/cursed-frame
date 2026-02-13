package usecase

import (
	"github.com/google/uuid"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
)

type GetResultUsecase struct {
	gm *core.GameManager
	resultStateMapper func(float32) int32
}

func (gru *GetResultUsecase) Execute(uid uuid.UUID, tid uint32) (resultState int32, personal core.Stats, team core.Stats, err error) {
	totalRate, personalStats, teamStats, err := gru.gm.GetResultStats(uid, core.TeamID(tid))
	if err != nil {
		return 0, core.Stats{}, core.Stats{}, err
	}
	return gru.resultStateMapper(totalRate), personalStats, teamStats, nil
}

func NewGetResultUsecase(gm *core.GameManager, mapper func(float32) int32) *GetResultUsecase {
	return &GetResultUsecase{
		gm: gm,
		resultStateMapper: mapper,
	}
}