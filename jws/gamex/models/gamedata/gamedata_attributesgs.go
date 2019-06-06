package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

func loadAttributesgsConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.ATTRIBUTESGS_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	items := dataList.GetItems()
	gdPlayerAtt.GSRadio = *items[0]
	//logs.Trace("loadAttributesgsConfig %v", gdPlayerAtt.GSRadio)
}
