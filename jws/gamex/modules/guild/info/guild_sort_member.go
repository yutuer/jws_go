package guild_info

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

// 用于对公会成员排序的结构
type GuildSortMember []*helper.AccountSimpleInfo

func (g GuildSortMember) Len() int {
	return len(g)
}

func (g GuildSortMember) Less(i, j int) bool {
	if g[i].GuildPosition != g[j].GuildPosition {
		leftPosition := g[i].GuildPosition
		rightPosition := g[j].GuildPosition
		if leftPosition == 0 {
			leftPosition = gamedata.Guild_Pos_Count
		}
		if rightPosition == 0 {
			rightPosition = gamedata.Guild_Pos_Count
		}
		return leftPosition < rightPosition
	}
	if g[i].CurrCorpGs != g[j].CurrCorpGs {
		return g[i].CurrCorpGs > g[j].CurrCorpGs
	}
	if g[i].GetOnline() {
		return true
	}
	if g[j].GetOnline() {
		return true
	}
	return g[i].LastLoginTime > g[j].LastLoginTime
}

func (g GuildSortMember) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}
