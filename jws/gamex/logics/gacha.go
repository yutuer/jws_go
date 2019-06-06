package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type rspGachaOneMsg struct {
	SyncResp
	RewardId    string `codec:"rid"`
	RewardCount uint32 `codec:"rc"`
	RewardData  string `codec:"rd"`

	GiveRewardId    string `codec:"grid"`
	GiveRewardCount uint32 `codec:"grc"`

	ExtRewardId    string `codec:"extid"`
	ExtRewardCount uint32 `codec:"extc"`
	ExtRewardData  string `codec:"extd"`
}

func giveGet2Client(p *Account, resp helper.ISyncRsp, itemID string, count uint32) (string, uint32, string) {
	gives := new(gamedata.CostData)
	rewardData := gamedata.MakeItemData(p.AccountID.String(), p.GetRand(), itemID)
	if !gamedata.IsFixedIDItemID(itemID) {
		gives.AddItemWithData(itemID, *rewardData, count)
	} else {
		gives.AddItem(itemID, count)
	}
	ok, res2Client := account.GiveBySyncWith2Client(p.Account, gives, resp, "GachaGive")
	if !ok || res2Client == nil {
		logs.Error("give Data Err By %v", gives)
		return "", 0, ""
	}
	return res2Client.Item2Client[0],
		res2Client.Count2Client[0],
		res2Client.Data2Client[0]
}

