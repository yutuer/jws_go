package gamedata

import (
	"math/rand"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	StrBattleStart = "2017/3/1 5:00"
	DayOffset      = -5 * 60 * 60 //每天5点刷次数和切换战场, 所以判断时间戳属于哪一天时, +DayOffset等于偏移-5小时

	CSRobTitleWeekReward = "QIANGLIANGSHI"
)

var CSRobRCConfig *ProtobufGen.RCCONFIG
var CSRobRefreshCrops []*ProtobufGen.REFRESHCROPS
var CSRobCropGift []*ProtobufGen.CROPSGIFT
var CSRobGiftLoot []*ProtobufGen.GIFTLOOT
var CSRobBattleHero []*ProtobufGen.BATTLEHERO

type CSRobConfig struct {
	DailyStartTime int64
	DailyEndTime   int64

	BestGrade uint32

	BattleStart int64
	DayOffset   int64

	NatList  []uint32
	NatCheck map[uint32]bool
}

var gCSRobConfig CSRobConfig

func loadCSRobRCConfigConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.RCCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("CSRob RCCONFIG_ARRAY empty")
	}

	CSRobRCConfig = dataList.Items[0]

	//解析需要解析的部分
	if 2 == len(strings.Split(CSRobRCConfig.GetStartTime(), ":")) &&
		2 == len(strings.Split(CSRobRCConfig.GetEndTime(), ":")) {
		prefix := "2017/3/29 "
		zero_time, err := time.ParseInLocation("2006/1/2 15:04", prefix+"0:00", util.ServerTimeLocal)
		errCheck(err)
		start_time, err := time.ParseInLocation("2006/1/2 15:04", prefix+CSRobRCConfig.GetStartTime(), util.ServerTimeLocal)
		errCheck(err)
		end_time, err := time.ParseInLocation("2006/1/2 15:04", prefix+CSRobRCConfig.GetEndTime(), util.ServerTimeLocal)
		errCheck(err)
		gCSRobConfig.DailyStartTime = start_time.Unix() - zero_time.Unix()
		gCSRobConfig.DailyEndTime = end_time.Unix() - zero_time.Unix()

		logs.Debug("[CSRob] CSRobConfig %v", gCSRobConfig)
	} else {
		panic("CSRob RCCONFIG Invalid time string")
	}

	//计算战场轮询的起点
	battleStart, err := time.ParseInLocation("2006/1/2 15:04", StrBattleStart, util.ServerTimeLocal)
	errCheck(err)
	gCSRobConfig.BattleStart = battleStart.Unix()

	//计算当日检查的偏移量
	gCSRobConfig.DayOffset = DayOffset

	logs.Debug("now [%d] starttime [%d] endtime[%d]", time.Now().Unix(), CSRobTodayStartTime(), CSRobTodayEndTime())
}

func loadCSRobRefreshCropsConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.REFRESHCROPS_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("CSRob REFRESHCROPS_ARRAY empty")
	}

	CSRobRefreshCrops = dataList.Items
}

func loadCSRobCropGiftConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.CROPSGIFT_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("CSRob CROPSGIFT_ARRAY empty")
	}

	CSRobCropGift = dataList.Items

	for _, item := range CSRobCropGift {
		if item.GetCropsID() > gCSRobConfig.BestGrade {
			gCSRobConfig.BestGrade = item.GetCropsID()
		}
	}
}

func loadCSRobGiftLootConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.GIFTLOOT_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("CSRob GIFTLOOT_ARRAY empty")
	}

	CSRobGiftLoot = dataList.Items
}

func loadCSRobBattleHeroConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.BATTLEHERO_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	if 0 == len(dataList.Items) {
		panic("CSRob BATTLEHERO_ARRAY empty")
	}

	CSRobBattleHero = dataList.Items

	gCSRobConfig.NatList = []uint32{}
	gCSRobConfig.NatCheck = map[uint32]bool{}
	for _, cfg := range CSRobBattleHero {
		gCSRobConfig.NatList = append(gCSRobConfig.NatList, cfg.GetHeroID())
		gCSRobConfig.NatCheck[cfg.GetHeroID()] = true
	}
}

func CSRobShowEnemiesLimit() uint32 {
	return CSRobRCConfig.GetEnemyAmount()
}

func CSRobShowGuildEnemiesLimit() uint32 {
	return CSRobRCConfig.GetOpposeGuildAmount()
}

func CSRobSkipBuildCost() uint32 {
	return CSRobRCConfig.GetQuickCropsCost()
}

func CSRobRobLimit() uint32 {
	return CSRobRCConfig.GetCountLimit()
}

func CSRobRankWeekDay() uint32 {
	return CSRobRCConfig.GetRankDay()
}

func CSRobDailyStartTime() int64 {
	return gCSRobConfig.DailyStartTime
}

