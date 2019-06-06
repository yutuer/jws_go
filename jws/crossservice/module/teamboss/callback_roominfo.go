package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamDamageRank ..
type ParamRoomInfo struct {
	Sid   uint
	Acids []string
	Info  helper.RoomDetailInfo
}

//CallbackDamageRank ..
type CallbackRoomInfo struct {
	module.BaseMethod
}

func newCallbackRoomInfo(m module.Module) *CallbackRoomInfo {
	return &CallbackRoomInfo{
		module.BaseMethod{Method: CallbackRoomInfoID, Module: m},
	}
}

//NewParam ..
func (m *CallbackRoomInfo) NewParam() module.Param {
	return &ParamRoomInfo{}
}

func (tb *TeamBoss) RoomInfo(sid uint, info helper.RoomDetailInfo, TgtID []string) error {
	p := &ParamRoomInfo{
		Sid:   sid,
		Info:  info,
		Acids: TgtID,
	}
	logs.Info("[TeamBoss] Refresh room info push: %v", *p)
	if err := tb.Push(uint32(sid), ModuleID, CallbackRoomInfoID, p); nil != err {
		logs.Error("[TeamBoss] Callback refresh room info, push to shard [%d] failed, %v ...param %v", p.Sid, err, p)
	}
	return nil
}