func (p *Account) GachaOne(r servers.Request) *servers.Response {
	req := &struct {
		Req
		GachaIdx int `codec:"id"`
	}{}
	resp := &rspGachaOneMsg{}

	initReqRsp(
		"PlayerAttr/GachaOneResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_                 = iota
		CODE_IDX_Err      // 失败:请求的索引不存在
		CODE_Cost_Err     // 失败:ID对应的数据不存在
		CODE_Reward_Err   // 失败:不满足完成条件
		CODE_Give_Err     // 失败:完成失败
		CODE_Bag_Full_Err // 失败：包裹满
		CODE_VIP_Err      // 失败:VIP等级不够
		Err_No_Act        // 限时神将活动未开启
	)

	now_time := p.Profile.GetProfileNowTime()

	if gamedata.IsHeroSurplusGacha(req.GachaIdx) {
		if retCode := p.surplusGachaOne(int(req.GachaIdx), resp); retCode != 0 {
			return rpcWarn(resp, retCode)
		} else {
			return rpcSuccess(resp)
		}
	}

	var hgr_act_id uint32
	if gamedata.GetHotDatas().HotLimitHeroGachaData.IsGachaHGR(uint32(req.GachaIdx)) {
		//活动是否开启
		if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_Value_Limit_Hero) {
			return rpcWarn(resp, errCode.ActivityTimeOut)
		}
		actId := gamedata.GetHotDatas().HotLimitHeroGachaData.GetHGRCurrOptValidActivityId()
		if actId <= 0 {
			logs.Warn("GachaOne Err_No_Act")
			return rpcWarn(resp, errCode.HeroGachaRaceTimeOut)
		}
		if !gamedata.GetHotDatas().HotLimitHeroGachaData.IsActivityGachaValid(actId, uint32(req.GachaIdx)) {
			return rpcWarn(resp, errCode.HeroGachaRaceTimeOut)
		}
		p.Profile.GetHeroGachaRaceInfo().CheckActivity(p.AccountID.String(), int64(actId))
		hgr_act_id = actId
	}

	acid := p.AccountID.String()
	corp_lv, _ := p.Profile.GetCorp().GetXpInfo()
	avatar_id := p.Profile.GetCurrAvatar()

	gacha_state := p.Profile.GetGacha(req.GachaIdx)
	data := gamedata.GetGachaData(corp_lv, req.GachaIdx)
	if gacha_state == nil || data == nil {
		return rpcError(resp, CODE_IDX_Err)
	}

	// VIP
	if req.GachaIdx == account.GachaVIP-1 {
		vipCfg := p.Profile.GetMyVipCfg()
		if !vipCfg.VIPGachaLimit {
			return rpcError(resp, CODE_VIP_Err)
		}
	}

	// 检查装备物品数量
	if req.GachaIdx != 2 &&
		//(p.BagProfile.GetEquipCount() >= gamedata.GetEquipCountUpLimit() ||
		p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
		return rpcError(resp, CODE_Bag_Full_Err)
	}

	// 单次免费抽奖全按照高概率抽奖
	is_cost_ok := false
	cost_hc_typ := helper.HC_From_Buy

	p.Profile.Gacha.Update(now_time)
	//标志玩家使用货币
	useCoin := false
	//标志玩家免费抽奖
	useNothing := false
	//标志玩家使用抽奖券
	useTicket := false
	if !gacha_state.IsCanFree(corp_lv, now_time, req.GachaIdx) {
		c := account.CostGroup{}
		costForOne := gamedata.CostData{}
		if p.Profile.GetSC().HasSC(helper.SCId(data.CostForOne_TTyp), int64(data.CostForOne_TCount)) {
			//有足够抽一次奖的抽奖券，消耗抽奖券
			costForOne = data.CostForOneTicket
			useTicket = true
		} else {
			//无足够抽一次奖的抽奖券，消耗货币
			costForOne = data.CostForOneCoin
			useCoin = true
		}
		if !c.AddCostData(p.Account, &costForOne) {
			logs.Warn("GachaOne CostForOneCoin err")
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}

		is_cost_ok, cost_hc_typ = c.CostWithHCBySync(p.Account, helper.HC_From_Buy, resp,
			helper.GachaTypeString(req.GachaIdx, false))
		if !is_cost_ok {
			logs.Warn("GachaOne CostHC err")
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}
	} else {
		//免费抽奖，什么都不消耗
		gacha_state.SetUseFreeNow(now_time)
		useNothing = true
	}
	resp.RewardId, resp.RewardCount, resp.RewardData = p.getGachaReward(data, gacha_state, resp, req.GachaIdx, cost_hc_typ, 0)
	if resp.RewardId == "" {
		return rpcError(resp, CODE_Reward_Err)
	}

	rewardForLog := make(map[string]uint32, 10)
	rewardItemForSN := make([]string, 0, 10)
	rewardCountForSN := make([]uint32, 0, 10)
	rewardForLog, rewardItemForSN, rewardCountForSN = _addItem(resp.RewardId, resp.RewardCount, rewardForLog,
		rewardItemForSN, rewardCountForSN)

	extGives := gacha_state.GetExtReward(corp_lv, req.GachaIdx, avatar_id, p.GetRand())
	if extGives != nil {
		logs.Trace("extGives %v", *extGives)
		resp.ExtRewardId, resp.ExtRewardCount, resp.ExtRewardData =
			giveGet2Client(p, resp, extGives.Id, extGives.Count)
		rewardForLog = _addItemM(resp.ExtRewardId, resp.ExtRewardCount, rewardForLog)
	}

	ok, _ := account.GiveBySyncWith2Client(p.Account, &data.GiveForOne.Cost,
		resp, helper.GachaTypeString(req.GachaIdx, false))
	if !ok {
		return rpcError(resp, CODE_Give_Err)
	}
	// 埋点数据，需要处理，判断玩家是使用货币还是使用抽奖券还是免费抽奖

	switch {
	case useNothing:
		logiclog.LogGacha(acid, avatar_id,
			p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
			helper.GachaTypeString(req.GachaIdx, false),
			"FreeGacha",
			0,
			rewardForLog,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	case useCoin:
		logiclog.LogGacha(acid, avatar_id,
			p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
			helper.GachaTypeString(req.GachaIdx, false),
			data.CostForOne_Typ,
			data.CostForOne_Count,
			rewardForLog,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	case useTicket:
		logiclog.LogGacha(acid, avatar_id,
			p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
			helper.GachaTypeString(req.GachaIdx, false),
			data.CostForOne_TTyp,
			data.CostForOne_TCount,
			rewardForLog,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	}

	resp.GiveRewardId = data.GiveForOne.PriceTyp
	resp.GiveRewardCount = data.GiveForOne.PriceCount

	p.updateCondition(account.COND_TYP_GachaOne, 1, req.GachaIdx+1, "", "", resp)
	p.Profile.GetHmtActivityInfo().AddGachaNum(p.GetProfileNowTime(), 1)
	// 限时名将
	if hgr_act_id > 0 {
		if p.OnGachaRace(hgr_act_id, 1) {
			resp.OnChangeHeroGachaRace()
		}
	}

	resp.OnChangeSC()
	resp.OnChangeHC()
	resp.OnChangeGachaAllChange()
	resp.mkInfo(p)

	// sysnotice
	cfgSN := gamedata.GachaHeroSysNotice()
	cfgWSN := gamedata.GachaHeroWholeSysNotice()
	for i, item := range rewardItemForSN {
		c := rewardCountForSN[i]
		if ok, _, _, cfg := gamedata.IsHeroPieceItem(item); ok {
			if uint32(cfg.GetRareLevel()) >= cfgSN.GetLampValueIP1() {
				sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgSN.GetServerMsgID())).
					AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
					AddParam(sysnotice.ParamType_ItemId, fmt.Sprintf("%s", item)).
					AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", c)).Send()
			}
		}
		if ok, heroId, _, cfg := gamedata.IsItemToWholeCharWhenAdd(item); ok {
			if id := gamedata.GetHeroByHeroID(heroId); id >= 0 {
				if uint32(cfg.GetRareLevel()) >= cfgWSN.GetLampValueIP1() {
					sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgWSN.GetServerMsgID())).
						AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
						AddParam(sysnotice.ParamType_Hero, fmt.Sprintf("%d", id)).Send()
				}
			}
		}
	}

	return rpcSuccess(resp)
}

