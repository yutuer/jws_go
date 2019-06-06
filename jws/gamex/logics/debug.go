package logics

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	ma1 "vcs.taiyouxi.net/jws/gamex/models/market_activity"
	"vcs.taiyouxi.net/jws/gamex/models/pay"
	"vcs.taiyouxi.net/jws/gamex/modules/balance_timer"
	"vcs.taiyouxi.net/jws/gamex/modules/city_fish"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice/teamboss"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	"vcs.taiyouxi.net/jws/gamex/modules/friend"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
	"vcs.taiyouxi.net/jws/gamex/modules/hero_diff"
	"vcs.taiyouxi.net/jws/gamex/modules/herogacharace"
	"vcs.taiyouxi.net/jws/gamex/modules/hour_log"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/modules/market_activity"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/modules/team_pvp"
	"vcs.taiyouxi.net/jws/gamex/modules/ws_pvp"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	jws_helper "vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

const (
	debugOp_AddItem                       string = "AddItem"
	debugOp_SaveSelfToInitAccount         string = "SaveSelfToInitAccount"
	debugOp_AddEnergy                     string = "AddEnergy"
	debugOp_ResetStage                    string = "ResetStage"
	debugOp_ResetSelfToInitAccount        string = "ResetSelfToInitAccount"
	debugOp_SetVipLv                      string = "SetVipLv"
	debugOp_SetDebugTime                  string = "SetDebugTime"
	debugOp_SetSeed                       string = "SetSeed"
	debugOp_ForceQuestCanReceive          string = "ForceQuestCanReceive"
	debugOp_ForceQuestFinish              string = "ForceQuestFinish"
	debugOp_SetCorpLv                     string = "SetCorpLv"
	debugOp_SetHeroLv                     string = "SetHeroLv"
	debugOp_AddHc                         string = "AddHc"
	debugOp_AddHcWithRMBpoints            string = "AddHcWithRMBpoints"
	debugOp_RankBalanceAll                string = "RankBalanceAll"
	debugOp_RankWeekBalanceAll            string = "RankWeekBalanceAll"
	debugOp_ResetGameModeTime             string = "ResetGameModeTime"
	debugOp_ResetGameModeCount            string = "ResetGameModeCount"
	debugOp_ResetDailyTaskBoss            string = "ResetDailyTaskBoss"
	debugOp_CleanStage                    string = "CleanStage"
	debugOp_SetDebugServerTime            string = "SetDebugServerTime"
	debugOp_SetAllStarLv                  string = "SetAllStarLv"
	debugOp_AddGuildLeaveTime             string = "AddGuildLeaveTime"
	debugOp_AddItems                      string = "AddItems"
	debugOp_SetNextBoss                   string = "SetNextBoss"
	debugOp_SetPost                       string = "SetPost"
	debugOp_AddIAPLogTest                 string = "AddIAPLogTest"
	debugOp_SetFashionEquipTimeout        string = "SetFashionEquipTimeout"
	debugOp_SetFashionBagTimeout          string = "SetFashionBagTimeout"
	debugOp_SetTrialCurLvl                string = "SetTrialCurLvl"
	debugOp_DebugSetLastBalanceTime       string = "DebugSetLastBalanceTime"
	debugOp_GateEnemy                     string = "DebugGateEnemy"
	debugOp_AddGuildXp                    string = "AddGuildXp"
	debugOp_SetFishReward                 string = "SetFishAward"
	debugOp_ResetFishReward               string = "ResetFishAward"
	debugOp_SetGlobalFishReward           string = "SetGlobalFishAward"
	debugOp_SetSevenDayRankEndToday       string = "SetSevenDayRankEndToday" // DebugSetSevenDayRankEndToday
	debugOp_TeamPvpExchange               string = "TeamPvpExchange"
	debugOp_TeamPvpCalcRes                string = "TeamPvpCalcResult"
	debugOp_ClearIAPFirstPay              string = "ClearIAPFirstPay"
	debugOp_SetHotActivityTime            string = "SetHotActivityTime"
	debugOp_ResetGameData                 string = "ResetGameData"
	debugOp_AddIAP                        string = "AddIAP"
	debugOp_SetGuildInventoryTime         string = "SetGuildInventoryTime"
	debugOp_ResetGuildInventoryTime       string = "ResetGuildInventoryTime"
	debugOp_AddGuildInventoryItem         string = "AddGuildInventoryItem"
	debugOp_AddGuildBossClean             string = "DebugGuildBossClean"
	debugOp_AddHeroTalentPoint            string = "AddHeroTalentPoint"
	debugOp_SetHeroTalentLevel            string = "SetHeroTalentLevel"
	debugOp_SetGSTLvl                     string = "SetGSTLvl"
	debugOp_ResetWantGeneral              string = "ResetWantGeneral"
	debugOp_SetGateEnemyTime              string = "SetGateEnemyTime"
	debugOp_SetGateEnemyPlayTime          string = "SetGateEnemyPlayTime"
	debugOp_GetGateEnemyTime              string = "GetGateEnemyTime"
	debugOp_SetSimplePvpWeek              string = "SetSimplePvpWeekDay"
	debugOp_AddSimplePvpBotScore          string = "AddSimplePvpBotScore"
	debugOp_TestItemGroup                 string = "TestItemGroup"
	debugOp_ResetGateEnemy                string = "ResetGateEnemy"
	debugOp_ReloadExpeditionEnemy         string = "ReloadExpeditionEnemy"
	debugOp_GetExpeditionEnemy            string = "GetExpeditionEnemy"
	debugOp_SetExpeditionPassNum          string = "SetExpeditionPassNum"
	debugOp_SetMoneyCatNum                string = "SetMoneyCatNum"
	debugOp_AutoChangeChiefTime           string = "AutoChangeChiefTime"
	debugOp_DebugSetGVGTime               string = "DebugSetGVGTime"
	debugOp_DebugResetGVGTime             string = "DebugResetGVGTime"
	debugOp_AutoChangeChief               string = "AutoChangeChief"
	debugOp_SyncGVGTime                   string = "SyncGVGTime"
	debugOp_ChangeLimitGood               string = "ChangeLimitGood"
	debugOp_HourLogOutput                 string = "HourLogOutput"
	debugOp_UnlockAllLevel                string = "UnlockAllLevel"
	debugOp_ResetRedPacketForGuild        string = "ResetRedPacketForGuild"
	debugOp_ResetRedPacketForPlayer       string = "ResetRedPacketForPlayer"
	debugOp_AllSendRePacket               string = "AllSendRedPacket"
	debugOp_AutoJoinGuild                 string = "AutoJoinGuild"
	debugOp_BatchAddFriend                string = "BatchAddFriend"
	debugOp_GetWorshipSign                string = "GetWorshipsign"
	debugOp_ResetHeroDiff                 string = "ResetHeroDiff"
	debugOp_ResetFashionWeapon            string = "ResetFashionWeapon"
	debugOp_ResetGuildWorship             string = "ResetGuildWorship"
	debugOp_FashionWeaponToMax            string = "FashionWeaponToMax"
	debugOp_AddHeroDiffBot                string = "AddHeroDiffBot"
	debugOp_SetMaterialLvl                string = "SetMaterialLvl"
	debugOp_SetWhiteGachaWish             string = "SetWhiteGachaWish"
	debugOp_GetWSPVPMarquee               string = "GetWSPVPMarquee"
	debugOp_SetWhiteGachaFree             string = "SetWhiteGachaFree"
	debugOp_WspvpExchangeRank             string = "ExchangeWspvpRank"
	debugOp_FinishWspvpRankTitle          string = "FinishWspvpRankTitle"
	debugOp_MarketActivityReward          string = "MarketActivityReward"
	debugOp_MarketActivityClear           string = "MarketActivityClear"
	debugOp_SetActivityRankByRobot        string = "SetActivityRank"
	debugOp_CSRobClearCount               string = "CSRobClearCount"
	debugOp_CSRobSetMaxCount              string = "CSRobSetMaxCount"
	debugOp_CSRobBuildCar                 string = "CSRobBuildCar"
	debugOp_CSRobAddEnemy                 string = "CSRobAddEnemy"
	debugOp_CSRobDoWeekReward             string = "CSRobDoWeekReward"
	debugOp_CSRobGuildList                string = "CSRobGuildList"
	debugOp_SetDestinyGeneralLv           string = "SetDestinyGeneralLv"
	debugOp_AddGuildLostInventoryItem     string = "AddGuildLostInventoryItem"
	debugOp_AstrologyGetInfo              string = "AstrologyGetInfo"
	debugOp_AstrologyInto                 string = "AstrologyInto"
	debugOp_AstrologyDestroyInHero        string = "AstrologyDestroyInHero"
	debugOp_AstrologyDestroyInBag         string = "AstrologyDestroyInBag"
	debugOp_AstrologyDestroySkip          string = "AstrologyDestroySkip"
	debugOp_AstrologySoulUpgrade          string = "AstrologySoulUpgrade"
	debugOp_AstrologyAugur                string = "AstrologyAugur"
	debugOp_AstrologyFillBag              string = "AstrologyFillBag"
	debugOp_AstrologyClearMyAstrologyData string = "AstrologyClearMyAstrologyData"
	debugOp_AstrologyClearMyAstrologyBag  string = "AstrologyClearMyAstrologyBag"
	debugOp_AstrologyTryAugurStatistic    string = "AstrologyTryAugurStatistic"
	debugOp_ResetMarketActivity           string = "ResetMarketActivity"
	debugOp_ClearGachaCount               string = "ClearGachaCount"
	debugOp_SimulateFreeGacha             string = "SimulateFreeGacha"
	debugOp_GetGachaCount                 string = "GetGachaCount"
	debugOp_SetOfflineTime                string = "SetOfflineTime"
	debugOp_CopyWspvpLog                  string = "CopyWspvpLog"
	debugOp_OpenSurplusGacha              string = "OpenSurplusGacha"
	debugOp_ResetSurplusInfo              string = "ResetSruplusInfo"
	debugOp_TeamBossFight                 string = "tb_fight"
	debugOp_TeamBossReady                 string = "tb_ready"
	debugOp_TeamBossCreate                string = "tb_create"
	debugOp_TeamBossJoin                  string = "tb_join"
	debugOp_TeamBossBox                   string = "tb_box"
	debugOp_TeamBossBoxNormal             string = "tb_box_normal"
	debugOp_MagicPetLevel                 string = "mp_lev"
	debugOp_MagicPetStar                  string = "mp_star"
	debugOp_MagicPetTalent                string = "mp_talent"
	debugOp_MagicPetSetChangeTalentCounte string = "mp_setCountTimes"
)

