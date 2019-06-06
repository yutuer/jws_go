package teamboss

import (
	"fmt"

	ts "vcs.taiyouxi.net/jws/crossservice/module/teamboss"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const DefaultRoomID = ts.DefaultRoomID

func init() {
	crossservice.RegGroupHandle(GetGroupIDbyShardID)
}

func StartFight(sid uint, acid string, player *helper.StartFightInfo) (*helper.StartFightRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := player.RoomID
	param := ts.ParamStartFight{
		Sid:  uint32(sid),
		Acid: acid,
		Info: helper.StartFightInfo(*player),
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodStartFightID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss StartFight CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss StartFight CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetStartFight)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss StartFight CrossService CallSync Return Un-match")
	}
	logs.Trace("TeamBoss StartFight get ret: %v", &ret.Info)
	return &ret.Info, crossservice.ErrOK, nil
}

func CreateRoom(sid uint, acid string, info *helper.CreateRoomInfo) (*helper.JoinRoomRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := helper.RoomDefaultSource
	param := ts.ParamCreateRoom{
		Sid:  sid,
		Acid: acid,
		Info: helper.CreateRoomInfo(*info),
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodCreateRoomID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss CreateRoom CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss CreateRoom CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetCreateRoom)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss CreateRoom CrossService CallSync Return Un-match")
	}
	logs.Trace("TeamBoss CreateRoom get ret: %v", &ret.Info)
	return &ret.Info, crossservice.ErrOK, nil
}

func JoinRoom(sid uint, acid string, info *helper.JoinRoomInfo) (*helper.JoinRoomRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := info.RoomID
	param := ts.ParamJoinRoom{
		Sid:  sid,
		Acid: acid,
		Info: helper.JoinRoomInfo(*info),
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodJoinRoomID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss JoinRoom CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss JoinRoom CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetJoinRoom)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss JoinRoom CrossService CallSync Return Un-match")
	}
	logs.Trace("TeamBoss JoinRoom get ret: %v", &ret.Info)

	return &ret.Info, crossservice.ErrOK, nil
}

func ReadyFight(sid uint, acid string, info *helper.ReadyFightInfo) (*helper.ReadyFightRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := info.RoomID
	param := ts.ParamReadyFight{
		Sid:  sid,
		Acid: acid,
		Info: *info,
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodReadyFightID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss GetRoomList CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss GetRoomList CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetReadyFight)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss GetRoomList CrossService CallSync Return Un-match")
	}
	logs.Trace("TeamBoss ReadyFight get ret: %v", &ret.Info)

	return &ret.Info, crossservice.ErrOK, nil
}

func GetRoomList(sid uint, acid string, info *helper.RoomListInfo) (*helper.RoomListRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := helper.RoomDefaultSource
	param := ts.ParamRoomList{
		Sid:  sid,
		Acid: acid,
		Info: helper.RoomListInfo(*info),
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodRoomListID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss GetRoomList CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss GetRoomList CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetRoomList)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss GetRoomList CrossService CallSync Return Un-match")
	}
	return &ret.Info, crossservice.ErrOK, nil
}

func LeaveRoom(sid uint, acid string, info *helper.LeaveRoomInfo) (*helper.LeaveRoomRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := info.RoomID
	param := ts.ParamLeaveRoom{
		Sid:  sid,
		Acid: acid,
		Info: *info,
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodLeaveRoomID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss LeaveRoom CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss LeaveRoom CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetLeaveRoom)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss LeaveRoom CrossService CallSync Return Un-match")
	}
	return &ret.Info, crossservice.ErrOK, nil
}

func ChangeAvatar(sid uint, acid string, info *helper.ChangeAvatarInfo) (*helper.ChangeAvatarRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := info.RoomID
	param := ts.ParamChangeAvatar{
		Sid:  sid,
		Acid: acid,
		Info: *info,
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodChangeAvatarID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss ChangeAvatar CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss ChangeAvatar CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetChangeAvatar)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss ChangeAvatar CrossService CallSync Return Un-match")
	}
	return &ret.Info, crossservice.ErrOK, nil
}

func CostAdvance(sid uint, acid string, info *helper.CostAdvanceInfo) (*helper.CostAdvanceRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := info.RoomID
	param := ts.ParamCostAdvance{
		Sid:  sid,
		Acid: acid,
		Info: *info,
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodCostAdvanceID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss CostAdvance CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss CostAdvance CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetCostAdvance)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss CostAdvance CrossService CallSync Return Un-match")
	}
	return &ret.Info, crossservice.ErrOK, nil
}

func ChangeRoomStatus(sid uint, acid string, info *helper.ChangeRoomStatusInfo) (*helper.ChangeRoomStatusRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := info.RoomID
	param := ts.ParamChangeRoomStatus{
		Sid:  sid,
		Acid: acid,
		Info: *info,
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodChangeRoomStatusID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss ChangeRoomStatus CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss ChangeRoomStatus CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetChangeRoomStatus)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss ChangeRoomStatus CrossService CallSync Return Un-match")
	}
	return &ret.Info, crossservice.ErrOK, nil
}

func EndFight(sid uint, acid string, info *helper.EndFightInfo) (*helper.EndFightRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := helper.Global2RoomID(info.GlobalRoomID)
	param := ts.ParamEndFight{
		Sid:  sid,
		Acid: acid,
		Info: *info,
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodEndFightID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss EndFight CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss EndFight CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetEndFight)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss EndFight CrossService CallSync Return Un-match")
	}
	return &ret.Info, crossservice.ErrOK, nil
}

func GetPlayerDetail(sid uint, acid string, info *helper.GetPlayerDetailInfo) (*helper.GetPlayerDetailRetInfo, int, error) {
	groupID := GetGroupIDbyShardID(sid)
	source := info.RoomID
	param := ts.ParamGetPlayerDetail{
		Sid:  sid,
		Acid: acid,
		Info: *info,
	}
	rsp, errcode, err := crossservice.GetModule(sid).CallSync(groupID, ts.ModuleID, ts.MethodGetPlayerDetailID, source, param)
	if nil != err {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss GetPlayerDetail CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	if crossservice.ErrOK != errcode {
		return nil, errcode, fmt.Errorf("TeamBoss GetPlayerDetail CrossService CallSync failed, errcode %d, %v", errcode, err)
	}
	ret, ok := rsp.(*ts.RetGetPlayerDetail)
	if false == ok {
		return nil, crossservice.ErrRemote, fmt.Errorf("TeamBoss GetPlayerDetail CrossService CallSync Return Un-match")
	}
	return &ret.Info, crossservice.ErrOK, nil
}

//GetGroupIDbyShardID ..
func GetGroupIDbyShardID(sid uint) uint32 {
	return gamedata.GetTBGroupId(uint32(sid))
}
