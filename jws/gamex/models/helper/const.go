package helper

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const CurrDBVersion int64 = 8

const AUTO_START_ID = uint32(0x8000000)

// 软通货种类, 未来根据需求添加新的
// 增加时查找所有helper.SC_Money gamedata.SC_Money 在相关地方添加特定逻辑
const SC_Money = 0
const SC_FineIron = 1    // 精铁
const SC_DestinyCoin = 2 // 扫荡券
const SC_BossCoin = 3
const SC_PvpCoin = 4
const SC_EquipCoin = 5
const SC_GuildCoin = 6
const SC_GachaTicket = 7
const SC_StarBlessCoin = 8
const SC_TPvpCoin = 9
const SC_EggKey = 10
const SC_BaoZi = 11
const SC_ACTIVE_ITEM = 12
const SC_EXPEDITION = 13
const SC_WEAPON_COIN = 14  // 神兵碎片
const SC_GB = 15           // 军魂
const SC_GuildSp = 16      // 建设点
const SC_HeroDiffPlay = 17 //出奇制胜
const SC_WuShuangCoin = 18
const SC_WhiteKey = 19 //白盒宝箱抽奖券
const SC_VI_XZ_SD = 20
const SC_VI_HDP_SD = 21
const SC_Wine3 = 22
const SC_Wine4 = 23
const SC_Wine5 = 24
const SC_SSXP = 25
const SC_SSLIGHT = 26
const SC_AC = 27
const SC_WB_BUFFCOIN = 28
const SC_HERO_SURPLUS3 = 29
const SC_HERO_SURPLUS4 = 30
const SC_HERO_SURPLUS5 = 31
const SC_PET_LEVEL = 32
const SC_PET_STAR = 33
const SC_PET_STAR2 = 34
const SC_PET_APTITUDE = 35
const SC_PET_APTITUDE2 = 36
const SC_PETCOIN = 37
const SC_GT_GACHA = 38   //   高级宝箱抽奖券
const SC_GT_VIP = 39     //   定向宝箱抽奖券
const SC_GT_HGC = 40     //   限时神将抽奖券
const SC_GT_GWCBOX = 41  //   神器再临抽奖券
const SC_GT_HEROBOX = 42 //   主将巡礼抽奖券
const SC_TYPE_COUNT = 43

const (
	FashionPart_Weapon = iota
	FashionPart_Armor
	FashionPart_Count
)

func SCString(scType int) string {
	switch scType {
	case SC_Money:
		return "VI_SC"
	case SC_FineIron:
		return "VI_FI"
	case SC_DestinyCoin:
		return "VI_DC"
	case SC_BossCoin:
		return "VI_BC"
	case SC_PvpCoin:
		return "VI_PVPC"
	case SC_EquipCoin:
		return "VI_EC"
	case SC_GuildCoin:
		return "VI_GC"
	case SC_GachaTicket:
		return "VI_GT"
	case SC_StarBlessCoin:
		return "VI_SB"
	case SC_TPvpCoin:
		return "VI_TPVPC"
	case SC_EggKey:
		return "VI_EggKey"
	case SC_BaoZi:
		return "VI_BAOZI"
	case SC_ACTIVE_ITEM:
		return "VI_ACTIVE_ITEM"
	case SC_EXPEDITION:
		return "VI_Expedition"
	case SC_WEAPON_COIN:
		return "VI_GWC"
	case SC_GB:
		return "VI_GB"
	case SC_GuildSp:
		return "VI_GuildSP"
	case SC_HeroDiffPlay:
		return "VI_HeroDiffPlay"
	case SC_WhiteKey:
		return "VI_WK"
	case SC_WuShuangCoin:
		return "VI_WSC"
	case SC_VI_XZ_SD:
		return "VI_XZ_SD"
	case SC_VI_HDP_SD:
		return "VI_HDP_SD"
	case SC_Wine3:
		return "VI_WINE3"
	case SC_Wine4:
		return "VI_WINE4"
	case SC_Wine5:
		return "VI_WINE5"
	case SC_SSXP:
		return "VI_SSXP"
	case SC_SSLIGHT:
		return "VI_SSLIGHT"
	case SC_AC:
		return "VI_AC"
	case SC_WB_BUFFCOIN:
		return "VI_WB_BUFFCOIN"
	case SC_HERO_SURPLUS3:
		return "VI_HERO_SURPLUS3"
	case SC_HERO_SURPLUS4:
		return "VI_HERO_SURPLUS4"
	case SC_HERO_SURPLUS5:
		return "VI_HERO_SURPLUS5"
	case SC_PET_LEVEL:
		return "VI_PET_LEVEL"
	case SC_PET_STAR:
		return "VI_PET_STAR"
	case SC_PET_STAR2:
		return "VI_PET_STAR2"
	case SC_PET_APTITUDE:
		return "VI_PET_APTITUDE"
	case SC_PET_APTITUDE2:
		return "VI_PET_APTITUDE2"
	case SC_PETCOIN:
		return "VI_PETCOIN"
	case SC_GT_GACHA:
		return "VI_GT_GACHA"
	case SC_GT_VIP:
		return "VI_GT_VIP"
	case SC_GT_HGC:
		return "VI_GT_HGC"
	case SC_GT_GWCBOX:
		return "VI_GT_GWCBOX"
	case SC_GT_HEROBOX:
		return "VI_GT_HEROBOX"
	default:
		logs.Error("SCString not define %d", scType)
		return ""
	}
}

