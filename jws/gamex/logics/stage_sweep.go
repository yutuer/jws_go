package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestStageSweep struct {
	Req
	LevelID       string `codec:"levelid"`
	AvatarIDs     []int  `codec:"avatar_id"`
	Count         int    `codec:"c"`
	ItemIDNeed    string `codec:"item_need"`
	ItemNeedCount uint32 `codec:"need_count"`
}

type RewardInfoStatistics struct {
	rewards map[string]uint32
}

func (r *RewardInfoStatistics) OnReward(itemID string, count uint32) {
	if r.rewards == nil {
		r.rewards = make(map[string]uint32, 8)
	}
	old, ok := r.rewards[itemID]
	if ok {
		r.rewards[itemID] = old + count
	} else {
		r.rewards[itemID] = count
	}
}

func (r *RewardInfoStatistics) IsGetted(itemID string, count uint32) bool {
	if r.rewards == nil {
		return false
	}
	c, ok := r.rewards[itemID]
	if !ok {
		return false
	}

	return c >= count
}

type ResponseStageSweep struct {
	SyncResp
	StageRewards [][][]byte `codec:"rewards"`
	ScType       []int      `codec:"sc_t"` // TBD 如果确认是发钱的话 可以移除
	ScValue      []int64    `codec:"sc_v"`
	CorpXpAdd    []uint32   `codec:"corp_xp_add"` // 战队经验
	statistics   RewardInfoStatistics
}

func (r *ResponseStageSweep) AddResReward(g *gamedata.CostData2Client) {
	res := new(ResponseDeclareLootForLvlEnmy)
	res.AddResReward(g)
	r.addStageReward(res)
	for idx, itemID := range g.Item2Client {
		count := g.Count2Client[idx]
		r.statistics.OnReward(itemID, count)
	}
}

func (r *ResponseStageSweep) MergeReward() {}

func (r *ResponseStageSweep) AddReward(vipLv int, rID string, count uint32, data string) {

}
func (r *ResponseStageSweep) AddRewards(vipLv int, d *gamedata.PriceDatas) {

}
func (r *ResponseStageSweep) init(count int) {
	r.StageRewards = make([][][]byte, 0, count)
	r.ScType = make([]int, 0, count)
	r.ScValue = make([]int64, 0, count)
	r.CorpXpAdd = make([]uint32, 0, count)
}

func (r *ResponseStageSweep) addStageReward(resp *ResponseDeclareLootForLvlEnmy) {
	r.StageRewards = append(r.StageRewards, resp.StageRewards)
	r.ScType = append(r.ScType, resp.ScType)
	r.ScValue = append(r.ScValue, resp.ScValue)
	r.CorpXpAdd = append(r.CorpXpAdd, resp.CorpXpAdd)
}

