package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type VIP struct {
	V        uint32 `json:"v"`
	RmbPoint uint32 `json:"rmb"`
	handler  events.Handler
}

func (e *VIP) SetHandler(handler events.Handler) {
	e.handler = handler
}

func (e *VIP) AddRmbPoint(a *Account, p uint32, reason string) {
	if p <= 0 {
		return
	}
	account := a.AccountID.String()

	// logiclog
	a.GetGiveCurrencyLog(account, a.Profile.GetCurrAvatar(), a.Profile.GetCorp().GetLvlInfo(),
		a.Profile.ChannelId, reason, helper.VI_HcByVIP, int64(e.RmbPoint), int64(p),
		func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")

	e.RmbPoint += p
	e.Update(a)
}

func (e *VIP) GetVIP() (uint32, uint32) {
	//e.update("") VIP初始为0级不需更新
	return e.V, e.RmbPoint
}

func (e *VIP) onVIPUp(account string) {
	logs.Trace("[%s]VIP Level Up %d", account, e.V)

	// data collect TODO
	// datacollector.VIPLevelChg(account, e.Level)

	if e.handler != nil {
		e.handler.OnVIPLvUp(e.V)
	}
}

func (e *VIP) Update(p *Account) {
	account := p.AccountID.String()
	info := gamedata.GetVIPCfg(int(e.V + 1)) //需要看下一级的
	if info == nil {
		//logs.Error("VIP Xp Max Or Data Lose In %d", e.V)
		return
	}

	old := e.V
	for e.RmbPoint >= info.RMBpoints {
		e.V += 1
		e.onVIPUp(account)

		info = gamedata.GetVIPCfg(int(e.V + 1))

		if info == nil {
			logs.Warn("Xp Max Or Data Lose In %d", e.V)
			break
		}
	}
	if e.V > old {
		p.Profile.GetTitle().OnVip(p)
		// 跑马灯
		sysNoticCfg := gamedata.VipSysNotic(e.V)
		if sysNoticCfg != nil && e.V == sysNoticCfg.GetLampValueIP1() {
			sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(sysNoticCfg.GetServerMsgID())).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).Send()
		}
	}
}
