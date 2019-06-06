package logics

import (
	"encoding/json"
	"time"

	"fmt"

	"strings"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/logics/tmp"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/update/data_update"
	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	"vcs.taiyouxi.net/jws/gamex/modules/friend"
	"vcs.taiyouxi.net/jws/gamex/modules/hour_log"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/push"
	"vcs.taiyouxi.net/jws/gamex/sdk/samsung"
	"vcs.taiyouxi.net/jws/gamex/sdk/vivo"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/chat"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
	"vcs.taiyouxi.net/platform/x/api_gateway/util"
)

type Account struct {
	//当前玩家的Account
	*account.Account

	_lastRequest      string
	_lastRequestStart int64
}

func NewAccount(r *servers.Mux, dbaccount db.Account, ip string) *Account {
	acc, _ := account.NewAccount(dbaccount, ip)
	p := &Account{
		Account: acc,
	}

	//XXX: 这个顺序很重要p的hook中需要先设置Debug时间
	r.RegisterRequestHook(p)
	r.RegisterRequestHook(p.Account)

	p.RegisterSystemRequestHandler(r)
	p.RegisterPlayerMsgHandler(r)
	r.HandleFunc("PlayerAttr/OnConnectRequest", p.OnConnect)
	r.HandleFunc("PlayerAttr/ResetTimeStampRequest", p.ResetTimeStamp)
	r.HandleFunc("PlayLevel/GetLvlEnmyLootRequest", p.PrepareLootForLevelEnemy)
	r.HandleFunc("PlayLevel/DeclareLvlEnmyLootRequest", p.DeclareLootForLevelEnemy)
	r.HandleFunc("PlayLevel/DeclareLvlEnmyLootRequest_v2", p.DeclareLootForLevelEnemyV2)
	r.HandleFunc("PlayLevel/ChapterAwardRequest", p.ChapterAward)
	r.HandleFunc("PlayerAttr/ChangeEquipRequest", p.ChangeEquip)
	r.HandleFunc("PlayerAttr/ChangeAvatarEquipRequest", p.ChangeAvatarEquip)
	r.HandleFunc("PlayerAttr/GetEnergyRequest", p.GetEnergy)
	r.HandleFunc("PlayerBag/SellRequest", p.SellRequest)
	r.HandleFunc("PlayerAttr/EquipOpReq", p.EquipOp)
	r.HandleFunc("PlayerAttr/EquipMatEnhanceAddReq", p.EquipMatEnhanceAdd)
	r.HandleFunc("PlayerAttr/EquipMatEnhanceAutoAddReq", p.EquipMatEnhanceAutoAdd)
	r.HandleFunc("PlayerAttr/EquipMatEnhanceLvlUpReq", p.EquipMatEnhanceLvlUp)
	r.HandleFunc("PlayerAttr/GetAvatarExpRequest", p.GetAvatarExp)
	r.HandleFunc("PlayerBag/EquipResolveReq", p.EquipResolveRequest)
	r.HandleFunc("PlayerAttr/GetInfoRequest", p.GetInfo)
	r.HandleFunc("PlayerBag/ComposeRequest", p.Compose)

	r.HandleFunc("PlayerAttr/ReceiveQuestRequest", p.ReceiveQuestReq)
	r.HandleFunc("PlayerAttr/FinishQuestRequest", p.FinishQuestReq)
	r.HandleFunc("PlayerAttr/FinishManyQuestsRequest", p.FinishManyQuestsReq)

	r.HandleFunc("PlayerAttr/GetGiftMonthlyRequest", p.GetGiftMonthly)
	r.HandleFunc("PlayerAttr/GetGiftActivityRequest", p.GetGiftActivity)
	r.HandleFunc("PlayerAttr/ActiveGiftActivityRequest", p.ActiveGiftActivity)

	r.HandleFunc("PlayerAttr/BuyInStoreRequest", p.BuyInStore)
	r.HandleFunc("PlayerAttr/RefreshStoreRequest", p.RefreshStore)

	r.HandleFunc("PlayerAttr/GachaOneRequest", p.GachaOne)
	r.HandleFunc("PlayerAttr/GachaTenRequest", p.GachaTen)

	r.HandleFunc("PlayerAttr/ChangeAvatarRequest", p.ChangeAvatarRequest)
	r.HandleFunc("PlayerAttr/AvatarArousalAddRequest", p.AvatarArousalAdd)
	r.HandleFunc("PlayerAttr/AvatarSkillAddRequest", p.AddSkill)
	r.HandleFunc("PlayerAttr/SkillAddReq", p.AddSkillPractice)

	r.HandleFunc("PlayerAttr/GetMailsRequest", p.GetMails)
	r.HandleFunc("PlayerAttr/ReceiveMailsRequest", p.ReceiveMails)
	r.HandleFunc("PlayerAttr/ReadMailsRequest", p.ReadMails)

	r.HandleFunc("PlayLevel/StageSweepReq", p.StageSweep)

	r.HandleFunc("PlayLevel/GameModeLevelSweepReq", p.GameModeLevelSweep)

	r.HandleFunc("PlayerAttr/PlayerRankReq", p.GetRank)

	r.HandleFunc("PlayerAttr/BuyReq", p.Buy)

	r.HandleFunc("PlayerAttr/UseItemReq", p.UseItem)

	r.HandleFunc("PlayLevel/ResetGameModeCDRequest", p.ResetGameModeCD)

	r.HandleFunc("PlayerAttr/BossBeginReq", p.BossFightBegin)
	r.HandleFunc("PlayerAttr/BossEndReq", p.BossFightEnd)
	r.HandleFunc("PlayerAttr/BossSweepReq", p.BossSweep)

	r.HandleFunc("PlayerAttr/GetProtoVerReq", p.GetProtoVer)
	r.HandleFunc("PlayerAttr/PrivilegeBuyReq", p.PrivilegeBuy)
	r.HandleFunc("PlayerAttr/ChangeNameRequest", p.ChangeNameRequest)
	r.HandleFunc("PlayerAttr/RandNameRequest", p.RandNameRequest)
	r.HandleFunc("PlayerAttr/SaveNewHandRequest", p.SaveNewHandRequest)
	r.HandleFunc("PlayerAttr/SettingRequest", p.SettingRequest)

	r.HandleFunc("PlayerAttr/ReadStoryReq", p.ReadStory)
	r.HandleFunc("PlayerAttr/GetStoryReq", p.GetStory)

	r.HandleFunc("PlayerAttr/EquipAbstractReq", p.EquipAbstract)
	r.HandleFunc("PlayerAttr/EquipAbstractCancelReq", p.EquipAbstractCancel)
	r.HandleFunc("PlayerAttr/EquipTrickSwapReq", p.EquipTrickSwap)

	r.HandleFunc("PlayerAttr/GetSimplePvpEnemyReq", p.GetSimplePvpEnemy)
	r.HandleFunc("PlayerAttr/BeginSimplePvpReq", p.BeginSimplePvp)
	r.HandleFunc("PlayerAttr/EndSimplePvpReq", p.EndSimplePvp)
	r.HandleFunc("Attr/OpenSimplePvpDayChestReq", p.OpenSimplePvpDayChest)

	r.HandleFunc("PlayerAttr/ChatAuthReq", p.ChatAuthRequest)
	r.HandleFunc("PlayerAttr/RedeemCodeExchangeReq", p.redeemCodeExchange)
	r.HandleFunc("PlayerAttr/GetActivityGiftByCondReq", p.getActGiftByCondReward)
	r.HandleFunc("PlayerAttr/GetActivityGiftByTimeReq", p.getActGiftByTimeReward)

	// general
	r.HandleFunc("PlayerAttr/AddGeneralStarlLvlReq", p.addGeneralStarLevel)
	r.HandleFunc("PlayerAttr/LevelupGeneralRelReq", p.levelupGeneralRelation)
	r.HandleFunc("PlayerAttr/GeneralQuestRevReq", p.generalQuestReceive)
	r.HandleFunc("PlayerAttr/GeneralQuestRefReq", p.generalQuestRefresh)
	r.HandleFunc("PlayerAttr/GeneralQuestFinishReq", p.generalQuestFinish)

	r.HandleFunc("PlayerAttr/SetClientTagReq", p.SetClientTag)

	r.HandleFunc("Log/LogReturnTownReq", p.LogReturnTownRequest)
	r.HandleFunc("Log/LogEnterBossReq", p.LogEnterBossRequest)
	r.HandleFunc("Log/LogClientTimeEventReq", p.LogClientTimeEventRequest)

	// guild 公会
	r.HandleFunc("PlayerGuild/CreateGuildReq", p.CreateGuildRequest)
	r.HandleFunc("PlayerGuild/GetGuildListReq", p.GetRandomGuildListRequest)
	r.HandleFunc("PlayerGuild/GetGuildInfoReq", p.GetGuildInfoRequest)
	r.HandleFunc("PlayerGuild/FindGuildReq", p.FindGuildRequest)
	r.HandleFunc("PlayerGuild/ApplyGuildReq", p.ApplyGuildRequest)
	r.HandleFunc("PlayerGuild/CancelApplyGuildReq", p.CancelApplyGuildRequest)
	r.HandleFunc("PlayerGuild/GetPlayerGuildApplyListReq", p.GetPlayerGuildApplyList)
	r.HandleFunc("PlayerGuild/ApproveGuildApplicantReq", p.ApproveGuildApplicant)
	r.HandleFunc("PlayerGuild/GetGuildApplyListReq", p.GetGuildApplyList)
	r.HandleFunc("PlayerGuild/GuildQuitReq", p.GuildQuit)
	r.HandleFunc("PlayerGuild/GuildKickReq", p.GuildKick)
	r.HandleFunc("PlayerGuild/GuildPosAppointReq", p.GuildPositionAppoint)
	r.HandleFunc("PlayerGuild/GuildDismissReq", p.GuildDismiss)
	r.HandleFunc("PlayerGuild/ChangeGuildNoticeReq", p.ChangeGuildNotice)
	r.HandleFunc("PlayerGuild/PlayerGuildRankReq", p.GetGuildRank)
	r.HandleFunc("PlayerGuild/GuildSignReq", p.GuildSign)
	r.HandleFunc("PlayerGuild/SetGuildApplySettingReq", p.SetGuildApplySettingRequest)

	r.HandleFunc("PlayerAttr/UnlockAvatarReq", p.unlockAvatar)
	r.HandleFunc("PlayerAttr/IOSPayReq", p.iOSPayRequest)

	r.HandleFunc("PlayerAttr/RefreshFashionRequest", p.RefreshFashionReq)
	r.HandleFunc("PlayerAttr/GetSimplePvpRecordReq", p.GetSimplePvpRecord)
	r.HandleFunc("PlayerAttr/BuyFashionReq", p.BuyFashion)

	// shop
	r.HandleFunc("PlayerAttr/RefreshShopRequest", p.RefreshShop)
	r.HandleFunc("PlayerAttr/BuyInShopRequest", p.BuyInShop)

	// timing sync
	r.HandleFunc("PlayerAttr/TimingSyncgReq", p.TimingSync)

	r.HandleFunc("Attr/UnlockDestinyGlReq", p.UnlockDestinyGeneral)
	r.HandleFunc("Attr/AddDestinyGlLvReq", p.AddDestinyGeneralLv)
	r.HandleFunc("Attr/SetDestinyGlSkillReq", p.SetDestinyGeneralSkill)

	// jade
	r.HandleFunc("PlayerAttr/ChgJadeReq", p.ChgJade)
	r.HandleFunc("PlayerAttr/AutoAddJadeReq", p.AutoAddJade)
	r.HandleFunc("PlayerAttr/LvlUpJadeReq", p.LvlUpJade)
	r.HandleFunc("PlayerAttr/LvlUpJadeInBagReq", p.LvlUpJadeInBag)

	// 兵临城下
	r.HandleFunc("PlayerGuild/GuildVisitReq", p.guildVisit)
	//r.HandleFunc("PlayerGuild/GuildGatesEnemyStartReq", p.guildGatesEnemyStart)
	r.HandleFunc("PlayerGuild/GuildGatesEnemyFightBReq", p.guildGatesEnemyFightBegin)
	r.HandleFunc("PlayerGuild/GuildGatesEnemyFightBossBReq", p.guildGatesEnemyFightBossBegin)
	r.HandleFunc("PlayerGuild/GuildGatesEnemyFightEReq", p.guildGatesEnemyFightEnd)
	r.HandleFunc("PlayerGuild/GuildGatesEnemyFightBossEReq", p.guildGatesEnemyFightBossEnd)
	r.HandleFunc("PlayerGuild/GuildGatesEnemyGetRewardReq", p.guildGatesEnemyGetReward)
	r.HandleFunc("PlayerGuild/GuildGatesEnemyEnterActReq", p.guildGatesEnemyEnterAct)
	r.HandleFunc("PlayerGuild/GuildGatesEnemyLeaveActReq", p.guildGatesEnemyLeaveAct)
	r.HandleFunc("PlayerGuild/GetGuildLogReq", p.GetGuildLog)
	r.HandleFunc("Attr/GetGateEnemyRankInfoReq", p.GetGateEnemyRankInfo)

	// trial
	r.HandleFunc("PlayerAttr/TrialEnterLvlReq", p.TrialEnterLvl)
	r.HandleFunc("PlayerAttr/TrialFightReq", p.TrialFight)
	r.HandleFunc("PlayerAttr/TrialBonusAwardReq", p.TrialBonusAward)
	r.HandleFunc("PlayerAttr/TrialSweepStartReq", p.TrialSweepStart)
	r.HandleFunc("PlayerAttr/TrialSweepAwardForShowReq", p.TrialSweepAwardForShow)
	r.HandleFunc("PlayerAttr/TrialSweepAwardReq", p.TrialSweepAward)
	r.HandleFunc("PlayerAttr/TrialSweepEndByHCReq", p.TrialSweepEndByHC)
	r.HandleFunc("PlayerAttr/TrialResetReq", p.TrialReset)
	r.HandleFunc("Attr/TrialFastPassReq", p.TrialFastPass)

	// recover
	r.HandleFunc("PlayerAttr/RecoverAwardReq", p.recoverAward)

	// daily award
	r.HandleFunc("PlayerAttr/GetAllDailyAwardsInfoReq", p.GetAllDailyAwardsInfo)
	r.HandleFunc("PlayerAttr/AwardDailyAwardReq", p.AwardDailyAward)

	r.HandleFunc("Attr/GetOtherAccountReq", p.GetOtherAccountData)
	r.HandleFunc("Attr/TryGetPhoneCodeReq", p.TryGetPhoneCode)
	r.HandleFunc("Attr/UsePhoneCodeForRewardReq", p.UsePhoneCodeForReward)

	// city fish
	r.HandleFunc("PlayerAttr/CityFishReq", p.CityFish)
	r.HandleFunc("PlayerAttr/CityFishTenReq", p.CityFishTen)
	r.HandleFunc("PlayerAttr/GetPCityFishInfoReq", p.GetPlayerCityFishInfo)
	r.HandleFunc("PlayerAttr/GetGCityFishLogReq", p.GetGlobalCityFishLog)

	// Gank
	r.HandleFunc("PlayerAttr/GankTownFightBeginReq", p.gankTownFightBegin)
	r.HandleFunc("PlayerAttr/GankTownFightEndReq", p.gankTownFightEnd)
	r.HandleFunc("PlayerAttr/GankGetRecordReq", p.gankGetRecord)

	// account 7day
	r.HandleFunc("PlayerAttr/Account7DayBugGoodReq", p.Account7DayBugGood)

	r.HandleFunc("Attr/StartMatchGVEReq", p.StartMatchGVE)
	r.HandleFunc("Attr/GetGVEStateReq", p.GetGVEState)

	// teampvp
	r.HandleFunc("PlayerAttr/GetTeamPvpInfoReq", p.GetTeamPvpInfo)
	r.HandleFunc("PlayerAttr/SetTeamPvpAvatarReq", p.SetTeamPvpAvatar)
	r.HandleFunc("PlayerAttr/GetTeamPvpEnemyReq", p.GetTeamPvpEnemy)
	r.HandleFunc("PlayerAttr/RefreshTeamPvpEnemyReq", p.RefreshTeamPvpEnemy)
	r.HandleFunc("PlayerAttr/TeamPvpFightReq", p.TeamPvpFight)
	r.HandleFunc("Attr/TeamPvpOverFightReq", p.TeamPvpOverFight)
	r.HandleFunc("PlayerAttr/GetTeamPvpRecordReq", p.GetTeamPvpRecord)
	r.HandleFunc("Attr/OpenTeamPvpDayChestReq", p.OpenTeamPvpDayChest)
	r.HandleFunc("Attr/HeroStarUpReq", p.heroStarUp)

	r.HandleFunc("PlayerAttr/SyncDeviceTokenReq", p.SyncDeviceToken)
	// 城镇演员
	r.HandleFunc("Attr/GetFromGSRankReq", p.GetSomeoneFromGSRank)

	r.HandleFunc("PlayerAttr/GetFirstPayRewardReq", p.GetFirstPayReward)

	r.HandleFunc("PlayerAttr/HitEggReq", p.hitEgg)

	r.HandleFunc("PlayerAttr/GetIAPOrderIdReq", p.getIAPOrderId)
	r.HandleFunc("PlayerAttr/IApCardRewardReq", p.iapCardRewardRequest)

	r.HandleFunc("PlayerAttr/TitleTakeOnOffReq", p.TitleTakeOnOff)
	r.HandleFunc("PlayerAttr/TitleActivateReq", p.TitleActivate)
	r.HandleFunc("PlayerAttr/TitleClearHintReq", p.TitleClearHint)

	// grow fund
	r.HandleFunc("PlayerAttr/ActivateGrowFundReq", p.activateGrowFund)
	r.HandleFunc("PlayerAttr/AwardGrowFundReq", p.awardGrowFund)

	r.HandleFunc("Attr/WorshipReqMsg", p.worship)
	r.HandleFunc("Attr/FirstPassRewardReq", p.firstPassReward)

	// Share Wechat
	r.HandleFunc("Attr/GetShareWeChatRewardsReq", p.GetShareWeChatRewards)

	// MoneyCat
	r.HandleFunc("Attr/GetMoneyCatBaseInfoReq", p.GetMoneyCatBaseInfo)
	r.HandleFunc("Attr/BuyMoneyCatReq", p.BuyMoneyCat)

	// Expedition
	r.HandleFunc("Attr/ExpeditionInfoReq", p.ExpeditionInfo)
	r.HandleFunc("Attr/ExpeditionOverFightReq", p.ExpeditionOverFight)
	r.HandleFunc("Attr/ExpeditionrestReq", p.Expeditionrest)
	r.HandleFunc("Attr/ExpeditionsimpleinfoReq", p.Expeditionsimpleinfo)
	r.HandleFunc("Attr/ExpeditionAwardReq", p.ExpeditionAward)
	r.HandleFunc("Attr/ExpeditionHeroChooseReq", p.ExpeditionHeroChoose)
	r.HandleFunc("Attr/ExpeditionFightReq", p.ExpeditionFight)

	//神翼
	r.HandleFunc("Attr/ChangeHeroSwingReq", p.ChangeHeroSwing)
	r.HandleFunc("Attr/HeroSwingActReq", p.HeroSwingAct)
	r.HandleFunc("Attr/HeroSwingLvUpReq", p.HeroSwingLvUp)
	r.HandleFunc("Attr/HeroSwingRestReq", p.HeroSwingRest)
	r.HandleFunc("Attr/HeroSwingShowReq", p.HeroSwingShow)

	//羁绊(情缘)
	r.HandleFunc("Attr/ActivateCompanionReq", p.ActivateCompanion)
	r.HandleFunc("Attr/EvolveCompanionReq", p.EvolveCompanion)

	//限时礼包
	r.HandleFunc("Attr/BuyLimitGoodReq", p.BuyLimitGood)

	//节日Boss
	r.HandleFunc("Attr/FestivalBossStartReq", p.FestivalBossStart)
	r.HandleFunc("Attr/FestivalBossEndReq", p.FestivalBossEnd)
	r.HandleFunc("Attr/FestivalShopReq", p.FestivalShop)
	r.HandleFunc("Attr/FestivalBossInfoReq", p.FestivalBossInfo)

	// 公会红包
	r.HandleFunc("Attr/ClaimRedPacketRewardReq", p.ClaimRedPacketReward)
	r.HandleFunc("Attr/GetMyRedPacketLogReq", p.GetMyRedPacketLog)
	r.HandleFunc("Attr/GetRedPacketLogReq", p.GetRedPacketLog)

	// 7日红包
	r.HandleFunc("Attr/RedPacket7daysReq", p.RedPacket7days)
	// 名将体验
	r.HandleFunc("Attr/ExperienceLevelReq", p.ExperienceLevel)
	// 白盒宝箱
	r.HandleFunc("Attr/WhiteGachaShowInfoReq", p.WhiteGachaShowInfo)
	r.HandleFunc("Attr/WhiteGachaoneReq", p.WhiteGachaone)
	r.HandleFunc("Attr/WhiteGachaTenReq", p.WhiteGachaTen)

	//FaceBook 绑定
	r.HandleFunc("Attr/FBactivateReq", p.FBactivate)
	r.HandleFunc("Attr/FBInvitationReq", p.FBInvitation)
	r.HandleFunc("Attr/FBShareReq", p.FBShare)

	//确定进入城镇
	r.HandleFunc("Attr/IsEnterTownReq", p.IsEnterTown)

	handleAllGenFunc(r, p)

	//Remove debug in production,
	//TODO maybe some AccountID could debug
	if !game.Cfg.IsRunModeProd() {
		r.HandleFunc("Debug/SCOpRequest", p.SCDebugOp)
		r.HandleFunc("Debug/AvatarExpOpRequest", p.AvatarExpOp)
		r.HandleFunc("Debug/DebugOpRequest", p.DebugOp)
		r.HandleFunc("Debug/DebugAddGeneralNumRequest", p.DebugAddGeneralNum)
	}
	p.addHandle()
	p.InitData()
	p.Profile.MarketActivitys.RegHandler(p.AutoExchangeShopProp)

	// 数据更新不涉及结构变化
	// 此处处理数据更新, version最后的更新也是在这里做的
	err := data_update.Update(p.Profile.Ver, false, p.Account)
	if err != nil {
		logs.SentryLogicCritical(p.AccountID.String(),
			"data_update Err By %s",
			err.Error())
		// ???如果出错的话,就不要进了
		panic(err)
	}

	//TODO --- debug
	// p.debugTest()

	return p
}

