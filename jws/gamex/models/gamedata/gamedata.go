package gamedata

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/astaxie/beego/utils"
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
)

type GDPlayerAttribute struct {
	BasicAtt  ProtobufGen.ATTRIBUTES
	LevelAtts []*ProtobufGen.LEVELATTRIBUTES
	GSRadio   ProtobufGen.ATTRIBUTESGS
}

func GetPlayerLevelAttr(level uint32) *ProtobufGen.LEVELATTRIBUTES {
	//直接取值
	/*
		for i, att := range gdPlayerAtt.LevelAtts {
			if uint32(i) == level {
				return att
			}
		}
	*/
	lv := int(level)
	if lv >= len(gdPlayerAtt.LevelAtts) || lv < 0 {
		logs.Error("GetPlayerLevelAttr Level To Large %d", level)
		return nil
	}
	return gdPlayerAtt.LevelAtts[lv]
}

func GetPlayerBasicAtt() ProtobufGen.ATTRIBUTES {
	return gdPlayerAtt.BasicAtt
}

func GetPlayerGSRadio() ProtobufGen.ATTRIBUTESGS {
	return gdPlayerAtt.GSRadio
}

var (
	gdPlayerAtt GDPlayerAttribute
)

type loadDataFunc func(dfilepath string, loadfunc func(string))

func loadBin(cfgname string) ([]byte, error) {
	errgen := func(err error, extra string) error {
		return fmt.Errorf("gamex.models.gamedata loadbin Error, %s, %s", extra, err.Error())
	}

	//	path := GetDataPath()
	//	appConfigPath := filepath.Join(path, cfgname)

	file, err := os.Open(cfgname)
	if err != nil {
		return nil, errgen(err, "open")
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, errgen(err, "stat")
	}

	buffer := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buffer) //read all content
	if err != nil {
		return nil, errgen(err, "readfull")
	}

	return buffer, nil
}

func GetDataPath() string {
	workPath, _ := os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	appConfigPath := filepath.Join(AppPath, "conf")
	if workPath != AppPath {
		if utils.FileExists(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = filepath.Join(workPath, "conf")
		}
	}
	return appConfigPath
}

func loadAttributesData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	buffer, err := loadBin(filepath)
	errcheck(err)

	plyAtt := &ProtobufGen.ATTRIBUTES_ARRAY{}
	err = proto.Unmarshal(buffer, plyAtt)
	errcheck(err)

	gdPlayerAtt.BasicAtt = *plyAtt.GetItems()[0]

	baseAttr = AvatarAttr{}

	baseAttr.CritRate = gdPlayerAtt.BasicAtt.GetCritRate()
	baseAttr.ResilienceRate = gdPlayerAtt.BasicAtt.GetResilienceRate()
	baseAttr.CritValue = gdPlayerAtt.BasicAtt.GetCritValue()
	baseAttr.ResilienceValue = gdPlayerAtt.BasicAtt.GetResilienceValue()
	baseAttr.IceDamage = gdPlayerAtt.BasicAtt.GetIceDamage()
	baseAttr.IceDefense = gdPlayerAtt.BasicAtt.GetIceDefense()
	baseAttr.IceBonus = gdPlayerAtt.BasicAtt.GetIceBonus()
	baseAttr.IceResist = gdPlayerAtt.BasicAtt.GetIceResist()
	baseAttr.FireDamage = gdPlayerAtt.BasicAtt.GetFireDamage()
	baseAttr.FireDefense = gdPlayerAtt.BasicAtt.GetFireDefense()
	baseAttr.FireBonus = gdPlayerAtt.BasicAtt.GetFireBonus()
	baseAttr.FireResist = gdPlayerAtt.BasicAtt.GetFireResist()
	baseAttr.LightingDamage = gdPlayerAtt.BasicAtt.GetLightingDamage()
	baseAttr.LightingDefense = gdPlayerAtt.BasicAtt.GetLightingDefense()
	baseAttr.LightingBonus = gdPlayerAtt.BasicAtt.GetLightingBonus()
	baseAttr.LightingResist = gdPlayerAtt.BasicAtt.GetLightingResist()
	baseAttr.PoisonDamage = gdPlayerAtt.BasicAtt.GetPoisonDamage()
	baseAttr.PoisonDefense = gdPlayerAtt.BasicAtt.GetPoisonDefense()
	baseAttr.PoisonBonus = gdPlayerAtt.BasicAtt.GetPoisonBonus()
	baseAttr.PoisonResist = gdPlayerAtt.BasicAtt.GetPoisonResist()
	baseAttr.HitRate = gdPlayerAtt.BasicAtt.GetHitRate()
	baseAttr.DodgeRate = gdPlayerAtt.BasicAtt.GetDodgeRate()

	baseAttrGs = baseAttr.GS()
}

