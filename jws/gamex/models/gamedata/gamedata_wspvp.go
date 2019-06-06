package gamedata

import (
	"github.com/golang/protobuf/proto"
	"math/rand"
	"strconv"
	"strings"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var WsPvpMatchConfigs []*ProtobufGen.WSPVPMATCH
var WsPvpBestRankRewardConfigs []*ProtobufGen.WSPVPFPASS
var WsPvpTimeRewardConfigs []*ProtobufGen.WSPVPSECTOR
var WsPvpChallengeRewardConfigs []*ProtobufGen.WSBOXREWARD
var WsPvpMainCfg *WsPvpMainConfig

type WsPvpMainConfig struct {
	avatarArray   []int
	starArray     []int
	Config        *ProtobufGen.WSPVPMAIN
	LockClock     int // 分钟为单位
	UnlockClock   int
	RankLockClock int
}

// wspvp时间是会跨过夜间凌晨
func IsInWsPvpRange(nowMin int) bool {
	return nowMin >= WsPvpMainCfg.LockClock || nowMin < WsPvpMainCfg.UnlockClock
}

// 22:30 - > 10:30
func IsRankInWspvpRange(nowMin int) bool {
	return nowMin >= WsPvpMainCfg.RankLockClock || nowMin <= WsPvpMainCfg.UnlockClock
}

func loadWsPvpMatchConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.WSPVPMATCH_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	WsPvpMatchConfigs = dataList.Items
}

func loadWsPvpBestRankRewardConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.WSPVPFPASS_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	WsPvpBestRankRewardConfigs = dataList.Items
}

func loadWsPvpTimeRewardConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.WSPVPSECTOR_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	WsPvpTimeRewardConfigs = dataList.Items
}

func loadWsPvpChallengeRewardConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.WSBOXREWARD_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	WsPvpChallengeRewardConfigs = dataList.Items
}

func loadWsPvpMainConfig(filePath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filePath)
	errCheck(err)

	dataList := &ProtobufGen.WSPVPMAIN_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	cfg := dataList.GetItems()[0]
	WsPvpMainCfg = new(WsPvpMainConfig)
	idStrs := strings.Split(cfg.GetBotHeroID(), "&")
	if len(idStrs) < 9 {
		panic("id strs should be more than 9")
	}
	WsPvpMainCfg.avatarArray = make([]int, len(idStrs))
	for i, id := range idStrs {
		WsPvpMainCfg.avatarArray[i], err = strconv.Atoi(id)
		if err != nil {
			panic(err)
		}
	}
	starStrs := strings.Split(cfg.GetBotStarLevel(), "&")
	WsPvpMainCfg.starArray = make([]int, len(starStrs))
	for i, id := range starStrs {
		WsPvpMainCfg.starArray[i], err = strconv.Atoi(id)
		if err != nil {
			panic(err)
		}
	}
	WsPvpMainCfg.Config = cfg
	logs.Debug("wspvpmainconfig %s, %s", cfg.GetLockingTime(), cfg.GetUnLockTime())
	hour, min := ParseWspvpLockingClock(cfg.GetLockingTime())
	WsPvpMainCfg.LockClock = hour*60 + min
	hour, min = ParseWspvpLockingClock(cfg.GetUnLockTime())
	WsPvpMainCfg.UnlockClock = hour*60 + min
	WsPvpMainCfg.RankLockClock = WsPvpMainCfg.LockClock + 5 // 延迟5分钟
	logs.Debug("wspvpmainconfig", WsPvpMainCfg.LockClock, WsPvpMainCfg.RankLockClock, WsPvpMainCfg.UnlockClock)
}

func ParseWspvpLockingClock(timeStr string) (int, int) {
	strs := strings.Split(timeStr, ":")
	hour, err := strconv.Atoi(strs[0])
	panicIfErr(err)
	min, err := strconv.Atoi(strs[1])
	panicIfErr(err)
	return hour, min
}

func GetWSPVPMatchConfig(rank int) *ProtobufGen.WSPVPMATCH {
	for _, cfg := range WsPvpMatchConfigs {
		if uint32(rank) >= cfg.GetStart() && uint32(rank) <= cfg.GetEnd() {
			return cfg
		}
	}
	return nil
}

func GetWSPVPFourOppopent(rank int) [4]int {
	if rank <= 0 || rank > 10000 {
		rank = 10000
	}
	index := 0
	for i, cfg := range WsPvpMatchConfigs {
		if uint32(rank) >= cfg.GetStart() && uint32(rank) <= cfg.GetEnd() {
			index = i
		}
	}
	var cfgForOne, cfgForThree *ProtobufGen.WSPVPMATCH
	var n1, n2, n3, n4 uint32
	if index == 0 {
		cfgForOne = WsPvpMatchConfigs[index]
		n1 = uint32(uint32(rank) - cfgForOne.GetRandomRange())
		n2 = uint32(int32(rank) + cfgForOne.GetRandomRange2())
	} else {
		cfgForOne = WsPvpMatchConfigs[index-1]
		n1 = cfgForOne.GetStart()
		n2 = cfgForOne.GetEnd()
	}
	cfgForThree = WsPvpMatchConfigs[index]
	n3 = uint32(uint32(rank) - cfgForThree.GetRandomRange())
	n4 = uint32(int32(rank) + cfgForThree.GetRandomRange2())
	return randomFour(n1, n2+1, n3, n4+1, uint32(rank))
}

// 从n1, n2里面随机1个，从n3, n4里面随机5个
// 然后两者判重，排除n5，取4个
func randomFour(n1, n2, n3, n4, n5 uint32) [4]int {
	match := [4]int{}
	rand := util.RandomInt(int32(n1), int32(n2))
	randList := util.RandomInts(int(n3), int(n4), 5)
	index := 0
	if rand != n5 {
		match[index] = int(rand)
		index++
	}
	for _, num := range randList {
		if num != int(rand) && num != int(n5) {
			match[index] = num
			index++
			if index == 4 {
				break
			}
		}
	}
	return match
}

func GetWsPvpBestRankRewardCfg(tableId int) *ProtobufGen.WSPVPFPASS {
	return WsPvpBestRankRewardConfigs[tableId-1]
}

func GetWsPvpTimeReward(rank int) *ProtobufGen.WSPVPSECTOR {
	for _, cfg := range WsPvpTimeRewardConfigs {
		if uint32(rank) >= cfg.GetStart() && uint32(rank) <= cfg.GetEnd() {
			return cfg
		}
	}
	return nil
}

func GetWsPvpChallengeReward(tableId int) *ProtobufGen.WSBOXREWARD {
	for _, cfg := range WsPvpChallengeRewardConfigs {
		if cfg.GetWinNum() == uint32(tableId) {
			return cfg
		}
	}
	return nil
}

func Random9RobotIds() []int64 {
	resArray := util.ShuffleArray(WsPvpMainCfg.avatarArray)
	ret := make([]int64, 9)
	for i, idx := range resArray[:9] {
		ret[i] = int64(idx)
	}
	return ret
}

func RandomRobotStar() int {
	rand := rand.Int31n(int32(len(WsPvpMainCfg.starArray)))
	return WsPvpMainCfg.starArray[rand]
}

func GetSidName(etcdRoot string, gid, sid uint) string {
	displayName1 := etcd.GetSidDisplayName(etcdRoot, gid, sid)
	displayName2 := etcd.ParseDisplayShardName(displayName1)
	index := strings.Index(displayName2, "-")
	if index == -1 {
		return displayName2
	}
	return displayName2[:index]
}
