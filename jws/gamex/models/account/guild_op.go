package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/guild"
	"vcs.taiyouxi.net/platform/planx/servers/db"
)

type PlayerGuild struct {
	guild.Guild
}

func NewPlayerGuild(account db.Account) PlayerGuild {
	AccountID := account

	my := PlayerGuild{
		Guild: *guild.NewPlayerGuild(AccountID),
	}
	return my
}
