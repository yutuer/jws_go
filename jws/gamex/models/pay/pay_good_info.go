package pay

import (
	"strconv"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/api_gateway/pay"
)

const (
	vn2diamond = 120.0 / 20000
)

type PayGoodInfo struct {
	GoodIdx         uint32 `json:"gidx"`
	FirstGiveSerial int32  `json:"s"`
}
type OrderRec struct {
	Money     uint32 `json:"my"`
	Idx       uint32 `json:"idx"`
	Order     string `json:"oder"`
	TimeStamp int64  `json:"ts"` // 客户端支付的时间戳
	Platflorm string `json:"plfm"`
	PayType   string `json:"payty"`
}

type PayGoodInfos struct {
	IAPOrderSeed int64         `json:"iapseed"` // iap订单seed
	Infos        []PayGoodInfo `json:"paygood"`
	CurrenSbatch int64         `json:"curren_sbatch"` //当前服务器批次
	RecentOrders []OrderRec    `json:"trecs"`
	MoneySum     uint32        `json:"money_sum"`

	MonthlyCardEndTime  int64 `json:"month_end"`
	MonthlyValidTime    int64 `json:"month_v_t"` // 本次有效时间
	IsLifeCard          bool  `json:"life_v"`
	LifeCardValidTime   int64 `json:"life_v_t"` // 本次有效时间
	WeekRewardEndTime   int64 `json:"week_end"`
	WeekRewardValidTime int64 `json:"week_v_t"` // 本次有效时间

	LevelGiftId          string   `json:"lvl_gift_id"`             // 当前可用礼包id
	LevelGiftEndTime     int64    `json:"lvl_gift_e_t"`            // 当前可用礼包过期时间
	LevelGiftIdWaitAward []string `json:"lvl_gift_id_wait_award_"` // 当前可用礼包已买,等待领奖
	BoughtLevelGiftId    []string `json:"bgt_lvl_gift_id"`         // 当前已买的等级礼包,用于防止多次购买

	DataBuild int `json:"data_build"`

	helper.SyncObj
}

type PayGoodInfoToClient struct {
	GoodIdx         uint32 `codec:"gidx"`
	FirstGiveSerial int32  `codec:"fs"`
}

// 支付成功，记录首次购买状态，并返回应该给的钻石数
func (pg *PayGoodInfos) OnPayGoodSuccess(shardid uint, goodIdx uint32,
	platform string, now_time int64, trueAmount uint32, channel string) (hcBuy, hcGive uint32, goodName string) {
	cfg := gamedata.GetIAPInfo(goodIdx)
	if goodIdx == 0 && (channel == pay.VNAndroidChannel || channel == pay.VNIOSChannel) {
		logs.Debug("condition0: %d", uint32(float32(trueAmount)*float32(vn2diamond)))
		return uint32(float32(trueAmount) * float32(vn2diamond)), 0, ""
	}
	logs.Debug("on pay good success: %v, %v, %v, %v, %v, %v", shardid, goodIdx, trueAmount, channel, cfg)
	if cfg == nil {
		logs.Debug("pay good idx %d not found", goodIdx)
		return 0, 0, ""
	}
	var extraPurchase uint32
	if channel == pay.VNAndroidChannel && trueAmount >= cfg.Android_Rmb_Price {
		extraPurchase = uint32(float32(trueAmount-cfg.Android_Rmb_Price) * float32(vn2diamond))
	} else if channel == pay.VNIOSChannel && trueAmount >= cfg.IOS_Rmb_Price {
		extraPurchase = uint32(float32(trueAmount-cfg.IOS_Rmb_Price) * float32(vn2diamond))
	}
	firstAdditionGive := cfg.Info.GetFirstAdditionalGive()
	additionGive := cfg.Info.GetAdditionalGive()
	var purchased = cfg.Info.GetPurchase()
	if channel == pay.VNAndroidChannel && trueAmount < cfg.Android_Rmb_Price {
		firstAdditionGive = 0
		purchased = 0
		additionGive = 0
	} else if channel == pay.VNIOSChannel && trueAmount < cfg.IOS_Rmb_Price {
		firstAdditionGive = 0
		purchased = 0
		additionGive = 0
	}
	logs.Debug("extraPurchase: %v, purchased: %v, firstAddtionGive: %v", extraPurchase, purchased, firstAdditionGive)
	if (channel == pay.VNAndroidChannel && cfg.Android_Rmb_Price <= trueAmount) ||
		(channel == pay.VNIOSChannel && cfg.IOS_Rmb_Price <= trueAmount) ||
		(channel != pay.VNIOSChannel && channel != pay.VNAndroidChannel) {
		// 月卡，终身卡
		if !pg._card(goodIdx, platform, now_time) {
			logs.Error("pay good card idx failed %d ", goodIdx)
			return 0, 0, ""
		}
		// 等级礼包
		pg._levelGift(goodIdx, platform, now_time)
		// 没有首次购买奖励的
		if cfg.Info.GetFirstGiveSerial() <= 0 {
			logs.Debug("condition1 no first give serial")
			return purchased + extraPurchase, additionGive, cfg.Info.GetIapID()
		}
		// 是否已购买过，但首次购买批次变了

		sbatch := pg.GetCurrenServerSbatch(shardid, now_time)
		pg.CurrenSbatch = int64(sbatch)
		for i, v := range pg.Infos {
			if v.GoodIdx == goodIdx {
				if int32(sbatch) > v.FirstGiveSerial {
					pg.Infos[i].FirstGiveSerial = int32(sbatch)
					logs.Debug("condition2")
					return purchased + extraPurchase, firstAdditionGive, cfg.Info.GetIapID()
				}
				logs.Debug("condition3")
				return purchased + extraPurchase, additionGive, cfg.Info.GetIapID()
			}
		}
		// 还没首次购买过, 现在客户端购买成功就会给服务器发协议,所以这里应该不会走到了，但如果客户端不发支付成功协议，这里会走到
		pg.Infos = append(pg.Infos, PayGoodInfo{
			GoodIdx:         goodIdx,
			FirstGiveSerial: int32(sbatch),
		})
	}
	logs.Debug("condition4")
	return purchased + extraPurchase, firstAdditionGive, cfg.Info.GetIapID()
}

