package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// By Fanyang 临时代码  T4711 服务器实现收集玩家信息的机制
// 要校验领取过7日登陆，战队达到等级
/*
func check7DayGiftAndCorpLv(p *Account) bool {
	gift := p.Profile.GetGifts().Gifts[:]
	corpLv, _ := p.Profile.GetCorp().GetXpInfo()
	for _, g := range gift {
		if g.Curr_activity_id == gamedata.ActivityID7DayLogin {
			_, count := g.IsHasReward()
			// 这个判断是为了最后一次领奖后，立刻返回空，即功能关闭
			if g.Curr_gift_idx == count-1 && g.Curr_gift_stat > 0 {
				if corpLv >= gamedata.GetCommonCfg().GetTempGiftLvReq() {
					acID := p.AccountID.String()
					//logs.Warn("[%s]check7DayGiftAndCorpLv %d %v", acID, corpLv, g)

					logiclog.Error(acID, logiclog.LogType_TempGiftLv,
						struct {
							AccountID string `json:"id"`
							Name      string `json:"name"`
						}{
							AccountID: acID,
							Name:      p.Profile.Name,
						}, "TempGiftLv")

					return true
				}
			}
		}
	}

	return false
}
*/

// SkipTutorialReq : 跳过特定组新手引导
// 跳过某组引导，服务器置账号到正常结束状态

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgSkipTutorialReq 跳过特定组新手引导请求消息定义
type reqMsgSkipTutorialReq struct {
	Req
	GroupId string `codec:"groupid"` // 跳过的引导组ID
}

// rspMsgSkipTutorialReq 跳过特定组新手引导回复消息定义
type rspMsgSkipTutorialReq struct {
	SyncResp
	IsSuccess int64 `codec:"issuccess"` // 是否成功，1为成功，0为失败
}

// SkipTutorialReq 跳过特定组新手引导: 跳过某组引导，服务器置账号到正常结束状态
func (p *Account) SkipTutorialReq(r servers.Request) *servers.Response {
	req := new(reqMsgSkipTutorialReq)
	rsp := new(rspMsgSkipTutorialReq)

	initReqRsp(
		"Attr/SkipTutorialReqRsp",
		r.RawBytes,
		req, rsp, p)
	acID := p.AccountID.String()

	// logic imp begin
	if p.Tmp.TmpSkipEquipCount < 10 {
		p.Tmp.TmpSkipEquipCount++
		if req.GroupId == "EQUIP" {
			account.GiveAndThenEquip(p.Account, "BR_ALL_1_1", gamedata.PartID_Bracers)
			account.GiveAndThenEquip(p.Account, "RG_ALL_1_1", gamedata.PartID_Ring)
		} else if req.GroupId == "EQUIP3" {
			account.GiveAndThenEquip(p.Account, "RG_ALL_1_2", gamedata.PartID_Ring)
		} else {
			logs.SentryLogicCritical(acID,
				"SkipTuReq Err By %s",
				req.GroupId)
		}
		p.Profile.AppendNewHand(req.GroupId)
		rsp.IsSuccess = 1
	}

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
