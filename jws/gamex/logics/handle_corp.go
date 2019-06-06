package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func addCorpLvUpHandle(a *Account) {
	acc := a.Account
	a.AddHandle(events.NewHandler().WithOnCorpLvUp(func(toLv, toExp uint32, reason string) {
		guild.GetModule(a.AccountID.ShardId).UpdateAccountInfo(a.GetSimpleInfo())

		profile := &acc.Profile
		for avatarID := 0; avatarID < gamedata.AVATAR_NUM_CURR; avatarID++ {
			if !acc.IsAvatarUnblock(avatarID) {
				logs.Trace("WithOnCorpLvUp avatarID %d unblock", avatarID)
				continue
			}

			// MaxGS可能变化 1. 角色升级
			profile.GetData().SetNeedCheckMaxGS()
		}
		//logs.Error("DestinyGeneralUnlockFirstLv %v %v", toLv, gamedata.DestinyGeneralUnlockFirstLv)
		if toLv == gamedata.DestinyGeneralUnlockFirstLv {
			acc.Profile.GetDestinyGeneral().AddNewGeneral(0)
			//logs.Trace("DestinyGeneralUnlockFirstLv %v", profile.GetDestinyGeneral())
		}

		if a.trialFirstActivate() {
			a.Tmp.TrialFirst = true
		}

		// 等级限购礼包
		a.Profile.GetIAPGoodInfo().OnPlayerLevelUp(toLv, profile.GetProfileNowTime())

		// 0.战队等级达到P1
		a.updateCondition(account.COND_TYP_Corp_lv,
			0, 0, "", "", nil)

		// 更新等级排行榜
		info := a.GetSimpleInfo()
		rank.GetModule(a.AccountID.ShardId).RankByCorpLv.Add(&info)

		// 清除出奇制胜武将上次最大积分
		if toLv > 60 && toLv%5 == 0 {
			a.Profile.GetHeroDiff().ClearMaxScore()
		}

	}))
}

func addCorpExpAddHandle(a *Account) {
	a.AddHandle(events.NewHandler().WithOnCorpExpAdd(func(oldV, chgV uint32, reason string) {
		logiclog.LogCorpExpChg(a.AccountID.String(), a.Profile.GetCurrAvatar(), a.Profile.GetCorp().GetLvlInfo(),
			a.Profile.ChannelId, reason, oldV, chgV,
			func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
	}))
}
