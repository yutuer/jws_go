package gs

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/general"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

type DroidAccountGsCalculateAdapter struct {
	D *gamedata.DroidAccountData
}

func (a *DroidAccountGsCalculateAdapter) GetAcID() string {
	return a.D.AcID
}

func (a *DroidAccountGsCalculateAdapter) GetCorpLv() uint32 {
	return a.D.CorpLv
}
func (a *DroidAccountGsCalculateAdapter) GetUnlockedAvatar() []int {
	return a.D.AvatarLockeds[:]
}
func (a *DroidAccountGsCalculateAdapter) GetArousalLv(t int) uint32 {
	if t < 0 || t >= len(a.D.Arousals) {
		return 0
	}
	return a.D.Arousals[t]
}

func (a *DroidAccountGsCalculateAdapter) CurrEquipInfo() ([]uint32, []uint32, []uint32, []uint32, [][]bool) {
	//eqs, lv_evo, lv_star
	eqs := make([]uint32, len(a.D.EquipsItemID), len(a.D.EquipsItemID))
	for i := 0; i < len(a.D.EquipsItemID); i++ {
		if a.D.EquipsItemID[i] != "" {
			eqs[i] = uint32(i + 1)
		} else {
			eqs[i] = 0
		}
	}
	mat_enh := make([][]bool, len(a.D.EquipMatEhn))
	for i, ms := range a.D.EquipMatEhn {
		m := make([]bool, len(ms))
		copy(m, ms[:])
		mat_enh[i] = m
	}
	return eqs[:], a.D.EquipUpgrade[:], a.D.EquipStar[:], a.D.EquipMatEnhLv[:], mat_enh
}
func (a *DroidAccountGsCalculateAdapter) GetItem(itemID uint32) *bag.BagItem {
	id := int(itemID)
	if id <= 0 || id >= len(a.D.EquipsItemID)+1 {
		return nil
	} else {

		return &bag.BagItem{
			ID:      itemID,
			TableID: a.D.EquipsItemID[id-1],
			ItemID:  a.D.EquipsItemID[id-1],
			Count:   1,
			ItemData: gamedata.BagItemData{
				TrickGroup: a.D.EquipsItemTrickGroup[id-1],
			},
		}
	}
}
func (a *DroidAccountGsCalculateAdapter) GetPost() (string, int64) {
	return "", 0
}
func (a *DroidAccountGsCalculateAdapter) GetProfileNowTime() int64 {
	return time.Now().Unix()
}

func (a *DroidAccountGsCalculateAdapter) GetAllGeneral() []general.General {
	res := make([]general.General, 0, len(a.D.Generals))
	for idx, generalID := range a.D.Generals {
		res = append(res, general.General{
			Id:     generalID,
			StarLv: a.D.GeneralsStar[idx],
			Num:    1,
		})
	}
	return res[:]
}
func (a *DroidAccountGsCalculateAdapter) GetAllGeneralRel() []general.Relation {
	res := make([]general.Relation, 0, len(a.D.GeneralRels))
	for idx, generalID := range a.D.GeneralRels {
		res = append(res, general.Relation{
			Id:    generalID,
			Level: a.D.GeneralRelsStar[idx],
		})
	}
	return res[:]
}

func (a *DroidAccountGsCalculateAdapter) GetSkillPractices() []uint32 {
	return a.D.SkillPractices[:]
}

func (a *DroidAccountGsCalculateAdapter) GetEquipJades() []uint32 {
	as := make([]uint32, len(a.D.EquipJade), len(a.D.EquipJade))
	for i := 0; i < len(a.D.EquipJade); i++ {
		if a.D.EquipJade[i] != "" {
			as[i] = uint32(i) + 1
		} else {
			as[i] = 0
		}
	}
	return as[:]
}
func (a *DroidAccountGsCalculateAdapter) GetDestinyGeneralJades() []uint32 {
	as := make([]uint32, len(a.D.DestGeneralJade), len(a.D.DestGeneralJade))
	for i := 0; i < len(a.D.DestGeneralJade); i++ {
		if a.D.DestGeneralJade[i] != "" {
			as[i] = uint32(i) + 1 + 1000
		} else {
			as[i] = 0
		}
	}
	return as[:]
}
func (a *DroidAccountGsCalculateAdapter) GetJadeData(itemID uint32) (*ProtobufGen.Item, bool) {
	id := int(itemID)
	jID := ""
	if id >= 1001 {
		if id-1001 >= len(a.D.DestGeneralJade) {
			return nil, false
		}
		jID = a.D.DestGeneralJade[id-1001]
	} else {
		if id-1 >= len(a.D.EquipJade) {
			return nil, false
		}
		jID = a.D.EquipJade[id-1]
	}

	if jID == "" {
		return nil, false
	}

	return gamedata.GetProtoItem(jID)
}

func (a *DroidAccountGsCalculateAdapter) GetLastGeneralGiveGs() *gamedata.DestinyGeneralLevelData {
	return gamedata.GetNewDestinyGeneralLevelData(a.D.DestinyGeneralID, a.D.DestinyGeneralLv)
}

func (a *DroidAccountGsCalculateAdapter) GetFashionAll() []helper.FashionItem {
	res := make([]helper.FashionItem, 0, len(a.D.Fashions))
	for idx, fIDs := range a.D.Fashions {
		for jidx, fID := range fIDs {
			res = append(res, helper.FashionItem{
				ID:              uint32(idx)*10 + uint32(jidx),
				TableID:         fID,
				ExpireTimeStamp: 99999,
			})
		}
	}
	return res[:]
}

func (a *DroidAccountGsCalculateAdapter) GetHeroStarLv() []uint32 {
	return []uint32{1, 1, 1}
}

func (a *DroidAccountGsCalculateAdapter) GetHeroLv() []uint32 {
	return []uint32{1, 1, 1}
}

func (a *DroidAccountGsCalculateAdapter) GetHeroTalent(avatarId int) []uint32 {
	return []uint32{0, 0, 0, 0}
}

func (a *DroidAccountGsCalculateAdapter) GetHeroTalentPointCost(avatarId int) uint32 {
	return 0
}

func (a *DroidAccountGsCalculateAdapter) GetHeroSoulLv() uint32 {
	return 0
}

func (a *DroidAccountGsCalculateAdapter) GetTitle() []string {
	return []string{}
}

func (a *DroidAccountGsCalculateAdapter) GetSwingLv(avatarId int) int {
	return 0
}

func (a *DroidAccountGsCalculateAdapter) GetSwingStarLv(avatarId int) int {
	return 0
}
