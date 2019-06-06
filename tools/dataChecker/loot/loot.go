package loot

import (
	"fmt"
	"sort"

	"vcs.taiyouxi.net/jws/gamex/logics"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/tools/dataChecker/utils"

	"github.com/golang/protobuf/proto"
)

var (
	repeatTimes = 100000 // 默认跑十万次，大概耗时15-25分钟
)

func init() {
	logs.Close()
}

// GetLevelInfoData 从data文件中读取并生成LEVEL_INFO
func GetLevelInfoData() []*ProtobufGen.LEVEL_INFO {
	levelInfoFilename := utils.GetDataFileFullPath("level_info")
	buff, err := utils.LoadBin2Buff(levelInfoFilename)
	if err != nil {
		panic(err)
	}

	levelInfo := &ProtobufGen.LEVEL_INFO_ARRAY{}
	err = proto.Unmarshal(buff, levelInfo)
	if err != nil {
		panic(err)
	}

	return levelInfo.Items
}

// GetAllLevelLootInfo 获得所有关卡的掉落并打印
func GetAllLevelLootInfo() {

	reporter := utils.NewReporter()
	account.InitDebuger()

	acc := new(logics.Account)
	acc.Account = account.Debuger.GetNewAccount()

	levelData := GetLevelInfoData()
	for _, level := range levelData {
		if len(level.GetDropItem_Template()) > 0 {
			loot := acc.DebugGetLevelLimitLootSummary(level.GetLevelID(), repeatTimes)

			lootLog := make([]string, 0, len(loot))
			for itemName, count := range loot {
				lootLog = append(lootLog, fmt.Sprintf("%v: %v", itemName, count))
			}
			sort.Strings(lootLog)

			reporter.Record(-1, level.GetLevelID(), utils.LOOT, utils.NONE, lootLog)
		}
	}

	reporter.Report()
}
