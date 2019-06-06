package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamDamageRank ..
type ParamPlayerStart struct {
	Sid          uint32
	ServerUrl    string   `codec:"server_url"`
	GlobalRoomID string   `codec:"room_id"`
	AcIDs        []string `codec:"acids"`
}

//CallbackDamageRank ..
type CallbackPlayerStart struct {
	module.BaseMethod
}

func newCallbackPlayerStart(m module.Module) *CallbackPlayerStart {
	return &CallbackPlayerStart{
		module.BaseMethod{Method: CallbackPlayerStartID, Module: m},
	}
}

//NewParam ..
func (m *CallbackPlayerStart) NewParam() module.Param {
	return &ParamPlayerStart{}
}

func (tb *TeamBoss) PlayerStart(url string, sid uint32, globalRoomID string, TgtID []string) error {
	p := &ParamPlayerStart{
		Sid:          sid,
		ServerUrl:    url,
		GlobalRoomID: globalRoomID,
		AcIDs: TgtID,
	}
	logs.Info("[TeamBoss] Player start fight push: %v", *p)
	if err := tb.Push(sid, ModuleID, CallbackPlayerStartID, p); nil != err {
		logs.Error("[TeamBoss] Callback player start, push to shard [%d] failed, %v ...param %v", p.Sid, err, p)
	}
	return nil
}
