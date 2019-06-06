package global_info

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func tableGlobalLevelFinish(sid uint) string {
	return fmt.Sprintf("global:levelfinish:%d:%d", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}
