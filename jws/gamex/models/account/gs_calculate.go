package account

import (
	"sort"
	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/battlearmy"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/general"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type ProfileData struct {
	CorpAttrs gamedata.AvatarAttr `json:"corp_attrs"` // To Be Delete

	SubModuleGs []int `json:"module_gs"` // To Be Delete

	HeroBaseAttrs         [AVATAR_NUM_MAX]gamedata.AvatarAttr `json:"hero_base_attrs"`
	HeroGs                [AVATAR_NUM_MAX]int                 `json:"hero_gs"` //每个英雄的战斗力
	HeroBaseGs            [AVATAR_NUM_MAX]int                 `json:"hero_base_gs"`
	CorpCurrGS            int                                 `json:"corp_gs"`
	CorpCurrGS_HistoryMax int                                 `json:"corp_gs_hmax"`
	HeroBaseGSSum         int                                 `json:"hero_base_gs_sum"`
	HeroBaseGSSum_Max     int                                 `json:"hero_base_gs_sum_max"`

	BestHeroAvatar []int `json:"best_hero_id"`

	isNeedCheckCorpMaxGS bool

	//	Attrs            [AVATAR_NUM_CURR]gamedata.AvatarAttr `json:"attrs"` // to be delete
	//	MaxGS            [AVATAR_NUM_CURR]int                 `json:"gs"`    // to be delete
	//	isNeedCheckMaxGS [AVATAR_NUM_CURR]bool                // to be delete

	is_inited bool

	// 在线奖励专用的计时
	OnlineTimeBegin int64 `json:"tbegin"`

	FarthestStageIndex      int32 `json:"fsi"`
	FarthestEliteStageIndex int32 `json:"fesi"`
	FarthestHellStageIndex  int32 `json:"fhsi"`

	Times_1V1 int `json:"ts_1v1"` // 历史1v1次数
	Times_3V3 int `json:"ts_3v3"` // 历史3v3次数

	isNeedCheckGS        bool
	lastGsUpdateTime     int64
	isNeedCheckCompanion bool

	needUpdateFriend bool
}

func (data *ProfileData) SetNeedUpdateFriend(flag bool) {
	data.needUpdateFriend = true
}

func (data *ProfileData) IsNeedUpdateFriend() bool {
	return data.needUpdateFriend
}

// 存档升级， HeroBaseGSSum_Max是新加的
func (data *ProfileData) OnAfterLogin() {
	if data.HeroBaseGSSum > data.HeroBaseGSSum_Max {
		data.HeroBaseGSSum_Max = data.HeroBaseGSSum
	}
}

func (data *ProfileData) GetHeroGsInfo() (
	bestHero []int, gsHeroGs []int, gsHeroBaseGs []int) {
	gsHeroGs = make([]int, 0, 3)
	gsHeroBaseGs = make([]int, 0, 3)
	for _, h := range data.BestHeroAvatar {
		gsHeroGs = append(gsHeroGs, data.HeroGs[h])
		gsHeroBaseGs = append(gsHeroBaseGs, data.HeroBaseGs[h])
	}
	return data.BestHeroAvatar, gsHeroGs, gsHeroBaseGs
}

func (data *ProfileData) GetSortHeroGsInfo() []int {
	gsHeroGs := make([]int, 0)
	for _, gs := range data.HeroGs {
		if gs == 0 {
			continue
		}
		gsHeroGs = append(gsHeroGs, gs)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(gsHeroGs)))
	logs.Debug("sortherogs %v", gsHeroGs)
	return gsHeroGs
}

func (data *ProfileData) GetCurrGS(a *Account) int {
	heroBase, heroAttr, heroBaseGs, heroGs, bestHero, gs, _ := gs.GetCurrAttr(NewAccountGsCalculateAdapter(a))

	for i, base := range heroBase {
		data.HeroBaseAttrs[i] = *base
	}
	var _bgss int
	for i, bgs := range heroBaseGs {
		_bgss += bgs
		data.HeroBaseGs[i] = bgs
	}
	for i, hgs := range heroGs {
		data.HeroGs[i] = hgs
	}
	data.CorpCurrGS = gs
	data.setCorpGsHistoryMax(gs)
	data.SetHeroBaseGSSum(_bgss)
	data.BestHeroAvatar = bestHero
	logs.Trace("%s GetCurrGS corpgs %d heroBase %v heroAttr %v "+
		"heroGs %v heroBaseGs %v HeroBaseGSSum %d BestHeroAvatar %v",
		a.AccountID.String(),
		gs, heroBase, heroAttr, heroGs, heroBaseGs,
		data.HeroBaseGSSum, data.BestHeroAvatar)
	return gs
}

