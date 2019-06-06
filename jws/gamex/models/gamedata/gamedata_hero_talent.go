package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	_ = iota
	Talent_Idx_Endurance
	Talent_Idx_Intellect
	Talent_Idx_Force
)

var (
	gdHeroTalent           map[uint32]*ProtobufGen.HEROTALENT
	gdHeroTalentLevel      map[uint32]*ProtobufGen.HEROTALENTLEVEL
	gdHeroStarUnlockTalent map[uint32]uint32
	HeroTalentCount        int
)

func GetHeroTalentConfig(id uint32) *ProtobufGen.HEROTALENT {
	return gdHeroTalent[id]
}

func GetHeroTalentLevelConfig(lvl uint32) *ProtobufGen.HEROTALENTLEVEL {
	return gdHeroTalentLevel[lvl]
}

func GetHeroStarUnlockTalent(star uint32) []uint32 {
	res := make([]uint32, 0, 4)
	for s, t := range gdHeroStarUnlockTalent {
		if star >= s {
			res = append(res, t)
		}
	}
	return res
}

func loadTalent(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.HEROTALENT_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	as := ar.GetItems()
	gdHeroTalent = make(map[uint32]*ProtobufGen.HEROTALENT, len(as))
	gdHeroStarUnlockTalent = make(map[uint32]uint32, len(as))
	for _, v := range as {
		gdHeroTalent[v.GetHeroTalentID()] = v
		gdHeroStarUnlockTalent[v.GetUnlockStarLevel()] = v.GetHeroTalentID()
	}
	HeroTalentCount = len(as)
}

func loadTalentLevel(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.HEROTALENTLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	as := ar.GetItems()
	gdHeroTalentLevel = make(map[uint32]*ProtobufGen.HEROTALENTLEVEL, len(as))
	for _, v := range as {
		gdHeroTalentLevel[v.GetHeroTalentLevel()] = v
	}

}
