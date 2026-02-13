package usecase

import (
	"github.com/google/uuid"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

const NoMoreQuestionText string = "これで質問は全て終わりです。御回答ありがとうございました。ゲーム開始まで暫くお待ちください。"

type ProfileQuestionDTO struct {
	QuestionID   uint
	QuestionText string
	NoMoreAnswer bool
}

type UserProfileDTO struct {
	UserID    uuid.UUID
	ProfileID uint
	Answer    string
}

func (profile UserProfileDTO) ToUserProfileModel() (*model.UserProfile, error) {
	return model.NewUserProfile(profile.UserID, profile.ProfileID, profile.Answer)
}

type RegistProfileUsecase struct {
	pqr IProfileQuestionRepository
	upr IUserProfileRepository
}

func (rpu *RegistProfileUsecase) Execute(profile UserProfileDTO) (ProfileQuestionDTO, error) {
	up, err := profile.ToUserProfileModel()
	if err != nil {
		return ProfileQuestionDTO{}, err
	}
	err = rpu.upr.Save(up)
	if err != nil {
		return ProfileQuestionDTO{}, err
	}
	nextQuestionID := profile.ProfileID + 1
	question, err := rpu.pqr.FetchByQuestionID(nextQuestionID)
	if err != nil {
		// Repositoryに次のIDがない＝今のAnswerが最後の質問のもの
		return ProfileQuestionDTO{
			QuestionID:   nextQuestionID,
			QuestionText: NoMoreQuestionText,
			NoMoreAnswer: true,
		}, nil
	}
	return ProfileQuestionDTO{
		QuestionID:   question.GetQuestionID(),
		QuestionText: question.GetQuestionText(),
		NoMoreAnswer: false,
	}, nil
}

func NewRegistProfileUsecase(pqr IProfileQuestionRepository, upr IUserProfileRepository) *RegistProfileUsecase {
	return &RegistProfileUsecase{
		pqr: pqr,
		upr: upr,
	}
}
