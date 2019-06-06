package guild_player_rank

import "vcs.taiyouxi.net/jws/gamex/models/helper"

type GuildPlayerRankInterface interface {
	OnPlayerSorce(p *helper.AccountSimpleInfo, sorce int64)
	OnPlayerSorceAdd(p *helper.AccountSimpleInfo, sorce int64)
	OnPlayerDel(accountID string)
	GetRank(acid string) (int, int64)
	GetSorce(acid string) int64
}

type GuildPlayerRankStoreInterface interface {
	Len() int
	Swap(i, j int)
	Add(i int, p *helper.AccountSimpleInfo, sorce int64)
	Del(i int)
	Clean()
	GetAcID(i int) string
	GetSorce(i int) int64
}
