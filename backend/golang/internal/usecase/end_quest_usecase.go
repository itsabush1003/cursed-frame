package usecase

import (
	"github.com/google/uuid"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
)

type UserStatsDTO struct {
	TeamID        core.TeamID
	UserID        uuid.UUID
	UserName      string
	CorrectRate   float32
	PersonalOrder uint32
}

type TeamStatsDTO struct {
	MembersStats []UserStatsDTO
	CorrectRate  float32
	TeamOrder    uint32
}

type EndQuestUsecase struct {
	gm                *core.GameManager
	ur                IUserRepository
	resultStateMapper func(float32) int32
}

func (equ *EndQuestUsecase) Execute() (int32, map[core.TeamID]TeamStatsDTO, error) {
	if err := equ.gm.EndQuest(); err != nil {
		return 0, nil, err
	}

	totalRate, usersStats, teamsStats, err := equ.gm.GetAllStats()
	if err != nil {
		return 0, nil, err
	}
	teamStats := make(map[core.TeamID]TeamStatsDTO, len(teamsStats))
	for tid, teamStat := range teamsStats {
		members, err := equ.ur.FetchByTeamID(uint32(tid))
		if err != nil {
			continue
		}
		userStatsList := make([]UserStatsDTO, 0, len(members))
		for _, user := range members {
			userStatsList = append(userStatsList, UserStatsDTO{
				TeamID:        tid,
				UserID:        user.GetUserID(),
				UserName:      user.GetName(),
				CorrectRate:   usersStats[user.GetUserID()].CorrectRate,
				PersonalOrder: uint32(usersStats[user.GetUserID()].Order),
			})
		}
		teamStats[tid] = TeamStatsDTO{
			MembersStats: userStatsList,
			CorrectRate:  teamStat.CorrectRate,
			TeamOrder:    uint32(teamStat.Order),
		}
	}

	return equ.resultStateMapper(totalRate), teamStats, nil
}

func NewEndQuestUsecase(gm *core.GameManager, ur IUserRepository, mapper func(float32) int32) *EndQuestUsecase {
	return &EndQuestUsecase{
		gm:                gm,
		ur:                ur,
		resultStateMapper: mapper,
	}
}
