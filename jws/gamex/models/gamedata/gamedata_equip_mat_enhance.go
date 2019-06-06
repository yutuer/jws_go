package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdEquipMatEnhance_Slot2Lvl2Cfg map[int]map[uint32]*ProtobufGen.MATERIALENHANCE
)

func loadEquipMaterialEnhance(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.MATERIALENHANCE_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdEquipMatEnhance_Slot2Lvl2Cfg = make(map[int]map[uint32]*ProtobufGen.MATERIALENHANCE, PartEquipCount)
	for _, v := range ar.GetItems() {
		slot := GetEquipSlot(v.GetPart())
		if slot < 0 {
			panic(fmt.Errorf("loadEquipMaterialEnhance index %s part %s not fount", v.GetIndex(), v.GetPart()))
		}
		lvl2Cfg := gdEquipMatEnhance_Slot2Lvl2Cfg[slot]
		if lvl2Cfg == nil {
			lvl2Cfg = make(map[uint32]*ProtobufGen.MATERIALENHANCE, 64)
		}
		lvl2Cfg[v.GetEnhanceLevel()] = v
		gdEquipMatEnhance_Slot2Lvl2Cfg[slot] = lvl2Cfg
		for _, needs := range v.GetMaterials_Table() {
			addNeed(needs.GetMaterialsID(), needs.GetMaterialsCount())
		}
	}

	logs.Trace("universalmaterial %v", gdUniversalMaterialDatas)
}

func GetEquipMatEnhCfg(slot int, lvl uint32) *ProtobufGen.MATERIALENHANCE {
	lvl2Cfg := gdEquipMatEnhance_Slot2Lvl2Cfg[slot]
	if lvl2Cfg != nil {
		return lvl2Cfg[lvl]
	}
	return nil
}