func SCId(scType string) int {
	switch scType {
	case "VI_SC":
		return SC_Money
	case "VI_FI":
		return SC_FineIron
	case "VI_DC":
		return SC_DestinyCoin
	case "VI_BC":
		return SC_BossCoin
	case "VI_PVPC":
		return SC_PvpCoin
	case "VI_EC":
		return SC_EquipCoin
	case "VI_GC":
		return SC_GuildCoin
	case "VI_GT":
		return SC_GachaTicket
	case "VI_SB":
		return SC_StarBlessCoin
	case "VI_TPVPC":
		return SC_TPvpCoin
	case "VI_EggKey":
		return SC_EggKey
	case "VI_BAOZI":
		return SC_BaoZi
	case "VI_ACTIVE_ITEM":
		return SC_ACTIVE_ITEM
	case "VI_Expedition":
		return SC_EXPEDITION
	case "VI_GWC":
		return SC_WEAPON_COIN
	case "VI_HeroDiffPlay":
		return SC_HeroDiffPlay
	case "VI_WK":
		return SC_WhiteKey
	case "VI_WSC":
		return SC_WuShuangCoin
	case "VI_XZ_SD":
		return SC_VI_XZ_SD
	case "VI_HDP_SD":
		return SC_VI_HDP_SD
	case "VI_WINE3":
		return SC_Wine3
	case "VI_WINE4":
		return SC_Wine4
	case "VI_WINE5":
		return SC_Wine5
	case "VI_SSXP":
		return SC_SSXP
	case "VI_SSLIGHT":
		return SC_SSLIGHT
	case "VI_AC":
		return SC_AC
	case "VI_WB_BUFFCOIN":
		return SC_WB_BUFFCOIN
	case "VI_HERO_SURPLUS3":
		return SC_HERO_SURPLUS3
	case "VI_HERO_SURPLUS4":
		return SC_HERO_SURPLUS4
	case "VI_HERO_SURPLUS5":
		return SC_HERO_SURPLUS5
	case "VI_PET_LEVEL":
		return SC_PET_LEVEL
	case "VI_PET_STAR":
		return SC_PET_STAR
	case "VI_PET_STAR2":
		return SC_PET_STAR2
	case "VI_PET_APTITUDE":
		return SC_PET_APTITUDE
	case "VI_PET_APTITUDE2":
		return SC_PET_APTITUDE2
	case "VI_PETCOIN":
		return SC_PETCOIN
	case "VI_GT_GACHA":
		return SC_GT_GACHA
	case "VI_GT_VIP":
		return SC_GT_VIP
	case "VI_GT_HGC":
		return SC_GT_HGC
	case "VI_GT_GWCBOX":
		return SC_GT_GWCBOX
	case "VI_GT_HEROBOX":
		return SC_GT_HEROBOX
	default:
		return SC_TYPE_COUNT
	}
}

