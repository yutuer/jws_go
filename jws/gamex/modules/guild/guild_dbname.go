package guild

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Table_GuildName        = "guild:name"          // 公会名称表，用来排查名字重复
	Table_GuildId2uuid     = "guild:id2uuid"       // 公会id->uuid，用于公会查找功能
	Table_GuildSeed        = "guild:idseed"        // 公会id自增发生器
	Table_Account2Guild    = "guild:account2guild" // player->公会，用于查找玩家是否属于某一公会
	Table_PlayerGuildApply = "guild:playerapply"   // 玩家申请的公会
	Table_GuildApply       = "guild:guildapply"    // 公会的申请列表
)

func TableGuildName(sid uint) string {
	return fmt.Sprintf("%s:%d:%d", Table_GuildName, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableGuildId2Uuid(sid uint) string {
	return fmt.Sprintf("%s:%d:%d", Table_GuildId2uuid, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func tableGuildSeed(sid uint) string {
	return fmt.Sprintf("%s:%d:%d", Table_GuildSeed, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}

func TableAccount2GuildByAccount(acid string) string {
	account, err := db.ParseAccount(acid)
	if err != nil {
		logs.Error("guild UpdateGuildToPlayer %v err %v", acid, err)
		return ""
	}
	return TableAccount2Guild(account.ShardId)
}

func TableAccount2Guild(sid uint) string {
	return fmt.Sprintf("%s:%d:%d", Table_Account2Guild, game.Cfg.Gid, game.Cfg.GetShardIdByMerge(sid))
}
