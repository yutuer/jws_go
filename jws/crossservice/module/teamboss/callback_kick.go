package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamDamageRank ..
type ParamKick struct {
	Sid   uint32
	Acids []string
	Param helper.LeaveRoomParam
}

//CallbackDamageRank ..
type CallbackKick struct {
	module.BaseMethod
}

func newCallbackKick(m module.Module) *CallbackKick {
	return &CallbackKick{
		module.BaseMethod{Method: CallbackKickID, Module: m},
	}
}

//NewParam ..
func (m *CallbackKick) NewParam() module.Param {
	return &ParamKick{}
}

func (tb *TeamBoss) Kick(sid uint32, param helper.LeaveRoomParam, TgtID []string) error {
	p := &ParamKick{
		Sid:   sid,
		Param: param,
		Acids: TgtID,
	}
	logs.Info("[TeamBoss] Kick player push: %v", *p)
	if err := tb.Push(sid, ModuleID, CallbackKickID, p); nil != err {
		logs.Error("[TeamBoss] Callback kick, push to shard [%d] failed, %v ...param %v", p.Sid, err, p)
	}
	return nil
}
