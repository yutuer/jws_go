package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
	"vcs.taiyouxi.net/jws/gamex/modules/hero_diff"
	"vcs.taiyouxi.net/jws/gamex/modules/herogacharace"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/modules/team_pvp"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const CODE_ERR_Save = 1

// warn, err
func (p *Account) rename(newName string, resp helper.ISyncRsp) (uint32, uint32) {
	// 检查消耗
	config := gamedata.GetRenameCostConfig(p.Profile.RenameCount + 1)
	if !p.Profile.GetHC().HasHC(int64(config.GetReNamePrice())) {
		return errCode.MaterialNotEnough, 0
	}

	// 数据库操作
	oldName := p.Profile.Name
	acid := p.AccountID.String()
	isSuccess, err := driver.RenameToRedis(oldName, newName, acid, p.AccountID.ShardId)
	if err != nil {
		logs.Warn("RenameToRedis err %s", err.Error())
		isSuccess, err = driver.RenameFor203(oldName, newName, acid, p.AccountID.ShardId)
		if err != nil {
			logs.SentryLogicCritical(acid, "rename RenameFor203 Err by %s", err.Error())
			return 0, CODE_ERR_Save
		}
	}
	if !isSuccess {
		return errCode.RenameNameHasExit, 0
	}

	// 消耗
	costData := &gamedata.CostData{}
	costData.AddItem(gamedata.VI_Hc, config.GetReNamePrice())

	reason := fmt.Sprintf("rename count %d", p.Profile.RenameCount+1)
	if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
		return errCode.ClickTooQuickly, 0
	}

	// 再修改内存
	p.Profile.RenameCount++
	p.Profile.Name = newName
	p.SimpleInfoProfile.Name = newName
	newSimpleInfo := p.GetSimpleInfo()
	simpleInfo := &newSimpleInfo

	logiclog.LogCommonInfo(p.getBIBaseInfo(), logiclog.PlayerChangeName{
		BeforeName: oldName,
		AfterName:  newName,
	}, logiclog.LogicTag_NicknameChg, "")

	// 最后更新各个功能， 这些功能即使这次失败也会由其他模块
	go func() {
		defer logs.PanicCatcherWithInfo("rename panic")
		p.updateGuild(simpleInfo)
		p.updateTeamPvp(simpleInfo)
		p.updateRank(simpleInfo)
		p.updateHeroGachaRace(oldName, simpleInfo)
		p.updateGvg(simpleInfo)
	}()

	return 0, 0
}

func (p *Account) updateGuild(simpleInfo *helper.AccountSimpleInfo) {
	guild.GetModule(p.AccountID.ShardId).UpdateAccountInfo(*simpleInfo)
}

func (p *Account) updateTeamPvp(simpleInfo *helper.AccountSimpleInfo) {
	team_pvp.GetModule(p.AccountID.ShardId).UpdateInfo(simpleInfo)
}

func (p *Account) updateRank(simpleInfo *helper.AccountSimpleInfo) {
	// 战力排行榜
	data := p.Profile.GetData()
	_, oldScore := p.OnMaybeChangeMaxGS()
	lv, _ := p.Profile.GetCorp().GetXpInfo()
	rankModule := rank.GetModule(p.AccountID.ShardId)
	if lv >= FirstIntoCorpLevel {
		rankModule.RankCorpGs.Add(simpleInfo, int64(data.CorpCurrGS), int64(oldScore))
	}

	// 单人竞技排行榜
	rankModule.RankSimplePvp.OnChangeName(simpleInfo)

	// 爬塔
	if simpleInfo.MaxTrialLv > 0 {
		rank.GetModule(p.AccountID.ShardId).RankByCorpTrial.UpdateIfInTopN(simpleInfo)
	}
	for i, _ := range rank.GetModule(p.AccountID.ShardId).RankByHeroDiff {
		extraData := hero_diff.HeroDiffRankData{
			AcID:       p.AccountID.String(),
			FreqAvatar: p.Profile.GetHeroDiff().GetTopNFreqHero(i, gamedata.HeroDiffRankShowAvatarCount),
		}
		rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[i].UpdateIfInTopNWithExtraData(simpleInfo, extraData)
	}

	// 主将星级
	rank.GetModule(p.AccountID.ShardId).RankByHeroStar.UpdateIfInTopN(simpleInfo)

	// 军团排行
	if p.GuildProfile.GuildUUID != "" && p.GuildProfile.GetCurrPosition() == gamedata.Guild_Pos_Chief {
		guildInfo, guildRet := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
		if !guildRet.HasError() {
			rank.GetModule(p.AccountID.ShardId).RankGuildGs.OnGuildOrLeaderRename(&guildInfo.Base)
		}
	}
}

func (p *Account) updateGvg(simpleInfo *helper.AccountSimpleInfo) {
	gvg.GetModule(p.AccountID.ShardId).RenamePlayer(simpleInfo.AccountID, simpleInfo.Name)
}

func (p *Account) updateHeroGachaRace(oldName string, simpleInfo *helper.AccountSimpleInfo) {
	hgr := p.Profile.GetHeroGachaRaceInfo()
	if score := hgr.GetCurScore(); score > 0 {
		herogacharace.Get(p.AccountID.ShardId).OnPlayerRename(simpleInfo.AccountID,
			oldName,
			simpleInfo.Name,
			uint64(score))
	}

}
