package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	_                = iota
	Err_HeroID       // 传入的主将ID错误,没有该主将
	Err_ActSwingID   // 传入的神翼还没有激活
	Err_HeroLv       // 英雄等级不足,还无法开启神翼系统
	Err_HeroStarLv   // 英雄星级不足,还无法开启神翼系统
	Err_ActSwingType // 神翼类型不需要手动激活
	Err_SwingType    // 没有这种神翼类型
	Err_SwingStar    // 神翼星级不足,无法操作
	Err_Cost         // 玩家材料不足,无法操作
	Err_SwingLv      // 神翼等级或星级异常,不合理
	Err_Give         // 给予玩家奖励失败
)

const (
	Wing_LvUp   = 1
	Wing_StarUp = 2
)

// ChangeHeroSwing : 控制玩家神翼的显示
// 客户端通过此协议来决定穿哪个神翼

// reqMsgChangeHeroSwing 控制玩家神翼的显示请求消息定义
type reqMsgChangeHeroSwing struct {
	Req
	SwingID int64 `codec:"swing_id"` // 需要装备的神翼ID，为0时代表不装备
	HeroID  int64 `codec:"hero_id"`  // 所操作的主将
}

// rspMsgChangeHeroSwing 控制玩家神翼的显示回复消息定义
type rspMsgChangeHeroSwing struct {
	SyncResp
}

