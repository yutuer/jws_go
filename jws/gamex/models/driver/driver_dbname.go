package driver

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

// names:0:10
func TableChangeName(sid uint) string {
	return fmt.Sprintf("%s:%d:%d", "names", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}
