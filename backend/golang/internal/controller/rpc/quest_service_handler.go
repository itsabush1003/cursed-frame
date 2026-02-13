package controller

import (
	"context"
	"errors"
	"html"

	"connectrpc.com/connect"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/controller/middleware"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
	commonv1 "github.com/itsuabush1003/cursed-frame/backend/golang/internal/gen/common/v1"
	questv1 "github.com/itsuabush1003/cursed-frame/backend/golang/internal/gen/quest/v1"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/gen/quest/v1/questv1connect"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

type QuestServiceHandler struct {
	questv1connect.UnimplementedQuestServiceHandler
	gsqu *usecase.GuestStartQuestUsecase
	au   *usecase.AnswerUsecase
	thu  *usecase.TakeHintUsecase
	gru  *usecase.GetResultUsecase
}

func (qsh *QuestServiceHandler) StartQuest(ctx context.Context, r *connect.Request[emptypb.Empty], stream *connect.ServerStream[questv1.StartQuestResponse]) error {
	user := middleware.GetUserFromCtx(ctx)
	if user == nil {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("Unauthenticated Access"))
	}

	if err := qsh.gsqu.Execute(
		ctx,
		user.GetUserID(),
		func(quiz core.Quiz) error {
			choices := make([]*commonv1.Choice, 0, len(quiz.Choices))
			for _, c := range quiz.Choices {
				choices = append(choices, &commonv1.Choice{
					ChoiceId:   uint32(c.ChoiceID),
					ChoiceText: c.ChoiceText,
				})
			}
			return stream.Send(&questv1.StartQuestResponse{
				TargetUserImageId: quiz.ImageID,
				TargetTeamId:      uint32(quiz.TeamID),
				QuestionId:        uint32(quiz.QuestionID),
				Question:          quiz.QuestionText,
				Choices:           choices,
				CanAnswer:         user.GetTeamID() != uint32(quiz.TeamID),
				LastTime:          int32(quiz.RemainedTime),
			})
		},
		func(err error) error {
			return connect.NewError(connect.CodeCanceled, err)
		},
	); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (qsh *QuestServiceHandler) Answer(ctx context.Context, r *connect.Request[questv1.AnswerRequest]) (*connect.Response[questv1.AnswerResponse], error) {
	user := middleware.GetUserFromCtx(ctx)
	if user == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("Unauthenticated Access"))
	}

	teamAnswer, answerMap, err := qsh.au.Execute(user, usecase.AnswerDTO{
		ChoiceID:   uint(r.Msg.Answer.ChoiceId),
		ChoiceText: r.Msg.Answer.ChoiceText,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	answerCount := make([]int32, len(answerMap))
	for cid, cnt := range answerMap {
		answerCount[cid] = int32(cnt)
	}
	return connect.NewResponse(&questv1.AnswerResponse{
		IsCorrect: teamAnswer.IsCorrect,
		TeamAnswer: &commonv1.Choice{
			ChoiceId:   uint32(teamAnswer.Answer.ChoiceID),
			ChoiceText: teamAnswer.Answer.ChoiceText,
		},
		AnswerCount: answerCount,
	}), nil
}

func (qsh *QuestServiceHandler) TakeHint(ctx context.Context, r *connect.Request[questv1.TakeHintRequest]) (*connect.Response[emptypb.Empty], error) {
	user := middleware.GetUserFromCtx(ctx)
	if user == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("Unauthenticated Access"))
	}
	hint := html.EscapeString(r.Msg.Hint)

	if err := qsh.thu.Execute(user.GetUserID(), hint); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (qsh *QuestServiceHandler) GetResult(ctx context.Context, r *connect.Request[emptypb.Empty]) (*connect.Response[questv1.GetResultResponse], error) {
	user := middleware.GetUserFromCtx(ctx)
	if user == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("Unauthenticated Access"))
	}

	resultState, personalStats, teamStats, err := qsh.gru.Execute(user.GetUserID(), user.GetTeamID())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&questv1.GetResultResponse{
		Result:        commonv1.Result(resultState),
		TeamOrder:     uint32(teamStats.Order),
		PersonalOrder: uint32(personalStats.Order),
		PersonalRate:  personalStats.CorrectRate,
	}), nil
}

func NewQuestServiceHandler(
	gsqu *usecase.GuestStartQuestUsecase,
	au *usecase.AnswerUsecase,
	thu *usecase.TakeHintUsecase,
	gru *usecase.GetResultUsecase,
) *QuestServiceHandler {
	return &QuestServiceHandler{
		gsqu: gsqu,
		au:   au,
		thu:  thu,
		gru:  gru,
	}
}
