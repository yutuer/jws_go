package gamedata

import (
	"fmt"

	"errors"
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

type general_star_k struct {
	generalId string
	star      uint32
}

type general_Rel_star_k struct {
	generalRelId string
	level        uint32
}

const (
	GeneralTypeInRelHero    = 0
	GeneralTypeInRelGeneral = 1
	GeneralTypeInRelNull    = 2
)

type GeneralRelInfo struct {
	Generals          []string
	GeneralTypes      []uint32
	HeroIdxIfTypeHero []int
}

func (g GeneralRelInfo) String() string {
	return fmt.Sprintf("%v", g.Generals)
}

var (
	gdGeneralInfo       map[string]*ProtobufGen.NEWGENERAL
	gdGeneralStarInfo   map[general_star_k]*ProtobufGen.GENERALSTART
	gdGeneralRelInfo    map[string]*GeneralRelInfo
	gdGeneralRelLvlInfo map[general_Rel_star_k]*ProtobufGen.RELATIONLEVEL
)

func IsGeneral(gid string) bool {
	_, ok := gdGeneralInfo[gid]
	return ok
}

func GeneralStarNeedNum(generalId string, star uint32) (bool, uint32) {
	info, ok := gdGeneralStarInfo[general_star_k{generalId, star}]
	if ok {
		return true, info.GetPieceNum()
	}
	return false, 0
}

func GeneralStarCfg(generalId string, star uint32) *ProtobufGen.GENERALSTART {
	return gdGeneralStarInfo[general_star_k{generalId, star}]
}

func GetGeneralInfo(generalId string) *ProtobufGen.NEWGENERAL {
	return gdGeneralInfo[generalId]
}

func GetGeneralStarAttr(generalId string, star uint32) (atk, def, hp float32) {
	cfg := gdGeneralStarInfo[general_star_k{generalId, star}]
	if cfg != nil {
		for _, attr := range cfg.GetGeneralProperty_Template() {
			switch attr.GetProperty() {
			case Attr_Atk:
				atk = attr.GetValue()
			case Attr_Def:
				def = attr.GetValue()
			case Attr_HP:
				hp = attr.GetValue()
			}
		}
	}
	return
}

func GetGeneralRelLvlAttr(generalRelId string, level uint32) (atk, def, hp float32) {
	cfg := gdGeneralRelLvlInfo[general_Rel_star_k{generalRelId, level}]
	if cfg != nil {
		for _, attr := range cfg.GetRelationProperty_Template() {
			switch attr.GetProperty() {
			case Attr_Atk:
				atk = attr.GetValue()
			case Attr_Def:
				def = attr.GetValue()
			case Attr_HP:
				hp = attr.GetValue()
			}
		}
	}
	return
}

func GetGeneralRelationInfo(relation string) *GeneralRelInfo {
	return gdGeneralRelInfo[relation]
}

func loadGeneralCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.NEWGENERAL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGeneralInfo = make(map[string]*ProtobufGen.NEWGENERAL, len(dataList.Items))
	for _, item := range dataList.Items {
		gdGeneralInfo[item.GetGeneralID()] = item
	}
}

func loadGeneralStarCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GENERALSTART_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGeneralStarInfo = make(map[general_star_k]*ProtobufGen.GENERALSTART, len(dataList.Items))
	for _, item := range dataList.Items {
		gdGeneralStarInfo[general_star_k{item.GetGeneralID(), item.GetStarLevel()}] = item
	}
}

func loadGeneralRelCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.GENERALRELATION_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGeneralRelInfo = make(map[string]*GeneralRelInfo, len(dataList.Items))
	for _, item := range dataList.Items {
		l := len(item.GetRelationGeneral_Template())
		v := GeneralRelInfo{
			Generals:          make([]string, 0, l),
			GeneralTypes:      make([]uint32, 0, l),
			HeroIdxIfTypeHero: make([]int, 0, l),
		}
		for _, gen := range item.GetRelationGeneral_Template() {
			v.Generals = append(v.Generals, gen.GetGeneralID())
			v.GeneralTypes = append(v.GeneralTypes, gen.GetGeneralType())
			if gen.GetGeneralType() == GeneralTypeInRelHero {
				heroIdx, ok := gdPlayerHeroID2IDx[gen.GetGeneralID()]
				if !ok {
					panic(errors.New("unknown heroid " + gen.GetGeneralID()))
				} else {
					v.HeroIdxIfTypeHero = append(v.HeroIdxIfTypeHero, heroIdx)
				}
			} else {
				v.HeroIdxIfTypeHero = append(v.HeroIdxIfTypeHero, -1)
			}

			if gen.GetGeneralType() == GeneralTypeInRelGeneral {
				if _, ok := gdGeneralInfo[gen.GetGeneralID()]; !ok {
					panic(fmt.Errorf("GeneralRelation[%s]'s general %s not define", item.GetRelationID(), gen.GetGeneralID()))
				}
			}
		}
		gdGeneralRelInfo[item.GetRelationID()] = &v
	}
	// check
	for k, v := range gdGeneralInfo {
		if v.GetGeneralRelation() != "" {
			relInfo, ok := gdGeneralRelInfo[v.GetGeneralRelation()]
			if !ok {
				panic(fmt.Errorf("general[%s]'s GeneralRelation %s not define in GENERALRELATION", k, v.GetGeneralRelation()))
			}
			right := false
			for _, gen_n := range relInfo.Generals {
				if gen_n == k {
					right = true
					break
				}
			}
			if !right {
				panic(fmt.Errorf("general[%s]'s GeneralRelation %s not contain this general %s", k, v.GetGeneralRelation(), k))
			}
		}
	}
}

func loadGeneralRelLevelCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.RELATIONLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdGeneralRelLvlInfo = make(map[general_Rel_star_k]*ProtobufGen.RELATIONLEVEL, len(dataList.Items))
	for _, item := range dataList.Items {
		gdGeneralRelLvlInfo[general_Rel_star_k{item.GetRelationID(), item.GetRelationLevel()}] = item
	}
}
