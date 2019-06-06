package friend

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

const (
	FRIEND_DB_NAME = "friend"
)

func TableFriend(shardId uint) string {
	return fmt.Sprintf("%d:%d:%s", game.Cfg.Gid, game.Cfg.GetShardIdByMerge(shardId), FRIEND_DB_NAME)
}
