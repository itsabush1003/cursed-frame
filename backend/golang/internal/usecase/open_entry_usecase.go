package usecase

import (
	"context"
	"time"

	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/core"
	"github.com/itsuabush1003/cursed-frame/backend/golang/internal/model"
)

type OpenEntryUsecase struct {
	gm *core.GameManager
	ur IUserRepository
}

func (oeu *OpenEntryUsecase) Execute(
	networkCtx context.Context,
	onTick func([]model.User) error,
	doneCallback func(),
	failedCallback func(error) error,
) error {
	ctx, err := oeu.gm.OpenLobby()
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
			// teamIDの確定を通知する
			// 本来はteamID確定後もtickが続くようにすべきだが、一旦これで暫定対応
			// users取得をselectの前に出して共通化すると情報が古くなってしまうのでコピペで
			uids := oeu.gm.GetLobbyUsers()
			users, _ := oeu.ur.FetchByUserIDs(uids)
			_ = onTick(users)
			return nil
		case <-networkCtx.Done():
			return failedCallback(networkCtx.Err())
		case <-ticker.C:
			uids := oeu.gm.GetLobbyUsers()
			users, _ := oeu.ur.FetchByUserIDs(uids)
			if err := onTick(users); err != nil {
				onTickFailedCount++
				if onTickFailedCount > MaxFailedCount {
					return failedCallback(err)
				}
			} else {
				onTickFailedCount = 0
			}
		}
	}
}

func NewOpenEntryUsecase(gm *core.GameManager, ur IUserRepository) *OpenEntryUsecase {
	return &OpenEntryUsecase{
		gm: gm,
		ur: ur,
	}
}
