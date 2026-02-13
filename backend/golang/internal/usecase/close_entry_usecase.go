package usecase

import (
	"errors"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

type CloseEntryUsecase struct {
	gm      *core.GameManager
	ur      IUserRepository
	teamNum int
}

func (ceu *CloseEntryUsecase) Execute() error {
	if err := ceu.gm.CloseLobby(); err != nil {
		return err
	}
	userIDs := ceu.gm.GetLobbyUsers()
	users, err := ceu.ur.FetchByUserIDs(userIDs)
	if err != nil {
		return err
	}

	for _, user := range users {
		if !user.GetIsReady() {
			return errors.New("Some users have not been ready yet")
		}
	}

	userTeam := ceu.gm.SplitTeams(userIDs, ceu.teamNum)
	teams := ceu.gm.GetTeams()
	for _, member := range teams {
		if len(member) < model.MinTeamUser {
			return errors.New("a team must have at least 3 users")
		}
	}
	for i, usr := range users {
		(&users[i]).SetTeamID(userTeam[usr.GetUserID()])
	}

	if err = ceu.ur.SaveBulk(users); err != nil {
		return err
	}

	ceu.gm.NotifyLobbyClosed()
	return nil
}

func NewCloseEntryUsecase(gm *core.GameManager, ur IUserRepository, N int) *CloseEntryUsecase {
	return &CloseEntryUsecase{
		gm:      gm,
		ur:      ur,
		teamNum: N,
	}
}