func (pg *PayGoodInfos) GetPayGoodInfoClient() []PayGoodInfoToClient {
	res := make([]PayGoodInfoToClient, len(pg.Infos))
	for i, v := range pg.Infos {
		res[i] = PayGoodInfoToClient{v.GoodIdx, v.FirstGiveSerial}
	}
	return res
}

func (pg *PayGoodInfos) GetCurrenServerSbatch(shardid uint, now_time int64) uint32 {
	var sbatch uint32
	cfgSbatch := gamedata.GetHotDatas().Activity.GetServerGroupSbatch(uint32(shardid))
	_ts, _ := time.ParseInLocation("20060102_15:04", cfgSbatch.GetEffectiveTime(), util.ServerTimeLocal)
	ts := _ts.Unix()
	if now_time > ts {
		sbatch = cfgSbatch.GetSbatch()
	} else {
		if cfgSbatch.GetSbatch() == 1 {
			sbatch = 1
		} else {
			sbatch = cfgSbatch.GetSbatch() - 1
		}
	}
	return sbatch
}

// 记录玩家的iap订单
func (pg *PayGoodInfos) AddRecOrder(now_time_s string, money, idx uint32,
	order, platform, payType string) {
	now_time, _ := strconv.ParseInt(now_time_s, 10, 64)
	pg.MoneySum += money
	pg.RecentOrders = append(pg.RecentOrders, OrderRec{
		Money:     money,
		Idx:       idx,
		Order:     order,
		TimeStamp: now_time,
		Platflorm: platform,
		PayType:   payType,
	})
}

func (pg *PayGoodInfos) InitGoodInfo() {
	if pg.Infos == nil {
		pg.Infos = make([]PayGoodInfo, 0, 64)
	}
	if pg.RecentOrders == nil {
		pg.RecentOrders = make([]OrderRec, 0, 16)
	}
	if pg.LevelGiftIdWaitAward == nil {
		pg.LevelGiftIdWaitAward = make([]string, 0, 16)
	}
	if pg.BoughtLevelGiftId == nil {
		pg.BoughtLevelGiftId = make([]string, 0, 16)
	}
}

