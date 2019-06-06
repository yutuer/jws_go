package city_fish

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const (
	Table_Fish_Reward = "fishreward"
)

func tableFishReward(shardId uint) string {
	return fmt.Sprintf("%s:%d:%d", Table_Fish_Reward, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(shardId))
}
