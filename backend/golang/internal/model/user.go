package model

import (
	"github.com/google/uuid"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/util"
)

const TokenLength int = 8

type User struct {
	userID      uuid.UUID
	name        string
	accessToken string
	teamID      uint32
	isReady     bool
	version     uint
}

func (u User) GetUserID() uuid.UUID {
	return u.userID
}

func (u User) GetName() string {
	return u.name
}

func (u User) GetAccessToken() string {
	return u.accessToken
}

func (u *User) RefreshAccessToken() error {
	newToken, err := util.CreateRandStr(TokenLength)
	if err != nil {
		return err
	}
	u.accessToken = newToken
	return nil
}

func (u User) GetTeamID() uint32 {
	return u.teamID
}

func (u *User) SetTeamID(tid uint32) {
	u.teamID = tid
}

func (u User) GetIsReady() bool {
	return u.isReady
}

func (u *User) SetReady() {
	u.isReady = true
}

func (u User) GetVersion() uint {
	return u.version
}

func (u *User) IncrementVersion() {
	u.version += 1
}

func NewUser(name string) (*User, error) {
	userID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	token, err := util.CreateRandStr(TokenLength)
	if err != nil {
		return nil, err
	}

	return &User{
		userID:      userID,
		name:        name,
		accessToken: token,
		teamID:      UNDEFINED.Raw(),
		isReady:     false,
		version:     1,
	}, nil
}

func ReconstructUser(userID string, name string, token string, teamID int, isReady bool, version int) (*User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return &User{
		userID:      uid,
		name:        name,
		accessToken: token,
		teamID:      uint32(teamID),
		isReady:     isReady,
		version:     uint(version),
	}, nil
}
