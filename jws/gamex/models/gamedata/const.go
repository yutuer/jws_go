package gamedata

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	Minute2Second = 60
	Hour2Second   = 60 * Minute2Second
	Day2Second    = 24 * Hour2Second
)

const (
	SC_Money         = helper.SC_Money
	SC_FineIron      = helper.SC_FineIron
	SC_DestinyCoin   = helper.SC_DestinyCoin
	SC_BossCoin      = helper.SC_BossCoin
	SC_PvpCoin       = helper.SC_PvpCoin
	SC_EquipCoin     = helper.SC_EquipCoin
	SC_GuildCoin     = helper.SC_GuildCoin
	SC_GachaTicket   = helper.SC_GachaTicket
	SC_StarBlessCoin = helper.SC_StarBlessCoin
	SC_TPvpCoin      = helper.SC_TPvpCoin
	SC_EggKey        = helper.SC_EggKey
	SC_BaoZi         = helper.SC_BaoZi
	SC_ACTIVE_ITEM   = helper.SC_ACTIVE_ITEM
	SC_EXPEDITION    = helper.SC_EXPEDITION
	SC_WEAPON_COIN   = helper.SC_WEAPON_COIN
	SC_GB            = helper.SC_GB
	SC_GuildSp       = helper.SC_GuildSp
	SC_HeroDiffPlay  = helper.SC_HeroDiffPlay
	SC_WhiteKey      = helper.SC_WhiteKey
	SC_WuShuangCoin  = helper.SC_WuShuangCoin
	SC_VI_XZ_SD      = helper.SC_VI_XZ_SD
	SC_VI_HDP_SD     = helper.SC_VI_HDP_SD
	SC_Wine3         = helper.SC_Wine3
	SC_Wine4         = helper.SC_Wine4
	SC_Wine5         = helper.SC_Wine5
	SC_SSXP          = helper.SC_SSXP
	SC_SSLIGHT       = helper.SC_SSLIGHT
	SC_AC            = helper.SC_AC
	SC_WB_BUFFCOIN   = helper.SC_WB_BUFFCOIN
	SC_HERO_SURPLUS3 = helper.SC_HERO_SURPLUS3
	SC_HERO_SURPLUS4 = helper.SC_HERO_SURPLUS4
	SC_HERO_SURPLUS5 = helper.SC_HERO_SURPLUS5
	SC_PET_LEVEL     = helper.SC_PET_LEVEL
	SC_PET_STAR      = helper.SC_PET_STAR
	SC_PET_STAR2     = helper.SC_PET_STAR2
	SC_PET_APTITUDE  = helper.SC_PET_APTITUDE
	SC_PET_APTITUDE2 = helper.SC_PET_APTITUDE2
	SC_PETCOIN       = helper.SC_PETCOIN
	SC_GT_GACHA      = helper.SC_GT_GACHA
	SC_GT_VIP        = helper.SC_GT_VIP
	SC_GT_HGC        = helper.SC_GT_HGC
	SC_GT_GWCBOX     = helper.SC_GT_GWCBOX
	SC_GT_HEROBOX    = helper.SC_GT_HEROBOX
	SC_TYPE_COUNT    = helper.SC_TYPE_COUNT
)

const (
	HC_From_Buy        = helper.HC_From_Buy
	HC_From_Give       = helper.HC_From_Give
	HC_From_Compensate = helper.HC_From_Compensate
	HC_TYPE_COUNT      = helper.HC_TYPE_COUNT
)

const (
	AVATAR_NUM_MAX   = helper.AVATAR_NUM_MAX
	AVATAR_NUM_CURR  = helper.AVATAR_NUM_CURR
	AVATAR_SLOT_MAX  = helper.EQUIP_SLOT_MAX
	AVATAR_SKILL_MAX = helper.AVATAR_SKILL_MAX
)

const CORP_SKILLPRACTICE_MAX = helper.CORP_SKILLPRACTICE_MAX

const EQUIP_SLOT_MAX = helper.EQUIP_SLOT_MAX

const TCJ_COUNT = helper.TCJ_Jin

// 一个要消耗/赠与东西的列表
// 这个主要是配合逻辑中得CostGroup和GiveGroup使用
// 虽然名字是CostData 但也可以根据这个给玩家赠送东西
//

