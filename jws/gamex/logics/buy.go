package logics

import (
	//"vcs.taiyouxi.net/jws/gamex/datacollector"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	//"vcs.taiyouxi.net/platform/planx/util"

	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
可以购买的(理论上)
	VI_SC 钱
	VI_FI 精铁
	VI_HC 硬通
	VI_XP 角色经验
	VI_CorpXP 战队经验
	VI_EN 体力
	VI_SweepTicket 扫荡券
	VI_BFP Boss战精力
	GENERAL_XXX 任意副将碎片
	任意装备

可以作为购买消耗的:
	VI_SC 钱
	VI_FI 精铁
	VI_HC 硬通
	VI_EN 体力
	VI_SweepTicket 扫荡券
	VI_BFP Boss战精力
*/

type RequestBuy struct {
	Req
	Typ    int    `codec:"t"`
	Param1 string `codec:"p1"`
}

type ResponseBuy struct {
	SyncRespWithRewards
}

func (p *Account) Buy(r servers.Request) *servers.Response {
	req := &RequestBuy{}
	resp := &ResponseBuy{}

	initReqRsp(
		"PlayerAttr/BuyRsp", r.RawBytes,
		req, resp, p)

	code, warncode := p.BuyImp(req.Typ, req.Param1, resp)
	if warncode != 0 {
		return rpcWarn(resp, warncode)
	}
	if code != 0 {
		logs.Warn("buy failed type %v code: %v", req.Typ, code)
		return rpcWarn(resp, errCode.CommonConditionFalse)
	}

	// market activity
	p.Profile.GetMarketActivitys().OnBuy(p.AccountID.String(),
		req.Typ, p.Profile.GetProfileNowTime())

	resp.OnChangeBuy()
	resp.OnChangeMarketActivity()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) BuyImp(typ int, param1 string, resp interfaces.ISyncRspWithRewards) (uint32, uint32) {
	logs.Trace("[%s]BuyImp:%d",
		p.AccountID, typ)

	const (
		_                       = iota
		CODE_Typ_ERROR          //失败:
		CODE_NoCount_ERROR      //失败:次数不足
		CODE_Cost_ERROR         //失败:消耗不足
		CODE_Give_ERROR         //失败:没有足够的物品
		CODE_No_Info            //失败:装备信息缺失
		CODE_Energy_Limit_ERROR //失败:体力最大值
		CODE_Hero_Talent_ERRR   //失败:主将天赋还有，不能买
		CODE_Baozi_Limit_ERROR  //失败:包子最大值
		COED_TIME_Limit_ERROR   //失败:次数最大值
	)

	now_t := p.Profile.GetProfileNowTime()
	player_buy := p.Profile.GetBuy()

	ok, price, give, count, limit := player_buy.GetInfo(
		typ,
		p.Profile.GetVipLevel(),
		now_t,
		param1)

	if !ok {
		return CODE_Typ_ERROR, 0
	}

	if limit > 0 && count >= limit {
		return CODE_NoCount_ERROR, 0
	}

	if price == nil || give == nil {
		return CODE_No_Info, 0
	}

	if typ == helper.Buy_Typ_GuildBossCount {
		limitCount, _ := p.Profile.GetCounts().Get(gamedata.CounterTypeGuildBossBuyTime, p)
		if limitCount <= 0 {
			return CODE_NoCount_ERROR, 0
		}
	} else if typ == helper.Buy_Typ_GuildBigBossCount {
		limitCount, _ := p.Profile.GetCounts().Get(counter.CounterTypeGuildBigBossBuyTime, p)
		if limitCount <= 0 {
			return CODE_NoCount_ERROR, 0
		}
	}

	// 判断所拥有包子数是否达到极限
	if typ == helper.Buy_Typ_BaoZi {
		curCount := p.Profile.GetSC().GetSC(gamedata.SC_BaoZi)
		if curCount >= int64(gamedata.GetCommonCfg().GetBaoZiGetLimit()) {
			return CODE_Baozi_Limit_ERROR, 0
		}
	}

	if typ == helper.Buy_Typ_EnergyBuy {
		v, _, _ := p.Profile.GetEnergy().Get()
		if v >= int64(gamedata.GetCommonCfg().GetMaxEnergy()) {
			logs.Warn("BuyImp CODE_Energy_Limit_ERROR")
			return 0, errCode.ClickTooQuickly
		}
	}

	if typ == helper.Buy_Typ_HeroTalentPoint {
		tal := p.Profile.GetHeroTalent()
		tal.UpdateTalentPoint(now_t)
		if tal.TalentPoint > 0 {
			return 0, errCode.ClickTooQuickly
		}
	}

	if typ == helper.Buy_Typ_FestivalBossCount {
		maxHas, _ := gamedata.GetGameModeCfgTimes(counter.CounterTypeFestivalBoss)

		if p.Profile.GetCounts().Counts[counter.CounterTypeFestivalBoss] >= int(maxHas) {
			return COED_TIME_Limit_ERROR, 0
		}
	}

	reason := helper.BuyTypeString(typ)
	if typ == helper.Buy_Typ_EliteTimes {
		stage_data := gamedata.GetStageData(param1)
		if stage_data.Type == gamedata.LEVEL_TYPE_HELL {
			reason = "BuyHellTimes"
		}
	}
	if !account.CostBySync(p.Account, &price.Cost, resp, reason) {
		return CODE_Cost_ERROR, errCode.ClickTooQuickly
	}

	if !account.GiveBySync(p.Account, &give.Cost, resp, reason) {
		return CODE_Give_ERROR, 0
	}

	if typ == helper.Buy_Typ_GuildBossCount {
		p.Profile.GetCounts().Use(counter.CounterTypeGuildBossBuyTime, p)
		resp.OnChangeGameMode(gamedata.CounterTypeGuildBossBuyTime)
	} else if typ == helper.Buy_Typ_GuildBigBossCount {
		p.Profile.GetCounts().Use(counter.CounterTypeGuildBigBossBuyTime, p)
		resp.OnChangeGameMode(gamedata.CounterTypeGuildBigBossBuyTime)
	}

	logs.Trace("buy %v %v %d %d", price, give, count, limit)

	player_buy.OnBuy(typ, p.Profile.GetProfileNowTime(), param1)

	switch typ {
	case helper.Buy_Typ_SC:
		p.updateCondition(account.COND_TYP_BuyMoney,
			1, 0, "", "", resp)
	case helper.Buy_Typ_EnergyBuy:
		p.updateCondition(account.COND_TYP_BuyEnergy,
			1, 0, "", "", resp)
	}
	return 0, 0
}

// BuyExpItem : 购买经验药
// 用来购买经验药的协议
// reqMsgBuyExpItem 购买经验药请求消息定义
type reqMsgBuyExpItem struct {
	Req
	ItemID string `codec:"itemid"` // 物品ID
	Count  int64  `codec:"count"`  // 购买数量
}

// rspMsgBuyExpItem 购买经验药回复消息定义
type rspMsgBuyExpItem struct {
	SyncRespWithRewards
	ItemID string `codec:"itemid"` // 物品ID
	Count  int64  `codec:"count"`  // 购买数量
}

// BuyExpItem 购买经验药: 用来购买经验药的协议
func (p *Account) BuyExpItem(r servers.Request) *servers.Response {
	req := new(reqMsgBuyExpItem)
	rsp := new(rspMsgBuyExpItem)

	initReqRsp(
		"Attr/BuyExpItemRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Config
		Err_Lvl
		Err_Cost
		Err_Give
	)

	if req.Count < 0 || req.Count > uutil.CHEAT_INT_MAX {
		return rpcErrorWithMsg(rsp, 99, "BuyExpItem Count cheat")
	}

	cfg := gamedata.GetHeroLevelItem(req.ItemID)
	if cfg == nil {
		return rpcErrorWithMsg(rsp, Err_Config, "Err_Config")
	}

	if p.Profile.GetCorp().GetLvlInfo() < cfg.GetHLIPurchaseLevelLimit() {
		return rpcErrorWithMsg(rsp, Err_Lvl, "Err_Lvl")
	}

	data := &gamedata.CostData{}
	data.AddItem(cfg.GetHLIPurchaseCoin(), cfg.GetHLIPurchasePrice()*uint32(req.Count))
	if !account.CostBySync(p.Account, data, rsp, "BuyExpItem") {
		logs.Warn("BuyExpItem Err_Cost")
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	give := &account.GiveGroup{}
	give.AddItem(req.ItemID, uint32(req.Count))
	if !give.GiveBySyncAuto(p.Account, rsp, "BuyExpItem") {
		return rpcErrorWithMsg(rsp, Err_Give, "Err_Give")
	}

	rsp.ItemID = req.ItemID
	rsp.Count = req.Count

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