//通过carbon-c-relay的聚合功能统计服务器上各种请求的平均响应时间

func (p *Account) graphiteRequestName(code, endfix string) string {
	lr := strings.Replace(code, "/", ".", -1)
	llr := strings.ToLower(lr)
	return fmt.Sprintf("requests.%d.%d.%s.%s", p.AccountID.GameId, p.AccountID.ShardId, llr, endfix)
}

func (p *Account) PreRequest(req servers.Request) {
	// for robot, set debugtime
	p._lastRequest = req.Code
	p._lastRequestStart = time.Now().UnixNano()
	name := p.graphiteRequestName(p._lastRequest, "count")
	metrics.SimpleSend(name, "1")

	if game.Cfg.DebugAccountTimeValid {
		var r Req
		cb := make([]byte, len(req.RawBytes))
		copy(cb, req.RawBytes)
		decode(cb, &r)
		if r.DebugTime > 0 {
			p.Account.Profile.DebugAbsoluteTime = r.DebugTime - time.Now().Unix()
		}
	}
}

func (p *Account) PostRequest(resp *servers.Response) {
	if p._lastRequest != "" {
		name := p.graphiteRequestName(p._lastRequest, "time")
		value := fmt.Sprintf("%d", (time.Now().UnixNano() - p._lastRequestStart))
		metrics.SimpleSend(name, value)
	}

	name := p.graphiteRequestName(resp.Code, "count")
	metrics.SimpleSend(name, "1")
}

