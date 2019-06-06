package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 根据副本Id发送Limit奖励
func (p *Account) sendStageLimitReward(stage_id string, is_first bool) *gamedata.PriceDataSet {

	rander := p.GetRand()
	acid := p.AccountID.String()

	stage_info := p.Profile.GetStage().GetStageInfo(
		gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()),
		stage_id,
		p.GetRand())
	logs.Trace("[%s]LimitReward:%s,%v %v.",
		p.AccountID,
		stage_id,
		stage_info, is_first)
	limit_rewards := gamedata.GetStageRewardLimitCfg(stage_id, is_first)
	logs.Trace("[%s]LimitReward:%v.",
		p.AccountID,
		limit_rewards)
	/*
		说明
		Limit奖励 包括
		ItemGroupID=""  物品组ID
		LootNum=1      （数量控制）掉落次数
		LootSpace	   （数量控制）随机区间
		Offset=0       （数量控制）区间偏移量
		MItemGroupID="" 补偿物品组ID

		首先会计算一个随机的区间值N 根据LootSpace和Offset随机值N
		这里表示每N次掉落中有LootNum次掉落ItemGroupID对应的物品，
		其他则掉落MItemGroupID对应的物品
	*/
	res := gamedata.NewPriceDataSet(8)
	for i := 0; i < len(limit_rewards); i++ {
		if i >= len(stage_info.Reward_state) {
			// 新添加了奖励，玩家的存储需要扩容
			stage_info.AppendRewardState()
		}

		stat := &stage_info.Reward_state[i]

		if i >= len(limit_rewards) {
			logs.SentryLogicCritical(acid, "No Limit Reward Info by %s %d",
				stage_id, i)
			continue
		}

		reward := limit_rewards[i]

		if stat.MN.IsNowNeedNewTurn() {
			stage_info.ResetReward(
				i,
				reward.Num,
				reward.Space,
				reward.Offset,
				rander)
		}

		isSpec := stat.MN.Selector(rander)

		stat.MN.LogicLog(acid,
			logiclog.LogType_StageMN,
			stage_id)

		if !isSpec {
			if reward.MItem_group_id != "" {
				res.AppendData(p.sendRewardByItemGroup(reward.MItem_group_id))
			}
		} else {
			res.AppendData(p.sendRewardByItemGroup(reward.Item_group_id))
		}
	}
	return res
}

// 根据ItemGroup发送奖励
func (p *Account) sendRewardByItemGroup(id string) gamedata.PriceDatas {
	acid := p.AccountID.String()
	gives, err := p.GetGivesByItemGroup(id)
	if err != nil {
		logs.SentryLogicCritical(acid, "sendRewardByItemGroup %s Err %s.",
			id, err.Error())
	}
	return gives
}

// 根据副本Id发送Rand奖励
func (p *Account) sendStageRandReward(stage_id string) *gamedata.PriceDataSet {

	//acid := p.AccountID.String()

	res := gamedata.NewPriceDataSet(8)
	rewards := gamedata.GetStageRewardRandCfg(stage_id)
	if rewards == nil {
		//logs.SentryLogicCritical(acid, "GetStageRewardRandCfg nil by %s", stage_id)
		return res
	}
	for idx, item_group_id := range rewards.Reward {
		if idx >= len(rewards.P) {
			logs.Error("GetStageRewardRandCfg P Err by %s %d",
				stage_id, idx)
			continue
		}

		// 这里是万分比 --> T1906
		is_send := util.RandIfTrue(p.GetRand(), int64(rewards.P[idx]), 10000)
		logs.Trace("Is Send By Rand : %v", is_send)
		if !is_send {
			continue
		}

		res.AppendData(p.sendRewardByItemGroup(item_group_id))
	}
	return res
}
