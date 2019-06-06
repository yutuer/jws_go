package buy

import (
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil/count"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//
// 类似于体力购买这样的购买机制
//   - 1. 花费消耗购买一样东西/执行一个操作
//   - 2. 每天有购买次数限制
//   - 3. 每次花费随着次数不同而不同
//
//

type buyData struct {
	BuyCount count.CountData `json:"c"`
	IsInited bool            `json:"i"`
	Param    string          `json:"pm"`
}

type buyDataToClient struct {
	Count    int    `codec:"c"`
	Limit    int    `codec:"l"`
	PriceTyp string `codec:"pt"`
	PriceNum uint32 `codec:"pn"`
}

type buyFunc_InitData func(data *buyData)                                // 初始化数据
type buyFunc_Get func(data *buyData, nowT int64) *gamedata.PriceData     // 获取商品 / 获取购买次数上限 / 获取消耗
type buyFunc_Cost func(data *buyData, nowT int64) *gamedata.PriceData    // 获取商品 / 获取购买次数上限 / 获取消耗
type buyFunc_GetLimit func(data *buyData, nowT int64, vip_lv uint32) int //  获取购买次数上限

type buyOpt struct {
	op_init      buyFunc_InitData
	op_get_give  buyFunc_Get
	op_get_limit buyFunc_GetLimit
	op_get_cost  buyFunc_Cost
}

var buyOpts [Buy_Typ_Count]buyOpt

func regBuyOpt(typ int,
	op_init buyFunc_InitData,
	op_get_give buyFunc_Get,
	op_get_limit buyFunc_GetLimit,
	op_get_cost buyFunc_Cost) {
	if typ < 0 || typ >= len(buyOpts) {
		logs.Error("regBuyOpt Error by Typ %d", typ)
		return
	}
	buyOpts[typ] = buyOpt{
		op_init:      op_init,
		op_get_give:  op_get_give,
		op_get_limit: op_get_limit,
		op_get_cost:  op_get_cost,
	}
	return
}

func mkGetCostFromBuyData(typ int) buyFunc_Cost {
	return func(data *buyData, nowT int64) *gamedata.PriceData {
		return gamedata.GetBuyCfg(
			typ,
			data.BuyCount.Get(nowT))
	}
}

func init() {

	// 购买体力
	regBuyOpt(Buy_Typ_EnergyBuy,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(
				gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			cfg := gamedata.GetCommonCfg()
			if cfg == nil {
				return nil
			}
			give := &gamedata.PriceData{}
			give.AddItem(gamedata.VI_EN, cfg.GetEnergyUnit())
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			if cfg == nil {
				return -1
			}
			return int(cfg.EnergyPurchaseLimit)
		}, mkGetCostFromBuyData(Buy_Typ_EnergyBuy))

	// 购买世界boss精力
	regBuyOpt(Buy_Typ_BossFightPoint,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			cfg := gamedata.GetCommonCfg()
			if cfg == nil {
				return nil
			}
			give := &gamedata.PriceData{}
			give.AddItem(gamedata.VI_BossFightPoint, cfg.GetSprintUnit())
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			if cfg == nil {
				return -1
			}
			return int(cfg.SprintPurchaseLimit)
		}, mkGetCostFromBuyData(Buy_Typ_BossFightPoint))

	// 购买SC
	regBuyOpt(Buy_Typ_SC,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			cfg := gamedata.GetCommonCfg()
			if cfg == nil {
				return nil
			}
			give := &gamedata.PriceData{}
			give.AddItem(gamedata.VI_Sc0, cfg.GetSCUnit())
			give.Cost.SetGST(gamedata.GST_GoldBonus)
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			if cfg == nil {
				return -1
			}
			return int(cfg.SCPurchaseLimit)
		}, mkGetCostFromBuyData(Buy_Typ_SC))

	// 购买精英关次数
	regBuyOpt(Buy_Typ_EliteTimes,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			stage_data := gamedata.GetStageData(data.Param)
			give := &gamedata.PriceData{}
			give.PriceTyp = "StageTimes"
			give.PriceCount = 1
			give.Cost.AddEStageTimes(data.Param,
				uint32(stage_data.MaxDailyAccess))
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			if cfg == nil {
				return -1
			}
			stage_data := gamedata.GetStageData(data.Param)
			if stage_data.Type == gamedata.LEVEL_TYPE_ELITE {
				return int(cfg.EliteStagePurchase)
			} else if stage_data.Type == gamedata.LEVEL_TYPE_HELL {
				return int(cfg.HellStagePurchase)
			}
			return 0
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			res := gamedata.GetStageTimesBuyCfg(
				data.Param,
				data.BuyCount.Get(nowT))
			if res == nil {
				logs.Error("GetStageTimesBuyCfg Nil By data.Param %v",
					data.Param)
			}
			return res
		})

	// 购买teampvp次数
	regBuyOpt(Buy_Typ_TeamPvp,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			give := &gamedata.PriceData{}
			give.PriceTyp = "TeamPvpTimes"
			give.PriceCount = 1
			give.Cost.AddGameModeTimes(counter.CounterTypeTeamPvp, 1)
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			if cfg == nil {
				return -1
			}
			return int(cfg.TPVPTimeLimit)
		}, mkGetCostFromBuyData(Buy_Typ_TeamPvp))

	// 购买simplepvp次数
	regBuyOpt(Buy_Typ_SimplePvp,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			give := &gamedata.PriceData{}
			give.PriceTyp = "SimplePvpTimes"
			give.PriceCount = 1
			give.Cost.AddGameModeTimes(counter.CounterTypeSimplePvp, 1)
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			if cfg == nil {
				return -1
			}
			return int(cfg.SimplePVPTimeLimit)
		}, mkGetCostFromBuyData(Buy_Typ_SimplePvp))

	// 购买军团普通次数
	regBuyOpt(Buy_Typ_GuildBossCount,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypGuildBoss)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			give := &gamedata.PriceData{}
			give.PriceTyp = "GuildBossNormalCount"
			give.PriceCount = 1
			give.Cost.AddGameModeTimes(counter.CounterTypeFreeGuildBoss, 1)
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetGameModeControlData(counter.CounterTypeGuildBossBuyTime)
			return cfg.GetCount
		}, mkGetCostFromBuyData(Buy_Typ_GuildBossCount))

	// 购买军团最终次数
	regBuyOpt(Buy_Typ_GuildBigBossCount,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypGuildBoss)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			give := &gamedata.PriceData{}
			give.PriceTyp = "GuildBigBossCount"
			give.PriceCount = 1
			give.Cost.AddGameModeTimes(counter.CounterTypeFreeGuildBigBoss, 1)
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetGameModeControlData(counter.CounterTypeGuildBigBossBuyTime)
			return cfg.GetCount
		}, mkGetCostFromBuyData(Buy_Typ_GuildBigBossCount))

	// 购买天赋点
	regBuyOpt(Buy_Typ_HeroTalentPoint,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			give := &gamedata.PriceData{}
			give.PriceTyp = "HeroTalentPoint"
			give.PriceCount = 1
			give.Cost.AddHeroTalentPoint(1)
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			return -1
		}, mkGetCostFromBuyData(Buy_Typ_HeroTalentPoint))

	// 购买包子
	regBuyOpt(Buy_Typ_BaoZi,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(
				gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			cfg := gamedata.GetCommonCfg()
			if cfg == nil {
				return nil
			}
			give := &gamedata.PriceData{}
			give.AddItem(gamedata.VI_BaoZi, cfg.GetBaoZiUnit())
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			if cfg == nil {
				return -1
			}
			return int(cfg.BaoZiPurchaseLimit)
		}, mkGetCostFromBuyData(Buy_Typ_BaoZi))
	// 购买节日Boss次数
	regBuyOpt(BUy_Typ_FestivalBossCount,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			give := &gamedata.PriceData{}
			give.PriceTyp = "FestivalBossCount"
			give.PriceCount = 1
			give.Cost.AddGameModeTimes(counter.CounterTypeFestivalBoss, 1)
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			maxCount := gamedata.GetFestivalBossMaxBuy()
			if cfg == nil {
				return -1
			}
			return maxCount
		}, mkGetCostFromBuyData(BUy_Typ_FestivalBossCount))
	// 购买无双争霸刷新次数
	regBuyOpt(Buy_Typ_WSPVP_Refresh,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			give := &gamedata.PriceData{}
			give.PriceTyp = "WSPVPRefreshCount"
			give.PriceCount = 1
			give.Cost.AddGameModeTimes(counter.CounterTypeWspvpRefresh, 1)
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			maxCount := gamedata.GetWSPVPRefreshMaxBuy()
			if cfg == nil {
				return -1
			}
			return maxCount
		}, mkGetCostFromBuyData(Buy_Typ_WSPVP_Refresh))
	regBuyOpt(Buy_Typ_WSPVP_Challenge,
		func(data *buyData) {
			data.BuyCount = count.NewDailyClear(gamedata.DailyStartTypCommon)
			data.IsInited = true
		}, func(data *buyData, nowT int64) *gamedata.PriceData {
			give := &gamedata.PriceData{}
			give.PriceTyp = "WSPVPChallengeCount"
			give.PriceCount = 1
			give.Cost.AddGameModeTimes(counter.CounterTypeWspvpChallenge, 1)
			return give
		}, func(data *buyData, nowT int64, vip_lv uint32) int {
			cfg := gamedata.GetVIPCfg(int(vip_lv))
			//maxCount := gamedata.GetWSPVPTimeMaxBuy()
			if cfg == nil {
				return -1
			}
			return int(cfg.WsChallengeLimit)
		}, mkGetCostFromBuyData(Buy_Typ_WSPVP_Challenge))
}