// 硬通货种类
const HC_From_Buy = 0        // 购买钻 --|
const HC_From_Give = 1       // 赠送钻 --|--顺序是固定的
const HC_From_Compensate = 2 // 补偿钻 --|
const HC_TYPE_COUNT = 3

func HCString(hcType int) string {
	switch hcType {
	case HC_From_Buy:
		return "VI_HC_Buy"
	case HC_From_Give:
		return "VI_HC_Give"
	case HC_From_Compensate:
		return "VI_HC_Compensate"
	case HC_TYPE_COUNT:
		return "VI_HC"
	default:
		logs.Error("HCString not define %d", hcType)
		return ""
	}
}

// 警告
// 这个常量表示一个角色身上可装备栏位的最大值
// 由于这个值一般是固定的，所以存储时使用连续数组存储
// 一旦修改了这个值，会导致存在数据库中的数据串位而不可用
// 这时需要增加逻辑读取升级老玩家数据
// TODO by FanYang 添加DBLoad功能已支出自动升级玩家数据 TBDYZH
const EQUIP_SLOT_MAX = 10
const EQUIP_SLOT_CURR = PartEquipCount
const AVATAR_EQUIP_SLOT_MAX = 5

const JADE_SLOT_MAX = 10

const BATTLE_ARMY_NUM_MAX = 4
const BATTLE_ARMYLOC_NUM_MAX = 7

//最大灵宠数量，武将们最多拥有的灵宠个数
const PET_NUM_MAX = 5

// 角色最大数量 这个主要是用来限制装备信息增长用的
const AVATAR_NUM_MAX = 40

// 当前角色数量
const AVATAR_NUM_CURR = 33
const ALL_AVATAR_EQUIP_SLOT_MAX = AVATAR_EQUIP_SLOT_MAX * AVATAR_NUM_MAX

const AVATAR_SKILL_MAX = 10
const CORP_SKILLPRACTICE_MAX = 4

// 装备材料强化最多材料位
const EQUIP_MAT_ENHANCE_MAT = 6

// 一个要消耗/赠与东西的列表
// 这个主要是配合逻辑中得CostGroup和GiveGroup使用
// 虽然名字是CostData 但也可以根据这个给玩家赠送东西
//