type handlerDebugOp struct {
	MethodName string
	MethodCall func(*Account, *RequestDebugOp) string
}

var handlerDebugOpMap = map[string]handlerDebugOp{}

func regDebugOpHandle(op handlerDebugOp) {
	handlerDebugOpMap[op.MethodName] = op
}

type RequestDebugOp struct {
	Req
	Type string   `codec:"typ"`
	P1   int64    `codec:"p1"`
	P2   int64    `codec:"p2"`
	P3   int64    `codec:"p3"`
	P4   int64    `codec:"p4"`
	Nums []int    `codec:"nums"`
	Strs []string `codec:"strs"`
}

type ResponseDebugOp struct {
	SyncRespWithRewards
}

func (p *Account) DebugOp(r servers.Request) *servers.Response {
	req := &RequestDebugOp{}
	resp := &ResponseDebugOp{}
	initReqRsp(
		"Debug/DebugOpResponse",
		r.RawBytes,
		req, resp, p)

	re := "ok"

	typs := strings.Split(req.Type, "#")
	if len(typs) >= 2 {
		typSub := typs[0]
		typP := typs[1]
		if typSub == debugOp_AddItem {
			gives := gamedata.CostData{}
			if req.P1 >= 0 {
				gives.AddItem(typP, uint32(req.P1))
				account.GiveBySync(p.Account, &gives, resp, "Debug")
			} else {
				gives.AddItem(typP, uint32(-req.P1))
				account.CostBySync(p.Account, &gives, resp, "Debug")
			}
		}
	}

	switch req.Type {
	case debugOp_SaveSelfToInitAccount:
		re = p.DebugSaveSelfToInitAccount(req.P1)
	case debugOp_AddEnergy:
		re = p.DebugAddEnergy(req.P1)
	case debugOp_ResetStage:
		re = p.DebugResetStage()
	case debugOp_ResetSelfToInitAccount:
		re = p.DebugResetSelfToInitAccount()
	case debugOp_SetVipLv:
		re = p.DebugSetVipLv(req.P1)
	case debugOp_SetDebugTime:
		re = p.DebugSetDebugTime(req.P1)
	case debugOp_SetSeed:
		re = p.DebugSetSeed(req.P1)
	case debugOp_ForceQuestCanReceive:
		re = p.DebugForceQuestCanReceive(req.P1)
	case debugOp_ForceQuestFinish:
		re = p.DebugForceQuestFinish(req.P1)
	case debugOp_SetCorpLv:
		re = p.DebugSetCorpLv(req.P1)
	case debugOp_SetHeroLv:
		p.DebugSetHeroLv(req.P1, req.P2)
		re = "ok"
	case debugOp_AddHc:
		re = p.DebugAddHc(req.P1, req.P2)
	case debugOp_AddHcWithRMBpoints:
		re = p.DebugAddHcWithRMBpoints(req.P1, req.P2, resp)
	case debugOp_RankBalanceAll:
		re = p.DebugRankBalanceAll()
	case debugOp_RankWeekBalanceAll:
		re = p.DebugRankWeekBalanceAll()
	case debugOp_ResetGameModeTime:
		re = p.DebugResetGameModuTime()
	case debugOp_ResetGameModeCount:
		re = p.DebugResetGameModeCount()
	case debugOp_ResetDailyTaskBoss:
		re = p.DebugResetDailyTaskBoss()
	case debugOp_CleanStage:
		re = p.CleanStage()
	case debugOp_SetDebugServerTime:
		re = p.DebugSetDebugServerTime(req.P1)
	case debugOp_SetAllStarLv:
		re = p.DebugSetAllStarLv(req.P1)
	case debugOp_AddGuildLeaveTime:
		re = p.DebugAddGuildLeaveTime(req.P1)
	case debugOp_AddItems:
		re = p.DebugAddItems(req.Strs, req.Nums, resp)
	case debugOp_SetNextBoss:
		re = p.DebugSetNextBoss(req.Strs[0])
	case debugOp_SetPost:
		re = p.DebugSetPost(req.Strs[0])
	case debugOp_AddIAPLogTest:
		re = p.DebugAddIAPLogTest()
	case debugOp_SetFashionEquipTimeout:
		re = p.DebugSetFashionEquipTimeout(req.P1)
	case debugOp_SetFashionBagTimeout:
		re = p.DebugSetFashionBagTimeout(req.P1)
	case debugOp_SetTrialCurLvl:
		re = p.DebugSetTrialCurLvl(int32(req.P1))
	case debugOp_DebugSetLastBalanceTime:
		re = p.DebugSetLastBalanceTime(req.P1)
	case debugOp_GateEnemy:
		re = p.DebugGateEnemy(req.P1, req.P2, req.P3)
	case debugOp_AddGuildXp:
		re = p.AddGuildXp(req.P1)
	case debugOp_SetFishReward:
		re = p.FishAward(req.P1, resp)
	case debugOp_ResetFishReward:
		re = p.FishResetAward()
	case debugOp_SetGlobalFishReward:
		if len(req.Strs) > 0 {
			re = p.FishSetGlobalAward(req.Strs[0])
		}
	case debugOp_SetSevenDayRankEndToday:
		rank.DebugSetSevenDayRankday(p.AccountID.ShardId, uint32(req.P1))
		re = "ok"
	case debugOp_TeamPvpExchange:
		team_pvp.GetModule(p.AccountID.ShardId).CommandExec(team_pvp.TeamPvpCmd{
			Typ:           team_pvp.TeamPvp_Cmd_Debug_Exchange,
			DebugOperRank: []int{int(req.P1), int(req.P2)},
		})
		re = "ok"
	case debugOp_TeamPvpCalcRes:
		team_pvp.CalcTeamPvpResult(int(req.P1), int(req.P2), p.GetRand())
		re = "ok"
	case debugOp_ClearIAPFirstPay:
		p.clearIAPFirstPay()
		re = "ok"
	case debugOp_SetHotActivityTime:
		p.setHotActivityTime(int(req.P1), req.P2, req.P3)
		p.Profile.GetMarketActivitys().DataBuild = 0
		re = "ok"
	case debugOp_ResetGameData:
		p.resetGamedata()
		re = "ok"
	case debugOp_AddIAP:
		re = p.DebugAddIAP(uint32(req.P1), uint32(req.P2), uint32(req.P3), resp)
	case debugOp_SetGuildInventoryTime:
		if p.GuildProfile.InGuild() {
			guildUuid := p.GuildProfile.GuildUUID
			guild.GetModule(p.AccountID.ShardId).DebugSetGuildInventoryTime(guildUuid, req.P1)
		}
		re = "ok"
	case debugOp_ResetGuildInventoryTime:
		if p.GuildProfile.InGuild() {
			guildUuid := p.GuildProfile.GuildUUID
			guild.GetModule(p.AccountID.ShardId).DebugResetGuildInventoryTime(guildUuid)
		}
		re = "ok"
	case debugOp_AddGuildInventoryItem:
		ids := make([]string, 0, len(req.Strs))
		counts := make([]uint32, 0, len(req.Nums))
		for i, loot := range req.Strs {
			if nil != gamedata.GetGuildInventoryCfg(loot) {
				ids = append(ids, loot)
				counts = append(counts, uint32(req.Nums[i]))
			}
		}
		if len(ids) > 0 && p.GuildProfile.InGuild() {
			guildUuid := p.GuildProfile.GuildUUID
			guild.GetModule(p.AccountID.ShardId).AddGuildInventory(guildUuid, ids, counts, "debug")
		}
		re = "ok"
	case debugOp_AddGuildLostInventoryItem:
		ids := make([]string, 0, len(req.Strs))
		counts := make([]uint32, 0, len(req.Nums))
		for i, loot := range req.Strs {
			if nil != gamedata.GetGuildLostInventoryCfg(loot) {
				ids = append(ids, loot)
				counts = append(counts, uint32(req.Nums[i]))
			}
		}
		logs.Debug("add guild inventory, %v, %v", ids, counts)
		if len(ids) > 0 && p.GuildProfile.InGuild() {
			guildUuid := p.GuildProfile.GuildUUID
			guild.GetModule(p.AccountID.ShardId).AddGuildInventory(guildUuid, ids, counts, "debug_lost")
		}
		re = "ok"
	case debugOp_AddGuildBossClean:
		guildId := p.GuildProfile.GuildUUID
		guild.GetModule(p.AccountID.ShardId).ActBossDebugClean(guildId, p.Account.AccountID.String(), req.P1)
		re = "ok"
	case debugOp_AddHeroTalentPoint:
		p.Profile.GetHeroTalent().DebugAddTP(uint32(req.P1), p.Profile.GetProfileNowTime())
		re = "ok"
	case debugOp_SetHeroTalentLevel:
		p.debugSetHeroTalentLevel(int(req.P1), uint32(req.P2), uint32(req.P3))
		re = "ok"
	case debugOp_SetGSTLvl:
		if p.GuildProfile.InGuild() {
			guild.GetModule(p.AccountID.ShardId).DebugSetGuildScienceLevel(
				p.GuildProfile.GuildUUID, p.AccountID.String(),
				gamedata.GST_Typ(req.P1), req.P2)
		}
		re = "ok"
	case debugOp_ResetWantGeneral:
		p.Profile.GetWantGeneralInfo().DebugReset(p.Account)
		re = "ok"
	case debugOp_SetGateEnemyTime:
		if len(req.Strs) > 0 { // "reset"为重置时间
			gamedata.DebugSetGEStartTime(req.Strs[0])
		}
		re = "ok"
	case debugOp_SetGateEnemyPlayTime:
		if len(req.Strs) > 0 { // "reset"为重置时间
			gamedata.DebugSetGEPlayTime(req.Strs[0], uint32(req.P1))
		}
		re = "ok"
	case debugOp_GetGateEnemyTime:
		p.Profile.GetGatesEnemy().DebugGetActTime()
		re = "ok"
	case debugOp_SetSimplePvpWeek:
		p.DebugSetSimplePvpWeekRestDay(int(req.P1))
	case debugOp_AddSimplePvpBotScore:
		p.DebugAddSimplePvpBotScore(int(req.P1), int(req.P2), int(req.P3))
	case debugOp_TestItemGroup:
		p.testItemGroup(req.P1, req.Strs)
		re = "ok"
	case debugOp_ResetGateEnemy:
		if p.GuildProfile.InGuild() {
			guild.GetModule(p.AccountID.ShardId).DebugResetGateEnemy(
				p.GuildProfile.GuildUUID, p.AccountID.String())
		}
	case debugOp_ReloadExpeditionEnemy:
		p.Profile.GetExpeditionInfo().GetEnemyNextTime = 0
		gs := int64(p.Profile.GetData().CorpCurrGS_HistoryMax)
		if req.P1 > 0 {
			gs = req.P1
		}
		p.Profile.GetExpeditionInfo().LoadEnemyToday(p.AccountID.String(),
			gs,
			p.Profile.GetProfileNowTime())
	case debugOp_GetExpeditionEnemy:
		p.Profile.GetExpeditionInfo().GetEnemyToday(p.AccountID.String(),
			p.Profile.GetProfileNowTime())
		p.setExpeditionEnmy()
	case debugOp_SetExpeditionPassNum:
		p.Profile.GetExpeditionInfo().ExpeditionNum = int32(req.P1)

	case debugOp_SetMoneyCatNum:
		p.Profile.GetMoneyCatInfo().MoneyCatTime = 0
	case debugOp_AutoChangeChiefTime:
		gamedata.ChiefMaxAbsentTime = req.P1
		gamedata.GuildSafeTimeOnAwake = req.P2
	case debugOp_DebugSetGVGTime:
		p.DebugSetGVGTime(req.P1)
	case debugOp_DebugResetGVGTime:
		p.DebugResetGVGTime()
	case debugOp_AutoChangeChief:
		if p.GuildProfile.InGuild() {
			guild.GetModule(p.AccountID.ShardId).DebugAutoChangeChief(
				p.GuildProfile.GuildUUID, p.AccountID.String())
		}
	case debugOp_SyncGVGTime:
		p.DebugSyncGVGTime()
	case debugOp_ChangeLimitGood:
		goodId := req.P1
		goodDuration := req.P2
		goodStartTime := req.Strs[0]
		for i, item := range gamedata.GetHotDatas().LimitGoodConfig.Items {
			if item.Item.GetLimitGoodsID() == uint32(goodId) {
				newTime, _ := time.ParseInLocation("20060102_15_04", goodStartTime, util.ServerTimeLocal)
				gamedata.GetHotDatas().LimitGoodConfig.Items[i].StartTime = newTime.Unix()
				gamedata.GetHotDatas().LimitGoodConfig.Items[i].Duration = int(goodDuration)
				logs.Debug("debug limit good, %d, %d, %d", goodId, item.StartTime, item.Duration)
				break
			}
		}
	case debugOp_HourLogOutput:
		hour_log.Get(p.AccountID.ShardId).DebugOutput()
	case debugOp_UnlockAllLevel:
		p.DebugUnlockAllLevel()
	case debugOp_ResetRedPacketForGuild:
		if p.GuildProfile.InGuild() {
			guild.GetModule(p.AccountID.ShardId).DebugResetRedPacketForGuild(p.GuildProfile.GuildUUID)
		}
	case debugOp_ResetRedPacketForPlayer:
		p.GuildProfile.RedPacketInfo.DailyReset(p.GetProfileNowTime())
		if p.GuildProfile.InGuild() {
			guild.GetModule(p.AccountID.ShardId).DebugResetRedPacketForPlayer(p.GuildProfile.GuildUUID, p.Profile.Name)
		}
	case debugOp_AllSendRePacket:
		if p.GuildProfile.InGuild() {
			guild.GetModule(p.AccountID.ShardId).DebugAllSendRedPacket(p.GuildProfile.GuildUUID)
		}
	case debugOp_AutoJoinGuild:
		if p.GuildProfile.InGuild() {
			ret := rank.GetModule(p.AccountID.ShardId).RankCorpGs.Get(p.AccountID.String())
			guild.GetModule(p.AccountID.ShardId).DebugAutoJoinGuild(p.GuildProfile.GuildUUID, req.P1, ret.TopN)
		}
	case debugOp_BatchAddFriend:
		p.DebugBatchAddFriend(req.P1, req.P2)

	case debugOp_GetWorshipSign:
		x := p.DebugGetWorshipGoodSign()
		return rpcErrorWithMsg(resp, x, "CODE_Cost_Er")
	case debugOp_ResetHeroDiff:
		p.DebugResetHeroDiff(req.P1)
	case debugOp_ResetFashionWeapon:
		heros := p.Profile.GetHero()
		heros.HeroExclusiveWeapon = [account.AVATAR_NUM_MAX]account.HeroExclusiveWeapon{}
	case debugOp_ResetGuildWorship:
		p.DebugResetGuildWorship()
		resp.OnChangeGuildWorshipInfo()
	case debugOp_FashionWeaponToMax:
		weapon := &p.Profile.GetHero().HeroExclusiveWeapon[int(req.P1)]
		p.DebugFashionWeaponMax(weapon)
	case debugOp_AddHeroDiffBot:
		p.DebugAddHeroDiffBot(int(req.P1), int(req.P2), int(req.P3))
	case debugOp_SetMaterialLvl:
		p.DebugSetMaterialLvl(int(req.P1), uint32(req.P2))
		resp.OnChangeEquip()
	case debugOp_SetWhiteGachaWish:
		p.DebugSetWhiteGachaWish(req.P1)
		resp.OnChangeWhiteGacha()
	case debugOp_GetWSPVPMarquee:
		p.DebugGetWSPVPMarquee()
	case debugOp_SetWhiteGachaFree:
		p.DebugSetWhiteGachaFree()
	case debugOp_WspvpExchangeRank:
		p.DebugExchangeWspvpRank(int(req.P1))
	case debugOp_FinishWspvpRankTitle:
		p.DebugFinishWspvpRankTitle()
	case debugOp_MarketActivityReward:
		p.DebugMarketActivityReward(req.P1, req.P2)
		re = "MarketActivityReward OK"
	case debugOp_MarketActivityClear:
		p.DebugMarketActivityClear(req.P1, req.P2)
		re = "MarketActivityClear OK"
	case debugOp_SetActivityRankByRobot:
		p.DebugSetActivityRankWithRobot(req.P1, req.P2, req.P3)
	case debugOp_CSRobClearCount:
		re = p.DebugCSRobClearCount()
	case debugOp_CSRobSetMaxCount:
		re = p.DebugCSRobSetMaxCount()
	case debugOp_CSRobBuildCar:
		re = p.DebugCSRobBuildCar()
	case debugOp_CSRobAddEnemy:
		re = p.DebugCSRobAddEnemy()
	case debugOp_CSRobDoWeekReward:
		re = p.DebugCSRobDoWeekReward()
	case debugOp_CSRobGuildList:
		re = p.DebugCSRobGuildList()
	case debugOp_SetDestinyGeneralLv:
		p.DebugSetDestinyGeneralLv(req.P1, req.P2, resp)
	case debugOp_AstrologyGetInfo:
		p.DebugAstrologyGetInfo()
	case debugOp_AstrologyInto:
		if 0 != len(req.Strs) {
			p.DebugAstrologyInto(uint32(req.P1), uint32(req.P2), req.Strs[0])
		}
	case debugOp_AstrologyDestroyInHero:
		p.DebugAstrologyDestroyInHero(uint32(req.P1), uint32(req.P2))
	case debugOp_AstrologyDestroyInBag:
		if 0 != len(req.Strs) {
			p.DebugAstrologyDestroyInBag(req.Strs[0])
		}
	case debugOp_AstrologyDestroySkip:
		rares := []uint32{}
		if 0 != req.P1 {
			rares = append(rares, uint32(req.P1))
		}
		if 0 != req.P2 {
			rares = append(rares, uint32(req.P2))
		}
		if 0 != req.P3 {
			rares = append(rares, uint32(req.P3))
		}
		if 0 != req.P4 {
			rares = append(rares, uint32(req.P4))
		}
		p.DebugAstrologyDestroySkip(rares)
	case debugOp_AstrologySoulUpgrade:
		p.DebugAstrologySoulUpgrade(uint32(req.P1), uint32(req.P2))
	case debugOp_AstrologyAugur:
		p.DebugAstrologyAugur(req.P1 == 1)
	case debugOp_AstrologyFillBag:
		p.DebugAstrologyFillBag()
	case debugOp_AstrologyClearMyAstrologyData:
		p.DebugAstrologyClearMyAstrologyData()
	case debugOp_AstrologyClearMyAstrologyBag:
		p.DebugAstrologyClearMyAstrologyBag()
	case debugOp_AstrologyTryAugurStatistic:
		p.DebugAstrologyTryAugurStatistic(int(req.P1))
	case debugOp_ResetMarketActivity:
		if req.P1 == 0 {
			p.Profile.MarketActivitys.Activitys = make([]ma1.PlayerMarketActivity, 0)
		} else {
			for i, data := range p.Profile.MarketActivitys.Activitys {
				if data.ActivityId == uint32(req.P1) {
					list := p.Profile.MarketActivitys.Activitys
					p.Profile.MarketActivitys.Activitys = append(list[:i], list[i+1:]...)
					break
				}
			}
		}
	case debugOp_ClearGachaCount:
		p.DebugClearGachaCount(int(req.P1))
	case debugOp_SimulateFreeGacha:
		p.DebugSimulateFreeGacha(int(req.P1), int(req.P2), resp)
	case debugOp_GetGachaCount:
		p.DebugGetGachaCount(int(req.P1))
	case debugOp_SetOfflineTime:
		playerName := req.Strs[0]
		offlineTime, err := time.ParseInLocation("20060102_1504", req.Strs[1], util.ServerTimeLocal)
		if err != nil {
			logs.Warn("<SetOfflineTime> time err, %s, %v", req.Strs[1], err)
			return rpcError(resp, 1)
		}
		db := driver.GetDBConn()
		defer db.Close()
		if !db.IsNil() {
			nameTable := fmt.Sprintf("names:%d:%d", p.AccountID.GameId, p.AccountID.ShardId)
			acid, err := redis.String(db.Do("HGET", nameTable, playerName))
			if err != nil {
				logs.Warn("<SetOfflineTime> acid err, %s, %v", acid, err)
				return rpcError(resp, 1)
			} else {
				profileTable := fmt.Sprintf("profile:%s", acid)
				now_t := time.Now().Unix()
				absoluteTime := offlineTime.Unix() - now_t
				db.Do("HSET", profileTable, "debugAbsoluteTime", absoluteTime)
			}
		}
	case debugOp_CopyWspvpLog:
		ws_pvp.DebugCopyWspvpLog(p.GetWSPVPGroupId(), p.AccountID.String(), int(req.P1))
	case debugOp_OpenSurplusGacha:
		p.Profile.GetHeroSurplusInfo().DebugOpen(p.Profile.GetProfileNowTime())
	case debugOp_ResetSurplusInfo:
		p.Profile.GetHeroSurplusInfo().DebugReset(p.Profile.GetProfileNowTime())
	case debugOp_TeamBossCreate:
		_, code, err := teamboss.CreateRoom(p.AccountID.ShardId, p.AccountID.String(), &jws_helper.CreateRoomInfo{
			RoomLevel: 0,
			JoinInfo: jws_helper.PlayerJoinInfo{
				PlayerDetailInfo: nil,
				AcID:             p.AccountID.String(),
				GS:               p.Profile.GetData().CorpCurrGS,
				Avatar:           p.Profile.GetCurrAvatar(),
				Name:             p.Profile.Name,
			},
			BossID:    "teamboss_4",
			TeamTypID: 1004,
			SceneID:   "teamboss_level_4",
		})
		if err != nil {
			logs.Error("!error: %v, code: %v", err, code)
		}
	case debugOp_TeamBossJoin:
		_, code, err := teamboss.JoinRoom(p.AccountID.ShardId, p.AccountID.String(), &jws_helper.JoinRoomInfo{
			RoomID: teamboss.DefaultRoomID,
			JoinInfo: jws_helper.PlayerJoinInfo{
				PlayerDetailInfo: nil,
				AcID:             p.AccountID.String(),
				GS:               p.Profile.GetData().CorpCurrGS,
				Avatar:           p.Profile.GetCurrAvatar(),
				Name:             p.Profile.Name,
			},
		})
		if err != nil {
			logs.Error("!error: %v, code: %v", err, code)
		}
	case debugOp_TeamBossBox:
		p.Profile.GetTeamBossStorageInfo().SetNewBoxForCheat(req.Strs[0], int(req.P1), req.P2)

	case debugOp_TeamBossBoxNormal:
		p.Profile.GetTeamBossStorageInfo().SetTBBoxNormalForCheat(req.Strs[0])

	case debugOp_TeamBossFight:
		accData := &helper.Avatar2ClientByJson{}
		account.FromAccount2Json(accData, p.Account, p.Profile.GetCurrAvatar())
		detail, err := json.Marshal(accData)
		if err != nil {
			logs.Error("json marshal error by %v", err)
		}
		_, code, err := teamboss.StartFight(p.AccountID.ShardId, p.AccountID.String(), &jws_helper.StartFightInfo{
			RoomID:     teamboss.DefaultRoomID,
			AcID:       p.AccountID.String(),
			BattleInfo: detail,
		})
		if err != nil {
			logs.Error("error: %v, code: %v", err, code)
		}
	case debugOp_TeamBossReady:
		accData := &helper.Avatar2ClientByJson{}
		account.FromAccount2Json(accData, p.Account, p.Profile.GetCurrAvatar())
		detail, err := json.Marshal(accData)
		if err != nil {
			logs.Error("json marshal error by %v", err)
		}
		_, code, err := teamboss.ReadyFight(p.AccountID.ShardId, p.AccountID.String(), &jws_helper.ReadyFightInfo{
			RoomID:     teamboss.DefaultRoomID,
			AcID:       p.AccountID.String(),
			BattleInfo: detail,
			Status:     jws_helper.TBPlayerStateReady,
		})
		if err != nil {
			logs.Error("error: %v, code: %v", err, code)
		}
	case debugOp_MagicPetLevel:
		p.Profile.GetHero().HeroMagicPets[req.P1].GetPets()[0].Lev = uint32(req.P2)
	case debugOp_MagicPetStar:
		p.Profile.GetHero().HeroMagicPets[req.P1].GetPets()[0].Star = uint32(req.P2)
	case debugOp_MagicPetTalent:
		pet := &p.Profile.GetHero().HeroMagicPets[req.P1].GetPets()[0]
		pet.CasualCompreTalent = int32(float32(req.P2) / float32(gamedata.GetMagicPetConfig().GetMinimumUnit()))
		//先确定平均单个属性资质
		preTalentsValue := float32(req.P2) / float32(gamedata.GetMagicPetConfig().GetRandomMulAptitude()) * float32(gamedata.GetMagicPetConfig().GetRandomAptitude())
		//生成资质值,talentsNum为资质数量
		talentsNum := gamedata.GetMagicPetConfig().GetAttributeAmount()
		myCasualTalentsValue := generateTalentsValues(preTalentsValue, talentsNum, p.Account.GetRand())
		//生成资质
		pet.CasualTalents = generateTalentsTypes(pet.CasualTalents, myCasualTalentsValue, p.Account.GetRand())
	case debugOp_MagicPetSetChangeTalentCounte:
		pet := &p.Profile.GetHero().HeroMagicPets[req.P1].GetPets()[0]
		//req.P2==0代表设置普通按键次数
		if req.P2 == 0 {
			pet.NormalChangeCountTimes = uint32(req.P3)
		} else if req.P2 == 1 {
			pet.SpecialChangeCountTimes = uint32(req.P3)
		}

	default:
		if handle, exist := handlerDebugOpMap[req.Type]; exist && nil != handle.MethodCall {
			re = handle.MethodCall(p, req)
		} else {
			logs.Warn("unknown DebugOp Method [%s]", req.Type)
		}
	}

	if req.Type != debugOp_SetDebugServerTime &&
		req.Type != debugOp_SetSeed &&
		req.Type != debugOp_AddIAPLogTest {
		resp.OnChangeSC()
		resp.OnChangeHC()
		resp.OnChangeEnergy()
		resp.OnChangeAvatarExp()
		resp.OnChangeCorpExp()
		resp.OnChangeBag()
		resp.OnChangeStageAll()
		resp.OnChangeEquip()
		resp.OnChangeQuestAll()
		resp.OnChangeGiftStateChange()
		resp.OnChangeMonthlyGiftStateChange()
		resp.OnChangeHC()
		resp.OnChangeGeneralAllChange()
		resp.OnChangeVIP()
		resp.OnChangeBuy()
		resp.OnChangeAllGameMode()
		resp.OnChangeSimplePvp()
		resp.OnChangeGuildInfo()
		resp.OnChangeGuildInventory()
		resp.OnChangeJadeFull()
		resp.OnChangeFashionBag()
		resp.OnChangeTrial()
		resp.OnChangeGatesEnemyData()
		resp.OnChangeGatesEnemyPushData()
		resp.OnChangeAllGameMode()
		resp.OnChangeSevenDayRank()
		resp.OnChangeIAPGoodInfo()
		resp.OnChangeTitle()
		resp.OnChangeMarketActivity()
		resp.OnChangeHeroTalent()
		resp.OnChangeGuildScience()
		resp.OnChangeWantGeneralInfo()
		resp.OnChangeDestinyGeneral()
		resp.OnChangerExpeditionInfo()
		resp.OnChangeFriendList()
		resp.OnChangeBlackList()
		resp.onChangeExclusiveWeaponInfo()
		resp.OnChangeWhiteGacha()
		resp.OnChangeMagicPetInfo()
		resp.mkInfo(p)
	}
	if req.Type == debugOp_SetDebugServerTime {
		resp.OnChangeAllGameMode()
		resp.mkInfo(p)
	}

	resp.SetMsg(re)
	return rpcSuccess(resp)
}

