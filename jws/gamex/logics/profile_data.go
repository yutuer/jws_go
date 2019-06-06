package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	modGuild "vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/modules/team_pvp"
	"vcs.taiyouxi.net/jws/gamex/modules/ws_pvp"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const GsCheckTime = 30
const FirstIntoCorpLevel = 5

func (a *Account) InitData() bool {
	data := a.Profile.GetData()
	if data.IsNeedInit() {
		data.CorpCurrGS = GetCurrGS(a.Account)
		data.SetNoNeedCheckMaxGS()
		logs.Trace("InitData %v", data.CorpCurrGS)

		//战力榜应该加入进榜条件, 防止当玩家第一次存档前进入TopN而导致的读取玩家存档错误
		lv, _ := a.Profile.GetCorp().GetXpInfo()
		if lv >= FirstIntoCorpLevel {
			simpleInfo := a.Account.GetSimpleInfo()
			rank.GetModule(a.AccountID.ShardId).RankCorpGs.Add(&simpleInfo, int64(data.CorpCurrGS), int64(data.CorpCurrGS))
		}
		data.SetNoNeedInit()

		a.updateCSRobPlayerRank(false)
		return true
	} else {
		return false
	}
}

func (a *Account) CheckChangeMaxGS() {
	data := a.Profile.GetData()
	mayChange, oldScore := a.OnMaybeChangeMaxGS()

	if mayChange || data.GetNeedCheckGS() {
		//战力榜应该加入进榜条件, 防止当玩家第一次存档前进入TopN而导致的读取玩家存档错误
		//战力本身更新会很频繁, 相关的排行榜和公会中的信息更新不能保持这样快的频率
		//所以这里采取延迟更新的方法, 另一方面登入和登出的更新不受此影响
		//nowT := time.Now().Unix()
		//if (nowT - data.GetLastGsUpdateTime()) < GsCheckTime {
		//	return
		//}
		//data.SetLastGsUpdateTime(nowT)
		data.SetNeedCheckGS(false)

		simpleInfo := a.GetSimpleInfo()

		lv, _ := a.Profile.GetCorp().GetXpInfo()
		if lv >= FirstIntoCorpLevel {
			logs.Trace("oldScore %v", oldScore)
			rank.GetModule(a.AccountID.ShardId).RankCorpGs.Add(&simpleInfo, int64(data.CorpCurrGS), int64(oldScore))
			rank.GetModule(a.AccountID.ShardId).RankByHeroWuShuangGs.Add(&simpleInfo)
			a.UpdateCountryGs(&simpleInfo)
		}
		if simpleInfo.MaxTrialLv > 0 {
			rank.GetModule(a.AccountID.ShardId).RankByCorpTrial.UpdateIfInTopN(&simpleInfo)
		}
		modGuild.GetModule(a.AccountID.ShardId).UpdateAccountInfo(simpleInfo)

		if account.CondCheck(gamedata.Mod_TeamPvp, a.Account) {
			team_pvp.GetModule(a.AccountID.ShardId).UpdateInfo(&simpleInfo)
		}
		a.Profile.GetData().SetNeedUpdateFriend(true)
		a.Profile.GetTitle().OnGs(a.Account)
		a.Profile.GetMarketActivitys().OnHeroFundByGs(a.AccountID.String(), simpleInfo.CurrCorpGs, a.GetProfileNowTime())
		if a.Profile.WSPVPPersonalInfo.Rank != 0 {
			ws_pvp.GetModule(a.AccountID.ShardId).UpdatePlayer(a.convertWspvpInfo())
		}

		a.updateCSRobPlayerRank(false)
	}
	//下次时间到了在更新 如果需要
	//data.SetNeedCheckGS(mayChange || data.GetNeedCheckGS())
}

func (a *Account) OnMaybeChangeMaxGS() (bool, int) {
	data := a.Profile.GetData()

	is_updated := a.InitData()
	if !is_updated && data.IsNeedCheckMaxGS() {
		oldScore := data.CorpCurrGS
		gs_now := GetCurrGS(a.Account)
		logs.Trace("Now GS %d", gs_now)
		data.SetNoNeedCheckMaxGS()
		return true, oldScore
	}
	return is_updated, data.CorpCurrGS
}

func GetCurrGS(a *account.Account) int {
	data := a.Profile.GetData()
	return data.GetCurrGS(a)
}

func (a *Account) updateCSRobPlayerRank(force bool) {
	if a.GetCorpLv() < gamedata.CSRobJoinLevelLimit() {
		return
	}
	// 自己是否不在公会中
	if !a.GuildProfile.InGuild() && false == force {
		return
	}

	data := a.Profile.GetData()
	natList := gamedata.CSRobNatList()
	for _, nat := range natList {
		heroList := map[int]int{}
		for idx, gs := range data.HeroGs {
			hero := gamedata.GetHeroData(idx)
			if nil == hero || gs < 1 || hero.Nationality != nat {
				continue
			}

			heroList[idx] = gs
		}

		list := util.SortIntMapKeyByValue(heroList)
		formation := list[:]
		if 3 < len(list) {
			formation = list[:3]
		}
		team := a.buildHeroListFromRank(formation)
		csrob.GetModule(a.AccountID.ShardId).PlayerRanker.Add(a.AccountID.String(), nat, team)
	}

	//通知刷新自己的信息
	csrob.GetModule(a.AccountID.ShardId).PlayerMod.RefreshPlayerCacheBySelf(
		a.AccountID.String(),
		a.GuildProfile.GuildUUID,
		a.Profile.Name,
		a.GuildProfile.GetCurrPosition(),
	)
}
