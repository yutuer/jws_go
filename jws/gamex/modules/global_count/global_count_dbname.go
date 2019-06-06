package global_count

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func tableGlobalCount(gc string, sid uint) string {
	return fmt.Sprintf("globalcount:%s:%d:%d", gc, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}
