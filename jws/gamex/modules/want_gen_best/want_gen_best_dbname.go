package want_gen_best

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const (
	Table_Want_Gen_Best = "wantgenbest"
)

func tableWantGenBest(shardId uint) string {
	return fmt.Sprintf("%s:%d:%d", Table_Want_Gen_Best, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(shardId))
}
