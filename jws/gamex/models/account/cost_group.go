package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 统一的消耗接口，游戏中经常要消耗玩家一部分物品软通去做一件事
// 这里是一个统一的接口，可以利用AddXXX添加要消耗的东西，如果返回false则表示玩家物品不足，没法消耗

type CostGroup struct {
	items          []uint32               // 要消耗的各种道具
	count          []uint32               // 要消耗的道具数量
	jades          []uint32               // 要消耗的宝石id
	jcount         []uint32               // 要消耗的宝石数量
	fashions       []uint32               // 要消耗的时装id
	sc             [SC_TYPE_COUNT]int64   // 要消耗的各种软通
	hc             int64                  // 要消耗的hc
	energy         int64                  // 要消耗的体力
	wheelcoin      int64                  // 要消耗的幸运抽奖币
	bossFightPoint int64                  // 要消耗的军令
	HeroStarPiece  [AVATAR_NUM_MAX]uint32 //要消耗的主将碎片
}

func CostBySync(
	p *Account,
	data *gamedata.CostData,
	sync helper.ISyncRsp,
	reason string) bool {

	g := CostGroup{}
	if !g.AddCostData(p, data) {
		return false
	} else {
		return g.CostBySync(p, sync, reason)
	}
}

func (c *CostGroup) AddCostData(account *Account, data *gamedata.CostData) bool {
	for sc_t, sc_v := range data.Sc {
		if !c.AddSc(account, sc_t, sc_v) {
			return false
		}
	}

	for idx, id := range data.Items {
		if !c.AddItemById(account, id, data.Count[idx]) {
			return false
		}
	}

	for _, hc_v := range data.Hc {
		if !c.AddHc(account, hc_v) {
			return false
		}
	}
	if data.WheelCoin > 0 {
		if !c.AddWheelCoin(account, int64(data.WheelCoin)) {
			return false
		}
	}
	if data.Energy > 0 {
		if !c.AddEnergy(account, int64(data.Energy)) {
			return false
		}
	}
	for sc_t, sc_v := range data.HeroPiece {
		if !c.AddHeroStarPiece(account, sc_t, sc_v) {
			return false
		}
	}

	return true
}

func (c *CostGroup) AddItem(account *Account, id string, count uint32) bool {
	data := gamedata.CostData{}
	data.AddItem(id, count)
	return c.AddCostData(account, &data)
}

func (c *CostGroup) AddSc(account *Account, sc_t int, sc_v int64) bool {
	if sc_t >= SC_TYPE_COUNT {
		logs.SentryLogicCritical(account.AccountID.String(),
			"AddSc Err Unknown Typ %d", sc_t)
		return false
	}

	if sc_v < 0 {
		logs.Warn("AddSc Err Sc Cost %d < 0", sc_v)
		return false
	}

	if sc_v == 0 {
		return true
	}

	if !account.Profile.GetSC().HasSC(sc_t, c.sc[sc_t]+sc_v) {
		return false
	}

	c.sc[sc_t] += sc_v
	return true
}

func (c *CostGroup) AddEnergy(account *Account, v int64) bool {
	if v < 0 {
		logs.Warn("AddEnergy Err Sc Cost %d < 0", v)
		return false
	}

	if v == 0 {
		return true
	}

	if !account.Profile.GetEnergy().Has(c.energy + v) {
		return false
	}

	c.energy += v
	return true
}

func (c *CostGroup) AddWheelCoin(account *Account, v int64) bool {
	if v < 0 {
		logs.Warn("AddWheelCoin Err Sc Cost %d < 0", v)
		return false
	}

	if v == 0 {
		return true
	}

	if !account.Profile.GetWheelGachaInfo().Has(c.wheelcoin + v) {
		return false
	}

	c.wheelcoin += v
	return true
}

func (c *CostGroup) AddBossFightPoint(account *Account, v int64) bool {
	if v < 0 {
		logs.Warn("AddBossFightPoint Err Sc Cost %d < 0", v)
		return false
	}

	if v == 0 {
		return true
	}

	if !account.Profile.GetBossFightPoint().Has(c.bossFightPoint + v) {
		return false
	}

	c.bossFightPoint += v
	return true
}

func (c *CostGroup) AddHc(account *Account, v int64) bool {
	if v < 0 {
		logs.Warn("AddHc Err Sc Cost %d < 0", v)
		return false
	}

	if v == 0 {
		return true
	}

	if !account.Profile.GetHC().HasHC(v) {
		return false
	}

	c.hc += v
	return true
}

