package logics

import (
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
)

// BuyGuildBossAbsentReward : 购买公会BOSS未参与的奖励
// 购买公会BOSS未参与的奖励
func (p *Account) BuyGuildBossAbsentRewardHandler(req *reqMsgBuyGuildBossAbsentReward, resp *rspMsgBuyGuildBossAbsentReward) uint32 {
	warnCode := p.CheckGuildStatus(true)
	if warnCode > 0 {
		return warnCode
	}

	isAllPassed, retCode := guild.GetModule(p.AccountID.ShardId).ActBossIsAllPassed(p.GuildProfile.GuildUUID, int(req.BossType))
	if retCode != 0 {
		return uint32(retCode)
	}
	if isAllPassed == 0 {
		// BOSS没有全部打完
		return errCode.GuildBossNotAllPassed
	}

	// 获取参数
	var leftCount, uintGB, costHc int // 剩余次数, 补偿系数, 付费补偿消耗
	if req.BossType == 0 && req.BuyType == 0 {
		leftCount, _ = p.Profile.GetCounts().Get(gamedata.CounterTypeFreeGuildBoss, p.Account)
		uintGB = int(gamedata.GetGuildBossCfg().GetOffsetMiniBossFreeGB())
	} else if req.BossType == 0 && req.BuyType == 1 {
		leftCount, _ = p.Profile.GetCounts().Get(gamedata.CounterTypeGuildBossBuyTime, p.Account)
		uintGB = int(gamedata.GetGuildBossCfg().GetOffsetMiniBossCostGB())
		costHc = leftCount * int(gamedata.GetGuildBossCfg().GetOffsetMiniBossCostHC())
	} else if req.BossType == 1 && req.BuyType == 0 {
		leftCount, _ = p.Profile.GetCounts().Get(gamedata.CounterTypeFreeGuildBigBoss, p.Account)
		uintGB = int(gamedata.GetGuildBossCfg().GetOffsetBigBossFreeGB())
	} else if req.BossType == 1 && req.BuyType == 1 {
		leftCount, _ = p.Profile.GetCounts().Get(gamedata.CounterTypeGuildBigBossBuyTime, p.Account)
		uintGB = int(gamedata.GetGuildBossCfg().GetOffsetBigBossCostGB())
		costHc = leftCount * int(gamedata.GetGuildBossCfg().GetOffsetBigBossCostHC())
	}

	if leftCount <= 0 {
		// 没有剩余次数
		return errCode.GuildBossCount
	}

	// 剩余购买次数，需要消耗HC才能获得奖励
	if costHc > 0 {
		costData := &gamedata.CostData{}
		costData.AddItem(gamedata.VI_Hc, uint32(costHc))

		reason := fmt.Sprintf("buy guild boss absent reward boss=%d, buy=%d", req.BossType, req.BuyType)
		if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
			return errCode.ClickTooQuickly
		}
	}

	// 扣掉相应次数
	if req.BossType == 0 && req.BuyType == 0 {
		p.Profile.GetCounts().UseN(gamedata.CounterTypeFreeGuildBoss, leftCount, p.Account)
		resp.OnChangeGameMode(gamedata.CounterTypeFreeGuildBoss)
	} else if req.BossType == 0 && req.BuyType == 1 {
		p.Profile.GetCounts().UseN(gamedata.CounterTypeGuildBossBuyTime, leftCount, p.Account)
		resp.OnChangeGameMode(gamedata.CounterTypeGuildBossBuyTime)
	} else if req.BossType == 1 && req.BuyType == 0 {
		p.Profile.GetCounts().UseN(gamedata.CounterTypeFreeGuildBigBoss, leftCount, p.Account)
		resp.OnChangeGameMode(gamedata.CounterTypeFreeGuildBigBoss)
	} else if req.BossType == 1 && req.BuyType == 1 {
		p.Profile.GetCounts().UseN(gamedata.CounterTypeGuildBigBossBuyTime, leftCount, p.Account)
		resp.OnChangeGameMode(gamedata.CounterTypeGuildBigBossBuyTime)
	}

	// 补偿军魂
	giveData := &gamedata.CostData{}
	giveData.AddItem(gamedata.VI_GuildBoss, uint32(leftCount*uintGB))

	reason := fmt.Sprintf("buy guild boss absent reward boss=%d, buy=%d", req.BossType, req.BuyType)
	if ok := account.GiveBySync(p.Account, giveData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	return 0
}
