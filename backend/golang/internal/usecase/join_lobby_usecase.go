package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
)

const TickDuration time.Duration = 5 * time.Second
const MaxFailedCount int = 3

type JoinLobbyUsecase struct {
	gm *core.GameManager
}

func (jlu *JoinLobbyUsecase) Execute(
	networkCtx context.Context,
	uid uuid.UUID,
	onTick func() error,
	doneCallback func(),
	failedCallback func(error) error,
) error {
	ctx, err := jlu.gm.JoinLobby(uid)
	if err != nil {
		return failedCallback(err)
	}
	ticker := time.NewTicker(TickDuration)
	defer ticker.Stop()
	var onTickFailedCount int = 0
	for {
		select {
		case <-ctx.Done():
			doneCallback()
			return nil
		case <-networkCtx.Done():
			jlu.gm.DisconnectLobby(uid)
			return failedCallback(networkCtx.Err())
		case <-ticker.C:
			if err := onTick(); err != nil {
				onTickFailedCount++
				if onTickFailedCount > MaxFailedCount {
					jlu.gm.DisconnectLobby(uid)
					return failedCallback(err)
				}
			} else {
				onTickFailedCount = 0
			}
		}
	}
}

func NewJoinLobbyUsecase(gm *core.GameManager) *JoinLobbyUsecase {
	return &JoinLobbyUsecase{
		gm: gm,
	}
}