type rspGachaTenMsg struct {
	SyncResp
	RewardId    []string `codec:"rid"`
	RewardCount []uint32 `codec:"rc"`
	RewardData  []string `codec:"rd"`

	GiveRewardId    string `codec:"grid"`
	GiveRewardCount uint32 `codec:"grc"`

	ExtRewardId    []string `codec:"extids"`
	ExtRewardCount []uint32 `codec:"extcs"`
	ExtRewardData  []string `codec:"extds"`
}

func (p *Account) GachaTen(r servers.Request) *servers.Response {
	req := &struct {
		Req
		GachaIdx int `codec:"id"`
	}{}
	resp := &rspGachaTenMsg{}

	initReqRsp(
		"PlayerAttr/GachaTenResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_                 = iota
		CODE_IDX_Err      // 失败:请求的索引不存在
		CODE_Cost_Err     // 失败:ID对应的数据不存在
		CODE_Reward_Err   // 失败:不满足完成条件
		CODE_SReward_Err  // 失败:不满足完成条件
		CODE_Give_Err     // 失败:完成失败
		CODE_Bag_Full_Err // 失败:包裹满
		CODE_VIP_Err      // 失败:VIP等级不够
		Err_No_Act
	)

	if gamedata.IsHeroSurplusGacha(req.GachaIdx) {
		if retCode := p.surplusGachaTen(int(req.GachaIdx), resp); retCode != 0 {
			return rpcWarn(resp, retCode)
		} else {
			return rpcSuccess(resp)
		}
	}

	var hgr_act_id uint32
	if gamedata.GetHotDatas().HotLimitHeroGachaData.IsGachaHGR(uint32(req.GachaIdx)) {
		//活动是否开启
		if !game.Cfg.GetHotActValidData(p.AccountID.ShardId, uutil.Hot_Value_Limit_Hero) {
			return rpcWarn(resp, errCode.ActivityTimeOut)
		}
		actId := gamedata.GetHotDatas().HotLimitHeroGachaData.GetHGRCurrOptValidActivityId()
		if actId <= 0 {
			return rpcWarn(resp, errCode.HeroGachaRaceTimeOut)
		}
		if !gamedata.GetHotDatas().HotLimitHeroGachaData.IsActivityGachaValid(actId, uint32(req.GachaIdx)) {
			return rpcWarn(resp, errCode.HeroGachaRaceTimeOut)
		}
		p.Profile.GetHeroGachaRaceInfo().CheckActivity(p.AccountID.String(), int64(actId))
		hgr_act_id = actId
	}

	corp_lv, _ := p.Profile.GetCorp().GetXpInfo()
	avatar_id := p.Profile.GetCurrAvatar()

	gacha_state := p.Profile.GetGacha(req.GachaIdx)
	data := gamedata.GetGachaData(corp_lv, req.GachaIdx)
	if gacha_state == nil || data == nil {
		return rpcError(resp, CODE_IDX_Err)
	}

	// VIP
	if req.GachaIdx == account.GachaVIP-1 {
		vipCfg := p.Profile.GetMyVipCfg()
		if !vipCfg.VIPGachaLimit {
			return rpcError(resp, CODE_VIP_Err)
		}
	}

	// 检查装备物品数量
	//if p.BagProfile.GetEquipCount() >= gamedata.GetEquipCountUpLimit() ||
	if p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
		return rpcError(resp, CODE_Bag_Full_Err)
	}

	c := account.CostGroup{}
	usingCoin := false
	usingTicket := false

	costForTen := gamedata.CostData{}
	if p.Profile.GetSC().HasSC(helper.SCId(data.CostForTen_TTyp), int64(data.CostForTen_TCount)) {
		//有足够抽一次十连抽的抽奖券，消耗抽奖券
		costForTen = data.CostForTenTicket
		usingTicket = true
	} else {
		//无足够抽一次十连抽的抽奖券，消耗货币
		costForTen = data.CostForTenCoin
		usingCoin = true
	}

	if !c.AddCostData(p.Account, &costForTen) {
		logs.Warn("GachaTen CODE_Cost_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	is_cost_ok, cost_hc_typ := c.CostWithHCBySync(p.Account, helper.HC_From_Buy, resp,
		helper.GachaTypeString(req.GachaIdx, true))
	if !is_cost_ok {
		logs.Warn("GachaTen CODE_Cost_Err")
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	resp.RewardId = make([]string, 0, 10)
	resp.RewardCount = make([]uint32, 0, 10)
	resp.RewardData = make([]string, 0, 10)

	resp.ExtRewardId = make([]string, 0, 10)
	resp.ExtRewardCount = make([]uint32, 0, 10)

	rewardForLog := make(map[string]uint32, 10)
	rewardItemForSN := make([]string, 0, 10)
	rewardCountForSN := make([]uint32, 0, 10)
	ct := 1
	var respRewardId string
	var respRewardCount uint32
	var respRewardData string
	for i := 0; i < 10; i++ {
		respRewardId, respRewardCount, respRewardData = p.getGachaReward(data, gacha_state, resp, req.GachaIdx, cost_hc_typ, ct)
		ct += 1
		if respRewardId == "" {
			return rpcError(resp, CODE_Reward_Err)
		}
		//logs.Trace("[%s]gacha ten %d reward %v controlcount %d",
		//	p.AccountID, i, reward, gacha_state.HeroGachaRaceCount)

		resp.RewardId = append(resp.RewardId, respRewardId)
		resp.RewardCount = append(resp.RewardCount, respRewardCount)
		resp.RewardData = append(resp.RewardData, respRewardData)
		rewardForLog, rewardItemForSN, rewardCountForSN = _addItem(respRewardId, respRewardCount, rewardForLog,
			rewardItemForSN, rewardCountForSN)
		extGives := gacha_state.GetExtReward(corp_lv, req.GachaIdx, avatar_id, p.GetRand())
		if extGives != nil {
			logs.Trace("extGives %v", *extGives)
			respExtRewardId, respExtRewardCount, respExtRewardData :=
				giveGet2Client(p, resp, extGives.Id, extGives.Count)
			resp.ExtRewardId = append(resp.ExtRewardId, respExtRewardId)
			resp.ExtRewardCount = append(resp.ExtRewardCount, respExtRewardCount)
			resp.ExtRewardData = append(resp.ExtRewardData, respExtRewardData)
			rewardForLog = _addItemM(respExtRewardId, respExtRewardCount, rewardForLog)
		}
	}

	g := account.GiveGroup{}
	g.AddCostData(&data.GiveForTen.Cost)
	rewardForLog = _addItemM(data.GiveForTen.PriceTyp, data.GiveForTen.PriceCount, rewardForLog)

	ok, _ := g.GiveBySyncWithRes(p.Account, resp, helper.GachaTypeString(req.GachaIdx, true))
	if !ok {
		return rpcError(resp, CODE_Give_Err)
	}

	resp.GiveRewardId = data.GiveForTen.PriceTyp
	resp.GiveRewardCount = data.GiveForTen.PriceCount
	rewardForLog = _addItemM(resp.GiveRewardId, resp.GiveRewardCount, rewardForLog)

	//埋点，判断消耗的是货币还是抽奖券
	switch {
	case usingCoin:
		logiclog.LogGacha(p.AccountID.String(), avatar_id,
			p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
			helper.GachaTypeString(req.GachaIdx, true),
			data.CostForTen_Typ,
			data.CostForTen_Count,
			rewardForLog,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	case usingTicket:
		logiclog.LogGacha(p.AccountID.String(), avatar_id,
			p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
			helper.GachaTypeString(req.GachaIdx, true),
			data.CostForTen_TTyp,
			data.CostForTen_TCount,
			rewardForLog,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	}

	p.updateCondition(account.COND_TYP_GachaOne, 10, req.GachaIdx+1, "", "", resp)
	p.Profile.GetHmtActivityInfo().AddGachaNum(p.GetProfileNowTime(), 10)
	// 限时名将
	if hgr_act_id > 0 {
		if p.OnGachaRace(hgr_act_id, 10) {
			resp.OnChangeHeroGachaRace()
		}
	}
	resp.OnChangeSC()
	resp.OnChangeHC()
	resp.OnChangeGachaAllChange()

	resp.mkInfo(p)

	// sysnotice
	cfgSN := gamedata.GachaHeroSysNotice()
	cfgWSN := gamedata.GachaHeroWholeSysNotice()
	for i, item := range rewardItemForSN {
		c := rewardCountForSN[i]
		if ok, _, _, cfg := gamedata.IsHeroPieceItem(item); ok {
			if uint32(cfg.GetRareLevel()) >= cfgSN.GetLampValueIP1() {
				sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgSN.GetServerMsgID())).
					AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
					AddParam(sysnotice.ParamType_ItemId, item).
					AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", c)).Send()
			}
		}
		if ok, heroId, _, cfg := gamedata.IsItemToWholeCharWhenAdd(item); ok {
			if id := gamedata.GetHeroByHeroID(heroId); id >= 0 {
				if uint32(cfg.GetRareLevel()) >= cfgWSN.GetLampValueIP1() {
					sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgWSN.GetServerMsgID())).
						AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
						AddParam(sysnotice.ParamType_Hero, fmt.Sprintf("%d", id)).Send()
				}
			}
		}
	}

	return rpcSuccess(resp)
}

