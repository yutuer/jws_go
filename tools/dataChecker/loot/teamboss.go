package loot

import (
	"fmt"
	"math/rand"
	"sort"

	"vcs.taiyouxi.net/jws/gamex/logics"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/tools/dataChecker/utils"
)

var (
	acc      *logics.Account
	reporter *utils.Reporter
)

func init() {
	account.InitDebuger()

	acc = new(logics.Account)
	acc.Account = account.Debuger.GetNewAccount()

	reporter = utils.NewReporter()
}

func simulateBoxDrop(levelId uint32, vipLevel uint32, repeatTimes int) map[string]int {
	boxes := make(map[string]int)
	controlHit := 0

	var rewardBoxId string
	bossInfo := acc.Profile.GetTeamBossStorageInfo()

	mainData := gamedata.GetTBossMainDataByDiff(levelId)
	acc.Profile.Vip.V = vipLevel
	boxCtrlCfg := gamedata.GetTBossVipCtrl(acc.Profile.Vip.V)

	// 抄袭来源：teambossHandler.go l80-l118
	for i := 0; i < repeatTimes; i++ {
		if bossInfo.BoxCtrlTimes >= boxCtrlCfg.GetGoodBoxControl() {
			bossInfo.ResetControlTimes()
			rewardBoxId = gamedata.RandomTBBox(mainData.GetSepcialDropGroup())
			controlHit++
		} else {
			rewardBoxId = gamedata.RandomTBBox(mainData.GetBoxDropGroup())
			if gamedata.IsRedOrGoldenBox(rewardBoxId) {
				bossInfo.ResetControlTimes()
			} else {
				bossInfo.IncreaseControlTimes()
			}
		}

		boxes[rewardBoxId] += 1
	}

	return boxes
}

func GetBoxesDropForAllDiff() {
	vipGroup := []uint32{0, 5, 8, 10}

	for key := range gamedata.GetTBossDiffMap() {
		for _, vipLevel := range vipGroup {
			recordLogs := make([]string, 0)
			res := simulateBoxDrop(key, vipLevel, repeatTimes)

			for boxId, count := range res {
				recordLogs = append(recordLogs, fmt.Sprintf("%v: %v", boxId, count))
			}

			sort.Strings(recordLogs)
			reporter.Record(-1, fmt.Sprintf("V%d 难度%v掉落", vipLevel, key), utils.TEAMBOSS, utils.NONE, recordLogs)
		}
	}
}

func simulateBoxOpen(boxId string, repeatTimes int) map[string]int {
	items := make(map[string]int)

	for i := 0; i < repeatTimes; i++ {
		loots := gamedata.GetTBBoxLootTableByBoxID(boxId)
		for _, item := range loots {
			if rand.Float32() < item.GetLootChance() && item.GetItemID() != "" {
				items[item.GetItemID()] += int(item.GetLootNumber())
			}
		}
	}

	return items
}

func GetAllBoxDropResults() {
	boxIds := make([]string, 0)

	// 生成箱子ID
	for diff := 1; diff < 4; diff++ {
		for rarity := 'a'; rarity < 'e'; rarity++ {
			boxIds = append(boxIds, fmt.Sprintf("TB_BOX_%d_%c", diff, rarity))
		}
	}

	// 记录掉落并输出log
	for _, boxId := range boxIds {
		lootLogs := make([]string, 0)
		items := simulateBoxOpen(boxId, repeatTimes)

		for item, count := range items {
			lootLogs = append(lootLogs, fmt.Sprintf("%v: %v", item, count))
		}

		sort.Strings(lootLogs)
		reporter.Record(-1, boxId, utils.TEAMBOSS, utils.NONE, lootLogs)
	}
}

func GetTeambossResult() {
	GetBoxesDropForAllDiff()
	GetAllBoxDropResults()
	reporter.Report()
}
