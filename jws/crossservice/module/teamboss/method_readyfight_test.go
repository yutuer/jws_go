package teamboss

import (
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"

	"github.com/stretchr/testify/assert"
)

func TestMethodReadyFight_Do(t *testing.T) {
	// 构建环境
	tb := module.LoadModulesList[0].NewModule(7)
	tb.Start()
	defer tb.Stop()

	create := tb.GetMethod("createroom")
	join := tb.GetMethod("joinroom")
	ready := tb.GetMethod("readyfight")

	acidLeader := "AcidWarcraft"
	acidMember := "AcidOverwatch"

	trans := new(module.Transaction)
	paramCreate := genCreateRoomParam(66, "Warcraft", acidLeader)
	create.Do(*trans, paramCreate)

	room := tb.(*TeamBoss).Room.Rooms[66][0]

	paramJoin := genJoinRoomParam(room.ID, 77, "Overwatch", acidMember)
	join.Do(*trans, paramJoin)

	// 构建参数
	paramReady := new(ParamReadyFight)
	paramReady.Info.Status = helper.TBPlayerStateReady

	// 错误的房间号
	paramReady.Info.RoomID = ""
	paramReady.Info.AcID = acidLeader
	errCode, ret := ready.Do(*trans, paramReady)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomNotExist, ret.(RetReadyFight).Info.Code)

	paramReady.Info.RoomID = room.ID

	// 错误的ACID
	paramReady.Info.AcID = ""
	errCode, ret = ready.Do(*trans, paramReady)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodePlayerNotInRoom, ret.(RetReadyFight).Info.Code)

	paramReady.Info.AcID = acidLeader

	// Leader不能Ready
	errCode, ret = ready.Do(*trans, paramReady)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeReadyFailed, ret.(RetReadyFight).Info.Code)

	// Member可以自由的Ready
	paramReady.Info.AcID = acidMember
	errCode, ret = ready.Do(*trans, paramReady)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetReadyFight).Info.Code)
	assert.Equal(t, helper.TBPlayerStateReady, room.GetPlayer(acidMember).SimpleInfo.Status)

	// member也可以自由的不Ready
	paramReady.Info.Status = helper.TBPlayerStateIdle
	errCode, ret = ready.Do(*trans, paramReady)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetReadyFight).Info.Code)
	assert.Equal(t, helper.TBPlayerStateIdle, room.GetPlayer(acidMember).SimpleInfo.Status)
}
