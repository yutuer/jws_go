package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func addChgAvatarHandle(a *Account) {
	acc := a.Account
	h := events.NewHandler()
	h.WithOnAvatarChg(func(avatarID int) {
		guild.GetModule(a.AccountID.ShardId).UpdateAccountInfo(a.GetSimpleInfo())
	})

	logs.Trace("addChgAvatarHandle %v", h)
	acc.AddHandle(h)
}