func _addItem(item string, count uint32, rm map[string]uint32, items []string, counts []uint32) (map[string]uint32,
	[]string, []uint32) {

	_addItemM(item, count, rm)
	items = append(items, item)
	counts = append(counts, count)
	return rm, items, counts
}

func _addItemM(item string, count uint32, rm map[string]uint32) map[string]uint32 {
	if o, ok := rm[item]; ok {
		rm[item] = o + count
	} else {
		rm[item] = count
	}
	return rm
}

func (p *Account) getGachaReward(data *gamedata.GachaData, gacha_state *account.GachaState,
	resp helper.ISyncRsp, gachaIdx int, cost_hc_typ int, ct int) (string, uint32, string) {
	corp_lv, _ := p.Profile.GetCorp().GetXpInfo()
	avatar_id := p.Profile.GetCurrAvatar()
	reward := gacha_state.Gacha(
		p.AccountID.String(),
		corp_lv,
		avatar_id,
		gachaIdx,
		cost_hc_typ,
		ct,
		p.GetRand())

	if reward == nil {
		return "", 0, ""
	}
	var respRewardId string
	var respRewardCount uint32
	var respRewardData string
	if data.AfricaNumber > 0 {
		gacha_state.HeroGachaRaceCount += 1
		if gacha_state.HeroGachaRaceCount < int64(data.AfricaNumber) {
			if reward.Id == data.ItemID && reward.Count == data.ItemNum {
				gacha_state.HeroGachaRaceCount = 0
			}
		}
		if gacha_state.HeroGachaRaceCount == int64(data.AfricaNumber) {
			//本次奖励替换为直接给道具
			respRewardId, respRewardCount, respRewardData = giveGet2Client(p, resp, data.ItemID, data.ItemNum)
			//次数归零
			gacha_state.HeroGachaRaceCount = 0
		} else {
			//正常奖励
			respRewardId, respRewardCount, respRewardData = giveGet2Client(p, resp, reward.Id, reward.Count)
		}
	} else {
		respRewardId, respRewardCount, respRewardData = giveGet2Client(p, resp, reward.Id, reward.Count)
	}
	return respRewardId, respRewardCount, respRewardData
}