func (p *Account) DebugSetSimplePvpWeekRestDay(day int) {
	gamedata.GetSimplePvpConfig().SetWeekRewardResetDay(day)
	logs.Debug("new set day %d", gamedata.GetSimplePvpConfig().GetWeekRewardResetDay())
}

func (p *Account) DebugAddSimplePvpBotScore(actId, count, score int) {
	cfg := gamedata.GetHotDatas().Activity.GetActivitySimpleInfoById(uint32(actId))
	for i := 0; i < count; i++ {
		_, err := herogacharace.Get(p.AccountID.ShardId).UpdateScore(
			herogacharace.HGRActivity{
				GroupID:    gamedata.GetHotDatas().Activity.GetShardGroup(uint32(p.AccountID.ShardId)),
				ActivityId: uint32(actId),
				StartTime:  cfg.StartTime,
				EndTime:    cfg.EndTime,
			}, uint64(score),
			herogacharace.HGRankMember{
				AccountID:  fmt.Sprintf("10:100:%d", i),
				PlayerName: fmt.Sprintf("Name:%d", i),
			})
		if err != nil {
			logs.Error("herogacharace UpdateScore act %d err %v",
				actId, err)
		}
	}
}

func (p *Account) DebugAddHeroDiffBot(score int, ID int, count int) {
	extraData := hero_diff.HeroDiffRankData{
		AcID:       p.AccountID.String(),
		FreqAvatar: []int{0, 1, 2},
	}
	info := p.GetSimpleInfo()
	for i := 0; i < count; i++ {
		info.HeroDiffScore[hero_diff.HeroDiffID2Index(ID)] = score + i
		info.AccountID = fmt.Sprintf("0:10:BOT%d", i)
		info.Name = fmt.Sprintf("BOT%d", i)
		rank.GetModule(p.AccountID.ShardId).RankByHeroDiff[hero_diff.HeroDiffID2Index(ID)].AddWithExtraData(&info, extraData)
	}
}

