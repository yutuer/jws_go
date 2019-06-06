package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
	"fmt"
	"math"
	"strconv"
)

var (
	gdPvpDailyRward       []PriceDatas
	gdPvpPoolSize         []int
	gdPvpPoolRandIdxArray [][]int
	gdPvpPoolRankMin      []int
	gdPvpPoolRankMax      []int
	gdPvpSwitchCost       []PriceData
	gdPvpDayWinReward     map[int]*PriceDatas
	gdPvpWeekReward       []PriceDatas
)

func GetPvpPoolSizeData() []int {
	return gdPvpPoolSize[:]
}

func GetPvpPoolMaxRank() int {
	return gdPvpPoolRankMax[len(gdPvpPoolRankMax)-1]
}

func GetPvpDayReward(index int) (*PriceDatas, bool) {
	value, ok := gdPvpDayWinReward[index]
	return value, ok
}

func GetPvpWeekReward(rank int) *PriceDatas {
	if rank < 0 || rank >= len(gdPvpWeekReward) {
		return nil
	}
	return &gdPvpWeekReward[rank]
}

func GetPvpPoolData(rank int) (bool, int, int, []int) {
	if rank < 0 || rank >= GetPvpPoolMaxRank() {
		return false, -1, -1, nil
	}
	for i := 0; i < len(gdPvpPoolSize); i++ {
		if gdPvpPoolRankMin[i] <= rank && rank <= gdPvpPoolRankMax[i] {
			return true, i, gdPvpPoolSize[i], gdPvpPoolRandIdxArray[i]
		}
	}

	return false, -1, -1, nil
}

func GetPvpDailyRewardData(rank int) *PriceDatas {
	if rank < 0 || rank >= len(gdPvpDailyRward) {
		return nil
	}
	return &gdPvpDailyRward[rank]
}

func GetPvpSwitchCostData(index int) *PriceData {
	// TODO by ljz 返回指针好么?
	return &gdPvpSwitchCost[int(math.Min(float64(index), float64(len(gdPvpSwitchCost))-1))]
}

func loadSimplePvpRewardConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.BSCPVPRANKREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	data_len := len(dataList.GetItems()) + 1 // 从1开始

	gdPvpDailyRward = make([]PriceDatas, data_len, data_len)

	for _, a := range dataList.GetItems() {
		datas := PriceDatas{}
		for _, d := range a.GetRwdCtt() {
			datas.AddItem(
				d.GetDailyReward_ID(),
				d.GetDailyReward_Value())
		}
		gdPvpDailyRward[int(a.GetRanking())] = datas
	}
	//logs.Trace("gdPvpDailyRwardId %v", gdPvpDailyRwardId)
	//logs.Trace("gdPvpDailyRwardCount %v", gdPvpDailyRwardCount)
}

// TODO By Fanyang 读取新的表格
func loadSimplePvpPoolConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.BSCPVPPOOL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	data_len := len(dataList.GetItems()) // 从0开始

	gdPvpPoolSize = make([]int, data_len, data_len)
	gdPvpPoolRandIdxArray = make([][]int, data_len, data_len)
	gdPvpPoolRankMin = make([]int, data_len, data_len) // rank是从一开始的
	gdPvpPoolRankMax = make([]int, data_len, data_len) // rank是从一开始的

	min := 0
	for _, a := range dataList.GetItems() {
		pidx := int(a.GetPoolID() - 1)
		gdPvpPoolSize[pidx] = int(a.GetPoolLength())
		r := util.Shuffle1ToN(int(a.GetPoolLength()))
		gdPvpPoolRandIdxArray[pidx] = r
		//logs.Trace("rand %v", r)
		gdPvpPoolRankMin[pidx] = min + 1
		min += int(a.GetPoolLength())
		gdPvpPoolRankMax[pidx] = min
	}
	//logs.Trace("gdPvpPoolRandIdxArray %v", gdPvpPoolRandIdxArray)
	//logs.Trace("gdPvpPoolRankMin %v", gdPvpPoolRankMin)
	//logs.Trace("gdPvpPoolRankMin %v", gdPvpPoolRankMax)

}

func loadSimplePvpSwitchCost(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.BSCPVPSWTCOST_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	data_len := len(dataList.GetItems()) // 从0开始
	gdPvpSwitchCost = make([]PriceData, data_len, data_len)
	for i, a := range dataList.GetItems() {
		costData := PriceData{}
		costData.AddItem(a.GetCostType(), a.GetCostValue())
		gdPvpSwitchCost[i] = costData
	}
}

type SimplePvpConfig struct {
	// 在周几清榜
	ResetDay int
	// 对手失效的时间间隔
	RefreshEnemyTime int64
	// 胜方的等级分增量乘以倍数(用于决斗结算分数计算)
	WinnerScoreX float64
	// 天梯算法增量参数(用于决斗结算分数计算)
	PVPElok uint32
}

var (
	simplePvpConfigData *SimplePvpConfig
)

// 返回指针不是很友好,可能会误操作修改该数据结构,有什么其他办法么
func GetSimplePvpConfig() *SimplePvpConfig {
	return simplePvpConfigData
}

func (sp *SimplePvpConfig) GetWeekRewardResetDay() int {
	return sp.ResetDay
}

// 正常情况下不需调用,for cheat
func (sp *SimplePvpConfig) SetWeekRewardResetDay(day int) {
	sp.ResetDay = day
}
func loadSimplePvpConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.BSCPVPCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)
	gdPvpConfig := dataList.GetItems()[0]
	day, err := strconv.Atoi(gdPvpConfig.GetResetDay())
	if err != nil {
		panic(fmt.Errorf("ResetDay Err Unknown str(%s)", gdPvpConfig.GetResetDay()))
	}
	simplePvpConfigData = &SimplePvpConfig{}
	simplePvpConfigData.ResetDay = util.TimeWeekDayTranslateFromCfg(day)
	simplePvpConfigData.RefreshEnemyTime = int64(gdPvpConfig.GetRefreshOpponentTime()) * util.HourSec
	simplePvpConfigData.WinnerScoreX = float64(gdPvpConfig.GetBasicPvpWinnerBonus())
	simplePvpConfigData.PVPElok = gdPvpConfig.GetBasicPvpEloK()

}

func loadSimplePvpWeekReward(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.BSCPVPREWARDWEEK_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)
	items := dataList.GetItems()
	data_len := len(dataList.GetItems()) + 1
	gdPvpWeekReward = make([]PriceDatas, data_len, data_len)
	for _, item := range items {
		fixLoots := item.GetFixed_Loot()
		priceDatas := PriceDatas{}
		for _, loot := range fixLoots {
			priceDatas.AddItem(loot.GetFixedLootID(), loot.GetFixedLootNumber())
		}
		gdPvpWeekReward[int(item.GetRanking())] = priceDatas
	}
}

func loadSimplePvpDayWinReward(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.BSCPVPWINREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)
	items := dataList.GetItems()
	gdPvpDayWinReward = make(map[int]*PriceDatas, len(items))
	for _, item := range items {
		count := int(item.GetWinNum())
		fixLoots := item.GetFixed_Loot()
		priceDatas := &PriceDatas{}
		for _, loot := range fixLoots {
			priceDatas.AddItem(loot.GetFixedLootID(), loot.GetFixedLootNumber())
		}
		gdPvpDayWinReward[count] = priceDatas
	}
}
