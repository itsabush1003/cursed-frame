package usecase

import (
	"errors"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

type GetTeamIDUsecase struct{}

func (gtu *GetTeamIDUsecase) Execute(user *model.User) (uint32, string, error) {
	if user.GetTeamID() == model.UNDEFINED.Raw() {
		return 0, model.UNDEFINED.String(), errors.New("Teams have not been splitted yet")
	}
	return user.GetTeamID(), model.TeamColor(user.GetTeamID()).String(), nil
}

func NewGetTeamIDUsecase() *GetTeamIDUsecase {
	return &GetTeamIDUsecase{}
}