func (p *Account) DebugSaveSelfToInitAccount(id int64) string {
	p.Profile.Name = ""

	profile_init_key := driver.GetInItProfileDBKey(p.Profile.DBName())
	bag_init_key := driver.GetInItProfileDBKey(p.BagProfile.DBName())

	// 清空profile，只留下equip
	curr_equip := make([]uint32, 0, 64)
	for _, in := range p.Profile.GetEquips().CurrEquips {
		curr_equip = append(curr_equip, in)
	}

	p.Profile = account.NewProfile(p.AccountID)
	p.Profile.GetEquips().CurrEquips = curr_equip

	// 要先把内存中的数据写进数据库
	cb := redis.NewCmdBuffer()
	p.Profile.DBSave(cb, true)
	p.BagProfile.DBSave(cb, true)
	db := driver.GetDBConn()
	defer db.Close()

	if _, err := db.DoCmdBuffer(cb, false); err != nil {
		logs.Error("DebugSaveSelfToInitAccount error %s", err.Error())
		return "failed"
	}

	is_profile_ok := driver.RedisSaveDataToOtherKey(p.Profile.DBName(), profile_init_key)
	is_bag_ok := driver.RedisSaveDataToOtherKey(p.BagProfile.DBName(), bag_init_key)

	if is_profile_ok && is_bag_ok {
		return "ok"
	}

	return "failed"
}

func (p *Account) DebugAddEnergy(num int64) string {
	p.Profile.GetEnergy().AddForce(p.AccountID.String(), "Debug", num)
	return "ok"
}

func (p *Account) DebugResetStage() string {
	p.Profile.GetStage().ResetStageCount()
	return "ok"
}

