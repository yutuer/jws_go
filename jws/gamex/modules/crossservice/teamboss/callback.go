package teamboss

import (
	"vcs.taiyouxi.net/jws/crossservice/module"
	tb "vcs.taiyouxi.net/jws/crossservice/module/teamboss"
	"vcs.taiyouxi.net/jws/gamex/models/codec"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const Prefix = "NOTIFY/Push"

type Msg struct {
	notifyAddress string
	PushMsgTyp    string `codec:"typ"`
	PushMsg       []byte `codec:"push_msg"`
}

func (pp *Msg) SetAddr(addr string) {
	pp.notifyAddress = addr
}

func (pp *Msg) GetAddr() string {
	return pp.notifyAddress
}

func init() {
	crossservice.RegCallbackHandle(tb.ModuleID, tb.CallbackPlayerStartID, CallBackPlayerStart)
	crossservice.RegCallbackHandle(tb.ModuleID, tb.CallbackRoomInfoID, CallBackRoomInfo)
	crossservice.RegCallbackHandle(tb.ModuleID, tb.CallbackKickID, CallBackKick)
}

type ParamPlayerStart tb.ParamPlayerStart

//CallBackPlayerStart ..
func CallBackPlayerStart(p module.Param) {
	param := p.(*tb.ParamPlayerStart)
	logs.Warn("[TeamBoss] Callback player start param %v", param)
	msg := Msg{
		PushMsgTyp: "tb_start_fight",
		PushMsg:    codec.Encode(param),
	}
	logs.Debug("param.GlobalRoomID: %v", param.GlobalRoomID)
	player_msg.SendToPlayers(param.AcIDs, player_msg.PlayerMsgTeamStartFight, msg)
}

func CallBackRoomInfo(p module.Param) {
	param := p.(*tb.ParamRoomInfo)
	msg := Msg{
		PushMsgTyp: "tb_refresh_room",
		PushMsg:    GenRoomPushInfo(param.Info),
	}
	logs.Warn("[TeamBoss] Call back room info param %v", param)
	player_msg.SendToPlayers(param.Acids, player_msg.PlayerMsgRefreshRoom, msg)
}

func CallBackKick(p module.Param) {
	param := p.(*tb.ParamKick)
	msg := Msg{
		PushMsgTyp: "tb_kicked",
		PushMsg:    codec.Encode(param.Param),
	}
	logs.Warn("[TeamBoss] Kick param %v", param)
	player_msg.SendToPlayers(param.Acids, player_msg.PlayerMsgTeamBossKicked, msg)
}
