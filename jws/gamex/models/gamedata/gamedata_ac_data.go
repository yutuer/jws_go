package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdAcData map[string]*ProtobufGen.ACDATA
)

func GetAcData(idx string) (*ProtobufGen.ACDATA, bool) {
	data, ok := gdAcData[idx]
	return data, ok
}

func GetAcLootsMaxTimes(idx string) uint32 {
	if templates, ok := gdAcData[idx]; ok {
		var sum uint32
		loots := templates.GetLoots()
		for _, t := range loots {
			sum += t.GetLootTimes()
		}
		return sum
	}
	return 0
}

func GetAcLoots(idx string) ([]*ProtobufGen.ACDATA_LootRule, bool) {
	data, ok := gdAcData[idx]
	if !ok {
		return nil, false
	}

	return data.GetLoots()[:], true
}

func loadAcData(filepath string) {
	gdAcData = make(map[string]*ProtobufGen.ACDATA)

	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	acDataList := &ProtobufGen.ACDATA_ARRAY{}
	err = proto.Unmarshal(buffer, acDataList)
	errcheck(err)

	for _, a := range acDataList.GetItems() {
		//logs.Trace("loots %s  ->  %v", a.GetID(), a.GetLoots())
		//for k, v := range a.Loots {
		//logs.Trace("loots kv %s, %s", k, v)
		//}
		gdAcData[a.GetID()] = a
	}

}
