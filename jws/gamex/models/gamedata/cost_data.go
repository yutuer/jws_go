package gamedata

import (
	"math/rand"

	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func IsItemVirtual(itemid string) bool {
	switch itemid {
	case VI_Sc0:
		return true
	case VI_Sc1:
		return true
	case VI_Hc_Buy:
		return true
	case VI_Hc_Give:
		return true
	case VI_Hc_Compensate:
		return true
	case VI_Hc:
		return true
	case VI_XP:
		return true
	case VI_CorpXP:
		return true
	case VI_EN:
		return true
	case VI_GoldLevelPoint:
		return true
	case VI_ExpLevelPoint:
		return true
	case VI_DCLevelPoint:
		return true
	case VI_BossFightPoint:
		return true
	case VI_HcByVIP:
		return true
	case VI_BossFightRankPoint:
		return true
	case VI_PvpCoin:
		return true
	case VI_BossCoin:
		return true
	case VI_EC:
		return true
	case VI_DC:
		return true
	case VI_GC:
		return true
	case VI_GachaTicket:
		return true
	case VI_StarBlessCoin:
		return true
	case VI_TPVPCoin:
		return true
	case VI_EggKey:
		return true
	case VI_GuildXP:
		return true
	case VI_GuildSP:
		return true
	case VI_BaoZi:
		return true
	case VI_GuildBoss:
		return true
	case VI_ACTIVE_ITEM:
		return true
	case VI_Expedition:
		return true
	case VI_WeaponCoin:
		return true
	case VI_HeroDiffPlay:
		return true
	case VI_WhiteKey:
		return true
	case VI_WuShuangCoin:
		return true
	case VI_HDP_SD:
		return true
	case VI_XZ_SD:
		return true
	case VI_WINE3:
		return true
	case VI_WINE4:
		return true
	case VI_WINE5:
		return true
	case VI_SSXP:
		return true
	case VI_SSLIGHT:
		return true
	case VI_AC:
		return true
	case VI_WB_BUFFCOIN:
		return true
	case VI_HERO_SURPLUS3:
		return true
	case VI_HERO_SURPLUS4:
		return true
	case VI_HERO_SURPLUS5:
		return true
	case VI_PET_LEVEL:
		return true
	case VI_PET_STAR:
		return true
	case VI_PET_STAR2:
		return true
	case VI_PET_APTITUDE:
		return true
	case VI_PET_APTITUDE2:
		return true
	case VI_PETCOIN:
		return true
	case VI_GT_GACHA:
		return true
	case VI_GT_VIP:
		return true
	case VI_GT_HGC:
		return true
	case VI_GT_GWCBOX:
		return true
	case VI_GT_HEROBOX:
		return true
	case VI_WheelCoin:
		return true
	default:
		return false
	}
}

type CostData struct {
	CostData2Client

	Items    []string
	Count    []uint32
	ItemData []BagItemData
	HasEquip bool // 是否Items里有固定id的物品，即装备
	HasJade  bool

	Generals       []string // 注意这两个只有Give的逻辑，Cost时不会扣除
	GCount         []uint32 // 注意这两个只有Give的逻辑，Cost时不会扣除
	Sc             [SC_TYPE_COUNT]int64
	GoldLevelPoint uint32 // 金币关积分
	ExpLevelPoint  uint32 // 经验关积分
	DCLevelPoint   uint32 // 天命关积分

	HeroPiece               [AVATAR_NUM_MAX]uint32
	HeroPieceWholeChar      [AVATAR_NUM_MAX]uint32
	HeroPieceWholeCharCount [AVATAR_NUM_MAX]uint32

	Hc [HC_TYPE_COUNT]int64
	// 注意 对于赠送来说 是分三种钻赠送的，
	// 对于消耗来说，怎样使用钻石是由功能具体决定的
	// 但是要注意这三个值都会扣除

	AvatarIds []int // 主将id，目前用于给主将加经验用

	StarSouls map[string]uint32

	AvatarXp           uint32 // 注意这两个只有Give的逻辑，Cost时不会扣除
	CorpXp             uint32 // 注意这两个只有Give的逻辑，Cost时不会扣除
	Energy             uint32
	WheelCoin          uint32
	BossFightPoint     uint32
	BossFightRankPoint uint32 // 注意这两个只有Give的逻辑，Cost时不会扣除
	EStageId           string // 注意这两个只有Give的逻辑，Cost时不会扣除
	EStageTimes        uint32 // 注意这两个只有Give的逻辑，Cost时不会扣除
	GameModeId         int    // 注意这两个只有Give的逻辑，Cost时不会扣除
	GameModeTimes      int    // 注意这两个只有Give的逻辑，Cost时不会扣除

	HeroTalentPoint uint32 // 注意只有Give的逻辑，Cost时不会扣除

	GuildXp           uint32  // 工会经验，只有在有工会的时候会加；只有Give的逻辑，Cost时不会扣除
	GuildScienceValid GST_Typ // 工会科技点加成，只有在有工会的时候会加；只有Give的逻辑，Cost时不会扣除

	GateEnemyBonus float32 // 兵临城下加成，只有在有工会的时候会加；只有Give的逻辑，Cost时不会扣除

	GiveHcFromVip int // 如果这个不为0 则按照vip等级中的Hc数量给hc, 之所以只加一个标签, 是因为给的数额是和VIP等级相关的

	IAPGoodIndex   uint32 // iap物品index
	IAPGameOrderId string
	IAPGoodOrder   string
	IAPMoney       uint32
	IAPPlatform    string
	IAPTrueAmount  uint32
	IAPChannel     string
	IAPPayTime     string // 客户端带过来的支付时间
	IAPPkgInfo     PackageInfo
	IAPPayType     string
	//itemToClient  []string
	//countToClient []uint32
}

func (c *CostData) addVirtualItem(itemid string, count uint32) {
	// http://wiki.taiyouxi.net/w/%E4%B8%89%E5%9B%BD-%E8%AE%BE%E8%AE%A1%E6%96%87%E6%A1%A3%E5%8C%BA/%E6%95%B0%E5%80%BC%E6%96%87%E6%A1%A3/%E8%B5%84%E6%BA%90%E7%B1%BB%E5%9E%8B%E5%AF%B9%E7%85%A7/
	switch itemid {
	case VI_Sc0:
		c.addSc(SC_Money, int64(count))
	case VI_Sc1:
		c.addSc(SC_FineIron, int64(count))
	case VI_BossCoin:
		c.addSc(SC_BossCoin, int64(count))
	case VI_PvpCoin:
		c.addSc(SC_PvpCoin, int64(count))
	case VI_Hc:
		c.addHc(1, int64(count))
	case VI_Hc_Buy:
		c.addHc(0, int64(count))
	case VI_Hc_Give:
		c.addHc(1, int64(count))
	case VI_Hc_Compensate:
		c.addHc(2, int64(count))
	case VI_XP:
		c.addAvatarXp(count)
	case VI_CorpXP:
		c.addCorpXp(count)
	case VI_EN:
		c.Energy += count
	case VI_DC:
		c.addSc(SC_DestinyCoin, int64(count))
	case VI_GoldLevelPoint:
		c.addGoldLevelPoint(count)
	case VI_ExpLevelPoint:
		c.addExpLevelPoint(count)
	case VI_DCLevelPoint:
		c.addDCLevelPoint(count)
	case VI_BossFightPoint:
		c.BossFightPoint += count
	case VI_BossFightRankPoint:
		c.BossFightRankPoint += count
	case VI_HcByVIP:
		c.GiveHcFromVip = 1
	case VI_EC:
		c.addSc(SC_EquipCoin, int64(count))
	case VI_GC:
		c.addSc(SC_GuildCoin, int64(count))
	case VI_GachaTicket:
		c.addSc(SC_GachaTicket, int64(count))
	case VI_StarBlessCoin:
		c.addSc(SC_StarBlessCoin, int64(count))
	case VI_TPVPCoin:
		c.addSc(SC_TPvpCoin, int64(count))
	case VI_EggKey:
		c.addSc(SC_EggKey, int64(count))
	case VI_GuildXP:
		c.GuildXp += count
	case VI_GuildSP:
		c.addSc(SC_GuildSp, int64(count))
	case VI_BaoZi:
		c.addSc(SC_BaoZi, int64(count))
	case VI_GuildBoss:
		c.addSc(SC_GB, int64(count))
	case VI_ACTIVE_ITEM:
		c.addSc(SC_ACTIVE_ITEM, int64(count))
	case VI_Expedition:
		c.addSc(SC_EXPEDITION, int64(count))
	case VI_WeaponCoin:
		c.addSc(SC_WEAPON_COIN, int64(count))
	case VI_HeroDiffPlay:
		c.addSc(SC_HeroDiffPlay, int64(count))
	case VI_WhiteKey:
		c.addSc(SC_WhiteKey, int64(count))
	case VI_WuShuangCoin:
		c.addSc(SC_WuShuangCoin, int64(count))
	case VI_XZ_SD:
		c.addSc(SC_VI_XZ_SD, int64(count))
	case VI_HDP_SD:
		c.addSc(SC_VI_HDP_SD, int64(count))
	case VI_WINE3:
		c.addSc(SC_Wine3, int64(count))
	case VI_WINE4:
		c.addSc(SC_Wine4, int64(count))
	case VI_WINE5:
		c.addSc(SC_Wine5, int64(count))
	case VI_SSXP:
		c.addSc(SC_SSXP, int64(count))
	case VI_SSLIGHT:
		c.addSc(SC_SSLIGHT, int64(count))
	case VI_AC:
		c.addSc(SC_AC, int64(count))
	case VI_WB_BUFFCOIN:
		c.addSc(SC_WB_BUFFCOIN, int64(count))
	case VI_HERO_SURPLUS3:
		c.addSc(SC_HERO_SURPLUS3, int64(count))
	case VI_HERO_SURPLUS4:
		c.addSc(SC_HERO_SURPLUS4, int64(count))
	case VI_HERO_SURPLUS5:
		c.addSc(SC_HERO_SURPLUS5, int64(count))
	case VI_PET_LEVEL:
		c.addSc(SC_PET_LEVEL, int64(count))
	case VI_PET_STAR:
		c.addSc(SC_PET_STAR, int64(count))
	case VI_PET_STAR2:
		c.addSc(SC_PET_STAR2, int64(count))
	case VI_PET_APTITUDE:
		c.addSc(SC_PET_APTITUDE, int64(count))
	case VI_PET_APTITUDE2:
		c.addSc(SC_PET_APTITUDE2, int64(count))
	case VI_PETCOIN:
		c.addSc(SC_PETCOIN, int64(count))
	case VI_GT_GACHA:
		c.addSc(SC_GT_GACHA, int64(count))
	case VI_GT_VIP:
		c.addSc(SC_GT_VIP, int64(count))
	case VI_GT_HGC:
		c.addSc(SC_GT_HGC, int64(count))
	case VI_GT_GWCBOX:
		c.addSc(SC_GT_GWCBOX, int64(count))
	case VI_GT_HEROBOX:
		c.addSc(SC_GT_HEROBOX, int64(count))
	case VI_WheelCoin:
		//c.addSc(SC_WheelCoin , int64(count))
		c.WheelCoin += count
	default:
		logs.Error("Unknown VirtualItem Id By %s", itemid)
	}
}

func (g *CostData) addSc(sc_t int, sc_v int64) {
	if sc_t >= SC_TYPE_COUNT || sc_t < 0 {
		logs.Error("CostData Add Sc Err Typ %d", sc_t)
		return
	}

	g.Sc[sc_t] += sc_v
}

func (g *CostData) addHc(hc_t int, hc_v int64) {
	if hc_t >= HC_TYPE_COUNT || hc_t < 0 {
		logs.Error("CostData Add Sc Err Typ %d", hc_t)
		return
	}

	g.Hc[hc_t] += hc_v
}

func (g *CostData) addCorpXp(xp uint32) {
	g.CorpXp += xp
}

func (g *CostData) addAvatarXp(xp uint32) {
	g.AvatarXp += xp
}

func (g *CostData) addGeneralNum(gid string, gcount uint32) {
	for idx, id := range g.Generals {
		if id == gid {
			g.GCount[idx] += gcount
			return
		}
	}

	g.Generals = append(g.Generals, gid)
	g.GCount = append(g.GCount, gcount)
}

func (g *CostData) addStarSoul(soulID string, num uint32) {
	if nil == g.StarSouls {
		g.StarSouls = map[string]uint32{}
	}
	g.StarSouls[soulID] = g.StarSouls[soulID] + num
}

func (g *CostData) AddItem(item_id string, count uint32) {
	if ok, it := GetPackage(item_id); ok {
		_, pkg := GetPackageGroup(it.GetAttrType())
		for i := int(count); i > 0; i-- {

			if len(pkg.GetStaticItem_Template()) != 0 {
				for _, item := range pkg.GetStaticItem_Template() {
					g.AddItemWithData(item.GetStaticItem(), BagItemData{}, item.GetStaticCount())
				}
			}
			if len(pkg.GetRandomItem_Template()) != 0 {
				for _, tid := range pkg.GetRandomItem_Template() {
					gives, err := GetGivesByTemplate(tid.GetLootTemplate(), "", nil)
					if err != nil {
						continue
					}
					if gives.IsNotEmpty() {
						for idx, itemID := range gives.Item2Client {
							g.AddItemWithData(itemID, BagItemData{}, gives.Count2Client[idx])
						}
					}

				}
			}
		}

	} else {
		g.AddItemWithData(item_id, BagItemData{}, count)
	}
}

// don't use
func (g *CostData) AddItemWithRes(item_id string, count uint32, res *CostData2Client, r *rand.Rand) {
	if ok, it := GetPackage(item_id); ok {
		_, pkg := GetPackageGroup(it.GetAttrType())
		for i := int(count); i > 0; i-- {

			if len(pkg.GetStaticItem_Template()) != 0 {
				for _, item := range pkg.GetStaticItem_Template() {
					g.AddItemWithData(item.GetStaticItem(), BagItemData{}, item.GetStaticCount())
					if !IsItemIdKnownBeforeGive(item.GetStaticItem()) {
						res.AddItemWithData2Client(item.GetStaticItem(), BagItemData{}, item.GetStaticCount())
					}
				}
			}
			if len(pkg.GetRandomItem_Template()) != 0 {
				for _, tid := range pkg.GetRandomItem_Template() {
					gives, err := GetGivesByTemplate(tid.GetLootTemplate(), "", r)
					if err != nil {
						continue
					}
					if gives.IsNotEmpty() {
						for idx, itemID := range gives.Item2Client {
							g.AddItemWithData(itemID, BagItemData{}, gives.Count2Client[idx])
							if !IsItemIdKnownBeforeGive(itemID) {
								res.AddItemWithData2Client(itemID, BagItemData{}, gives.Count2Client[idx])
							}
						}
					}

				}
			}
		}

	} else {
		g.AddItemWithData(item_id, BagItemData{}, count)
		if !IsItemIdKnownBeforeGive(item_id) {
			res.AddItemWithData2Client(item_id, BagItemData{}, count)
		}
	}
}

func (g *CostData) AddItemWithData(item_id string, data BagItemData, count uint32) {
	g.AddItemWithData2Client(item_id, data, count)
	g.addItemWithData(item_id, data, count)
}

func (g *CostData) addItemWithData(item_id string, data BagItemData, count uint32) {
	avatarID, ok := gdPlayerHeroPieceID2IDx[item_id]
	if ok {
		g.HeroPiece[avatarID] += count
		return
	}

	if IsItemVirtual(item_id) {
		g.addVirtualItem(item_id, count)
		return
	}

	if CheckItemIsStarSoul(item_id) {
		g.addStarSoul(item_id, count)
	}

	// General GoodWill
	// 要在IsItemToSCWhenAdd之前判断
	isGeneral, generalID, _ := IsGeneralGoodwillItem(item_id)
	if isGeneral && generalID != "" {
		g.addGeneralNum(generalID, count)
		return
	}

	is_make, vt, vcount := IsItemToSCWhenAdd(item_id)
	if is_make {
		g.addVirtualItem(vt, vcount*count)
		return
	}

	if IsItemToBuffWhenAdd(item_id) {
		// BuffItem 不作处理
		return
	}

	if IsGeneral(item_id) {
		g.addGeneralNum(item_id, count)
		return
	}

	is_wholeChar, heroTyp, heroC, _ := IsItemToWholeCharWhenAdd(item_id)
	if is_wholeChar {
		heroIdx, ok := gdPlayerHeroID2IDx[heroTyp]
		if !ok && heroIdx >= AVATAR_NUM_MAX {
			logs.Error("heroTyp Nil By %s", heroTyp)
		}
		g.HeroPieceWholeChar[heroIdx] += count
		g.HeroPieceWholeCharCount[heroIdx] = heroC
		return
	}

	for idx, id := range g.Items {
		if id == item_id {
			g.Count[idx] += count
			return
		}
	}

	g.Items = append(g.Items, item_id)
	g.Count = append(g.Count, count)
	g.ItemData = append(g.ItemData, data)

	if IsItemEquip(item_id) {
		g.HasEquip = true
	}
	if ok, _ := IsJade(item_id); ok {
		g.HasJade = true
	}
}

func (g *CostData) AddGroup(other *CostData) {
	for sc_t, sc_v := range other.Sc {
		g.addSc(sc_t, sc_v)
	}

	for idx, item_id := range other.Items {
		g.addItemWithData(item_id, other.ItemData[idx], uint32(other.Count[idx]))
	}

	g.HasEquip = g.HasEquip || other.HasEquip
	g.HasJade = g.HasJade || other.HasJade

	for i := 0; i < len(other.Generals); i++ {
		g.addGeneralNum(other.Generals[i], other.GCount[i])
	}

	g.GoldLevelPoint += other.GoldLevelPoint
	g.ExpLevelPoint += other.ExpLevelPoint
	g.DCLevelPoint += other.DCLevelPoint

	for hc_t, hc_v := range other.Hc {
		g.addHc(hc_t, hc_v)
	}

	for idx, v := range other.HeroPiece {
		g.HeroPiece[idx] += v
	}

	g.AvatarXp += other.AvatarXp
	g.CorpXp += other.CorpXp
	g.Energy += other.Energy
	g.WheelCoin += other.WheelCoin
	g.BossFightPoint += other.BossFightPoint
	g.BossFightRankPoint += other.BossFightRankPoint

	for idx, v := range other.HeroPieceWholeChar {
		g.HeroPieceWholeChar[idx] += v
	}

	for idx, v := range other.HeroPieceWholeCharCount {
		g.HeroPieceWholeCharCount[idx] = v
	}

	if nil != other.StarSouls {
		for id, c := range other.StarSouls {
			g.addStarSoul(id, c)
		}
	}

	g.AddOther2Client(other)
}

func (g *CostData) addGoldLevelPoint(glp uint32) {
	g.GoldLevelPoint += glp
}

func (g *CostData) addExpLevelPoint(elp uint32) {
	g.ExpLevelPoint += elp
}

func (g *CostData) addDCLevelPoint(dclp uint32) {
	g.DCLevelPoint += dclp
}

func (g *CostData) AddEStageTimes(estageId string, times uint32) {
	g.EStageId = estageId
	g.EStageTimes = times
}

func (g *CostData) AddHeroTalentPoint(p uint32) {
	g.HeroTalentPoint = p
}

func (g *CostData) AddGameModeTimes(gameModeId int, times int) {
	g.GameModeId = gameModeId
	g.GameModeTimes = times
}

func (g *CostData) AddIAPGood(iapGoodIndex uint32, gameOrderId, order string, money uint32,
	IAPChannel, IAPPayTime, platformLogicTyp string, amount uint32, pkginfo PackageInfo, payType string) {
	g.IAPGoodIndex = iapGoodIndex
	g.IAPGameOrderId = gameOrderId
	g.IAPGoodOrder = order
	g.IAPMoney = money
	g.IAPChannel = IAPChannel
	g.IAPPayTime = IAPPayTime
	g.IAPTrueAmount = amount
	g.IAPPlatform = platformLogicTyp
	g.IAPPkgInfo = pkginfo
	g.IAPPayType = payType
}

func (g *CostData) AddIAPGoodByID(iapGoodId string, order string, money uint32, IAPPayTime string) {
	idx := GetIAPIdxByID(iapGoodId)
	if idx == 0 {
		logs.Error("AddIAPGoodByID Err By No %s", iapGoodId)
	} else {
		g.IAPGoodIndex = idx
		g.IAPGoodOrder = order
		g.IAPMoney = money
		g.IAPPayTime = IAPPayTime
		g.IAPPlatform = uutil.IOS_Platform
	}
}

// 这个求差的适合累计奖励返还或者消耗时用 为了简便只计算Sc!!!!!!
// return g - other --> 这里假定g的所有奖励都要大于other
func (g *CostData) SubSCGroup(other *CostData) *CostData {
	re := &CostData{}
	for sc_t, sc_v := range other.Sc {
		re.Sc[sc_t] = g.Sc[sc_t] - sc_v
	}
	return re
}

func (g *CostData) AddAvatar(avatar int) {
	g.AvatarIds = append(g.AvatarIds, avatar)
}
func (g *CostData) AddAvatars(avatars []int) {
	g.AvatarIds = append(g.AvatarIds, avatars...)
}

func (g *CostData) SetGST(gst_t GST_Typ) {
	g.GuildScienceValid = gst_t
}

func (g *CostData) SetGateEnemyBonue(bonus float32) {
	g.GateEnemyBonus = bonus
}
