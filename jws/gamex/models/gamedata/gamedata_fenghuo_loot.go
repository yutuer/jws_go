package gamedata

import (
	"errors"
	"math/rand"
	"time"

	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const FenghuoStageMaxNum = 8
const FegnhuoRoomMaxPlayer = 2

type FenghuoLootData struct {
	LootTemplate string
	Count        uint32
}

var (
	gdFenghuoDropList map[string][]FenghuoLootData
)

func loadFengHuoDropData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar1 := new(ProtobufGen.FGDROPLIST_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar1))
	dropData := ar1.GetItems()
	gdFenghuoDropList = make(map[string][]FenghuoLootData, len(dropData))

	for _, c := range dropData {
		loots := make([]FenghuoLootData, 0, 8)
		for _, ci := range c.GetLoot_Table() {
			loots = append(loots, FenghuoLootData{
				LootTemplate: ci.GetLootTemplateID(),
				Count:        ci.GetLootTime(),
			})
		}
		gdFenghuoDropList[c.GetDropID()] = loots
	}
}

type FenghuoLevelData struct {
	// * 大王关
	LevelInfoID string
	// * 最高出现次数
	MaxNum int
	// * 普通掉落（普通扫荡也用此列）
	NormalDrop []FenghuoLootData
	// * 高级扫荡掉落
	SeniorSweepDrop []FenghuoLootData
	// * 关卡类型（1=大boss，2=双子boss，3=小怪
	LevelType uint32
}

type FenghuoMain struct {
	//SubLevel金币消耗
	SubLevelCostSC uint32
	//团队翻倍效率产出消耗
	TeamSubLevelCostHC uint32

	LevelBigBoss    FenghuoLevelData
	LevelDoubleBoss FenghuoLevelData
	LevelOthers     []FenghuoLevelData

	//最终轮奖励
	FinalLoot []FenghuoLootData
}

type fenghuoMainMap map[uint32]FenghuoMain

var (
	gdFenghuoMainMap fenghuoMainMap
	fenghuoRand      *rand.Rand
)

func FenghuoHasEnoughCurrency(sc, hc int64, rewardPower, BattleHard uint32, master bool) bool {
	if rewardPower <= 0 {
		rewardPower = 1
	}
	csc, chc := GetFenghuoSCHC(BattleHard)
	nsc := csc * rewardPower
	nhc := chc * (rewardPower - 1)

	if master {
		if int64(nsc) <= sc && int64(nhc) <= hc {
			return true
		}
	} else {
		if int64(nsc) <= sc {
			return true
		}
	}
	return false
}

func GetFenghuoSCHC(BattleHard uint32) (sc, hc uint32) {
	fm, ok := gdFenghuoMainMap[BattleHard]
	if !ok {
		logs.Error("Fenghuo GetFenghuoSCHC error! %d do not exist!", BattleHard)
		return 0, 0
	}

	return fm.SubLevelCostSC, fm.TeamSubLevelCostHC
}

func GetFenghuoSubLevels(BattleHard uint32) []FenghuoLevelData {
	fm, ok := gdFenghuoMainMap[BattleHard]
	if !ok {
		logs.Error("Fenghuo GetFenghuoSubLevels error! %d do not exist!", BattleHard)
		return nil
	}

	shidx := util.Shuffle1ToNSelf(FenghuoStageMaxNum, fenghuoRand)

	sublvls := make([]FenghuoLevelData, FenghuoStageMaxNum, FenghuoStageMaxNum)
	nBigBoss := fenghuoRand.Intn(fm.LevelBigBoss.MaxNum) + 1
	nDoubleBoss := fenghuoRand.Intn(fm.LevelDoubleBoss.MaxNum) + 1
	idx := 0
	for i := 0; i < nBigBoss; i++ {
		sublvls[shidx[idx]] = fm.LevelBigBoss
		idx++
	}
	for i := 0; i < nDoubleBoss; i++ {
		sublvls[shidx[idx]] = fm.LevelDoubleBoss
		idx++
	}

	for i := nBigBoss + nDoubleBoss; i < FenghuoStageMaxNum; i++ {
		l := len(fm.LevelOthers)
		got := fenghuoRand.Intn(l)
		sublvls[shidx[idx]] = fm.LevelOthers[got]
		idx++
	}
	return sublvls
}

func GetFinalReward(rnd *rand.Rand, BattleHard uint32) (PriceDatas, error) {
	fm, ok := gdFenghuoMainMap[BattleHard]
	if !ok {
		logs.Error("Fenghuo GetFenghuoSubLevels error! %d do not exist!", BattleHard)
		return PriceDatas{}, fmt.Errorf("fenghuo room GetFinalReward, BattleHard %d not found", BattleHard)
	}
	return MakeFenghuoGives(rnd, fm.FinalLoot)
}

