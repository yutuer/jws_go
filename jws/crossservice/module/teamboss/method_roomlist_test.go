package teamboss

import (
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/helper"

	"github.com/stretchr/testify/assert"
)

func TestMethodRoomList_Do(t *testing.T) {
	tb := module.LoadModulesList[0].NewModule(2)
	tb.Start()
	defer tb.Stop()

	c := tb.GetMethod("createroom")
	r := tb.GetMethod("roomlist")

	// 新建房间
	param := genCreateRoomParam(110, "Wings", "ACIDTest002")
	trans := new(module.Transaction)
	c.Do(*trans, param)

	// 正常获取
	param2 := new(ParamRoomList)

	// 没有符合当前等级的房间
	info2 := helper.RoomListInfo{}
	info2.RoomLevel = 99
	param2.Info = info2

	errCode, ret := r.Do(*trans, param2)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetRoomList).Info.Code)
	assert.Empty(t, ret.(RetRoomList).Info.List)

	// 存在符合等级的房间
	param2.Info.RoomLevel = 110
	errCode, ret = r.Do(*trans, param2)

	t.Logf("ret: ", ret.(RetRoomList).Info)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetRoomList).Info.Code)
	assert.Equal(t, 1, len(ret.(RetRoomList).Info.List))

	room := ret.(RetRoomList).Info.List[0]

	assert.Equal(t, uint32(110), room.Level)
	assert.Equal(t, "Wings", room.LeadName)
	assert.EqualValues(t, testGroupId1, room.ServerID)

	// 超过gamedata设置的最大房间数量
	param3 := genCreateRoomParam(110, "VG", "ACIDTest003")
	c.Do(*trans, param3)
	errCode, ret = r.Do(*trans, param2)

	// 确保目前存在2个房间
	assert.Equal(t, 2, len(ret.(RetRoomList).Info.List))

	// 修改房间最大值为1
	preAmount := gamedata.BoxCfg.RoomMax
	gamedata.BoxCfg.RoomMax = 1

	// 验证
	c.Do(*trans, param3)
	errCode, ret = r.Do(*trans, param2)
	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetRoomList).Info.Code)
	assert.Equal(t, 1, len(ret.(RetRoomList).Info.List))

	// 房间最大值改回去
	gamedata.BoxCfg.RoomMax = preAmount

	// 异常处理
}
