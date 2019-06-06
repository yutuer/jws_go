package guild_player_rank

import (
	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

type PlayerInGuildRank struct {
	data  GuildPlayerRankStoreInterface
	index map[string]int
}

func (p *PlayerInGuildRank) InitData(d GuildPlayerRankStoreInterface) {
	p.data = d
	len := p.data.Len()
	if len < 64 {
		len = 64
	}
	p.index = make(map[string]int, len)
	for i := 0; i < p.data.Len(); i++ {
		p.index[p.data.GetAcID(i)] = i
	}
}

func (p *PlayerInGuildRank) sort() {
	sort.Sort(p)
	p.InitData(p.data)
}

func (p *PlayerInGuildRank) get(acID string) (int, bool) {
	if p.index == nil {
		p.InitData(p.data)
	}
	res, ok := p.index[acID]
	return res, ok
}

// Len is the number of elements in the collection.
func (p *PlayerInGuildRank) Len() int {
	return p.data.Len()
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p *PlayerInGuildRank) Less(i, j int) bool {
	is := p.data.GetSorce(i)
	js := p.data.GetSorce(j)
	if is < js {
		return true
	} else if is == js {
		return p.data.GetAcID(i) < p.data.GetAcID(j)
	} else {
		return false
	}
}

// Swap swaps the elements with indexes i and j.
func (p *PlayerInGuildRank) Swap(i, j int) {
	p.data.Swap(i, j)
}

func (p *PlayerInGuildRank) OnPlayerSorce(acc *helper.AccountSimpleInfo, sorce int64) {
	idx, ok := p.get(acc.AccountID)
	if !ok {
		p.data.Add(p.data.Len(), acc, sorce)
	} else {
		p.data.Add(idx, acc, sorce)
	}
	p.sort()
}

func (p *PlayerInGuildRank) OnPlayerSorceNoSort(acc *helper.AccountSimpleInfo, sorce int64) {
	idx, ok := p.get(acc.AccountID)
	if !ok {
		p.data.Add(p.data.Len(), acc, sorce)
	} else {
		p.data.Add(idx, acc, sorce)
	}
}

func (p *PlayerInGuildRank) Sort() {
	p.sort()
}

func (p *PlayerInGuildRank) OnPlayerSorceAdd(acc *helper.AccountSimpleInfo, sorce int64) {
	if sorce == 0 {
		return
	}
	idx, ok := p.get(acc.AccountID)
	if !ok {
		p.data.Add(p.data.Len(), acc, sorce)
	} else {
		sorceOld := p.data.GetSorce(idx)
		p.data.Add(idx, acc, sorce+sorceOld)
	}
	p.sort()
}

func (p *PlayerInGuildRank) OnPlayerDel(accountID string) {
	idx, ok := p.get(accountID)
	if ok {
		p.data.Del(idx)
		delete(p.index, accountID)
		p.sort()
	}
}

func (p *PlayerInGuildRank) GetRank(acid string) int {
	idx, ok := p.get(acid)
	if !ok {
		return -1
	} else {
		return idx
	}
}

func (p *PlayerInGuildRank) GetSorce(acid string) int64 {
	idx, ok := p.get(acid)
	if !ok {
		return 0
	} else {
		return p.data.GetSorce(idx)
	}
}

func (p *PlayerInGuildRank) GetRankWithSorce(acid string) (int, int64) {
	idx, ok := p.get(acid)
	if !ok {
		return 0, 0
	} else {
		return idx, p.data.GetSorce(idx)
	}
}
