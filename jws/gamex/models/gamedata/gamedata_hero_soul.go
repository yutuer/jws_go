package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdHeroSoulLv2Cfg map[uint32]*ProtobufGen.HEROSOULLEVEL
)

func GetHeroSoulLvlConfig(lvl uint32) *ProtobufGen.HEROSOULLEVEL {
	return gdHeroSoulLv2Cfg[lvl]
}
func loadHeroSoul(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.HEROSOULLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	as := ar.GetItems()

	gdHeroSoulLv2Cfg = make(map[uint32]*ProtobufGen.HEROSOULLEVEL, len(as))
	for _, r := range as {
		gdHeroSoulLv2Cfg[r.GetHeroSoulLevel()] = r
	}
}
