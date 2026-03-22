package usecase

import (
	"errors"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

type GetTeamInfoUsecase struct {
	ur IUserRepository
}

func (gtu *GetTeamInfoUsecase) Execute(user *model.User) (uint32, string, []string, error) {
	if user.GetTeamID() == model.UNDEFINED.Raw() {
		return 0, model.UNDEFINED.String(), []string{}, errors.New("Teams have not been splitted yet")
	}

	members, err := gtu.ur.FetchByTeamID(user.GetTeamID())
	if err != nil {
		return 0, model.UNDEFINED.String(), []string{}, err
	}
	memberNames := make([]string, 0, len(members)-1)
	for _, member := range members {
		if member.GetUserID() == user.GetUserID() {
			continue
		}
		memberNames = append(memberNames, member.GetName())
	}
	return user.GetTeamID(), model.TeamColor(user.GetTeamID()).String(), memberNames, nil
}

func NewGetTeamInfoUsecase(ur IUserRepository) *GetTeamInfoUsecase {
	return &GetTeamInfoUsecase{
		ur: ur,
	}
}