func CSRobDailyEndTime() int64 {
	return gCSRobConfig.DailyEndTime
}

func CSRobTodayStartTime() int64 {
	return util.DailyBeginUnix(time.Now().Unix()) + gCSRobConfig.DailyStartTime
}

func CSRobTodayEndTime() int64 {
	return util.DailyBeginUnix(time.Now().Unix()) + gCSRobConfig.DailyEndTime
}

func CSRobThisWeekRankTime(now time.Time) int64 {
	now_t := now.Unix()
	weekDay := int(CSRobRankWeekDay())
	return util.DailyBeginUnix(now_t) + gCSRobConfig.DailyEndTime + int64(((weekDay-util.GetWeek(now_t))%7)*util.DaySec)
}

func CSRobNextWeekStartTime(now time.Time) int64 {
	now_t := now.Unix()
	weekDay := int(CSRobRankWeekDay())
	return util.DailyBeginUnix(now_t) - DayOffset + int64((((weekDay-util.GetWeek(now_t))%7)+1)*util.DaySec)
}

func CSRobBestGrade() uint32 {
	return gCSRobConfig.BestGrade
}

//CSRobGradeFirstRefresh ..
func CSRobGradeFirstRefresh(rf float32) uint32 {
	if 0 == len(CSRobRefreshCrops) {
		return 1
	}

	tRf := rf
	for _, elem := range CSRobRefreshCrops[0].GetFixed_Loot() {
		if tRf < elem.GetRefreshProbability() {
			return elem.GetRefreshQuality()
		}
		tRf -= elem.GetRefreshProbability()
	}

	return 1
}

//CSRobGradeRefresh param: 次数  return: 概率,消耗
func CSRobGradeRefresh(count uint32, curGrade uint32) (float32, uint32) {
	index := int(count)
	if index >= len(CSRobRefreshCrops) {
		index = len(CSRobRefreshCrops) - 1
	}
	cost := CSRobRefreshCrops[index].GetRefreshCost()
	prob := float32(0)
	for _, elem := range CSRobRefreshCrops[index].GetFixed_Loot() {
		if elem.GetRefreshQuality() == curGrade+1 {
			prob = elem.GetRefreshProbability()
		}
	}
	return prob, cost
}

func CSRobGradeTopEdge() uint32 {
	return CSRobRCConfig.GetMaxRefresh()
}

func CSRobGiftConfig(grade uint32) *ProtobufGen.CROPSGIFT {
	for _, gift := range CSRobCropGift {
		if gift.GetCropsID() == grade {
			return gift
		}
	}

	return CSRobCropGift[len(CSRobCropGift)-1]
}

func CSRobLootConfig(lootID uint32) []*ProtobufGen.GIFTLOOT {
	list := []*ProtobufGen.GIFTLOOT{}
	for _, loot := range CSRobGiftLoot {
		if loot.GetGiftID() == lootID {
			list = append(list, loot)
		}
	}

	return list
}

func CSRobRewardForDriver(history map[string]uint32, grade uint32, sub uint32) (newHistory map[string]uint32, goods map[string]uint32, dark bool) {
	newHistory = make(map[string]uint32)
	goods = make(map[string]uint32)

	for id, num := range history {
		newHistory[id] = num
	}

	gift := CSRobGiftConfig(grade)

	lootID1 := gift.GetCropsGift1ID()
	lootNum1 := gift.GetCropsGift1Amount()
	if lootNum1 > sub {
		lootNum1 -= sub
	} else {
		lootNum1 = 0
	}
	lootList1 := CSRobLootConfig(lootID1)
	newHistory, rg, d := randFromLoot(lootList1, lootNum1, newHistory)
	for id, num := range rg {
		goods[id] += num
	}
	if true == d {
		dark = d
	}

	lootID2 := gift.GetCropsGift2ID()
	lootNum2 := gift.GetCropsGift2Amount()
	if lootNum2 > sub {
		lootNum2 -= sub
	} else {
		lootNum2 = 0
	}
	lootList2 := CSRobLootConfig(lootID2)
	newHistory, rg, d = randFromLoot(lootList2, lootNum2, newHistory)
	for id, num := range rg {
		goods[id] += num
	}
	if true == d {
		dark = d
	}

	return
}

func CSRobRewardForRob(history map[string]uint32, grade uint32) (newHistory map[string]uint32, goods map[string]uint32, dark bool) {
	newHistory = make(map[string]uint32)
	goods = make(map[string]uint32)

	for id, num := range history {
		newHistory[id] = num
	}

	gift := CSRobGiftConfig(grade)

	lootID1 := gift.GetRobsGift1ID()
	lootNum1 := gift.GetRobGift1Amount()
	lootList1 := CSRobLootConfig(lootID1)
	newHistory, rg, d := randFromLoot(lootList1, lootNum1, newHistory)
	for id, num := range rg {
		goods[id] += num
	}
	if true == d {
		dark = d
	}

	lootID2 := gift.GetRobsGift2ID()
	lootNum2 := gift.GetRobGift2Amount()
	lootList2 := CSRobLootConfig(lootID2)
	newHistory, rg, d = randFromLoot(lootList2, lootNum2, newHistory)
	for id, num := range rg {
		goods[id] += num
	}
	if true == d {
		dark = d
	}

	return
}