func (p *Account) DebugResetSelfToInitAccount() string {
	profile_init_key := driver.GetInItProfileDBKey(p.Profile.DBName())
	bag_init_key := driver.GetInItProfileDBKey(p.BagProfile.DBName())

	// 先转存
	is_profile_ok := driver.RedisSaveDataToOtherKey(profile_init_key, p.Profile.DBName())
	is_bag_ok := driver.RedisSaveDataToOtherKey(bag_init_key, p.BagProfile.DBName())

	if is_profile_ok && is_bag_ok {
		// 读取所有信息到内存
		err_profile := p.Profile.DBLoad(false)
		err_bag := p.BagProfile.DBLoad(false)

		if err_profile != nil {
			logs.Error("profile load err %s", err_profile.Error())
			return "failed"
		}

		if err_bag != nil {
			logs.Error("bag load err %s", err_bag.Error())
			return "failed"
		}

		return "ok"
	}

	return "failed"
}

func (p *Account) DebugSetVipLv(lv int64) string {
	info := gamedata.GetVIPCfg(int(lv))
	if info != nil {
		p.Profile.GetVip().RmbPoint = 0
		p.Profile.GetVip().AddRmbPoint(p.Account, info.RMBpoints+1, "Debug")
	}
	return "failed"
}

func (p *Account) DebugSetDebugTime(t int64) string {
	now_t := time.Now().Unix()
	p.Profile.DebugAbsoluteTime = t - now_t
	return "ok"
}

func (p *Account) DebugSetLastBalanceTime(t int64) string {
	now_t := time.Now().Unix()
	balance.GetModule(p.AccountID.ShardId).DebugSetBalanceTimeOffset(now_t - util.WeekSec - 1)
	return "ok"
}

func (p *Account) DebugSetSeed(t int64) string {
	p.Profile.Rng.Seed(t)
	return "ok"
}

func (p *Account) DebugForceQuestCanReceive(t int64) string {
	player_quest := p.Profile.GetQuest()
	player_quest.DebugSetCanReceiveForce(uint32(t))
	return "ok"
}

func (p *Account) DebugForceQuestFinish(t int64) string {
	param := []int{int(t)}
	p.FinishQuest(param, nil, true)
	return "ok"
}

func (p *Account) DebugSetCorpLv(lv int64) string {
	c := p.Profile.GetCorp()
	if lv <= 0 {
		lv = 1
	}
	ll := int64(gamedata.GetCommonCfg().GetCorpLevelUpperLimit())
	if lv > ll {
		lv = ll
	}
	oldLevel := c.Level
	c.Xp = 0

	if uint32(lv) > oldLevel {
		for i := oldLevel + 1; i <= uint32(lv); i++ {
			c.Level = uint32(i)
			c.OnLevelUp(p.AccountID.String(), c.Level, "Debug")
		}
	} else {
		c.Level = uint32(lv)
	}

	return "ok"
}

func (p *Account) DebugSetHeroLv(heroId, lv int64) {
	id := int(heroId)
	hero := p.Profile.GetHero()
	if id >= len(hero.HeroLevel) ||
		!p.Profile.GetCorp().IsAvatarHasUnlock(int(heroId)) {
		return
	}
	hero.HeroLevel[id] = uint32(lv)
	hero.SyncObj.SetNeedSync()
	p.Profile.GetData().SetNeedCheckMaxGS()
}

// 最好不要用
func (p *Account) DebugAddHc(t, v int64) string {
	p.Profile.GetHC().AddHC(p.AccountID.String(), int(t), v, p.Profile.GetProfileNowTime(), "Debug")
	return "ok"
}

func (p *Account) DebugAddHcWithRMBpoints(hc, rmb int64, sync interfaces.ISyncRspWithRewards) string {
	data := &gamedata.CostData{}
	data.AddItem("VI_HC_Buy", uint32(hc))
	give := &account.GiveGroup{}
	give.AddCostData(data)
	give.GiveBySyncAuto(p.Account, sync, "debug")
	return "ok"
}

func (p *Account) DebugRankBalanceAll() string {
	balance.GetModule(p.AccountID.ShardId).BalanceNotifyAll()
	return "ok"
}

func (p *Account) DebugRankWeekBalanceAll() string {
	req := servers.Request{
		Code: "MSG/TEST",
		RawBytes: encode(struct {
			AA string `codec:"aa"`
			BB string `codec:"bb"`
			CC string `codec:"cc"`
			DD string `codec:"dd"`
		}{
			AA: "dasfadsfadsf",
			BB: "fadfadsfadsf",
			CC: "ss",
			DD: "adss",
		}),
	}
	player_msg.GetModule(p.AccountID.ShardId).SendMsg(p.AccountID.String(), req)
	return "ok"
}

func (p *Account) DebugResetGameModuTime() string {
	for i := range p.Profile.GameMode.Info {
		info := &p.Profile.GameMode.Info[i]
		info.LastSlotBeginUnit = 0
		info.LastSecTimeAfC = 0
	}
	return "ok"
}

func (p *Account) DebugResetGameModeCount() string {
	// gamemode次数
	c := p.Profile.GetCounts()
	c.Counts = [counter.CounterTypeCountMax]int{}
	c.CountLastUpdateTime = [counter.CounterTypeCountMax]int64{}
	// 关卡次数
	for i := range p.Profile.GetStage().Stages {
		stage := &p.Profile.GetStage().Stages[i]
		stage.T_count = 0
	}
	// buy times
	p.Profile.GetBuy().DebugResetCount()
	// 公会离开冷却
	p.GuildProfile.NextEnterGuildTime = 0
	// team pvp 敌人刷新时间
	p.Profile.GetTeamPvp().NextEnemyRefTime = 0
	return "ok"
}

func (p *Account) DebugResetDailyTaskBoss() string {
	p.Profile.GetQuest().LastRefreshN[helper.Quest_PVE_Boss] = 0
	// remove from has_close and reveived
	for _, qid := range gamedata.GetDailyTaskQuestId(helper.Quest_PVE_Boss) {
		delete(p.Profile.GetQuest().Has_closed, qid)
		for i := 0; i < len(p.Profile.GetQuest().Received); i++ {
			if qid == p.Profile.GetQuest().Received[i].Id {
				p.Profile.GetQuest().Received[i].Id = 0
				p.Profile.GetQuest().Received_len -= 1
			}
		}
		p.Profile.ReceiveQuestByQid(qid, time.Now().Unix())
	}
	return "ok"
}

func (p *Account) CleanStage() string {
	p.Profile.GetStage().DebugCleanStage()
	return "ok"
}

func (p *Account) DebugSetDebugServerTime(t int64) string {
	shard := p.AccountID.ShardId
	now_t := time.Now().Unix()
	logs.Debug("DebugSetDebugServerTime %d %d -> %d", t, now_t, game.ServerStartTime(shard))
	oldt := time.Unix(game.ServerStartTime(shard), 0)
	util.SetServerStartTime(shard, t-now_t+game.ServerStartTime(shard))
	newt := time.Unix(game.ServerStartTime(shard), 0)
	logs.Debug("DebugSetDebugServerTime %v -> %v", oldt, newt)
	return "ok"
}

func (p *Account) DebugSetAllStarLv(t int64) string {
	p.Profile.GetEquips().DebugSetAllStarLv(uint32(t))
	return "ok"
}

func (p *Account) DebugAddGuildLeaveTime(t int64) string {
	p.Account.GuildProfile.NextEnterGuildTime += t
	return "ok"
}

func (p *Account) DebugSetNextBoss(t string) string {
	return "ok"
}

func (p *Account) DebugAddItems(itemIDs []string, counts []int, sync interfaces.ISyncRspWithRewards) string {
	gives := &gamedata.CostData{}
	for idx, id := range itemIDs {
		if idx < len(counts) {
			gives.AddItem(id, uint32(counts[idx]))
		}
	}
	account.GiveBySync(p.Account, gives, sync, "Debug")
	return "ok"
}

func (p *Account) DebugSetPost(t string) string {
	p.Account.GuildProfile.Post = t
	return "ok"
}