func (c *CostGroup) AddHeroStarPiece(account *Account, v_a int, v_n uint32) bool {
	if v_n < 0 {
		logs.Warn("AddHeroStarPiece Err Sc Cost %d < 0", v_n)
		return false
	}

	if v_n == 0 {
		return true
	}
	if account.Profile.GetHero().HeroStarPiece[v_a] < v_n {
		return false
	}
	c.HeroStarPiece[v_a] += v_n
	return true
}

// 只能消耗除宝石之外的物品
func (c *CostGroup) AddItemByBagId(account *Account, id uint32, count uint32) bool {
	idx_in_group := c.getItemIdxById(id)
	if bag.IsFixedID(id) {
		//看看是不是已经加了一部分了
		//道具，非武器这样类型同一类不会重的
		has_count := account.BagProfile.GetCountByBagId(id)
		// 注意和下面代码的区别，这两个功能面向的需求不一样
		// AddItemByBagId会特指背包中的一件道具，既可以是武器也可以是材料
		// AddItemById是消耗一种类型中的几个道具，目前只能消耗材料

		if idx_in_group < 0 {
			// 没加过，新加一下
			if count > has_count {
				return false
			}
			c.addItem(id, count)
		} else {
			if (count + c.count[idx_in_group]) > has_count {
				return false
			}
			c.count[idx_in_group] += count
		}

	} else {
		// 装备武器 count 认为为1
		if idx_in_group < 0 {
			c.addItem(id, 1)
		} else {
			return false
		}
	}

	return true
}

func (c *CostGroup) AddJadeByBagId(account *Account, id uint32, count uint32) bool {
	j := account.Profile.GetJadeBag().GetJade(id)
	if j == nil {
		return false
	}
	idx_in_group := c.getJadeIdxById(id)
	if idx_in_group < 0 {
		// 没加过，新加一下
		if int64(count) > j.CountInBag() {
			return false
		}
		c.addJade(id, count)
	} else {
		if int64(count+c.jcount[idx_in_group]) > j.CountInBag() {
			return false
		}
		c.jcount[idx_in_group] += count
	}
	return true
}

func (c *CostGroup) AddFashionByBagId(account *Account, id uint32) bool {
	has := account.Profile.GetFashionBag().HasFashionByBagId(id)
	idx_in_group := c.getFashionIdxById(id)
	if idx_in_group < 0 && has {
		c.addFashion(id)
		return true
	}
	return false
}

func (c *CostGroup) AddItemById(account *Account, item_id string, count uint32) bool {
	// 根据item_id
	// 将account的实际所有的item加进表中，如果没有多余的就返回false
	// 注意类似weapon类型的item，每个uint32的id只对应一个，
	// 所以要先看看这个id是不是已经加上了
	bag_id, is_fixed := gamedata.GetFixedBagID(item_id)
	if is_fixed {

		//道具，非武器这样类型同一类不会重的
		has_count := account.BagProfile.GetCount(item_id)

		//看看是不是已经加了一部分了
		idx_in_group := c.getItemIdxById(bag_id)
		if idx_in_group < 0 {
			// 没加过，新加一下
			if count > has_count {
				return false
			}
			c.addItem(bag_id, count)
		} else {
			if (count + c.count[idx_in_group]) > has_count {
				return false
			}
			c.count[idx_in_group] += count
		}
	} else {
		//武器类型的，注意这种同一类型有多个
		// TBD 这种类型暂时不这样自动加
		// 毕竟武器之间很不一样，让一个自动的系统消耗太不合适了
		// 另一方面暂时没有这种按名字消耗武器的逻辑
		return false
	}

	return true
}

func (c *CostGroup) addItem(id uint32, count uint32) {
	c.items = append(c.items, id)
	c.count = append(c.count, count)
}

func (c *CostGroup) addJade(id uint32, count uint32) {
	c.jades = append(c.jades, id)
	c.jcount = append(c.jcount, count)
}

func (c *CostGroup) addFashion(id uint32) {
	c.fashions = append(c.fashions, id)
}

func (c *CostGroup) getItemIdxById(id uint32) int {
	for idx, item_id := range c.items {
		if item_id == id {
			return idx
		}
	}
	return -1
}

func (c *CostGroup) getJadeIdxById(id uint32) int {
	for idx, jade_id := range c.jades {
		if jade_id == id {
			return idx
		}
	}
	return -1
}