func CSRobRewardForHelp(history map[string]uint32, grade uint32) (newHistory map[string]uint32, goods map[string]uint32, dark bool) {
	newHistory = make(map[string]uint32)
	goods = make(map[string]uint32)

	for id, num := range history {
		newHistory[id] = num
	}

	gift := CSRobGiftConfig(grade)

	lootID1 := gift.GetEscortGift1ID()
	lootNum1 := gift.GetEscortGift1Amount()
	lootList1 := CSRobLootConfig(lootID1)
	newHistory, rg, d := randFromLoot(lootList1, lootNum1, newHistory)
	for id, num := range rg {
		goods[id] += num
	}
	if true == d {
		dark = d
	}

	return
}

func randFromLoot(loots []*ProtobufGen.GIFTLOOT, num uint32, history map[string]uint32) (newHistory map[string]uint32, goods map[string]uint32, dark bool) {
	goods = map[string]uint32{}
	newHistory = make(map[string]uint32)
	dark = false

	for id, num := range history {
		newHistory[id] = num
	}

	for i := uint32(0); i < num; i++ {
		sumRand := float32(0)
		randPool := []*ProtobufGen.GIFTLOOT{}
		for _, loot := range loots {
			if loot.GetMaxDailyLoot() <= newHistory[loot.GetRandomLootID()] {
				dark = true
				continue
			}
			randPool = append(randPool, loot)
			sumRand += loot.GetRandomLootWeight()
		}

		randRet := rand.Float32() * sumRand
		for _, loot := range randPool {
			if randRet < loot.GetRandomLootWeight() {
				n := loot.GetRandomLootAmount()
				if n > loot.GetMaxDailyLoot()-newHistory[loot.GetRandomLootID()] {
					n = loot.GetMaxDailyLoot() - newHistory[loot.GetRandomLootID()]
					dark = true
				}
				newHistory[loot.GetRandomLootID()] += n
				goods[loot.GetRandomLootID()] += n
				break
			} else {
				randRet -= loot.GetRandomLootWeight()
			}
		}
	}

	return newHistory, goods, dark
}

func CSRobBattleIDAndHeroID(now int64) (uint32, uint32) {
	num := len(CSRobBattleHero)
	index := int((now-gCSRobConfig.BattleStart)/util.DaySec) % num
	return CSRobBattleHero[index].GetBattleID(), CSRobBattleHero[index].GetHeroID()
}

func CSRobBattleIDList() []uint32 {
	list := []uint32{}

	for _, c := range CSRobBattleHero {
		list = append(list, c.GetBattleID())
	}

	return list
}

//CSRobNatList ..
func CSRobNatList() []uint32 {
	return gCSRobConfig.NatList
}

//CSRobNatCheck ..
func CSRobNatCheck(nat uint32) bool {
	return gCSRobConfig.NatCheck[nat]
}

func CSRobCheckSameDay(ta int64, tb int64) bool {
	return util.IsSameDayUnix(ta+gCSRobConfig.DayOffset, tb+gCSRobConfig.DayOffset)
}

func CSRobCheckSameWeek(ta int64, tb int64) bool {
	return util.IsSameWeekUnix(ta+gCSRobConfig.DayOffset, tb+gCSRobConfig.DayOffset)
}

func CSRobCheckSameDayAndHour(ta int64, tb int64) bool {
	if false == util.IsSameDayUnix(ta+gCSRobConfig.DayOffset, tb+gCSRobConfig.DayOffset) {
		return false
	}
	return util.DailyBeginUnix(ta)/util.HourSec == util.DailyBeginUnix(tb)/util.HourSec
}

func CSRobMarqueeCost() uint32 {
	return gdGankConf.GetBroadcastCost()
}

func CSRobGuildListRefreshOffset() int64 {
	return 0 - DayOffset + 5*util.MinSec
}

func CSRobAppealLimit() uint32 {
	return CSRobRCConfig.GetRescuetimes()
}

func CSRobRobTimeout() int64 {
	return int64(CSRobRCConfig.GetLevelTime())
}

func CSRobBuildCarOffsetTime() int64 {
	return int64(CSRobRCConfig.GetDelayTime())
}

func CSRobRecordTrim() int {
	return int(CSRobRCConfig.GetMaxLogAmount())
}

func CSRobJoinLevelLimit() uint32 {
	return CSRobRCConfig.GetLevelLimit()
}