type RequestResetTimeStamp struct {
	Req
}

type ResponseResetTimeStamp struct {
	Resp
	TimeStamp int64 `codec:"ts"`
}

func (p *Account) ResetTimeStamp(r servers.Request) *servers.Response {
	req := &RequestResetTimeStamp{}
	resp := &ResponseResetTimeStamp{}
	initReqRsp(
		"PlayerAttr/ResetTimeStampResponse",
		r.RawBytes,
		req, resp, p)

	resp.TimeStamp = time.Now().Unix()
	return rpcSuccess(resp)
}

// 体力相关
type RequestGetEnergy struct {
	Req
}

type ResponseGetEnergy struct {
	Resp
	Value       int64 `codec:"value"`
	RefershTime int64 `codec:"refersh_time"`
	LastTime    int64 `codec:"last_time"`
}

func (p *Account) GetEnergy(r servers.Request) *servers.Response {
	req := &RequestGetEnergy{}
	resp := &ResponseGetEnergy{}
	initReqRsp(
		"PlayerAttr/GetEnergyResponse",
		r.RawBytes,
		req, resp, p)

	resp.Value, resp.RefershTime, resp.LastTime = p.Profile.GetEnergy().Get()

	return rpcSuccess(resp)
}

type RequestBagSell struct {
	Req
	SellList  []uint32 `codec:"sells"`
	CountList []uint32 `codec:"counts"`
}

type ResponseBagSell struct {
	SyncResp
	NewSc int64 `codec:"new_sc"`
}

