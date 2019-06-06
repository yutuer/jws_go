package balance

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func tableRankBalance(sid uint) string {
	return fmt.Sprintf("%d:%d:RankBalance", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}