const (
	VI_Sc0                = helper.VI_Sc0
	VI_Sc1                = helper.VI_Sc1
	VI_Hc_Buy             = helper.VI_Hc_Buy
	VI_Hc_Give            = helper.VI_Hc_Give
	VI_Hc_Compensate      = helper.VI_Hc_Compensate
	VI_Hc                 = helper.VI_Hc
	VI_XP                 = helper.VI_XP
	VI_CorpXP             = helper.VI_CorpXP
	VI_EN                 = helper.VI_EN
	VI_GoldLevelPoint     = helper.VI_GoldLevelPoint
	VI_ExpLevelPoint      = helper.VI_ExpLevelPoint
	VI_DCLevelPoint       = helper.VI_DCLevelPoint
	VI_BossFightPoint     = helper.VI_BossFightPoint
	VI_BossFightRankPoint = helper.VI_BossFightRankPoint
	VI_HcByVIP            = helper.VI_HcByVIP
	VI_BossCoin           = helper.VI_BossCoin
	VI_PvpCoin            = helper.VI_PvpCoin
	VI_EC                 = helper.VI_EC
	VI_DC                 = helper.VI_DC
	VI_GC                 = helper.VI_GC
	VI_GachaTicket        = helper.VI_GachaTicket
	VI_StarBlessCoin      = helper.VI_StarBlessCoin
	VI_TPVPCoin           = helper.VI_TPVPCoin
	VI_EggKey             = helper.VI_EggKey
	VI_GuildXP            = helper.VI_GuildXP
	VI_GuildSP            = helper.VI_GuildSP
	VI_BaoZi              = helper.VI_BaoZi
	VI_GuildBoss          = helper.VI_GuildBoss
	VI_ACTIVE_ITEM        = helper.VI_ACTIVE_ITEM
	VI_Expedition         = helper.VI_Expedition
	VI_WeaponCoin         = helper.VI_WeaponCoin
	VI_HeroDiffPlay       = helper.VI_HeroDiffPlay
	VI_WhiteKey           = helper.VI_WhiteKey
	VI_WuShuangCoin       = helper.VI_WuShuangCoin
	VI_XZ_SD              = helper.VI_XZ_SD
	VI_HDP_SD             = helper.VI_HDP_SD
	VI_WINE3              = helper.VI_WINE3
	VI_WINE4              = helper.VI_WINE4
	VI_WINE5              = helper.VI_WINE5
	VI_SSXP               = helper.VI_SSXP
	VI_SSLIGHT            = helper.VI_SSLIGHT
	VI_AC                 = helper.VI_AC
	VI_WB_BUFFCOIN        = helper.VI_WB_BUFFCOIN
	VI_HERO_SURPLUS3      = helper.VI_HERO_SURPLUS3
	VI_HERO_SURPLUS4      = helper.VI_HERO_SURPLUS4
	VI_HERO_SURPLUS5      = helper.VI_HERO_SURPLUS5
	VI_PET_LEVEL          = helper.VI_PET_LEVEL
	VI_PET_STAR           = helper.VI_PET_STAR
	VI_PET_STAR2          = helper.VI_PET_STAR2
	VI_PET_APTITUDE       = helper.VI_PET_APTITUDE
	VI_PET_APTITUDE2      = helper.VI_PET_APTITUDE2
	VI_PETCOIN            = helper.VI_PETCOIN
	VI_GT_GACHA           = helper.VI_GT_GACHA
	VI_GT_VIP             = helper.VI_GT_VIP
	VI_GT_HGC             = helper.VI_GT_HGC
	VI_GT_GWCBOX          = helper.VI_GT_GWCBOX
	VI_GT_HEROBOX         = helper.VI_GT_HEROBOX
	VI_WheelCoin          = helper.VI_WheelCoin
)

const (
	Attr_Atk = "ATK"
	Attr_Def = "DEF"
	Attr_HP  = "HP"
)

const (
	Build       = ProtobufGen.Build
	BuildHash   = ProtobufGen.BuildHash
	BuildBranch = ProtobufGen.BuildBranch
	BuildDate   = ProtobufGen.BuildDate
)

