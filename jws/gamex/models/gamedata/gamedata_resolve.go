package gamedata

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdEquipResolveGiveData map[string]CostData
)

func GetEquipResolveGive(tier, rare int32) *CostData {
	// TODO 优化查询
	tire_rare := fmt.Sprintf("%d_%d", tier, rare)
	data, is_ok := gdEquipResolveGiveData[tire_rare]
	if is_ok {
		return &data
	} else {
		return nil
	}
}

func loadEquipResolveGiveConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.EQUIPRESOLVE_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdEquipResolveGiveData = make(map[string]CostData, len(dataList.GetItems()))

	for _, a := range dataList.GetItems() {
		tire_rare := a.GetEResolveIndex()
		fine_iron := a.GetFineIron()
		material_typ := a.GetExtraMaterial()
		material_count := a.GetCount()

		//logs.Trace("equip Resolve %s  ->  %v %v %v",
		//	tire_rare, fine_iron, material_typ, material_count)

		//TODO 优化内存分配
		need := CostData{}

		// 这里需要注意，升级的消耗的配的累计值
		if fine_iron > 0 {
			need.AddItem(VI_Sc1, fine_iron)

		}
		if material_count > 0 && material_typ != "" {
			need.AddItem(material_typ, material_count)
		}
		gdEquipResolveGiveData[tire_rare] = need

		//logs.Trace("Resolve Give %s : %v", tire_rare, need)
	}

}
