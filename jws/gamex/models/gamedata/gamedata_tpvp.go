package gamedata

import (
	"math/rand"
	"sort"

	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
)

const (
	sector_rand_count      = 3
	last_sector_rand_count = 1
	SumEnemyCountInList    = sector_rand_count + last_sector_rand_count
)

var (
	gdTPvpCommonCfg       *ProtobufGen.TPVPMAIN
	gdTPvpMatch           []*MatchCfgInfo
	gdTPvpFirstPassReward map[uint32]*ProtobufGen.TPVPFPASS
	gdTPvpSectorReward    []TPvpSectorReward
	gdTPvpRewardMaxRank   uint32
	gdTPvpRankMax         uint32
	gdTPvpRefreshCost     map[uint32]*CostData
	gdTPvpTimeBalance     util.TimeToBalance
	gdTPvpDayWinReward    map[int]*PriceDatas
)

type TPvpSectorReward struct {
	Cfg    *ProtobufGen.TPVPSECTOR
	Items  []string
	Counts []uint32
}

type MatchCfgInfo struct {
	MatchCfg     *ProtobufGen.TPVPMATCH
	shuffleArray []int
	idsMap       map[uint32]struct{}
}

func (c MatchCfgInfo) GetTPvpMatchCfgShuffle(idx int) int {
	_r := c.shuffleArray[idx]
	r := _r - int(c.MatchCfg.GetRandomRange())
	if r == 0 {
		r = int(c.MatchCfg.GetRandomRange2())
	}
	return r
}

func loadTPvpMain(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.TPVPMAIN_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdTPvpCommonCfg = ar.GetItems()[0]
	gameMode := gdGameModeConfig[CounterTypeTeamPvpRefresh]
	gdTPvpTimeBalance = util.TimeToBalance{
		DailyTime: util.DailyTimeFromString(gameMode.info.GetGetTicketTime()),
	}
}

func loadTPvpMatch(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.TPVPMATCH_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	as := ar.GetItems()
	gdTPvpMatch = make([]*MatchCfgInfo, 0, len(as))
	for _, item := range as {
		if item.GetEnd() > gdTPvpRankMax {
			gdTPvpRankMax = item.GetEnd()
		}
		ids := make(map[uint32]struct{}, item.GetEnd()-item.GetStart()+1)
		for i := item.GetStart(); i <= item.GetEnd(); i++ {
			ids[i] = struct{}{}
		}
		if item.GetRandomRange() < 0 || item.GetRandomRange2() < 0 {
			panic(fmt.Errorf("loadTPvpMatch item.GetRandomRange() < 0 || "+
				"item.GetRandomRange2() < 0 %d", item.GetIndex()))
		}
		if item.GetIndex() == 1 && item.GetRandomRange() != 0 {
			panic(fmt.Errorf("loadTPvpMatch item.GetIndex() == 1 && item.GetRandomRange() != 0  err"))
		}
		gdTPvpMatch = append(gdTPvpMatch, &MatchCfgInfo{
			MatchCfg:     item,
			shuffleArray: util.Shuffle1ToN(int(item.GetRandomRange() + item.GetRandomRange2())),
			idsMap:       ids,
		})
	}
}

func loadTPvpPass(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.TPVPFPASS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	items := ar.GetItems()
	gdTPvpFirstPassReward = make(map[uint32]*ProtobufGen.TPVPFPASS, len(items))
	for _, item := range items {
		gdTPvpFirstPassReward[item.GetIndex()] = item
	}
}

func loadTPvpSector(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.TPVPSECTOR_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	items := ar.GetItems()
	gdTPvpSectorReward = make([]TPvpSectorReward, 0, len(items))
	for _, item := range items {
		items := make([]string, 0, len(item.GetPass_Loot()))
		counts := make([]uint32, 0, len(item.GetPass_Loot()))
		for _, cc := range item.GetPass_Loot() {
			items = append(items, cc.GetPassLootID())
			counts = append(counts, cc.GetPassLootNumber())
		}
		gdTPvpSectorReward = append(gdTPvpSectorReward, TPvpSectorReward{
			Cfg:    item,
			Items:  items,
			Counts: counts,
		})
		if item.GetEnd() > gdTPvpRewardMaxRank {
			gdTPvpRewardMaxRank = item.GetEnd()
		}
	}
}

