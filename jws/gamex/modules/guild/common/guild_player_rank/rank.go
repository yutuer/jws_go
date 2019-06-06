package guild_player_rank

import (
	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GuildPlayerInfoInRank struct {
	Name          string `json:"name" codec:"name"`
	AccountID     string `json:"aid" codec:"aid"`
	CorpLv        uint32 `json:"corplv" codec:"corplv"`
	GuildPosition int    `json:"position" codec:"position"` // 这个信息和其他信息不同是从工会向玩家存档更新, 注意其变化逻辑
	LastLoginTime int64  `json:"l_login" codec:"l_login"`
	CurrAvatar    int    `json:"curr" codec:"curr"`
	CurrAvatarGs  int    `json:"currgs" codec:"currgs"`
	TitleOn       string `json:"ttlo" codec:"ttlo"`   // 当前头顶的称号
	TitleTimeOut  int64  `json:"ttlto" codec:"ttlto"` // 称号过期时间，只影响有过期时间的称号
	Score         int64  `json:"score" codec:"score"`
}

func (g *GuildPlayerInfoInRank) FromSimpleInfo(s *helper.AccountSimpleInfo, score int64) {
	g.Name = s.Name
	g.AccountID = s.AccountID
	g.CorpLv = s.CorpLv
	g.GuildPosition = s.GuildPosition
	g.LastLoginTime = s.LastLoginTime
	g.CurrAvatar = s.CurrAvatar
	g.CurrAvatarGs = s.CurrCorpGs
	g.TitleOn = s.TitleOn
	g.TitleTimeOut = s.TitleTimeOut
	if score > g.Score {
		g.Score = score
	}
}

type GuildPlayerRank struct {
	Players   [helper.MaxGuildMember + 1]GuildPlayerInfoInRank
	playerLen int
	index     map[string]int
}

func (g *GuildPlayerRank) Clean() {
	for i := 0; i < len(g.Players); i++ {
		g.Players[i] = GuildPlayerInfoInRank{}
	}
	g.playerLen = 0
	g.index = make(map[string]int, len(g.Players))
}

// Len is the number of elements in the collection.
func (g *GuildPlayerRank) Len() int {
	return g.playerLen
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (g *GuildPlayerRank) Less(i, j int) bool {
	return g.Players[i].Score > g.Players[j].Score
}

// Swap swaps the elements with indexes i and j.
func (g *GuildPlayerRank) Swap(i, j int) {
	if i == j {
		return
	}
	t := g.Players[i]
	g.Players[i] = g.Players[j]
	g.Players[j] = t
}

func (g *GuildPlayerRank) Sort() {
	sort.Stable(g)

	g.index = make(map[string]int, len(g.Players))
	for i := 0; i < g.playerLen; i++ {
		g.index[g.Players[i].AccountID] = i
	}
}

func (g *GuildPlayerRank) InitOnRestart() {
	for i := 0; i < len(g.Players); i++ {
		if g.Players[i].AccountID != "" {
			g.playerLen++
		}
	}
	g.Sort()
	logs.Debug("init guild boss, %v", g)
}

func (g *GuildPlayerRank) OnPlayerSorce(p *helper.AccountSimpleInfo, sorce int64) {
	if sorce <= 0 {
		return
	}
	pRank, ok := g.index[p.AccountID]
	if !ok {
		if g.playerLen >= len(g.Players) {
			logs.Error("OnPlayerSorce player too large by %s", p.AccountID)
			return
		}
		g.Players[g.playerLen].FromSimpleInfo(p, sorce)
		g.playerLen++
	} else {
		g.Players[pRank].FromSimpleInfo(p, sorce)
	}
	g.Sort()
}

func (g *GuildPlayerRank) OnPlayerSorceAdd(p *helper.AccountSimpleInfo, sorce int64) {
	if sorce <= 0 {
		return
	}
	pRank, ok := g.index[p.AccountID]
	if !ok {
		if g.playerLen >= len(g.Players) {
			logs.Error("OnPlayerSorce player too large by %s", p.AccountID)
			return
		}
		g.Players[g.playerLen].FromSimpleInfo(p, sorce)
		g.playerLen++
	} else {
		g.Players[pRank].FromSimpleInfo(p, g.Players[pRank].Score+sorce)
	}
	g.Sort()
}

func (g *GuildPlayerRank) UpdatePlayerInfo(p *helper.AccountSimpleInfo) {
	if pRank, ok := g.index[p.AccountID]; ok {
		g.Players[pRank].FromSimpleInfo(p, g.Players[pRank].Score)
	}
}

func (g *GuildPlayerRank) OnPlayerNew(p *helper.AccountSimpleInfo) {
	if g.playerLen >= len(g.Players) {
		logs.Error("OnPlayerSorce player too large by %s", p.AccountID)
		return
	}
	g.Players[g.playerLen].FromSimpleInfo(p, 0)
	g.index[p.AccountID] = g.playerLen
	g.playerLen++
	return
}

func (g *GuildPlayerRank) OnPlayerDel(accountID string) {
	pRank, ok := g.index[accountID]
	logs.Debug("guild boss on player del, %d", accountID, pRank)
	if ok {
		g.playerLen--
		for i := pRank; i < g.playerLen; i++ {
			g.Players[i] = g.Players[i+1]
		}
		g.Players[g.playerLen] = GuildPlayerInfoInRank{}
		g.Sort()
	}
}

func (g *GuildPlayerRank) GetTop(n int) []GuildPlayerInfoInRank {
	res := make([]GuildPlayerInfoInRank, 0, n)
	for i := 0; i < n; i++ {
		res = append(res, g.Players[i])
	}
	return res[:]
}

func (g *GuildPlayerRank) GetRank(acid string) (int, int64) {
	rank, ok := g.index[acid]
	if ok {
		return rank + 1, g.Players[rank].Score
	} else {
		return 0, 0
	}
}

func (g *GuildPlayerRank) GetSorce(acid string) int64 {
	rank, ok := g.index[acid]
	if ok {
		return g.Players[rank].Score
	} else {
		return 0
	}
}
