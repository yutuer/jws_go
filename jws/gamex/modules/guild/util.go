package guild

import (
	"fmt"
	"strconv"

	"strings"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	MaxRandGuilds = helper.MaxRandGuilds

	Guild_DB_Counter_Name = "Guild_DB"

	GuildId_Min       = 100000
	GuildId_Bit_Count = 5
)

func genGuildIdByShard(shardId uint, id int64) int64 {
	if id >= GuildId_Min {
		logs.Error("%d guildId greater than %d !!!", id, GuildId_Min)
	}
	return int64(shardId)*GuildId_Min + id
}

func GuildIdAtoi(s string) int64 {
	if len(s) < GuildId_Bit_Count+1 {
		logs.Error("GuildIdAtoi %s guild illegal", s)
		return 0
	}
	s_id := s[len(s)-GuildId_Bit_Count:]
	id, err := strconv.Atoi(s_id)
	if err != nil {
		logs.Warn("GuildIdAtoi Atoi err by %s", s)
		return 0
	}
	prefix := s[:len(s)-GuildId_Bit_Count]
	iprefix, err := strconv.ParseInt(prefix, 32, 64)
	if err != nil {
		logs.Warn("GuildIdAtoi ParseInt err by %s", s)
		return 0
	}
	return iprefix*GuildId_Min + int64(id)
}

func GuildItoa(guildId int64) string {
	if guildId < GuildId_Min {
		//logs.Error("GuildItoa %d guildId illegal", guildId)
		return ""
	}
	prefix := guildId / GuildId_Min
	s := strconv.FormatInt(prefix, 32)
	return fmt.Sprintf("%s%05d", strings.ToUpper(s), guildId%GuildId_Min)
}
