package teamboss

import (
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"

	"github.com/stretchr/testify/assert"
)

func TestMethodJoinRoom_Do(t *testing.T) {
	tb := module.LoadModulesList[0].NewModule(4)
	tb.Start()
	defer tb.Stop()

	create := tb.GetMethod("createroom")
	join := tb.GetMethod("joinroom")

	acid1 := "TESTACID007"
	acid2 := "TESTACID008"

	param := genCreateRoomParam(108, "EG", acid1)
	trans := new(module.Transaction)
	_, ret := create.Do(*trans, param)

	room := tb.(*TeamBoss).Room.Rooms[108][0]

	param2 := genJoinRoomParam(room.ID, 101, "Musket", acid2)

	// 房间不存在
	param2.Info.RoomID = "222_222_2222222"
	errCode, ret := join.Do(*trans, param2)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomNotExist, ret.(RetJoinRoom).Info.Code)

	// 房间已经开战
	param2.Info.RoomID = room.ID
	room.RoomState = 1
	errCode, ret = join.Do(*trans, param2)
	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomInBattle, ret.(RetJoinRoom).Info.Code)

	room.RoomState = 0

	// 房间已锁，没有邀请
	room.RoomSetting = 1

	errCode, ret = join.Do(*trans, param2)
	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomCantEntry, ret.(RetJoinRoom).Info.Code)

	// 有邀请
	param2.Info.JoinInfo.IsInvited = true
	errCode, ret = join.Do(*trans, param2)

	info := ret.(RetJoinRoom).Info.Info

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetJoinRoom).Info.Code)
	assert.Equal(t, room.ID, info.RoomID)
	assert.Equal(t, 2, len(info.SimpleInfo)) // 两个人

	// 房间已满时邀请
	param3 := genJoinRoomParam(room.ID, 105, "Nano", "TestAcidNano")
	param3.Info.JoinInfo.IsInvited = true
	errCode, ret = join.Do(*trans, param3)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomPlayerFull, ret.(RetJoinRoom).Info.Code)

	// 新建两个param，测直接加入
	paramCreate := genCreateRoomParam(100, "Creator", "AcidCreator")

	_, ret = create.Do(*trans, paramCreate)
	roomID := ret.(RetCreateRoom).Info.Info.RoomID

	// 正常
	paramJoin1 := genJoinRoomParam(roomID, 77, "Joiner1", "AcidJoiner1")
	errCode, ret = join.Do(*trans, paramJoin1)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetJoinRoom).Info.Code)
	assert.Equal(t, room.ID, info.RoomID)
	assert.Equal(t, 2, len(info.SimpleInfo)) // 两个人

	// 已满
	paramJoin2 := genJoinRoomParam(roomID, 111, "Joiner2", "AcidJoiner2")
	errCode, ret = join.Do(*trans, paramJoin2)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomPlayerFull, ret.(RetJoinRoom).Info.Code)
}
