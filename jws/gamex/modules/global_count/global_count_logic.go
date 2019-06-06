package global_count

import (
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func AddRecordCount(shardId, gameId uint, id uint32) (bool, map[uint32]uint32) {
	ret := GetModule(shardId).CommandExec(GlobalCountCmd{
		CmdTyp:   GlobalCount_Cmd_AddAndGet,
		CountTyp: GlobalCount_Typ_Record,
		Gid:      gameId,
		Sid:      shardId,
		Key:      GlobalCountKey{IId: id},
	})
	if !ret.Success {
		return false, ret.Counti2c
	}
	return true, ret.Counti2c
}

func GetRecordCount(shardId, gameId uint) map[uint32]uint32 {
	ret := GetModule(shardId).CommandExec(GlobalCountCmd{
		CmdTyp:   GlobalCount_Cmd_GetInfo,
		CountTyp: GlobalCount_Typ_Record,
		Gid:      gameId,
		Sid:      game.Cfg.GetShardIdByMerge(shardId),
	})
	if !ret.Success {
		logs.Error("getRecordCount err")
		return map[uint32]uint32{}
	}
	return ret.Counti2c
}
