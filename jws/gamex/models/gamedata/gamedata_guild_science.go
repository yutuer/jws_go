package gamedata

import (
	"math"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GST_Typ uint32

const (
	GST_NULL        GST_Typ = 0
	GST_MemCap      GST_Typ = 1
	GST_BossFight   GST_Typ = 2
	GST_DailyTask   GST_Typ = 3
	GST_GoldBonus   GST_Typ = 5
	GST_GateEnemy   GST_Typ = 6
	GST_WantGeneral GST_Typ = 7
)

type GuildScienceConfig struct {
	GuildLvlReq uint32
	NeedSP      uint32

	MemCap           *ProtobufGen.GSTMEMBERCAP
	BossFightBonus   *ProtobufGen.GSTGBOSSFIGHTBONUS
	DailyTaskExp     *ProtobufGen.GSTDAILYTASKEXP
	GoldBonus        *ProtobufGen.GSTGOLDBONUS
	GateEnemyBonus   *ProtobufGen.GSTGATEENEMYBONUS
	WantGeneralBonus *ProtobufGen.GSTWANNAHERO
}

var (
	gdGuildScienceId2Lvl2Info map[uint32]map[uint32]*GuildScienceConfig
	gdGSTConfig               *ProtobufGen.GSTCONFIG
)

func init() {
	gdGuildScienceId2Lvl2Info = make(map[uint32]map[uint32]*GuildScienceConfig, 10)
}

func (science *GuildScienceConfig) GetLvlNeedSP(typ GST_Typ) uint32 {
	switch typ {
	case GST_MemCap:
		return science.MemCap.GetGSP()
	case GST_BossFight:
		return science.BossFightBonus.GetGSP()
	case GST_DailyTask:
		return science.DailyTaskExp.GetGSP()
	case GST_GoldBonus:
		return science.GoldBonus.GetGSP()
	case GST_GateEnemy:
		return science.GateEnemyBonus.GetGSP()
	case GST_WantGeneral:
		return science.WantGeneralBonus.GetGSP()
	}
	logs.Error("GuildScienceConfig GetLvlNeedSP no type %d", typ)
	return math.MaxUint32
}

func (science *GuildScienceConfig) GetLvlNeedGLv(typ GST_Typ) uint32 {
	switch typ {
	case GST_MemCap:
		return science.MemCap.GetGuildLevelReq()
	case GST_BossFight:
		return science.BossFightBonus.GetGuildLevelReq()
	case GST_DailyTask:
		return science.DailyTaskExp.GetGuildLevelReq()
	case GST_GoldBonus:
		return science.GoldBonus.GetGuildLevelReq()
	case GST_GateEnemy:
		return science.GateEnemyBonus.GetGuildLevelReq()
	case GST_WantGeneral:
		return science.WantGeneralBonus.GetGuildLevelReq()
	}
	logs.Error("GuildScienceConfig GetLvlNeedGLv no type %d", typ)
	return math.MaxUint32
}

func GetGuildScienceConfig(typ GST_Typ, lvl uint32) *GuildScienceConfig {
	r := gdGuildScienceId2Lvl2Info[uint32(typ)]
	if r != nil {
		return r[lvl]
	}
	return nil
}

func GetGuildLevelMemLimit(lvl uint32) uint32 {
	r := gdGuildScienceId2Lvl2Info[uint32(GST_MemCap)]
	return r[lvl].MemCap.GetGuildMemberCap()
}

func GetGuildLevelNewMemLimit(lvl uint32) uint32 {
	r := gdGuildScienceId2Lvl2Info[uint32(GST_MemCap)]
	return r[lvl].MemCap.GetGuildMemberCapNew()
}

func GetGuildScienceBonus(typ GST_Typ, lvl uint32) []float32 {
	cfg := GetGuildScienceConfig(typ, lvl)
	if cfg == nil {
		logs.Error("gamedata GetGuildScienceBonus not found, typ %d lvl %d", typ, lvl)
		return []float32{0.0, 0.0}
	}
	switch typ {
	case GST_BossFight:
		return []float32{cfg.BossFightBonus.GetGoldRewardBonus()}
	case GST_DailyTask:
		return []float32{cfg.DailyTaskExp.GetDailyExpBonusRate()}
	case GST_GoldBonus:
		return []float32{cfg.GoldBonus.GetGoldPurchaseBonusRate()}
	case GST_GateEnemy:
		return []float32{cfg.GateEnemyBonus.GetGuildCoinBonus()}
	case GST_WantGeneral:
		return []float32{
			float32(cfg.WantGeneralBonus.GetFreeParticipateBonus()),
			float32(cfg.WantGeneralBonus.GetFreeResetBonus())}
	}
	return []float32{0.0, 0.0}
}

func GetGSTConfig() *ProtobufGen.GSTCONFIG {
	return gdGSTConfig
}

func loadGuildGSTMemCap(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GSTMEMBERCAP_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	var gstId uint32
	m := make(map[uint32]*GuildScienceConfig, 10)
	for _, v := range data {
		gstId = v.GetGSTid()
		m[v.GetLevel()] = &GuildScienceConfig{
			GuildLvlReq: v.GetGuildLevelReq(),
			NeedSP:      v.GetGSP(),
			MemCap:      v,
		}
	}
	gdGuildScienceId2Lvl2Info[gstId] = m
}

func loadGuildGSTBossFightBonus(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GSTGBOSSFIGHTBONUS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	var gstId uint32
	m := make(map[uint32]*GuildScienceConfig, 10)
	for _, v := range data {
		gstId = v.GetGSTid()
		m[v.GetLevel()] = &GuildScienceConfig{
			GuildLvlReq:    v.GetGuildLevelReq(),
			NeedSP:         v.GetGSP(),
			BossFightBonus: v,
		}
	}
	gdGuildScienceId2Lvl2Info[gstId] = m
}

func loadGuildGSTDailyTaskExp(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GSTDAILYTASKEXP_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	var gstId uint32
	m := make(map[uint32]*GuildScienceConfig, 10)
	for _, v := range data {
		gstId = v.GetGSTid()
		m[v.GetLevel()] = &GuildScienceConfig{
			GuildLvlReq:  v.GetGuildLevelReq(),
			NeedSP:       v.GetGSP(),
			DailyTaskExp: v,
		}
	}
	gdGuildScienceId2Lvl2Info[gstId] = m
}

func loadGuildGSTGoldBonus(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GSTGOLDBONUS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	var gstId uint32
	m := make(map[uint32]*GuildScienceConfig, 10)
	for _, v := range data {
		gstId = v.GetGSTid()
		m[v.GetLevel()] = &GuildScienceConfig{
			GuildLvlReq: v.GetGuildLevelReq(),
			NeedSP:      v.GetGSP(),
			GoldBonus:   v,
		}
	}
	gdGuildScienceId2Lvl2Info[gstId] = m
}

func loadGuildGSTGateEnemyBonus(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GSTGATEENEMYBONUS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	var gstId uint32
	m := make(map[uint32]*GuildScienceConfig, 10)
	for _, v := range data {
		gstId = v.GetGSTid()
		m[v.GetLevel()] = &GuildScienceConfig{
			GuildLvlReq:    v.GetGuildLevelReq(),
			NeedSP:         v.GetGSP(),
			GateEnemyBonus: v,
		}
	}
	gdGuildScienceId2Lvl2Info[gstId] = m
}

func loadGuildGSTWannaHero(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GSTWANNAHERO_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	var gstId uint32
	m := make(map[uint32]*GuildScienceConfig, 10)
	for _, v := range data {
		gstId = v.GetGSTid()
		m[v.GetLevel()] = &GuildScienceConfig{
			GuildLvlReq:      v.GetGuildLevelReq(),
			NeedSP:           v.GetGSP(),
			WantGeneralBonus: v,
		}
	}
	gdGuildScienceId2Lvl2Info[gstId] = m
}

func loadGuildGSTConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GSTCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdGSTConfig = ar.Items[0]
}
