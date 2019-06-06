package gamedata

import (
	"sort"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	LampType_Trial                  = 42
	LampType_SimplePvp              = 999
	LampType_TeamPvp                = 998
	LampType_GachaHero              = 997
	LampType_Vip                    = 15
	LampType_EquipStar              = 35
	LampType_HeroUnlock             = 57
	LampType_HeroStar               = 58
	LampType_DestinyGeneralActivate = 56
	LampType_GachaHeroWhole         = 996
	LampType_GVGWinStreak           = 74
)

type SysNoticeEquipStar struct {
	Star uint32
	Cond Condition
	Cfg  *ProtobufGen.HORSELAMP
}
type SysNoticeEquipStarSlice []SysNoticeEquipStar

func (pq SysNoticeEquipStarSlice) Len() int      { return len(pq) }
func (pq SysNoticeEquipStarSlice) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }
func (pq SysNoticeEquipStarSlice) Less(i, j int) bool {
	return pq[i].Star > pq[j].Star
}

var (
	gdTrialLevelId   map[uint32]uint32
	gdSimplePvp      map[uint32]*ProtobufGen.HORSELAMP
	gdTeamPvp        map[uint32]*ProtobufGen.HORSELAMP
	gdVip            map[uint32]*ProtobufGen.HORSELAMP
	gdEquipStar      []SysNoticeEquipStar
	gdHeroUnlock     *ProtobufGen.HORSELAMP
	gdHeroStar       map[uint32]*ProtobufGen.HORSELAMP
	gdDGAct          map[uint32]*ProtobufGen.HORSELAMP
	gdGachaHero      *ProtobufGen.HORSELAMP
	gdGachaHeroWhole *ProtobufGen.HORSELAMP
	gdGVGWinStreak   *ProtobufGen.HORSELAMP
)

func loadHorseLamp(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := &ProtobufGen.HORSELAMP_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	panicIfErr(err)

	gdTrialLevelId = make(map[uint32]uint32, 5)
	gdVip = make(map[uint32]*ProtobufGen.HORSELAMP, 5)
	gdEquipStar = make([]SysNoticeEquipStar, 0, 5)
	gdHeroStar = make(map[uint32]*ProtobufGen.HORSELAMP, 5)
	gdSimplePvp = make(map[uint32]*ProtobufGen.HORSELAMP, 5)
	gdTeamPvp = make(map[uint32]*ProtobufGen.HORSELAMP, 5)
	gdDGAct = make(map[uint32]*ProtobufGen.HORSELAMP, 5)
	data := ar.GetItems()
	for _, item := range data {
		switch int(item.GetLampType()) {
		case LampType_Trial:
			gdTrialLevelId[item.GetLampValueIP1()] = item.GetServerMsgID()
		case LampType_SimplePvp:
			gdSimplePvp[item.GetLampValueIP1()] = item
		case LampType_TeamPvp:
			gdTeamPvp[item.GetLampValueIP1()] = item
		case LampType_Vip:
			gdVip[item.GetLampValueIP1()] = item
		case LampType_EquipStar:
			gdEquipStar = append(gdEquipStar, SysNoticeEquipStar{
				Star: item.GetLampValueIP1(),
				Cond: Condition{
					Ctyp:   item.GetLampType(),
					Param1: int64(item.GetLampValueIP1()),
					Param2: int64(item.GetLampValueIP2()),
				},
				Cfg: item,
			})
		case LampType_HeroUnlock:
			gdHeroUnlock = item
		case LampType_HeroStar:
			gdHeroStar[item.GetLampValueIP1()] = item
		case LampType_DestinyGeneralActivate:
			gdDGAct[item.GetLampValueIP1()] = item
		case LampType_GachaHero:
			gdGachaHero = item
		case LampType_GachaHeroWhole:
			gdGachaHeroWhole = item
		case LampType_GVGWinStreak:
			gdGVGWinStreak = item
		}
	}
	sort.Sort(SysNoticeEquipStarSlice(gdEquipStar))
}

func IsTrialLevelSysNotice(lvl uint32) bool {
	_, ok := gdTrialLevelId[lvl]
	return ok
}

func Trial2SysNotice(lvl uint32) uint32 {
	return gdTrialLevelId[lvl]
}

func SimplePvpSysNotic(rank uint32) *ProtobufGen.HORSELAMP {
	return gdSimplePvp[rank]
}

func TeamPvpSysNotic(rank uint32) *ProtobufGen.HORSELAMP {
	return gdTeamPvp[rank]
}

func VipSysNotic(vip uint32) *ProtobufGen.HORSELAMP {
	return gdVip[vip]
}

func EquipStarSysNotice() []SysNoticeEquipStar {
	return gdEquipStar
}

func HeroUnlockSysNotice() *ProtobufGen.HORSELAMP {
	return gdHeroUnlock
}

func HeroStarSysNotice(star uint32) *ProtobufGen.HORSELAMP {
	return gdHeroStar[star]
}

func ActDestingGeneralSysNotice(id uint32) *ProtobufGen.HORSELAMP {
	return gdDGAct[id]
}

func GachaHeroSysNotice() *ProtobufGen.HORSELAMP {
	return gdGachaHero
}

func GachaHeroWholeSysNotice() *ProtobufGen.HORSELAMP {
	return gdGachaHeroWhole
}

func GVGWinStreakNotice() *ProtobufGen.HORSELAMP {
	return gdGVGWinStreak
}
