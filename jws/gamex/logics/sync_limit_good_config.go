package logics

import "vcs.taiyouxi.net/jws/gamex/models/gamedata"

// 热更限时商品表给客户端
func (s *SyncResp) mkLimitGoodConfig(p *Account) {
	itemLen := len(gamedata.GetHotDatas().LimitGoodConfig.Items)
	s.SyncLimitGoodConfigs = make([][]byte, itemLen)
	for i, item := range gamedata.GetHotDatas().LimitGoodConfig.GetAllLimitGoodForClient() {
		item.GoodItems = make([][]byte, len(item.GoodItemsArray))
		for j := 0; j < len(item.GoodItemsArray); j++ {
			item.GoodItems[j] = encode(item.GoodItemsArray[j])
		}
		s.SyncLimitGoodConfigs[i] = encode(item)
	}
}

func (s *SyncResp) mkLimitGoodBuy(p *Account) {
	if s.SyncBuyLimitShopNeed {
		p.StoreProfile.LimitShop.CheckConsistence()
		itemLen := len(p.StoreProfile.LimitShop.HasBoughtGoods)
		s.SyncBuyLimitShopInfo = make([]int, itemLen)
		s.SyncBuyLimitShopCount = make([]int, itemLen)
		for i, buyId := range p.StoreProfile.LimitShop.HasBoughtGoods {
			s.SyncBuyLimitShopInfo[i] = int(buyId)
			s.SyncBuyLimitShopCount[i] = int(p.StoreProfile.LimitShop.HasBoughtCount[i])
		}
	}
}
