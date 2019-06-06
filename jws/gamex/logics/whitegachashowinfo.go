package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// WhiteGachaShowInfo : 白盒宝箱奖励展示
// 白盒宝箱奖励展示

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgWhiteGachaShowInfo 白盒宝箱奖励展示请求消息定义
type reqMsgWhiteGachaShowInfo struct {
	Req
}

// rspMsgWhiteGachaShowInfo 白盒宝箱奖励展示回复消息定义
type rspMsgWhiteGachaShowInfo struct {
	SyncResp
	WhiteGachaItemId      []string `codec:"wg_id"`  // 非终极大奖物品Id
	WhiteGachaItemCount   []int64  `codec:"wg_ct"`  // 非终极大奖物品count
	WhiteGachaIsUnusual   []int64  `codec:"wg_iu"`  // 非终极大奖物品是否稀有
	WhiteGachaFinalReward string   `codec:"wgf_id"` // 终极大奖物品Id
	WhiteGachaFinalCount  int64    `codec:"wgf_ct"` // 终极大奖物品count
	WhiteGachaSpecilNum   []int64  `codec:"wgs_n"`  // 保底奖励抽取到第几次给奖励
	WhiteGachaSpecilId    []string `codec:"wgs_id"` // 保底奖励Id
	WhiteGachaSpecilCount []int64  `codec:"wgs_c"`  // 保底奖励数量
	WhiteGachaMaxWish     int64    `codec:"wg_mw"`  // 祝福值最大值
	WhiteGachaOneHc       int64    `codec:"wg_ohc"` // 抽一次消耗的钻石
	WhiteGachaTenHc       int64    `codec:"wg_thc"` // 抽十次消耗的钻石
	WhiteGachaOneKey      int64    `codec:"wg_ok"`  // 抽一次消耗的钥匙
	WhiteGachaTenKey      int64    `codec:"wg_tk"`  // 抽十次消耗的钥匙
}

// WhiteGachaShowInfo 白盒宝箱奖励展示: 白盒宝箱奖励展示
func (p *Account) WhiteGachaShowInfo(r servers.Request) *servers.Response {
	req := new(reqMsgWhiteGachaShowInfo)
	rsp := new(rspMsgWhiteGachaShowInfo)

	initReqRsp(
		"Attr/WhiteGachaShowInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// 活动时间检查
	now_t := p.Profile.GetProfileNowTime()
	ga := gamedata.GetHotDatas().Activity
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

	showInfo := ga.GetActivityWhiteGachaShow(actInfo.ActivityId)
	gachaSetting := ga.GetActivityGachaSeting(actInfo.ActivityId)
	for _, info := range showInfo {
		rsp.WhiteGachaItemId = append(rsp.WhiteGachaItemId, info.GetItemID())
		rsp.WhiteGachaItemCount = append(rsp.WhiteGachaItemCount, int64(info.GetItemCount()))
		rsp.WhiteGachaIsUnusual = append(rsp.WhiteGachaIsUnusual, int64(info.GetUnusual()))
	}
	rsp.WhiteGachaFinalReward = gachaSetting.GetFinalRewardID()
	rsp.WhiteGachaFinalCount = int64(gachaSetting.GetFinalRewardNum())

	lowest := ga.GetWhiteGachaLowest(actInfo.ActivityId)
	for _, info := range lowest {
		rsp.WhiteGachaSpecilNum = append(rsp.WhiteGachaSpecilNum, int64(info.GetLowestTimes()))
		rsp.WhiteGachaSpecilId = append(rsp.WhiteGachaSpecilId, info.GetItemID())
		rsp.WhiteGachaSpecilCount = append(rsp.WhiteGachaSpecilCount, int64(info.GetItemCount()))

	}
	rsp.WhiteGachaMaxWish = int64(gachaSetting.GetWishMax())
	rsp.WhiteGachaOneHc = int64(gachaSetting.GetAPrice1())
	rsp.WhiteGachaTenHc = int64(gachaSetting.GetTenPrice1())
	rsp.WhiteGachaOneKey = int64(gachaSetting.GetAPrice2())
	rsp.WhiteGachaTenKey = int64(gachaSetting.GetTenPrice2())

	return rpcSuccess(rsp)
}
