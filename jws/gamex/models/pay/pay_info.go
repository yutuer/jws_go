package pay

import "time"

type Order struct {
	TransID   string `json:"id"`
	GoodID    string `json:"good_id"`
	StartTime int64  `json:"start"`
	ErrInfo   string `json:"err"`
}

type PlayerPayInfo struct {
	Goods []Order `json:"goods"`
}

func (p *PlayerPayInfo) Init() {
	p.Goods = make([]Order, 0, 8)
}

func (p *PlayerPayInfo) AddIAPOrder(transID, goodID string) {
	if p.Goods == nil {
		p.Goods = make([]Order, 0, 8)
	}
	p.Goods = append(p.Goods, Order{
		TransID:   transID,
		GoodID:    goodID,
		StartTime: time.Now().Unix(),
	})
}

func (p *PlayerPayInfo) IsHasTransID(transID string) bool {
	for _, g := range p.Goods {
		if g.TransID == transID {
			return true
		}
	}
	return false
}
