package team_pvp

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func TableTeamPvpRank(sid uint) string {
	return fmt.Sprintf("teampvp:%d:%d", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}
