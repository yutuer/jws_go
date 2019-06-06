package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	Unlock_Typ_Auto   = 0
	Unlock_Typ_Manual = 1
)

var (
	heroSwingStarLvInfo []*ProtobufGen.HEROWINGSTAR
	heroSwingLvInfo     []*ProtobufGen.HEROWINGLEVEL
	swingTypeInfo       map[uint32]*ProtobufGen.HEROWINGTABLE
	heroOwnSwing        []*ProtobufGen.HEROWINGLIST
)

func loadHeroSwingStarLevelData(filepath string) {

	buffer, err := loadBin(filepath)
	panicIfErr(err)
	// 神翼星级数据
	protoData := &ProtobufGen.HEROWINGSTAR_ARRAY{}
	err = proto.Unmarshal(buffer, protoData)
	panicIfErr(err)

	data := protoData.GetItems()
	dataLen := len(data)
	heroSwingStarLvInfo = make([]*ProtobufGen.HEROWINGSTAR, dataLen+1, dataLen+1)
	for _, item := range data {
		heroSwingStarLvInfo[int(item.GetHWStar())] = item
	}
}

func loadHeroSwingLevelData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)
	// 神翼等级数据
	protoData := &ProtobufGen.HEROWINGLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, protoData)
	panicIfErr(err)

	data := protoData.GetItems()
	dataLen := len(data)
	heroSwingLvInfo = make([]*ProtobufGen.HEROWINGLEVEL, dataLen+1, dataLen+1)
	for _, item := range data {
		heroSwingLvInfo[int(item.GetHWLevel())] = item
	}
}

func loadHeroSwingTypeData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)
	//神翼外观种类数据
	protoData := &ProtobufGen.HEROWINGTABLE_ARRAY{}
	err = proto.Unmarshal(buffer, protoData)
	panicIfErr(err)

	data := protoData.GetItems()
	swingTypeInfo = make(map[uint32]*ProtobufGen.HEROWINGTABLE, len(data))
	for _, item := range data {
		swingTypeInfo[item.GetHWID()] = item
	}
}

func loadHeroSwingOwnData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)
	// 英雄所能拥有神翼外观种类数据
	protoData := &ProtobufGen.HEROWINGLIST_ARRAY{}
	err = proto.Unmarshal(buffer, protoData)
	panicIfErr(err)

	data := protoData.GetItems()
	heroOwnSwing = make([]*ProtobufGen.HEROWINGLIST, 0, len(data))
	for _, item := range data {
		heroOwnSwing = append(heroOwnSwing, item)
	}
}

func GetHeroSwingCanAct(starLv uint32) []int {
	ret := make([]int, 0, 4)
	for _, item := range swingTypeInfo {
		if starLv >= item.GetHWUnlockStar() && item.GetHWUnlockType() == Unlock_Typ_Auto {
			ret = append(ret, int(item.GetHWID()))
		}
	}
	return ret
}

func GetHeroSwingInfo(id int) *ProtobufGen.HEROWINGTABLE {
	return swingTypeInfo[uint32(id)]
}

func GetHeroSwingLvUpInfo(lv int) *ProtobufGen.HEROWINGLEVEL {
	if lv < 0 || lv >= len(heroSwingLvInfo) {
		return nil
	}
	return heroSwingLvInfo[lv]
}

func GetHeroSwingStarLvUpInfo(starLv int) *ProtobufGen.HEROWINGSTAR {
	if starLv < 0 || starLv > len(heroSwingStarLvInfo) {
		return nil
	}
	return heroSwingStarLvInfo[starLv]
}

func GetHeroSwingResetCost(lv, starLv int) (int, bool) {
	if lv < 0 || lv > len(heroSwingLvInfo) {
		return 0, false
	}
	if starLv < 0 || starLv > len(heroSwingStarLvInfo) {
		return 0, false
	}
	return int(heroSwingLvInfo[lv].GetRebornCost() + heroSwingStarLvInfo[starLv].GetRebornCost()), true
}

func GetHeroSwingResetReward(lv, starLv int) (*CostData, bool) {
	if lv < 0 || lv > len(heroSwingLvInfo) {
		return nil, false
	}
	if starLv < 0 || starLv > len(heroSwingStarLvInfo) {
		return nil, false
	}
	costData := &CostData{}
	for i := 1; i <= lv; i++ {
		info := heroSwingLvInfo[i]
		costData.AddItem(info.GetHWLevelupMaterial(), info.GetHWLevelupMaterialCount())
	}
	for i := 1; i <= starLv; i++ {
		info := heroSwingStarLvInfo[i]
		for _, item := range info.GetHWStarup_Template() {
			costData.AddItem(item.GetHWStarupMaterial(), item.GetHWStarupMaterialCount())
		}
	}
	return costData, true
}