type PlayerBuy struct {
	BuyDatas    [Buy_Typ_Count]buyData `json:"d"`
	EStageTimes []buyData              `json:"stage_d"`

	estageTimes map[string]*buyData
}

func (p *PlayerBuy) OnAfterLogin() {
	stagePolicy := gamedata.StagesPurchasePolicy()
	if p.EStageTimes == nil {
		p.EStageTimes = make([]buyData, len(stagePolicy))
		i := 0
		for stageId, _ := range stagePolicy {
			p.EStageTimes[i].Param = stageId
			i++
		}
	} else if len(p.EStageTimes) < len(stagePolicy) {
		tmp := make(map[string]struct{}, len(p.EStageTimes))
		for _, ot := range p.EStageTimes {
			tmp[ot.Param] = struct{}{}
		}
		old := p.EStageTimes
		p.EStageTimes = make([]buyData, 0, len(stagePolicy))
		p.EStageTimes = append(p.EStageTimes, old...)
		for stageId, _ := range stagePolicy {
			if _, ok := tmp[stageId]; !ok {
				p.EStageTimes = append(p.EStageTimes, buyData{
					Param: stageId,
				})
			}
		}
	}

	p.estageTimes = make(map[string]*buyData, len(p.EStageTimes))
	for i := 0; i < len(p.EStageTimes); i++ {
		t := &p.EStageTimes[i]
		p.estageTimes[t.Param] = t
	}
}

