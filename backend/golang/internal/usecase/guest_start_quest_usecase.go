package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
)

type GuestStartQuestUsecase struct {
	gm *core.GameManager
}

func (gsqu *GuestStartQuestUsecase) Execute(
	networkCtx context.Context,
	uid uuid.UUID,
	onRead func(core.Quiz) error,
	failedCallback func(error) error,
) error {
	ctx, quizCh, err := gsqu.gm.EnterQuestRoom(uid)
	if err != nil {
		return failedCallback(err)
	}
	var onReadFailedCount int = 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-networkCtx.Done():
			return failedCallback(networkCtx.Err())
		case quiz := <-quizCh:
			if err := onRead(quiz); err != nil {
				onReadFailedCount++
				if onReadFailedCount > MaxFailedCount {
					return failedCallback(err)
				}
			} else {
				onReadFailedCount = 0
			}
		}
	}
}

func NewGuestStartQuestUsecase(gm *core.GameManager) *GuestStartQuestUsecase {
	return &GuestStartQuestUsecase{
		gm: gm,
	}
}