func GetProtoDataVer() string {
	return fmt.Sprintf("(%v) %s %s %s",
		Build, BuildHash, BuildBranch, BuildDate)
}

// 洗练数据
const (
	EquipTrickInitCount = 3
	EquipTrickMaxCount  = 5
)

// GameMode Counter Type
const (
	CounterTypeNull = iota
	CounterTypeGoldLevel
	CounterTypeFineIronLevel
	CounterTypeTrial
	CounterTypeBoss
	CounterTypeDCLevel
	CounterTypeFish
	CounterTypeGeneralQuest
	CounterTypeFishHC
	CounterTypeGVE
	CounterTypeTeamPvp // 10
	CounterTypeTeamPvpRefresh
	CounterTypeSimplePvp
	CounterTypeHitHammerDailyLimit
	CounterTypeWorshipTimes
	CounterTypeFreeGuildBoss // 15
	CounterTypeGuildBossBuyTime
	CounterTypeFreeGuildBigBoss
	CounterTypeGuildBigBossBuyTime
	CounterTypeEatBaozi
	CounterTypeFengHuoFreeExtraReward // 20
	CounterTypeFengHuoFreeSeniorSweep
	CounterTypeFengHuoMaxSeniorSweep
	CounterTypeExpedition
	CounterTypeMain
	CounterTypeElite
	CounterTypeHell
	CounterGateEnemy
	CounterFitMe
	CounterFestivalBoss
	ConnterHeroDiff // 30
	CounterTypeWspvpChallenge
	CounterTypeWspvpRefresh
	CounterTypeFBInvitation
	ConterTypeWBoss
	CounterTypeCountMax
)

// data conf
type DataVerConf struct {
	Build       int    `toml:"Build"`
	BuildHash   string `toml:"BuildHash"`
	BuildBranch string `toml:"BuildBranch"`
	BuildDate   string `toml:"BuildDate"`
}

var (
	DataVerCfg    DataVerConf
	HotDataVerCfg DataVerConf
	HotDataValid  bool
)

func GetGameDataConfVer() string {
	return fmt.Sprintf("(%v) %s %s %s",
		DataVerCfg.Build,
		DataVerCfg.BuildHash,
		DataVerCfg.BuildBranch,
		DataVerCfg.BuildDate)
}

// 相应的配置显示+1
const (
	GachaType_Surplus3 = 12 + iota
	GachaType_Surplus4
	GachaType_Surplus5
)

//跑马灯相关
const (
	_                                   = iota
	SN_MaxVip                           // 最高vip
	SN_FirstFinishLevel                 // 首个通关
	SN_Fish_Bouns                       // 钓鱼
	SN_Fish_Normal                      // 钓鱼
	IDS_GANK_WIN_1                      // 切磋
	IDS_GANK_WIN_2                      // 切磋
	IDS_GANK_WIN_3                      // 切磋
	IDS_HITEGG_1                        // 砸蛋
	IDS_PAY_XUANWU                      // 付费激活玄武
	IDS_GUILDBUFF                       = 32
	IDS_SHENSHOUDIYI                    = 33
	IDS_SHENSHOUMANJI                   = 34
	IDS_IWANTYOU                        = 35
	IDS_ROLLINFO_PLAYER_MONEYGOD        = 36 //招财猫
	IDS_GVG_MARQUEE_CONSECUTIVE_VICTORY = 37 // GVG军团战连胜
	IDS_GVG_MARQUEE_CHANGAN_OCCUPIED    = 38 // GVG军团战占领长安城
	IDS_GUILD_WORSHIPCRIT_MARQUEE       = 39 //  军团膜拜
	IDS_WORLDBOSS_CHAMPINE              = 47 // 世界boss第一名

	IDS_White_Gacha = 41 //
	IDS_WuShuang    = 42 //无双

	IDS_CSRob_Rob_Without_Helper = 43
	IDS_CSRob_Rob_With_Helper    = 44

	IDS_BLACK_GACHA_HERO   = 45
	IDS_BLACK_GACHA_WEAPON = 46

	IDS_Astrology_Augur = 48
)

const (
	IDS_KILLHERO    = 11   // 首次击杀boss
	IDS_PAY_JIANGLI = 1002 // 首冲礼包
)