func (p *Account) DebugAddIAPLogTest() string {
	sec := time.Now().Unix()
	logiclog.LogIAP(p.AccountID.String(), p.Profile.AccountName, p.Profile.Name,
		p.Profile.CurrAvatar, p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		13, fmt.Sprintf("%s:%d:13", p.AccountID.String(), 0),
		"android.com.taiyouxi.ifsg.10", "order-test", 1, uutil.Android_Platform, "0", fmt.Sprintf("%d", sec),
		10, 0, p.GetIp(), p.Profile.GetVipLevel(),
		p.Profile.IAPGoodInfo.MoneySum, p.Profile.HC.GetHCFromBy(), p.Profile.HC.GetHCFromGive(),
		p.Profile.HC.GetHCFromCompensate(), func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	order := timail.IAPOrder{
		Order_no:      "order-test",
		Game_order:    "13",
		Game_order_id: fmt.Sprintf("%s:%d:13", p.AccountID.String(), 0),
		Amount:        "1",
		Channel:       "0",
		PayTime:       fmt.Sprintf("%d", sec),
		PkgInfo:       timail.PackageInfo{0, 0}}

	info, _ := json.Marshal(&order)
	mail_sender.SendAndroidIAPMail(p.AccountID.String(), string(info))
	return "ok"
}

func (p *Account) DebugAddIAP(idx uint32, pkgid uint32, subpkgid uint32, sync interfaces.ISyncRspWithRewards) string {
	give := &account.GiveGroup{}
	give.IAPData = &account.GiveIAPData{}
	give.IAPData.IAPGoodIndex = idx
	give.IAPData.IAPPkgInfo.PkgId = int(pkgid)
	give.IAPData.IAPPkgInfo.SubPkgId = int(subpkgid)
	cfg := gamedata.GetIAPInfo(idx)
	if cfg != nil {
		give.IAPData.IAPPrice = cfg.IOS_Rmb_Price
		give.GiveBySyncAuto(p.Account, sync, "debug")
	}
	return "ok"
}

func (p *Account) DebugSetFashionEquipTimeout(minute int64) string {
	now_time := p.Profile.GetProfileNowTime()
	for i := gamedata.FashionPart_Weapon; i <= gamedata.FashionPart_Armor; i++ {
		for aid := 0; aid < account.AVATAR_NUM_CURR; aid++ {
			id := p.Profile.GetAvatarEquips().GetEquip(aid, i)
			if id <= 0 {
				continue
			}
			ok, item := p.Profile.GetFashionBag().GetFashionInfo(id)
			if ok {
				ok, _ := gamedata.IsFashion(item.TableID)
				if !ok {
					continue
				}
				//				if gamedata.IsFashionPerm(itemCfg, item.Count) {
				//					continue
				//				}
				item.ExpireTimeStamp = now_time + minute*60
				p.Profile.GetFashionBag().Items[id] = item
			}
		}
	}
	return "ok"
}

func (p *Account) DebugSetFashionBagTimeout(minute int64) string {
	now_time := p.Profile.GetProfileNowTime()
	eq := make(map[uint32]uint32)
	for _, id := range p.Profile.GetAvatarEquips().Curr_equips {
		if id > 0 {
			eq[id] = id
		}
	}
	for id, item := range p.Profile.GetFashionBag().Items {
		if _, ok := eq[id]; ok {
			continue
		}
		if ok, _ := gamedata.IsFashion(item.TableID); ok {
			//			if gamedata.IsFashionPerm(itemCfg, item.Count) {
			//				continue
			//			}
			item.ExpireTimeStamp = now_time + minute*60
			p.Profile.GetFashionBag().Items[id] = item
		}
	}
	return "ok"
}

func (p *Account) DebugSetTrialCurLvl(lvl int32) string {
	p.Profile.GetPlayerTrial().DebugSetCurLvl(p.Account, lvl)
	return "ok"
}

func (p *Account) DebugGateEnemy(p1, p2, p3 int64) string {
	acID := p.AccountID.String()
	guildID := p.GuildProfile.GetCurrGuildUUID()
	p.Profile.GetGatesEnemy().OnDebugOp(acID, guildID, p1, p2, p3)
	return "ok"
}

func (p *Account) AddGuildXp(xp int64) string {
	guild.GetModule(p.AccountID.ShardId).DebugOp(p.GuildProfile.GuildUUID, p.AccountID.String(), 0, xp, 0, "")
	return "ok"
}

func (p *Account) FishAward(p1 int64, sync interfaces.ISyncRspWithRewards) string {
	give := &account.GiveGroup{}
	if p1 > 0 {
		// 全服奖励里抽奖
		fc := city_fish.FishCmd{
			Typ:            city_fish.CityFish_Cmd_Award,
			AName:          p.Profile.Name,
			ARand:          p.GetRand(),
			AwardCount:     1,
			DebugRewardIdx: int(p1),
		}
		res := city_fish.GetModule(p.AccountID.ShardId).CommandExec(fc)
		if !res.Success {
			return "ok"
		}
		if len(res.AwardId) > 0 { // 有全服奖
			give.AddItem(res.AwardItem[0], res.AwardCount[0])
		}
	} else { // 全服奖没有了, 给个人奖
		fCfg := gamedata.FishCost()

		// 随机物品
		gives, err := p.GetGivesByItemGroup(fCfg.GetLootDataID())
		if err != nil {
			return "ok"
		}

		if !gives.IsNotEmpty() {
			return "ok"
		}

		// 加物品
		give.AddCostData(gives.Gives())
	}

	give.GiveBySyncAuto(p.Account, sync, "DebugFishAward")
	return "ok"
}

func (p *Account) FishResetAward() string {
	fc := city_fish.FishCmd{
		Typ: city_fish.CityFish_Cmd_Debug_Reset_Award,
	}
	city_fish.GetModule(p.AccountID.ShardId).CommandExec(fc)
	return "ok"
}

func (p *Account) FishSetGlobalAward(s string) string {
	fc := city_fish.FishCmd{
		Typ:               city_fish.CityFish_Cmd_Debug_Set_Global_Award,
		DebugGlobalReward: s,
	}
	city_fish.GetModule(p.AccountID.ShardId).CommandExec(fc)
	return "ok"
}

func (p *Account) clearIAPFirstPay() {
	p.Profile.GetIAPGoodInfo().Infos = make([]pay.PayGoodInfo, 0, 64)
}

func (p *Account) setHotActivityTime(actId int, s, e int64) {
	now_t := time.Now().Unix()
	info := gamedata.GetHotDatas().Activity.GetActivitySimpleInfoById(uint32(actId))
	if info == nil {
		return
	}
	if info.ActivityType >= gamedata.ActHeroGachaRace_Begin && info.ActivityType <= gamedata.ActHeroGachaRace_End {
		// 检查限时名将，不能同时开两个
		nid := gamedata.GetHGRCurrValidActivityId()
		if nid > 0 && nid != uint32(actId) {
			logs.Error("debugSetActivityTime HeroGachaRace cant open two act at the same time")
			return
		}
		info := gamedata.GetHotDatas().Activity.GetActivitySimpleInfoById(uint32(actId))
		last_t := int64(gamedata.GetHotDatas().HotLimitHeroGachaData.GetHGRConfig().GetPublicityTime())
		if info != nil {
			if info.EndTime-last_t > now_t && e-last_t <= now_t { // 之前没结束，现在结束了
				// 强制结算
				herogacharace.Get(p.AccountID.ShardId).CheatBalance()
			}
		}
	}
	gamedata.DebugSetActivityInfo(actId, s, e)
}

func (p *Account) DebugSetGVGTime(t int64) string {
	now_t := time.Now().Unix()
	logs.Debug("DebugChgGVGTimeLine %d -> %d", now_t, t)
	gvg.GetModule(p.AccountID.ShardId).SetDebugOffSetTime(t - now_t)
	p.Profile.DebugAbsoluteTime = t - now_t
	return "ok"
}

func (p *Account) DebugResetGVGTime() string {
	gvg.GetModule(p.AccountID.ShardId).ResetDebugTime()
	return "ok"
}

func (p *Account) resetGamedata() {
	gamedata.DebugResetData()
	p.Profile.GetMarketActivitys().DebugReset(p.Profile.ChannelQuickId)
}
func (p *Account) DebugSyncGVGTime() {
	logs.Debug("DebugSyncGVGTime")
	p.Profile.DebugAbsoluteTime = gvg.GetModule(p.AccountID.ShardId).GetDebugOffsetTime()
}

func (p *Account) DebugGetWorshipGoodSign() uint32 {
	res, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
	if code.HasError() {
		logs.SentryLogicCritical(p.AccountID.String(), "SyncGuildInfoNeed Err %v %s",
			code, p.GuildProfile.GuildUUID)
	}
	for i, x := range res.GuildWorship.WorshipMember {
		if x.MemberSign == 2 {
			return uint32(i)
		}
	}
	return 0
}

func (p *Account) DebugResetGuildWorship() {
	res, code := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
	if code.HasError() {
		logs.SentryLogicCritical(p.AccountID.String(), "SyncGuildInfoNeed Err %v %s",
			code, p.GuildProfile.GuildUUID)
	}
	if res.GuildInfoBase.Base.Level >= gamedata.GetGuildActivityLvl(gamedata.GuildActivity_GUILD_WORSHIPCRIT_NAME) {
		guild.GetModule(p.AccountID.ShardId).DebugResetGuildWorship(p.GuildProfile.GuildUUID)
		p.GuildProfile.WorshipInfo.DailyReset(p.GetProfileNowTime())
	}
}

func (p *Account) DebugResetHeroDiff(t int64) {
	hero_diff.GetModule(p.AccountID.ShardId).DebugCleanRank()
	p.Profile.DebugAbsoluteTime = t - time.Now().Unix()
	logs.Debug("set debug time: %d by herodiff", t)
}
func (p *Account) testItemGroup(n int64, strs []string) {
	if len(strs) <= 0 {
		return
	}
	mode := os.FileMode(0644)
	f, err := os.OpenFile("test_itemgroup.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		logs.Error("debug testItemGroup OpenFile err %v", err)
		return
	}
	defer f.Close()
	f.Write([]byte(strs[0] + "\r\n"))
	for i := 0; i < int(n); i++ {
		loot, err := gamedata.LootItemGroupRandSelect(p.GetRand(), strs[0])
		if err == nil {
			f.Write([]byte(loot.ItemID + "\r\n"))
		} else {
			logs.Error("debug testItemGroup LootItemGroupRandSelect err %v", err)
		}
	}
}

func (p *Account) DebugUnlockAllLevel() {
	corp_lv, _ := p.Profile.GetCorp().GetXpInfo()
	for k, stage_data := range gamedata.GetAllStageData() {
		if stage_data.CorpLvRequirement != 0 &&
			corp_lv < uint32(stage_data.CorpLvRequirement) {
			continue
		}
		info := p.Profile.GetStage().GetStageInfo(
			gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()),
			k,
			p.GetRand())
		info.MaxStar = 7
		info.Sum_count = 1
		info.T_count = 1
		if !p.Profile.GetStage().IsStagePass(k) {
			logs.Debug("unlock failed: %d, %v", corp_lv, stage_data)
		}
		p.Profile.GetStage().AddChapterStarFromStage(stage_data.Id, 3)
	}
}

func (p *Account) DebugBatchAddFriend(count int64, isBlack int64) {
	info := friend.GetModule(p.AccountID.ShardId).GetGSClosePlayer(p.Profile.GetData().CorpCurrGS)
	i := int64(0)
	for _, v := range info {
		if v.IsNil() {
			continue
		}
		i++
		if isBlack > 0 {
			p.Friend.AddBlack(v)
		} else {
			p.Friend.AddFriend(v)
		}
		if i >= count {
			break
		}
	}
}

func (p *Account) DebugSetMaterialLvl(slot int, lvl uint32) {
	logs.Debug("slot %d lvl %d", slot, lvl)

	p.Profile.GetEquips().LvMaterialEnhance[slot] = lvl
	p.Profile.GetData().SetNeedCheckMaxGS()

}

func (p *Account) DebugSetWhiteGachaWish(wish int64) {
	p.Profile.GetWhiteGachaInfo().GachaBless = wish
	p.Profile.GetWhiteGachaInfo().GachaNum = wish
}

func (p *Account) DebugGetWSPVPMarquee() {
	p.Profile.GetWSPVPInfo().IsWSPVPMarquee(p.AccountID.GameId, p.AccountID.ShardId, p.AccountID.String(), p.Profile.Name)
}

func (p *Account) DebugSetWhiteGachaFree() {
	p.Profile.GetWhiteGachaInfo().LastGachaTime = 0
}

// 按照顺序养满当前品质的1条属性
func (p *Account) DebugFashionWeaponMax(weapon *account.HeroExclusiveWeapon) {
	promoteCfg := gamedata.GetEvolveGloryWeaponCfg(weapon.Quality)
	for i := 0; i < account.ExclusiveWeaponMaxAttr; i++ {
		if i+1 == gamedata.CRI_RATE || i+1 == gamedata.CRI_DAMAGE {
			continue
		}
		attrCfg := gamedata.GetWeaponAttrById(promoteCfg, i+1)
		if weapon.Attr[i] == attrCfg.GetValue() {
			continue
		}
		weapon.Attr[i] = attrCfg.GetValue()
		break
	}
}

const (
	DebugPorts_start = 7081
	DebugPorts_end   = 7099
)

