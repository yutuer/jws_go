package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var RecoverConfigList []*ProtobufGen.RECOVERRESOURCES

func loadOfflineRecover(filepath string) {
	errCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errCheck(err)

	dataList := &ProtobufGen.RECOVERRESOURCES_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)
	RecoverConfigList = dataList.Items
}

func GetOfflineRecoverConfig(id string) *ProtobufGen.RECOVERRESOURCES {
	for _, cfg := range RecoverConfigList {
		if cfg.GetResourcesID() == id {
			return cfg
		}
	}
	return nil
}

func GetAllOfflineRecoverConfigs() []*ProtobufGen.RECOVERRESOURCES {
	return RecoverConfigList
}
