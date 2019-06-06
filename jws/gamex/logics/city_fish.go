package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/city_fish"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type CityFish2Client struct {
	GlobalFishCount int `codec:"g_fish_c"`
}

const (
	TenTimes = 10
)

func (p *Account) CityFishTen(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncRespWithRewards
		FishNextRefTime     int64    `codec:"nreft"`
		FishRewardLeftCount []uint32 `codec:"rlc"`
	}{}

	initReqRsp(
		"PlayerAttr/CityFishTenResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Free_Times_Still_Have
		Err_Times_NotEnough
		Err_Hc_Not_Enough
		Err_Inner
		Err_Loot
		Err_Give
		Err_Jade_Full
	)

	// 次数检查
	if lc, nt := p.Profile.GetCounts().Get(counter.CounterTypeFish, p.Account); nt < 0 || lc > 0 {
		return rpcErrorWithMsg(resp, Err_Free_Times_Still_Have, "Err_Free_Times_Still_Have")
	}
	lc, nt := p.Profile.GetCounts().Get(counter.CounterTypeFishHC, p.Account)
	if nt < 0 || lc < TenTimes {
		return rpcErrorWithMsg(resp, Err_Times_NotEnough, "Err_Times_NotEnough when GetCounts")
	}
	// 背包检查
	if p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
		return rpcErrorWithMsg(resp, Err_Jade_Full, "Err_Jade_Full")
	}
	// 花hc
	fc := gamedata.FishCost()
	var hc_s uint32
	c := gamedata.GetGameModeControlData(counter.CounterTypeFishHC)
	ut := c.GetCount - lc + 1
	for i := 0; i < TenTimes; i++ {
		hc := fc.GetCostHC() + uint32(ut-1)*fc.GetAddHC()
		if hc > fc.GetLimit() {
			hc = fc.GetLimit()
		}
		hc_s += hc
		ut++
	}
	cost := &account.CostGroup{}
	if !cost.AddHc(p.Account, int64(hc_s)) || !cost.CostBySync(p.Account, resp, "CityFish") {
		return rpcErrorWithMsg(resp, Err_Hc_Not_Enough,
			fmt.Sprintf("Err_Hc_Not_Enough from %d, hc %d", ut, hc_s))
	}
	// 扣次数
	if !p.Profile.GetCounts().UseN(counter.CounterTypeFishHC, TenTimes, p.Account) {
		logs.Error("CityFishTen cost times failed")
	}

	// 抽奖，先大奖
	res := execCmd(city_fish.CityFish_Cmd_Award, p, 10)
	if res == nil {
		return rpcErrorWithMsg(resp, Err_Inner, "Err_Inner")
	}

	data := &gamedata.CostData{}
	resp.FishNextRefTime = res.NextRefTime
	resp.FishRewardLeftCount = res.GlobalRewardCount
	i2c := make(map[string]uint32, 10)
	for idx, _ := range res.AwardItem {
		item := res.AwardItem[idx]
		itemc := res.AwardCount[idx]
		data.AddItem(item, itemc)
		i2c[item] = i2c[item] + itemc
	}
	// 再个人奖
	for i := 0; i < res.AwardLeftCount; i++ {
		gives := p._personalAward()
		if gives == nil {
			return rpcErrorWithMsg(resp, Err_Loot, "Err_Loot")
		} else {
			data.AddGroup(gives.Gives())
		}
	}
	give := &account.GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(p.Account, resp, "CityFish") {
		return rpcErrorWithMsg(resp, Err_Give,
			fmt.Sprintf("Err_Give count %d item_id %d", res.AwardCount, res.AwardItem))
	}

	// 46.钓鱼次数，大于等于P1
	p.updateCondition(account.COND_TYP_FishTimes,
		0, 0, "", "", nil)

	resp.OnChangeGameMode(counter.CounterTypeFishHC)
	resp.mkInfo(p)
	// log
	logiclog.LogFish(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, res.AwardId, true, true,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}

