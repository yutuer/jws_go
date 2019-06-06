package update

import (
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func V7toV8(FromVersion int64, justLoad bool, acc *account.Account) error {
	if FromVersion != acc.Profile.Ver {
		return fmt.Errorf("V7toV8 err FromVersion %d Profile.Ver %d",
			FromVersion, acc.Profile.Ver)
	}

	heroDestinyV7ToV8(acc)

	acc.Profile.Ver = FromVersion + 1
	logs.Debug("V7toV8 %s %d", acc.AccountID.String(), acc.Profile.Ver)
	return nil
}

func heroDestinyV7ToV8(acc *account.Account) {
	destiny := acc.Profile.GetHeroDestiny()
	destiny.NewActivateDestiny = make([]account.DestinyInfo, 0)
	for _, id := range destiny.ActivateDestiny {
		maxLevel := gamedata.GetFateMaxLevel(id)
		destiny.NewActivateDestiny = append(destiny.NewActivateDestiny, account.DestinyInfo{
			Id:    id,
			Level: maxLevel,
		})
	}
}
