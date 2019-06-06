package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestResetGameModeCD struct {
	Req
	GameModeId uint32 `codec:"gameModeId"`
}

type ResponseResetGameModeCD struct {
	SyncResp
}

func (p *Account) ResetGameModeCD(r servers.Request) *servers.Response {
	req := &RequestResetGameModeCD{}
	resp := &ResponseResetGameModeCD{}

	initReqRsp(
		"PlayLevel/ResetGameModeCDResponse",
		r.RawBytes,
		req, resp, p)

	const (
		CODE_MIN = 20
	)

	if ok, errcode := p.Profile.GameMode.ResetCD(p.Account, req.GameModeId); !ok {
		return rpcError(resp, errcode+CODE_MIN)
	}

	resp.OnChangeGameMode(req.GameModeId)
	resp.OnChangeHC()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
