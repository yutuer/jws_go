package moneycat_marquee

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const (
	Table_Dest_MoneyCat = "moneycat"
)

func tableDestMoneyCat(shardId uint) string {
	return fmt.Sprintf("%s:%d:%d", Table_Dest_MoneyCat, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(shardId))
}