func (data *ProfileData) SetHeroBaseGSSum(base_gs_sum int) {
	data.HeroBaseGSSum = base_gs_sum
	if base_gs_sum > data.HeroBaseGSSum_Max {
		data.HeroBaseGSSum_Max = base_gs_sum
	}
}

func (p *ProfileData) setCorpGsHistoryMax(gs int) {
	if gs > p.CorpCurrGS_HistoryMax {
		p.CorpCurrGS_HistoryMax = gs
	}
}

func (p *ProfileData) SetLastGsUpdateTime(t int64) {
	p.lastGsUpdateTime = t
}

func (p *ProfileData) GetLastGsUpdateTime() int64 {
	return p.lastGsUpdateTime
}

func (p *ProfileData) SetNeedCheckGS(t bool) {
	p.isNeedCheckGS = t
}

func (p *ProfileData) GetNeedCheckGS() bool {
	return p.isNeedCheckGS
}

func (p *ProfileData) SetOnlineTimeBegin(t int64) {
	p.OnlineTimeBegin = t
}

func (p *ProfileData) GetOnlineTimeBegin() int64 {
	return p.OnlineTimeBegin
}

// 最高GS在如下地方会变化
//  1. 角色升级
//  2. 角色突破
//  3. 装备新装备
//  4. 装备 强化 精炼
//  5. 技能升级
//  6. 装备洗练
//  7. 装备升星
//  8. 时装
//  9. 宝石
// 10. 神将
// 11. 神翼
// 关键字 : MaxGS可能变化
// XXX by Fanyang 注意更新这些位置

func (p *ProfileData) IsNeedInit() bool {
	return !p.is_inited
}
func (p *ProfileData) SetNoNeedInit() {
	p.is_inited = true
}
func (p *ProfileData) IsNeedCheckMaxGS() bool {
	return p.isNeedCheckCorpMaxGS
}

func (p *ProfileData) SetNeedCheckMaxGS() {
	p.isNeedCheckCorpMaxGS = true
}

func (p *ProfileData) SetNoNeedCheckMaxGS() {
	p.isNeedCheckCorpMaxGS = false
}
func (p *ProfileData) IsNeedCheckCompanion() bool {
	return p.isNeedCheckCompanion
}

func (p *ProfileData) SetNeedCheckCompanion(flag bool) {
	p.isNeedCheckCompanion = flag
}

type AccountGsCalculateAdapter struct {
	acc *Account
}

func NewAccountGsCalculateAdapter(a *Account) *AccountGsCalculateAdapter {
	return &AccountGsCalculateAdapter{a}
}

func (a *AccountGsCalculateAdapter) GetAcID() string {
	return a.acc.AccountID.String()
}

func (a *AccountGsCalculateAdapter) GetCorpLv() uint32 {
	corp_lvl, _ := a.acc.Profile.GetCorp().GetXpInfoNoUpdate()
	return corp_lvl
}
func (a *AccountGsCalculateAdapter) GetUnlockedAvatar() []int {
	return a.acc.Profile.GetCorp().GetUnlockedAvatar()
}
func (a *AccountGsCalculateAdapter) GetArousalLv(t int) uint32 {
	return a.acc.Profile.GetAvatarExp().GetArousalLv(t)
}

func (a *AccountGsCalculateAdapter) CurrEquipInfo() ([]uint32, []uint32, []uint32, []uint32, [][]bool) {
	eqs, _, lv_evo, lv_star, lv_mat_enh, mat_enh := a.acc.Profile.GetEquips().CurrInfo()
	return eqs, lv_evo, lv_star, lv_mat_enh, mat_enh
}

