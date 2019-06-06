package dest_gen_first

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const (
	Table_Dest_Gen_First = "destgenfirst"
)

func TableDestGenFirst(shardId uint) string {
	return fmt.Sprintf("%s:%d:%d", Table_Dest_Gen_First, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(shardId))
}
