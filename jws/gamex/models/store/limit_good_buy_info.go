package store

// 限时商城购买信息
// 有两个因素导致这样的设计
// 1 之前只有已经购买的，并且已经上线
// 2 go1.6并不支持map的序列话
// 相同的index分别表示ID,count
type LimitShopBuyInfo struct {
	HasBoughtGoods []int `json:"buy_limit_good"`       // 购买的物品ID
	HasBoughtCount []int `json:"buy_limit_good_count"` // 购买的物品数量
}

// 旧版本->新版本 保证物品和数量数据一致
func (info *LimitShopBuyInfo) CheckConsistence() {
	if info.HasBoughtGoods != nil && info.HasBoughtCount == nil {
		info.HasBoughtCount = make([]int, len(info.HasBoughtGoods))
		for i := range info.HasBoughtCount {
			info.HasBoughtCount[i] = 1
		}
	}
}
