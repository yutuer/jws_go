package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/global_count"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Good7DayToClient struct {
	PromotionID     uint32 `codec:"pid"`
	LeftTimes       uint32 `codec:"lt"`
	ServerLeftCount uint32 `codec:"slc"`
}

func (p *Account) Account7DayBugGood(r servers.Request) *servers.Response {
	req := &struct {
		Req
		PromotionID uint32 `codec:"pid"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/Account7DayBugGoodResp", r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param
		Err_Money_Not_Enough
		Err_LimitCount_Not_Enough
		Err_Give
		Err_Time_Over
	)

	// 检查时间
	now_time := p.Profile.GetProfileNowTime()
	if now_time >= gamedata.GetAccount7DayOverTime(p.Profile.CreateTime) {
		return rpcErrorWithMsg(resp, Err_Time_Over, "Err_Time_Over")
	}

	a7d := p.Profile.GetAccount7Day()
	a7d.UpdateGoods(now_time)

	cfg := gamedata.GetAccount7DayGood(req.PromotionID)
	if cfg == nil {
		return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
	}
	// 钱是否够
	cost := &account.CostGroup{}
	if !cost.AddItem(p.Account, cfg.GetCoinItemID(), cfg.GetCurrentPrice()) {
		return rpcErrorWithMsg(resp, Err_Money_Not_Enough, "Err_Money_Not_Enough")
	}
	// 是否本身还有次数
	needSync := false
	if cfg.GetCountLimit() > 0 {
		good := a7d.Goods[req.PromotionID]
		if good.LeftTimes <= 0 {
			return rpcErrorWithMsg(resp, Err_LimitCount_Not_Enough, "Err_LimitCount_Not_Enough")
		}
		// 是否服务器限制数量, 并扣数量
		if cfg.GetServerCountLimit() > 0 {
			ret, _ := p.delAccount7DayServGoodCount(req.PromotionID)
			if !ret {
				return rpcWarn(resp, errCode.Account7DaySevGoodCountNotEnough)
			}
		}
		// 扣本身次数
		good.LeftTimes--
		a7d.Goods[req.PromotionID] = good
		needSync = true
	}
	// 扣钱
	if !cost.CostBySync(p.Account, resp, "Account7DayBugGood") {
		return rpcErrorWithMsg(resp, Err_Money_Not_Enough, "Err_Money_Not_Enough")
	}
	// 给东西
	data := &gamedata.CostData{}
	data.AddItem(cfg.GetItemID(), cfg.GetGoodsCount())
	give := &account.GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(p.Account, resp, "Account7DayBugGood") {
		return rpcErrorWithMsg(resp, Err_Give, "Err_Give")
	}
	// 加积分
	point := cfg.GetActiveValue()
	if point > 0 {
		p.Profile.GetQuest().AddAccount7DayQuestPoint(p.Account, int(point), "shop")
	}
	if needSync {
		resp.OnChangeAccount7Day()
	}
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) getAccount7DayServGoodCount() map[uint32]uint32 {
	ret := global_count.GetModule(p.AccountID.ShardId).CommandExec(global_count.GlobalCountCmd{
		CmdTyp:   global_count.GlobalCount_Cmd_GetInfo,
		CountTyp: global_count.GlobalCount_Typ_Account7DayGood,
		Gid:      p.AccountID.GameId,
		Sid:      game.Cfg.GetShardIdByMerge(p.AccountID.ShardId),
	})
	if !ret.Success {
		logs.Error("getAccount7DayServGoodCount err")
		return map[uint32]uint32{}
	}
	return ret.Counti2c
}

func (p *Account) delAccount7DayServGoodCount(id uint32) (bool, map[uint32]uint32) {
	ret := global_count.GetModule(p.AccountID.ShardId).CommandExec(global_count.GlobalCountCmd{
		CmdTyp:   global_count.GlobalCount_Cmd_DelAndGet,
		CountTyp: global_count.GlobalCount_Typ_Account7DayGood,
		Gid:      p.AccountID.GameId,
		Sid:      game.Cfg.GetShardIdByMerge(p.AccountID.ShardId),
		Key:      global_count.GlobalCountKey{IId: id},
	})
	if !ret.Success {
		return false, ret.Counti2c
	}
	return true, ret.Counti2c
}
