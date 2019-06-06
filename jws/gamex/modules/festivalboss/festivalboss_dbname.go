package festivalboss

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const (
	Table_Dest_FestivalBoss = "festivalboss"
)

func tableDestFestivalBoss(shardId uint) string {
	return fmt.Sprintf("%s:%d:%d", Table_Dest_FestivalBoss, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(shardId))
}
