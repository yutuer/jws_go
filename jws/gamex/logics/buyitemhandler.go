package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
)

// BuyItem : 购买部分道具
//
const (
	buy_VI_XZ_SD       = "VI_XZ_SD"
	buy_VI_HDP_SD      = "VI_HDP_SD"
	buy_VI_WB_BUFFCOIN = "VI_WB_BUFFCOIN"
)

func (p *Account) BuyItemHandler(req *reqMsgBuyItem, resp *rspMsgBuyItem) uint32 {
	costData := &gamedata.CostData{}
	giveData := &gamedata.CostData{}
	reason := ""
	switch req.PropID {
	case buy_VI_HDP_SD:
		cfg := gamedata.GetHeroDiffConfig()
		costData.AddItem(cfg.GetBuySweepItemUse(), uint32(req.PropCount)*cfg.GetBuySweepItemPrice())
		giveData.AddItem(gamedata.VI_HDP_SD, uint32(req.PropCount))
		reason = "Buy VI_HDP_SD"
	case buy_VI_XZ_SD:
		cfg := gamedata.GetExpeditionCfg()
		costData.AddItem(cfg.GetBuySweepItemUse(), uint32(req.PropCount)*cfg.GetBuySweepItemPrice())
		giveData.AddItem(gamedata.VI_XZ_SD, uint32(req.PropCount))
		reason = "Buy VI_XZ_SD"
	case buy_VI_WB_BUFFCOIN:
		cfg := gamedata.GetWBConfig()
		costData.AddItem(cfg.GetBuySweepItemUse(), uint32(req.PropCount)*cfg.GetBuySweepItemPrice())
		giveData.AddItem(gamedata.VI_WB_BUFFCOIN, uint32(req.PropCount))
		reason = "Buy VI_WB_BUFFCOIN"
	}
	if !account.CostBySync(p.Account, costData, resp, reason) {
		return errCode.CommonLessMoney
	}
	if !account.GiveBySync(p.Account, giveData, resp, reason) {
		return errCode.CommonLessMoney
	}
	return 0
}
