package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type HeroGachaRaceInfo struct {
	Rank       int64                          `json:"rank"`
	Score      int64                          `json:"score"`
	ChestInfo  [gamedata.HGR_Chest_Count]bool `json:"chest_info"`
	ActivityID int64                          `json:"act_id"`
}

func (h *HeroGachaRaceInfo) GetCurScore() int64 {
	return h.Score
}

func (h *HeroGachaRaceInfo) AddCurScore(newScore int64) {
	h.Score += newScore
}

func (h *HeroGachaRaceInfo) GetChestInfo() [gamedata.HGR_Chest_Count]bool {
	return h.ChestInfo
}

func (h *HeroGachaRaceInfo) SetChestInfo(index int, flag bool) {
	if index >= 0 && index < len(h.ChestInfo) {
		h.ChestInfo[index] = flag
	}
}

func (h *HeroGachaRaceInfo) CheckActivity(acID string, nowActivityID int64) {
	// 检查上次活动是否得到了宝箱奖励,并发送邮件
	if h.ActivityID != 0 && h.ActivityID != nowActivityID {
		// 发送限时名将未领取宝箱奖励邮件
		rewardMap := make(map[string]uint32, 10)
		for index, info := range h.ChestInfo {
			needScore, had := gamedata.GetHotDatas().HotLimitHeroGachaData.GetHeroGachaRaceChestInfo(h.ActivityID, uint32(index))
			if !had {
				logs.Debug("HeroGachaRace activity or index not valid, id is %d and index is %d", h.ActivityID, index)
				continue
			}
			if uint32(h.Score) >= needScore && info == false {
				rewards := gamedata.GetHotDatas().HotLimitHeroGachaData.GetHeroGachaRaceChestReward(h.ActivityID, uint32(index))
				for _, reward := range rewards {
					id := reward.GetItemID()
					count := reward.GetItemNum()
					value, ok := rewardMap[id]
					if ok {
						rewardMap[id] = value + count
					} else {
						rewardMap[id] = count
					}
				}
			}
		}
		logs.Debug("HeroGachaRace Mail Chest Rewards Len= %d", len(rewardMap))
		account, _ := db.ParseAccount(acID)
		if len(rewardMap) != 0 && game.Cfg.GetHotActValidData(account.ShardId, uutil.Hot_Value_Limit_Hero) {
			mail_sender.SendHeroGachaRaceChestMail(acID, rewardMap)
			logs.Warn("Send HeroGachaRaceChest Mail")
		}
		h.ResetInfo()
	}

	// 更新活动ID
	h.ActivityID = nowActivityID
}

// 重置
func (h *HeroGachaRaceInfo) ResetInfo() {
	h.Rank = 0
	h.Score = 0
	h.ChestInfo = [gamedata.HGR_Chest_Count]bool{}
	h.ActivityID = 0
}
