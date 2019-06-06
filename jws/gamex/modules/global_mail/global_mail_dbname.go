package global_mail

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const (
	table_global_mail = "all"
)

func TableGlobalMailName(sid uint) string {
	return fmt.Sprintf("%s:%d", table_global_mail, game.Cfg.GetShardIdByMerge(sid))
}