func (p *Account) StageSweep(r servers.Request) *servers.Response {
	const (
		_                  = iota
		CODE_SendRewardErr // 警告:奖励物品发送失败
		CODE_LootIdErr     // 警告:申请的掉落ID无效
		CODE_LootCountErr  // 警告:次数错误
		CODE_VIP_NOT_FOUND // 警告:配置错误
		CODE_VIP_ERR       // 警告:vip级别不够
		CODE_Bag_Full_Err  // 失败：包裹满
	)
	const (
		CODE_MIN = 20
	)

	acid := p.AccountID.String()
	req := &RequestStageSweep{}
	resp := &ResponseStageSweep{}

	initReqRsp(
		"PlayLevel/StageSweepRsp",
		r.RawBytes,
		req, resp, p)

	if req.Count <= 0 {
		return rpcError(resp, CODE_LootCountErr)
	}

	// 检查装备物品数量
	if p.BagProfile.GetEquipCount() >= gamedata.GetEquipCountUpLimit() {
		return rpcErrorWithMsg(resp, CODE_Bag_Full_Err, fmt.Sprintf("CODE_Bag_Full_Err for equip"))
	}
	//if p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
	//	return rpcErrorWithMsg(resp, CODE_Bag_Full_Err, fmt.Sprintf("CODE_Bag_Full_Err for jade"))
	//}

	stage_id := req.LevelID

	curVip, _ := p.Profile.GetVip().GetVIP()
	vipInfo := gamedata.GetVIPCfg(int(curVip))
	if vipInfo == nil {
		logs.SentryLogicCritical(acid, "Stage Sweep Error %s %d", stage_id, CODE_VIP_NOT_FOUND)
		return rpcError(resp, CODE_VIP_NOT_FOUND)
	}
	if req.Count > 1 { // 连扫
		if !vipInfo.SweepTenValid {
			logs.SentryLogicCritical(acid, "Stage Sweep Error %s %d", stage_id, CODE_VIP_ERR)
			return rpcError(resp, CODE_VIP_ERR)
		}
	}

	for i := 0; i < req.Count; i++ {
		ok, loots := mkLootDataForSweep(p, req.LevelID, req.AvatarIDs)
		if !ok {
			return rpcError(resp, CODE_SendRewardErr)
		} else {
			code, warnCode := p.CostStagePay(stage_id, req.AvatarIDs, true, resp)
			if warnCode != 0 {
				return rpcWarn(resp, warnCode)
			}
			if code != 0 {
				// 扣体力，校验次数等失败了
				logs.Warn(acid, "%s StageSweep Cost Error %s %d", stage_id, code)
				return rpcWarn(resp, errCode.ClickTooQuickly)
			}

			// 副本结算 发送奖励
			rewards := gamedata.NewPriceDataSet(8)
			rewards.AppendOther(p.sendStageLimitReward(stage_id, false))
			rewards.AppendOther(p.sendStageRandReward(stage_id))
			allrewards := rewards.Mk2One()

			stage_reward_data := gamedata.GetStageReward(stage_id, false)
			stage_data := gamedata.GetStageData(stage_id)
			if stage_reward_data != nil && stage_data != nil {
				allrewards.AddItem(gamedata.VI_Sc0, uint32(stage_reward_data.SCReward))
				allrewards.AddItem(gamedata.VI_CorpXP, stage_reward_data.CorpXpReward)
				allrewards.AddItem(gamedata.VI_XP, stage_reward_data.XpReward)
				if stage_reward_data.SweepItem != "" {
					allrewards.AddItem(
						stage_reward_data.SweepItem,
						stage_reward_data.SweepCount)
				}
			}

			for _, reward := range loots {
				if !gamedata.IsFixedIDItemID(reward.Item_id) {
					item_data := gamedata.MakeItemData(p.AccountID.String(), p.Account.GetRand(), reward.Item_id)
					allrewards.AddItemWithData(reward.Item_id, *item_data, reward.Count)
				} else {
					allrewards.AddItem(reward.Item_id, reward.Count)
				}
			}

			// 兑换商店道具掉落
			// 兑换商店道具掉落
			giveExchangeData := p.getExchangeShopLoot(req.LevelID, uint32(stage_data.Type))
			allrewards.Gives().AddGroup(giveExchangeData)
			if !account.GiveBySync(p.Account, allrewards.Gives(), resp, "stageSweep") {
				logs.Error("GiveBySync Err By %v", allrewards)
			}
			// 星级计算
			player_stage_info := p.Profile.GetStage().GetStageInfo(
				gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()),
				stage_id,
				p.GetRand())
			player_stage_info.T_count += 1
			player_stage_info.Sum_count += 1

			p.updateCondition(account.COND_TYP_Stage_Pass,
				1, MAX_STAR, stage_id, "", resp)

			// 主线关卡通过次数条件
			if stage_data.Type == gamedata.LEVEL_TYPE_MAIN {
				p.updateCondition(account.COND_TYP_Any_Stage_Pass,
					1, 1, "", "", resp)
				p.Profile.GetMarketActivitys().OnGameMode(acid, gamedata.CounterTypeMain, 1, p.Profile.GetProfileNowTime())
			}
			// 精英关卡通过次数条件
			if stage_data.Type == gamedata.LEVEL_TYPE_ELITE {
				p.updateCondition(account.COND_TYP_Any_Stage_Pass,
					1, 2, "", "", resp)
				p.Profile.GetMarketActivitys().OnGameMode(acid, gamedata.CounterTypeElite, 1, p.Profile.GetProfileNowTime())
				p.Profile.GetHmtActivityInfo().AddDungeonJy(p.GetProfileNowTime(), 1)
			}
			// 地狱关卡通过次数条件
			if stage_data.Type == gamedata.LEVEL_TYPE_HELL {
				p.updateCondition(account.COND_TYP_Any_Stage_Pass,
					1, 3, "", "", resp)
				p.Profile.GetMarketActivitys().OnGameMode(acid, gamedata.CounterTypeHell, 1, p.Profile.GetProfileNowTime())
				p.Profile.GetHmtActivityInfo().AddDungeonDy(p.GetProfileNowTime(), 1)
			}

			// 任意关卡通过次数条件
			p.updateCondition(account.COND_TYP_Any_Stage_Pass,
				1, 0, "", "", resp)

			//检查是否需要停止扫荡
			if req.ItemIDNeed != "" && req.ItemNeedCount > 0 && req.Count > 1 {
				if resp.statistics.IsGetted(req.ItemIDNeed, req.ItemNeedCount) {
					break
				}
			}
		}
	}
	// logiclog
	logiclog.LogStageFinish(acid, p.Profile.GetCurrAvatar(), stage_id, true, 0, req.Count, true, 0,
		p.Profile.GetCorp().Level, p.Profile.ChannelId, 0, [helper.DestinyGeneralSkillMax]int{},
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	// 给客户端刷新
	resp.OnChangeSC()
	resp.OnChangeEnergy()
	resp.OnChangeCorpExp()
	resp.OnChangeStage(stage_id)
	//resp.OnChangeBoss()
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) GameModeLevelSweep(r servers.Request) *servers.Response {
	req := &struct {
		Req
		GameModeId uint32 `codec:"gmid"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayLevel/GameModeLevelSweepResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Vip_Not_Enough
		Err_give
		Err_Not_Finish
	)
	curVip, _ := p.Profile.GetVip().GetVIP()
	vipCfg := gamedata.GetVIPCfg(int(curVip))
	// 次数减少
	ok, errcode, warnCode, leftCount := p.Profile.GetGameMode().GameModeLevelSweep(p.Account, req.GameModeId, resp)
	if !ok {
		if warnCode > 0 {
			return rpcWarn(resp, errCode.ClickTooQuickly)
		}
		return rpcError(resp, errcode+20)
	}
	// 奖励
	data := &gamedata.CostData{}
	var levelId string
	switch req.GameModeId {
	case gamedata.CounterTypeGoldLevel:
		for _, lvlId := range gamedata.GoldLevelOrderByDiffDesc() {
			if p.Profile.GetStage().IsStagePass(lvlId) {
				levelId = lvlId
				break
			}
		}
		if levelId == "" {
			return rpcErrorWithMsg(resp, Err_Not_Finish, fmt.Sprintf("Err_Not_Finish %d", req.GameModeId))
		}

		p.updateCondition(account.COND_TYP_GoldLevel,
			leftCount, 0, "", "", resp)
		p.updateCondition(account.COND_TYP_Try_Test,
			leftCount, 0, "", "", resp)

		cfg := gamedata.GetGoldLevelCfg(levelId)
		sc0 := cfg.GetTotal()
		difclt := cfg.GetDifficulty()
		sc0 += uint32(vipCfg.GoldLevelAdd * float32(sc0))
		sc0 *= uint32(leftCount)
		data.AddItem(gamedata.VI_Sc0, sc0)
		logs.Trace("GameModeLevelSweep GoldLevel difficult %d times %d", difclt, leftCount)
	case gamedata.CounterTypeFineIronLevel:
		for _, lvlId := range gamedata.ExpLevelOrderByDiffDesc() {
			if p.Profile.GetStage().IsStagePass(lvlId) {
				levelId = lvlId
				break
			}
		}
		if levelId == "" {
			return rpcErrorWithMsg(resp, Err_Not_Finish, fmt.Sprintf("Err_Not_Finish %d", req.GameModeId))
		}

		p.updateCondition(account.COND_TYP_FiLevel,
			leftCount, 0, "", "", resp)
		p.updateCondition(account.COND_TYP_Try_Test,
			leftCount, 0, "", "", resp)

		cfg := gamedata.GetExpLevelCfg(levelId)
		sc1 := cfg.GetTotal()
		difclt := cfg.GetDifficulty()
		sc1 += uint32(vipCfg.IronLevelAdd * float32(sc1))
		sc1 *= uint32(leftCount)
		data.AddItem(gamedata.VI_Sc1, sc1)
		logs.Trace("GameModeLevelSweep FineIronLevel difficult %d times %d", difclt, leftCount)
	case gamedata.CounterTypeDCLevel:
		for _, lvlId := range gamedata.DCLevelOrderByDiffDesc() {
			if p.Profile.GetStage().IsStagePass(lvlId) {
				levelId = lvlId
				break
			}
		}
		if levelId == "" {
			return rpcErrorWithMsg(resp, Err_Not_Finish, fmt.Sprintf("Err_Not_Finish %d", req.GameModeId))
		}

		p.updateCondition(account.COND_TYP_ExpLevel,
			leftCount, 0, "", "", resp)
		p.updateCondition(account.COND_TYP_Try_Test,
			leftCount, 0, "", "", resp)

		cfg := gamedata.GetDCLevelCfg(levelId)
		dc := cfg.GetTotal()
		difclt := cfg.GetDifficulty()
		dc += uint32(vipCfg.DcLevelAdd * float32(dc))
		dc *= uint32(leftCount)
		data.AddItem(gamedata.VI_DC, dc)
		logs.Trace("GameModeLevelSweep DCLevel difficult %d times %d", difclt, leftCount)
	}
	give := &account.GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(p.Account, resp, "GameModeLevelSweep") {
		return rpcErrorWithMsg(resp, Err_give, "Err_give")
	}

	// 任意关卡通过次数条件
	p.updateCondition(account.COND_TYP_Any_Stage_Pass,
		leftCount, 0, "", "", resp)

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		req.GameModeId, leftCount,
		p.Profile.GetProfileNowTime())

	resp.OnChangeGameMode(req.GameModeId)
	resp.OnChangeMarketActivity()
	resp.mkInfo(p)

	// logiclog
	logiclog.LogStageFinish(p.AccountID.String(), p.Profile.GetCurrAvatar(), levelId,
		true, 0, leftCount, true, 0,
		p.Profile.GetCorp().Level, p.Profile.ChannelId, 0, [helper.DestinyGeneralSkillMax]int{},
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}
