package account

type PlayerGrowFund struct {
	IsActivate bool
	Bought     []uint32
}

func (gf *PlayerGrowFund) IfNotBuyThenBuy(lvl uint32) bool {
	gf._init()
	for _, id := range gf.Bought {
		if id == lvl {
			return false
		}
	}
	gf.Bought = append(gf.Bought, lvl)
	return true
}

func (gf *PlayerGrowFund) _init() {
	if gf.Bought == nil {
		gf.Bought = make([]uint32, 0, 10)
	}
}