func (c *CostGroup) getFashionIdxById(id uint32) int {
	for idx, fashion_id := range c.fashions {
		if fashion_id == id {
			return idx
		}
	}
	return -1
}

func (c *CostGroup) CostBySync(p *Account, sync helper.ISyncRsp, reason string) bool {
	// 这里返回的是是否全部cost完成，考虑
	acid := p.AccountID.String()
	logs.Trace("[%s]Cost by %s:%v", acid, reason, c)

	// 消耗软通
	for sc_t, sc_v := range c.sc {
		if !p.Profile.GetSC().UseSC(sc_t, sc_v, reason) {
			logs.SentryLogicCritical(acid, "NoScToCost:%d,%d", sc_t, sc_v)
			return false
		} else {
			if sync != nil {
				sync.OnChangeSC()
			}
		}
	}

	// 消耗道具
	for idx, id := range c.items {
		count := c.count[idx]
		res, isRemove, itemId, oldCount := p.BagProfile.UseByID(acid, id, count)
		if !res {
			logs.SentryLogicCritical(acid, "NoItemToCost:%d,%d", id, count)
			return false
		} else {
			if sync != nil {
				if isRemove {
					sync.OnChangeDelItems(helper.Item_Inner_Type_Basic, id, itemId, int64(oldCount), reason)
				} else {
					sync.OnChangeUpdateItems(helper.Item_Inner_Type_Basic, id, int64(oldCount), reason)
				}
			}
		}
	}

	// 消耗宝石
	for idx, id := range c.jades {
		count := c.jcount[idx]
		res, isRemove, itemId, oldCount := p.Profile.GetJadeBag().RemoveJade(id, int64(count))
		if !res {
			logs.SentryLogicCritical(acid, "NoJadeToCost:%d,%d", id, count)
			return false
		} else {
			if sync != nil {
				if isRemove {
					sync.OnChangeDelItems(helper.Item_Inner_Type_Jade, id, itemId, int64(oldCount), reason)
				} else {
					sync.OnChangeUpdateItems(helper.Item_Inner_Type_Jade, id, int64(oldCount), reason)
				}
			}
		}
	}

	// 消耗时装
	for _, id := range c.fashions {
		ok, itemId := p.Profile.GetFashionBag().RemoveFashion(id)
		if !ok {
			logs.SentryLogicCritical(acid, "NoFashionToCost:%d", id)
			return false
		} else {
			if sync != nil {
				sync.OnChangeDelItems(helper.Item_Inner_Type_Fashion, id, itemId, 1, reason)
			}
		}
	}

	if c.hc > 0 {
		// 一般的功能按照通常先消耗赠送钻，同时不需要知道到底消耗的那些类型的钻
		if !p.Profile.GetHC().UseHcGiveFirst(acid, c.hc, p.Profile.GetProfileNowTime(), reason) {
			logs.SentryLogicCritical(acid, "NoHcToCost:%d", c.hc)
			return false
		} else {
			if sync != nil {
				sync.OnChangeHC()
			}
		}
	}

	if c.energy > 0 {
		if !p.Profile.GetEnergy().Use(acid, reason, c.energy) {
			logs.SentryLogicCritical(acid, "NoEnergyToCost:%d", c.energy)
			return false
		} else {
			if sync != nil {
				sync.OnChangeEnergy()
			}
		}
	}

	if c.wheelcoin > 0 {
		if !p.Profile.GetWheelGachaInfo().Use(acid, reason, c.wheelcoin) {
			logs.SentryLogicCritical(acid, "NoWheelCoinToCost:%d", c.wheelcoin)
			return false
		} else {
			if sync != nil {
				sync.OnChangeWheel()
			}
		}
	}

	if c.bossFightPoint > 0 {
		if !p.Profile.GetBossFightPoint().Use(acid, reason, c.bossFightPoint) {
			logs.SentryLogicCritical(acid, "NoBossFightPointToCost:%d", c.bossFightPoint)
			return false
		}
		if sync != nil {
			sync.OnChangeBossFightPoint()
		}
	}
	//消耗主将碎片
	hero := p.Profile.GetHero()
	for sc_t, sc_v := range c.HeroStarPiece {
		if !hero.Remove(p, sc_t, sc_v, reason) {
			logs.SentryLogicCritical(acid, "NoHeroPieceToCost:%d,%d", sc_t, sc_v)
			return false
		} else {
			if sync != nil {
				sync.OnChangeHC()
			}
		}
	}
	return true
}

