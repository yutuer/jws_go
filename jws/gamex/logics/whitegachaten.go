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

// WhiteGachaTen : 白盒宝箱抽取十次
// 白盒宝箱抽取十次

// reqMsgWhiteGachaTen 白盒宝箱抽取十次请求消息定义
type reqMsgWhiteGachaTen struct {
	Req
	GachaId int64 `codec:"gachaid"` // 消耗方式
}

// rspMsgWhiteGachaTen 白盒宝箱抽取十次回复消息定义
type rspMsgWhiteGachaTen struct {
	SyncRespWithRewards
	GoodsId         []string `codec:"goodsid"`    // 物品ID
	GoodsCount      []int64  `codec:"goodscount"` // 物品数量
	GiveGoodsId     string   `codec:"givegoodsid"`
	GiveGoodesCount int64    `codec:"givegoodscount"`
}

const MaxGacha = 10

// WhiteGachaTen 白盒宝箱抽取十次: 白盒宝箱抽取十次
func (p *Account) WhiteGachaTen(r servers.Request) *servers.Response {
	req := new(reqMsgWhiteGachaTen)
	rsp := new(rspMsgWhiteGachaTen)

	initReqRsp(
		"Attr/WhiteGachaTenRsp",
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

	if req.GachaId == KeyWhiteGacha {
		data := &gamedata.CostData{}
		data.AddItem(gachaSeting.GetGachaCoin2(), gachaSeting.GetTenPrice2())
		if !account.CostBySync(p.Account, data, rsp, "WhiteGacha Key") {
			return rpcErrorWithMsg(rsp, CODE_Key_Cost_Err, "CODE_Key_Cost_Err")
		}
	} else {
		data := &gamedata.CostData{}
		data.AddItem(gachaSeting.GetGachaCoin1(), gachaSeting.GetTenPrice1())
		if !account.CostBySync(p.Account, data, rsp, "WhiteGacha HC") {
			return rpcErrorWithMsg(rsp, CODE_Hc_Cost_Err, "CODE_Hc_Cost_Err")
		}
	}

	data := &gamedata.CostData{}
	for i := 0; i < MaxGacha; i++ {
		gachaNum := wg.GachaNum

		lowestId, lowestCount, isLowest := ga.IsWhiteGachaLowest(actInfo.ActivityId, uint32(gachaNum))
		specilId, IsSpecil := ga.IsWhiteGachaSpecial(actInfo.ActivityId, uint32(gachaNum))
		if isLowest {
			data.AddItem(lowestId, lowestCount)

			rsp.GoodsId = append(rsp.GoodsId, lowestId)
			rsp.GoodsCount = append(rsp.GoodsCount, int64(lowestCount))
		} else if IsSpecil {
			cfg := ga.GetActivityGachaNormal(specilId).RandomConfig()
			data.AddItem(cfg.GetItemID(), cfg.GetItemCount())

			rsp.GoodsId = append(rsp.GoodsId, cfg.GetItemID())
			rsp.GoodsCount = append(rsp.GoodsCount, int64(cfg.GetItemCount()))
			if cfg.GetItemID() == gachaSeting.GetFinalRewardID() && cfg.GetItemCount() == gachaSeting.GetFinalRewardNum() {
				sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_White_Gacha).
					AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
					AddParam(sysnotice.ParamType_ItemId, cfg.GetItemID()).
					AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", cfg.GetItemCount())).Send()
			}
		} else {
			cfg := ga.GetActivityGachaNormal(gachaSeting.GetItemGroupID()).RandomConfig()
			data.AddItem(cfg.GetItemID(), cfg.GetItemCount())

			rsp.GoodsId = append(rsp.GoodsId, cfg.GetItemID())
			rsp.GoodsCount = append(rsp.GoodsCount, int64(cfg.GetItemCount()))
		}
		wg.UpdateWhiteGacha()
	}
	if !account.GiveBySync(p.Account, data, rsp, "WhiteGachaTen") {
		logs.Error("WhiteGachaTen GiveBySync Err")
	}
	if !uutil.IsOverseaVer() {
		rsp.GiveGoodsId = gachaSeting.GetGachaShowItem()
		rsp.GiveGoodesCount = int64(gachaSeting.GetGachaShowCount()) * 10
		giveData := &gamedata.CostData{}
		giveData.AddItem(rsp.GiveGoodsId, uint32(rsp.GiveGoodesCount))
		if !account.GiveBySync(p.Account, giveData, rsp, "WhiteGachaTenGive") {
			logs.Error("WhiteGachaOne GiveBySync Err")
		}
	}
	rsp.OnChangeWhiteGacha()

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
