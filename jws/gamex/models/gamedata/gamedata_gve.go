package gamedata

import (
	"errors"
	"math/rand"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
)

var (
	gdGVEDailyDoubleRewardFightCount int
	gdGVEFightTimeLimit              int64

	gdGVEDoublePrice     int
	gdGVEHardDoublePrice int
	gdGVECost            []uint32

	gdGVERwardNomarlTemple      []string
	gdGVERwardNomarlTempleCount []uint32
	gdGVERwardHardTemple        []string
	gdGVERwardHardTempleCount   []uint32

	gdGVEEnemyGroup     []GVEEnemyGroup
	gdGVEEnemy          []*ProtobufGen.GVEENEMY
	gdGVEEnemyModel     []*ProtobufGen.GVEMODEL
	gdGVEStartDailyTime int64
	gdGVEEndSec         int64
	gdGVEBotTimeMin     int64
	gdGVEBotTimeMax     int64
	gdGVEBotGSMax       float32
	gdGVEBotGSMin       float32
)

type GVEEnemyGroup struct {
	LevelMin       uint32
	LevelMax       uint32
	LevelID        string
	BossRander     util.RandIntSet
	HardBossRander util.RandIntSet
}

func (g *GVEEnemyGroup) FromData(d *ProtobufGen.GVEENEMYGROUP) {
	g.LevelID = d.GetEGLevelID()
	g.LevelMin = d.GetLevelMin()
	g.LevelMax = d.GetLevelMax()
	g.BossRander.Init(3)
	g.BossRander.Add(int(d.GetBoss1()), 100)
	g.BossRander.Add(int(d.GetBoss2()), 100)
	g.BossRander.Add(int(d.GetBoss3()), 100)
	if !g.BossRander.Make() {
		panic(errors.New("BossRander Make Err"))
	}
	g.HardBossRander.Init(3)
	g.HardBossRander.Add(int(d.GetHardBoss1()), 100)
	g.HardBossRander.Add(int(d.GetHardBoss2()), 100)
	g.HardBossRander.Add(int(d.GetHardBoss3()), 100)
	if !g.HardBossRander.Make() {
		panic(errors.New("HardBossRander Make Err"))
	}
}

type GVEGameData struct {
	Ok        bool
	LevelID   string
	Boss      *ProtobufGen.GVEENEMY
	BossModel *ProtobufGen.GVEMODEL
}

func GetGVEGameCfg() (int64, int) {
	return gdGVEFightTimeLimit,
		gdGVEDailyDoubleRewardFightCount
}

func GetGVEGameCostCfg() (int, int, []uint32) {
	return gdGVEDoublePrice, gdGVEHardDoublePrice, gdGVECost[:]
}

func GetGVEGameBotTime() (int64, int64) {
	return gdGVEBotTimeMin, gdGVEBotTimeMax
}

func GetGVEGameBotGSFloat() (float32, float32) {
	return gdGVEBotGSMin, gdGVEBotGSMax
}

func GetGVEGameRewardCfg(isHard bool) ([]string, []uint32) {
	if isHard {
		return gdGVERwardHardTemple[:], gdGVERwardHardTempleCount[:]
	} else {
		return gdGVERwardNomarlTemple[:], gdGVERwardNomarlTempleCount[:]
	}
}

//GetGVEGameData 获取本次游戏的boss属性,有随机部分
func GetGVEGameData(corpLv uint32, isHard bool, rd *rand.Rand) GVEGameData {
	var (
		idx     int
		bossIdx int
	)
	for i := 0; i < len(gdGVEEnemyGroup); i++ {
		if (gdGVEEnemyGroup[i].LevelMin <= corpLv) &&
			(corpLv <= gdGVEEnemyGroup[i].LevelMax) {
			idx = i
		}
	}
	res := GVEGameData{}
	res.LevelID = gdGVEEnemyGroup[idx].LevelID

	if isHard {
		bossIdx = gdGVEEnemyGroup[idx].HardBossRander.Rand(rd)
	} else {
		bossIdx = gdGVEEnemyGroup[idx].BossRander.Rand(rd)
	}

	if bossIdx > 0 && bossIdx < len(gdGVEEnemy) && int(corpLv) > 0 && int(corpLv) < len(gdGVEEnemyModel) {
		res.Boss = gdGVEEnemy[bossIdx]
		res.BossModel = gdGVEEnemyModel[corpLv]
		res.Ok = true
	}

	return res
}

