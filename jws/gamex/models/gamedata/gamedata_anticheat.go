package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gbAntiCheatCommon *ProtobufGen.ANTIRATIO
	gbAntiSkillCfg    map[uint32]*ProtobufGen.SKILLRATIO
)

func loadAntiCheatRatioConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	arr := &ProtobufGen.ANTIRATIO_ARRAY{}
	err = proto.Unmarshal(buffer, arr)
	errcheck(err)
	items := arr.GetItems()

	gbAntiCheatCommon = items[0]
}

func GetAntiCheatCommon() *ProtobufGen.ANTIRATIO {
	return gbAntiCheatCommon
}

func loadAntiCheatSkillRatioConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	array := &ProtobufGen.SKILLRATIO_ARRAY{}
	err = proto.Unmarshal(buffer, array)
	errcheck(err)

	items := array.GetItems()
	gbAntiSkillCfg = make(map[uint32]*ProtobufGen.SKILLRATIO, len(items))
	for _, skill := range items {
		gbAntiSkillCfg[skill.GetSkillLevel()] = skill
	}
}

func GetAntiCheatSkillCfg(skillLevel uint32) *ProtobufGen.SKILLRATIO {
	return gbAntiSkillCfg[skillLevel]
}