func DebugTest() {
	if game.Cfg.IsRunModeProd() {
		return
	}
	g := gin.Default()
	//g.GET("/push/:acid/:content", func(c *gin.Context) {
	//	acid := c.Param("acid")
	//	content := c.Param("content")
	//
	//	if push.Push2Account(acid, content) {
	//		c.String(200, "ok")
	//	}
	//	c.String(200, "failed")
	//})

	go func() {
		for port := DebugPorts_start; port <= DebugPorts_end; port++ {
			url := fmt.Sprintf(":%d", port)
			logs.Info("debugtest try listen port %d ", port)
			mode := os.FileMode(0644)
			f, err := os.OpenFile("debugtest", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				logs.Error("debugtest start failed for can't open new file: %s", err)
				return
			}
			f.Write([]byte(url))
			f.Close()

			if err := g.Run(url); err != nil {
				logs.Warn("debugtest server try %s err %v", url, err)
			} else {
				break
			}
		}
	}()
}

func (p *Account) DebugExchangeWspvpRank(newRank int) {
	rank := ws_pvp.DeubgSetRank(p.GetWSPVPGroupId(), p.AccountID.String(), newRank)
	p.Profile.GetWSPVPInfo().Rank = rank
	ws_pvp.SavePlayerInfo(p.GetWSPVPGroupId(), p.convertWspvpInfo())
}

func (p *Account) DebugFinishWspvpRankTitle() {
	wspvpModule := ws_pvp.GetModule(p.AccountID.ShardId)
	topN := wspvpModule.GetTopN()
	wspvpModule.UpdateTitle(topN)
}

func (p *Account) DebugResetWspvpLockTime(beginTime, endTime string) {

}

func (p *Account) DebugMarketActivityReward(actType int64, actID int64) {
	market_activity.GetModule(p.AccountID.ShardId).DebugSnapShootAndReward(uint32(actType), uint32(actID))
	logs.Debug("DebugMarketActivityReward --")
}

func (p *Account) DebugMarketActivityClear(actType int64, actID int64) {
	market_activity.GetModule(p.AccountID.ShardId).DebugClearSnapShoot(uint32(actType), uint32(actID))
	logs.Debug("DebugMarketActivityClear --")
}

func (p *Account) DebugSetActivityRankWithRobot(actType int64, newRank int64, newScore int64) {
	var res *rank.RankByCorpGetRes
	si := p.GetSimpleInfo()
	res = market_activity.GetModule(p.AccountID.ShardId).GetRank(uint32(actType), p.AccountID.String(), &si)
	if res.Pos > int(newRank) {
		simpleInfo := p.GetSimpleInfo()
		switch actType {
		case gamedata.ActDestinyActivityRank:
			simpleInfo.DestinyLv = newScore
			rank.GetModule(p.AccountID.ShardId).RankByDestiny.Add(&simpleInfo)
		case gamedata.ActJadeActivityRank:
			simpleInfo.JadeLv = newScore
			rank.GetModule(p.AccountID.ShardId).RankByJade.Add(&simpleInfo)
		case gamedata.ActEquipStarLvActivityRank:
			simpleInfo.EquipStarLv = newScore
			rank.GetModule(p.AccountID.ShardId).RankByEquipStarLv.Add(&simpleInfo)
		case gamedata.ActHeroStarActivityRank:
			for i := 0; i < len(simpleInfo.AvatarStarLvl); i++ {
				simpleInfo.AvatarStarLvl[i] = 0
			}
			simpleInfo.AvatarStarLvl[0] = uint32(newScore)
			rank.GetModule(p.AccountID.ShardId).RankByHeroStar.Add(&simpleInfo)
		case gamedata.ActCorpLvActivityRank:
			simpleInfo.CorpLv = uint32(newScore)
			rank.GetModule(p.AccountID.ShardId).RankByCorpLv.Add(&simpleInfo)
		case gamedata.ActCorpGsActivityRank:
			simpleInfo.CurrCorpGs = int(newScore)
			_, oldScore := p.OnMaybeChangeMaxGS()
			rank.GetModule(p.AccountID.ShardId).RankCorpGs.Add(&simpleInfo, newScore, int64(oldScore))
		}

	} else if res.Pos < int(newRank) {
		simpleInfo := p.GetSimpleInfo()
		oldScore := res.Score + 1
		//robotC := int(newRank) - res.Pos
		logs.Debug("debug activity rank info: %d, %d", newRank, res.Pos)
		switch actType {
		case gamedata.ActDestinyActivityRank:
			simpleInfo.DestinyLv = oldScore
			for i := res.Pos; i < int(newRank); i++ {
				simpleInfo.AccountID = fmt.Sprintf("%d:%d:ROBOTROBOT%d", p.AccountID.GameId, p.AccountID.ShardId, i)
				simpleInfo.Name = "ROBOTROBOT"
				rank.GetModule(p.AccountID.ShardId).RankByDestiny.Add(&simpleInfo)
			}
		case gamedata.ActJadeActivityRank:
			simpleInfo.JadeLv = oldScore
			for i := res.Pos; i < int(newRank); i++ {
				simpleInfo.AccountID = fmt.Sprintf("%d:%d:ROBOTROBOT%d", p.AccountID.GameId, p.AccountID.ShardId, i)
				simpleInfo.Name = "ROBOTROBOT"
				rank.GetModule(p.AccountID.ShardId).RankByJade.Add(&simpleInfo)
			}
		case gamedata.ActEquipStarLvActivityRank:
			simpleInfo.EquipStarLv = oldScore
			for i := res.Pos; i < int(newRank); i++ {
				simpleInfo.AccountID = fmt.Sprintf("%d:%d:ROBOTROBOT%d", p.AccountID.GameId, p.AccountID.ShardId, i)
				simpleInfo.Name = "ROBOTROBOT"
				rank.GetModule(p.AccountID.ShardId).RankByEquipStarLv.Add(&simpleInfo)
			}

		case gamedata.ActHeroStarActivityRank:
			for i := 0; i < len(simpleInfo.AvatarStarLvl); i++ {
				simpleInfo.AvatarStarLvl[i] = 0
			}
			simpleInfo.AvatarStarLvl[0] = uint32(oldScore)
			for i := res.Pos; i < int(newRank); i++ {
				simpleInfo.AccountID = fmt.Sprintf("%d:%d:ROBOTROBOT%d", p.AccountID.GameId, p.AccountID.ShardId, i)
				simpleInfo.Name = "ROBOTROBOT"
				rank.GetModule(p.AccountID.ShardId).RankByHeroStar.Add(&simpleInfo)
			}
		case gamedata.ActCorpLvActivityRank:
			simpleInfo.CorpLv = uint32(oldScore)
			for i := res.Pos; i < int(newRank); i++ {
				simpleInfo.AccountID = fmt.Sprintf("%d:%d:ROBOTROBOT%d", p.AccountID.GameId, p.AccountID.ShardId, i)
				simpleInfo.Name = "ROBOTROBOT"
				rank.GetModule(p.AccountID.ShardId).RankByCorpLv.Add(&simpleInfo)
			}

		case gamedata.ActCorpGsActivityRank:
			simpleInfo.CurrCorpGs = int(oldScore)
			_, old := p.OnMaybeChangeMaxGS()
			for i := res.Pos; i < int(newRank); i++ {
				simpleInfo.AccountID = fmt.Sprintf("%d:%d:ROBOTROBOT%d", p.AccountID.GameId, p.AccountID.ShardId, i)
				simpleInfo.Name = "ROBOTROBOT"
				rank.GetModule(p.AccountID.ShardId).RankCorpGs.Add(&simpleInfo, oldScore, int64(old))
			}
		}
	}

}

func (p *Account) DebugCSRobClearCount() string {
	logs.Debug("DebugCSRobClearCount --")
	simpleInfo := p.Account.GetSimpleInfo()
	param := &csrob.PlayerParam{
		Acid:              p.AccountID.String(),
		GuildID:           p.GuildProfile.GuildUUID,
		Name:              simpleInfo.Name,
		GuildPosition:     simpleInfo.GuildPosition,
		Vip:               simpleInfo.Vip,
		FormationNew:      makeTodayFormation(p),
		FormationTeamFunc: p.buildHeroList,
	}
	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.PlayerWithNew(param)
	if nil == player {
		return "InitFailed"
	}
	if false == player.DebugClearCount() {
		return "Failed"
	}
	return "OK"
}

func (p *Account) DebugCSRobSetMaxCount() string {
	logs.Debug("DebugCSRobClearCount --")
	simpleInfo := p.Account.GetSimpleInfo()
	param := &csrob.PlayerParam{
		Acid:              p.AccountID.String(),
		GuildID:           p.GuildProfile.GuildUUID,
		Name:              simpleInfo.Name,
		GuildPosition:     simpleInfo.GuildPosition,
		Vip:               simpleInfo.Vip,
		FormationNew:      makeTodayFormation(p),
		FormationTeamFunc: p.buildHeroList,
	}
	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.PlayerWithNew(param)
	if nil == player {
		return "InitFailed"
	}

	vipCfg := gamedata.GetVIPCfg(int(p.Profile.Vip.V))
	count := csrob.PlayerCount{
		Build: vipCfg.CSRobBuildCarTimes,
		Help:  vipCfg.CSRobHelpLimit,
		Rob:   vipCfg.CSRobRobTimes,
	}
	if false == player.DebugSetMaxCount(count) {
		return "Failed"
	}
	return "OK"
}

func (p *Account) DebugCSRobBuildCar() string {
	logs.Debug("DebugCSRobClearCount --")
	simpleInfo := p.Account.GetSimpleInfo()
	param := &csrob.PlayerParam{
		Acid:              p.AccountID.String(),
		GuildID:           p.GuildProfile.GuildUUID,
		Name:              simpleInfo.Name,
		GuildPosition:     simpleInfo.GuildPosition,
		Vip:               simpleInfo.Vip,
		FormationNew:      makeTodayFormation(p),
		FormationTeamFunc: p.buildHeroList,
	}
	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.PlayerWithNew(param)
	if nil == player {
		return "InitFailed"
	}
	//当前阵容筹备
	formation := player.GetFormation()
	team := p.buildHeroList(formation)
	if nil == team || 0 == len(team) {
		return "Failed"
	}
	vipCfg := gamedata.GetVIPCfg(int(p.Profile.Vip.V))
	if _, err := player.BuildCar(team, int64(vipCfg.CSRobCarKeep*60)); nil != err {
		logs.Error("%v", err)
		return "Failed"
	}
	return "OK"
}

