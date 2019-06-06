package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type CompanionEvolveConfig struct {
	HeroIdx int
	Config  *ProtobufGen.RELATIONEVOLUTION
}

var companionEvolveArray []*CompanionEvolveConfig
var MaxCompanionLevel uint32 // 记录配置表的最大情缘等级

func loadCompanionEvolve(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.RELATIONEVOLUTION_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)
	companionEvolveArray = make([]*CompanionEvolveConfig, len(data_ar.GetItems()))
	for i, data := range data_ar.GetItems() {
		companionEvolveArray[i] = &CompanionEvolveConfig{
			HeroIdx: GetHeroByHeroID(data.GetHeroID()),
			Config:  data,
		}
		if data.GetRelationLevel() > MaxCompanionLevel {
			MaxCompanionLevel = data.GetRelationLevel()
		}
	}
}

func GetCompanionEvolveConfigByLess(heroIdx, evolveLevel int) []*CompanionEvolveConfig {
	retArray := make([]*CompanionEvolveConfig, 0, MaxCompanionLevel) // TODO 3的处理
	for _, config := range companionEvolveArray {
		if config.HeroIdx == heroIdx && config.Config.GetRelationLevel() <= uint32(evolveLevel) {
			retArray = append(retArray, config)
		}
	}
	logs.Debug("calc companion %d %d %d", heroIdx, evolveLevel, len(retArray))
	return retArray
}