func (p *Account) surplusGachaOne(gachaIdx int, resp *rspGachaOneMsg) uint32 {
	// 检查条件
	if errCode := p.checkSurplusGacha(gachaIdx, 1); errCode != 0 {
		return errCode
	}

	p.Profile.Gacha.Update(p.Profile.GetProfileNowTime())

	acid := p.AccountID.String()
	corp_lv, _ := p.Profile.GetCorp().GetXpInfo()
	avatar_id := p.Profile.GetCurrAvatar()
	data := gamedata.GetGachaData(corp_lv, gachaIdx)

	// 消耗道具
	c := account.CostGroup{}
	if !c.AddCostData(p.Account, &data.CostForOneCoin) {
		return errCode.ClickTooQuickly
	}
	reason := fmt.Sprintf("surplus gacha one %d", gachaIdx)
	if !c.CostBySync(p.Account, resp, reason) {
		return errCode.ClickTooQuickly
	}

	gacha_state := p.Profile.GetGacha(gachaIdx)

	// 奖励
	resp.RewardId, resp.RewardCount, resp.RewardData = p.getGachaReward(data, gacha_state, resp, gachaIdx, helper.HC_From_Buy, 0)
	if resp.RewardId == "" {
		return errCode.CommonInner
	}

	// 记录抽奖次数
	p.Profile.GetHeroSurplusInfo().AddDailyDrawCount(gachaIdx-12, 1)

	rewardForLog := make(map[string]uint32, 10)
	rewardItemForSN := make([]string, 0, 10)
	rewardCountForSN := make([]uint32, 0, 10)
	rewardForLog, rewardItemForSN, rewardCountForSN = _addItem(resp.RewardId, resp.RewardCount, rewardForLog,
		rewardItemForSN, rewardCountForSN)

	extGives := gacha_state.GetExtReward(corp_lv, gachaIdx, avatar_id, p.GetRand())
	if extGives != nil {
		logs.Trace("extGives %v", *extGives)
		resp.ExtRewardId, resp.ExtRewardCount, resp.ExtRewardData =
			giveGet2Client(p, resp, extGives.Id, extGives.Count)
		rewardForLog = _addItemM(resp.ExtRewardId, resp.ExtRewardCount, rewardForLog)
	}

	ok, _ := account.GiveBySyncWith2Client(p.Account, &data.GiveForOne.Cost,
		resp, reason)

	if !ok {
		return errCode.ClickTooQuickly
	}

	// logiclog
	logiclog.LogGacha(acid, avatar_id,
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		helper.GachaTypeString(gachaIdx, false),
		data.CostForOne_Typ,
		data.CostForOne_Count,
		rewardForLog,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	resp.GiveRewardId = data.GiveForOne.PriceTyp
	resp.GiveRewardCount = data.GiveForOne.PriceCount

	resp.OnChangeGachaAllChange()
	resp.mkInfo(p)

	cfgSN := gamedata.GachaHeroSysNotice()
	cfgWSN := gamedata.GachaHeroWholeSysNotice()
	for i, item := range rewardItemForSN {
		c := rewardCountForSN[i]
		if ok, _, _, cfg := gamedata.IsHeroPieceItem(item); ok {
			if uint32(cfg.GetRareLevel()) >= cfgSN.GetLampValueIP1() {
				sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgSN.GetServerMsgID())).
					AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
					AddParam(sysnotice.ParamType_ItemId, fmt.Sprintf("%s", item)).
					AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", c)).Send()
			}
		}
		if ok, heroId, _, cfg := gamedata.IsItemToWholeCharWhenAdd(item); ok {
			if id := gamedata.GetHeroByHeroID(heroId); id >= 0 {
				if uint32(cfg.GetRareLevel()) >= cfgWSN.GetLampValueIP1() {
					sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgWSN.GetServerMsgID())).
						AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
						AddParam(sysnotice.ParamType_Hero, fmt.Sprintf("%d", id)).Send()
				}
			}
		}
	}

	return 0
}