func (a *AccountGsCalculateAdapter) GetItem(id uint32) *bag.BagItem {
	return a.acc.BagProfile.GetItem(id)
}
func (a *AccountGsCalculateAdapter) GetPost() (string, int64) {
	return a.acc.GuildProfile.GetPost()
}
func (a *AccountGsCalculateAdapter) GetProfileNowTime() int64 {
	return a.acc.Profile.GetProfileNowTime()
}

func (a *AccountGsCalculateAdapter) GetAllGeneral() []general.General {
	return a.acc.GeneralProfile.GetAllGeneral()
}
func (a *AccountGsCalculateAdapter) GetAllGeneralRel() []general.Relation {
	return a.acc.GeneralProfile.GetAllGeneralRel()
}

func (a *AccountGsCalculateAdapter) GetSkillPractices() []uint32 {
	return a.acc.Profile.GetAvatarSkill().SkillPractices[:]
}

func (a *AccountGsCalculateAdapter) GetEquipJades() []uint32 {
	return a.acc.Profile.GetEquipJades().Jades[:]
}
func (a *AccountGsCalculateAdapter) GetDestinyGeneralJades() []uint32 {
	return a.acc.Profile.GetDestGeneralJades().DestinyGeneralJade[:]
}
func (a *AccountGsCalculateAdapter) GetJadeData(id uint32) (*ProtobufGen.Item, bool) {
	return a.acc.Profile.GetJadeBag().GetJadeData(id)
}

func (a *AccountGsCalculateAdapter) GetLastGeneralGiveGs() *gamedata.DestinyGeneralLevelData {
	return a.acc.Profile.GetDestinyGeneral().GetLastGeneralGiveGs()
}
func (a *AccountGsCalculateAdapter) GetFashionAll() []helper.FashionItem {
	return a.acc.Profile.GetFashionBag().GetFashionAll()
}
func (a *AccountGsCalculateAdapter) GetHeroStarLv() []uint32 {
	return a.acc.Profile.GetHero().HeroStarLevel[:]
}
func (a *AccountGsCalculateAdapter) GetHeroLv() []uint32 {
	return a.acc.Profile.GetHero().HeroLevel[:]
}
func (a *AccountGsCalculateAdapter) GetHeroTalent(avatarId int) []uint32 {
	return a.acc.Profile.GetHeroTalent().HeroTalentLevel[avatarId][:]
}
func (a *AccountGsCalculateAdapter) GetHeroTalentPointCost(avatarId int) uint32 {
	var res uint32
	for _, tl := range a.acc.Profile.GetHeroTalent().HeroTalentLevel[avatarId] {
		res += tl
	}
	return res
}
func (a *AccountGsCalculateAdapter) GetHeroSoulLv() uint32 {
	return a.acc.Profile.GetHeroSoul().HeroSoulLevel
}
func (a *AccountGsCalculateAdapter) GetTitle() []string {
	return a.acc.Profile.GetTitle().GetTitles()
}

func (a *AccountGsCalculateAdapter) GetSwingLv(avatarId int) int {
	return a.acc.Profile.GetHero().HeroSwings[avatarId].Lv
}

func (a *AccountGsCalculateAdapter) GetSwingStarLv(avatarId int) int {
	return a.acc.Profile.GetHero().HeroSwings[avatarId].StarLv
}

func (a *AccountGsCalculateAdapter) GetCompanionActiveData(avatarId int) []*gamedata.CompanionActiveConfig {
	companionInfo := a.acc.Profile.GetHero().HeroCompanionInfos[avatarId]
	ret := make([]*gamedata.CompanionActiveConfig, 0)
	for _, companion := range companionInfo.NewCompanions {
		configs := gamedata.GetAllActiveConfigByLess(avatarId, companion.GetCompanionId(), companion.GetLevel())
		ret = append(ret, configs...)
	}
	return ret
}

func (a *AccountGsCalculateAdapter) GetCompanionEvolveData(avatarId int) []*gamedata.CompanionEvolveConfig {
	companionInfo := a.acc.Profile.GetHero().HeroCompanionInfos[avatarId]
	return gamedata.GetCompanionEvolveConfigByLess(avatarId, companionInfo.EvolveLevel)
}

func (a *AccountGsCalculateAdapter) GetExclusiveWeaponData(avatarId int) (int, []float32) {
	weapon := a.acc.Profile.GetHero().HeroExclusiveWeapon[avatarId]
	return weapon.Quality, weapon.Attr[:]
}