func (p *Account) CityFish(r servers.Request) *servers.Response {
	req := &struct {
		Req
		IsCostHc bool `codec:"is_hc"`
	}{}
	resp := &struct {
		SyncRespWithRewards
		FishRid             uint32   `codec:"fish_rid"`
		FishNextRefTime     int64    `codec:"nreft"`
		FishRewardLeftCount []uint32 `codec:"rlc"`
	}{}

	initReqRsp(
		"PlayerAttr/CityFishResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_HcTimes_Not_Enough
		Err_Free_Times_Still_Has
		Err_Get_Loot_Cfg
		Err_Loot
		Err_Give
		Err_Cfg
		Err_Hc_Not_Enough
		Err_Inner
		Err_Jade_Full
	)

	// 背包检查
	if p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
		return rpcErrorWithMsg(resp, Err_Jade_Full, "Err_Jade_Full")
	}

	if !req.IsCostHc { // 免费的
		// 次数检查
		if !p.Profile.GetCounts().Use(counter.CounterTypeFish, p.Account) {
			logs.Error("CityFish but times not enough") // 防止连点情况，温柔处理，不返回错误
			return rpcSuccess(resp)
			//			return rpcErrorWithMsg(resp, Err_Times_Not_Enough, "Err_Times_Not_Enough")
		}
	} else { // 付费的
		if p.Profile.GetCounts().Has(counter.CounterTypeFish, p.Account) {
			return rpcErrorWithMsg(resp, Err_Free_Times_Still_Has, "Err_Free_Times_Still_Has")
		}
		if !p.Profile.GetCounts().Use(counter.CounterTypeFishHC, p.Account) {
			return rpcErrorWithMsg(resp, Err_HcTimes_Not_Enough, "Err_HcTimes_Not_Enough")
		}
		lc, nt := p.Profile.GetCounts().Get(counter.CounterTypeFishHC, p.Account)
		if nt < 0 {
			return rpcErrorWithMsg(resp, Err_Cfg, "Err_Cfg when GetCounts")
		}
		c := gamedata.GetGameModeControlData(counter.CounterTypeFishHC)
		ut := c.GetCount - lc
		// cost hc
		fc := gamedata.FishCost()
		hc := fc.GetCostHC() + uint32(ut-1)*fc.GetAddHC()
		if hc > fc.GetLimit() {
			hc = fc.GetLimit()
		}
		cost := &account.CostGroup{}
		if !cost.AddHc(p.Account, int64(hc)) || !cost.CostBySync(p.Account, resp, "CityFish") {
			return rpcErrorWithMsg(resp, Err_Hc_Not_Enough, "Err_Hc_Not_Enough")
		}
	}

	// 全服奖励里抽奖
	res := execCmd(city_fish.CityFish_Cmd_Award, p, 1)
	if res == nil {
		return rpcErrorWithMsg(resp, Err_Inner, "Err_Inner")
	}
	resp.FishNextRefTime = res.NextRefTime
	resp.FishRewardLeftCount = res.GlobalRewardCount
	data := &gamedata.CostData{}
	if len(res.AwardId) > 0 { // 有全服奖
		resp.FishRid = res.AwardId[0]
		for i := 0; i < len(res.AwardItem); i++ {
			item := res.AwardItem[i]
			itemc := res.AwardCount[i]
			data.AddItem(item, itemc)
		}
	} else { // 全服奖没有了, 给个人奖
		gives := p._personalAward()
		if gives == nil {
			return rpcErrorWithMsg(resp, Err_Loot, "Err_Loot")
		} else {
			data.AddGroup(gives.Gives())
		}
	}
	give := &account.GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(p.Account, resp, "CityFish") {
		return rpcErrorWithMsg(resp, Err_Give,
			fmt.Sprintf("Err_Give count %v item_id %v", res.AwardCount, res.AwardItem))
	}

	// 46.钓鱼次数，大于等于P1
	p.updateCondition(account.COND_TYP_FishTimes,
		0, 0, "", "", nil)

	resp.OnChangeGameMode(counter.CounterTypeFish)
	resp.OnChangeGameMode(counter.CounterTypeFishHC)
	resp.mkInfo(p)
	// log
	logiclog.LogFish(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, res.AwardId, false, req.IsCostHc,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}

func (p *Account) GetPlayerCityFishInfo(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		FishNextRefTime     int64    `codec:"nreft"`
		FishRewardLeftCount []uint32 `codec:"rlc"`
	}{}

	initReqRsp(
		"PlayerAttr/GetPCityFishInfoResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Inner
	)

	res := execCmd(city_fish.CityFish_Cmd_Get_Info, p, 0)
	if res == nil {
		return rpcErrorWithMsg(resp, Err_Inner, "Err_Inner")
	}
	resp.FishNextRefTime = res.NextRefTime
	resp.FishRewardLeftCount = res.GlobalRewardCount
	return rpcSuccess(resp)
}

func (p *Account) GetGlobalCityFishLog(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		Time  []int64  `codec:"t"`
		Names []string `codec:"nam"`
		Item  []string `codec:"item"`
		Count []uint32 `codec:"count"`
	}{}

	initReqRsp(
		"PlayerAttr/GetGCityFishLogResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Inner
	)

	res := execCmd(city_fish.CityFish_Cmd_Get_Record, p, 0)
	if res == nil {
		return rpcErrorWithMsg(resp, Err_Inner, "Err_Inner")
	}
	resp.Time = make([]int64, 0, len(res.Logs))
	resp.Names = make([]string, 0, len(res.Logs))
	resp.Item = make([]string, 0, len(res.Logs))
	resp.Count = make([]uint32, 0, len(res.Logs))
	for _, log := range res.Logs {
		resp.Time = append(resp.Time, log.Time)
		resp.Names = append(resp.Names, log.Name)
		resp.Item = append(resp.Item, log.Item)
		resp.Count = append(resp.Count, log.Count)
	}
	return rpcSuccess(resp)
}

func execCmd(typ int, p *Account, c int) *city_fish.FishRet {
	fc := city_fish.FishCmd{
		Typ:        typ,
		AName:      p.Profile.Name,
		ARand:      p.GetRand(),
		AwardCount: c,
	}
	res := city_fish.GetModule(p.AccountID.ShardId).CommandExec(fc)
	if res.Success {
		return res
	}
	return nil
}

func (p *Account) _personalAward() *gamedata.PriceDatas {
	fCfg := gamedata.FishCost()
	// 随机物品
	gives, err := gamedata.LootItemGroupRand(p.GetRand(), fCfg.GetLootDataID())
	if err != nil {
		return nil
	}

	if fCfg.GetItemID() != "" && fCfg.GetAmount() > 0 {
		gives.AddItem(fCfg.GetItemID(), fCfg.GetAmount())
	}
	// 加物品
	return &gives
}
