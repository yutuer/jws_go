package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// FestivalShop : 节日商店兑换
// 节日商店兑换宝箱

// reqMsgFestivalShop 节日商店兑换请求消息定义
type reqMsgFestivalShop struct {
	Req
	FestivalId int64 `codec:"fes_id"` // 节日Id
	GoodsId    int64 `codec:"g_id"`   // 所兑换的物品id
}

// rspMsgFestivalShop 节日商店兑换回复消息定义
type rspMsgFestivalShop struct {
	SyncRespWithRewards
}

// FestivalShop 节日商店兑换: 节日商店兑换宝箱
func (p *Account) FestivalShop(r servers.Request) *servers.Response {
	req := new(reqMsgFestivalShop)
	rsp := new(rspMsgFestivalShop)

	initReqRsp(
		"Attr/FestivalShopRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		CODE_Cost_Err
		COED_MAX_LIMIT
		CODE_Active_Closed
	)

	// 活动时间检查
	now_t := p.Profile.GetProfileNowTime()

	var actInfo *gamedata.HotActivityInfo
	_actInfo := gamedata.GetHotDatas().Activity.GetActivityInfo(int(req.FestivalId), p.Profile.ChannelQuickId)
	for _, v := range _actInfo {
		if now_t > v.StartTime && now_t < v.EndTime {
			actInfo = v
		}
	}
	if actInfo == nil {
		return rpcWarn(rsp, errCode.ActivityTimeOut)
	}

	e := p.Profile.GetFestivalBossInfo()

	e.UpdateFestivalActId(actInfo.ActivityId)
	if len(e.FbShopRewardTime) == 0 {
		e.FbShopRewardTime = make([]int64, gamedata.GetFestivalShopGoodsCount(uint32(req.FestivalId)))
	}

	rewardTime := e.GetFbShopRewardTime()
	maxLimit := gamedata.GetFestivalShopMaxRewardTime(uint32(req.FestivalId), uint32(req.GoodsId))

	if rewardTime[req.GoodsId-1] > int64(maxLimit) {
		return rpcErrorWithMsg(rsp, COED_MAX_LIMIT, "COED_MAX_LIMIT")
	}

	rewardTmp, rewardCount, needitem, needcount := gamedata.GetFestivalShopReward(uint32(req.FestivalId), uint32(req.GoodsId))

	data := &gamedata.CostData{}
	for idx, rid := range needitem {
		c := needcount[idx]
		data.AddItem(rid, c)
	}

	if !account.CostBySync(p.Account, data, rsp, "FestivalShop Cost") {
		return rpcErrorWithMsg(rsp, CODE_Cost_Err, "CODE_Cost_Er")
	}

	givesAll := gamedata.NewPriceDatas(32)
	for idx, tid := range rewardTmp {
		tc := rewardCount[idx]
		for i := 0; i < int(tc); i++ {
			gives, err := p.GetGivesByTemplate(tid)
			if err != nil {
				logs.Error("OnPlayerMsgGVEStart Loot err by %v",
					gives)
				continue
			}
			givesAll.AddOther(&gives)
		}
	}

	fbGives := gamedata.PriceDatas{}
	for idx, rid := range givesAll.Item2Client {
		c := givesAll.Count2Client[idx]
		fbGives.AddItem(rid, c)
	}

	if !account.GiveBySync(p.Account, fbGives.Gives(), rsp, "FESTIVAL Shop Rewards") {
		logs.Error("FESTIVAL Shop GiveBySync Err")
	}

	e.UpdateFbShopRewardTime(uint32(req.FestivalId), req.GoodsId-1)
	rsp.onChangeFestivalBossInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
