package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type LuckyWheel struct {
	RemainHc int64    `json:"costhc"`
	HaveItem []string `json:"have_item"`
	CurNum   int64    `json:"cur_nums"`
	AllCoin  uint32   `json:"all_coin"`
	CostCoin uint32   `json:"cost_coin"`
	ActId    uint32   `json:"act_id"`
}

//清除数据
func (lw *LuckyWheel) ClearInfo() {
	lw.RemainHc = 0
	lw.CurNum = 0
	lw.CostCoin = 0
	lw.AllCoin = 0
	lw.HaveItem = make([]string, 0)
}

func (lw *LuckyWheel) Use(account, reason string, num int64) bool {
	if lw.Has(num) {
		logs.Debug("[%s] %s use wheel coin %d", reason, account, num)
		lw.CostCoin += uint32(num)
		return true
	} else {
		logs.Debug("dont have enough WheelCoins %d", num)
		return false
	}
}

func (lw *LuckyWheel) Has(coin int64) bool {
	if coin < 0 {
		logs.Error("Param num:%d is not right", coin)
		return false
	}
	data := gamedata.GetHotDatas().Activity.GetWheelSeting(lw.ActId)
	if data == nil {
		logs.Error("Can not find WheelSeting")
		return false
	}
	//确保不超过消耗上限
	if coin > int64(data.GetGetCoinItemMax()) {
		return false
	}

	if lw.CostCoin+uint32(coin) > lw.AllCoin {
		return false
	}
	return true
}

func (lw *LuckyWheel) GetCurCoin() int64 {
	if lw.AllCoin < lw.CostCoin {
		logs.Error("Profile WheelInfo Error:AllCoin less than CostCoin")
		return 0
	}
	logs.Debug("Have %d Coins now", int64(lw.AllCoin-lw.CostCoin))
	return int64(lw.AllCoin - lw.CostCoin)
}

func (lw *LuckyWheel) UpdateActId(id uint32) {
	lw.ActId = id
}

func (lw *LuckyWheel) InitInfo() {
	lw.ClearInfo()
}

//通过HC获得的代币
func (lw *LuckyWheel) UpdataCoinByHc(num int64) int64 {
	oldCoin := lw.AllCoin
	data := gamedata.GetHotDatas().Activity.GetWheelSeting(lw.ActId)
	if data == nil {
		logs.Error("Can not find WheelSeting")
		return 0
	}
	HcBase := int64(data.GetHCBase())
	GiveCoin := int64(data.GetGetCoinItem())
	limitCoin := int64(data.GetGetCoinItemMax())
	lw.RemainHc += num
	if lw.RemainHc >= HcBase {
		lw.AllCoin += uint32(lw.RemainHc / HcBase * GiveCoin)
		lw.RemainHc %= HcBase
	}
	if lw.AllCoin >= uint32(limitCoin) {
		lw.AllCoin = uint32(limitCoin)
	}
	return int64(lw.AllCoin - oldCoin)
}

func (lw *LuckyWheel) UpdataCoin(num uint32) bool {
	data := gamedata.GetHotDatas().Activity.GetWheelSeting(lw.ActId)
	if data == nil {
		logs.Error("Can not find WheelSeting")
		return false
	}
	limitCoin := data.GetGetCoinItemMax()
	lw.AllCoin += num
	if lw.AllCoin >= limitCoin || num >= limitCoin {
		lw.AllCoin = limitCoin
	}
	return true
}

func (lw *LuckyWheel) CoinFull(limitCoin uint32) bool {
	if lw.AllCoin == limitCoin {
		return true
	}
	return false
}

func (lw *LuckyWheel) GenItem() []string {
	return lw.HaveItem
}

func (lw *LuckyWheel) UpdateItem(name string) {
	lw.HaveItem = append(lw.HaveItem, name)
}

func (lw *LuckyWheel) UpdateUseWheel() {
	lw.CurNum += 1
	if lw.CurNum > 10 {
		logs.Error("Profile:Wheel use numbers out 10 times")
		lw.CurNum = 10
	}
}
