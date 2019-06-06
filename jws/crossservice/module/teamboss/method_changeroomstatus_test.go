package teamboss

import (
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"

	"github.com/stretchr/testify/assert"
)

func TestMethodChangeRoomStatus_Do(t *testing.T) {
	tb := module.LoadModulesList[0].NewModule(3)
	tb.Start()
	defer tb.Stop()

	c := tb.GetMethod("createroom")
	change := tb.GetMethod("changeroomstatus")

	// 新建房间
	acid := "ACIDTest004"

	param := genCreateRoomParam(110, "EG", acid)
	trans := new(module.Transaction)
	errCode, ret := c.Do(*trans, param)

	roomId := ret.(RetCreateRoom).Info.Info.RoomID

	// 构建参数
	param2 := new(ParamChangeRoomStatus)
	param2.Info = helper.ChangeRoomStatusInfo{}
	param2.Info.RoomID = roomId
	param2.Info.AcID = acid
	param2.Info.RoomStatus = helper.TBRoomFight

	errCode, ret = change.Do(*trans, param2)

	// 成功
	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetChangeRoomStatus).Info.Code)
	assert.Equal(t, 1, ret.(RetChangeRoomStatus).Info.RoomStatus)

	// 不存在的roomID
	param2.Info.RoomID = "99_99_999999"
	errCode, ret = change.Do(*trans, param2)
	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomNotExist, ret.(RetChangeRoomStatus).Info.Code)

	param2.Info.RoomID = roomId

	// 玩家不在此房间
	param2.Info.AcID = "ACIDTest005"
	errCode, ret = change.Do(*trans, param2)
	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodePlayerNotInRoom, ret.(RetChangeRoomStatus).Info.Code)

	param2.Info.AcID = acid

	// 玩家不是队长
	room := tb.(*TeamBoss).Room.Rooms[110][0]
	room.LeadAcID = "ACIDNewLeader"
	errCode, ret = change.Do(*trans, param2)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeOptLimitPermission, ret.(RetChangeRoomStatus).Info.Code)
}
