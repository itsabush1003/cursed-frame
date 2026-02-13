package usecase

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/util"
)

type AdminStartQuestUsecase struct {
	gm  *core.GameManager
	ur  IUserRepository
	uir IUserImageRepository
	upr IUserProfileRepository
	pqr IProfileQuestionRepository
}

func (asqu *AdminStartQuestUsecase) Execute(
	networkCtx context.Context,
	onTick func(core.Quiz, string) error,
	failedCallback func(error) error,
) error {
	goNext, err := asqu.gm.QuestStart()
	if err != nil {
		return failedCallback(err)
	}
	questions, err := asqu.pqr.FetchAllQuestions()
	if err != nil {
		return failedCallback(err)
	}
	teams := asqu.gm.GetTeams()
	teamIDs := make([]core.TeamID, 0, len(teams))
	for tid := range teams {
		teamIDs = append(teamIDs, tid)
	}
	shuffledTIDs := util.ShuffleSlice(teamIDs)
	for _, tid := range shuffledTIDs {
		teamUsers := teams[tid]
		shuffledUsers := util.ShuffleSlice(teamUsers)
		shuffledQuestions := util.ShuffleSlice(questions)
		if len(shuffledUsers) > len(shuffledQuestions) {
			for {
				shuffledQuestions = append(shuffledQuestions, util.ShuffleSlice(questions)...)
				if len(shuffledQuestions) >= len(shuffledUsers) {
					break
				}
			}
		}
		shuffledQuestions = shuffledQuestions[:len(shuffledUsers)]
		for i, uid := range shuffledUsers {
			imageID, err := asqu.uir.FetchByUserID(uid)
			if err != nil {
				imageID = "NotFoundImage"
			}
			question := shuffledQuestions[i]
			correctProfile, err := asqu.upr.FetchByProfileIDWithUserGroup(question.GetQuestionID(), []uuid.UUID{uid})
			if err != nil {
				// 正答が取得できないとクイズにならないので仕方なくスキップ
				continue
			}
			profileCandidates, err := asqu.upr.FetchByProfileIDWithUserGroup(question.GetQuestionID(), shuffledUsers)
			if err != nil {
				// とりあえず正答に適当な選択肢を足して２択に
				profileCandidates = correctProfile
				profile, _ := model.NewUserProfile(uid, (&correctProfile[0]).GetProfileID(), question.GetSampleAnswer())
				profileCandidates = append(profileCandidates, *profile)
			}
			choiceCandidates := make([]string, 0, len(profileCandidates))
			for _, profile := range profileCandidates {
				choiceCandidates = append(choiceCandidates, profile.GetAnswer())
			}
			// Sort -> Compactで重複を削除
			slices.Sort(choiceCandidates)
			choiceCandidates = util.ShuffleSlice(slices.Compact(choiceCandidates))
			if len(choiceCandidates) > core.MaxChoiceNum {
				choiceCandidates = choiceCandidates[:core.MaxChoiceNum]
				if !slices.Contains(choiceCandidates, correctProfile[0].GetAnswer()) {
					choiceCandidates[0] = correctProfile[0].GetAnswer()
					choiceCandidates = util.ShuffleSlice(choiceCandidates)
				}
			}
			quiz := core.Quiz{
				ImageID:      imageID,
				TeamID:       core.TeamID(tid),
				QuestionID:   question.GetQuestionID(),
				QuestionText: question.GetQuestionText(),
				Choices:      make([]core.Choice, len(choiceCandidates)),
			}
			var correctChoiceID uint
			for i, choice := range choiceCandidates {
				quiz.Choices[i] = core.Choice{
					Target:     uid,
					ChoiceID:   uint(i),
					ChoiceText: choice,
				}
				if choice == correctProfile[0].GetAnswer() {
					correctChoiceID = uint(i)
				}
			}
			var remaindTime int = core.InitialRemaindTime
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			var onTickFailedCount int = 0
			var hint string = ""
		quizLoop:
			for {
				select {
				case <-goNext:
					break quizLoop
				case <-networkCtx.Done():
					return failedCallback(networkCtx.Err())
				case hint = <-asqu.gm.CheckHint():
					if remaindTime > 0 {
						remaindTime += core.IncreaseTimeHintTaken
					}
				case <-ticker.C:
					quiz.RemainedTime = remaindTime
					asqu.gm.Broadcast(uid, quiz, core.Choice{
						Target:     uid,
						ChoiceID:   correctChoiceID,
						ChoiceText: correctProfile[0].GetAnswer(),
					})
					if err := onTick(quiz, hint); err != nil {
						onTickFailedCount++
						if onTickFailedCount > MaxFailedCount {
							return failedCallback(err)
						}
					} else {
						onTickFailedCount = 0
					}
					remaindTime--
				}
			}
		}
	}

	return nil
}

func NewAdminStartQuestUsecase(
	gm *core.GameManager,
	ur IUserRepository,
	uir IUserImageRepository,
	upr IUserProfileRepository,
	pqr IProfileQuestionRepository,
) *AdminStartQuestUsecase {
	return &AdminStartQuestUsecase{
		gm:  gm,
		ur:  ur,
		uir: uir,
		upr: upr,
		pqr: pqr,
	}
}
