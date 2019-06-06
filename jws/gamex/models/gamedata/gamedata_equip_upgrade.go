package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdEquipUpgradeConfig  []*ProtobufGen.EQUIPUPGRADE
	gdEquipEvolutionCofig [AVATAR_SLOT_MAX][]*ProtobufGen.EVOLUTION

	gdEquipUpgradeAllNeedData []CostData // 注意这个是累计值
	gdEquipEvolutionNeedData  [AVATAR_SLOT_MAX][]CostData
)

func GetEquipUpgrade(lv int) *ProtobufGen.EQUIPUPGRADE {
	if lv < 0 || lv >= len(gdEquipUpgradeConfig) {
		logs.Error("No EquipUpgrade data by %d", lv)
		return nil
	}
	return gdEquipUpgradeConfig[lv]
}

func GetEquipEvolution(slot, lv int) *ProtobufGen.EVOLUTION {

	if slot < 0 || slot >= len(gdEquipEvolutionCofig) {
		logs.Error("No EquipEvolution data by slot %d", slot)
		return nil
	}

	if lv < 0 || lv >= len(gdEquipEvolutionCofig[slot]) {
		//logs.Error("No EquipEvolution data by  lv %d %d", slot, lv)
		return nil
	}

	return gdEquipEvolutionCofig[slot][lv]
}

func loadEquipUpgradeCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.EQUIPUPGRADE_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdEquipUpgradeConfig = make([]*ProtobufGen.EQUIPUPGRADE, 0, len(dataList.GetItems()))
	gdEquipUpgradeAllNeedData = make([]CostData, 0, len(dataList.GetItems()))

	for _, a := range dataList.GetItems() {
		//uplv := a.GetUpgradeLevel()
		//logs.Trace("equip upgrade %s  ->  %v", uplv, a.GetFineIronCost())
		gdEquipUpgradeConfig = append(gdEquipUpgradeConfig, a)

		need := CostData{}

		// 这里需要注意，升级的消耗的配的累计值
		need.AddItem(VI_Sc0, a.GetSC())
		need.AddItem(VI_Sc1, a.GetFineIronCost())

		gdEquipUpgradeAllNeedData = append(gdEquipUpgradeAllNeedData, need)
		//logs.Trace("AllNeedData %d : %v", uplv, need)
	}

}

func loadEquipEvolutionCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.EVOLUTION_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)
	for _, a := range dataList.GetItems() {
		//logs.Trace("equip evolution %v", a.GetEvolutionLevel())

		data := CostData{}
		for _, item := range a.GetMaterials_Table() {
			data.AddItem(item.GetMaterialsID(), item.GetMaterialsCount())
		}
		data.AddItem(VI_Sc0, a.GetSC())
		data.AddItem(VI_Sc1, a.GetFineIronCost())

		part_idx := GetEquipSlot(a.GetPart())
		if part_idx < 0 {
			logs.Error("GetEquipSlot Err by %s", a.GetPart())
			continue
		}
		if gdEquipEvolutionNeedData[part_idx] == nil {
			gdEquipEvolutionNeedData[part_idx] = make([]CostData, 10, 64)
		}

		if gdEquipEvolutionCofig[part_idx] == nil {
			gdEquipEvolutionCofig[part_idx] = make([]*ProtobufGen.EVOLUTION, 10, 64)
		}

		eidx := int(a.GetEvolutionLevel())

		for len(gdEquipEvolutionNeedData[part_idx]) <= eidx {
			gdEquipEvolutionNeedData[part_idx] = append(gdEquipEvolutionNeedData[part_idx], CostData{})
		}

		for len(gdEquipEvolutionCofig[part_idx]) <= eidx {
			gdEquipEvolutionCofig[part_idx] = append(gdEquipEvolutionCofig[part_idx], nil)
		}

		gdEquipEvolutionNeedData[part_idx][eidx] = data
		gdEquipEvolutionCofig[part_idx][eidx] = a

		//gdEquipEvolutionNeedData = append(gdEquipEvolutionNeedData, data)
		//logs.Trace("Evo Data %d --> %v  %v", a.GetEvolutionLevel(), data, eidx)
	}

	//logs.Trace("Evo Data %v", gdEquipEvolutionNeedData)

}

func GetEquipUpgradeNeed(from, to uint32) *CostData {
	// 再次注意 这个是累计值
	if to <= from {
		return &gdEquipUpgradeAllNeedData[0]
	}

	from_cost := &gdEquipUpgradeAllNeedData[from]
	to_cost := &gdEquipUpgradeAllNeedData[to]

	return to_cost.SubSCGroup(from_cost)
}

func GetEquipEvolutionNeed(slot_idx int, from, to uint32) *CostData {
	if to <= from {
		return &CostData{}
	}

	if slot_idx < 0 || slot_idx >= len(gdEquipEvolutionNeedData) {
		return nil
	}

	c := gdEquipEvolutionNeedData[slot_idx]

	if from == (to - 1) {
		return &c[to]
	}

	re := &CostData{}
	// 注意 表中i项表示从i-1级升到第i级所需的
	for i := from; i < to; i++ {
		re.AddGroup(&c[i+1])
	}

	return re
}
