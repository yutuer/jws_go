package store

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

type Shop struct {
	ShopTyp  uint32 `json:"st"`
	Goods    []Good `json:"gs"`
	LastTime int64  `json:"lt"`
}

type Good struct {
	GoodId   string `json:"g"`
	UseTimes int    `json:"t"`
}

func (shop *Shop) ShopRefresh(now_time int64) bool {
	shop.initGoods()
	isRef := false
	isSameDay := gamedata.IsSameDayCommon(now_time, shop.LastTime)
	for i, good := range shop.Goods {
		if !gamedata.IsGoodInShop(shop.ShopTyp, good.GoodId) {
			continue
		}
		goodCfg := gamedata.GetGoodCfg(good.GoodId)
		if goodCfg == nil {
			continue
		}
		if gamedata.IsGoodDailyRefresh(goodCfg) && !isSameDay {
			shop.Goods[i].UseTimes = 0
			isRef = true
		}
	}
	if !isSameDay {
		shop.LastTime = now_time
		isRef = true
	}
	return isRef
}

func (shop *Shop) GetGoodUseTimes(good string) int {
	shop.initGoods()
	for _, g := range shop.Goods {
		if g.GoodId == good {
			return g.UseTimes
		}
	}
	return 0
}

func (shop *Shop) AddGoodUseTimes(good string, count int) bool {
	if !gamedata.IsGoodInShop(shop.ShopTyp, good) {
		return false
	}
	shop.initGoods()
	for i, g := range shop.Goods {
		if g.GoodId == good {
			shop.Goods[i].UseTimes = g.UseTimes + count
			return true
		}
	}
	shop.Goods = append(shop.Goods, Good{GoodId: good, UseTimes: count})
	return true
}

func (shop *Shop) initGoods() {
	if shop.Goods == nil {
		shop.Goods = make([]Good, 0, 10)
	}
}
