package gvg

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const (
	GVG_DB_NAME  = "gvg"
	GVG_DB_MERGE = "gvgmerge"
)

func TableGVG(shardId uint) string {
	return fmt.Sprintf("%d:%d:%s", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(shardId), GVG_DB_NAME)
}

func TableGVGMerge(shardId uint) string {
	return fmt.Sprintf("%d:%d:%s", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(shardId), GVG_DB_MERGE)
}
