package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type trickAbstractLimit struct {
	Lower int
	Upper int
	Num   int64
	Space int64
}

var gdTrickAbstractLimit []trickAbstractLimit

func GetEquipTrickSelectData(c int) trickAbstractLimit {
	for _, t := range gdTrickAbstractLimit {
		if t.Lower <= c && c <= t.Upper {
			return t
		}
	}
	return gdTrickAbstractLimit[len(gdTrickAbstractLimit)-1]
}

func loadEquipTrickSelectData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.EUIPTRICKSELECT_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	item_data := ar.GetItems()
	l := len(item_data)
	gdTrickAbstractLimit = make([]trickAbstractLimit, l, l)

	for i, a := range item_data {
		gdTrickAbstractLimit[i] = trickAbstractLimit{
			Lower: int(a.GetTrickLowerLimit()) - 1,
			Upper: int(a.GetTrickUpperLimit()) - 1,
			Num:   int64(a.GetTrickSelectN()),
			Space: int64(a.GetTrickSelectM()),
		}
	}

	logs.Trace("gdTrickAbstractLimit %v", gdTrickAbstractLimit)
}
