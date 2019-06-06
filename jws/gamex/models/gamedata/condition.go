package gamedata

import (
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Condition struct {
	Ctyp   uint32 `json:"typ"`
	Param1 int64  `json:"p1"`
	Param2 int64  `json:"p2"`
	Param3 string `json:"p3"`
	Param4 string `json:"p4"`

	cidx int //标志在playerCondition中的位置
}

const (
	player_conditon_init_len = 16
)

type PlayerCondition struct {
	Conds     []*Condition
	Cond_next []int
}

func (p *PlayerCondition) Init() {
	p.Conds = make([]*Condition, 0, player_conditon_init_len)
	p.Cond_next = make([]int, 0, player_conditon_init_len)
}

func (p *PlayerCondition) RegCondition(c *Condition) {
	if len(p.Cond_next) > 0 {
		idx := p.Cond_next[len(p.Cond_next)-1]
		if p.Conds[idx] != nil {
			logs.Error("p.Conds[idx] Err %v %d - %v", p.Conds[idx], idx, p.Cond_next)
			p.Conds = append(p.Conds, c)
			c.cidx = len(p.Conds) - 1
		} else {
			p.Conds[idx] = c
			p.Cond_next = p.Cond_next[:len(p.Cond_next)-1]
			c.cidx = idx
		}
	} else {
		p.Conds = append(p.Conds, c)
		c.cidx = len(p.Conds) - 1
	}
}
