package market_activity

import "sort"

type pair struct {
	Acid  string
	Score float64
}

type pairlist []pair

func (p pairlist) Len() int {
	return len(p)
}

func (p pairlist) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p pairlist) Less(i, j int) bool {
	return p[i].Score > p[j].Score
}

func sortFromAcid2score(acid2score map[string]float64) []pair {
	list := []pair{}
	for k, v := range acid2score {
		list = append(list, pair{k, v})
	}

	sort.Sort(pairlist(list))

	return list
}

func (p rewardParam) Len() int {
	return len(p.conds)
}

func (p rewardParam) Swap(i, j int) {
	p.conds[i], p.conds[j] = p.conds[j], p.conds[i]
}

func (p rewardParam) Less(i, j int) bool {
	return p.conds[i].rankTop < p.conds[j].rankTop
}