func MakeFenghuoGives(rnd *rand.Rand, drops []FenghuoLootData) (PriceDatas, error) {
	giveDatas := NewPriceDatas(8)

	for _, lootTemplate := range drops {
		for c := 0; c < int(lootTemplate.Count); c++ {
			gives, err := LootTemplateRand(rnd, lootTemplate.LootTemplate)
			if err != nil {
				return PriceDatas{}, err
			}
			giveDatas.AddOther(&gives)
		}
	}
	return giveDatas, nil
}

func makeFenghuoLootData(drop string) []FenghuoLootData {
	Reward, ok := gdFenghuoDropList[drop]
	if !ok {
		panic(errors.New("makeFenghuoLootData no found:" + drop))
	}
	return Reward
}

func loadFengHuoMainData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar0 := new(ProtobufGen.FENGHUOMAIN_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar0))

	mainData := ar0.GetItems()
	gdFenghuoMainMap = make(fenghuoMainMap, len(mainData))

	for _, v := range mainData {

		bb := v.GetBigBossData_Table()
		bigboss := FenghuoLevelData{
			LevelInfoID: bb.GetLevelInfoID(),
			LevelType:   bb.GetLevelType(),
			MaxNum:      int(bb.GetMaxNum()),
		}
		bigboss.NormalDrop = makeFenghuoLootData(bb.GetNormalDrop())
		if bb.GetSeniorSweepDrop() == "" {
			bigboss.SeniorSweepDrop = []FenghuoLootData{}
		} else {
			bigboss.SeniorSweepDrop = makeFenghuoLootData(bb.GetSeniorSweepDrop())
		}

		db := v.GetDoubleBossData_Table()
		doubleboss := FenghuoLevelData{
			LevelInfoID: db.GetLevelInfoID(),
			LevelType:   db.GetLevelType(),
			MaxNum:      int(db.GetMaxNum()),
		}
		doubleboss.NormalDrop = makeFenghuoLootData(db.GetNormalDrop())
		if db.GetSeniorSweepDrop() == "" {
			doubleboss.SeniorSweepDrop = []FenghuoLootData{}
		} else {
			doubleboss.SeniorSweepDrop = makeFenghuoLootData(db.GetSeniorSweepDrop())
		}

		fl := make([]FenghuoLevelData, 0, v.GetSubLevelNum())
		for _, lvl := range v.GetLevelINfoData_Table() {
			fld := FenghuoLevelData{
				LevelInfoID: lvl.GetLevelInfoID(),
				LevelType:   lvl.GetLevelType(),
				MaxNum:      0,
			}
			fld.NormalDrop = makeFenghuoLootData(lvl.GetNormalDrop())
			if lvl.GetSeniorSweepDrop() == "" {
				fld.SeniorSweepDrop = []FenghuoLootData{}
			} else {
				fld.SeniorSweepDrop = makeFenghuoLootData(lvl.GetSeniorSweepDrop())
			}
			fl = append(fl, fld)
		}
		var mainFinalLoot []FenghuoLootData
		if v.GetFinalLoot() != "" {
			mainFinalLoot = makeFenghuoLootData(v.GetFinalLoot())
		}
		gdFenghuoMainMap[v.GetBattleHard()] = FenghuoMain{
			LevelBigBoss:    bigboss,
			LevelDoubleBoss: doubleboss,
			LevelOthers:     fl[:],
			FinalLoot:       mainFinalLoot,

			SubLevelCostSC:     v.GetSubLevelCostSC(),
			TeamSubLevelCostHC: v.GetTeamSubLevelCostHC(),
		}
	}
}

type FenghuoConfig struct {
	//组队最高可以花钻至几倍（每日前两次免费加一倍不计算在此数内）
	NumLimitTeamCostHC uint32
	//小关卡获得免费复活次数的概率（100%=100）0.0-1.0
	SubLevelFreeRevivalChance float64
	//最多获得免费复活次数上限
	FreeRevivalNumLimit uint32
}

var (
	FenghuoConfigData FenghuoConfig
)

func loadFengHuoConfigData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar0 := new(ProtobufGen.FENGHUOCONFIG_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar0))
	cfg := ar0.GetItems()[0]
	FenghuoConfigData.NumLimitTeamCostHC = cfg.GetNumLimitTeamCostHC()
	FenghuoConfigData.FreeRevivalNumLimit = cfg.GetFreeRevivalNumLimit()
	FenghuoConfigData.SubLevelFreeRevivalChance = float64(cfg.GetSubLevelFreeRevivalChance()) / 100.0
}

func mkFenghuoDatas(loadFunc func(dfilepath string, loadfunc func(string))) {
	if fenghuoRand == nil {
		fenghuoRand = rand.New(
			rand.NewSource(time.Now().Unix()))
	}
	loadFunc("fgdroplist.data", loadFengHuoDropData)
	loadFunc("fenghuomain.data", loadFengHuoMainData)
	loadFunc("fenghuoconfig.data", loadFengHuoConfigData)
}