// ChangeHeroSwing 控制玩家神翼的显示: 客户端通过此协议来决定穿哪个神翼
func (p *Account) ChangeHeroSwing(r servers.Request) *servers.Response {
	req := new(reqMsgChangeHeroSwing)
	rsp := new(rspMsgChangeHeroSwing)

	initReqRsp(
		"Attr/ChangeHeroSwingRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	ok, errCode := p.commonWCheck(int(req.HeroID))
	if !ok {
		return rpcError(rsp, errCode)
	}
	hero := p.Profile.GetHero()
	if !hero.HasSwingAct(int(req.HeroID), int(req.SwingID)) {
		return rpcError(rsp, Err_ActSwingID)
	}
	hero.SetCurSwing(int(req.HeroID), int(req.SwingID))
	rsp.OnChangeUpdateHeroWing(int(req.HeroID))
	// logic imp end
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// HeroSwingAct : 主将手动神翼外观激活
// 激活神翼外观

// reqMsgHeroSwingAct 主将手动神翼外观激活请求消息定义
type reqMsgHeroSwingAct struct {
	Req
	SwingID int64 `codec:"swing_id"` // 需要激活的神翼ID
	HeroID  int64 `codec:"hero_id"`  // 所操作的主将ID
}

// rspMsgHeroSwingAct 主将手动神翼外观激活回复消息定义
type rspMsgHeroSwingAct struct {
	SyncResp
}

// HeroSwingAct 主将手动神翼外观激活: 激活神翼外观
func (p *Account) HeroSwingAct(r servers.Request) *servers.Response {
	req := new(reqMsgHeroSwingAct)
	rsp := new(rspMsgHeroSwingAct)

	initReqRsp(
		"Attr/HeroSwingActRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	ok, errCode := p.commonWCheck(int(req.HeroID))
	if !ok {
		return rpcError(rsp, errCode)
	}
	info := gamedata.GetHeroSwingInfo(int(req.SwingID))
	if info == nil {
		return rpcError(rsp, Err_SwingType)
	}
	if info.GetHWUnlockType() != gamedata.Unlock_Typ_Manual {
		return rpcError(rsp, Err_ActSwingType)
	}
	hero := p.Profile.GetHero()
	if hero.HeroSwings[int(req.HeroID)].StarLv < int(info.GetHWUnlockStar()) {
		return rpcError(rsp, Err_SwingStar)
	}
	if hero.HeroSwings[req.HeroID].HasAct(int(req.SwingID)) {
		logs.Warn("Have already Act Swing, HeroID: %d, SwingID: %d", req.HeroID, req.SwingID)
		return rpcSuccess(rsp)
	}
	costData := gamedata.CostData{}
	costData.AddItem(info.GetHWUnlockMaterial(), info.GetHWUnlockCount())
	if !account.CostBySync(p.Account, &costData, rsp, "Active Hero Wing") {
		return rpcError(rsp, Err_Cost)
	}
	// 更新宝石等级排行榜
	simpleInfo := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByWingStar.Add(&simpleInfo)

	hero.HeroSwings[req.HeroID].ActSwing(int(req.SwingID))
	rsp.OnChangeUpdateHeroWing(int(req.HeroID))
	logiclog.LogHeroWingAct(p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, 0,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		int(req.SwingID), int(req.HeroID), "")
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// HeroSwingLvUp : 主将神翼升级升星
// 提升神翼等级星级

// reqMsgHeroSwingLvUp 主将神翼升级升星请求消息定义
type reqMsgHeroSwingLvUp struct {
	Req
	HeroID int64 `codec:"hero_id"` // 所操作的主将ID
	Type   int64 `codec:"type"`    // 升级或者升星，type=1时为升级， type=2时为升星
}

// rspMsgHeroSwingLvUp 主将神翼升级升星回复消息定义
type rspMsgHeroSwingLvUp struct {
	SyncResp
}

// HeroSwingLvUp 主将神翼升级升星: 提升神翼等级星级
func (p *Account) HeroSwingLvUp(r servers.Request) *servers.Response {
	req := new(reqMsgHeroSwingLvUp)
	rsp := new(rspMsgHeroSwingLvUp)

	initReqRsp(
		"Attr/HeroSwingLvUpRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	ok, err_code := p.commonWCheck(int(req.HeroID))
	if !ok {
		return rpcError(rsp, err_code)
	}
	hero := p.Profile.GetHero()
	if req.Type == Wing_LvUp {
		curStarLvl := hero.HeroSwings[int(req.HeroID)].StarLv
		if uint32(curStarLvl) < gamedata.GetHeroCommonConfig().GetHWLevelUnlockStar() {
			return rpcWarn(rsp, errCode.ClickTooQuickly)
		}
		curLvl := hero.HeroSwings[int(req.HeroID)].Lv
		info := gamedata.GetHeroSwingLvUpInfo(curLvl + 1)
		costData := gamedata.CostData{}
		costData.AddItem(info.GetHWLevelupMaterial(), info.GetHWLevelupMaterialCount())
		if !account.CostBySync(p.Account, &costData, rsp, "LevelUp Hero Wing") {
			return rpcWarn(rsp, errCode.ClickTooQuickly)
		}
		hero.HeroSwings[int(req.HeroID)].Lv += 1

		//p.updateCondition(account.COND_TYP_Swing_Lvl_Together, 0, 0, "", "", rsp)
		logiclog.LogHeroWingCostProp(p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId, 0,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			costData.Items, costData.Count, int(req.HeroID), "")

		logiclog.LogHeroWingLvUp(p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId, 0,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			int(hero.HeroSwings[int(req.HeroID)].Lv), int(req.HeroID), "")
	} else if req.Type == Wing_StarUp {
		curStarLvl := hero.HeroSwings[int(req.HeroID)].StarLv
		info := gamedata.GetHeroSwingStarLvUpInfo(curStarLvl + 1)
		costData := gamedata.CostData{}
		for _, item := range info.GetHWStarup_Template() {
			costData.AddItem(item.GetHWStarupMaterial(), item.GetHWStarupMaterialCount())
		}
		if !account.CostBySync(p.Account, &costData, rsp, "StarLevelUp Hero Wing") {
			return rpcWarn(rsp, errCode.ClickTooQuickly)
		}
		hero.HeroSwings[int(req.HeroID)].StarLv += 1
		ret := hero.HeroSwings[req.HeroID].UpdateAct()
		if ret != -1 {
			logiclog.LogHeroWingAct(p.AccountID.String(),
				p.Profile.GetCurrAvatar(),
				p.Profile.GetCorp().GetLvlInfo(),
				p.Profile.ChannelId, 0,
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
				int(ret), int(req.HeroID), "")
		}

		//p.updateCondition(account.COND_TYP_Swing_Star_Together, 0, 0, "", "", rsp)
		logiclog.LogHeroWingCostProp(p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId, 0,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			costData.Items, costData.Count, int(req.HeroID), "")
		logiclog.LogHeroWingStarLvUp(p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId, 0,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			int(hero.HeroSwings[int(req.HeroID)].StarLv), int(req.HeroID), "")
	}
	// logic imp end
	p.Profile.GetData().SetNeedCheckMaxGS()
	//更新神翼信息
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByWingStar.Add(&info)
	rsp.OnChangeUpdateHeroWing(int(req.HeroID))
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// HeroSwingRest : 神翼退化到初始等级和星级
// 重置神翼

// reqMsgHeroSwingRest 神翼退化到初始等级和星级请求消息定义
type reqMsgHeroSwingRest struct {
	Req
	HeroID int64 `codec:"hero_id"` // 所操作的主将ID
}

// rspMsgHeroSwingRest 神翼退化到初始等级和星级回复消息定义
type rspMsgHeroSwingRest struct {
	SyncRespWithRewards
}

// HeroSwingRest 神翼退化到初始等级和星级: 重置神翼
func (p *Account) HeroSwingRest(r servers.Request) *servers.Response {
	req := new(reqMsgHeroSwingRest)
	rsp := new(rspMsgHeroSwingRest)

	initReqRsp(
		"Attr/HeroSwingRestRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	ok, errCode := p.commonWCheck(int(req.HeroID))
	if !ok {
		return rpcError(rsp, errCode)
	}
	hero := p.Profile.GetHero()
	wingLv := hero.HeroSwings[req.HeroID].Lv
	wingStarLv := hero.HeroSwings[req.HeroID].StarLv
	cost, ok := gamedata.GetHeroSwingResetCost(wingLv, wingStarLv)
	if !ok {
		return rpcError(rsp, Err_SwingLv)
	}
	costData := gamedata.CostData{}
	costData.AddItem(helper.VI_Hc, uint32(cost))
	if !account.CostBySync(p.Account, &costData, rsp, "Reset Hero Wing") {
		return rpcError(rsp, Err_Cost)
	}
	giveData, ok := gamedata.GetHeroSwingResetReward(wingLv, wingStarLv)
	if !ok {
		return rpcError(rsp, Err_SwingLv)
	}
	logs.Debug("Before merge and sort: %v", *giveData)
	if !account.GiveBySync(p.Account, giveData, rsp, "Reset Hero Wing rewards") {
		return rpcError(rsp, Err_Give)
	}
	logs.Debug("After merge and sort: %v", rsp.RewardID)

	hero.HeroSwings[req.HeroID].Reset()
	p.Profile.GetData().SetNeedCheckMaxGS()
	rsp.OnChangeUpdateHeroWing(int(req.HeroID))
	logiclog.LogHeroWingReset(p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, 0,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		int(req.HeroID), giveData.Items, giveData.Count, "")
	// logic imp end
	//更新神翼信息
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByWingStar.Add(&info)
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// HeroSwingShow : 控制主将神翼是否显示
// 神翼展示

// reqMsgHeroSwingShow 控制主将神翼是否显示请求消息定义
type reqMsgHeroSwingShow struct {
	Req
	IsShow bool `codec:"is_show"` // 是否显示神翼(注意工具不支持自动生成bool,手动修改,xixi)
}

// rspMsgHeroSwingShow 控制主将神翼是否显示回复消息定义
type rspMsgHeroSwingShow struct {
	SyncResp
}

// HeroSwingShow 控制主将神翼是否显示: 神翼展示
func (p *Account) HeroSwingShow(r servers.Request) *servers.Response {
	req := new(reqMsgHeroSwingShow)
	rsp := new(rspMsgHeroSwingShow)

	initReqRsp(
		"Attr/HeroSwingShowRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	p.Profile.IsHideSwing = !req.IsShow
	rsp.OnChangeHeroSwing()
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) commonWCheck(avatarID int) (bool, uint32) {
	if avatarID < 0 || avatarID >= account.AVATAR_NUM_MAX {
		return false, Err_HeroID
	}
	if p.Profile.GetHero().HeroLevel[avatarID] < gamedata.GetHeroCommonConfig().GetHWUnlockLevel() {
		return false, Err_HeroLv
	}
	if p.Profile.GetHero().HeroStarLevel[avatarID] < gamedata.GetHeroCommonConfig().GetHWUnlockStar() {
		return false, Err_HeroStarLv
	}
	return true, 0
}
