package teamboss

import (
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"

	"github.com/stretchr/testify/assert"
)

func TestMethodLeaveRoom_Do(t *testing.T) {
	// 构建环境
	tb := module.LoadModulesList[0].NewModule(5)
	tb.Start()
	defer tb.Stop()

	create := tb.GetMethod("createroom")
	join := tb.GetMethod("joinroom")
	leave := tb.GetMethod("leaveroom")

	acidLeader := "AcidMaybe"
	acidMember := "AcidRTK"

	trans := new(module.Transaction)
	paramCreate := genCreateRoomParam(100, "Maybe", acidLeader)
	create.Do(*trans, paramCreate)

	room := tb.(*TeamBoss).Room.Rooms[100][0]

	paramJoin := genJoinRoomParam(room.ID, 102, "RTK", acidMember)
	join.Do(*trans, paramJoin)

	// 构建参数
	paramLeave := new(ParamLeaveRoom)
	info := helper.LeaveRoomInfo{}
	paramLeave.Info = info

	// 不存在的房间
	paramLeave.Info.RoomID = "NotExistRoomID"
	errCode, ret := leave.Do(*trans, paramLeave)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomNotExist, ret.(RetLeaveRoom).Info.Code)

	paramLeave.Info.RoomID = room.ID

	// 玩家不在此房间
	paramLeave.Info.OptAcID = "NoBody"
	paramLeave.Info.TgtAcID = acidMember
	errCode, ret = leave.Do(*trans, paramLeave)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodePlayerNotInRoom, ret.(RetLeaveRoom).Info.Code)

	paramLeave.Info.OptAcID = acidMember

	// 没有权限踢出玩家
	paramLeave.Info.TgtAcID = acidLeader
	errCode, ret = leave.Do(*trans, paramLeave)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeOptLimitPermission, ret.(RetLeaveRoom).Info.Code)

	paramLeave.Info.TgtAcID = acidMember

	// 已经开战
	paramLeave.Info.OptAcID = acidLeader
	room.RoomState = helper.TBRoomFight
	errCode, ret = leave.Do(*trans, paramLeave)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeKickFightingRoom, ret.(RetLeaveRoom).Info.Code)

	room.RoomState = helper.TBRoomIdle

	// 队长正常离开
	paramLeave.Info.TgtAcID = acidLeader
	errCode, ret = leave.Do(*trans, paramLeave)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetLeaveRoom).Info.Code)
	assert.Equal(t, acidMember, room.LeadAcID)  // 房间状态验证：剩下那只自动变队长
	assert.Equal(t, acidLeader, ret.(RetLeaveRoom).Info.Param.TgtAcID)

	// 只有一个人时自己离开
	paramLeave.Info.OptAcID = acidMember
	paramLeave.Info.TgtAcID = acidMember
	errCode, ret = leave.Do(*trans, paramLeave)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetLeaveRoom).Info.Code)
	assert.Nil(t, tb.(*TeamBoss).Room.GetRoom(room.ID))

	// 再创建好并加入房间...
	create.Do(*trans, paramCreate)
	join.Do(*trans, paramJoin)
	room = tb.(*TeamBoss).Room.Rooms[100][0]

	// 队员正常离开
	paramLeave.Info.OptAcID = acidMember
	paramLeave.Info.TgtAcID = acidMember
	errCode, ret = leave.Do(*trans, paramLeave)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetLeaveRoom).Info.Code)
	assert.Equal(t, acidLeader, room.LeadAcID)  // 房间状态验证：队长还是队长
	assert.Equal(t, acidMember, ret.(RetLeaveRoom).Info.Param.TgtAcID)

	join.Do(*trans, paramJoin)

	// 队员正常被踢出
	paramLeave.Info.TgtAcID = acidMember
	errCode, ret = leave.Do(*trans, paramLeave)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetLeaveRoom).Info.Code)
	assert.Equal(t, acidMember, ret.(RetLeaveRoom).Info.Param.TgtAcID)
}
