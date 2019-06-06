package account

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

type PlayerFirstPassRankReward struct {
	RewardStat [gamedata.FirstPassRankTypCount][]int `json:"rs"`
	MaxRank    [gamedata.FirstPassRankTypCount]int   `json:"mr"`
}

func (p *PlayerFirstPassRankReward) OnAccountInit() {
	for i := 0; i < len(p.RewardStat); i++ {
		p.RewardStat[i] = make([]int, 0, 64)
	}
}

func (p *PlayerFirstPassRankReward) OnAfterLogin() {

}

func (p *PlayerFirstPassRankReward) OnRank(t, rank int) {
	is1MineTheMax := gamedata.GetFirstPassRankIs1MineTheMax(t)
	if is1MineTheMax {
		if p.MaxRank[t] <= 0 {
			p.MaxRank[t] = rank
		} else if rank < p.MaxRank[t] {
			p.MaxRank[t] = rank
		}
	} else {
		if p.MaxRank[t] < rank {
			p.MaxRank[t] = rank
		}
	}
}

func (p *PlayerFirstPassRankReward) IsRankMaxCanGetReward(t, rankNeed int) bool {
	is1MineTheMax := gamedata.GetFirstPassRankIs1MineTheMax(t)
	//logs.Warn("IsRankMaxCanGetReward %v %v %v %v", is1MineTheMax, t, rankNeed, p.MaxRank[t])
	if is1MineTheMax {
		if p.MaxRank[t] <= 0 {
			return false
		}
		return p.MaxRank[t] <= rankNeed
	} else {
		return p.MaxRank[t] >= rankNeed
	}
}

func (p *PlayerFirstPassRankReward) GetRankMax(t int) int {
	return p.MaxRank[t]
}

func (p *PlayerFirstPassRankReward) AddFirstPassReward(t, idx int) bool {
	if idx < 0 || t < 0 || t >= len(p.RewardStat) {
		return false
	}

	for i := 0; i < len(p.RewardStat[t]); i++ {
		if p.RewardStat[t][i] == idx {
			return false
		}
	}

	p.RewardStat[t] = append(p.RewardStat[t], idx)
	return true
}

func (p *PlayerFirstPassRankReward) GetReward(t, idx int) *gamedata.FirstPassRewardData {
	r := gamedata.GetFirstPassRewardData(t)
	if r == nil {
		return nil
	}
	if idx < 0 || idx >= len(r) {
		return nil
	}
	return &r[idx]
}
