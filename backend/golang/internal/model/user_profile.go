package model

import (
	"errors"

	"github.com/google/uuid"
)

const MaxAnswerLength int = 30

type UserProfile struct {
	userID    uuid.UUID
	profileID uint
	answer    string
}

func (up *UserProfile) GetUserID() uuid.UUID {
	return up.userID
}

func (up *UserProfile) GetProfileID() uint {
	return up.profileID
}

func (up *UserProfile) GetAnswer() string {
	return up.answer
}

func NewUserProfile(userID uuid.UUID, profileID uint, answer string) (*UserProfile, error) {
	if len(answer) > MaxAnswerLength {
		return nil, errors.New("Answer is too long")
	}
	return &UserProfile{
		userID:    userID,
		profileID: profileID,
		answer:    answer,
	}, nil
}
