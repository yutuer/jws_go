package gamedata

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	HeroUnlockTypAuto = iota
	HeroUnlockTypAutoByCond
	HeroUnlockTypByOp
	HeroUnlockTypCount
)

type playerHeroLvInfo struct {
	Star               uint32
	PieceNumToThisStar uint32
	ATK                float32
	DEF                float32
	HP                 float32
	CoinCost           string
	CoinCount          uint32
	IdId               string
	Cfg                *ProtobufGen.HEROSTAR
}

type playerHeroInfo struct {
	HeroId                string
	HeroIdx               int
	RareLv                uint32
	Ranging               uint32
	Piece                 string
	UnlockInitLv          uint32
	UnlockTyp             uint32
	UnlockPieceNeed       uint32
	IsInCurrVersion       bool
	GsAddon               float32
	LvData                []playerHeroLvInfo
	Nationality           uint32 //国籍
	Sex                   uint32 //性别
	StatSoulPart          string //星魂属性类型
	SurplusCurrencyId     string // 主将多余碎片兑换的软通ID
	SurplusCurrencyCount  uint32 // 兑换比例 N
	SurplusCurrencyCount2 uint32 // 兑换比例 M   N个碎片兑换M个货币
}

var (
	gdPlayerHeroInfo        []playerHeroInfo
	gdPlayerHeroID2IDx      map[string]int
	gdPlayerHeroPieceID2IDx map[string]int
	gdPlayerHeroLevelExp    map[int32]int32
	gdHeroConfig            *ProtobufGen.HEROCONFIG
	gdHeroLevelItem         map[string]*ProtobufGen.HEROLEVELITEM
	gdHeroIdxCountry        map[int]int // <武将idx，国家>
	gdHeroIdxType           map[int]int //<武将idx，种类>
)

func GetHeroData(idx int) *playerHeroInfo {
	if idx < 0 || idx >= len(gdPlayerHeroInfo) {
		return nil
	}
	return &gdPlayerHeroInfo[idx]
}
func GetHeroCommonConfig() *ProtobufGen.HEROCONFIG {
	return gdHeroConfig
}

func loadHeroData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)
	ar := &ProtobufGen.HERO_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	data := ar.GetItems()
	gdPlayerHeroInfo = make([]playerHeroInfo, len(data), len(data))
	gdPlayerHeroID2IDx = make(map[string]int, len(data))
	gdPlayerHeroPieceID2IDx = make(map[string]int, len(data))
	gdHeroIdxCountry = make(map[int]int, len(data))
	gdHeroIdxType = make(map[int]int, len(data))
	for _, v := range data {
		gdPlayerHeroInfo[int(v.GetID())] = playerHeroInfo{
			HeroId:                v.GetHeroID(),
			HeroIdx:               int(v.GetID()),
			RareLv:                v.GetRareLevel(),
			Ranging:               v.GetHeroRanging(),
			Piece:                 v.GetHeroPiece(),
			UnlockInitLv:          v.GetHeroInitialStar(),
			UnlockTyp:             v.GetActivateType(),
			IsInCurrVersion:       v.GetIsInThisVision() != 0,
			GsAddon:               v.GetTPVPBonus(),
			LvData:                make([]playerHeroLvInfo, 0, 8),
			Nationality:           v.GetHeroNationality(),
			Sex:                   v.GetHeroSex(),
			StatSoulPart:          v.GetStarSoulPart(),
			SurplusCurrencyId:     v.GetHeroCurrency(),
			SurplusCurrencyCount:  v.GetCurrencyAmount(),
			SurplusCurrencyCount2: v.GetCurrencyAmount2(),
		}
		gdPlayerHeroPieceID2IDx[v.GetHeroPiece()] = int(v.GetID())
		gdPlayerHeroID2IDx[v.GetHeroID()] = int(v.GetID())
		gdHeroIdxCountry[int(v.GetID())] = int(v.GetHeroNationality())
		gdHeroIdxType[int(v.GetID())] = int(v.GetHeroType())
	}
	logs.Trace("gdPlayerHeroID2IDx %v", gdPlayerHeroID2IDx)
}