func loadLevelAttributesData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lvlAtt := &ProtobufGen.LEVELATTRIBUTES_ARRAY{}
	err = proto.Unmarshal(buffer, lvlAtt)
	errcheck(err)

	gdPlayerAtt.LevelAtts = lvlAtt.GetItems()
	//logs.Trace("loadLevelAttributesData: %d", len(gdPlayerAtt.LevelAtts))
}

func LoadGameData(rootPath string) {
	dataRelPath := "data"
	dataAbsPath := filepath.Join(GetDataPath(), dataRelPath)
	hotDataValid := false
	// etcd
	err := LoadHotDataVerFromEtcd()
	if err == nil && game.Cfg.IsHotDataValid() {
		// load data from s3
		if err := uutil.LoadHotData2LocalFromS3(
			game.Cfg.HotDataS3Bucket(),
			game.Cfg.GetHotDataVerC(),
			GetHotDataPath()); err == nil {
			hotDataValid = true
			logs.Info("LoadGameData get hot data success %s", game.Cfg.GetHotDataVerC())
		}
	}
	// 在非prod，并确认开启denger更新情况下，启动全部用热更的数据
	if hotDataValid && !game.Cfg.IsRunModeProd() {
		dataAbsPath = GetHotDataPath()
		dataRelPath = GetHotDataRelPath()
	}

	type Ver struct {
		Ver DataVerConf
	}
	var ver Ver
	cfg := config.NewConfigToml(filepath.Join(dataRelPath, "proto_ver.toml"), &ver)
	DataVerCfg = ver.Ver
	if cfg == nil {
		logs.Error("gamedata load ver config nil")
		panic(fmt.Errorf("gamedata load ver config nil"))
	}
	logs.Info("gamedata load ver config: %v", DataVerCfg)

	load := func(dfilepath string, loadfunc func(string)) {
		//logs.Info("LoadGameData %s start", filepath)
		loadfunc(filepath.Join(rootPath, dataAbsPath, dfilepath))
		logs.Trace("LoadGameData %s success", dfilepath)
	}

	mkHeroData(load)

	load("allrefreshtime.data", loadRefreshTimeData)
	//注意副将有顺序
	load("newgeneral.data", loadGeneralCofig)
	load("generalstart.data", loadGeneralStarCofig)
	load("generalrelation.data", loadGeneralRelCofig)
	load("relationlevel.data", loadGeneralRelLevelCofig)

	load("ngqdetail.data", loadGeneralQuestConfig)
	load("ngqrefreshtime.data", loadGeneralQuestRefTimeConfig)
	load("ngqsettings.data", loadGeneralQuestSettingConfig)

	load("levelattributes.data", loadLevelAttributesData)
	// 注意顺序，item
	load("jadexp.data", loadJadeData)
	load("item.data", loadItemData)

	// 注意顺序 universalmaterial要在所有loot之前
	load("universalmaterial.data", loadUniversalMaterialDatas)

	//Loot
	load("itemgroup.data", loadLootItemGroup)
	load("template.data", loadLootTemplate)
	load("stagerandreward.data", loadStageRewardRand)
	load("stagelimitreward.data", loadStageRewardLimit)
	load("firststagelimitreward.data", loadFirstStageRewardLimit)
	load("levelenemyconfig.data", loadLevelEnemyConfig)
	load("acdata.data", loadAcData)
	load("equipupgrade.data", loadEquipUpgradeCofig)
	load("evolution.data", loadEquipEvolutionCofig)
	load("equipresolve.data", loadEquipResolveGiveConfig)
	load("materialenhance.data", loadEquipMaterialEnhance)
	load("level_info.data", loadStageData)
	load("chapteraward.data", loadChapterReward)
	load("corplevel.data", loadCorpLevelInfo)
	load("formula.data", loadComposeCofig)
	load("quest.data", loadQuestDetailed)
	load("bornquest.data", loadQuestInit)

	// 下面次序有依赖
	load("euiptrickdetail.data", loadEquipTrickConfig)
	load("euiptricksettings.data", loadEquipTrickSettingConfig)
	load("euiptrickrule.data", loadTrickRandPool)

	load("newstarlvup.data", loadStarUpConfig)
	load("newstarlvupsettings.data", loadStarUpSettingConfig)

	load("condition.data", loadConditionConfig)

	// 1v1竞技场
	load("bscpvppool.data", loadSimplePvpPoolConfig)
	load("bscpvprankreward.data", loadSimplePvpRewardConfig)
	load("bscpvpswtcost.data", loadSimplePvpSwitchCost)
	load("bscpvpconfig.data", loadSimplePvpConfig)
	load("bscpvprewardweek.data", loadSimplePvpWeekReward)
	load("bscpvpwinreward.data", loadSimplePvpDayWinReward)

	//注意下面两个顺序有依赖
	load("monthlyactivity.data", loadMonthlyActivityData)
	load("monthlygift.data", loadMonthlyGiftData)
	//注意上面两个顺序有依赖

	//注意下面两个顺序有依赖
	load("giftactivitylist.data", loadGiftActivityData)
	load("dailygift.data", loadDailyGiftData)
	//注意上面两个顺序有依赖

	//注意下面顺序有依赖
	load("storeblank.data", loadStoreDataConfig)
	load("storegroup.data", loadStorePoolDataConfig)
	load("refreshprice.data", loadStoreRefreshCostDataConfig)
	load("refreshtime.data", loadStoreAutoRefreshDataConfig)
	//注意上面顺序有依赖

	//注意下面**四**个顺序有依赖
	load("normalgacha.data", loadGachaDataConfig)
	load("rewardserial.data", loadGachaRewardDataConfig)
	load("gachagroup.data", loadGachaPoolDataConfig)
	load("gachasettings.data", loadGachaCommonDataConfig)
	load("gachagroup.data", loadGachaExtPoolDataConfig)
	//注意上面**四**顺序有依赖

	//注意下面两个顺序有依赖
	load("firstgive.data", loadPayFirstGiveCofig)
	load("pay.data", loadPayCofig)
	//注意上面两个顺序有依赖

	load("roleweight.data", loadStageAvatarPower)

	load("arousal.data", loadAvatarArousalConfig)
	load("skillupgrade.data", loadSkillLevelInfo)
	load("skillpractice.data", loadSkillPracticeLevelInfo)
	load("hcinfluence.data", loadGachaProbabilityToSpecPoolDataConfig)

	load("attributesgs.data", loadAttributesgsConfig)
	load("attributes.data", loadAttributesData)
	load("vipsettings.data", loadVIPConfig)

	load("energypurchase.data", loadEnergyPurchaseConfig)
	load("eatbaozicost.data", loadBaoZiPurchaseConfig)
	load("sprintpurchase.data", loadBossFightPointPurchaseConfig)
	load("scpurchase.data", loadScPurchaseConfig)
	load("stagepurchase.data", loadBuyEStageTimes)
	load("tpvptime.data", loadTeamPvpTimesConfig)
	load("pvptime.data", loadSimplePvpTimesConfig)
	load("htpointpurchase.data", loadHeroTalentPointConfig)

	load("config.data", loadCommonConfig)
	load("rankforgwc.data", loadRankForGWC)

	load("bossfight.data", loadBossPool)

	load("goldlevel.data", loadGoldLevelConfig)
	load("modecontrol.data", loadModeControlConfig)
	load("explevel.data", loadExpLevelCfg)
	load("dclevel.data", loadDCLevelCfg)

	load("privilege.data", loadPrivilegeData)
	load("initializesave.data", loadAvatarInitConfig)
	load("initializeitem.data", loadAvatarInitBagConfig)
	load("initializefashion.data", loadAvatarInitFashion)

	load("zhhans.data", loadSensitiveWordConfig)
	load("namezhhans.data", loadNameZhans)
	load("namehmt.data", LoadNameHmt)
	load("nameja.data", loadNameJapan)
	load("nameen.data", LoadNameEN)
	load("namevi.data", LoadNameVN)
	load("nameko.data", LoadNameKO)
	load("nameth.data", LoadNameTH)
	load("story.data", loadStoryDetailed)
	load("section.data", loadStorySection)

	// 下面次序有依赖
	load("codegiftpatch.data", loadGiftCodeBatchData)
	load("codegiftgroup.data", loadGiftCodeGroupData)

	// 下面次序有依赖
	load("cdtgiftmain.data", loadActivityGiftByCondMain)
	load("cdtgiftvalue.data", loadActivityGiftByCond)
	load("cdtgifttiming.data", loadActivityGiftByTime)

	load("conditionrole.data", loadFteConditionRoleConfig)
	load("jadeconditon.data", loadJadeConditionConfig)

	// iap 下面次序有依赖
	load("iapbase.data", loadIAPBaseData)
	load("iapmain.data", loadIAPMainData)
	load("iapconfig.data", loadIAPConfig)
	load("dubbleraward.data", loadIAPCard)

	load("guildnumbers.data", loadGuildMemNumberData)

	load("antiratio.data", loadAntiCheatRatioConfig)
	load("skillratio.data", loadAntiCheatSkillRatioConfig)

	load("commonids.data", loadIDSData)

	load("guildrankingaward.data", loadGuildRankingData)
	load("guildconfig.data", loadGuildConfig)

	// shop 下面依赖顺序
	load("shopgoods.data", loadShopGood)
	load("shopdisplay.data", loadShop)

	load("modecontrol.data", loadGameModeControlData)

	load("destinygenerallevel.data", loadDestinyGeneralLevelData)
	load("destinygeneralunlock.data", loadDestinyGeneralUnlockData)
	load("newdestinygenerallevel.data", loadNewDestinyGeneralLevelData)
	load("destinyconfig.data", loadDestinyGeneralConfig)

	load("guildposition.data", loadGuildMemPosData)
	load("guildbag.data", loadGuildInventory)
	load("lostgoodshop.data", loadGuildLostInventory)

	load("gstmembercap.data", loadGuildGSTMemCap)
	load("gstgbossfightbonus.data", loadGuildGSTBossFightBonus)
	load("gstdailytaskexp.data", loadGuildGSTDailyTaskExp)
	load("gstgoldbonus.data", loadGuildGSTGoldBonus)
	load("gstgateenemybonus.data", loadGuildGSTGateEnemyBonus)
	load("gstwannahero.data", loadGuildGSTWannaHero)
	load("gstconfig.data", loadGuildGSTConfig)

	// trial
	load("level_trial.data", loadTrial)

	// GatesEnemy兵临城下 以下有顺序
	load("geconfig.data", loadGatesEnemyConfig)
	load("geenemygroup.data", loadGatesEnemyGroup)
	load("geenemy.data", loadGatesEnemyEnemy)
	load("geloot.data", loadGatesEnemyLoot)
	load("gegift.data", loadGatesEnemyGift)

	load("guildsign.data", loadGuildSignData)

	// recover
	load("recover.data", loadRecover)
	load("recoverdetail.data", loadRecoverRetail)
	load("recoversettings.data", loadRecoverSetting)

	load("dailyaward.data", loadDailyAward)
	load("spcialactivity.data", loadActivitySpecRewards)

	load("bscpvpbot.data", loadDroidDatas)

	// city
	load("fishingcost.data", loadFishCostConfig)
	load("fishingreward.data", loadFishReward)

	load("gveconfig.data", loadGVEConfigData)
	load("gveenemygroup.data", loadGVEEnemyGroupData)
	load("gveenemy.data", loadGVEEnemyData)
	load("gvemodel.data", loadGVEEnemyModelData)
	load("gveloot.data", loadGVELootData)

	// server open activity
	load("dayrank.data", loadSevOpnRankConfig)
	load("rankaward.data", loadSevOpnRankAwardConfig)
	load("fightaward.data", loadSevOpnFightAwardConfig)
	load("guildrank.data", loadSevOpnGuildRankConfig)
	load("guildlead.data", loadSevOpnGuildRankLeadConfig)

	// gank
	load("gankconfig.data", loadGankData)
	load("gankids.data", loadGankIDS)

	// account 7day
	load("activityshop.data", loadAccount7DayShop)
	load("activityquest.data", loadAccount7DayQuest)

	// server channel
	load("branchandroid13.data", loadSerChannelConfig)
	load("channelconst.data", loadChannelConstConfig)

	// team pvp
	load("tpvpmain.data", loadTPvpMain)
	load("tpvpmatch.data", loadTPvpMatch)
	load("tpvpfpass.data", loadTPvpPass)
	load("tpvpsector.data", loadTPvpSector)
	load("tpvprefresh.data", loadTPvpRefresh)
	load("tpvpwinreward.data", loadTPvpDayReards)

	// first pay
	load("firstpay.data", loadFirstPayCofig)

	// hit egg
	load("rebatecost.data", loadHitEggCost)
	load("rebatereward.data", loadHitEggReward)

	load("euiptrickselect.data", loadEquipTrickSelectData)

	// title
	load("titlelist.data", loadTitle)

	// grow fund
	load("growfund.data", loadGrowFundData)

	load("horselamp.data", loadHorseLamp)

	load("tpvpfpass.data", loadTeamPvpFirstPassReward)
	load("bscpvpsector.data", loadSimplePvpFirstPassReward)

	// LEVELGIFTPURCHASE
	load("levelgiftpurchase.data", loadLevelGift)

	load("gbconfig.data", loadGuildBossTimesConfig)

	load("newstarlvupstagecost.data", loadStarUpHcCostConfig)

	//Share WeChat
	load("sclshare.data", loadShareWeChatData)

	//神翼
	load("herowingstar.data", loadHeroSwingStarLevelData)
	load("herowinglevel.data", loadHeroSwingLevelData)
	load("herowingtable.data", loadHeroSwingTypeData)
	load("herowinglist.data", loadHeroSwingOwnData)

	//远征
	load("expeditionbot.data", loadExpeditionDroidDatas)
	load("expeditionlevel.data", loadExpeditionData)
	load("expeditionreward.data", loadExpeditionRewardData)
	load("passaward.data", loadExpeditionPassAwardData)
	load("expeditionconfig.data", loadExpeditionConfig)
	load("expeditionsweep.data", loadExpeditionSweep)

	//GVG军团战
	load("gvgcity.data", loadGVGCityIDData)
	load("gvgwinspoint.data", loadGVGWinsScore)
	load("gvgconfig.data", loadGVGConfigData)
	load("gvgguard.data", loadGVGDroidDatas)
	load("gvgacitygift.data", loadGVGActivityGift)
	load("gvgguildgift.data", loadGVGGuildGift)
	load("gvgdailygift.data", loadGVGDailyGift)
	load("gvgpointgift.data", loadGVGPointGift)

	//招财猫
	//load("moneygod.data",loadMoneyCatData)

	// 情缘
	load("relationactive.data", loadCompanionActive)
	load("relationevolution.data", loadCompanionEvolve)

	//节日Boss
	load("fbconfig.data", loadFestivallBossCfgData)
	load("fbreward.data", loadFestivallBossLootData)
	load("fbshop.data", loadFestivallShopData)
	load("fbbuycount.data", loadFestivalBossTimesConfig)

	// 好友
	load("friendcheat.data", loadFriendConfig)
	// 改名
	load("renamecost.data", loadRenameCostData)

	// 军团改名
	load("guildrenamechase.data", loadGuildRenameData)

	// 军团膜拜
	load("guildworshipreward.data", loadGuildWorshipData)
	load("guildworshipcrit.data", loadGuildWorshipCritData)
	load("guildworshipbox.data", loadGuildWorshipBoxRewardData)
	load("guildworshipcost.data", loadGuildWorshipCostData)
	load("guildactivity.data", loadGuildActivity)

	// 神兵
	load("gloryweapon.data", loadGloryWeaponData)
	load("gloryweaponlist.data", loadGloryWeaponListData)
	load("gwdeveloprandpolicy.data", loadGWDevelopRandPolicyData)

	// 出奇制胜 武将差异化
	load("hdplevel.data", loadHeroDiffLevelData)
	load("hdpreward.data", loadHeroDiffRewardData)
	load("hdpenemy.data", loadHeroDiffEnemyData)
	load("hdprewardlist.data", loadHeroDiffRewardList)
	load("hdplevelsection.data", loadHeroDiffLevelSection)
	load("hdpmodel.data", loadHeroDiffModelData)
	load("hdpconfig.data", loadHeroDiffConfig)
	load("fzbtu.data", loadHeroDiffFZBTU)
	load("fzbzhan.data", loadHeroDiffFZBZHAN)
	load("fzbhu.data", loadHeroDiffFZBHU)
	load("fzbshi.data", loadHeroDiffFZBSHI)
	load("zhanpoint.data", loadHeroDiffZHANPointData)
	load("tupoint.data", loadHeroDiffTUPointData)

	// 无双争霸
	load("wspvpmatch.data", loadWsPvpMatchConfig)
	load("wspvpfpass.data", loadWsPvpBestRankRewardConfig)
	load("wspvpsector.data", loadWsPvpTimeRewardConfig)
	load("wsboxreward.data", loadWsPvpChallengeRewardConfig)
	load("wspvprefresh.data", loadWsPvpRefreshConfig)
	load("wspvptime.data", loadWsPvpChallengeConfig)
	load("wspvpmain.data", loadWsPvpMainConfig)

	// OPPO 相关
	load("opposevendays.data", loadOPPOSignData)
	load("oppoeveryday.data", loadOPPODailyQuestData)

	// 押运粮草
	load("battlehero.data", loadCSRobBattleHeroConfig)
	load("cropsgift.data", loadCSRobCropGiftConfig)
	load("giftloot.data", loadCSRobGiftLootConfig)
	load("rcconfig.data", loadCSRobRCConfigConfig)
	load("refreshcrops.data", loadCSRobRefreshCropsConfig)

	load("fatetable.data", loadHeroDestinyData)
	load("fatelevel.data", loadFateLevelData)

	//星图系统
	load("starmap.data", loadAstrologyStarMapConfig)
	load("starsoul.data", loadAstrologyStarSoulConfig)
	load("starsoulupgrade.data", loadAstrologyStarSoulUpgradeConfig)
	load("augur.data", loadAstrologyAugurConfig)
	load("starmapconfig.data", loadAstrologyStarMapConfigConfig)

	//世界Boss
	load("wbossconfig.data", loadWBConfigData)
	load("wbrankreward.data", loadWBRankRewardData)
	load("wbosslevel.data", loadWBBossLevelData)
	load("wbdemagereward.data", loadWBDamageRewardData)
	load("wbossdata.data", loadWBBossData)

	//英雄灵宠系统
	load("magicpetconfig.data", loadMagicPetConfig)
	load("petaptitude.data", loadPetAptitude)
	load("petlevel.data", loadPetLevel)
	load("petstar.data", loadPetStar)
	load("typeaptitude.data", loadTypeAptitude)

	//战阵系统
	load("battlearmy.data",loadBattleArmy)
	load("battlearmylevel.data",loadBattleArmyLevel)

	// 离线资源找回
	load("recoverresources.data", loadOfflineRecover)

	// 掉落礼包
	load("packagegroup.data", loadPackageGroup)

	//组队BOSS
	load("tbossboxdata.data", loadTBBoxData)
	load("tbossconfig.data", loadTBBossConfig)
	load("tbossmaindata.data", loadTBBossMainData)
	load("tbossdungeon.data", loadTBBossDungeon)
	load("tbossenemy.data", loadTBBossEnemy)
	load("tbossboxloot.data", loadTBBoxLoot)
	load("tbossvipcontrol.data", loadTBBossVipCtrl)
	load("tbossherotype.data", loadTBBossHeroType)
	load("troopsmessage.data", loadTRoopsMessage)

	//跑马灯
	load("rollinfo.data", loadRollInfo)

	mkIdentityDatas(load)
	mkGuildBossDatas(load)
	mkFenghuoDatas(load)
	mkEatBaoziDatas(load)
	mkWantGeneralDatas(load)

	//mkServerGroupDatas(load) 已改为热更处理

	//TODO: temp 将会同其他限时名将表格数据合并处理

	// hot data
	if hotDataValid {
		dataAbsPath = GetHotDataPath()
		dataRelPath = GetHotDataRelPath()
	}
	loadHotGameDataFromInit(dataAbsPath, dataRelPath)
	HotDataValid = hotDataValid

	checkGameData()
	processDataBeforeAll()
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func IsGameDevMode() bool {
	return game.Cfg.RunMode == "dev"
}

// 2015/5/8
func getDataTimeInData(str string) int64 {
	if str == "" {
		return 0
	}

	t, err := time.ParseInLocation("2006/1/2", str, util.ServerTimeLocal)
	if err != nil {
		logs.Error("getDataTimeInData %s err by %s", str, err.Error())
	}
	return t.Unix()
}

func setDataVer2Etcd() {
	for _, sid := range game.Cfg.ShardId {
		prefix := GetHotDataEtcdRoot(game.Cfg.EtcdRoot, version.Version, fmt.Sprintf("%d", game.Cfg.Gid))
		key := fmt.Sprintf("%s/%d/%s", prefix, sid, etcd.KeyBaseDataBuild)
		if err := etcd.Set(key, fmt.Sprintf("%d", DataVerCfg.Build), 0); err != nil {
			logs.Error("setDataVer2Etcd key %s err %s", key, err.Error())
		}
		key = fmt.Sprintf("%s/%d/%s", prefix, sid, etcd.KeyHotDataBuild)
		if err := etcd.Set(key, fmt.Sprintf("%d", GetHotDataVerCfg().Build), 0); err != nil {
			logs.Error("setDataVer2Etcd key %s err %s", key, err.Error())
		}
		key = fmt.Sprintf("%s/%d/%s", prefix, sid, etcd.KeyHotDataSeq)
		if err := etcd.Set(key, game.Cfg.GetHotDataVerC(), 0); err != nil {
			logs.Error("setDataVer2Etcd key %s err %s", key, err.Error())
		}
		key = fmt.Sprintf("%s/%d/%s", prefix, sid, etcd.KeyHotDataGid)
		if err := etcd.Set(key, fmt.Sprintf("%d", game.Cfg.Gid), 0); err != nil {
			logs.Error("setDataVer2Etcd key %s err %s", key, err.Error())
		}
		game.Cfg.HotDataVerC_Suc = game.Cfg.GetHotDataVerC()
	}
}

func DebugResetData() {
	dataRelPath := "data"
	dataAbsPath := filepath.Join(GetDataPath(), dataRelPath)
	loadHotGameDataFromInit(dataAbsPath, dataRelPath)
}
