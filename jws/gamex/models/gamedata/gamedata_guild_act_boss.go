package gamedata

import (
	"github.com/golang/protobuf/proto"
	"math/rand"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdGuildBossData     map[uint32]*ProtobufGen.GUILDBOSS
	gdGuildGBConfig     *ProtobufGen.GBCONFIG
	gdGuildGBEnemyGroup []*ProtobufGen.GBENEMYGROUP
	gdGuildBossDroplist map[string]*ProtobufGen.GBDROPLIST
)

type guildGBDropsData struct {
	Ids    []string
	Counts []uint32
}

var (
	gdGuildGBEnemyGroupRander map[string]*util.RandSet
	gdGuildGBRestartTime      util.TimeToBalance
	gdGuildGBFinishTime       util.TimeToBalance
	gdGuildGBDrops            map[string]guildGBDropsData
)

func GetGuildBossDataByLv(lv uint32) *ProtobufGen.GUILDBOSS {
	return gdGuildBossData[lv]
}

func GetGuildBossCfg() *ProtobufGen.GBCONFIG {
	return gdGuildGBConfig
}

func RandGuildBossEnemy(groupID string, rd *rand.Rand) (string, bool) {
	res, ok := gdGuildGBEnemyGroupRander[groupID]
	if ok {
		return res.Rand(rd), true
	} else {
		return "", false
	}
}

func GetGuildBossRestartTime() util.TimeToBalance {
	return gdGuildGBRestartTime
}

func GetGuildBossFinishTime() util.TimeToBalance {
	return gdGuildGBFinishTime
}

func GetGuildBossDroplist(dropID string) guildGBDropsData {
	return gdGuildGBDrops[dropID]
}

func GetGuildBossDataCfg(lv, idx int) *ProtobufGen.GUILDBOSS_BossData {
	d, ok := gdGuildBossData[uint32(lv)]
	if !ok || d == nil {
		return nil
	}
	t := d.GetBossData_Table()
	if idx >= len(t) || idx < 0 {
		return nil
	}

	return t[idx]
}

func loadGuildBossData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.GUILDBOSS_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	data := ar.GetItems()
	gdGuildBossData = make(map[uint32]*ProtobufGen.GUILDBOSS, len(data))

	for _, c := range data {
		gdGuildBossData[c.GetDifficultyLevel()] = c
	}
}

func loadGuildGBConfig(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.GBCONFIG_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	data := ar.GetItems()

	gdGuildGBConfig = data[0]
	logs.Trace("gdGuildGBConfig %v", gdGuildGBConfig)
}

func loadGuildGBEnemyGroup(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.GBENEMYGROUP_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	data := ar.GetItems()
	gdGuildGBEnemyGroup = make([]*ProtobufGen.GBENEMYGROUP, 0, len(data))

	for _, c := range data {
		gdGuildGBEnemyGroup = append(gdGuildGBEnemyGroup, c)
	}
}

func loadGuildBossDroplist(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.GBDROPLIST_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	data := ar.GetItems()
	gdGuildBossDroplist = make(map[string]*ProtobufGen.GBDROPLIST, len(data))

	for _, c := range data {
		gdGuildBossDroplist[c.GetDropID()] = c
	}
}

func mkGuildBossDatas(loadFunc func(dfilepath string, loadfunc func(string))) {
	loadFunc("guildboss.data", loadGuildBossData)
	loadFunc("gbconfig.data", loadGuildGBConfig)
	loadFunc("gbenemygroup.data", loadGuildGBEnemyGroup)
	loadFunc("gbdroplist.data", loadGuildBossDroplist)

	gdGuildGBEnemyGroupRander = make(map[string]*util.RandSet)
	for _, data := range gdGuildGBEnemyGroup {
		groupID := data.GetBossGroupID()
		grander, ok := gdGuildGBEnemyGroupRander[groupID]
		if ok {
			grander.Add(data.GetBossID(), data.GetWeight())
		} else {
			n := new(util.RandSet)
			n.Init(64)
			n.Add(data.GetBossID(), data.GetWeight())
			gdGuildGBEnemyGroupRander[groupID] = n
		}
	}

	for _, r := range gdGuildGBEnemyGroupRander {
		if !r.Make() {
			logs.Error("guild boss rander make err %v", r)
		}
		logs.Trace("gdGuildGBEnemyGroupRander %v", r)

	}

	gdGuildGBRestartTime = util.TimeToBalance{
		DailyTime: util.DailyTimeFromString(gdGuildGBConfig.GetRestartTime()),
	}

	gdGuildGBFinishTime = util.TimeToBalance{
		DailyTime: util.DailyTimeFromString(gdGuildGBConfig.GetFinishTime()),
	}

	gdGuildGBDrops = make(map[string]guildGBDropsData, len(gdGuildBossDroplist))
	for did, d := range gdGuildBossDroplist {
		data := guildGBDropsData{
			Ids:    make([]string, 0, 8),
			Counts: make([]uint32, 0, 8),
		}

		for _, dd := range d.GetLoot_Table() {
			data.Ids = append(data.Ids, dd.GetGuildBagID())
			data.Counts = append(data.Counts, dd.GetGuildBagNum())
		}

		gdGuildGBDrops[did] = data
	}
}
