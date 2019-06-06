package currency

import (
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// SC_TYPE_COUNT
// 0 - 金钱
// 1 - 精铁
// 2 - 扫荡券
//

type SoftCurrency struct {
	Currency [SC_TYPE_COUNT]int64
	handler  events.Handler
}

// 货币 软通 金钱接口
func (p *SoftCurrency) UseSC(t int, m int64, reason string) bool {
	if m <= 0 {
		return true
	}

	if p.HasSC(t, m) {
		tTotal := p.GetSC(t)

		if p.handler != nil {
			p.handler.OnScChg(false, t, tTotal, m, reason)
		}

		p.Currency[t] -= m
		return true
	} else {
		return false
	}
}

func (p *SoftCurrency) AddSC(t int, m int64, reason string) {
	if t >= len(p.Currency) {
		logs.Error("No SC Type %d", t)
		return
	}

	tTotal := p.GetSC(t)

	if p.handler != nil {
		p.handler.OnScChg(true, t, tTotal, m, reason)
	}

	p.Currency[t] += m
}

func (p *SoftCurrency) GetSC(t int) int64 {
	if t >= len(p.Currency) {
		return 0
	} else {
		return p.Currency[t]
	}
}

func (p *SoftCurrency) HasSC(t int, m int64) bool {
	if t >= len(p.Currency) {
		return false
	} else {
		return p.Currency[t] >= m
	}
}

func (p *SoftCurrency) GetAll() []int64 {
	return p.Currency[:]
}

func (p *SoftCurrency) SetHandler(handler events.Handler) {
	p.handler = handler
}
