package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdBuyConfig      [helper.Buy_Typ_Count][]PriceData
	gdBuyEStageTimes map[string][]PriceData // key: stagePolicyName
)

func loadEnergyPurchaseConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.ENERGYPURCHASE_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_EnergyBuy] = make([]PriceData, len(lv_data), len(lv_data))

	for _, c := range lv_data {
		t := int(c.GetPurchaseTime() - 1)
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", c.GetEnergyPrice())
		gdBuyConfig[helper.Buy_Typ_EnergyBuy][t] = cost_data
	}

	//logs.Trace("gdBuyConfig %v", gdBuyConfig)

}

func loadBaoZiPurchaseConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.EATBAOZICOST_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_BaoZi] = make([]PriceData, len(lv_data), len(lv_data))

	for _, c := range lv_data {
		t := int(c.GetHTPointPurchaseTime() - 1)
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", c.GetHTPointPrice())
		gdBuyConfig[helper.Buy_Typ_BaoZi][t] = cost_data
	}

	//logs.Trace("gdBuyConfig %v", gdBuyConfig)
}

func loadBossFightPointPurchaseConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.SPRINTPURCHASE_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_BossFightPoint] = make([]PriceData, len(lv_data), len(lv_data))

	for _, c := range lv_data {
		t := int(c.GetSprintPurchaseTime() - 1)
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", c.GetSprintEnergyPrice())
		gdBuyConfig[helper.Buy_Typ_BossFightPoint][t] = cost_data
	}

	//logs.Trace("gdBuyConfig %v", gdBuyConfig)
}

func loadScPurchaseConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.SCPURCHASE_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_SC] = make([]PriceData, len(lv_data), len(lv_data))

	for _, c := range lv_data {
		t := int(c.GetSCPurchaseTime() - 1)
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", c.GetSCPrice())
		gdBuyConfig[helper.Buy_Typ_SC][t] = cost_data
	}

	//logs.Trace("gdBuyConfig %v", gdBuyConfig)
}

func loadBuyEStageTimes(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.STAGEPURCHASE_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	gdBuyEStageTimes = map[string][]PriceData{}
	estageBuyTimesCount := map[string]int{}
	for _, sp := range lv_ar.GetItems() {
		count, ok := estageBuyTimesCount[sp.GetStagePurchasePolicy()]
		if !ok {
			estageBuyTimesCount[sp.GetStagePurchasePolicy()] = 1
		} else {
			estageBuyTimesCount[sp.GetStagePurchasePolicy()] = count + 1
		}
	}
	for _, sp := range lv_ar.GetItems() {
		policyName := sp.GetStagePurchasePolicy()
		if _, ok := gdBuyEStageTimes[policyName]; !ok {
			gdBuyEStageTimes[policyName] = make([]PriceData, estageBuyTimesCount[policyName])
		}
		ps := gdBuyEStageTimes[policyName]
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", sp.GetStagePrice())
		ps[sp.GetStagePurchaseTime()-1] = cost_data
	}
}

func loadTeamPvpTimesConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.TPVPTIME_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_TeamPvp] = make([]PriceData, len(lv_data), len(lv_data))
	for _, c := range lv_data {
		t := int(c.GetTPVPTime() - 1)
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", c.GetTPVPPrice())
		gdBuyConfig[helper.Buy_Typ_TeamPvp][t] = cost_data
	}
}

func loadSimplePvpTimesConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.PVPTIME_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_SimplePvp] = make([]PriceData, len(lv_data), len(lv_data))
	for _, c := range lv_data {
		t := int(c.GetPVPTime() - 1)
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", c.GetPVPPrice())
		gdBuyConfig[helper.Buy_Typ_SimplePvp][t] = cost_data
	}
}

func loadHeroTalentPointConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.HTPOINTPURCHASE_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_HeroTalentPoint] = make([]PriceData, len(lv_data), len(lv_data))
	for _, c := range lv_data {
		t := int(c.GetHTPointPurchaseTime() - 1)
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", c.GetHTPointPrice())
		gdBuyConfig[helper.Buy_Typ_HeroTalentPoint][t] = cost_data
	}
}

