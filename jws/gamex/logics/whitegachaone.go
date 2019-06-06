package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// WhiteGachaone : 白盒宝箱抽取一次
// 白盒宝箱抽取一次

// reqMsgWhiteGachaone 白盒宝箱抽取一次请求消息定义
type reqMsgWhiteGachaone struct {
	Req
	GachaId int64 `codec:"gachaid"` // 消耗方式
}

// rspMsgWhiteGachaone 白盒宝箱抽取一次回复消息定义
type rspMsgWhiteGachaone struct {
	SyncRespWithRewards
	GoodsId         string `codec:"goodsid"`    // 物品ID
	GoodsCount      int64  `codec:"goodscount"` // 物品数量
	GiveGoodsId     string `codec:"givegoodsid"`
	GiveGoodesCount int64  `codec:"givegoodscount"`
}

const (
	FreeWhiteGacha = 0
	KeyWhiteGacha  = 1
	HcWhiteGacha   = 2
)

// WhiteGachaone 白盒宝箱抽取一次: 白盒宝箱抽取一次
func (p *Account) WhiteGachaone(r servers.Request) *servers.Response {
	req := new(reqMsgWhiteGachaone)
	rsp := new(rspMsgWhiteGachaone)

	initReqRsp(
		"Attr/WhiteGachaoneRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_                 = iota
		CODE_Bag_Full_Err // 失败：包裹满
		CODE_Key_Cost_Err // key
		CODE_Hc_Cost_Err
		CODE_Condition_Err
		CODE_wish_Err
	)

	// 活动时间检查
	now_t := p.Profile.GetProfileNowTime()
	ga := gamedata.GetHotDatas().Activity
	wg := p.Profile.GetWhiteGachaInfo()
	var actInfo *gamedata.HotActivityInfo
	_actInfo := gamedata.GetHotDatas().Activity.GetActivityInfo(gamedata.ActWhiteGacha, p.Profile.ChannelQuickId)

	for _, v := range _actInfo {
		if now_t > v.StartTime && now_t < v.EndTime {
			actInfo = v
		}
	}
	if actInfo == nil {
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}
	// 跟新活动Id
	wg.UpdateWhiteGachaActId(actInfo.ActivityId)

	gachaSeting := gamedata.GetHotDatas().Activity.GetActivityGachaSeting(actInfo.ActivityId)

	// 检查解锁条件
	if !account.CondCheck(gamedata.Mod_WhiteGacha, p.Account) {
		return rpcError(rsp, CODE_Condition_Err)
	}
	// 检查祝福值
	if wg.GachaBless >= int64(gachaSeting.GetWishMax()) {
		return rpcError(rsp, CODE_wish_Err)
	}
	// 检查装备物品数量
	if p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
		return rpcError(rsp, CODE_Bag_Full_Err)
	}

	if req.GachaId == FreeWhiteGacha && wg.IsCanFree(actInfo.ActivityId, p.GetProfileNowTime(), req.GachaId) {
		wg.SetUseFreeNow(p.GetProfileNowTime())
	} else if req.GachaId == KeyWhiteGacha {
		data := &gamedata.CostData{}
		data.AddItem(gachaSeting.GetGachaCoin2(), gachaSeting.GetAPrice2())
		if !account.CostBySync(p.Account, data, rsp, "WhiteGacha Key") {
			return rpcErrorWithMsg(rsp, CODE_Key_Cost_Err, "CODE_Key_Cost_Err")
		}
	} else {
		data := &gamedata.CostData{}
		data.AddItem(gachaSeting.GetGachaCoin1(), gachaSeting.GetAPrice1())
		if !account.CostBySync(p.Account, data, rsp, "WhiteGacha HC") {
			return rpcErrorWithMsg(rsp, CODE_Hc_Cost_Err, "CODE_Hc_Cost_Err")
		}
	}

	gachaNum := wg.GachaNum

	data := &gamedata.CostData{}

	lowestId, lowestCount, isLowest := ga.IsWhiteGachaLowest(actInfo.ActivityId, uint32(gachaNum))
	specilId, IsSpecil := ga.IsWhiteGachaSpecial(actInfo.ActivityId, uint32(gachaNum))
	if isLowest {
		data.AddItem(lowestId, lowestCount)
		if !account.GiveBySync(p.Account, data, rsp, "WhiteGachaOnelowestId") {
			logs.Error("WhiteGachaOne GiveBySync Err")
		}
		rsp.GoodsId = lowestId
		rsp.GoodsCount = int64(lowestCount)
	} else if IsSpecil {
		cfg := ga.GetActivityGachaNormal(specilId).RandomConfig()
		data.AddItem(cfg.GetItemID(), cfg.GetItemCount())
		if !account.GiveBySync(p.Account, data, rsp, "WhiteGachaOnespecilId") {
			logs.Error("WhiteGachaOne GiveBySync Err")
		}
		rsp.GoodsId = cfg.GetItemID()
		rsp.GoodsCount = int64(cfg.GetItemCount())
		if cfg.GetItemID() == gachaSeting.GetFinalRewardID() && cfg.GetItemCount() == gachaSeting.GetFinalRewardNum() {
			sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_White_Gacha).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
				AddParam(sysnotice.ParamType_ItemId, cfg.GetItemID()).
				AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", cfg.GetItemCount())).Send()
		}
	} else {
		cfg := ga.GetActivityGachaNormal(gachaSeting.GetItemGroupID()).RandomConfig()
		data.AddItem(cfg.GetItemID(), cfg.GetItemCount())
		if !account.GiveBySync(p.Account, data, rsp, "WhiteGachaOne") {
			logs.Error("WhiteGachaOne GiveBySync Err")
		}
		rsp.GoodsId = cfg.GetItemID()
		rsp.GoodsCount = int64(cfg.GetItemCount())
	}
	if !uutil.IsOverseaVer() {
		rsp.GiveGoodsId = gachaSeting.GetGachaShowItem()
		rsp.GiveGoodesCount = int64(gachaSeting.GetGachaShowCount())
		giveData := &gamedata.CostData{}
		giveData.AddItem(rsp.GiveGoodsId, uint32(rsp.GiveGoodesCount))
		if !account.GiveBySync(p.Account, giveData, rsp, "WhiteGachaOneGive") {
			logs.Error("WhiteGachaOne GiveBySync Err")
		}
	}
	wg.UpdateWhiteGacha()

	rsp.OnChangeWhiteGacha()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
