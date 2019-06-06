package currency

import (
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var HC_Cost_Typ_Seq [HC_TYPE_COUNT + 1][HC_TYPE_COUNT]int = [HC_TYPE_COUNT + 1][HC_TYPE_COUNT]int{
	[HC_TYPE_COUNT]int{HC_From_Buy, HC_From_Give, HC_From_Compensate},
	[HC_TYPE_COUNT]int{HC_From_Give, HC_From_Buy, HC_From_Compensate},
	[HC_TYPE_COUNT]int{HC_From_Compensate, HC_From_Buy, HC_From_Give},
	[HC_TYPE_COUNT]int{HC_From_Compensate, HC_From_Give, HC_From_Buy}, // 常规扣除次序
}

type HardCurrency struct {
	Currency          [HC_TYPE_COUNT + 1]int64 // 最后一位缓存和
	BuyFromHc         int64
	BuyFromHcToday    int64
	BuyFromHcLastTime int64
	CostHcToday       int64
	CostHcLastTime    int64
	handler           events.Handler
}

func (p *HardCurrency) useByTyp(
	AccountId string,
	t int,
	m int64,
	last_cost_typ int,
	reason string) (
	to_be_used int64, // 扣除完之后还有多少需要扣除
	cost_typ int, // 这次消费算哪种
) {
	cost_typ = last_cost_typ
	if m == 0 {
		to_be_used = 0
		return
	}

	//	old := p.Currency[t]

	if p.Currency[t] >= m {
		p.Currency[t] -= m
		to_be_used = 0

	} else {
		to_be_used = m - p.Currency[t]
		p.Currency[t] = 0
	}

	// 本次消费扣除类型按照 HC_From_Buy 》 HC_From_Give 》 HC_From_Compensate 顺序来确定
	if to_be_used < m && t < last_cost_typ {
		cost_typ = t
	}

	// TBD hc详细的变化记录，看是否还要？
	//	helper.LogPlayerCost(AccountId, reason,
	//		"UseTypHc %d %d-%d TureUse %d From %d With TobeUse %d",
	//		m, cost_typ, t, m-to_be_used, old, to_be_used)

	return
}

// 根据typ_cost_first所对应的扣除次序(在 HC_Cost_Typ_Seq 定义)扣除硬通，返回是否成功及本次扣除的类型
func (p *HardCurrency) UseHc(AccountId string, typ_cost_first int, m int64, now_time int64, reason string) (bool, int) {
	if m <= 0 {
		return true, typ_cost_first
	}

	if p.HasHC(m) {
		logs.Trace("[%s]Use %d First Hc %d By %s", AccountId, typ_cost_first, m, reason)

		oldHc := p.Currency

		cost_seq := HC_Cost_Typ_Seq[typ_cost_first]

		var lasted int64 = m
		var cost_typ int = HC_TYPE_COUNT

		lasted, cost_typ = p.useByTyp(AccountId, cost_seq[0], lasted, cost_typ, reason)
		lasted, cost_typ = p.useByTyp(AccountId, cost_seq[1], lasted, cost_typ, reason)
		lasted, cost_typ = p.useByTyp(AccountId, cost_seq[2], lasted, cost_typ, reason)

		if lasted != 0 {
			logs.Error("[%s]Use Buy First lasted %d No 0 !", AccountId, lasted)
			// TBD by Fanyang 暂时不返还
		}

		p.Currency[HC_TYPE_COUNT] -= m

		// logiclog
		for i, _ := range p.Currency {
			if p.Currency[i] != oldHc[i] {
				if p.handler != nil {
					p.handler.OnHcChg(false, i, oldHc[i], oldHc[i]-p.Currency[i], reason)
				}
			}
		}

		// 记录每日消耗钻数量，为任务用
		p.hcCostToday(m, now_time)

		return true, cost_typ
	} else {
		return false, 0
	}
}

// 常规扣除方式，适用于出宝箱之外的扣除HC逻辑
func (p *HardCurrency) UseHcGiveFirst(AccountId string, m int64, now_time int64, reason string) bool {
	ok, _ := p.UseHc(AccountId, HC_TYPE_COUNT, m, now_time, reason)
	return ok
}

func (p *HardCurrency) AddHC(AccountId string, t int, m int64, now_time int64, reason string) {
	if t >= HC_TYPE_COUNT {
		logs.Error("No SC Type %d", t)
		return
	}

	if p.handler != nil {
		p.handler.OnHcChg(true, t, p.Currency[HC_TYPE_COUNT], m, reason)
	}

	p.Currency[t] += m
	p.Currency[HC_TYPE_COUNT] += m

	// 记录每日购买钻石数，为任务用
	if t == HC_From_Buy {
		p.addHcFromBuyToday(m, now_time)
	}
}

func (p *HardCurrency) GetHC() int64 {
	return p.Currency[HC_TYPE_COUNT]
}

func (p *HardCurrency) GetHCFromBy() int64 {
	return p.Currency[HC_From_Buy]
}

func (p *HardCurrency) GetHCFromGive() int64 {
	return p.Currency[HC_From_Give]
}

func (p *HardCurrency) GetHCFromCompensate() int64 {
	return p.Currency[HC_From_Compensate]
}

func (p *HardCurrency) HasHC(m int64) bool {
	return p.Currency[HC_TYPE_COUNT] >= m
}

func (p *HardCurrency) GetAll() []int64 {
	return p.Currency[:]
}

func (p *HardCurrency) SetHandler(handler events.Handler) {
	p.handler = handler
}

func (p *HardCurrency) addHcFromBuyToday(m, now_time int64) {
	p.BuyFromHc += m
	if !gamedata.IsSameDayCommon(p.BuyFromHcLastTime, now_time) {
		p.BuyFromHcToday = m
	} else {
		p.BuyFromHcToday += m
	}
	p.BuyFromHcLastTime = now_time
}

func (p *HardCurrency) GetBuyFromHcToday(now_time int64) int64 {
	if !gamedata.IsSameDayCommon(p.BuyFromHcLastTime, now_time) {
		return 0
	}
	return p.BuyFromHcToday
}

func (p *HardCurrency) hcCostToday(m, now_time int64) {
	if !gamedata.IsSameDayCommon(p.CostHcLastTime, now_time) {
		p.CostHcToday = m
	} else {
		p.CostHcToday += m
	}
	p.CostHcLastTime = now_time
}

func (p *HardCurrency) GetCostHcToday(now_time int64) int64 {
	if !gamedata.IsSameDayCommon(p.CostHcLastTime, now_time) {
		return 0
	}
	return p.CostHcToday
}