const (
	VI_Sc0                = "VI_SC"
	VI_Sc1                = "VI_FI"
	VI_Hc_Buy             = "VI_HC_Buy"
	VI_Hc_Give            = "VI_HC_Give"
	VI_Hc_Compensate      = "VI_HC_Compensate"
	VI_Hc                 = "VI_HC"
	VI_XP                 = "VI_XP"
	VI_CorpXP             = "VI_CorpXP"
	VI_EN                 = "VI_EN" // 体力
	VI_GoldLevelPoint     = "VI_GLP"
	VI_ExpLevelPoint      = "VI_ELP"
	VI_DCLevelPoint       = "VI_DCLP"
	VI_BossFightPoint     = "VI_BFP"
	VI_BossFightRankPoint = "VI_BFRankP"
	VI_HcByVIP            = "VI_HcByVIP"
	VI_BossCoin           = "VI_BC"
	VI_PvpCoin            = "VI_PVPC"
	VI_EC                 = "VI_EC" // SC_EquipCoin 装备代币
	VI_DC                 = "VI_DC" // SC_DestinyCoin 神将代币
	VI_GC                 = "VI_GC" // SC_GuildCoin 公会代币
	VI_GachaTicket        = "VI_GT" // SC_GachaTicket Gacha抽奖券
	VI_StarBlessCoin      = "VI_SB"
	VI_TPVPCoin           = "VI_TPVPC" // team pvp 代币
	VI_EggKey             = "VI_EK"
	VI_GuildXP            = "VI_GuildXP"
	VI_GuildSP            = "VI_GuildSP"
	VI_BaoZi              = "VI_BAOZI"
	VI_GuildBoss          = "VI_GB"
	VI_ACTIVE_ITEM        = "VI_ACTIVE_ITEM"
	VI_Expedition         = "VI_Expedition"
	VI_WeaponCoin         = "VI_GWC"          //远征代币
	VI_HeroDiffPlay       = "VI_HeroDiffPlay" //出奇制胜代币
	VI_WhiteKey           = "VI_WK"           // 白盒宝箱抽奖券
	VI_WuShuangCoin       = "VI_WSC"          //无双代币
	VI_XZ_SD              = "VI_XZ_SD"
	VI_HDP_SD             = "VI_HDP_SD"
	VI_WINE3              = "VI_WINE3"
	VI_WINE4              = "VI_WINE4"
	VI_WINE5              = "VI_WINE5"
	VI_SSXP               = "VI_SSXP"          // 星魂分解材料(星魂经验):星图系统
	VI_SSLIGHT            = "VI_SSLIGHT"       // 七星灯:星图系统
	VI_AC                 = "VI_AC"            // 香火:星图系统
	VI_WB_BUFFCOIN        = "VI_WB_BUFFCOIN"   // 世界boss上古之魂
	VI_HERO_SURPLUS3      = "VI_HERO_SURPLUS3" // 普通举贤令
	VI_HERO_SURPLUS4      = "VI_HERO_SURPLUS4" // 精锐举贤令
	VI_HERO_SURPLUS5      = "VI_HERO_SURPLUS5" // 无双举贤令
	VI_PET_LEVEL          = "VI_PET_LEVEL"     //灵石
	VI_PET_STAR           = "VI_PET_STAR"      //星石
	VI_PET_STAR2          = "VI_PET_STAR2"     //保星符
	VI_PET_APTITUDE       = "VI_PET_APTITUDE"  //洗髓丹
	VI_PET_APTITUDE2      = "VI_PET_APTITUDE2" //高级洗髓符
	VI_PETCOIN            = "VI_PETCOIN"       //灵宠代币
	VI_GT_GACHA           = "VI_GT_GACHA"      //高级宝箱抽奖券
	VI_GT_VIP             = "VI_GT_VIP"        //定向宝箱抽奖券
	VI_GT_HGC             = "VI_GT_HGC"        //限时神将抽奖券
	VI_GT_GWCBOX          = "VI_GT_GWCBOX"     //主将巡礼抽奖券
	VI_GT_HEROBOX         = "VI_GT_HEROBOX"    //神兵巡礼抽奖券
	VI_WheelCoin          = "VI_LWC"           //幸运转盘抽奖币
)

const (
	TCJ_Nil = iota
	TCJ_Ten // 天
	TCJ_Chi // 地
	TCJ_Jin // 人
)

// 购买类型, 在Buy中使用
const (
	Buy_Typ_EnergyBuy = iota
	Buy_Typ_BossFightPoint
	Buy_Typ_SC
	Buy_Typ_EliteTimes
	Buy_Typ_TeamPvp
	Buy_Typ_SimplePvp
	Buy_Typ_GuildBossCount
	Buy_Typ_GuildBigBossCount
	Buy_Typ_HeroTalentPoint
	Buy_Typ_BaoZi
	Buy_Typ_FestivalBossCount // 10
	Buy_Typ_WSPVP_Refresh
	Buy_Typ_WSPVP_Challenge
	Buy_Typ_Count
)

// Boss品质, 在Boss遭遇战中使用
const (
	Boss_Class_EnergyBuy = iota
	Boss_Class_Green
	Boss_Class_Blue
	Boss_Class_Purple
	Boss_Class_Golden
	Boss_Class_Count
)

// 任务类型
const (
	Quest_Main = iota
	Quest_Branch
	Quest_PVE_Boss
	Quest_Story_No_Use_Now
	Quest_Daily
	Quest_QuestPoint
	Quest_7Day
	Quest_QuestPoint_7Day
	Quest_Typ_count
)