func (pg *PayGoodInfos) _card(goodIdx uint32, platform string, now_time int64) bool {
	if !gamedata.IsIAPCardIdx(goodIdx) {
		goodIdx -= uutil.IAPID_ONESTORE_2_GOOGLE
		if !gamedata.IsIAPCardIdx(goodIdx) {
			return true
		}
	}
	buyMonthCard := true
	if platform == uutil.Android_Platform {
		switch goodIdx {
		case gamedata.IAPMonth.GetConditionValue1ForAndroid():
			buyMonthCard = true
		case gamedata.IAPLife.GetConditionValue1ForAndroid():
			buyMonthCard = false
		default:
			logs.Error("IAP buy card idx not found in android %d", goodIdx)
			return false
		}
	} else {
		switch goodIdx {
		case gamedata.IAPMonth.GetConditionValue1ForIOS():
			buyMonthCard = true
		case gamedata.IAPLife.GetConditionValue1ForIOS():
			buyMonthCard = false
		default:
			logs.Error("IAP buy card idx not found in ios %d", goodIdx)
			return false
		}
	}

	// 月卡或终身卡生效
	if buyMonthCard {
		logs.Debug("buy month card")
		if pg.MonthlyCardEndTime < now_time { // 第一次买月卡或已经过期
			pg.MonthlyValidTime = gamedata.GetCommonDayBeginSec(now_time)
			pg.MonthlyCardEndTime = pg.MonthlyValidTime +
				int64(gamedata.IAPMonth.GetGetDay())*int64(util.DaySec)
		} else { // 还有月卡生效着
			pg.MonthlyCardEndTime = pg.MonthlyCardEndTime +
				int64(gamedata.IAPMonth.GetGetDay())*int64(util.DaySec)
		}
	} else {
		logs.Debug("buy life card")
		if pg.IsLifeCard {
			logs.Error("IAP buy life card duplicate %d", goodIdx)
			return false
		}
		pg.IsLifeCard = true
		pg.LifeCardValidTime = gamedata.GetCommonDayBeginSec(now_time)
	}
	// 周奖励生效
	if pg.IsLifeCard && pg.MonthlyCardEndTime > now_time {
		if pg.WeekRewardEndTime < now_time {
			pg.WeekRewardValidTime = WeekStartCommonSecBaseMonday(now_time)
		}
		pg.WeekRewardEndTime = WeekStartCommonSecBaseMonday(pg.MonthlyCardEndTime) + int64(util.WeekSec)
	}
	logs.Debug("buy card sucess %v", pg)
	return true
}

func WeekStartCommonSecBaseMonday(now_time int64) int64 {
	dayBegin := gamedata.GetCommonDayBeginSec(now_time)
	w := util.GetWeek(dayBegin)
	if w == 0 {
		w = 7
	}
	return dayBegin - int64((w-1)*util.DaySec)
}

func (pg *PayGoodInfos) _levelGift(goodIdx uint32, platform string, now_time int64) {
	if pg.LevelGiftId == "" {
		return
	}
	var cfg *ProtobufGen.LEVELGIFTPURCHASE
	if platform == uutil.Android_Platform {
		cfg = gamedata.IapLevelGiftAndroid(goodIdx)
		if cfg == nil {
			goodIdx -= uutil.IAPID_ONESTORE_2_GOOGLE
			cfg = gamedata.IapLevelGiftAndroid(goodIdx)
		}
	} else {
		cfg = gamedata.IapLevelGiftIOS(goodIdx)
	}
	if cfg == nil {
		return
	}
	if pg.isBought(cfg) {
		logs.Error("PayGoodInfos _levelGift already boughtLevelGiftId %v %v", goodIdx, pg.BoughtLevelGiftId)
		return
	} else {
		pg.BoughtLevelGiftId = append(pg.BoughtLevelGiftId, cfg.GetLevelGiftID())
	}

	pg.LevelGiftIdWaitAward = append(pg.LevelGiftIdWaitAward, cfg.GetLevelGiftID())
	pg.LevelGiftId = ""
	pg.LevelGiftEndTime = 0
	logs.Debug("LevelGift buy iap %v %v ", pg.LevelGiftIdWaitAward, pg.BoughtLevelGiftId)
}

func (pg *PayGoodInfos) OnPlayerLevelUp(lvl uint32, now_t int64) {
	id := gamedata.LevelGiftOnLvlUp(lvl)
	if id == "" {
		return
	}
	cfg := gamedata.GetLevelGiftCfg(id)
	if pg.isBought(cfg) {
		return
	}
	if pg.LevelGiftId == id { // 防止本次礼包过期后，再升级再次弹出
		return
	}
	pg.LevelGiftId = id
	pg.LevelGiftEndTime = now_t + int64(cfg.GetTimeLimit())*util.HourSec
	pg.SyncObj.SetNeedSync()
	logs.Debug("LevelGift OnPlayerLevelUp %d %v %v", lvl, pg.LevelGiftId, pg.LevelGiftEndTime)
}

func (pg *PayGoodInfos) isBought(cfg *ProtobufGen.LEVELGIFTPURCHASE) bool {
	for _, v := range pg.BoughtLevelGiftId {
		if v == cfg.GetLevelGiftID() {
			return true
		}
	}
	return false
}