func (a *AccountGsCalculateAdapter) GetHeroDestinyData() ([]int, []int) {
	ids := make([]int, 0)
	levels := make([]int, 0)
	infoList := a.acc.Profile.GetHeroDestiny().GetActivateDestiny()
	for _, info := range infoList {
		ids = append(ids, info.Id)
		levels = append(levels, info.Level)
	}
	return ids, levels
}

func (a *AccountGsCalculateAdapter) GetHeroAstrologySouls(avatarId int) map[uint32]float32 {
	attr := map[uint32]float32{}
	hero := a.acc.Profile.GetAstrology().CheckHero(uint32(avatarId))
	if nil == hero {
		return attr
	}

	holes := hero.GetHoles()
	for _, hole := range holes {
		holeAttr := gamedata.GetAstrologySoulAttr(hole.HoleID, hole.Rare, hole.Upgrade)
		for p, v := range holeAttr {
			attr[p] = attr[p] + v
		}
	}

	logs.Debug("[AccountGsCalculateAdapter] GetHeroAstrologySouls %v", attr)

	return attr
}
func (a *AccountGsCalculateAdapter) GetMagicPet(avatarID int) map[uint32]float32 {
	attr := map[uint32]float32{}
	pet := &a.acc.Profile.GetHero().HeroMagicPets[avatarID].GetPets()[0]
	if nil == pet {
		return attr
	}

	//灵宠战力加成
	MagicPetLvInfo := gamedata.GetMagicPetLvInfo(pet.Lev)
	MagicPetStarInfo := gamedata.GetMagicPetStarInfo(pet.Star)
	MagicPetConfig := gamedata.GetMagicPetConfig()

	if MagicPetLvInfo != nil && MagicPetStarInfo != nil && MagicPetConfig != nil {
		var AtkTalent uint32
		var DefTalent uint32
		var HPTalent uint32

		var AtkNums uint32
		var DefNums uint32
		var HPNums uint32
		for _, data := range pet.Talents {
			if data.Type == 0 {
				AtkTalent += data.Value
				AtkNums += MagicPetLvInfo.GetPetATK()
			}
			if data.Type == 1 {
				DefTalent += data.Value
				DefNums += MagicPetLvInfo.GetPetDEF()
			}
			if data.Type == 2 {
				HPTalent += data.Value
				HPNums += MagicPetLvInfo.GetPetHP()
			}
		}
		//升星百分比
		AtkStarNum := float32(MagicPetStarInfo.GetAttributeScaling()) / 100
		DefStarNum := float32(MagicPetStarInfo.GetAttributeScaling()) / 100
		HPStarNum := float32(MagicPetStarInfo.GetAttributeScaling()) / 100

		//洗练百分比

		talent := float32(pet.CompreTalent) * gamedata.GetMagicPetConfig().GetMinimumUnit()
		preTalentsValue := talent / float32(gamedata.GetMagicPetConfig().GetRandomMulAptitude()) * float32(gamedata.GetMagicPetConfig().GetRandomAptitude())

		AtkChangeTalentsNum := preTalentsValue / float32(MagicPetConfig.GetRandomAptitude()) * MagicPetConfig.GetAptitudeScaling()
		DefChangeTalentsNum := preTalentsValue / float32(MagicPetConfig.GetRandomAptitude()) * MagicPetConfig.GetAptitudeScaling()
		HPChangeTalentsNum := preTalentsValue / float32(MagicPetConfig.GetRandomAptitude()) * MagicPetConfig.GetAptitudeScaling()

		attr[gamedata.ATK] = float32(AtkNums) * (1 + AtkStarNum) * (1 + AtkChangeTalentsNum)
		attr[gamedata.DEF] = float32(DefNums) * (1 + DefStarNum) * (1 + DefChangeTalentsNum)
		attr[gamedata.HP] = float32(HPNums) * (1 + HPStarNum) * (1 + HPChangeTalentsNum)
	}
	return attr
}
func (a *AccountGsCalculateAdapter) GetBattleArmyData(avatarId int) *battlearmy.BattleArmy {
	country := gamedata.GetHeroCountry(int(avatarId))
	return a.acc.Profile.BattleArmys.GetBattleArmy(country)
}