// 使用物品的类型
const (
	UseItem_Exp = "XPPotion"
)

// 商店类型
const (
	Store_Town = 0
	Store_Boss = 1
	Store_Pvp  = 2
)

// pveboss刷新时间
const Pve_Boss_Refresh_Time = "22:00"

// 每日任务跨天时间
const Daily_Quest_Refresh_Time = "00:00"

// 出战神将数量
const DestinyGeneralSkillMax = 3

// GVG出战数量
const GVG_AVATAR_COUNT = 3

func StoreString(storeId uint32) string {
	switch storeId {
	case 0:
		return "StoreGuild"
	case 1:
		return "StoreBoss"
	case 2:
		return "StorePvp"
	case 3:
		return "StoreEquip"
	case 4:
		return ""
	case 5:
		return "StoreTPvp"
	case 6:
		return "StoreVip"
	default:
		return fmt.Sprintf("Store%d", storeId)
	}
}

func ShopString(shopId uint32) string {
	switch shopId {
	case 0:
		return "Shop"
	case 1:
		return "ShopJade"
	case 2:
		return "ShopVIP"
	case 3:
		return "ShopGuild"
	default:
		return fmt.Sprintf("Shop%d", shopId)
	}
}

const (
	Gacha_HC = iota
	Gacha_Normal
	Gacha_General
	Gacha_
	Gacha_Vip
)

func GachaTypeString(gachaType int, isTen bool) string {
	switch gachaType {
	case Gacha_HC:
		if isTen {
			return "HCTen"
		}
		return "HCOne"
	case Gacha_Normal:
		if isTen {
			return "NormalTen"
		}
		return "NormalOne"
	case Gacha_General:
		if isTen {
			return "GeneralTen"
		}
		return "GeneralOne"
	case Gacha_:
		return "_Ten"
	case Gacha_Vip:
		if isTen {
			return "HCVipTen"
		}
		return "HCVipOne"
	}
	return fmt.Sprintf("gachaType %d isTen %v", gachaType, isTen)
}

func BuyTypeString(buyType int) string {
	switch buyType {
	case Buy_Typ_EnergyBuy:
		return "BuyEnergy"
	case Buy_Typ_BossFightPoint:
		return "BuyBFP"
	case Buy_Typ_SC:
		return "BuySC"
	case Buy_Typ_EliteTimes:
		return "BuyEliteTimes"
	case Buy_Typ_TeamPvp:
		return "BuyTeamPvpTimes"
	case Buy_Typ_SimplePvp:
		return "BuySimplePvp"
	case Buy_Typ_GuildBossCount:
		return "BuyGuildBossCount"
	case Buy_Typ_GuildBigBossCount:
		return "BuyGuildBigBossCount"
	case Buy_Typ_HeroTalentPoint:
		return "BuyHeroTalentPoint"
	case Buy_Typ_BaoZi:
		return "BuyBaoZi"
	case Buy_Typ_FestivalBossCount:
		return "BuyFestivalBossCount"
	case Buy_Typ_WSPVP_Refresh:
		return "BuyWSPVPRefresh"
	case Buy_Typ_WSPVP_Challenge:
		return "BuyWSPVPChallenge"
	}
	return fmt.Sprintf("BuyType-%d", buyType)
}

const (
	MaxGuildMember       = 50
	MaxRandGuilds        = 10
	MaxGuildScienceCount = 10
)

// 服务器物品内部逻辑分类
const (
	Item_Inner_Type_Basic = iota
	Item_Inner_Type_Jade
	Item_Inner_Type_Fashion
)

// itemadd返回码
const (
	RES_AddToBag_Success = iota
	RES_AddToBag_Err
	RES_AddToBag_NoItemADD
	RES_AddToBag_MaxCount // 超过物品最大拥有数量
)

const (
	PartID_Weapon = iota
	PartID_Chest
	PartID_Necklace
	PartID_Belt
	PartID_Ring
	PartID_Leggings
	PartID_Bracers
	PartEquipCount
)

