package logics

import (
	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const lootContentInitSize = 32

func mkLootDataForSweep(p *Account, stage_id string, avatar_ids []int) (bool, []StageReward) {
	res := make([]StageReward, 0, lootContentInitSize)
	es := gamedata.GetLevelEnemyConfig(stage_id)

	if es != nil {
		for _, e := range es {
			loot_templates, ok := gamedata.GetAcLoots(e.ID)
			if !ok {
				logs.Error("enemy has no loot info by %s", e.ID)
				continue
			}
			for i := 0; i < e.Count; i++ {
				for _, t := range loot_templates {
					template_id := t.GetLootTemplate()
					times := int(t.GetLootTimes())

					for ti := 0; ti < times; ti++ {
						gives, err := p.GetGivesByTemplate(template_id)
						if err != nil {
							continue
						}
						if gives.IsNotEmpty() {
							for idx, itemID := range gives.Item2Client {
								res = append(res, StageReward{
									Item_id: itemID,
									Count:   gives.Count2Client[idx],
								})
							}
						}

					}

				}
			}
		}
	} else {
		return false, res[:]
	}

	return true, res[:]
}

func isHeroDiffStage(data *gamedata.StageData) bool {
	if data != nil && (data.Type == gamedata.LEVEL_TYPE_HERODIFF_TU ||
		data.Type == gamedata.LEVEL_TYPE_HERODIFF_ZHAN ||
		data.Type == gamedata.LEVEL_TYPE_HERODIFF_HU ||
		data.Type == gamedata.LEVEL_TYPE_HERODIFF_SHI) {
		return true
	}
	return false
}

func mkLootData(p *Account, stage_id string, avatar_ids []int, resp *ResponsePrepareLootForLevelEnemy) bool {
	// 保证uint16不溢出，所以限制随机最大值为32768
	// 假设单次掉落物品不会超过32768个
	es := gamedata.GetLevelEnemyConfig(stage_id)
	logs.Trace("PrepareLootForLevelEnemy for %s, es : %v",
		stage_id, es)
	rander := p.GetRand()
	lootRand := uint16(rander.Int31n(32768))

	logs.Trace("[%s][RandRes]lootRand %d", p.AccountID, lootRand)
	lootContent := make([]loot, 0, lootContentInitSize)

	if es != nil {
		lootID := uint16(0) //必须从0开始
		result := make(map[string]interface{})
		resultN := make(map[string]int)

		for _, e := range es {
			elootb := make([][][]byte, 0, e.Count)
			resultN[e.ID] = int(e.Count)
			// prepare loot templates
			stageData := gamedata.GetStageData(stage_id)
			ar := make([]struct {
				id    string
				count int
			}, 0, 2)

			// 出奇制胜特殊逻辑处理
			if isHeroDiffStage(stageData) {
				data := gamedata.GetHeroDiffEnemyData(e.ID)
				if data == nil {
					logs.Warn("[herodiff]enemy has no loot info by %s", e.ID)
					continue
				}
				loot_templates := data.GetLoots()

				for _, t := range loot_templates {
					ar = append(ar, struct {
						id    string
						count int
					}{
						id:    t.GetLootTemplate(),
						count: int(t.GetLootTimes()),
					})
				}
			} else {
				loot_templates, ok := gamedata.GetAcLoots(e.ID)
				if !ok {
					logs.Warn("enemy has no loot info by %s", e.ID)
					continue
				}
				// 2. 根据掉落数据随机掉落
				for _, t := range loot_templates {
					ar = append(ar, struct {
						id    string
						count int
					}{
						id:    t.GetLootTemplate(),
						count: int(t.GetLootTimes()),
					})
				}
			}
			for i := 0; i < e.Count; i++ {
				elootb_ar := make([][]byte, 0, gamedata.GetAcLootsMaxTimes(e.ID))
				for _, item := range ar {
					//logs.Trace("loot template %s - %s", template_id, times)
					for ti := 0; ti < item.count; ti++ {
						gives, err := p.GetGivesByTemplate(item.id)
						if err != nil {
							continue
						}
						if gives.IsNotEmpty() {
							for idx, itemID := range gives.Item2Client {
								nloot := loot{
									ID:    lootID + lootRand,
									Data:  itemID,
									Count: gives.Count2Client[idx],
								}

								lootID++
								lootContent = append(lootContent, nloot)
								elootb_ar = append(elootb_ar, encode(&nloot))
							}
						}
					}
				}
				if len(elootb_ar) > 0 {
					//如果什么都没有掉落则不传输
					elootb = append(elootb, elootb_ar)
				}
			}
			result[e.ID] = encode(elootb)
		}
		resp.Result = result
		resp.ResultN = resultN
	} else {
		logs.Error("mkLootData stage_id %s not found", stage_id)
		return false
	}

	p.Tmp.LootRand = lootRand
	logs.Trace("lootContent %v", lootContent)

	// 加上副本id共结算用
	jdata, err := json.Marshal(lootContent_t{stage_id, lootContent})
	if err != nil {
		logs.SentryLogicCritical(p.AccountID.String(), "jdata error %s", err.Error())
		return false
	} else {
		p.Tmp.LootContent = jdata
	}

	return true
}

func mkStageRewards(p *Account, stageID string, avatarIDs []int, resp *ResponsePrepareLootForLevelEnemy) bool {
	lootTmpData := &(p.Tmp.StageRewards)
	lootTmpData.Init(8)
	// 星级计算
	player_stage_info := p.Profile.GetStage().GetStageInfo(
		gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()),
		stageID,
		p.GetRand())
	is_first_pass := player_stage_info.MaxStar == 0

	// 副本结算 发送奖励
	lootTmpData.
		AppendOther(p.sendStageLimitReward(stageID, is_first_pass)).
		AppendOther(p.sendStageRandReward(stageID))

	//logs.Warn("StageRewards %v", p.Tmp.StageRewards)

	stage_reward_data := gamedata.GetStageReward(stageID, is_first_pass)
	stage_data := gamedata.GetStageData(stageID)
	if stage_reward_data != nil && stage_data != nil {
		scReward := gamedata.NewPriceDatas(1)
		scReward.AddItem(gamedata.VI_Sc0, uint32(stage_reward_data.SCReward))
		scReward.AddItem(gamedata.VI_CorpXP, stage_reward_data.CorpXpReward)
		scReward.AddItem(gamedata.VI_XP, stage_reward_data.XpReward)
		lootTmpData.AppendData(scReward)
	}

	rewardCount := len(lootTmpData.Datas)
	res2client := make([][][]byte, rewardCount, rewardCount)
	for i := 0; i < rewardCount; i++ {
		oneRewardCount := len(lootTmpData.Datas[i].Item2Client)
		res2client[i] = make([][]byte, 0, oneRewardCount)
		for j := 0; j < oneRewardCount; j++ {
			res2client[i] = append(res2client[i], encode(
				loot{
					ID:    0,
					Data:  lootTmpData.Datas[i].Item2Client[j],
					Count: lootTmpData.Datas[i].Count2Client[j],
				}))
		}
	}
	resp.ResultS = encode(res2client)

	return true

}

func giveReward(p *Account, g *account.GiveGroup, item_id string, count uint32, item_data *gamedata.BagItemData) {
	c := gamedata.CostData{}
	if item_data != nil {
		c.AddItemWithData(item_id, *item_data, count)
	} else {
		c.AddItem(item_id, count)
	}
	g.AddCostData(&c)
}
