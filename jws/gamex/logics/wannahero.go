package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/want_gen_best"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// WannaHero : 我要名将结算协议
// 我要名将结算， Mode为1时代表，请求掷骰子，为2时代表重置，为3时代表结算
// reqMsgWannaHero 我要名将结算协议请求消息定义
type reqMsgWannaHero struct {
	Req
	Mode int64 `codec:"_p1_"` // 1-投掷, 2-重置, 3-结算
}

// rspMsgWannaHero 我要名将结算协议回复消息定义
type rspMsgWannaHero struct {
	SyncRespWithRewards
}

// WannaHero 我要名将结算协议: 我要名将结算， Mode为1时代表，请求掷骰子，为2时代表重置，为3时代表结算
func (p *Account) WannaHero(r servers.Request) *servers.Response {
	req := new(reqMsgWannaHero)
	rsp := new(rspMsgWannaHero)

	initReqRsp(
		"Attr/WannaHeroRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		Err_Param
		Err_Last_Play_Not_Finish
		Err_Times
		Err_Not_Reset
		Err_Cost
		Err_Reset_Full
		Err_Give
	)

	wg := p.Profile.GetWantGeneralInfo()
	now_t := p.Profile.GetProfileNowTime()
	wg.UpdateInfo(p.Account, now_t)

	switch req.Mode {
	case 1: // 请求掷骰子
		if !wg.IsCanCast() {
			logs.Warn("WannaHero Err_Last_Play_Not_Finish")
			return rpcWarn(rsp, errCode.ClickTooQuickly)
		}
		if wg.CanPlayCountCurr <= 0 {
			return rpcErrorWithMsg(rsp, Err_Times, "Err_Times")
		}
		wg.CanPlayCountCurr--
		wg.CastDice(p.GetRand(), now_t)
		p.updateCondition(account.COND_TYP_IWant_Hero, 1, 0, "", "", rsp)
	case 2: // 重置
		if !wg.IsCanReset() {
			logs.Warn("WannaHero Err_Not_Reset %s", p.AccountID.String())
			return rpcSuccess(rsp)
		}
		if wg.CanFreeResetCountCurr > 0 {
			wg.CanFreeResetCountCurr--
		} else {
			cfg := gamedata.GetWantGeneralAwardResetCostConfig(wg.CurrHcResetCount + 1)
			if cfg == nil {
				return rpcErrorWithMsg(rsp, Err_Reset_Full, "Err_Reset_Full")
			}
			cost := &account.CostGroup{}
			if !cost.AddHc(p.Account, int64(cfg.GetResetCost())) ||
				!cost.CostBySync(p.Account, rsp, "WantHero") {
				return rpcErrorWithMsg(rsp, Err_Cost, "Err_Cost")
			}
			wg.CurrHcResetCount++
		}
		wg.CurrResetCount++
		wg.ResetCast(p.GetRand())
	case 3: // 结算
		res := wg.AwardClear()
		awCfg := gamedata.GetWantGeneralAwardConfig(res)

		// 跑马灯展示, 摇出6个6
		if res == gamedata.WantGeneralDiceCount {
			sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_IWANTYOU).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).Send()
		}

		data := &gamedata.CostData{}
		for _, aw := range awCfg.Fixed_Loot {
			data.AddItem(aw.GetGiftID(), aw.GetGiftNumber())
			if gamedata.IsHeroPiece(aw.GetGiftID()) {
				wg.AddTodayHeroPiece(aw.GetGiftNumber())
			}
		}
		if !account.GiveBySync(p.Account, data, rsp, "WantHero") {
			return rpcErrorWithMsg(rsp, Err_Give, "Err_Give")
		}
		want_gen_best.GetModule(p.AccountID.ShardId).CheckAndReplace(
			now_t, wg.DailyAllAward,
			p.AccountID.String(), p.Profile.Name)
	default:
		return rpcErrorWithMsg(rsp, Err_Param, "Err_Param")
	}

	rsp.OnChangeWantGeneralInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
