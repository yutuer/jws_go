package gamedata

import (
	"time"

	"math/rand"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var worldBossData WorldBossData

type WorldBossData struct {
	config       *ProtobufGen.WBOSSCONFIG
	rankRewards  []*ProtobufGen.WBRANKREWARD
	damageReards []*ProtobufGen.WBDEMAGEREWARD
	bossData     []*ProtobufGen.WBOSSDATA
	bossLevel    []*ProtobufGen.WBOSSLEVEL
	maxRankLimit uint32
}

func loadWBConfigData(filepath string) {
	ar := &ProtobufGen.WBOSSCONFIG_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	worldBossData.config = data[0]
}

func loadWBRankRewardData(filepath string) {
	ar := &ProtobufGen.WBRANKREWARD_ARRAY{}
	_common_load(filepath, ar)
	worldBossData.rankRewards = ar.GetItems()
	for _, item := range worldBossData.rankRewards {
		if item.GetEnd() > worldBossData.maxRankLimit {
			worldBossData.maxRankLimit = item.GetEnd()
		}
	}
}

func loadWBDamageRewardData(filepath string) {
	ar := &ProtobufGen.WBDEMAGEREWARD_ARRAY{}
	_common_load(filepath, ar)
	worldBossData.damageReards = ar.GetItems()
}

func loadWBBossData(filepath string) {
	ar := &ProtobufGen.WBOSSDATA_ARRAY{}
	_common_load(filepath, ar)
	worldBossData.bossData = ar.GetItems()
}

func loadWBBossLevelData(filepath string) {
	ar := &ProtobufGen.WBOSSLEVEL_ARRAY{}
	_common_load(filepath, ar)
	worldBossData.bossLevel = ar.GetItems()
}

func GetWBConfig() *ProtobufGen.WBOSSCONFIG {
	return worldBossData.config
}

func GetWBRankRewards(rank uint32) []*ProtobufGen.WBRANKREWARD_LootRule {
	for _, item := range worldBossData.rankRewards {
		if rank <= item.GetEnd() && rank >= item.GetStart() {
			return item.GetLoot_Table()
		}
	}
	logs.Warn("no worldboss reward for rank: %v", rank)
	return nil
}

func GetWBDamageRewards(id uint32) *ProtobufGen.WBDEMAGEREWARD {
	for _, item := range worldBossData.damageReards {
		if item.GetID() == id {
			return item
		}
	}
	return nil
}

func GetWBDamageRewardsWithDmg(damage uint64) []*ProtobufGen.WBDEMAGEREWARD {
	ret := make([]*ProtobufGen.WBDEMAGEREWARD, 0)
	for _, item := range worldBossData.damageReards {
		if item.GetNeedDemage() <= uint32(damage) {
			ret = append(ret, item)
		}
	}
	return ret
}

func GetWBBossCfg(bossLevel uint32) *ProtobufGen.WBOSSDATA {
	for _, item := range worldBossData.bossData {
		if item.GetBossLevel() == bossLevel {
			return item
		}
	}
	logs.Warn("no world boss cfg for level: %v", bossLevel)
	return nil
}

func GetMaxRankLimit() uint32 {
	return worldBossData.maxRankLimit
}

//GetWBBossHP ..
func GetWBBossHP(bossLevel uint32) int64 {
	for _, item := range worldBossData.bossData {
		if item.GetBossLevel() == bossLevel {
			return item.GetHitPoint()
		}
	}
	logs.Warn("no world boss cfg for level: %v", bossLevel)
	ret := worldBossData.bossData[len(worldBossData.bossData)-1]
	return ret.GetHitPoint()
}

func GetNextWBBoss(bossLv uint32) (nextLv uint32, nextHP int64) {
	logs.Debug("get next wb boss lv for LV.%d", bossLv)
	for _, item := range worldBossData.bossData {
		if item.GetBossLevel() == bossLv+1 {
			return item.GetBossLevel(), item.GetHitPoint()
		}
	}
	for _, item := range worldBossData.bossData {
		if item.GetBossLevel() == bossLv {
			return item.GetBossLevel(), item.GetHitPoint()
		}
	}
	// no boss data return last
	ret := worldBossData.bossData[len(worldBossData.bossData)-1]
	return ret.GetBossLevel(), ret.GetHitPoint()
}

func GetTodayWBBoss(nowT time.Time) (bossID, sceneID string) {
	weekDay := nowT.Weekday()
	if weekDay == 0 {
		weekDay = 7
	}
	boss := getWBBoss(uint32(weekDay))
	rd := uint32(rand.Int31n(100))
	c := uint32(0)
	for _, item := range boss.GetChoose_Table() {
		c = c + item.GetChooseChance()
		if c > rd {
			return item.GetWbossID(), item.GetLevelInfoID()
		}

	}
	logs.Error("data chance err for rd: %d", rd)
	return

}

func GetTodayWBResetTime(nowT time.Time) int64 {
	cfg := GetWBConfig()
	_bt, err := time.ParseInLocation("2006-1-2 15:04",
		fmt.Sprintf("%d-%d-%d %s", nowT.Year(), nowT.Month(), nowT.Day(),
			cfg.GetResetTime()), util.ServerTimeLocal)
	if err != nil {
		logs.Error("GetTodayResetTime time.ParseInLocation err %v", err)
		return 0
	}
	return _bt.Unix()
}

func GetTodayWBStartTime(nowT time.Time) int64 {
	cfg := GetWBConfig()
	_bt, err := time.ParseInLocation("2006-1-2 15:04",
		fmt.Sprintf("%d-%d-%d %s", nowT.Year(), nowT.Month(), nowT.Day(),
			cfg.GetStartTime()), util.ServerTimeLocal)
	if err != nil {
		logs.Error("GetTodayStartTime time.ParseInLocation err %v", err)
		return 0
	}
	return _bt.Unix()
}

func GetTodayWBEndTime(nowT time.Time) int64 {
	cfg := GetWBConfig()
	_bt, err := time.ParseInLocation("2006-1-2 15:04",
		fmt.Sprintf("%d-%d-%d %s", nowT.Year(), nowT.Month(), nowT.Day(),
			cfg.GetEndTime()), util.ServerTimeLocal)
	if err != nil {
		logs.Error("GetTodayEndTime time.ParseInLocation err %v", err)
		return 0
	}
	return _bt.Unix()
}

func GetTodayWBRewardTime(nowT time.Time) int64 {
	cfg := GetWBConfig()
	_bt, err := time.ParseInLocation("2006-1-2 15:04",
		fmt.Sprintf("%d-%d-%d %s", nowT.Year(), nowT.Month(), nowT.Day(),
			cfg.GetRewardTime()), util.ServerTimeLocal)
	if err != nil {
		logs.Error("GetTodayRewardTime time.ParseInLocation err %v", err)
		return 0
	}
	return _bt.Unix()
}

func GetWBBattleValidTime() uint32 {
	cfg := GetWBConfig()
	return cfg.GetSeverFightTime()
}

func GetTopLength() uint32 {
	cfg := GetWBConfig()
	return cfg.GetRankListNum()
}

func GetKillBossRewardRankLimit() uint32 {
	cfg := GetWBConfig()
	return cfg.GetKillBossRank()
}

func GetTodayValidHero(nowT time.Time) uint32 {
	weekDay := nowT.Weekday()
	if weekDay == 0 {
		weekDay = 7
	}
	for _, item := range worldBossData.bossLevel {
		if item.GetDateID() == uint32(weekDay) {
			return item.GetLineUp()
		}
	}
	logs.Error("no wb boss data for weekDay: %v", weekDay)
	return 0
}

// param day from 1 to 7
func getWBBoss(day uint32) *ProtobufGen.WBOSSLEVEL {
	for _, item := range worldBossData.bossLevel {
		if item.GetDateID() == day {
			return item
		}
	}
	logs.Warn("no boss for day: %v", day)
	return nil
}