func (p *Account) DebugCSRobAddEnemy() string {
	logs.Debug("DebugCSRobClearCount --")
	simpleInfo := p.Account.GetSimpleInfo()
	param := &csrob.PlayerParam{
		Acid:              p.AccountID.String(),
		GuildID:           p.GuildProfile.GuildUUID,
		Name:              simpleInfo.Name,
		GuildPosition:     simpleInfo.GuildPosition,
		Vip:               simpleInfo.Vip,
		FormationNew:      makeTodayFormation(p),
		FormationTeamFunc: p.buildHeroList,
	}
	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.PlayerWithNew(param)
	if nil == player {
		return "InitFailed"
	}
	if false == player.DebugAddEnemy() {
		return "Failed"
	}
	return "OK"
}

func (p *Account) DebugCSRobDoWeekReward() string {
	title := p.Profile.GetTitle()
	title.CSRobTitleTime = 0
	csrob.GetModule(p.AccountID.ShardId).Ranker.DebugDoWeekReward()
	return "OK"
}

//DebugCSRobGuildList ..
func (p *Account) DebugCSRobGuildList() string {
	if !p.GuildProfile.InGuild() {
		return "fail"
	}
	guild := csrob.GetModule(p.AccountID.ShardId).GuildMod.GuildWithNew(p.Account.GuildProfile.GuildUUID, p.Account.GuildProfile.GuildName)
	if nil == guild {
		return "fail"
	}
	list := guild.GetEnemies()
	recommends := guild.GetList()

	//发debug邮件
	mail_sender.BatchSendMail2Account(p.AccountID.String(),
		timail.Mail_send_By_Debug,
		mail_sender.IDS_MAIL_GUILD_DECLINE_TITLE,
		[]string{
			fmt.Sprintf("Enemies:%+v\nRecommendList:%+v", list, recommends),
		},
		nil,
		"CSROB: DebugCSRobGuildList", false)

	logs.Debug("DebugCSRobGuildList")

	return "OK"
}

func (p *Account) DebugSetDestinyGeneralLv(lv int64, destID int64, rsp *ResponseDebugOp) {
	logs.Debug("Set DestinyGeneral lv: %d", lv)
	destiny := p.Profile.GetDestinyGeneral()
	for i := 0; i <= int(destID); i++ {
		destiny.AddNewGeneral(i)
	}
	dg := destiny.GetGeneral(int(destID))
	dg.LevelIndex = int(lv)
	dg.Exp = 0
	// 首次满级记录
	p._firstDest(int(destID))

	// MaxGS可能变化 10.神将
	p.Profile.GetData().SetNeedCheckMaxGS()

	// 更新神兽等级排行榜
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByDestiny.Add(&info)

	rsp.OnChangeDestinyGeneral()
	// 检查神兽任务是否完成
	rsp.OnChangeQuestAll()
	rsp.mkInfo(p)
}

//DebugAstrologyGetInfo ..
func (p *Account) DebugAstrologyGetInfo() string {
	req := new(reqMsgAstrologyGetInfo)
	r := servers.Request{}
	r.RawBytes = encode(req)

	p.AstrologyGetInfo(r)

	logs.Debug("DebugAstrologyGetInfo After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologyInto ..
func (p *Account) DebugAstrologyInto(hero, hole uint32, soul string) string {
	req := new(reqMsgAstrologyInto)
	req.HeroID = hero
	req.HoleID = hole
	req.SoulID = soul
	r := servers.Request{}
	r.RawBytes = encode(req)

	p.AstrologyInto(r)

	logs.Debug("DebugAstrologyInto After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologyDestroyInHero ..
func (p *Account) DebugAstrologyDestroyInHero(hero, hole uint32) string {
	req := new(reqMsgAstrologyDestroyInHero)
	req.HeroID = hero
	req.HoleID = hole
	r := servers.Request{}
	r.RawBytes = encode(req)

	p.AstrologyDestroyInHero(r)

	logs.Debug("DebugAstrologyDestroyInHero After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologyDestroyInBag ..
func (p *Account) DebugAstrologyDestroyInBag(soul string) string {
	req := new(reqMsgAstrologyDestroyInBag)
	req.SoulID = soul
	r := servers.Request{}
	r.RawBytes = encode(req)

	p.AstrologyDestroyInBag(r)

	logs.Debug("DebugAstrologyDestroyInBag After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologyDestroySkip ..
func (p *Account) DebugAstrologyDestroySkip(rares []uint32) string {
	req := new(reqMsgAstrologyDestroySkip)
	req.Rares = rares
	r := servers.Request{}
	r.RawBytes = encode(req)

	p.AstrologyDestroySkip(r)

	logs.Debug("DebugAstrologyDestroySkip After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologySoulUpgrade ..
func (p *Account) DebugAstrologySoulUpgrade(hero, hole uint32) string {
	req := new(reqMsgAstrologySoulUpgrade)
	req.HeroID = hero
	req.HoleID = hole
	r := servers.Request{}
	r.RawBytes = encode(req)

	p.AstrologySoulUpgrade(r)

	logs.Debug("DebugAstrologySoulUpgrade After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologyAugur ..
func (p *Account) DebugAstrologyAugur(skip bool) string {
	req := new(reqMsgAstrologyAugur)
	req.Skip = skip
	r := servers.Request{}
	r.RawBytes = encode(req)

	p.AstrologyAugur(r)

	logs.Debug("DebugAstrologyAugur After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologyFillBag ..
func (p *Account) DebugAstrologyFillBag() string {
	astrology := p.Profile.GetAstrology()
	bag := astrology.GetBag()
	ids := gamedata.GetAstrologyAllSoulIDs()
	for _, id := range ids {
		bag.AddSoul(id, 100)
	}

	logs.Debug("DebugAstrologyFillBag After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologyClearMyAstrologyData ..
func (p *Account) DebugAstrologyClearMyAstrologyData() string {
	astrology := p.Profile.GetAstrology()
	astrology.ClearData()

	logs.Debug("DebugAstrologyClearMyAstrologyData After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologyClearMyAstrologyBag ..
func (p *Account) DebugAstrologyClearMyAstrologyBag() string {
	astrology := p.Profile.GetAstrology()
	astrology.ClearBag()

	logs.Debug("DebugAstrologyClearMyAstrologyBag After, Astrology:%#v", p.Profile.GetAstrology())

	return "OK"
}

//DebugAstrologyTryAugurStatistic ..
func (p *Account) DebugAstrologyTryAugurStatistic(count int) string {
	astrology := p.Profile.GetAstrology()
	statis, goods := astrology.DebugAugurUpStatistic(count)

	str := "DebugAstrologyTryAugurStatistic:\n"
	str += "AugurLevel Statistic:\n"
	for l, c := range statis {
		str += fmt.Sprintf("%10d:%d\n", l, c)
	}
	str += "StarSoul Statistic:\n"
	for id, c := range goods {
		str += fmt.Sprintf("%20s:%d\n", id, c)
	}

	//发debug邮件
	mail_sender.BatchSendMail2Account(p.AccountID.String(),
		timail.Mail_send_By_Debug,
		mail_sender.IDS_MAIL_GUILD_DECLINE_TITLE,
		[]string{
			str,
		},
		nil,
		"DebugAstrologyClearMyAstrologyBag", false)

	logs.Debug("DebugAstrologyClearMyAstrologyBag, AugurLevel Statistic: %+v, StarSoul Statistic: %+v", statis, goods)

	return "OK"
}

func (p *Account) DebugClearGachaCount(index int) {
	if index >= gamedata.GachaMaxCount {
		logs.Error("exceed max gacha count for idx: %v", index)
		return
	}
	p.Profile.GetGacha(index).HeroGachaRaceCount = 0
}

func (p *Account) DebugSimulateFreeGacha(index int, times int, resp helper.ISyncRsp) {
	if times > 100 {
		logs.Error("too much times")
		return
	}
	items := make([]map[string]uint32, 0)
	for i := 0; i < times; i++ {
		corp_lv, _ := p.Profile.GetCorp().GetXpInfo()
		data := gamedata.GetGachaData(corp_lv, index)
		c := account.CostGroup{}
		if !c.AddCostData(p.Account, &data.CostForTenCoin) {
			logs.Error("Gacha CODE_Cost_Err")
			return
		}
		is_cost_ok, cost_hc_typ := c.CostWithHCBySync(p.Account, helper.HC_From_Buy, resp,
			helper.GachaTypeString(index, true))
		if !is_cost_ok {
			logs.Error("Gacha CODE_Cost_Err")
			return
		}
		gacha_state := p.Profile.GetGacha(index)
		id, count, _ := p.getGachaReward(data, gacha_state, resp, index, cost_hc_typ, i)
		item := make(map[string]uint32, 0)
		item[id] = count
		items = append(items, item)
	}
	go func(items []map[string]uint32) {
		for i, item := range items {
			mail_sender.BatchSendMail2Account(p.AccountID.String(),
				timail.Mail_Send_By_Sys, 0, []string{fmt.Sprintf("GachaTimes: %d", i+1), ""}, item, "debug", false)
			time.Sleep(time.Second)
		}
	}(items)

}

func (p *Account) DebugGetGachaCount(index int) {
	if index >= gamedata.GachaMaxCount {
		logs.Error("exceed max gacha count for idx: %v", index)
		return
	}
	mail_sender.BatchSendMail2Account(p.AccountID.String(),
		timail.Mail_Send_By_Sys, 0, []string{fmt.Sprintf("GachaCount: %d", p.Profile.GetGacha(index).HeroGachaRaceCount), ""}, nil, "debug",
		false)
}

// DebugGetLevelLimitLootSummary 返回指定关卡n次掉落的物品统计
func (p *Account) DebugGetLevelLimitLootSummary(levelId string, repeatTimes int) map[string]uint32 {
	itemCounts := make(map[string]uint32)

	for i := 0; i < repeatTimes; i++ {
		items := p.sendStageLimitReward(levelId, false)
		for _, item := range items.Datas {
			n := len(item.Item2Client)
			for j := 0; j < n; j++ {
				itemCounts[item.Item2Client[j]] += item.Count2Client[j]
			}
		}
	}

	return itemCounts
}

// DebugGetGachaReward 返回指定Gacha n次掉落的物品统计
func (p *Account) DebugGetGachaReward(gachaIdx int, repeatTimes int) map[string]uint32 {
	resp := &ResponseDebugOp{}
	itemCounts := make(map[string]uint32)

	for i := 0; i < repeatTimes; i++ {
		gachaData := gamedata.GetGachaData(p.Account.GetCorpLv(), gachaIdx)
		gachaState := p.Account.Profile.GetGacha(gachaIdx)
		item, count, _ := p.getGachaReward(gachaData, gachaState, resp, gachaIdx, 0, 1)
		itemCounts[item] += count
	}

	return itemCounts
}