// 如果有Hc消耗时，有些地方需要制定Hc消耗方式和最终消耗的Hc类型，就用这个接口
func (c *CostGroup) CostWithHCBySync(p *Account, hc_typ_first int, sync helper.ISyncRsp, reason string) (cost_ok bool, hc_typ int) {
	// 这里返回的是是否全部cost完成，考虑
	logs.Trace("[%s]Cost by %s:%v", p.AccountID, c)

	acid := p.AccountID.String()

	hc_typ = gamedata.HC_From_Buy

	// 消耗软通
	for sc_t, sc_v := range c.sc {
		if !p.Profile.GetSC().UseSC(sc_t, sc_v, reason) {
			logs.SentryLogicCritical(acid, "NoScToCost:%d,%d", sc_t, sc_v)
			cost_ok = false
			return
		}
		sync.OnChangeSC()
	}

	// 消耗道具
	for idx, id := range c.items {
		count := c.count[idx]
		res, isRemove, itemId, oldCount := p.BagProfile.UseByID(acid, id, count)
		if !res {
			logs.SentryLogicCritical(acid, "NoItemToCost:%d,%d", id, count)
			cost_ok = false
			return
		}
		if isRemove {
			sync.OnChangeDelItems(helper.Item_Inner_Type_Basic, id, itemId, int64(oldCount), reason)
		} else {
			sync.OnChangeUpdateItems(helper.Item_Inner_Type_Basic, id, int64(oldCount), reason)
		}
	}

	// 消耗宝石
	for idx, id := range c.jades {
		count := c.count[idx]
		res, isRemove, itemId, oldCount := p.Profile.GetJadeBag().RemoveJade(id, int64(count))
		if !res {
			logs.SentryLogicCritical(acid, "NoJadeToCost:%d,%d", id, count)
			cost_ok = false
			return
		} else {
			if isRemove {
				sync.OnChangeDelItems(helper.Item_Inner_Type_Jade, id, itemId, int64(oldCount), reason)
			} else {
				sync.OnChangeUpdateItems(helper.Item_Inner_Type_Jade, id, int64(oldCount), reason)
			}
		}
	}

	// 消耗时装
	for _, id := range c.fashions {
		ok, itemId := p.Profile.GetFashionBag().RemoveFashion(id)
		if !ok {
			logs.SentryLogicCritical(acid, "NoFashionToCost:%d", id)
			cost_ok = false
			return
		} else {
			sync.OnChangeDelItems(helper.Item_Inner_Type_Fashion, id, itemId, 1, reason)
		}
	}

	if c.hc > 0 {
		hc_typ = gamedata.HC_From_Compensate
		// 特殊的功能需要按照一定的规则消耗，同时需要知道到底消耗的那些类型的钻
		ok, typ := p.Profile.GetHC().UseHc(acid, hc_typ_first, c.hc, p.Profile.GetProfileNowTime(), reason)
		if !ok {
			logs.SentryLogicCritical(acid, "NoHcToCost:%d", c.hc)
			cost_ok = false
			return
		}
		hc_typ = typ
		sync.OnChangeHC()
	}

	if c.energy > 0 {
		if !p.Profile.GetEnergy().Use(acid, reason, c.energy) {
			logs.SentryLogicCritical(acid, "NoEnergyToCost:%d", c.energy)
			cost_ok = false
			return
		}
		sync.OnChangeEnergy()
	}

	if c.wheelcoin > 0 {
		if !p.Profile.GetWheelGachaInfo().Use(acid, reason, c.wheelcoin) {
			logs.SentryLogicCritical(acid, "NoWheelCoinToCost:%d", c.wheelcoin)
			cost_ok = false
			return
		}
		sync.OnChangeWheel()
	}

	if c.bossFightPoint > 0 {
		if !p.Profile.GetBossFightPoint().Use(acid, reason, c.bossFightPoint) {
			logs.SentryLogicCritical(acid, "NoBossFightPointToCost:%d", c.bossFightPoint)
			cost_ok = false
			return
		}
		sync.OnChangeBossFightPoint()
	}

	cost_ok = true
	return
}

// 由于对于一个玩家来说所有情求可被视为同步的所以这一步没必要
// 只要之前检测是成功的，除非有逻辑要再CostGroup构建过程中消耗物品
// 这个逻辑才是有用的
func (c *CostGroup) Has(p *Account) bool {
	return true
}
