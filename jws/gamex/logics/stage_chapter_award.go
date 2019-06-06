package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	warn "vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestChapterAward struct {
	Req
	ChapterId string `codec:"chap"`
	Index     int32  `codec:"xid"`
}

type ResponseChapterAward struct {
	SyncRespWithRewards
}

func (p *Account) ChapterAward(r servers.Request) *servers.Response {
	req := &RequestChapterAward{}
	resp := &ResponseChapterAward{}

	initReqRsp(
		"PlayLevel/ChapterAwardResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_                         = iota
		Err_Award_Index           // 失败:请求的奖励index错误，配置里没有
		Err_Chapter_Not_Found     // 失败：chapter错误，配置里没有
		Err_Award_Star_Not_Enough // 失败：想领奖的星数不足
		Err_Award_Repeat          // 失败：重复领奖
		Err_Award_Not_Found       // 失败：奖励配置里没有
		Err_Give_Fail
	)

	acid := p.AccountID.String()
	var errCode uint32
	// 奖励对应星数
	star_need := gamedata.ChapterAwardId2Star(req.ChapterId, req.Index-1)
	if star_need < 0 {
		errCode = Err_Award_Index
		logs.SentryLogicCritical(acid, "Chapter award, err [%d] %v ", errCode, req)
		return rpcError(resp, errCode)
	}

	// 找到对应章节
	chapter := p.Profile.GetStage().GetChapterInfoWithInit(req.ChapterId)
	if chapter == nil {
		errCode = Err_Chapter_Not_Found
		logs.SentryLogicCritical(acid, "Chapter award, err [%d] %v ", errCode, req)
		return rpcError(resp, errCode)
	}

	// 星数是否达到要求
	if chapter.Star < uint32(star_need) {
		errCode = Err_Award_Star_Not_Enough
		logs.SentryLogicCritical(acid, "Chapter award, err [%d] %v ", errCode, req)
		return rpcError(resp, errCode)
	}

	// 是否已经领过奖
	for _, index := range chapter.Has_awardId {
		if index == uint32(req.Index) {
			return rpcWarn(resp, warn.ClickTooQuickly)
		}
	}

	// 领奖
	cfgAward := gamedata.ChapterGoalAward(req.ChapterId, uint32(star_need))
	if cfgAward == nil {
		errCode = Err_Award_Not_Found
		logs.SentryLogicCritical(acid, "Chapter award, err [%d] %v ", errCode, req)
		return rpcError(resp, errCode)
	}

	chapter.Has_awardId = append(chapter.Has_awardId, uint32(req.Index))
	cost_data := &gamedata.CostData{}
	give := account.GiveGroup{}
	for _, aAward := range cfgAward.GetGoalAward_Template() {
		cost_data.AddItem(aAward.GetReward(), aAward.GetCount())
	}
	give.AddCostData(cost_data)
	if !give.GiveBySyncAuto(p.Account, resp, "ChapterAward") {
		return rpcErrorWithMsg(resp, Err_Give_Fail, "ChapterAward give award fail")
	}

	resp.OnChangeChapter(req.ChapterId)
	resp.mkInfo(p)
	return rpcSuccess(resp)
}
