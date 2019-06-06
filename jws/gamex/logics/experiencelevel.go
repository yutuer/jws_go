package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// ExperienceLevel : 体验关卡
// 体验关卡

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgExperienceLevel 体验关卡请求消息定义
type reqMsgExperienceLevel struct {
	Req
	LevleId string `codec:"el_id"` // 关卡ID
}

// rspMsgExperienceLevel 体验关卡回复消息定义
type rspMsgExperienceLevel struct {
	SyncResp
}

// ExperienceLevel 体验关卡: 体验关卡
func (p *Account) ExperienceLevel(r servers.Request) *servers.Response {
	req := new(reqMsgExperienceLevel)
	rsp := new(rspMsgExperienceLevel)

	initReqRsp(
		"Attr/ExperienceLevelRsp",
		r.RawBytes,
		req, rsp, p)

	el := p.Profile.GetExperienceLevelInfo()

	el.AddExperiendeLevel(req.LevleId)

	rsp.OnChangeExperienceLevel()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