func loadTPvpRefresh(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.TPVPREFRESH_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdTPvpRefreshCost = make(map[uint32]*CostData, len(ar.GetItems()))
	for _, item := range ar.GetItems() {
		c := &CostData{}
		c.AddItem(VI_Sc0, item.GetRefreshPrice())
		gdTPvpRefreshCost[item.GetTPVPRefreshTime()] = c
	}
}

func loadTPvpDayReards(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)
	dataList := &ProtobufGen.TPVPWINREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)
	items := dataList.GetItems()
	gdTPvpDayWinReward = make(map[int]*PriceDatas, len(items))
	for _, item := range items {
		count := int(item.GetWinNum())
		fixLoots := item.GetFixed_Loot()
		priceDatas := &PriceDatas{}
		for _, loot := range fixLoots {
			priceDatas.AddItem(loot.GetFixedLootID(), loot.GetFixedLootNumber())
		}
		gdTPvpDayWinReward[count] = priceDatas
	}
}

func GetTPvpCommonCfg() *ProtobufGen.TPVPMAIN {
	return gdTPvpCommonCfg
}

func GetTPvpMatchCfg() []*MatchCfgInfo {
	return gdTPvpMatch
}

func GetTPvpRankMax() uint32 {
	return gdTPvpRankMax
}

func GetTPvpRefreshCost(times uint32) *CostData {
	return gdTPvpRefreshCost[times]
}

func GetTPvpBalance() util.TimeToBalance {
	return gdTPvpTimeBalance
}

func TPvpMatchEnemies(r uint32) []int {
	res := make([]int, 0, SumEnemyCountInList)

	var curSector int
	var curMatchCfg *MatchCfgInfo
	for i, m := range gdTPvpMatch {
		if r >= m.MatchCfg.GetStart() && r <= m.MatchCfg.GetEnd() {
			curSector = i
			curMatchCfg = m
			break
		}
	}

	if curSector <= 0 { // 第一名，只从自己后面取
		ib := rand.Intn(len(curMatchCfg.shuffleArray) - SumEnemyCountInList + 1)
		for i := ib; i < ib+SumEnemyCountInList; i++ {
			res = append(res, int(r)+curMatchCfg.GetTPvpMatchCfgShuffle(i))
		}
	} else { // 非第一名，从当前段取，在从上面段取
		lastMatchCfg := gdTPvpMatch[curSector-1]
		// 当前段浮动已完全包含上一段，则一次性随机出结果
		if int32(r)-curMatchCfg.MatchCfg.GetRandomRange() <= int32(lastMatchCfg.MatchCfg.GetStart()) {
			ib := rand.Intn(len(curMatchCfg.shuffleArray) - SumEnemyCountInList + 1)
			for i := ib; i < ib+SumEnemyCountInList; i++ {
				res = append(res, int(r)+curMatchCfg.GetTPvpMatchCfgShuffle(i))
			}
		} else { // 否则分两次随机
			// 从当前段根据浮动值取
			ib := rand.Intn(len(curMatchCfg.shuffleArray) - sector_rand_count + 1)
			for i := ib; i < ib+sector_rand_count; i++ {
				res = append(res, int(r)+curMatchCfg.GetTPvpMatchCfgShuffle(i))
			}
			// 从上一段随机取一个
			for lid, _ := range lastMatchCfg.idsMap {
				if !_findDup(res, int(lid)) {
					res = append(res, int(lid))
					break
				}
			}
		}
	}
	sort.Ints(res)
	return res
}

func _findDup(ids []int, id int) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}

func GetTPvpSectorReward(rank uint32) ([]string, []uint32) {
	for _, r := range gdTPvpSectorReward {
		if rank >= r.Cfg.GetStart() && rank <= r.Cfg.GetEnd() {
			return r.Items, r.Counts
		}
	}
	return nil, nil
}

func GetTPvpFirstPassRewardCfg(idx uint32) *ProtobufGen.TPVPFPASS {
	return gdTPvpFirstPassReward[idx]
}

func GetTPvpDayReward(index int) (*PriceDatas, bool) {
	value, ok := gdTPvpDayWinReward[index]
	return value, ok

}

func GetTPvpRewardRankMax() uint32 {
	return gdTPvpRewardMaxRank
}
