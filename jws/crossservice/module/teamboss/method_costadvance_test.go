package teamboss

import (
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"

	"github.com/stretchr/testify/assert"
)

func TestMethodCostAdvance_Do(t *testing.T) {
	// 构建环境
	tb := module.LoadModulesList[0].NewModule(6)
	tb.Start()
	defer tb.Stop()

	create := tb.GetMethod("createroom")
	join := tb.GetMethod("joinroom")
	cost := tb.GetMethod("costadvance")
	leave := tb.GetMethod("leaveroom")

	acidLeader := "Viper"
	acidMember := "Lina"

	trans := new(module.Transaction)
	paramCreate := genCreateRoomParam(88, "Viper", acidLeader)
	create.Do(*trans, paramCreate)

	room := tb.(*TeamBoss).Room.Rooms[88][0]

	paramJoin := genJoinRoomParam(room.ID, 110, "Lina", acidMember)
	join.Do(*trans, paramJoin)

	// 构建参数
	paramCost := new(ParamCostAdvance)
	info := helper.CostAdvanceInfo{}
	info.RoomID = room.ID
	paramCost.Info = info

	// 房间不存在
	paramCost.Info.RoomID = "NotExistsRoom"
	errCode, ret := cost.Do(*trans, paramCost)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomNotExist, ret.(RetCostAdvance).Info.Code)

	paramCost.Info.RoomID = room.ID

	// 玩家不在房间内
	paramCost.Info.AcID = "NotExistACID"
	errCode, ret = cost.Do(*trans, paramCost)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodePlayerNotInRoom, ret.(RetCostAdvance).Info.Code)

	// 勾选
	paramCost.Info.AcID = acidLeader
	paramCost.Info.BoxStatus = helper.BoxStatusAdvance
	errCode, ret = cost.Do(*trans, paramCost)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetCostAdvance).Info.Code)
	assert.Equal(t, acidLeader, room.AdvanceCostID)
	assert.Equal(t, helper.TBBoxStateAdvance, room.BoxStatus)

	// 另一人尝试勾选
	paramCost.Info.AcID = acidMember
	errCode, ret = cost.Do(*trans, paramCost)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeAlreadyTickRedBox, ret.(RetCostAdvance).Info.Code)
	assert.Equal(t, acidLeader, room.AdvanceCostID)

	// 试图取消非自己勾选的
	paramCost.Info.AcID = acidMember
	paramCost.Info.BoxStatus = helper.BoxStatusLow
	errCode, ret = cost.Do(*trans, paramCost)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeAlreadyTickRedBox, ret.(RetCostAdvance).Info.Code)
	assert.Equal(t, acidLeader, room.AdvanceCostID)

	// 取消
	paramCost.Info.AcID = acidLeader
	errCode, ret = cost.Do(*trans, paramCost)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeSuccess, ret.(RetCostAdvance).Info.Code)
	assert.Equal(t, 0, room.BoxStatus)
	assert.Equal(t, "", room.AdvanceCostID)

	// 非勾选者离开
	paramCost.Info.BoxStatus = helper.BoxStatusAdvance
	cost.Do(*trans, paramCost)

	paramLeave := new(ParamLeaveRoom)
	paramLeave.Info.RoomID = room.ID
	paramLeave.Info.OptAcID = acidMember
	paramLeave.Info.TgtAcID = acidMember
	leave.Do(*trans, paramLeave)

	assert.Equal(t, acidLeader, room.AdvanceCostID)
	assert.Equal(t, helper.TBBoxStateAdvance, room.BoxStatus)

	join.Do(*trans, paramJoin)

	// 勾选者离开
	paramLeave.Info.TgtAcID = acidLeader
	paramLeave.Info.OptAcID = acidLeader
	leave.Do(*trans, paramLeave)

	assert.Equal(t, "", room.AdvanceCostID)
	assert.Equal(t, 0, room.BoxStatus)

	// 勾选者被踢 理论上和上面是一样的，安全起见还是跑一下
	paramJoin.Info.RoomID = room.ID
	paramJoin.Info.JoinInfo.AcID = acidLeader
	join.Do(*trans, paramJoin)

	paramCost.Info.AcID = acidLeader
	cost.Do(*trans, paramCost)

	assert.Equal(t, acidLeader, room.AdvanceCostID)
	assert.Equal(t, helper.TBBoxStateAdvance, room.BoxStatus)

	paramLeave.Info.OptAcID = acidMember
	paramLeave.Info.TgtAcID = acidLeader
	leave.Do(*trans, paramLeave)

	assert.Equal(t, "", room.AdvanceCostID)
	assert.Equal(t, 0, room.BoxStatus)
}
