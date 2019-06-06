package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/general"
	"vcs.taiyouxi.net/platform/planx/servers/db"
)

type PlayerGenerals struct {
	general.PlayerGenerals
}

func NewPlayerGenerals(account db.Account) PlayerGenerals {
	AccountID := account

	my := PlayerGenerals{
		PlayerGenerals: *general.NewPlayerGenerals(AccountID),
	}
	return my
}
