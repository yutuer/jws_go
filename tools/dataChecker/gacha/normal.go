package gacha

import (
	"fmt"
	"sort"

	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/tools/dataChecker/utils"

	"github.com/golang/protobuf/proto"
)

// GetNormalGachaData 从data文件中读取gacha数据
func GetNormalGachaData() []*ProtobufGen.NORMALGACHA {
	normalGachaFilename := utils.GetDataFileFullPath("normalgacha")
	buff, err := utils.LoadBin2Buff(normalGachaFilename)
	if err != nil {
		panic(err)
	}

	normalGacha := &ProtobufGen.NORMALGACHA_ARRAY{}
	err = proto.Unmarshal(buff, normalGacha)
	if err != nil {
		panic(err)
	}

	return normalGacha.Items
}

// GetGachaSettings 从data文件中读取gacha数据
func GetGachaSettings() []*ProtobufGen.GACHASETTINGS {
	fName := utils.GetDataFileFullPath("gachasettings")
	buff, err := utils.LoadBin2Buff(fName)
	if err != nil {
		panic(err)
	}

	gachaSettings := &ProtobufGen.GACHASETTINGS_ARRAY{}
	err = proto.Unmarshal(buff, gachaSettings)
	if err != nil {
		panic(err)
	}

	return gachaSettings.Items
}

func GetAllNormalGachaLoot() {
	gachas := GetNormalGachaData()

	for _, gacha := range gachas {
		if gacha.GetLevelMax() == 200 {
			acc.Account.Profile.CorpInf.Level = 100
		} else {
			acc.Account.Profile.CorpInf.Level = 20
		}
		GetGachaLoot(gacha.GetGachaType())
	}

	reporter.Report()
}

func GetGachaLoot(gachaType uint32) {
	// gacha type 是实际减一，真坑啊……
	result := acc.DebugGetGachaReward(int(gachaType)-1, repeatTimes)
	reportLogs := make([]string, 0, len(result))

	for item, count := range result {
		reportLogs = append(reportLogs, fmt.Sprintf("%v: %v", item, count))
	}
	sort.Strings(reportLogs)

	reporter.Record(int(gachaType), fmt.Sprintf("等级：%v", acc.Account.Profile.CorpInf.Level), utils.GACHA, utils.NONE, reportLogs)
}

// GetGachaLootDetails 显示每个Gacha 10连抽的结果，重复hitRtimes次
func GetGachaLootDetails(gachaType uint32) {
	rpt := utils.NewReporter()
	var hitItem string
	var last int

	// 获得必中的东西
	gachas := GetGachaSettings()
	for _, gacha := range gachas {
		if gacha.GetGachaType() == gachaType {
			hitItem = gacha.GetItemID()
		}
	}

	for i := 0; i < hitRTimes; i++ {
		result := acc.DebugGetGachaReward(int(gachaType)-1, 1)

		lootLog := make([]string, 0)
		for item, count := range result {
			if item == hitItem {
				lootLog = append(lootLog, fmt.Sprintf("%v: %v", item, count))
			}
		}

		if len(lootLog) > 0 {
			rpt.Record(int(gachaType), fmt.Sprintf("第%d次抽中，间隔%d", i, i-last), utils.GACHA, utils.NONE, lootLog)
			last = i
		}
	}

	rpt.Report()
}
