package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var renameCostConfigs []*ProtobufGen.RENAMECOST

func loadRenameCostData(filepath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errCheck(err)

	dataList := &ProtobufGen.RENAMECOST_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)
	renameCostConfigs = dataList.Items
}

func GetRenameCostConfig(count int) *ProtobufGen.RENAMECOST {
	for _, config := range renameCostConfigs {
		if config.GetReNameTime() == uint32(count) {
			return config
		}
	}

	return renameCostConfigs[len(renameCostConfigs)-1]
}
