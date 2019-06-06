package gamedata

import (
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var hmtgift map[string]string

func LoadHMTGiftData(filepath string) {
	ar := &ProtobufGen.HMTITEMLIST_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()

	hmtgift = make(map[string]string, 0)
	for _, v := range data {
		hmtgift[v.GetId()] = v.GetName()
	}
	logs.Debug("load hmtgift data: %v", hmtgift)
}

func IsHMTGiftItem(id string) bool {
	for key, _ := range hmtgift {
		if key == id {
			return true
		}
	}
	return false
}