func loadHeroStarData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.HEROSTAR_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	data := ar.GetItems()
	for _, v := range data {
		idx, ok := gdPlayerHeroID2IDx[v.GetGeneralID()]
		if !ok {
			panic(errors.New(
				fmt.Sprintf("PlayerHeroID No Found %s", v.GetGeneralID())))
		}

		lv := playerHeroLvInfo{
			Star:               v.GetStarLevel(),
			PieceNumToThisStar: v.GetPieceNum(),
			CoinCost:           v.GetCoinCost(),
			CoinCount:          v.GetCoinCount(),
			IdId:               v.GetIdid(),
			Cfg:                v,
		}
		for _, attr := range v.GetGeneralProperty_Template() {
			switch attr.GetProperty() {
			case Attr_Atk:
				lv.ATK = attr.GetValue()
			case Attr_Def:
				lv.DEF = attr.GetValue()
			case Attr_HP:
				lv.HP = attr.GetValue()
			default:
				logs.Error("UnKnown Attr In Hero for type: %s", attr.GetProperty())
				//panic(errors.New("UnKnown Attr In Hero"))
			}
		}

		for int(v.GetStarLevel()) >= len(gdPlayerHeroInfo[idx].LvData) {
			gdPlayerHeroInfo[idx].LvData = append(gdPlayerHeroInfo[idx].LvData,
				playerHeroLvInfo{})
		}
		gdPlayerHeroInfo[idx].LvData[int(v.GetStarLevel())] = lv
	}

	logs.Trace("gdPlayerHeroPieceID2IDx %v", gdPlayerHeroPieceID2IDx)
}

func loadHeroConfigData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.HEROCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	data := ar.GetItems()
	gdHeroConfig = data[0]
}

func loadHeroLevelData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.HEROLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	data := ar.GetItems()
	gdPlayerHeroLevelExp = make(map[int32]int32, len(data))
	for _, v := range data {
		gdPlayerHeroLevelExp[v.GetLevel()] = v.GetXP()
	}
}

func mkHeroData(load loadDataFunc) {
	load("hero.data", loadHeroData)
	load("herostar.data", loadHeroStarData)
	load("heroconfig.data", loadHeroConfigData)
	load("herolevel.data", loadHeroLevelData)
	load("herotalent.data", loadTalent)
	load("herotalentlevel.data", loadTalentLevel)
	load("herosoullevel.data", loadHeroSoul)
	load("herolevelitem.data", loadHeroLevelItem)

	for i := 0; i < len(gdPlayerHeroInfo); i++ {
		initLv := int(gdPlayerHeroInfo[i].UnlockInitLv)
		lvData := gdPlayerHeroInfo[i].LvData[:]
		var piece uint32
		piece = lvData[initLv].PieceNumToThisStar
		logs.Trace("Hero %d PieceNeed %d by %v", i, piece, lvData)
		gdPlayerHeroInfo[i].UnlockPieceNeed = piece
	}
}

func GetHeroByHeroID(heroId string) int {
	if id, ok := gdPlayerHeroID2IDx[heroId]; ok {
		return id
	} else {
		return -1
	}
}

func GetHeroLevelExpLimit(lvl int32) (r int32, b bool) {
	r, b = gdPlayerHeroLevelExp[lvl]
	return
}

func loadHeroLevelItem(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.HEROLEVELITEM_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	data := ar.GetItems()
	gdHeroLevelItem = make(map[string]*ProtobufGen.HEROLEVELITEM, len(data))
	for _, v := range data {
		gdHeroLevelItem[v.GetHeroLevelItemID()] = v
	}
}

func GetHeroLevelItem(item string) *ProtobufGen.HEROLEVELITEM {
	return gdHeroLevelItem[item]
}

func IsHeroPiece(itemid string) bool {
	_, ok := gdPlayerHeroPieceID2IDx[itemid]
	return ok
}

func GetHeroCountry(heroIdx int) int {
	return gdHeroIdxCountry[heroIdx]
}

func GetHeroType(heroIdx int) int {
	return gdHeroIdxType[heroIdx]
}
