package guild_player_rank

import "vcs.taiyouxi.net/jws/gamex/models/helper"

type playerSimpleInfoRankStore struct {
	Datas  []helper.AccountSimpleInfo `json:"datas"`
	Sorces []int64                    `json:"sorces"`
}

type PlayerSimpleInfoRank struct {
	Rank playerSimpleInfoRankStore `json:"r"`
	PlayerInGuildRank
}

func NewPlayerSimpleInfoRank() *PlayerSimpleInfoRank {
	return NewPlayerSimpleInfoRankByCap(64)
}

func NewPlayerSimpleInfoRankByCap(cap int) *PlayerSimpleInfoRank {
	res := new(PlayerSimpleInfoRank)
	res.Rank.Datas = make([]helper.AccountSimpleInfo, 0, cap)
	res.Rank.Sorces = make([]int64, 0, cap)
	res.Init()
	return res
}

func (p *PlayerSimpleInfoRank) Init() {
	p.InitData(&(p.Rank))
}

func (p *PlayerSimpleInfoRank) GetSimpleInfo(i int) *helper.AccountSimpleInfo {
	return &(p.Rank.Datas[i])
}

func (p *playerSimpleInfoRankStore) Len() int {
	return len(p.Datas)
}

func (p *playerSimpleInfoRankStore) Swap(i, j int) {
	if i == j {
		return
	}
	td := p.Datas[i]
	p.Datas[i] = p.Datas[j]
	p.Datas[j] = td

	ts := p.Sorces[i]
	p.Sorces[i] = p.Sorces[j]
	p.Sorces[j] = ts
}

func (p *playerSimpleInfoRankStore) Add(i int, acc *helper.AccountSimpleInfo, sorce int64) {
	if i == len(p.Datas) {
		p.Datas = append(p.Datas, *acc)
		p.Sorces = append(p.Sorces, sorce)
	} else if i < len(p.Datas) && i >= 0 {
		p.Datas[i] = *acc
		p.Sorces[i] = sorce
	}
}

func (p *playerSimpleInfoRankStore) Del(i int) {
	if i < len(p.Datas) && i >= 0 {
		last := len(p.Datas) - 1
		p.Datas[i] = p.Datas[last]
		p.Sorces[i] = p.Sorces[last]
		p.Datas[last] = helper.AccountSimpleInfo{}
		p.Sorces[last] = 0
		p.Datas = p.Datas[:last]
		p.Sorces = p.Sorces[:last]
	}
}
func (p *playerSimpleInfoRankStore) Clean() {
	p.Datas = p.Datas[0:0]
	p.Sorces = p.Sorces[0:0]
}
func (p *playerSimpleInfoRankStore) GetAcID(i int) string {
	if i < 0 || i >= len(p.Datas) {
		return ""
	} else {
		return p.Datas[i].AccountID
	}
}
func (p *playerSimpleInfoRankStore) GetSorce(i int) int64 {
	if i < 0 || i >= len(p.Sorces) {
		return 0
	} else {
		return p.Sorces[i]
	}
}
