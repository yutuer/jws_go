package gamedata

import (
	"fmt"
	"strconv"
	"strings"

	"sort"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
)

var (
	gdSevOpnRank         *ProtobufGen.DAYRANK
	gdSevOpenTimeBalance util.TimeToBalance
	gdRankAward          map[uint32]*ProtobufGen.RANKAWARD
	gdFightAward         sliceFightAward
	gdGuildRankAward     map[uint32]*ProtobufGen.GUILDRANK
	gdGuildRankLeadAward map[uint32]*ProtobufGen.GUILDLEAD
	gdGuildRankCount     int
)

func loadSevOpnRankConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.DAYRANK_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdSevOpnRank = dataList.Items[0]
	gdSevOpenTimeBalance = util.TimeToBalance{
		DailyTime: util.DailyTimeFromString(gdSevOpnRank.GetAwardTime()),
	}

	// check
	if err, _, _ := CalcSevOpnRankEndTime(0); err != nil {
		panic(err)
	}
}

func loadSevOpnRankAwardConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.RANKAWARD_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdRankAward = make(map[uint32]*ProtobufGen.RANKAWARD, len(dataList.Items))
	for _, item := range dataList.Items {
		gdRankAward[item.GetRank()] = item
	}
}

func loadSevOpnFightAwardConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.FIGHTAWARD_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdFightAward = make([]fightAward, 0, len(dataList.Items))
	for _, item := range dataList.Items {
		gdFightAward = append(gdFightAward, fightAward{
			gs:  item.GetFight(),
			cfg: item,
		})
	}
	sort.Sort(gdFightAward)
}

func loadSevOpnGuildRankConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GUILDRANK_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGuildRankAward = make(map[uint32]*ProtobufGen.GUILDRANK, len(dataList.Items))
	for _, item := range dataList.Items {
		gdGuildRankAward[item.GetGuildRank()] = item
	}
	gdGuildRankCount = len(gdGuildRankAward)
}

func loadSevOpnGuildRankLeadConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GUILDLEAD_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGuildRankLeadAward = make(map[uint32]*ProtobufGen.GUILDLEAD, len(dataList.Items))
	for _, item := range dataList.Items {
		gdGuildRankLeadAward[item.GetGuildRank()] = item
	}
}

func GetSevOpnRankConfg() *ProtobufGen.DAYRANK {
	return gdSevOpnRank
}

func GetSevOpnRankTimeBalance() util.TimeToBalance {
	return gdSevOpenTimeBalance
}

func CalcSevOpnRankEndTime(startTime int64) (err error, endTime, closeTime int64) {
	st := util.DailyBeginUnix(startTime)
	ss := strings.Split(gdSevOpnRank.GetAwardTime(), ":")
	if len(ss) < 2 {
		return fmt.Errorf("server open rank: AwardTime cfg err %s",
			gdSevOpnRank.GetAwardTime()), 0, 0
	}
	h, err := strconv.Atoi(ss[0])
	if err != nil {
		return err, 0, 0
	}
	m, err := strconv.Atoi(ss[1])
	if err != nil {
		return err, 0, 0
	}
	endTime = st +
		int64(gdSevOpnRank.GetAwardDay()-1)*util.DaySec +
		int64(h)*util.HourSec + int64(m)*util.MinSec
	closeTime = st + int64(gdSevOpnRank.GetOpenDay())*util.DaySec
	return
}

func GetRankAward(rank uint32, gs int64) *ProtobufGen.RANKAWARD {
	if gs < int64(gdSevOpnRank.GetRankAward()) {
		return nil
	}
	return gdRankAward[rank]
}

func GetFightAward(gs int64) *ProtobufGen.FIGHTAWARD {
	for _, fa := range gdFightAward {
		if gs >= int64(fa.gs) {
			return fa.cfg
		}
	}
	return nil
}

func GetGuildRankAwardCfg(rank int) (*ProtobufGen.GUILDRANK, *ProtobufGen.GUILDLEAD) {
	if rank > len(gdGuildRankAward) {
		return nil, nil
	}
	return gdGuildRankAward[uint32(rank)], gdGuildRankLeadAward[uint32(rank)]
}

func GetGuildRankCount() int {
	return gdGuildRankCount
}

type fightAward struct {
	gs  uint32
	cfg *ProtobufGen.FIGHTAWARD
}

type sliceFightAward []fightAward

func (p sliceFightAward) Len() int           { return len(p) }
func (p sliceFightAward) Less(i, j int) bool { return p[i].gs > p[j].gs } // 逆序
func (p sliceFightAward) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
