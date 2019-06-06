package teamboss

import (
	"testing"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/helper"

	"github.com/stretchr/testify/assert"
)

func TestMethodStartFight_Do(t *testing.T) {
	// 构建环境
	tb := module.LoadModulesList[0].NewModule(8)
	tb.Start()
	defer tb.Stop()

	create := tb.GetMethod("createroom")
	join := tb.GetMethod("joinroom")
	start := tb.GetMethod("startfight")
	ready := tb.GetMethod("readyfight")

	acidLeader := "AcidWarcraft"
	acidMember := "AcidOverwatch"

	trans := new(module.Transaction)
	paramCreate := genCreateRoomParam(55, "Diablo", acidLeader)
	create.Do(*trans, paramCreate)

	room := tb.(*TeamBoss).Room.Rooms[55][0]

	// 构建Param
	paramStart := new(ParamStartFight)
	paramStart.Info.AcID = acidLeader
	paramStart.Info.RoomID = room.ID

	// 房间ID错误
	paramStart.Info.RoomID = ""
	errCode, ret := start.Do(*trans, paramStart)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeRoomNotExist, ret.(RetStartFight).Info.Code)

	paramStart.Info.RoomID = room.ID

	// ACID错误
	paramStart.Info.AcID = ""
	errCode, ret = start.Do(*trans, paramStart)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeOptInvalid, ret.(RetStartFight).Info.Code)

	paramStart.Info.AcID = acidLeader

	// 人不满
	errCode, ret = start.Do(*trans, paramStart)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeOptInvalid, ret.(RetStartFight).Info.Code)

	// 加入Member
	paramJoin := genJoinRoomParam(room.ID, 97, "Hearthstone", acidMember)
	join.Do(*trans, paramJoin)

	// 非Leader
	paramStart.Info.AcID = acidMember
	errCode, ret = start.Do(*trans, paramStart)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeOptInvalid, ret.(RetStartFight).Info.Code)

	// Member未准备
	paramStart.Info.AcID = acidLeader
	errCode, ret = start.Do(*trans, paramStart)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeStartFightFailed, ret.(RetStartFight).Info.Code)

	// BattleInfo为空
	paramReady := new(ParamReadyFight)
	paramReady.Info.AcID = acidMember
	paramReady.Info.Status = helper.TBPlayerStateReady
	ready.Do(*trans, paramReady)

	errCode, ret = start.Do(*trans, paramStart)

	assert.Equal(t, uint32(0), errCode)
	assert.Equal(t, helper.RetCodeStartFightFailed, ret.(RetStartFight).Info.Code)

	// MultPlay服务报错
	playerLeader := room.GetPlayer(acidLeader)
	playerLeader.BattleData = []byte{'t', 'e', 's', 't', '1'}
	playerMember := room.GetPlayer(acidMember)
	playerMember.BattleData = []byte{'t', 'e', 's', 't', '2'}

	errCode, ret = start.Do(*trans, paramStart)

	assert.Equal(t, uint32(2), errCode)
	assert.Equal(t, helper.RetCodeStartFightFailed, ret.(RetStartFight).Info.Code)

	// 到这里没搭好后续，没法继续测返回值了
	// TODO: 实现Multiplay和Teamboss交互
}
