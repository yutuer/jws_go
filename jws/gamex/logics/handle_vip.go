package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/modules/friend"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
)

func addVipHandle(a *Account) {
	acc := a.Account
	acc.AddHandle(events.NewHandler().WithOnVIPLvUp(func(toLv uint32) {
		// 神将
		a.Profile.GetDestinyGeneral().OnVipLvUp(int(toLv), a.Profile.GetProfileNowTime())
		// 工会
		guild.GetModule(a.AccountID.ShardId).UpdateAccountInfo(a.GetSimpleInfo())
		// 好友
		simpleInfo := a.GetSimpleInfo()
		friend.GetModule(a.AccountID.ShardId).UpdateFriendInfo(&simpleInfo, 0)

	}))
}
