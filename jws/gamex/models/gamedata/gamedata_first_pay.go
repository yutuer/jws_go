package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdFirstPay map[uint32]*ProtobufGen.FIRSTPAY
)

func loadFirstPayCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.FIRSTPAY_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdFirstPay = make(map[uint32]*ProtobufGen.FIRSTPAY, len(dataList.GetItems()))
	for _, d := range dataList.GetItems() {
		gdFirstPay[d.GetFirstPayID()] = d
	}
}

func GetFirstPayConfig(id uint32) *ProtobufGen.FIRSTPAY {
	return gdFirstPay[id]
}