func (p *Account) checkSurplusGacha(gachaId int, count int) uint32 {
	nowTime := p.Profile.GetProfileNowTime()
	surplusInfo := p.Profile.GetHeroSurplusInfo()
	surplusInfo.TryDailyReset(nowTime)
	if nowTime > surplusInfo.EndTime {
		logs.Warn("<hero surplus gacha> time is over, now = %d, end = %d", nowTime, surplusInfo.EndTime)
		return errCode.CommonNotInTime
	}
	vipCfg := p.Profile.GetMyVipCfg()
	surplusType := gamedata.GetSurplusTypeById(gachaId)
	if surplusType == -1 {
		return errCode.CommonConditionFalse
	}
	if surplusInfo.DailyDrawCount[surplusType]+count > vipCfg.SurplusGachaLimit[surplusType] {
		logs.Warn("<hero surplus gacha> count is limit, now count %d", surplusInfo.DailyDrawCount)
		return errCode.CommonCountLimit
	}
	return 0
}

func (p *Account) surplusGachaTen(gachaIdx int, resp *rspGachaTenMsg) uint32 {
	if errCode := p.checkSurplusGacha(gachaIdx, 10); errCode != 0 {
		return errCode
	}

	p.Profile.Gacha.Update(p.Profile.GetProfileNowTime())

	corp_lv, _ := p.Profile.GetCorp().GetXpInfo()
	avatar_id := p.Profile.GetCurrAvatar()
	data := gamedata.GetGachaData(corp_lv, gachaIdx)

	// 消耗道具
	c := account.CostGroup{}
	if !c.AddCostData(p.Account, &data.CostForTenCoin) {
		return errCode.ClickTooQuickly
	}
	reason := fmt.Sprintf("surplus gacha ten %d", gachaIdx)
	if !c.CostBySync(p.Account, resp, reason) {
		return errCode.ClickTooQuickly
	}

	gacha_state := p.Profile.GetGacha(gachaIdx)

	resp.RewardId = make([]string, 0, 10)
	resp.RewardCount = make([]uint32, 0, 10)
	resp.RewardData = make([]string, 0, 10)

	resp.ExtRewardId = make([]string, 0, 10)
	resp.ExtRewardCount = make([]uint32, 0, 10)

	rewardForLog := make(map[string]uint32, 10)
	rewardItemForSN := make([]string, 0, 10)
	rewardCountForSN := make([]uint32, 0, 10)
	ct := 1
	var respRewardId string
	var respRewardCount uint32
	var respRewardData string
	for i := 0; i < 10; i++ {
		respRewardId, respRewardCount, respRewardData = p.getGachaReward(data, gacha_state, resp, gachaIdx, helper.HC_From_Buy, ct)
		ct += 1
		if respRewardId == "" {
			return errCode.ClickTooQuickly
		}
		//logs.Trace("[%s]gacha ten %d reward %v controlcount %d",
		//	p.AccountID, i, reward, gacha_state.HeroGachaRaceCount)

		resp.RewardId = append(resp.RewardId, respRewardId)
		resp.RewardCount = append(resp.RewardCount, respRewardCount)
		resp.RewardData = append(resp.RewardData, respRewardData)
		rewardForLog, rewardItemForSN, rewardCountForSN = _addItem(respRewardId, respRewardCount, rewardForLog,
			rewardItemForSN, rewardCountForSN)
		extGives := gacha_state.GetExtReward(corp_lv, gachaIdx, avatar_id, p.GetRand())
		if extGives != nil {
			logs.Trace("extGives %v", *extGives)
			respExtRewardId, respExtRewardCount, respExtRewardData :=
				giveGet2Client(p, resp, extGives.Id, extGives.Count)
			resp.ExtRewardId = append(resp.ExtRewardId, respExtRewardId)
			resp.ExtRewardCount = append(resp.ExtRewardCount, respExtRewardCount)
			resp.ExtRewardData = append(resp.ExtRewardData, respExtRewardData)
			rewardForLog = _addItemM(respExtRewardId, respExtRewardCount, rewardForLog)
		}
	}

	g := account.GiveGroup{}
	g.AddCostData(&data.GiveForTen.Cost)
	rewardForLog = _addItemM(data.GiveForTen.PriceTyp, data.GiveForTen.PriceCount, rewardForLog)

	p.Profile.GetHeroSurplusInfo().AddDailyDrawCount(gachaIdx-12, 10)

	ok, _ := g.GiveBySyncWithRes(p.Account, resp, helper.GachaTypeString(gachaIdx, true))
	if !ok {
		return errCode.ClickTooQuickly
	}

	resp.GiveRewardId = data.GiveForTen.PriceTyp
	resp.GiveRewardCount = data.GiveForTen.PriceCount
	rewardForLog = _addItemM(resp.GiveRewardId, resp.GiveRewardCount, rewardForLog)

	// logiclog
	logiclog.LogGacha(p.AccountID.String(), avatar_id,
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		helper.GachaTypeString(gachaIdx, true),
		data.CostForTen_Typ,
		data.CostForTen_Count,
		rewardForLog,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	resp.OnChangeSC()
	resp.OnChangeHC()
	resp.OnChangeGachaAllChange()

	resp.mkInfo(p)

	// sysnotice
	cfgSN := gamedata.GachaHeroSysNotice()
	cfgWSN := gamedata.GachaHeroWholeSysNotice()
	for i, item := range rewardItemForSN {
		c := rewardCountForSN[i]
		if ok, _, _, cfg := gamedata.IsHeroPieceItem(item); ok {
			if uint32(cfg.GetRareLevel()) >= cfgSN.GetLampValueIP1() {
				sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgSN.GetServerMsgID())).
					AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
					AddParam(sysnotice.ParamType_ItemId, item).
					AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", c)).Send()
			}
		}
		if ok, heroId, _, cfg := gamedata.IsItemToWholeCharWhenAdd(item); ok {
			if id := gamedata.GetHeroByHeroID(heroId); id >= 0 {
				if uint32(cfg.GetRareLevel()) >= cfgWSN.GetLampValueIP1() {
					sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgWSN.GetServerMsgID())).
						AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
						AddParam(sysnotice.ParamType_Hero, fmt.Sprintf("%d", id)).Send()
				}
			}
		}
	}

	return 0
}

//func (p *Account) sendSysNotice(rewardItemForSN []string) {
//	cfgSN := gamedata.GachaHeroSysNotice()
//	cfgWSN := gamedata.GachaHeroWholeSysNotice()
//	for i, item := range rewardItemForSN {
//		c := rewardCountForSN[i]
//		if ok, _, _, cfg := gamedata.IsHeroPieceItem(item); ok {
//			if uint32(cfg.GetRareLevel()) >= cfgSN.GetLampValueIP1() {
//				sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgSN.GetServerMsgID())).
//					AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
//					AddParam(sysnotice.ParamType_ItemId, item).
//					AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", c)).Send()
//			}
//		}
//		if ok, heroId, _, cfg := gamedata.IsItemToWholeCharWhenAdd(item); ok {
//			if id := gamedata.GetHeroByHeroID(heroId); id >= 0 {
//				if uint32(cfg.GetRareLevel()) >= cfgWSN.GetLampValueIP1() {
//					sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(cfgWSN.GetServerMsgID())).
//						AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
//						AddParam(sysnotice.ParamType_Hero, fmt.Sprintf("%d", id)).Send()
//				}
//			}
//		}
//	}
//}
