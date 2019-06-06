package hero_diff

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func TableHeroDiff(shardId uint) string {
	return fmt.Sprintf("%d:%d:HeroDiff", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(shardId))
}