func GetBuyCfg(typ, count int) *PriceData {
	if typ < 0 || typ >= len(gdBuyConfig) {
		logs.Error("GetBuyCfg Err typ by %d %d", typ, count)
		return nil
	}

	if count < 0 {
		logs.Error("GetBuyCfg Err count by %d %d", typ, count)
		return nil
	}

	if count >= len(gdBuyConfig[typ]) {
		return &gdBuyConfig[typ][len(gdBuyConfig[typ])-1]
	}

	return &gdBuyConfig[typ][count]
}

func GetStageTimesBuyCfg(stageId string, count int) *PriceData {
	if count < 0 {
		logs.Error("GetStageTimesBuyCfg Err count by %s %d", stageId, count)
		return nil
	}

	policy, ok := StagesPurchasePolicy()[stageId]
	if !ok {
		logs.Error("GetStageTimesBuyCfg Err not found stage policy  by %s %d", stageId, count)
		return nil
	}

	timesPrice, ok := gdBuyEStageTimes[policy]
	if !ok {
		logs.Error("GetStageTimesBuyCfg Err not found policy data by %s %d", stageId, count)
		return nil
	}

	if count >= len(timesPrice) {
		return &timesPrice[len(timesPrice)-1]
	}

	return &timesPrice[count]
}

func loadGuildBossTimesConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.GBCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()[0]
	const NumMax = 32
	gdBuyConfig[helper.Buy_Typ_GuildBossCount] = make([]PriceData, NumMax, NumMax)
	gdBuyConfig[helper.Buy_Typ_GuildBigBossCount] = make([]PriceData, NumMax, NumMax)
	for i := 0; i < NumMax; i++ {
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", lv_data.GetLittleBossCostHC())
		cost_dataBig := PriceData{}
		cost_dataBig.AddItem("VI_HC", lv_data.GetBigBossCostHC())
		gdBuyConfig[helper.Buy_Typ_GuildBossCount][i] = cost_data
		gdBuyConfig[helper.Buy_Typ_GuildBigBossCount][i] = cost_dataBig
	}
}

func loadFestivalBossTimesConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.FBBUYCOUNT_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_FestivalBossCount] = make([]PriceData, len(lv_data), len(lv_data))
	for _, c := range lv_data {
		t := int(c.GetBuyNum() - 1)
		cost_data := PriceData{}
		cost_data.AddItem("VI_HC", c.GetCostHC())
		gdBuyConfig[helper.Buy_Typ_FestivalBossCount][t] = cost_data
	}
}

func GetFestivalBossMaxBuy() int {
	return len(gdBuyConfig[helper.Buy_Typ_FestivalBossCount])
}

func loadWsPvpRefreshConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.WSPVPREFRESH_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_WSPVP_Refresh] = make([]PriceData, len(lv_data), len(lv_data))
	for _, c := range lv_data {
		t := int(c.GetWSPVPRefreshTime() - 1)
		cost_data := PriceData{}
		cost_data.AddItem(helper.VI_Sc0, c.GetWsPvPPrice())
		gdBuyConfig[helper.Buy_Typ_WSPVP_Refresh][t] = cost_data
	}
}

func GetWSPVPRefreshMaxBuy() int {
	return len(gdBuyConfig[helper.Buy_Typ_WSPVP_Refresh])
}

func loadWsPvpChallengeConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.WSPVPTIME_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdBuyConfig[helper.Buy_Typ_WSPVP_Challenge] = make([]PriceData, len(lv_data), len(lv_data))
	for _, c := range lv_data {
		t := int(c.GetWsPvPTime() - 1)
		cost_data := PriceData{}
		cost_data.AddItem(helper.VI_Hc, c.GetWsPvPPrice())
		gdBuyConfig[helper.Buy_Typ_WSPVP_Challenge][t] = cost_data
	}
}

func GetWSPVPTimeMaxBuy() int {
	return len(gdBuyConfig[helper.Buy_Typ_WSPVP_Challenge])
}
