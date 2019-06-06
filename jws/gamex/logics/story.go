package logics

import (
	//"vcs.taiyouxi.net/jws/gamex/models"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	//"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestGetStory struct {
	Req
	SId int `codec:"sid"`
}

type ResponseGetStory struct {
	Resp
	Story [][]byte `codec:"story"`
}

func (p *Account) GetStory(r servers.Request) *servers.Response {
	req := &RequestGetStory{}
	resp := &ResponseGetStory{}

	initReqRsp(
		"PlayerAttr/GetStoryRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_                = iota
		CODE_No_Data_Err // 失败:数据错误
	)

	res, ok := p.Profile.GetStory().GetStoryState(p.Account, req.SId)
	if !ok {
		return rpcError(resp, CODE_No_Data_Err)
	}

	resp.Story = make([][]byte, 0, len(res))
	for _, r := range res {
		resp.Story = append(resp.Story, encode(r))
	}

	return rpcSuccess(resp)
}

type RequestReadStory struct {
	Req
	StoryId []int `codec:"sid"`
}

type ResponseReadStory struct {
	Resp
}

func (p *Account) ReadStory(r servers.Request) *servers.Response {
	req := &RequestReadStory{}
	resp := &ResponseReadStory{}

	initReqRsp(
		"PlayerAttr/ReadStoryRsp",
		r.RawBytes,
		req, resp, p)

	for _, ids := range req.StoryId {
		p.Profile.GetStory().SetHasRead(ids)
	}

	return rpcSuccess(resp)
}