const (
	GuildMedal = iota + PartEquipCount
	GuildMedalCount
)

const (
	JadePart_0 = iota
	JadePart_1
	JadePart_2
	JadePart_3
	JadePart_4
	JadePart_5
	JadePartCount
)

var (
	// part字符串到idx得索引
	part2SlotIdxMap = map[string]int{
		"Weapon":          PartID_Weapon,
		"Chest":           PartID_Chest,
		"Necklace":        PartID_Necklace,
		"Belt":            PartID_Belt,
		"Ring":            PartID_Ring,
		"Leggings":        PartID_Leggings,
		"Bracers":         PartID_Bracers,
		"GuildMedal":      GuildMedal,
		"FWeapon":         FashionPart_Weapon,
		"FAmor":           FashionPart_Armor,
		"EvoMaterials":    10,
		"ArouMaterials":   11,
		"Treasurebox":     12,
		"Baozi":           13,
		"Defence":         14,
		"Attack":          15,
		"HardCoin":        16,
		"SoftCoin":        17,
		"XP":              18,
		"CorpXP":          19,
		"FineIron":        20,
		"Energy":          21,
		"SweepTicket":     22,
		"RandomItem":      23,
		"EquipChips":      24,
		"GoldLevelPoint":  25,
		"BossFightPoint":  26,
		"XPPotion":        27,
		"ExpLevelPoint":   28,
		"StarMaterials":   29,
		"BossCoin":        30,
		"PVPCoin":         31,
		"GeneralItem":     32,
		"GeneralGoodwill": 33,

		"None": -1,
		"":     -1,
	}

	part2AvatarSlotIdxMap = map[string]int{
		"FWeapon": FashionPart_Weapon,
		"FAmor":   FashionPart_Armor,
	}

	part2JadeSlotIdxMap = map[string]int{
		"JD0": JadePart_0,
		"JD1": JadePart_1,
		"JD2": JadePart_2,
		"JD3": JadePart_3,
		"JD4": JadePart_4,
		"JD5": JadePart_5,
	}
)

// 龙玉装备part
func GetJadeSlot(part string) int {
	slot, ok := part2JadeSlotIdxMap[part]
	if !ok {
		return -1
	}
	return slot
}

const equip_slot_max = GuildMedal // part2SlotIdxMap里装备slot的最大值；目前装备slot分为了两部分，这个值是两部分的最大值
// 获取一个Part代表的装备位置索引，如果不是装备，返回-1
func GetEquipSlot(part string) int {
	slot, ok := part2SlotIdxMap[part]
	if !ok || slot > equip_slot_max {
		return -1
	}
	return slot
}

// 获取avatar专属装备Part
func GetAvatarEquipSlot(part string) int {
	slot, ok := part2AvatarSlotIdxMap[part]
	if !ok {
		return -1
	}
	return slot
}

// 自适应掉落标记类型 遇到掉这个ItemID时转为自适应掉落
const MatEevoUniversalItemID = "MAT_EEVO_UNIVERSAL"

const TeamPvpAvatarsCount = 3

// 神兽最多多少个
const MaxDestingGeneralCount = 10 // 超过则需要修改这个数

// 天赋最多多少个
const MaxTalentCount = 6
const CurTalentCount = 4

const GateEnemyBuffCount = 4 // 一共3个buff，0不用，所以是4个
const GateEnemyBuffMaxLv = 3

const CorpHeroGsNum = 3 // 战队计算拿几个英雄的gs
const WspvpBestHeroCount = 9

const (
	HeroDiff_TU = iota
	HeroDiff_ZHAN
	HeroDiff_HU
	HeroDiff_SHI
	HeroDiff_Count
)

// 势力
const (
	Country_Invalid = iota
	Country_Shu
	Country_Wei
	Country_Wu
	Country_Qun
	Country_Count
)

// 武将剩余碎片抽奖
const (
	Hero_Surplus_3 = iota
	Hero_Surplus_4
	Hero_Surplus_5
	Hero_Surplus_Count
)
