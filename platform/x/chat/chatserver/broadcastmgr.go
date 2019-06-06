package chatserver

import "vcs.taiyouxi.net/platform/planx/util/logs"

const (
	Typ_SysNotice  = "SysRollNotice"
	Typ_FishInfo   = "FishInfo"
	Typ_Gve        = "Gve"
	Typ_FengHuo    = "FengHuo"
	Typ_FilterRoom = "FilterRoom"
	Typ_GuildRoom  = "GuildRoom"
)

func RouteBroadCast(typ, shard, msg string, acids []string) {
	shard = GetGidSidMerge(shard)
	city_mgr.broadCast(typ, shard, msg, acids)
	logs.Info("broadcast msg %v %v %v %v", typ, shard, msg, acids)
}

func ClearRoomsTooFewPlayer(shard string) {
	city_mgr.broadCast(Typ_FilterRoom, shard, "", nil)
	logs.Info("ClearRoomsTooFewPlayer !!!")
}
