package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const Companion_Count_Max = 5

type CompanionActiveConfig struct {
	HeroIdx         int
	CompanionIdx    int
	Config          *ProtobufGen.RELATIONACTIVE
	OldCompanionIdx int
}

var companionActiveArray []*CompanionActiveConfig

func loadCompanionActive(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	data_ar := &ProtobufGen.RELATIONACTIVE_ARRAY{}
	err = proto.Unmarshal(buffer, data_ar)
	errcheck(err)
	companionActiveArray = make([]*CompanionActiveConfig, len(data_ar.GetItems()))
	for i, data := range data_ar.GetItems() {
		companionActiveArray[i] = &CompanionActiveConfig{HeroIdx: GetHeroByHeroID(data.GetHeroID()),
			CompanionIdx:    GetHeroByHeroID(data.GetCompanionID()),
			Config:          data,
			OldCompanionIdx: GetHeroByHeroID(data.GetMemoryRelation()),
		}
	}
}

func GetAllActiveConfig(heroIdx, relationLv int) []*CompanionActiveConfig {
	retArray := make([]*CompanionActiveConfig, 0, Companion_Count_Max)
	for _, it := range companionActiveArray {
		if it.HeroIdx == heroIdx && it.Config.GetRelationLevel() == uint32(relationLv) {
			retArray = append(retArray, it)
		}
	}
	return retArray
}

func GetCompanionActiveConfig(heroIdx, companionIdx, evolveLevel int) *CompanionActiveConfig {
	for _, it := range companionActiveArray {
		if it.HeroIdx == heroIdx &&
			it.Config.GetRelationLevel() == uint32(evolveLevel) &&
			it.OldCompanionIdx == companionIdx {
			return it
		}
	}
	return nil
}

// 获取一个武将所有小于指定进化等级的情缘配置
func GetAllActiveConfigByLess(heroIdx, companionIdx, evolveLevel int) []*CompanionActiveConfig {
	retArray := make([]*CompanionActiveConfig, 0, Companion_Count_Max*3) // 目前3个等级, 每个等级最多5个
	for _, it := range companionActiveArray {
		if it.HeroIdx == heroIdx && it.CompanionIdx == companionIdx && it.Config.GetRelationLevel() <= uint32(evolveLevel) {
			retArray = append(retArray, it)
		}
	}
	return retArray
}

func GetCompanionActiveConfigById(id int) *CompanionActiveConfig {
	for _, it := range companionActiveArray {
		if it.Config.GetUniqueID() == uint32(id) {
			return it
		}
	}
	return nil
}