func (a *Account) SellRequest(r servers.Request) *servers.Response {
	req := &RequestBagSell{}
	resp := &ResponseBagSell{}

	initReqRsp(
		"PlayerBag/SellResponse",
		r.RawBytes,
		req, resp, a)

	var price_sum int64 = 0

	// 作弊检查
	for _, n := range req.CountList {
		if n > uutil.CHEAT_INT_MAX {
			return rpcErrorWithMsg(resp, 99, "SellRequest count cheat")
		}
	}

	for i, id := range req.SellList {
		if i >= len(req.CountList) {
			continue
		}
		count := req.CountList[i]
		code, sum := a.Sell(id, count, resp)

		if code != 0 {
			logs.SentryLogicCritical(a.AccountID.String(),
				"SellItemErr:%d,%d,%d", code, id, count)
			resp.MsgOK = "no"
			resp.SetCode(CODE_ERR, code)
			break
		}

		price_sum += sum
	}

	resp.NewSc = price_sum

	return rpcSuccess(resp)
}

func (a *Account) Sell(id uint32, count uint32, sync helper.ISyncRsp) (code uint32, sum int64) {
	const (
		_                    = iota
		CODE_To_Sell_No_Item //失败:要买的物品不是道具
		CODE_Item_Info_Err   //失败:服务器没找到道具对应的数据（ItemID不正确）
		CODE_No_Enough_Item  //失败:玩家没有足够的道具
	)

	idx := gamedata.ItemIdx_t(id)
	item_data := gamedata.GetItemDataByIdx(idx)

	code = 0
	sum = 0

	if count == 0 {
		return
	}

	if !bag.IsFixedID(id) {
		code = CODE_To_Sell_No_Item
		return
	}

	if item_data == nil {
		code = CODE_Item_Info_Err
		return
	}

	price := item_data.GetLootScore()

	c := account.CostGroup{}
	if c.AddItemByBagId(a.Account, id, count) && c.CostBySync(a.Account, sync, "Sell") {
		sum = int64(price) * int64(count)
		a.Profile.GetSC().AddSC(helper.SC_Money, sum, "Sell")
	} else {
		code = CODE_No_Enough_Item
	}

	return
}

type RequestBagEquipResolve struct {
	Req
	Equips []uint32 `codec:"equips"`
}

type ResponseBagEquipResolve struct {
	SyncRespWithRewards
}

func (a *Account) EquipResolveRequest(r servers.Request) *servers.Response {
	req := &RequestBagEquipResolve{}
	resp := &ResponseBagEquipResolve{}

	initReqRsp(
		"PlayerBag/EquipResolveRsp",
		r.RawBytes,
		req, resp, a)

	need_update := true
	for _, id := range req.Equips {
		code := a.EquipResolve(id, resp, "EquipResolve")
		if code != 0 {
			logs.Warn("EquipResolveErr:%d,%d %s", code, id, a.AccountID.String())
			resp.MsgOK = "no"
			resp.Code = code
			need_update = false
			break
		}
	}

	if need_update {
		// 表示是否需要全量更新道具和软通信息到客户端，装备穿脱会引发产出
		logs.Trace("Need To Update Client")
		resp.OnChangeSC()
		resp.mkInfo(a)
	}

	return rpcSuccess(resp)

}

func (a *Account) EquipResolve(id uint32, sync interfaces.ISyncRspWithRewards, reason string) uint32 {
	const (
		_                        = iota
		CODE_To_Resolve_No_Equip //失败:要买的物品不是装备
		CODE_Equip_Info_Err      //失败:服务器没找到装备对应的数据（ItemID不正确）
		CODE_No_Resolve_data     //失败:服务器没找到装备对应的熔炼数据（ItemID不正确）
		CODE_No_Equip            //失败:玩家没有这件装备
		CODE_Equip_Cannot_Cost   //失败:玩家装备已锁定
		CODE_Give_Err            //失败:赠送时失败
	)

	// 不是装备
	if bag.IsFixedID(id) {
		return mkCode(CODE_ERR, CODE_To_Resolve_No_Equip)
	}

	// 压根没有装备
	is_has := a.BagProfile.IsHasBagId(id)
	if !is_has {
		return mkCode(CODE_WARN, errCode.ClickTooQuickly)
	}

	// 装备正在用着呢
	if a.Profile.GetEquips().IsHasEquip(id) {
		logs.Warn("EquipResolve CODE_Equip_Cannot_Cost")
		return mkCode(CODE_WARN, errCode.ClickTooQuickly)
	}

	// 装备数据找不到
	item_data, data_ok := a.BagProfile.GetItemData(id)
	if !data_ok {
		return mkCode(CODE_ERR, CODE_Equip_Info_Err)
	}

	// 熔炼数据找不到
	tier := item_data.GetTier()
	rare := item_data.GetRareLevel()
	cost_data := gamedata.GetEquipResolveGive(tier, rare)
	if cost_data == nil {
		return mkCode(CODE_ERR, CODE_No_Resolve_data)
	}

	cost := account.CostGroup{}
	if cost.AddItemByBagId(a.Account, id, 1) && cost.CostBySync(a.Account, sync, reason) {
		gg := account.GiveGroup{}
		gg.AddCostData(cost_data)
		is_success := gg.GiveBySyncAuto(a.Account, sync, reason)
		if is_success {
			logs.Trace("[%s]EquipResolve:%d", a.AccountID, id)
			return 0
		} else {
			return mkCode(CODE_ERR, CODE_Give_Err)
		}
	}

	return mkCode(CODE_WARN, errCode.ClickTooQuickly)
}

type RequestProtoVer struct {
	Req
	Info string `codec:"info"`
}
type ResponseProtoVer struct {
	Resp
	ServerVer            string `codec:"serverver"`
	ProtoDataVer         string `codec:"protodataver"`
	GroupId              string `codec:"groupId,omitempty"`
	ProtoDataVerNew      string `codec:"protodatavernew,omitempty"`
	LoginTimes           int64  `codec:"logintimes"`
	ChatAddr             string `codec:"chataddr"`
	HotData              string `codec:"hotdata"`
	CreateTime           int64  `codec:"createtime"`
	AntiCheatGsThreshold int64  `codec:"anticheatgs"`
	Lang                 string `codec:"lang"`
	TimeLocal            string `codec:"timelocal"`
}

