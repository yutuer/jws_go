package logics

import (
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ActivateHeroDestiny : 激活指定的宿命
// 激活指定的宿命
func (p *Account) ActivateHeroDestinyHandler(req *reqMsgActivateHeroDestiny, resp *rspMsgActivateHeroDestiny) uint32 {
	destinyId := int(req.DestinyId)
	heroDestiny := p.Profile.GetHeroDestiny()

	// 检查是否达到最大等级
	destinyInfo := heroDestiny.GetHeroDestinyById(destinyId)
	if destinyInfo != nil && destinyInfo.Level >= gamedata.GetFateMaxLevel(destinyId) {
		logs.Warn("hero destiny is already max level")
		return errCode.HeroDestinyHasActivated
	}

	// 设置要激活的等级
	var nextLevel int
	if destinyInfo == nil {
		nextLevel = 1
	} else {
		nextLevel = destinyInfo.Level + 1
	}

	// 读取配置
	cfg := gamedata.GetHeroDestinyById(destinyId)
	if cfg == nil {
		logs.Warn("hero destiny cfg is not exsit %d", destinyId)
		return errCode.CommonInner
	}

	// 检查条件满足
	if activateCode := canActivate(p, cfg); activateCode != 0 {
		return activateCode
	}

	// 读取等级配置
	levelCfg := gamedata.GetFateLevelConfig(destinyId, nextLevel)
	if levelCfg == nil {
		logs.Warn("hero fate level config is not found, %d, %d", destinyId, nextLevel)
		return errCode.CommonInner
	}

	// 消费
	costData := &gamedata.CostData{}
	costData.AddItem(levelCfg.GetFateActiveCoin(), levelCfg.GetFateActiveCount())
	reason := fmt.Sprintf("activate hero destiny %d", heroDestiny)
	if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
		logs.Warn("activate hero destiny cost err")
		return errCode.ClickTooQuickly
	}

	// 设置已经激活
	heroDestiny.AddOrUpdate(destinyId, nextLevel)

	logiclog.LogHeroDestiny(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		p.Profile.GetVipLevel(),
		p.Profile.GetData().CorpCurrGS,
		destinyId,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	p.Profile.GetData().SetNeedCheckMaxGS()
	//更新羁绊信息
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByHeroDestiny.Add(&info)

	resp.OnChangeHeroDestiny()
	return 0
}

func canActivate(p *Account, cfg *gamedata.HeroDestinyConfig) uint32 {
	if p.Profile.GetCorp().GetLvlInfo() < gamedata.GetHeroCommonConfig().GetFateOpenLevel() {
		logs.Warn("hero destiny corp lv is not reached %d", p.Profile.GetCorp().GetLvlInfo())
		return errCode.HeroDestinyConditionError
	}
	heroList := cfg.AvatarIds
	for _, avatarId := range heroList {
		if !p.Profile.GetCorp().IsAvatarHasUnlock(avatarId) {
			logs.Warn("hero destiny condition is not enough, avatarId %d", avatarId)
			return errCode.HeroDestinyConditionError
		}
	}
	return 0
}
