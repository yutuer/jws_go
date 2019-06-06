package ws_pvp

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*WSPVPModule
)

func init() {
	mInstance = make(map[uint]*WSPVPModule)
	modules.RegModule(modules.Module_WsPvp, newWsPvpModule)
}

func GetModule(shard uint) *WSPVPModule {
	mergeShardId := game.Cfg.GetShardIdByMerge(shard)
	groupId := gamedata.GetWSPVPGroupId(uint32(mergeShardId))
	return mInstance[uint(groupId)]
}

func newWsPvpModule(sid uint) modules.ServerModule {
	m := genModule(sid)
	mInstance[uint(m.groupId)] = m
	return m
}

func genModule(sid uint) *WSPVPModule {
	mergeShardId := game.Cfg.GetShardIdByMerge(sid)
	groupId := gamedata.GetWSPVPGroupId(uint32(mergeShardId))
	return &WSPVPModule{
		sid:     sid,
		groupId: int(groupId),
	}
}