func loadGVEConfigData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GVECONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdGVEFightTimeLimit = int64(data[0].GetFightTimeLimit())
	gdGVEDailyDoubleRewardFightCount = int(data[0].GetDailyDoubleTime())
	gdGVEDoublePrice = int(data[0].GetDoublePrice())
	gdGVEHardDoublePrice = int(data[0].GetHardDoublePrice())
	gdGVECost = make([]uint32, len(data[0].GetCost()), len(data[0].GetCost()))
	for idx, c := range data[0].GetCost() {
		gdGVECost[idx] = c.GetGCost()
	}
	gdGVEStartDailyTime = util.DailyTimeFromString(data[0].GetGEStartTime())
	gdGVEEndSec = int64(data[0].GetGEPlayTime()) * util.MinSec
	gdGVEBotTimeMax, gdGVEBotTimeMin = int64(data[0].GetBotTimeMax()), int64(data[0].GetBotTimeMin())
	if gdGVEBotTimeMax <= gdGVEBotTimeMin {
		panic("GVE BotTime max time < min time")
	}
	gdGVEBotGSMin, gdGVEBotGSMax = data[0].GetGSMin(), data[0].GetGSMax()

}

func loadGVEEnemyGroupData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GVEENEMYGROUP_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdGVEEnemyGroup = make([]GVEEnemyGroup, len(data), len(data))
	for idx, v := range data {
		gdGVEEnemyGroup[idx].FromData(v)
	}
}

func loadGVEEnemyData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GVEENEMY_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdGVEEnemy = make([]*ProtobufGen.GVEENEMY, len(data)+1, len(data)+1)
	for _, v := range data {
		gdGVEEnemy[int(v.GetID())] = v
	}
}

func loadGVEEnemyModelData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GVEMODEL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdGVEEnemyModel = make([]*ProtobufGen.GVEMODEL, len(data)+1, len(data)+1)
	for _, v := range data {
		gdGVEEnemyModel[int(v.GetPlayLevel())] = v
	}
}

func loadGVELootData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GVELOOT_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	for _, v := range data {
		if v.GetEGLevelID() == "team_boss_1" {
			if len(gdGVERwardNomarlTemple) > 0 {
				panic(errors.New("GVELootData normal err"))
			}
			gdGVERwardNomarlTemple = make([]string, 0, len(v.GetLoots()))
			gdGVERwardNomarlTempleCount = make([]uint32, 0, len(v.GetLoots()))
			for _, loot := range v.GetLoots() {
				gdGVERwardNomarlTemple = append(gdGVERwardNomarlTemple, loot.GetLootTemplate())
				gdGVERwardNomarlTempleCount = append(gdGVERwardNomarlTempleCount, loot.GetLootTimes())
			}
		} else if v.GetEGLevelID() == "team_boss_2" {
			if len(gdGVERwardHardTemple) > 0 {
				panic(errors.New("GVELootData hard err"))
			}
			gdGVERwardHardTemple = make([]string, 0, len(v.GetLoots()))
			gdGVERwardHardTempleCount = make([]uint32, 0, len(v.GetLoots()))
			for _, loot := range v.GetLoots() {
				gdGVERwardHardTemple = append(gdGVERwardHardTemple, loot.GetLootTemplate())
				gdGVERwardHardTempleCount = append(gdGVERwardHardTempleCount, loot.GetLootTimes())
			}
		} else {
			panic(errors.New("unknown GVELootData GetEGLevelID"))
		}
	}
}

func IsCanGVE(nowT int64) bool {
	start := util.DailyTime2UnixTime(nowT, gdGVEStartDailyTime)
	return (start <= nowT) && (nowT <= (start + gdGVEEndSec))
}

func GetGVETime(nowT int64) (int64, int64) {
	start := util.DailyTime2UnixTime(nowT, gdGVEStartDailyTime)
	end := start + gdGVEEndSec
	if nowT >= end {
		start += Day2Second
		end += Day2Second
	}
	return start, end
}
