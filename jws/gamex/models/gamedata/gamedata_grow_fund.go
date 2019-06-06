package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdGrowFund map[uint32]*ProtobufGen.GROWFUND
)

func loadGrowFundData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GROWFUND_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdGrowFund = make(map[uint32]*ProtobufGen.GROWFUND, len(data))
	for _, item := range data {
		gdGrowFund[item.GetGroupLevel()] = item
	}
}

func GetGrowFund(lvl uint32) *ProtobufGen.GROWFUND {
	return gdGrowFund[lvl]
}