func (a *Account) GetProtoVer(r servers.Request) *servers.Response {
	req := &RequestProtoVer{}
	resp := &ResponseProtoVer{}

	initReqRsp(
		"PlayerAttr/GetProtoVerResp",
		r.RawBytes,
		req, resp, a)

	acid := a.AccountID.String()
	/* req.Info=
	{"MarketVer":"Invalid",
	"Build":"0",
	"BuildHash":"",
	"BuildBranch":"",
	"Data":"0",
	"DataHash":"",
	"DataBranch":"",
	"DeviceInfo":"MacBookPro12,1",
	"NetInfo":"NetNone",
	"S":["ddd", "ddd", "ddd"]   }
	*/
	a.Profile.LastGetProtoTS = time.Now().UnixNano()

	a.Tmp.Last_GetProto_TS = time.Now().Unix()

	info := account.ClientInfo{}
	if err := json.Unmarshal([]byte(req.Info), &info); err != nil {
		return rpcErrorWithMsg(resp, 1, fmt.Sprintf("unmarshal Info err %s", err.Error()))
	}

	a.Profile.ClientInfo = info

	a.Profile.AccountName = info.AccountName
	a.Profile.DeviceId = info.DeviceId
	a.Profile.MemSize = info.MemSize
	a.Profile.ChannelQuickId = info.ChannelId
	// 获取到渠道ID，更新运营活动, 放到这里很糟糕
	mka := a.Profile.GetMarketActivitys()
	oldChannel := mka.ChannelID
	mka.ChannelID = a.Profile.ChannelQuickId
	if oldChannel == "" {
		a.Profile.GetMarketActivitys().ForceUpdateMarketActivity(a.AccountID.String(), a.Profile.GetProfileNowTime())
	}

	a.Profile.ChannelId = gamedata.GetChannelId(info.ChannelId)
	a.Profile.PlatformId = info.PlatformId
	a.Profile.DeviceToken = info.DeviceToken
	a.Profile.IDFA = info.IDFA
	a.LeaveGVG()
	a.Profile.SetClientMarketVer(info.MarketVer)

	resp.ServerVer = version.GetVersion()
	resp.ProtoDataVer = gamedata.GetProtoDataVer()
	resp.ProtoDataVerNew = gamedata.GetGameDataConfVer()
	resp.LoginTimes = a.Profile.LoginTimes
	resp.ChatAddr = chat.Cfg.ChatAddr
	hotBuild := gamedata.GetHotDataVerCfg().Build
	if hotBuild > 0 && gamedata.HotDataValid {
		resp.HotData = fmt.Sprintf("%s %d",
			game.Cfg.HotDataVerC_Suc, hotBuild)
	}
	isReg := false
	if a.Profile.LoginTimes == 1 {
		isReg = true
	}
	resp.CreateTime = a.Profile.CreateTime
	anticheatGsCfg := game.Cfg.AntiCheat[account.CheckerHeroGS]
	resp.AntiCheatGsThreshold = anticheatGsCfg.ParamInt
	resp.Lang = game.Cfg.Lang
	resp.TimeLocal = game.Cfg.TimeLocal
	// group
	//resp.GroupId = a.Profile.PlayerGroup

	// bi
	loginType := "login"
	if info.ConnectReason == "token" {
		loginType = "relogin"
	}
	logiclog.LogLogin(acid, info.AccountName, a.Profile.GetCurrAvatar(),
		a.Profile.GetHC().GetHC(), info.DeviceId, info.MemSize, a.Profile.Name, a.Account.GetIp(),
		info.BundleUpdate, info.DataUpdate, isReg,
		info.MarketVer, info.DeviceInfo, info.PhoneNum,
		logiclog.LogicInfo_ProfileInfo{ChannelId: a.Profile.ChannelId, Group: resp.GroupId}, info.IDFA,
		a.Profile.LoginTimes, loginType, a.Profile.GetCorp().GetLvlInfo(),
		func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
	if a.Profile.LoginTimes == 1 {
		hour_log.Get(a.AccountID.ShardId).OnRegister(a.AccountID.String(),
			a.Profile.ChannelId, a.Profile.DeviceId)
	}

	// 存device信息到dynamo
	if info.PlatformId == push.Platform_ios || info.PlatformId == push.Platform_android {
		logs.Info("%s DevicePlatformIdToken %s DeviceToken %s DeviceInfo %s",
			acid, info.PlatformId, info.DeviceToken, info.DeviceInfo)
		if info.DeviceToken == "0" || info.DeviceToken == "null" {
			logs.Warn("DeviceToken=0 %s %s %s %v", acid, a.Profile.Name,
				info.PlatformId, info.DeviceInfo)
		}
		// 不用下面代码了，直接放到profile里
		//if info.DeviceToken != "" && info.DeviceInfo != "" {
		//	account_info.AccountInfoDynamo.SetAccountInfoData(acid, info.PlatformId, info.DeviceToken, info.DeviceInfo)
		//}
	}
	a._notifyUserInfo()

	return rpcSuccess(resp)
}

type RequestOnConnect struct {
	Req
	DeviceId    string `codec:"devId"`
	AccountName string `codec:"acnm"`
	ChannelId   string `codec:"chan"`
	PlatformId  string `codec:"plfm"`
}

type ResponseOnConnect struct {
	Resp
}

// 为记录大数据用
func (p *Account) OnConnect(r servers.Request) *servers.Response {
	req := &RequestOnConnect{}
	resp := &ResponseOnConnect{}
	initReqRsp(
		"PlayerAttr/OnConnectResponse",
		r.RawBytes,
		req, resp, p)

	channel := gamedata.GetChannelId(req.ChannelId)
	p.Profile.ChannelId = channel
	hour_log.AddCCU(channel)
	hour_log.Get(p.AccountID.ShardId).OnLogin(p.AccountID.String(),
		channel, req.DeviceId)
	return rpcSuccess(resp)
}

type profileInfo struct {
	ChannelId string `json:"ChannelId"`
	Group     string `json:"Group"`
}

// 通用的信息获取接口 一些经常需要刷新且不大的（数值型）信息通过这个一次获取
type RequestGetInfo struct {
	Req
	NeedBag               bool `codec:"need_bag"`
	NeedSc                bool `codec:"need_sc"`
	NeedHc                bool `codec:"need_hc"`
	NeedAvatarExps        bool `codec:"need_avatar_exps"`
	NeedCorp              bool `codec:"need_corp"`
	NeedEnergy            bool `codec:"need_energy"`
	NeedBossFightPoint    bool `codec:"need_boss_fp"`
	NeedStageAll          bool `codec:"need_stage_all"`
	NeedChapterAll        bool `codec:"need_chapter_all"`
	NeedEquip             bool `codec:"need_equip"`
	NeedAvatarEquip       bool `codec:"need_avatar_equip"`
	NeedQuestAll          bool `codec:"need_quest"`
	NeedGift              bool `codec:"need_gift"`
	NeedMonthlyGift       bool `codec:"need_m_gift"`
	NeedStoreAll          bool `codec:"need_store_all"`
	NeedGachaAll          bool `codec:"need_gacha_all"`
	NeedSkillAll          bool `codec:"need_skill_all"`
	NeedGeneralAll        bool `codec:"need_general_all"`
	NeedGeneralRelAll     bool `codec:"need_general_rel_all"`
	NeedGeneralQuestAll   bool `codec:"need_general_q_all"`
	NeedGeneralTeamAll    bool `codec:"need_general_t_all"`
	NeedMailAll           bool `codec:"need_mail_all"`
	NeedBuyAll            bool `codec:"need_buy_all"`
	NeedBossAll           bool `codec:"need_boss_all"`
	NeedGameModeAll       bool `codec:"need_gm_all"`
	NeedPrivilegeBuy      bool `codec:"need_pvlgby_all"`
	NeedNewHandStep       bool `codec:"need_new_hand"`
	NeedPvp               bool `codec:"need_pvp"`
	NeedActGiftByCond     bool `codec:"need_act_gift_cond_all"`
	NeedActGiftByTime     bool `codec:"need_act_gift_time_all"`
	NeedIAPGoodInfo       bool `codec:"need_iap_good_info_all"`
	NeedPlayerGuildInfo   bool `codec:"need_player_guild_info"`
	NeedGuildInventory    bool `codec:"need_guild_inventory"`
	NeedGuildBoss         bool `codec:"need_guild_Boss"`
	NeedClientTagInfo     bool `codec:"need_client_tag_info"`
	NeedShopsAll          bool `codec:"need_shop_all"`
	NeedJadeAll           bool `codec:"need_jade_all"`
	NeedAvatarJadeAll     bool `codec:"need_av_jade_all"`
	NeedDestinyGeneral    bool `codec:"need_destiny"`
	NeedDestinyGenJadeAll bool `codec:"need_dg_jade_all"`
	NeedTrialAll          bool `codec:"need_trail_all"`
	NeedGatesEnemyAll     bool `codec:"need_gates_enemy_all"`
	NeedRecoverAll        bool `codec:"need_recover_all"`
	NeedFashionBagAll     bool `codec:"need_fashion_bag_all"`
	NeedPhone             bool `codec:"need_phone"`
	NeedSevenDayRank      bool `codec:"need_seven_day_rank"`
	NeedAccount7Day       bool `codec:"need_account_7day"`
	NeedDailyAward        bool `codec:"need_daily_award"`
	NeedTeamPvp           bool `codec:"need_team_pvp"`
	NeedHero              bool `codec:"need_hero"`
	NeedHeroTalent        bool `codec:"need_hero_talent"`
	NeedHeroSoul          bool `codec:"need_hero_soul"`
	NeedFirstPayReward    bool `codec:"need_first_pay"`
	NeedHitEgg            bool `codec:"need_hit_egg"`
	NeedTitle             bool `codec:"need_title"`
	NeedGrowFund          bool `codec:"need_grow_fund"`
	NeedFirstPassReward   bool `codec:"need_first_pass_reward"`
	NeedMarketActivity    bool `codec:"need_market_activity"`
	NeedWantGeneralInfo   bool `codec:"need_want_general_info"`
	NeedHeroTeam          bool `codec:"need_hero_team"`
	NeedHeroGachaRace     bool `codec:"need_hero_gacha_race"`
	NeedGVGInfo           bool `codec:"need_gvg_info"`
	NeedUpdateFriendList  bool `codec:"need_update_friend_list"`
	NeedUpdateBlackList   bool `codec:"need_update_black_list"`
	NeedGuildRedPacket    bool `codec:"need_guild_red_pakcet"`
	NeedFestivalShop      bool `codec:"need_festival_shop"`
	NeedExclusiveWeapon   bool `codec:"need_exclusive_weapon"`
	NeedGuildWorship      bool `codec:"need_guild_worship"`
	NeedHeroDiff          bool `codec:"need_hero_diff"`
	NeedRedPacket7Days    bool `codec:"need_red_packet_days"`
	NeedExperienceLevel   bool `codec:"need_experience_level"`
	NeedWhiteGacha        bool `codec:"need_white_gacha"`
	NeedWSPVP             bool `codec:"need_wspvp"`
	NeedFaceBook          bool `codec:"need_facebook"`
	NeedOppoRelated       bool `codec:"need_oppo_related"`
	NeedHeroDestiny       bool `codec:"need_hero_destiny"`
	NeedExpedition        bool `codec:"need_expedition"`
	NeedHeroStarMap       bool `codec:"need_hero_star_map"`
	NeedBindMailReward    bool `codec:"need_bind_mail_reward"`
	NeedMagicPetInfo      bool `codec:"need_mp_i"`
	NeedBattleArmyInfo    bool `codec:"need_ba_i"`
	NeedTwitterShare      bool `codec:"need_twitter"`
	NeedLineShare         bool `codec:"need_line"`
}

type ResponseGetInfo struct {
	SyncResp
}

func (p *Account) GetInfo(r servers.Request) *servers.Response {
	req := &RequestGetInfo{}
	resp := &ResponseGetInfo{}

	initReqRsp(
		"PlayerAttr/GetInfoResponse",
		r.RawBytes,
		req, resp, p)

	p.Tmp.Last_GetInfo_TS = time.Now().Unix()

	if req.NeedBag {
		resp.OnChangeBag()
	}

	if req.NeedSc {
		resp.OnChangeSC()
	}

	if req.NeedHc {
		resp.OnChangeHC()
	}

	if req.NeedAvatarExps {
		resp.OnChangeAvatarExp()
	}

	if req.NeedCorp {
		resp.OnChangeCorpExp()
	}

	if req.NeedEnergy {
		resp.OnChangeEnergy()
	}

	if req.NeedBossFightPoint {
		resp.OnChangeBossFightPoint()
	}

	if req.NeedStageAll {
		resp.OnChangeStageAll()
	}

	if req.NeedChapterAll {
		resp.OnChangeChapterAll()
	}

	if req.NeedEquip {
		resp.OnChangeEquip()
	}

	if req.NeedAvatarEquip {
		resp.OnChangeAvatarEquip()
	}

	if req.NeedQuestAll {
		resp.OnChangeQuestAll()
	}

	if req.NeedGift {
		resp.OnChangeGiftStateChange()
	}

	if req.NeedMonthlyGift {
		resp.OnChangeMonthlyGiftStateChange()
	}

	if req.NeedStoreAll {
		resp.OnChangeStoreAllChange()
	}

	if req.NeedGachaAll {
		resp.OnChangeGachaAllChange()
	}

	resp.OnChangeAvatarArousal()

	if req.NeedSkillAll {
		resp.OnChangeSkillAllChange()
	}

	if req.NeedGeneralAll {
		resp.OnChangeGeneralAllChange()
	}

	if req.NeedGeneralRelAll {
		resp.OnChangeGeneralRelAllChange()
	}

	if req.NeedGeneralQuestAll {
		resp.OnChangeGeneralQuest()
	}

	if req.NeedGeneralTeamAll {
		resp.OnChangeGeneralTeamAllChange()
	}

	if req.NeedMailAll {
		resp.OnChangeMail()
	}

	resp.OnChangeVIP()

	if req.NeedBuyAll {
		resp.OnChangeBuy()
	}

	if req.NeedBossAll {
		resp.OnChangeBoss()
	}

	if req.NeedGameModeAll {
		resp.OnChangeAllGameMode()
	}

	if req.NeedPrivilegeBuy {
		resp.OnChangePrivilegeBuy()
	}

	if req.NeedNewHandStep {
		resp.OnChangeNewHand()
	}

	if req.NeedPvp {
		resp.OnChangeSimplePvp()
	}

	if req.NeedActGiftByCond {
		resp.OnChangeActivityByCond()
	}

	if req.NeedActGiftByTime {
		resp.OnChangeActivityByTime()
	}

	if req.NeedIAPGoodInfo {
		resp.OnChangeIAPGoodInfo()
	}

	if req.NeedPlayerGuildInfo {
		resp.OnChangePlayerGuild()
		resp.OnChangePlayerGuildApply()
		resp.OnChangeGuildInfo()
		resp.OnChangeGuildMemsInfo()
		resp.OnChangeGuildScience()
	}

	if req.NeedGuildInventory {
		resp.OnChangeGuildInventory()
	}

	if req.NeedClientTagInfo {
		resp.OnChangeClientTagInfo()
	}

	if req.NeedShopsAll {
		resp.OnChangeShopAllChange()
	}

	if req.NeedDestinyGeneral {
		resp.OnChangeDestinyGeneral()
	}

	if req.NeedJadeAll {
		resp.OnChangeJadeFull()
	}
	if req.NeedAvatarJadeAll {
		resp.OnChangeAvatarJade()
	}
	if req.NeedDestinyGenJadeAll {
		resp.OnChangeDestinyGenJade()
	}

	if req.NeedTrialAll {
		resp.OnChangeTrial()
	}

	if req.NeedGatesEnemyAll {
		resp.OnChangeGatesEnemyData()
		resp.OnChangeGatesEnemyPushData()
	}

	if req.NeedRecoverAll {
		resp.OnChangeRecover()
	}

	if req.NeedFashionBagAll {
		resp.OnChangeFashionBag()
	}

	if req.NeedPhone {
		resp.OnChangePhone()
	}

	if req.NeedSevenDayRank {
		resp.OnChangeSevenDayRank()
	}

	if req.NeedAccount7Day {
		resp.OnChangeAccount7Day()
	}

	if req.NeedDailyAward {
		resp.OnChangeDailyAward()
	}

	if req.NeedTeamPvp {
		resp.OnChangeTeamPvp()
	}

	if req.NeedHero {
		p.Profile.GetHero().SetNeedSync()
	}

	if req.NeedHeroTalent {
		resp.OnChangeHeroTalent()
	}

	if req.NeedHeroSoul {
		resp.OnChangeHeroSoul()
	}

	if req.NeedFirstPayReward {
		resp.OnChangeFirstPayReward()
	}

	if req.NeedHitEgg {
		resp.OnChangeHitEgg()
	}

	if req.NeedTitle {
		resp.OnChangeTitle()
	}

	if req.NeedGrowFund {
		resp.OnChangeGrowFund()
	}

	if req.NeedFirstPassReward {
		resp.OnChangeFirstPassRewardInfo()
	}

	if req.NeedMarketActivity {
		resp.OnChangeMarketActivity()
	}

	if req.NeedWantGeneralInfo {
		resp.OnChangeWantGeneralInfo()
	}

	if req.NeedHeroTeam {
		resp.OnChangeHeroTeam()
	}

	if req.NeedHeroGachaRace {
		resp.OnChangeHeroGachaRace()
	}
	if req.NeedUpdateFriendList {
		resp.OnChangeFriendList()
	}
	if req.NeedUpdateBlackList {
		resp.OnChangeBlackList()
	}
	if req.NeedGuildRedPacket {
		resp.OnChangeGuildRedPacket()
	}

	if req.NeedFestivalShop {
		resp.onChangeFestivalBossInfo()
	}

	if req.NeedGuildBoss {
		resp.SetNeedSyncGuildActBoss()
	}

	if req.NeedExclusiveWeapon {
		resp.onChangeExclusiveWeaponInfo()
	}
	if req.NeedHeroDiff {
		resp.OnChangeHeroDiff()
	}

	if req.NeedGuildWorship {
		resp.OnChangeGuildWorshipInfo()
	}

	if req.NeedRedPacket7Days {
		resp.OnChangeRedPacket7Days()
	}

	if req.NeedExperienceLevel {
		resp.OnChangeExperienceLevel()
	}
	if req.NeedWhiteGacha {
		resp.OnChangeWhiteGacha()
	}
	if req.NeedWSPVP {
		resp.OnChangeWSPVP()
	}

	if req.NeedFaceBook {
		resp.OnChangeFaceBook()
	}

	if req.NeedTwitterShare {
		resp.OnChangeTwitter()
	}

	if req.NeedLineShare {
		resp.OnChangeLine()
	}

	if req.NeedOppoRelated {
		resp.OnChangeOppoRelated()
	}

	if req.NeedHeroDestiny {
		resp.OnChangeHeroDestiny()
	}
	if req.NeedExpedition {
		resp.OnChangerExpeditionInfo()
	}

	if req.NeedHeroStarMap {
		resp.OnChangeHeroStarMap()
	}
	if req.NeedBindMailReward {
		resp.OnChangeBindMailReward()
	}
	if req.NeedMagicPetInfo {
		resp.OnChangeMagicPetInfo()
	}
	if req.NeedBattleArmyInfo {
		resp.OnChangeBattleArmyInfo()
	}

	resp.OnChangeBaseData()
	resp.OnChangeUnlockAvatar()

	resp.OnChangeShareWeChat()
	resp.OnChangeHeroSwing()
	if req.NeedGVGInfo {
		resp.OnChangeGVG()
	}
	resp.OnChangeCompanion()
	resp.onChangeLimitShop()
	resp.onChangeMoneyCat()
	resp.onChangeFestivalBossInfo()
	resp.OnChangeGuildWorshipInfo()
	resp.OnChangeNewHandIgnoreNeed()
	resp.OnChangeOfflineRecover()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

type RequestChangeAvatar struct {
	Req
	AvatarId int `codec:"a"`
}

type ResponseChangeAvatar struct {
	Resp
}

func (a *Account) ChangeAvatarRequest(r servers.Request) *servers.Response {
	req := &RequestChangeAvatar{}
	resp := &ResponseChangeAvatar{}

	initReqRsp(
		"PlayerAttr/ChangeAvatarResponse",
		r.RawBytes,
		req, resp, a)

	if a.IsAvatarUnblock(req.AvatarId) {
		a.Profile.CurrAvatar = req.AvatarId
		a.GetHandle().OnAvatarChg(req.AvatarId)
		// 更新排行榜
		simpleInfo := a.GetSimpleInfo()
		lv, _ := a.Profile.GetCorp().GetXpInfo()
		if lv >= FirstIntoCorpLevel {
			rank.GetModule(a.AccountID.ShardId).RankCorpGs.Add(&simpleInfo,
				int64(simpleInfo.CurrCorpGs), int64(simpleInfo.CurrCorpGs))
		}
	}

	return rpcSuccess(resp)
}

type RequestChangeName struct {
	Req
	Name string `codec:"name"`
}

type ResponseChangeName struct {
	SyncResp
}

type PackageReq struct {
	PkgId    int64 `codec:"package_id"`
	SubPkgId int64 `codec:"sub_package_id"`
}

func (a *Account) ChangeNameRequest(r servers.Request) *servers.Response {
	const (
		_ = iota
		CODE_ERR_Save
		CODE_ERR_Name_Err
		CODE_ERR_Name_Len_Err
	)

	acid := a.AccountID.String()
	req := &RequestChangeName{}
	resp := &ResponseChangeName{}

	initReqRsp(
		"PlayerAttr/ChangeNameResponse",
		r.RawBytes,
		req, resp, a)

	if req.Name == "" {
		return rpcError(resp, CODE_ERR_Name_Err)
	}

	if len(req.Name) > 64 {
		return rpcErrorWithMsg(resp, CODE_ERR_Name_Len_Err, fmt.Sprintf("CODE_ERR_Name_Len_Err size %v", len(req.Name)))
	}

	// 检查特殊字符，敏感词
	if gamedata.CheckSymbol(req.Name) || gamedata.CheckSensitive(req.Name) {
		return rpcWarn(resp, errCode.RenameSensitve)
	}

	logs.Trace("[%s]ChangeName %s From %s", acid, req.Name, a.Profile.Name)

	if a.Profile.Name == "" {
		err := driver.ChangeNameToRedis(req.Name, acid, a.AccountID.ShardId)

		if err == driver.ChangeNameErrByNameExist {
			return rpcWarn(resp, errCode.RenameNameHasExit)
		}
		if err != nil {
			logs.SentryLogicCritical(acid, "ChangeNameRequest Err by %s", err.Error())
			return rpcError(resp, CODE_ERR_Save)
		}

		a.Profile.Name = req.Name
		logiclog.LogCreateRole(a.AccountID.String(), a.Profile.CurrAvatar,
			a.Profile.Name, a.GetCorpLv(), a.Profile.ChannelId,
			func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")

	} else {
		warnCode, errCode := a.rename(req.Name, resp)
		if errCode != 0 {
			return rpcError(resp, errCode)
		}
		if warnCode != 0 {
			return rpcWarn(resp, warnCode)
		}
	}

	// 新名字更新好友系统cache
	simpleInfo := a.GetSimpleInfo()
	friend.GetModule(a.AccountID.ShardId).UpdateFriendInfo(&simpleInfo, 0)

	//各个排行榜更新，异步
	rank.GetModule(a.AccountID.ShardId).RankByEquipStarLv.UpdateIfInTopN(&simpleInfo)
	rank.GetModule(a.AccountID.ShardId).RankByJade.UpdateIfInTopN(&simpleInfo)
	rank.GetModule(a.AccountID.ShardId).RankByDestiny.UpdateIfInTopN(&simpleInfo)
	rank.GetModule(a.AccountID.ShardId).RankByCorpLv.UpdateIfInTopN(&simpleInfo)

	rank.GetModule(a.AccountID.ShardId).RankByWingStar.UpdateIfInTopN(&simpleInfo)
	rank.GetModule(a.AccountID.ShardId).RankByHeroDestiny.UpdateIfInTopN(&simpleInfo)
	rank.GetModule(a.AccountID.ShardId).RankByHeroWuShuangGs.UpdateIfInTopN(&simpleInfo)
	rank.GetModule(a.AccountID.ShardId).RankByAstrology.UpdateIfInTopN(&simpleInfo)
	rank.GetModule(a.AccountID.ShardId).RankByExclusiveWeapon.UpdateIfInTopN(&simpleInfo)
	rank.GetModule(a.AccountID.ShardId).RankByHeroJadeTwo.UpdateIfInTopN(&simpleInfo)

	//其他需要更新名字的模块
	csrob.GetModule(a.AccountID.ShardId).PlayerMod.Rename(a.AccountID.String(), a.GuildProfile.GuildUUID, req.Name)
	if gamedata.Guild_Pos_Chief == a.GuildProfile.GuildPosition {
		csrob.GetModule(a.AccountID.ShardId).GuildMod.MasterRename(a.GuildProfile.GuildUUID, req.Name)
	}

	resp.OnChangeBaseData()
	resp.mkInfo(a)
	return rpcSuccess(resp)
}

func (a *Account) _notifyUserInfo() {
	uid := a.AccountID.UserId.String()
	var canGot bool
	var f1, f2, money float64
	if !a.Profile.GotPayFeedBack {
		canGot, f1, f2, money = game.GetPayFeedBackByUid(uid)
	}
	if !canGot { // 正常服务器,只通知auth有角色,采用异步方式
		a._notifyUserInfoToAuth()
	} else { // 有反钻状态的服务器,需要等待auth查询,决定是否要发邮件,采用同步方式
		tmp.PayFeedBack(a.Account, f1, f2, money)
	}
}

func (a *Account) _notifyUserInfoToAuth() {
	go func() {
		req := httplib.Post(game.Cfg.AuthNotifyNewRoleUrl)
		req.Param("gid", fmt.Sprintf("%d", a.AccountID.GameId))
		req.Param("shardid", fmt.Sprintf("%d", a.AccountID.ShardId))
		req.Param("uid", a.AccountID.UserId.String())
		req.Param("newrole", fmt.Sprintf("%d", a.Profile.GetCorp().GetLvlInfo()))

		ret := struct {
			Res            string
			HadGotFeedBack bool
		}{}
		_s, err := req.String()
		if err != nil {
			logs.Error("[AuthNotifyNewRole] resp string failed err with %v %s",
				err, _s)
			return
		}

		err = json.Unmarshal([]byte(_s), &ret)
		if err != nil {
			logs.Error("[AuthNotifyNewRole] json.Unmarshal failed err with %v %s",
				err, _s)
			return
		}

		rsp, _ := req.Response()
		defer rsp.Body.Close()

		if ret.Res != "ok" {
			logs.Error("[AuthNotifyNewRole] res failed with %v", ret)
			return
		}
		logs.Trace("[AuthNotifyNewRole] success %s", a.AccountID.String())
	}()
}

func (a *Account) onExit() {
	a._notifyUserInfoToAuth()
}

type RequestRandName struct {
	Req
}

type ResponseRandName struct {
	Resp
	Names []string `codec:"names"`
}

func (a *Account) RandNameRequest(r servers.Request) *servers.Response {
	req := &RequestRandName{}
	resp := &ResponseRandName{}

	initReqRsp(
		"PlayerAttr/RandNameResponse",
		r.RawBytes,
		req, resp, a)

	const randNameCount = 10
	resp.Names = gamedata.RandNames(randNameCount, a.Account.Profile.SystemLanguage)

	return rpcSuccess(resp)
}

type RequestSetting struct {
	Req
	SysLang string `codec:"lang"`
}

type ResponseSetting struct {
	Resp
}

func (a *Account) SettingRequest(r servers.Request) *servers.Response {
	req := &RequestSetting{}
	resp := &ResponseSetting{}

	initReqRsp(
		"PlayerAttr/SettingResponse",
		r.RawBytes,
		req, resp, a)

	a.Account.Profile.SystemLanguage = req.SysLang
	return rpcSuccess(resp)
}

type RequestSaveNewHand struct {
	Req
	Step string `codec:"stp"`
}

type ResponseSaveNewHand struct {
	Resp
}

func (a *Account) SaveNewHandRequest(r servers.Request) *servers.Response {
	req := &RequestSaveNewHand{}
	resp := &ResponseSaveNewHand{}

	initReqRsp(
		"PlayerAttr/SaveNewHandResponse",
		r.RawBytes,
		req, resp, a)

	err := a.Profile.SetNewHand(req.Step)
	if err != nil {
		logs.SentryLogicCritical(a.AccountID.String(), err.Error())
		return rpcError(resp, 1)
	}

	// logiclog
	logiclog.LogTutorial(a.AccountID.String(), a.Profile.GetCurrAvatar(),
		a.Profile.GetCorp().GetLvlInfo(), a.Profile.ChannelId, req.Step,
		func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}

type RequestChatAuth struct {
	Req
}

type ResponseChatAuth struct {
	Resp
	AuthKey string `codec:"auth"`
}

func (a *Account) ChatAuthRequest(r servers.Request) *servers.Response {
	req := &RequestChatAuth{}
	resp := &ResponseChatAuth{}

	initReqRsp(
		"PlayerAttr/ChatAuthRsp",
		r.RawBytes,
		req, resp, a)

	acid := a.AccountID.String()
	reqChatAuth := httplib.Post(game.Cfg.ChatAuthUrl).SetTimeout(8*time.Second, 8*time.Second)
	reqChatAuth.Param("acid", acid)

	chatAuth, err := reqChatAuth.String()
	if err != nil {
		logs.SentryLogicCritical(acid, "ChatAuthRequest Err by %s", err.Error())
		return rpcWarn(resp, errCode.ChatEnterWarn)
	}

	resp.AuthKey = chatAuth

	rsp, _ := reqChatAuth.Response()
	defer rsp.Body.Close()

	return rpcSuccess(resp)
}

func (p *Account) SyncDeviceToken(r servers.Request) *servers.Response {
	req := &struct {
		Req
		PlatformId  string `codec:"PlatformId"`
		DeviceToken string `codec:"DeviceToken"`
		DeviceInfo  string `codec:"DeviceInfo"`
	}{}
	resp := &struct {
		Resp
	}{}

	initReqRsp(
		"PlayerAttr/SyncDeviceTokenResp",
		r.RawBytes,
		req, resp, p)

	acid := p.AccountID.String()
	switch req.PlatformId {
	case push.Platform_ios:
		fallthrough
	case push.Platform_android:
		logs.Trace("%s DevicePlatformIdToken %s DeviceToken %s DeviceInfo %s",
			acid, req.PlatformId, req.DeviceToken, req.DeviceInfo)
		p.Profile.PlatformId = req.PlatformId
		if req.DeviceToken != "" {
			p.Profile.DeviceToken = req.DeviceToken
		}

	}
	//if req.PlatformId == push.Platform_ios || req.PlatformId == push.Platform_android {
	//	logs.Trace("%s DevicePlatformIdToken %s DeviceToken %s DeviceInfo %s",
	//		acid, req.PlatformId, req.DeviceToken, req.DeviceInfo)
	//	if req.DeviceToken != "" && req.DeviceInfo != "" {
	//		account_info.AccountInfoDynamo.SetAccountInfoData(acid, req.PlatformId, req.DeviceToken, req.DeviceInfo)
	//	}
	//}

	return rpcSuccess(resp)
}

func (p *Account) GetFirstPayReward(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Id uint32 `codec:"idx"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/GetFirstPayRewardRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_No_Cfg
		Err_Already_Get
		Err_Give
		Err_Hc_Not_Enough
	)

	cfg := gamedata.GetFirstPayConfig(req.Id)
	if cfg == nil {
		return rpcErrorWithMsg(resp, Err_No_Cfg, "Err_No_Cfg")
	}

	if p.Profile.FirstPayReward.HadGot(req.Id) {
		logs.Warn("GetFirstPayReward Err_Already_Get")
		return rpcWarn(resp, errCode.ClickTooQuickly)

	}

	if uint32(p.Profile.GetHC().BuyFromHc) < cfg.GetHCBuyCount() {
		return rpcErrorWithMsg(resp, Err_Hc_Not_Enough, "Err_Already_Get")
	}

	// 给奖励
	data := &gamedata.CostData{}
	for _, award := range cfg.GetPayAward() {
		data.AddItem(award.GetReward(), award.GetCount())
	}
	give := &account.GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(p.Account, resp, "GetFirstPayReward") {
		return rpcErrorWithMsg(resp, Err_Give, "Err_Give")
	}

	// 记录
	p.Profile.FirstPayReward.GotReward(req.Id)

	resp.OnChangeFirstPayReward()
	resp.mkInfo(p)

	// sysnotice
	sysnotice.NewSysRollNotice(p.AccountID.ServerString(), gamedata.IDS_PAY_JIANGLI).
		AddParam(sysnotice.ParamType_RollName, p.Profile.Name).Send()
	return rpcSuccess(resp)
}

// 获取支付订单号
func (p *Account) getIAPOrderId(r servers.Request) *servers.Response {
	req := &struct {
		Req
		IAPIndex uint32 `codec:"iap_idx"`
		ExtrInfo string `codec:"extr_info"`
	}{}
	resp := &struct {
		Resp
		IAPOrderId  string   `codec:"iap_order"`
		ParamNames  []string `codec:"pns"`
		ParamValues []string `codec:"pvs"`
		OrderTitle  string   `codec:"ord_tl"`
		OrderDesc   string   `codec:"ord_desc"`
	}{}

	initReqRsp(
		"PlayerAttr/GetIAPOrderIdRsp",
		r.RawBytes,
		req, resp, p)

	acid := p.AccountID.String()
	cfg := gamedata.GetIAPInfo(req.IAPIndex)
	if nil == cfg {
		return rpcErrorWithMsg(resp, 1, "IAPIndex not found in config")
	}
	p.Profile.GetIAPGoodInfo().IAPOrderSeed++
	resp.IAPOrderId = fmt.Sprintf("%s:%d:%d", acid,
		p.Profile.GetIAPGoodInfo().IAPOrderSeed, req.IAPIndex)

	logiclog.LogIAPTry(p.AccountID.String(), p.Profile.AccountName, p.Profile.Name,
		p.Profile.CurrAvatar, p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		req.IAPIndex, resp.IAPOrderId,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	// 准备参数
	iapBaseCfg := gamedata.GetIAPBaseConfig(cfg.Info.GetIapID())
	if iapBaseCfg == nil { // 说明是ios
		return rpcSuccess(resp)
	}
	title := iapBaseCfg.GetAppleOfficialName()
	desc := iapBaseCfg.GetAppleOfficialDesc()
	if title == "" {
		title = "empty"
		logs.Error("getIAPOrderId iap title empty %s", cfg.Info.GetIapID())
	}
	if desc == "" {
		desc = "empty"
		logs.Error("getIAPOrderId iap desc empty %s", cfg.Info.GetIapID())
	}

	// 各个渠道
	var res map[string]string
	if p.Profile.ChannelQuickId == util.VivoChannel {
		res = vivo.TryPay(resp.IAPOrderId, cfg.Android_Rmb_Price, req.ExtrInfo,
			title, title)
	} else if p.Profile.ChannelQuickId == util.SamsungChannel {
		res = samsung.TryPay(acid, req.IAPIndex, resp.IAPOrderId, cfg.Android_Rmb_Price,
			req.ExtrInfo, title, title)
	}
	// 处理结果
	if res != nil {
		resp.ParamNames = make([]string, 0, len(res))
		resp.ParamValues = make([]string, 0, len(res))
		for k, v := range res {
			resp.ParamNames = append(resp.ParamNames, k)
			resp.ParamValues = append(resp.ParamValues, v)
		}
		resp.OrderTitle = title
		resp.OrderDesc = desc
	}
	return rpcSuccess(resp)
}

func (p *Account) GetWSPVPGroupId() int {
	return int(gamedata.GetWSPVPGroupId(uint32(p.AccountID.ShardId)))
}

