package gamedata

import (
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var ucgift map[string]*ProtobufGen.UCGIFT

func LoadUCGiftData(filepath string) {
	ar := &ProtobufGen.UCGIFT_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	gdGVGCity = make([]int, 0, len(data))
	ucgift = make(map[string]*ProtobufGen.UCGIFT, 0)
	for _, v := range data {
		ucgift[v.GetGiftID()] = v
	}
	logs.Debug("load ucgift data: %v", ucgift)
}

func GetUCGiftData(id string) *ProtobufGen.UCGIFT {
	return ucgift[id]
}
