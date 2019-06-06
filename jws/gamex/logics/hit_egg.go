package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 砸蛋
func (p *Account) hitEgg(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Index int `codec:"idx"` // 从0开始
	}{}
	resp := &struct {
		SyncRespWithRewards
		IsSpec bool `codec:"is_spec"`
	}{}

	initReqRsp(
		"PlayerAttr/HitEggRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
		Err_Cost
		Err_Loot
		Err_give
	)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_Value_HitEeg) {
		return rpcWarn(resp, errCode.ActivityTimeOut)
	}
	// 活动时间检查
	now_t := p.Profile.GetProfileNowTime()
	var actInfo *gamedata.HotActivityInfo
	_actInfo := gamedata.GetHotDatas().Activity.GetActivityInfo(gamedata.ActHitEgg, p.Profile.ChannelQuickId)
	for _, v := range _actInfo {
		if now_t > v.StartTime && now_t < v.EndTime {
			actInfo = v
		}
	}
	if actInfo == nil {
		p.Profile.GetHitEgg().EndHitEggActivity(p.Account, now_t)
		return rpcWarn(resp, errCode.ActivityTimeOut)
	}
	p.Profile.GetHitEgg().UpdateHitEggActivityTime(p.Account,
		actInfo.StartTime, actInfo.EndTime, now_t)
	hitEgg := p.Profile.GetHitEgg()
	if hitEgg.IsEnd {
		return rpcWarn(resp, errCode.ActivityTimeOut)
	}
	hitEgg.UpdateHitEgg(now_t)

	// 参数检查
	if req.Index >= len(hitEgg.EggsShow) || !hitEgg.EggsShow[req.Index] {
		logs.Warn("hitEgg Err_Param %d ", req.Index)
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	// 消耗是否够
	costCfg := gamedata.GetHitEggCost(hitEgg.CurIdx)
	dataC := &gamedata.CostData{}
	dataC.AddItem(costCfg.GetCoinItem1ID(), costCfg.GetCurrentPrice1())
	dataC.AddItem(costCfg.GetCoinItem2ID(), costCfg.GetCurrentPrice2())
	cost := &account.CostGroup{}
	if !cost.AddCostData(p.Account, dataC) || !cost.CostBySync(p.Account, resp, "HitEgg") {
		return rpcErrorWithMsg(resp, Err_Cost, "Err_Cost")
	}

	// 砸蛋
	hitEgg.EggsShow[req.Index] = false
	isSpec, lootId, weight := hitEgg.RandHitEgg(p.GetRand())
	if isSpec {
		resp.IsSpec = true
		// 跑马灯
		sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_HITEGG_1).
			AddParam(sysnotice.ParamType_RollName, p.Profile.Name).Send()
	}
	itemId, count := p._trialLootGroup(lootId)
	if itemId == "" || count <= 0 {
		return rpcErrorWithMsg(resp, Err_Loot, "Err_Loot")
	}
	// give
	oldhc := p.Profile.GetHC().GetHC()
	dataG := &gamedata.CostData{}
	dataG.AddItem(itemId, count)
	give := &account.GiveGroup{}
	give.AddCostData(dataG)
	if !give.GiveBySyncAuto(p.Account, resp, "HitEgg") {
		return rpcErrorWithMsg(resp, Err_give, "Err_give")
	}
	chgHc := p.Profile.GetHC().GetHC() - oldhc
	hitEgg.TodayGotHc += chgHc

	resp.OnChangeHitEgg()
	resp.mkInfo(p)

	// logiclog
	logiclog.LogHitEgg(p.AccountID.String(), p.Profile.CurrAvatar,
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		req.Index, isSpec, lootId, weight,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}
