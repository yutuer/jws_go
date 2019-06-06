package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 新的新的升星逻辑

const (
	EquipStarLevelUpBonus_Null = iota
	EquipStarLevelUpBonus_Little
	EquipStarLevelUpBonus_Big
	EquipStarLevelUpBonusCount
)

var (
	gdEquipStarData        []*ProtobufGen.NEWSTARLVUP
	gdEquipStarBonusPool   []util.RandIntSet
	gdExPerStarLevelUp     uint32
	gdEquipStarLittleBonus uint32
	gdEquipStarBigBonus    uint32
	gdEquipStarCostSCType  string
	gdEquipStarCostSCCost  uint32
	gdEquipStarHCRatio     float32
	gdEquipStarHcCost      []uint32
)

func loadStarUpConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.NEWSTARLVUP_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	item_data := ar.GetItems()
	l := len(item_data) + 1
	gdEquipStarData = make([]*ProtobufGen.NEWSTARLVUP, l, l)
	gdEquipStarBonusPool = make([]util.RandIntSet, l, l)

	for _, a := range item_data {
		lv := int(a.GetStarLV())
		gdEquipStarData[lv] = a
		gdEquipStarBonusPool[lv].Init(EquipStarLevelUpBonusCount)
		gdEquipStarBonusPool[lv].Add(EquipStarLevelUpBonus_Little, uint32(a.GetLittleBonusRate()*10000))
		gdEquipStarBonusPool[lv].Add(EquipStarLevelUpBonus_Big, uint32(a.GetBigBonusRate()*10000))
		nilPower := uint32((1.0 - a.GetLittleBonusRate() - a.GetBigBonusRate()) * 10000)
		gdEquipStarBonusPool[lv].Add(EquipStarLevelUpBonus_Null, nilPower)
		gdEquipStarBonusPool[lv].Make()
	}

	logs.Trace("gdEquipStarData %v", gdEquipStarData)
}

func loadStarUpSettingConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.NEWSTARLVUPSETTINGS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	item_data := ar.GetItems()

	gdExPerStarLevelUp = item_data[0].GetEXPPerOperate()
	gdEquipStarLittleBonus = item_data[0].GetLittleBonus()
	gdEquipStarBigBonus = item_data[0].GetBigBonus()
	gdEquipStarCostSCType = item_data[0].GetCostCoin()
	gdEquipStarCostSCCost = item_data[0].GetCostForEach()
	gdEquipStarHCRatio = item_data[0].GetHCRatio()

	logs.Trace("gdEquipStarData %v", item_data)
}

func loadStarUpHcCostConfig(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.NEWSTARLVUPSTAGECOST_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	item_data := ar.GetItems()

	gdEquipStarHcCost = make([]uint32, len(item_data), len(item_data))
	for _, data := range item_data {
		gdEquipStarHcCost[int(data.GetHCOperateTimes())-1] = data.GetHCCost()
	}
	logs.Trace("gdEquipStarHcCost %v", gdEquipStarHcCost)
}

func GetStarUpHcCount(starUpCount int) uint32 {
	return gdEquipStarHcCost[starUpCount]
}

func GetEquipStarData(star uint32) *ProtobufGen.NEWSTARLVUP {
	s := int(star)
	if s < 0 || s >= len(gdEquipStarData) {
		return nil
	}

	return gdEquipStarData[s]
}

func GetEquipStarLvUpData(star uint32) (uint32, *util.RandIntSet) {
	s := int(star)
	if s < 0 || s >= len(gdEquipStarData) {
		return 0, nil
	}

	return gdEquipStarData[s].GetStarUpExp(), &gdEquipStarBonusPool[s]
}

func GetEquipStarUpSettings() (string, uint32, uint32, uint32, uint32, float32) {
	return gdEquipStarCostSCType,
		gdEquipStarCostSCCost,
		gdExPerStarLevelUp,
		gdEquipStarLittleBonus,
		gdEquipStarBigBonus,
		gdEquipStarHCRatio
}
