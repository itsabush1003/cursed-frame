package usecase

import (
	"context"
	"maps"
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
	startCount, goNext, err := asqu.gm.QuestStart()
	if err != nil {
		return failedCallback(err)
	}
	questions, err := asqu.pqr.FetchAllQuestions()
	if err != nil {
		return failedCallback(err)
	}
	teams := asqu.gm.GetTeams()
	teamIDs := slices.Collect(maps.Keys(teams))
	shuffledTIDs := util.ShuffleSlice(teamIDs)
checkConnectionLoop:
	for {
		select {
		case <-networkCtx.Done():
			return failedCallback(networkCtx.Err())
		case <-time.After(time.Second):
			connected := asqu.gm.GetConnectedMembers()
			// 全チーム少なくとも一人以上の接続があるか
			if slices.Min(slices.Collect(maps.Values(connected))) > 0 {
				break checkConnectionLoop
			}
		}
	}
	ticker := time.NewTicker(time.Second)
	// quizLoopに入るまで止めておく
	ticker.Stop()
	defer ticker.Stop()
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
					ChoiceID:   uint(i + 1),
					ChoiceText: choice,
				}
				if choice == correctProfile[0].GetAnswer() {
					correctChoiceID = uint(i + 1)
				}
			}
			var remaindTime int = core.InitialRemaindTime
			var onTickFailedCount int = 0
			var hint string = ""
			var canCountdown bool = false
			ticker.Reset(time.Second)
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
				case <-startCount:
					canCountdown = true
				case <-ticker.C:
					quiz.RemainedTime = remaindTime
					asqu.gm.Broadcast(uid, quiz, core.Choice{
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
					if canCountdown {
						remaindTime--
					}
				}
			}
			ticker.Stop()
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