func (p *PlayerBuy) DebugResetCount() {
	for i := 0; i < len(p.BuyDatas); i++ {
		p.BuyDatas[i].BuyCount.Count = 0
	}
	for i := 0; i < len(p.EStageTimes); i++ {
		p.EStageTimes[i].BuyCount.Count = 0
	}
}

func (p *PlayerBuy) checkTypThenInit(typ int, param1 string) *buyOpt {
	if typ < 0 || typ >= len(buyOpts) {
		logs.Error("checkTypThenInit Err by typ %d", typ)
		return nil
	}

	opt := &buyOpts[typ]
	bd := p.getBuyData(typ, param1)
	if bd == nil {
		logs.Error("checkTypThenInit Err by typ %d param %s", typ, param1)
		return nil
	}
	if !bd.IsInited {
		opt.op_init(bd)
	}
	return opt
}

func (p *PlayerBuy) GetInfo(typ int, vip_lv uint32,
	time_now int64, param1 string) (
	ok bool,
	price, give *gamedata.PriceData,
	curr_count, curr_limit int) {
	opt := p.checkTypThenInit(typ, param1)
	if opt == nil {
		ok = false
		return
	}

	data := p.getBuyData(typ, param1)
	price = opt.op_get_cost(data, time_now)
	give = opt.op_get_give(data, time_now)
	curr_count = data.BuyCount.Get(time_now)
	curr_limit = opt.op_get_limit(data, time_now, vip_lv)
	ok = true
	return
}

func (p *PlayerBuy) GetInfoToClient(typ int, vip_lv uint32, time_now int64, realLimit int) buyDataToClient {
	if typ == Buy_Typ_EliteTimes {
		return buyDataToClient{}
	}
	ok, price, _, c, l := p.GetInfo(typ, vip_lv, time_now, "")
	if !ok || price == nil {
		return buyDataToClient{}
	}
	if realLimit != -10 {
		l = realLimit
	}
	return buyDataToClient{
		Count:    c,
		Limit:    l,
		PriceTyp: price.PriceTyp,
		PriceNum: price.PriceCount,
	}
}

func (p *PlayerBuy) GetStageTimesInfoToClient(vip_lv uint32, time_now int64) map[string]buyDataToClient {
	res := make(map[string]buyDataToClient, len(p.estageTimes))
	for stageId, _ := range p.estageTimes {
		ok, price, _, c, l := p.GetInfo(Buy_Typ_EliteTimes, vip_lv, time_now, stageId)
		if !ok || price == nil {
			res[stageId] = buyDataToClient{}
		} else {
			res[stageId] = buyDataToClient{
				Count:    c,
				Limit:    l,
				PriceTyp: price.PriceTyp,
				PriceNum: price.PriceCount,
			}
		}
	}
	return res
}

func (p *PlayerBuy) OnBuy(typ int, time_now int64, param1 string) {
	opt := p.checkTypThenInit(typ, param1)
	if opt == nil {
		return
	}
	bd := p.getBuyData(typ, param1)
	bd.BuyCount.Add(time_now, 1)
}

func (p *PlayerBuy) getBuyData(typ int, param string) *buyData {
	if typ == Buy_Typ_EliteTimes {
		return p.estageTimes[param]
	} else {
		return &p.BuyDatas[typ]
	}
}
