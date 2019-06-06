package gamedata

import (
	"fmt"
	"math"
	"time"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type gEEnemyGiftData struct {
	PointNeed int
	Gifts     PriceDatas
}

var (
	gdGECommonConfig   *ProtobufGen.GECONFIG
	gdGEEnemyGroupInfo []*ProtobufGen.GEENEMYGROUP
	gdGEEnemy          map[string]*ProtobufGen.GEENEMY
	gdGEEnemyLoot      map[string]*ProtobufGen.GELOOT
	gdGEEnemyGift      []gEEnemyGiftData
	gdGEStartTime      string
	gdGEPlayTime       uint32
	gdGEWaitTime       uint32
)

func loadGatesEnemyConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GECONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGECommonConfig = dataList.Items[0]
	gdGEStartTime = gdGECommonConfig.GetGEStartTime()
	gdGEPlayTime = gdGECommonConfig.GetGEPlayTime()
	gdGEWaitTime = gdGECommonConfig.GetGEWaitTime()
}

func loadGatesEnemyGroup(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GEENEMYGROUP_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGEEnemyGroupInfo = make([]*ProtobufGen.GEENEMYGROUP, len(dataList.Items), len(dataList.Items))
	for _, e := range dataList.Items {
		gdGEEnemyGroupInfo[int(e.GetEnemyGroupID())] = e
	}
}

func loadGatesEnemyEnemy(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GEENEMY_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGEEnemy = make(map[string]*ProtobufGen.GEENEMY, len(dataList.Items))
	for _, e := range dataList.Items {
		gdGEEnemy[e.GetBossID()] = e
	}
}

func loadGatesEnemyLoot(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GELOOT_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGEEnemyLoot = make(map[string]*ProtobufGen.GELOOT, len(dataList.Items))
	for _, e := range dataList.Items {
		gdGEEnemyLoot[e.GetEGLevelID()] = e
	}
}

func loadGatesEnemyGift(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GEGIFT_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGEEnemyGift = make([]gEEnemyGiftData, 0, len(dataList.Items))
	for _, e := range dataList.Items {
		d := gEEnemyGiftData{}
		d.PointNeed = int(e.GetGuildGEPoint())
		for _, g := range e.GetFixed_Loot() {
			d.Gifts.AddItem(g.GetGiftID(), g.GetGiftNumber())
		}
		gdGEEnemyGift = append(gdGEEnemyGift, d)
	}
}

func GetGEConfig() *ProtobufGen.GECONFIG {
	return gdGECommonConfig
}

func GetGETime(now_t int64) (s, e int64) {
	t := time.Unix(now_t, 0)
	t = t.In(util.ServerTimeLocal)
	t_in_time, err := time.ParseInLocation("2006/1/2 15:04",
		fmt.Sprintf("%d/%d/%d %s", t.Year(), int(t.Month()), t.Day(),
			gdGEStartTime),
		util.ServerTimeLocal)
	if err != nil {
		logs.Error("GatesEnemyData GetActTime time.ParseInLocation err %v", err)
		return math.MaxInt64, math.MaxInt64
	}
	e = util.GetNextDailyTime(t_in_time.Unix()+
		int64(gdGEPlayTime*util.MinSec), now_t)
	s = e - int64(gdGEPlayTime*util.MinSec)
	return
}

func GetGeReadyBETime(now_t int64) (btime, etime int64) {
	s, e := GetGETime(now_t)
	return s - int64(gdGEWaitTime*util.MinSec), e
}

func DebugSetGEStartTime(s string) {
	logs.Debug("DebugSetGEStartTime %s", s)
	if s == "reset" {
		gdGEStartTime = gdGECommonConfig.GetGEStartTime()
	} else {
		_, err := time.ParseInLocation("2006/1/2 15:04",
			fmt.Sprintf("2006/1/2 %s", s),
			util.ServerTimeLocal)
		if err != nil {
			logs.Error("DebugSetGEStartTime err %s %v", s, err)
			return
		}
		gdGEStartTime = s
	}
}

func DebugSetGEPlayTime(typ string, min uint32) {
	logs.Debug("DebugSetGEPlayTime %d", min)
	if typ == "reset" {
		gdGEPlayTime = gdGECommonConfig.GetGEPlayTime()
	} else {
		gdGEPlayTime = min
	}
}

func GetAllGEEnemyGroupCfg() []*ProtobufGen.GEENEMYGROUP {
	return gdGEEnemyGroupInfo[:]
}

func GetGEEnemyCfg(enemyId string) *ProtobufGen.GEENEMY {
	return gdGEEnemy[enemyId]
}

func GetGEEnemyLootCfg(enemyId string) *ProtobufGen.GELOOT {
	return gdGEEnemyLoot[enemyId]
}

func GetGEEnemyGiftCfg(point int) *gEEnemyGiftData {
	var res *gEEnemyGiftData
	currPoint := 0
	for i := 0; i < len(gdGEEnemyGift); i++ {
		gift := &gdGEEnemyGift[i]
		if gift == nil {
			continue
		}

		p := gift.PointNeed

		if point >= p && p > currPoint {
			currPoint = p
			res = gift
		}
	}

	return res
}

func GetGEPlayTime() uint32 {
	return gdGEPlayTime
}

func GetGEWaitTimeSec() int64 {
	return int64(gdGEWaitTime) * Minute2Second
}
