package logics

import (
	"fmt"

	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/logics/notify"
	"vcs.taiyouxi.net/jws/gamex/logics/sync"
	"vcs.taiyouxi.net/jws/gamex/models/buy"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/dest_gen_first"
	"vcs.taiyouxi.net/jws/gamex/modules/hero_diff"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/modules/want_gen_best"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

type StageInfo2Client struct {
	Id           string `codec:"id"`
	TodayCount   int32  `codec:"count"`
	RefreshCount int32  `codec:"refresh"`
	SumCount     int32  `codec:"sumcount"`
	Star         int32  `codec:"star"`
}

type activityGiftInfo2Client struct {
	// 签到ID、进度和当前状态（0 为当天未领取过 1 为已领取基本奖励 2 为已领取VIP奖励）、下一次签到时间
	GiftId              uint32   `codec:"id"`
	GiftActivityTime    int64    `codec:"act_t"`
	GiftCurrIdx         int      `codec:"idx"`
	GiftCurrStat        int      `codec:"stat"`
	GiftNextUpdateSec   int64    `codec:"nxupsec"`
	GiftAllRewardStat   []int    `codec:"ga_ss"`
	GiftAllRewardCount  []int    `codec:"ga_cs"`
	GiftAllRewardAID    []string `codec:"ga_aid"`
	GiftAllRewardACount []uint32 `codec:"ga_ac"`
	GiftAllRewardAData  []string `codec:"ga_ad"`
	// TODO delete
	GiftRewardID    []string `codec:"rid"`
	GiftRewardCount []uint32 `codec:"rc"`
	GiftRewardData  []string `codec:"rd"`
}

type guildPlayerInfoToClient struct {
	GuildUUID     string `codec:"uuid"`
	NextEnterTime int64  `codec:"lev_t"`
}

type guildPlayerApplyInfoToClient struct {
	ApplyGuild_UUID   []string `codec:"aply_g_uid"`
	ApplyGuild_Name   []string `codec:"aply_g_nm"`
	ApplyGuild_Lvl    []uint32 `codec:"aply_g_lvl"`
	ApplyGuild_Notice []string `codec:"aply_g_ntc"`
	ApplyGuild_Time   []int64  `codec:"aply_g_t"`
}

type guildBasicInfoToClient struct {
	GuildUUID                  string `codec:"uuid"`
	GuildID                    string `codec:"id"`
	Name                       string `codec:"na"`
	Level                      uint32 `codec:"lvl"`
	Exp                        int64  `codec:"lvexp"`
	NextExp                    int64  `codec:"mlvexp"`
	ApplyGsLimit               int    `codec:"aply_gs"`
	ApplyAuto                  bool   `codec:"aply_auto"`
	SignCount                  int    `codec:"signc"`
	Icon                       string `codec:"icon"`
	Rank                       int    `codec:"rank"`
	GuildGSSum                 int64  `codec:"gssum"`
	MemNum                     uint32 `codec:"mem_num"`
	MaxMem                     uint32 `codec:"max_mem"`
	Notice                     string `codec:"notc"`
	WeekRank                   int    `codec:"weekrank"`
	GatesEnemyCount            int    `codec:"gecount"`
	GatesEnemyMemCount         int    `codec:"gememc"`
	GatesEnemyMemRewardedCount int    `codec:"gememrc"`
	RenameTimes                int    `codec:"renamet"`
	GuildTmpVer                int    `codec:"guildtmpver"`
}

type guildMemToClient struct {
	Acid                    string `codec:"aid"`
	Name                    string `codec:"na"`
	GuildContribution       int64  `codec:"act_t"`
	GuildSp                 int64  `codec:"sp"`
	GuildSignLastTs         int64  `codec:"signlastt"`
	Level                   uint32 `codec:"lvl"`
	Gs                      int    `codec:"gs"`
	Vip                     uint32 `codec:"vip"`
	CurrAvatar              int    `codec:"cur_avatar"`
	Position                int    `codec:"pos"`
	LastLoginTime           int64  `codec:"l_li"`
	IsGettedGateEnemyReward bool   `codec:"isger"`
	GuildBossDamage         int    `codec:"lastBossDamage"`
	Online                  bool   `codec:"online"`
	GVGScore                int    `codec:"GVGScore"`
}

type SyncResp struct {
	SyncRespNotify

	syncData.SyncPhone

	// common 每次sync都带的信息
	SyncCommonNeed bool `codec:"comn_b_"`

	// 数据版本号，大于0有效，每个协议带回
	DataVer int32 `codec:"date_ver_"`
	// bundle版本号，大于0有效，每个协议带回
	BundleVer int32 `codec:"bundle_ver_"`
	// 服务器shardid，主要合服用，没合服的就是本身的id
	ShardId uint `codec:"shard_"`
	// 热更版本号，大于0有效，每个协议带回
	HotDataVer int `codec:"hot_ver_"`
	// 数据版本号，大于0有效，每个协议带回
	DataMin   int32 `codec:"data_min_"`
	BundleMin int32 `codec:"bundle_min_"`

	// 活动开启标志
	ActValid []bool `codec:"act_v_"`

	SyncCurrAvatar int `codec:"ca_"`

	Name                 string `codec:"name_"`
	SyncRenameCount      int    `codec:"rename_count"`
	SyncGuildChangeToday bool   `codec:"guild_ct"`

	// 背包更新
	SyncFullItemsNeed   bool     `codec:"items_full_n"`
	Items               []byte   `codec:"items_full_"`
	SyncUpdateItemsNeed bool     `codec:"items_up_n"`
	ItemsUpdate         []byte   `codec:"items_up_"`
	SyncDelItemsNeed    bool     `codec:"items_del_n"`
	ItemsDel            []uint32 `codec:"items_del_"`

	// 软通
	SyncSC []int64 `codec:"sc_"`
	SyncHC int64   `codec:"hc_"`
	// bundle版本号，大于0有效，每个协议带回
	SyncHCBuy      int64 `codec:"hc_buy"` // 购买对应hc的数量，不会减少
	SyncHCBuyToday int64 `codec:"hc_buy_today_"`

	// 经验
	SyncExps []uint32 `codec:"exps_"`

	// 觉醒
	SyncAvatarArousal []uint32 `codec:"arousal_"`

	// 技能等级
	SyncAvatarSkillLevel    []uint32 `codec:"skill_"`
	SyncAvatarSkillLevelLen int      `codec:"skill_l_"`
	SyncSkillPracticeLv     []uint32 `codec:"skill_p_"`

	// 战队等级
	SyncCorpLv uint32 `codec:"clv_"`
	SyncCorpXp uint32 `codec:"cxp_"`

	// 体力信息
	SyncValue       int64 `codec:"evalue_"`
	SyncRefershTime int64 `codec:"erefersh_"`
	SyncLastTime    int64 `codec:"elast_"`

	// 世界boss fight point
	SyncBossFightPointValue       int64 `codec:"bfp_value_"`
	SyncBossFightPointRefershTime int64 `codec:"bfp_refersh_"`
	SyncBossFightPointLastTime    int64 `codec:"bfp_last_"`

	// 关卡
	SyncStage     [][]byte `codec:"stages_"`
	SyncLastStage string   `codec:"lstage_"`
	SyncChapter   [][]byte `codec:"chapters_"`

	// 当前装备
	CurrEquips     [][]byte `codec:"cequips_"`
	SlotMax        int      `codec:"slot_len_"`
	EquipMatEnh    []bool   `codec:"eq_me_"`
	EquipMatEnhMax int      `codec:"eq_me_max_"`

	CurrAvatarEquips  [][]byte `codec:"caequips_"`
	AvatarSlotMax     int      `codec:"avatar_slot_len_"`
	StarHcUpLastCount int      `codec:"starhcc_"`

	// 龙玉
	SyncFullJadeNeed   bool     `codec:"jade_f_n_"`
	SyncFullJade       [][]byte `codec:"jade_f_"`
	SyncUpdateJadeNeed bool     `codec:"jade_u_n_"`
	SyncUpdateJade     [][]byte `codec:"jade_u_"`
	SyncDelJadeNeed    bool     `codec:"jade_d_n_"`
	SyncDelJade        []uint32 `codec:"jade_d_"`

	AvatarIdJade []int    `codec:"aid_jade_"`
	AvatarJade   [][]byte `codec:"a_jade_"`
	DGIdJade     []int    `codec:"dgid_jade_"`
	DGJade       [][]byte `codec:"dg_jade_"`
	JadeSlotMax  int      `codec:"jade_slot_len_"`

	// 任务
	// 当前所有可接任务
	SyncQuestCanReceiveAll [][]byte `codec:"q_can_rec_all_"`
	SyncQuestCaneceiveNeed bool     `codec:"q_can_rec_need_"`
	// 当前所有已接任务(未领取的)
	SyncQuestReceivedAll [][]byte `codec:"q_reced_all_"`
	SyncQuestReceivedAdd [][]byte `codec:"q_reced_add_"`
	// 当前所有完成的每日任务
	SyncQuestDailyClosed      [][]byte `codec:"q_daily_clo_"`
	SyncQuestPoint            int      `codec:"quest_point_"`
	SyncAccount7DayQuestPoint int      `codec:"quest_7day_point_"`

	// 月签到ID、进度和当前状态 0 为当天未领取过 1 为已领取基本奖励 2 为已领取VIP奖励 3 为已补签
	SyncGiftMonthlyNeed    bool   `codec:"m_gift_need_"`
	SyncGiftMonthlyId      uint32 `codec:"m_gift_id_"`
	SyncGiftMonthlyCurrIdx int    `codec:"m_gift_idx_"`
	// 0 为当天未领取过 1 为已领取基本奖励 2 为已领取VIP奖励 3 为已补签
	SyncGiftMonthlyCurrStat int `codec:"m_gift_stat_"`
	SyncGiftLeftReSignCount int `codec:"m_gift_resign_c"`

	SyncGiftsNeed bool     `codec:"gifts_need_"`
	SyncGifts     [][]byte `codec:"gifts_"`

	// 玩家商店信息，注意不止一个商店 []byte是store
	SyncStore [][]byte `codec:"stores_"`

	syncStoreDelta     map[uint32]syncStoreElem `codec:"-"`
	syncStoreDeltaNeed bool
	// SyncStoreNextRefershCostTyp []string   `codec:"stores_uct_"`
	// SyncStoreNextRefershCost    []int64    `codec:"stores_uc_"`
	// SyncStoreNextRefershTime    []int64    `codec:"stores_ut_"`
	// SyncStoreRefershTime        []int      `codec:"stores_rc_"`
	// SyncStoreRefershTimeLast    []int      `codec:"stores_rcl_"`

	SyncGacha [][]byte `codec:"gacha_"`

	// 玩家副将信息
	SyncGeneral          bool     `codec:"general_"`
	SyncGeneralName      []string `codec:"general_n_"`
	SyncGeneralStarLevel []uint32 `codec:"general_sl_"`
	SyncGeneralNum       []uint32 `codec:"general_num_"`

	// 玩家副将羁绊
	SyncGeneralRel      bool     `codec:"general_r_b_"`
	SyncGeneralRelId    []string `codec:"general_r_"`
	SyncGeneralRelLevel []uint32 `codec:"general_r_l_"`

	// 玩家副将派遣
	SyncGeneralTeamNeed bool     `codec:"general_t_n_"`
	SyncGeneralTeam     [][]byte `codec:"general_t_"`

	// 玩家副将任务
	SyncGeneralQuestNeed bool     `codec:"general_q_n_"`
	QuestListId          []int64  `codec:"qlId_"`
	QuestListName        []string `codec:"qln_"`
	QuestListReved       []bool   `codec:"qlr_"`
	QuestListNextRefTime int64    `codec:"qlnrt_"`
	QuestRevId           []int64  `codec:"qrn_id_"`
	QuestRevName         []string `codec:"qrn_"`
	QuestRevFinishTime   []int64  `codec:"qrft_"`
	QuestRevGeneralNum   []int    `codec:"qrgn_"`
	QuestRevGenerals     []string `codec:"qrgs_"`

	// 玩家副将任务信息
	SyncGQuest     [][]byte `codec:"gqs_"`
	SyncGQuestNeed bool     `codec:"gqs_n_"`

	// 邮件系统信息
	SyncMail     [][]byte `codec:"mail_"`
	SyncMailNeed bool     `codec:"mail_need_"`

	// VIP信息
	SyncVIPLv       uint32 `codec:"vip_"`
	SyncVIPRMBPoint uint32 `codec:"rmb_"`

	// 各种购买信息
	SyncBuy           [][]byte `codec:"buy_"`
	SyncBuyStageTimes []byte   `codec:"buy_stage_times_"`

	// 名将乱入
	SyncBossIDs          []string `codec:"bossids_"`
	SyncBossRewardIDs    []string `codec:"bossrids_"`
	SyncBossRewardCounts []uint32 `codec:"bossrcs_"`
	SyncBossRewardsCount int      `codec:"bossrc_"`
	SyncBossCount        int      `codec:"bossc_"`
	SyncBossCountRefTime int64    `codec:"bossc_rt_"`
	SyncBossMaxDegree    int      `codec:"bossm_"`

	// 活动gamemode
	SyncGameModeInfo [][]byte `codec:"gm_"`
	SyncGameModeNeed bool     `codec:"gm_n_"`

	// privilege buy
	SyncPrivilegeBuyNeed bool     `codec:"pvlby_n"`
	SyncPrivilegeBuy     [][]byte `codec:"pvlby_"`

	// new hand
	SyncNewHandNeed       bool   `codec:"newhad_n"`
	SyncNewHand           string `codec:"newhad"`
	SyncNewHandIgnoreNeed bool   `codec:"newhadign_n"`
	SyncNewHandIgnore     bool   `codec:"newhadign"`

	// pvp次数
	SyncPvpCount            int      `codec:"pvp_n_"`
	SyncPvpCountNextRefTime int64    `codec:"pvp_rt_"`
	SyncPvpDefAvatarId      int      `codec:"pvp_def_"`
	SyncPvpCountToday       int      `codec:"pvp_d_c"`
	SyncPvpOpendChests      []uint32 `codec:"pvp_chest"`

	// 活动条件奖励
	SyncActGiftByCondCount int      `codec:"actgs_n_"`
	SyncActGiftByCond      [][]byte `codec:"actgs_"`

	SyncActGiftByTimeCount int    `codec:"actg_t_n_"`
	SyncActGiftByTime      []byte `codec:"actg_t_"`

	// iap good info
	SyncIAPGoodInfoNeed      bool     `codec:"iap_g_n_"`
	SyncIAPGoodInfo          [][]byte `codec:"iap_g_"`
	SyncCurrenSbatch         int64    `codec:"fsc"`
	SyncMonthlyEndTime       int64    `codec:"month_card_end_t"`
	SyncMonthlyValidTime     int64    `codec:"month_card_valid_t"`
	SyncIsLifeCard           bool     `codec:"is_life_card"`
	SyncLifeCardValidTime    int64    `codec:"life_card_valid_t"`
	SyncWeekRewardEndTime    int64    `codec:"week_card_end_t"`
	SyncWeekRewardValidTime  int64    `codec:"week_card_valid_t"`
	SyncLevelGiftId          string   `codec:"lvl_gift_gf_id"`
	SyncLevelGiftEndTime     int64    `codec:"lvl_gift_gf_et"`
	SyncLevelGiftIdWaitAward []string `codec:"lvl_gift_gf_w_a"`

	/*
		// guild
		SyncPlayerGuild     bool   `codec:"guild_p_n_"`
		SyncPlayerGuildInfo []byte `codec:"guild_p_i_"`

		SyncGuildInfoNeed bool   `codec:"guild_i_n_"`
		SyncGuildInfo     []byte `codec:"guild_i_"`
		SyncPost          string `codec:"post_"`

		SyncGuildMemsNeed bool     `codec:"guild_mm_n_"`
		SyncGuildMems     [][]byte `codec:"guild_mm_"`

		SyncApplyGuildMemsNeed bool     `codec:"guild_a_mm_n_"`
		SyncGuildApplyList     [][]byte `codec:"guild_a_mm_"`

		SyncClientTagNeed   bool   `codec:"tag_n_"`
		SyncClientTag       []int  `codec:"tag_"`
		SyncUnlockedAvatars []int  `codec:"unlock_"`
		SyncGuildRanking    string `codec:"ranking_"`

		// red point
		SyncRedPoint []int `codec:"redp_"`

	*/
	SyncUnlockedAvatars    []int `codec:"unlock_"`
	SyncCanUnlockedAvatars []int `codec:"canunlock_"`

	// shop
	ShopTodayRefreshTime int64    `codec:"shops_ref_t_"`
	SyncShops            [][]byte `codec:"shops_"`
	SyncShopGoods        [][]byte `codec:"shop_gs_"`

	//DestinyGeneral
	SyncDestinyGenerals                   []int    `codec:"dgs_"`
	SyncDestinyGeneralLvs                 []int    `codec:"dglvs_"`
	SyncDestinyGeneralMaxLvs              []int    `codec:"dgmlvs_"`
	SyncDestinyGeneralExps                []uint32 `codec:"dgmexps_"`
	SyncDestinySkills                     []int    `codec:"dgss_"`
	SyncDestinyGeneralCurr                int      `codec:"dgcurr_"`
	SyncDestinyGeneralFirstIds            []int    `codec:"fst_dgs_"`
	SyncDestingGeneralFisrtPassNames      []string `codec:"fstnms_"`
	SyncDestingGeneralFisrtPassAvatarIds  []int    `codec:"fstavids_"`
	SyncDestingGeneralFisrtPassTimeStamps []int64  `codec:"fstts_"`
	SyncDestingGeneralVipFreeTimes        uint32   `codec:"dgvipfts_"`
	SyncDestingGeneralVipRefTS            int64    `codec:"dgviprefts_"`

	// trial
	SyncTrialNeed bool   `codec:"trial_b_"`
	SyncTrialInfo []byte `codec:"trial_"`

	// recover
	SyncRecoverNeed bool     `codec:"recvr_b_"`
	SyncRecoverInfo [][]byte `codec:"recvr_"`

	// daily award
	SyncDailyAwardNeed        bool    `codec:"da_b_"`
	SyncDailyAwardsStates     []int   `codec:"da_s_"`
	SyncDailyAwardsStatesLen  []int   `codec:"da_s_l_"`
	SyncDailyAwardNextRefTime []int64 `codec:"da_n_r_t_"`

	// fashion
	SyncFashionBagAllNeed    bool     `codec:"fb_b_"`
	SyncFashionBagAll        [][]byte `codec:"fb_"`
	SyncFashionBagUpdateNeed bool     `codec:"fb_u_b_"`
	SyncFashionBagUpdate     [][]byte `codec:"fb_u_"`
	SyncFashionBagDelNeed    bool     `codec:"fb_d_b_"`
	SyncFashionBagDel        []uint32 `codec:"fb_d_"`

	// sevenday rank
	SyncSevenDayRankNeed      bool  `codec:"sdr_b_"`
	SyncSevenDayRankEndTime   int64 `codec:"sdr_et_"`
	SyncSevenDayRankCloseTime int64 `codec:"sdr_ct_"`

	// account sevenday
	SyncAccountSevenDayNeed            bool     `codec:"asd_b_"`
	SyncAccountSevenDayRefTime         []int64  `codec:"asd_rt_"`
	SyncAccountSevenDayGoodInfo        [][]byte `codec:"asd_good_"`
	SyncAccountSevenDayGoodNextRefTime int64    `codec:"asd_good_nrt_"`

	// day zero time
	SyncTodayBeginTime int64 `codec:"day_begin_time_"`

	// hero
	SyncHeroStarLevel []uint32 `codec:"herostar_"`
	SyncHeroStarPiece []uint32 `codec:"herop_"`
	SyncHeroLevel     []uint32 `codec:"herolvl_"`
	SyncHeroExp       []uint32 `codec:"heroexp_"`
	// hero star
	SyncHeroIdentiy1Len []int    `codec:"hero_idt1_l_"`
	SyncHeroIdentiy1n   []string `codec:"hero_idt1_"`
	SyncHeroIdentiy2len []int    `codec:"hero_idt2_l_"`
	SyncHeroIdentiy2n   []string `codec:"hero_idt2_"`
	SyncHeroIdentiy3len []int    `codec:"hero_idt3_l_"`
	SyncHeroIdentiy3n   []string `codec:"hero_idt3_"`

	// hero talent
	SyncHeroTalentNeed            bool     `codec:"hero_tal_n_"`
	SyncHeroTalentLevel           []uint32 `codec:"hero_tal_lv_"`
	SyncHeroTalentPoint           uint32   `codec:"hero_tal_p_"`
	SyncHeroTalentPointUpdateTick int64    `codec:"hero_tal_t_"`
	SyncHeroTalentCount           int      `codec:"hero_tal_c_"`

	// hero soul
	SyncHeroSoulNeed bool   `codec:"hero_soul_n_"` // 服务器与客户端都已经停止使用，现在sync必定会传输武魂点数。
	SyncHeroSoulLvl  uint32 `codec:"hero_soul_lv"` // iscommoninfo的时候都带

	// first pay
	SyncFirstPayRewardNeed bool     `codec:"fst_p_n_"`
	SyncFirstPayReward     []uint32 `codec:"fst_p_"`

	// 砸蛋
	SyncHitEggNeed           bool   `codec:"hg_n_"`
	SyncHitEggNextUpdateTime int64  `codec:"hg_nu_t_"`
	SyncHitEggCurIdx         uint32 `codec:"hg_c_idx_"`
	SyncHitEggShow           []bool `codec:"hg_sw_t_"`
	SyncHitEggTodayGotHc     int64  `codec:"hg_ghc_"`
	SyncHitEggTodayGotHammer int64  `codec:"hg_ghmr_"`

	// 活动时间
	SyncHotConfigNeed           bool     `codec:"hot_config_need"`
	SyncHotActivityInfo         [][]byte `codec:"ha_i_"`
	SyncMarketSubActivity       [][]byte `codec:"ha_mrkt_s_"`   // MarketSubActivityConfig2Client
	SyncMarketSubActivityReward [][]byte `codec:"ha_mrkt_s_r_"` // MarketSubActivityConfigReward2Client
	// 运营活动
	SyncMarketActivityNeed   bool     `codec:"mrkt_n_"`
	SyncMarketActivityInfo   [][]byte `codec:"mrkt_i_"`
	SyncMarketActivityStates []int    `codec:"mrkt_g_sts_"`

	// 成长基金
	SyncGrowFundNeed     bool     `codec:"gf_n_"`
	SyncGrowFundActivate bool     `codec:"gf_act_"`
	SyncGrowFundHadBuy   []uint32 `codec:"gf_buy_"`

	// 本日登陆次数
	SyncTodayLoginTimes int `codec:"day_login_ts_"`

	// 主将和战队gs
	SyncHeroGs        [helper.AVATAR_NUM_CURR]int `codec:"hero_gs_"`
	SyncCorpGs        int                         `codec:"corp_gs"`
	SyncHeroBaseGSSum int                         `codec:"hero_base_gs_sum"`

	//我要名将信息
	SyncWantGeneralInfoNeed        bool     `codec:"wt_g_info_need"`
	SyncWantGeneralRefreshTime     int64    `codec:"wt_g_ref_t"`
	SyncWantGeneralPlayCount       uint32   `codec:"wt_g_play_count"`
	SyncWantGeneralPlayTotal       uint32   `codec:"wt_g_play_total"`
	SyncWantGeneralFreeResetCount  uint32   `codec:"wt_g_free_reset_count"`
	SyncWantGeneralFreeResetTotal  uint32   `codec:"wt_g_free_reset_total"`
	SyncWantGeneralCurrHcPlayCount uint32   `codec:"wt_g_cur_hc_p_c"`
	SyncWantGeneralPlayResult      []uint32 `codec:"wt_g_play_result"`
	SyncWantGeneralNo1Acid         string   `codec:"wt_g_no1_acid"`
	SyncWantGeneralNo1Name         string   `codec:"wt_g_no1_nm"`
	SyncWantGeneralNo1Award        uint32   `codec:"wt_g_no1_award"`

	// 主将阵容
	SyncHeroTeamNeed bool  `codec:"hero_tm_n_"`
	SyncHeroTeamC    []int `codec:"hero_tm_c_"`
	SyncHeroTeam     []int `codec:"hero_tm_"`

	// 限时名将信息
	SyncHeroGachaRaceInfoNeed   bool   `codec:"h_gc_i_need"`
	SyncHeroGachaRaceActivityId int64  `codec:"h_gc_actid"`
	SyncHeroGachaRaceRank       int64  `codec:"h_gc_rank"`
	SyncHeroGachaRaceScore      int64  `codec:"h_gc_score"`
	SyncHeroGachaRaceChestInfo  []bool `codec:"h_gc_chest_i"`

	//微信分享信息
	SyncShareWeChatNeed bool     `codec:"shr_wc_need"`
	SyncShareWeChat     [][]byte `codec:"shr_wechat"`

	SyncHeroSwing     [][]byte `codec:"hero_sw"`
	SyncHeroSwingNeed bool     `codec:"hero_sw_need"`
	SyncShowHeroSwing bool     `codec:"hero_sw_show"`

	// debug info
	DebugTimeOffSet int64 `codec:"debug_time_offset"`

	//MoneyCat 招财猫
	MoneyCatNeed   bool  `codec:"mc_info_need"`
	MoneyCatStatus int64 `codec:"mc_s"`

	// 远征
	SyncExpeditionInfoNeed bool  `codec:"e_info_need"`
	ExpeditionState        int64 `codec:"edstate"` // 当前最远关卡
	ExpeditionAvard        int64 `codec:"edavard"` // 当前最远宝箱
	ExpeditionNum          int64 `codec:"ednum"`   // 远征通关总计次数
	ExpeditionStep         bool  `codec:"edstep"`  //是否通过九关

	// 节日Boss
	SyncFestivalBossInfoNeed bool    `codec:"fb_info_need"`
	FestivalShopRewardTime   []int64 `codec:"fb_sr_t"`

	// gve
	GveStartTime int64 `codec:"gve_st"`
	GveEndTime   int64 `codec:"gve_et"`

	// 情缘
	SyncHeroCompanion     [][]byte `codec:"hero_hcp"`
	SyncHeroCompanionNeed bool     `codec:"hero_hcp_need"`

	// gvg
	SyncGVGNeed     bool     `codec:"need_sync_gvg"`
	SyncGVGTime     []byte   `codec:"gvg_time_"`
	SyncGVGCityData [][]byte `codec:"gvg_city_data"`

	// 限时商店 配置表
	SyncLimitGoodConfigs [][]byte `codec:"limit_good_config"`
	// 限时商店 购买信息
	SyncBuyLimitShopNeed  bool  `codec:"buy_limit_shop_need"`
	SyncBuyLimitShopInfo  []int `codec:"buy_limit_shop_info"`
	SyncBuyLimitShopCount []int `codec:"buy_limit_shop_count"`
	// 神兵系统
	SyncExclusiveWeaponInfo [][]byte `codec:"ex_wea_info"`
	SyncExclusiveWeaponNeed bool     `codec:"ex_wea_need"`

	// 出奇制胜 武将差异化
	SyncHeroDiffNeed bool   `codec:"hero_diff_need"`
	SyncHeroDiff     []byte `codec:"sync_hero_diff"`

	// 7日红包信息
	SyncRedPacket7DaysNeed  bool    `codec:"redp_need"`
	SyncRedPacket7Days      []int64 `codec:"redp_days"`
	SyncRedPacketCreatTime  int64   `codec:"redp_ctime"`
	SyncRedPacket7DaysPoint bool    `codec:"redp_p"`
	SyncRedPacket7DaysSumHc int64   `codec:"redp_s"`

	// 白盒宝箱
	SyncWhiteGachaInfoNeed     bool  `codec:"wg_need"`
	SyncNextWhiteGachaFreeTime int64 `codec:"wg_t"`
	SyncWhiteGachaBless        int64 `codec:"wg_b"`
	SyncWhiteGachaNum          int64 `codec:"wg_n"`
	SyncWhiteGachaMaxWish      int64 `codec:"wg_mw"` // 祝福值最大值
	// 名将体验关
	SyncExperienceLevelNeed bool     `codec:"el_need"`
	SyncExperienceLevelId   []string `codec:"el_id"`

	// 无双争霸
	WsRank             int     `codec:"ws_rank"`
	NotClaimedReward   int     `codec:"not_claimed_reward"`
	LastRankChangeTime int64   `codec:"last_rank_change_time"`
	HasClaimedBox      []int   `codec:"has_claimed_box"`
	HasChallengeCount  int     `codec:"has_challenge_count"`
	BestRank           int     `codec:"best_rank"`
	HasClaimedBestRank []int   `codec:"has_claimed_best_rank"`
	DefenseFormation   []int64 `codec:"defense_formation"`
	SyncWSPVPNeed      bool    `codec:"wspvp_need"`

	// facebook 是否绑定
	SyncFaceBookNeed bool `codec:"facebook_need"`
	SyncIsFaceBook   bool `codec:"is_facebook"`

	//Twitter分享
	SyncTwitterNeed    bool `codec:"twitter_need"`
	SyncTwitterIsShare bool `codec:"is_twitter_share"`

	//Line分享
	SyncLineNeed    bool `codec:"line_need"`
	SyncLineIsShare bool `codec:"is_line_share"`

	// Oppo 相关
	SyncOppoNeed               bool `codec:"sync_oppo_need"`
	SyncOppoTodaySigned        bool `codec:"oppo_today_sign"`
	SyncOppoDailyQuestFineshed bool `codec:"oppo_quest_finished"`
	SyncOppoSignDays           int  `codec:"oppo_sign_days"`

	// 武将宿命
	SyncHeroDestinyNeed  bool  `codec:"sync_hd_need"`
	HeroDestinyActivated []int `codec:"hd_active"`
	HeroDestinyLevel     []int `codec:"hd_level"`

	// 武将相关属性消息优化
	ChangedHeroAvatar       int  `codec:"ch_hero_id"`
	UpdateHeroStarLevelNeed bool `codec:"up_hsl_need"`
	UpdateHeroStarPieceNeed bool `codec:"up_hsp_need"`
	UpdateHeroLevelNeed     bool `codec:"up_hl_need"`
	UpdateHeroExpNeed       bool `codec:"up_he_need"`
	UpdateHeroSkillsNeed    bool `codec:"up_hs_need"`
	UpdateHeroWingsNeed     bool `codec:"up_hw_need"`
	UpdateHeroCompanionNeed bool `codec:"up_hc_need"`
	UpdateHeroExclusiveNeed bool `codec:"up_hew_need"`

	UpdateHeroStarPiece int    `codec:"up_hsp"`
	UpdateHeroStarLevel int    `codec:"up_hsl"`
	UpdateHeroLevel     int    `codec:"up_hl"`
	UpdateHeroExp       int64  `codec:"up_he"`
	UpdateHeroSkills    []byte `codec:"up_hs"`
	UpdateHeroWings     []byte `codec:"up_hw"`
	UpdateHeroCompanion []byte `codec:"up_hc"`
	UpdateHeroExclusive []byte `codec:"up_hew"`

	// 黑盒宝箱
	BlackGachaSettings [][]byte `codec:"bg_settings"`
	BlackGachaShows    [][]byte `codec:"bg_shows"`
	BlackGachaLowest   [][]byte `codec:"bg_lowest"`

	// IAPGift
	IsIAPGift bool `codec:"iap_gift"`

	//幸运转盘
	SyncWheelInfoNeed bool `codec:"wheel_need"`

	//限时神将
	HotLimitGachaBox  [][]byte `codec:"boxitem_ids"`
	HotLimitGachaRank [][]byte `codec:"hgrranks_ids"`
	HgrOptions        [][]byte `codec:"hgroptions_ids"` // 活动ID信息
	GachaRacePoint    int64    `codec:"gr_point"`       // 每次购买涨多少积分
	PublicityTime     int64    `codec:"publicitytime"`  // 公示时间（秒）

	//军团红包
	RedPacketInfo [][]byte `codec:"red_packet_info"` // 红包信息

	// 武将星魂
	SyncHeroStarMapNeed bool     `codec:"sync_hsm_need"`
	HeroStarMap         [][]byte `codec:"sync_hsm"`

	SyncBindMailRewardNeed   bool `codec:"sync_b_ml_rw"`
	SyncHasGetBindMailReward bool `codec:"has_get_b_ml_rw"`
	SyncHasGetBindEGReward   bool `codec:"has_get_eg_ml_rw"`

	HasOfflineRecovers bool `codec:"has_off_recover"`
	SyncOffRecoverNeed bool `codec:"off_recover_need"`
	OffRecoverRedpoint bool `codec:"off_recover_redpoint"`

	//灵宠
	SyncMagicPetInfo      bool     `codec:"sync_mp_i"`
	SyncHeroMagicPetsInfo [][]byte `codec:"sync_hmp_i"`

	//战阵
	SyncBattleArmyInfo         bool     `codec:"sync_ba_i"`
	SyncAccountBattleArmysInfo [][]byte `codec:"sync_aba_i"`

	// 武将碎片抽奖功能
	SurplusGachaEndTime   int64 `codec:"surg_et"`
	SurplusDrawGachaCount []int `codec:"surg_dc"`
	SurplusGachaFirstOpen bool  `codec:"surg_fo"`

	sc_need_sync               bool
	hc_need_sync               bool
	exps_need_sync             bool
	corp_lv_need_sync          bool
	energy_need_sync           bool
	boss_fight_point_Sync      bool
	boss_fight_need_sync       bool
	stage_update_all           bool
	chapter_update_all         bool
	equip_need_sync            bool
	avatar_equip_need_sync     bool
	avatar_jade_need_sync      bool
	destinyGen_jade_need_sync  bool
	quest_need_all             bool
	store_all_sync             bool
	shop_all_sync              bool
	shop_sync                  bool
	gacha_all_sync             bool
	arousal_all_sync           bool
	skill_need_sync            bool
	general_need_sync          bool
	general_rel_need_sync      bool
	general_quest_need_sync    bool
	general_team_need_sync     bool
	last_stage_update          bool
	vip_need_sync              bool
	buy_need_sync              bool
	base_data_need_sync        bool
	pvp_count_need_sync        bool
	act_gift_by_cond_need_sync bool
	act_gift_by_time_need_sync bool
	unlock_avatar_need_sync    bool
	destiny_need_sync          bool
	update_friend_list_sync    bool
	update_black_list_sync     bool

	SyncFriendListNeed      bool     `codec:"s_fr_l_n"`
	SyncBlackListNeed       bool     `codec:"s_bl_l_n"`
	SyncFriendList          [][]byte `codec:"sync_friend_list"`
	SyncBlackList           [][]byte `codec:"sync_black_list"`
	SyncFriendGiftListNeed  bool     `codec:"s_fr_gt_t"`
	SyncFriendGift          []string `codec:"s_fr_gt"`
	SyncReceiveGiftListNeed bool     `codec:"s_rc_gt_l_n"`
	SyncReceiveGiftList     [][]byte `codec:"s_rc_gt_l"`
	SyncReceiveGiftTimes    int      `codec:"s_rc_gt_t"`

	/*
		gates_enemy_data_need_sync      bool
		gates_enemy_push_data_need_sync bool
	*/

	stage_update   string
	chapter_update string

	gameMode_update     []uint32
	gameMode_update_all bool

	items_update        []uint32
	items_update_o      map[uint32]int64
	items_update_reason string
	items_del           []uint32
	items_del_o         map[string]int64
	items_del_reason    string

	jades_update        []uint32
	jades_update_o      map[uint32]int64
	jades_update_reason string
	jades_del           []uint32
	jades_del_o         map[string]int64
	jades_del_reason    string

	fashion_update        []uint32
	fashion_update_o      map[uint32]int64
	fashion_update_reason string
	fashion_del           []uint32
	fashion_del_o         map[string]int64
	fashion_del_reason    string

	shop_id uint32
}

// 处理更新数据
func (s *SyncResp) mkInfo(p *Account) {
	profile := &p.Profile
	acid := p.AccountID.String()
	shard := p.AccountID.ShardId
	vip_lv := profile.GetVipLevel()

	now_t := profile.GetProfileNowTime()

	commonNextDayBeginTime := util.GetNextDailyTime(
		gamedata.GetCommonDayBeginSec(now_t), now_t)
	pvpNextDayBeginTime := util.GetNextDailyTime(
		gamedata.GetPVPBalanceBeginSec(now_t), now_t)
	s.ShardId = game.Cfg.ShardId[0]

	// 每次Sync都带的信息，在下面
	s.SyncCommonNeed = true

	s.DataVer, s.BundleVer, s.ActValid, s.DataMin, s.BundleMin = game.Cfg.GetHotCfgData(
		p.Profile.GetClientMarketVer(), p.AccountID.ShardId)

	s.SyncCurrAvatar = p.Profile.GetCurrAvatar() + 1 // 与默认值区别

	s.SyncTodayBeginTime = util.DailyBeginUnix(now_t)

	s.SyncGuildChangeToday = p.GuildProfile.IsTodayChangeGuild(p.GetProfileNowTime())

	if p.Tmp.TrialFirst {
		p.Tmp.TrialFirst = false
		s.OnChangeGameMode(counter.CounterTypeTrial)
	}
	if p.Tmp.ExpeditionFirst {
		p.Tmp.ExpeditionFirst = false
		s.OnChangeGameMode(counter.CounterTypeExpedition)
	}

	if profile.GetCorp().IsNeedSyncUnlocked() {
		s.OnChangeUnlockAvatar()
		s.OnChangeFashionBag()
		s.OnChangeAvatarExp()
		s.OnChangeAvatarEquip()
		s.OnChangeEquip()

		profile.GetCorp().SetNoNeedSync()
	}

	if profile.GetDestinyGeneral().IsNeedSync() {
		s.OnChangeDestinyGeneral()
		logs.Trace("OnChangeDestinyGeneral %v", profile.GetDestinyGeneral())
		profile.GetDestinyGeneral().HasSync()
	}

	s.mkFashionAllInfo(p)

	// 检查活动（7日领奖）红点
	if p.ActGiftRedPoint() {
		s.OnChangeRedPoint(notify.RedPointTyp_ActGift)
	}

	//skill自动解锁
	player_skill := profile.GetAvatarSkill()
	if player_skill.IsHasSkillUnlockThenNeedSync() {
		s.OnChangeSkillAllChange()
		player_skill.SetSkillUnlockHasSync()
	}

	// title
	s.mkTitleInfo(p)

	if s.corp_lv_need_sync {
		s.SyncCorpLv, s.SyncCorpXp = p.Profile.GetCorp().GetXpInfo()
	}

	//
	// GS更新, 因为所有涉及到数据变更的操作都会走同步,
	// 所以这里来检查最大GS是否变更, 需要先检查一下防止里面有用的
	// 注意：所有可能引起gs变化的要在之前做
	//
	p.CheckChangeMaxGS()

	if s.SyncFullItemsNeed {
		s.Items = encode(p.BagProfile.ItemToClients())
	}

	if s.SyncUpdateItemsNeed {
		items := p.BagProfile.GetItemsToClient(s.items_update)
		s.ItemsUpdate = encode(items)
		// logiclog
		addItems := make(map[string]int64, len(s.items_update_o))
		delItems := make(map[string]int64, len(s.items_update_o))
		for idx, oldCount := range s.items_update_o {
			item := items[idx]
			newCount := item.Count
			if newCount > oldCount {
				addItems[item.TableID] = newCount - oldCount
				logiclog.LogGiveItemUseSelf(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
					p.Profile.ChannelId, s.items_update_reason, item.TableID, oldCount, newCount-oldCount, newCount,
					p.Profile.GetVipLevel(), func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

			} else if newCount < oldCount {
				delItems[item.TableID] = oldCount - newCount
				logiclog.LogCostItemUseSelf(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
					p.Profile.ChannelId, s.items_update_reason, item.TableID, oldCount, newCount-oldCount, newCount,
					p.Profile.GetVipLevel(), func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			}
		}
		if len(addItems) > 0 {
			logiclog.LogGiveItem(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
				p.Profile.ChannelId, s.items_update_reason, addItems, p.Profile.GetVipLevel(),
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
		}
		if len(delItems) > 0 {
			logiclog.LogCostItem(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
				p.Profile.ChannelId, s.items_update_reason, delItems, p.Profile.GetVipLevel(),
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
		}
	}

	if s.SyncDelItemsNeed {
		s.ItemsDel = s.items_del
		// logiclog
		logiclog.LogCostItem(acid, p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId, s.items_del_reason, s.items_del_o, p.Profile.GetVipLevel(),
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	}

	if s.sc_need_sync {
		s.SyncSC = p.Profile.GetSC().GetAll()
	}

	if s.hc_need_sync {
		s.SyncHC = p.Profile.GetHC().GetHC() + 1 // 与默认值区别
		s.SyncHCBuy = p.Profile.GetHC().BuyFromHc
		s.SyncHCBuyToday = p.Profile.GetHC().GetBuyFromHcToday(p.Profile.GetProfileNowTime())
	}

	if s.exps_need_sync {
		s.SyncExps = p.Profile.GetAvatarExp().GetAll()
	}

	if s.energy_need_sync {
		s.SyncValue, s.SyncRefershTime, s.SyncLastTime = p.Profile.GetEnergy().Get()
	}

	if s.boss_fight_point_Sync {
		s.SyncBossFightPointValue, s.SyncBossFightPointRefershTime, s.SyncBossFightPointLastTime =
			profile.GetBossFightPoint().Get()
	}

	if s.arousal_all_sync {
		s.SyncAvatarArousal = p.Profile.GetAvatarExp().GetAvatarArousalLv()
	}

	if s.skill_need_sync {
		s.SyncAvatarSkillLevel, s.SyncAvatarSkillLevelLen = p.Profile.GetAvatarSkill().GetAll()
		s.SyncSkillPracticeLv = p.Profile.GetAvatarSkill().GetPracticeLevel()
	}

	// stage, chapter
	s.mkStageInfo(p)

	if s.equip_need_sync {
		eq := p.Profile.GetEquips()
		equips, glv, elv, star, star_xp, lv_mat_enh, mat_enh, slot_max := eq.Curr()
		mat_enh_l := helper.EQUIP_MAT_ENHANCE_MAT

		s.CurrEquips = make([][]byte, 0, len(equips))
		s.EquipMatEnh = make([]bool, 0, len(equips)*mat_enh_l)
		s.EquipMatEnhMax = mat_enh_l
		for idx, eid := range equips {
			e := Equip{eid, glv[idx], elv[idx], star[idx], star_xp[idx], lv_mat_enh[idx]}
			s.CurrEquips = append(s.CurrEquips, encode(e))
			s.EquipMatEnh = append(s.EquipMatEnh, mat_enh[idx][:mat_enh_l]...)
		}
		s.SlotMax = slot_max
		all := p.Profile.GetMyVipCfg().StarHcUpLimitDaily
		s.StarHcUpLastCount = all - eq.StarHcUpCount.Get(now_t)
	}

	if s.avatar_equip_need_sync {
		equips, slot_max := p.Profile.GetAvatarEquips().Curr()
		s.CurrAvatarEquips = make([][]byte, 0, len(equips))
		for _, eid := range equips {
			e := AvatarEquip{eid}
			s.CurrAvatarEquips = append(s.CurrAvatarEquips, encode(e))
		}
		s.AvatarSlotMax = slot_max
	}

	// gift
	s.mkGiftInfo(p)

	// general
	s.mkGeneralInfo(p)

	if s.SyncMailNeed {
		mails := p.Profile.GetMails().GetAllMail()

		s.SyncMail = make([][]byte, 0, len(mails))
		for i := 0; i < len(mails); i++ {
			s.SyncMail = append(s.SyncMail, encode(
				newMailToClient(mailToClient{
					mails[i].Idx,
					mails[i].IdsID,
					mails[i].Param,
					mails[i].Tag,
					mails[i].IsRead,
					mails[i].ItemId,
					mails[i].Count,
					timail.GetTimeFromMailID(mails[i].Idx),
					mails[i].TimeEnd,
				})))
		}
	}

	if s.vip_need_sync {
		v, rmb := p.Profile.GetVip().GetVIP()
		s.SyncVIPLv = v + 1         // 这里是实际值加1 区分0值
		s.SyncVIPRMBPoint = rmb + 1 // 这里是实际值加1 区分0值
	}

	if s.buy_need_sync {
		s.SyncBuy = make([][]byte, buy.Buy_Typ_Count, buy.Buy_Typ_Count)
		player_buy := profile.GetBuy()
		for i := 0; i < buy.Buy_Typ_Count; i++ {
			realLimitCount := -10
			if i == buy.Buy_Typ_GuildBigBossCount {
				realLimitCount, _ = p.Profile.Counts.Get(counter.CounterTypeGuildBigBossBuyTime, p)
			}
			s.SyncBuy[i] = encode(player_buy.GetInfoToClient(
				i,
				vip_lv,
				now_t,
				realLimitCount))
		}

		stageTimesBuy := profile.GetBuy().GetStageTimesInfoToClient(vip_lv, now_t)
		s.SyncBuyStageTimes = encode(stageTimesBuy)
	}

	s.mkExpeditionInfo(p)

	if s.SyncGameModeNeed {
		game_mode := profile.GetGameMode()
		if s.gameMode_update_all {
			infos := game_mode.GetAllSyncInfo(p.Account)
			c := len(infos)
			s.SyncGameModeInfo = make([][]byte, c, c)
			for i := 0; i < c; i++ {
				s.SyncGameModeInfo[i] = encode(infos[i])
			}
		} else {
			c := len(s.gameMode_update)
			s.SyncGameModeInfo = make([][]byte, c, c)
			for i := 0; i < c; i++ {
				s.SyncGameModeInfo[i] = encode(
					game_mode.GetSyncInfo(
						p.Account,
						s.gameMode_update[i]))
			}
		}
	}

	// quest
	s.mkQuestInfo(p)

	if s.SyncPrivilegeBuyNeed {
		prvlg_by := profile.GetPrivilegeBuy()
		infos := prvlg_by.GetAllSyncInfo()
		l := len(infos)
		s.SyncPrivilegeBuy = make([][]byte, l, l)
		for i := 0; i < l; i++ {
			s.SyncPrivilegeBuy[i] = encode(infos[i])
		}
	}

	if s.base_data_need_sync {
		s.Name = p.Profile.Name
		s.SyncRenameCount = p.Profile.RenameCount
	}

	if s.SyncNewHandNeed {
		// 解压
		s.SyncNewHand = profile.GetNewHand()
	}
	if s.SyncNewHandIgnoreNeed {
		s.SyncNewHandIgnore = p.Profile.NewHandIgnore
	}

	// 检查1v1竞技场每日竞技宝箱是否重置
	p.Profile.GetSimplePvp().UpdateChestInfo(now_t)

	// 检查1v1竞技场每日换人次数
	p.Profile.GetSimplePvp().UpdateSwitchCount(now_t)

	// 检查3v3竞技场每日竞技宝箱是否重置
	p.Profile.GetTeamPvp().UpdateChestInfo(now_t)

	// Hmt活动 是否登录
	p.Profile.GetHmtActivityInfo().IsLogin(p.GetProfileNowTime())
	if s.pvp_count_need_sync {
		// +1来区别没有同步
		//s.SyncPvpCount = profile.GetSimplePvp().GetPvpCount(now_t) + 1

		s.SyncPvpCountNextRefTime = pvpNextDayBeginTime + 1
		pvpInfo := p.Profile.GetSimplePvp()
		s.SyncPvpDefAvatarId = pvpInfo.PvpDefAvatar + 1
		s.SyncPvpCountToday = pvpInfo.PvpCountToday
		s.SyncPvpOpendChests = make([]uint32, 0, len(pvpInfo.OpenedChestIDs))
		for _, id := range pvpInfo.OpenedChestIDs {
			s.SyncPvpOpendChests = append(s.SyncPvpOpendChests, id)
		}
	}

	iap := p.Account.Profile.GetIAPGoodInfo()
	hotBuild := gamedata.GetHotDataVerCfg().Build
	if s.SyncIAPGoodInfoNeed || iap.SyncObj.IsNeedSync() ||
		iap.DataBuild != hotBuild {
		s.SyncIAPGoodInfoNeed = true
		infos := iap.GetPayGoodInfoClient()
		s.SyncIAPGoodInfo = make([][]byte, len(infos))
		for i := 0; i < len(infos); i++ {
			s.SyncIAPGoodInfo[i] = encode(infos[i])
		}
		s.SyncCurrenSbatch = int64(iap.GetCurrenServerSbatch(p.AccountID.ShardId, p.Profile.GetProfileNowTime()))
		s.SyncMonthlyEndTime = iap.MonthlyCardEndTime
		s.SyncMonthlyValidTime = iap.MonthlyValidTime
		s.SyncIsLifeCard = iap.IsLifeCard
		s.SyncLifeCardValidTime = iap.LifeCardValidTime
		s.SyncWeekRewardEndTime = iap.WeekRewardEndTime
		s.SyncWeekRewardValidTime = iap.WeekRewardValidTime

		s.SyncLevelGiftId = iap.LevelGiftId
		s.SyncLevelGiftEndTime = iap.LevelGiftEndTime
		s.SyncLevelGiftIdWaitAward = iap.LevelGiftIdWaitAward
		iap.DataBuild = hotBuild
		iap.SyncObj.SetHadSync()
	}

	// guild
	s.mkGuildInfo(p)

	if s.unlock_avatar_need_sync {
		corp := p.Profile.GetCorp()
		s.SyncUnlockedAvatars = corp.GetUnlockedAvatar()
	}

	s.ShopTodayRefreshTime = commonNextDayBeginTime
	if s.shop_all_sync {
		shops, goods := p.getAllShopsInfoForClient()
		s.SyncShops = make([][]byte, 0, len(shops))
		for i := 0; i < len(shops); i++ {
			s.SyncShops = append(s.SyncShops, encode(shops[i]))
		}
		s.SyncShopGoods = make([][]byte, 0, len(goods))
		for i := 0; i < len(goods); i++ {
			s.SyncShopGoods = append(s.SyncShopGoods, encode(goods[i]))
		}
	}

	if s.shop_sync {
		shops, goods := p.getShopsInfoForClient(s.shop_id)
		s.SyncShops = make([][]byte, 0, len(shops))
		for i := 0; i < len(shops); i++ {
			s.SyncShops = append(s.SyncShops, encode(shops[i]))
		}
		s.SyncShopGoods = make([][]byte, 0, len(goods))
		for i := 0; i < len(goods); i++ {
			s.SyncShopGoods = append(s.SyncShopGoods, encode(goods[i]))
		}
	}

	if s.destiny_need_sync {
		distiny := profile.GetDestinyGeneral()
		distiny.UpdateDGTimes(int(p.Profile.GetVipLevel()), now_t)

		s.SyncDestinyGeneralCurr = distiny.CurrGeneralIdx
		s.SyncDestinyGeneralLvs = make([]int, 0, len(distiny.Generals))
		s.SyncDestinyGeneralMaxLvs = make([]int, 0, len(distiny.Generals))
		s.SyncDestinyGenerals = make([]int, 0, len(distiny.Generals))
		for i := 0; i < len(distiny.Generals); i++ {
			maxLv := 1
			data := gamedata.GetNewDestinyGeneralLevelDatas(distiny.Generals[i].Id)
			if data != nil {
				maxLv = len(data) - 1
			}
			s.SyncDestinyGeneralMaxLvs = append(s.SyncDestinyGeneralMaxLvs, maxLv)
			s.SyncDestinyGeneralLvs = append(s.SyncDestinyGeneralLvs, distiny.Generals[i].LevelIndex)
			s.SyncDestinyGeneralExps = append(s.SyncDestinyGeneralExps, distiny.Generals[i].Exp)
			s.SyncDestinyGenerals = append(s.SyncDestinyGenerals, distiny.Generals[i].Id)
		}
		l := gamedata.GetDestingGeneralIdCount()
		s.SyncDestinyGeneralFirstIds = make([]int, 0, l)
		s.SyncDestingGeneralFisrtPassNames = make([]string, 0, l)
		s.SyncDestingGeneralFisrtPassAvatarIds = make([]int, 0, l)
		s.SyncDestingGeneralFisrtPassTimeStamps = make([]int64, 0, l)
		for i := 0; i < l; i++ {
			s.SyncDestinyGeneralFirstIds = append(s.SyncDestinyGeneralFirstIds, i)
			//SyncResp GetModule
			bs := time.Now().UnixNano()
			firstInfo := dest_gen_first.GetModule(p.AccountID.ShardId).GetFirstDestGen(i)
			metric_send(p.AccountID, "DestGenFirst", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
			s.SyncDestingGeneralFisrtPassNames = append(s.SyncDestingGeneralFisrtPassNames,
				firstInfo.FirstPlayerName)
			s.SyncDestingGeneralFisrtPassAvatarIds = append(s.SyncDestingGeneralFisrtPassAvatarIds,
				firstInfo.FirstPlayerAvatarId)
			s.SyncDestingGeneralFisrtPassTimeStamps = append(s.SyncDestingGeneralFisrtPassTimeStamps,
				firstInfo.FirstPlayerTimeStamp)
		}
		s.SyncDestinySkills = distiny.SkillGenerals[:]
		s.SyncDestingGeneralVipFreeTimes = distiny.VipTimes
		s.SyncDestingGeneralVipRefTS = distiny.VipRefreshTimeStamp
		logs.Trace("OnChangeDestinyGeneral %v", profile.GetDestinyGeneral())
	}

	if s.SyncRecoverNeed {
		infos := profile.GetRecover().GetRecoverForClient()
		s.SyncRecoverInfo = make([][]byte, 0, len(infos))
		for _, v := range infos {
			if v.GiveScTyp != "" {
				s.SyncRecoverInfo = append(s.SyncRecoverInfo, encode(v))
			}
		}
	}

	// daily award
	if s.SyncDailyAwardNeed {
		info := profile.GetDailyAwards().GetDailyAwardsForClient()
		s.SyncDailyAwardsStates = info.AwardsStates
		s.SyncDailyAwardsStatesLen = info.AwardsStatesLen
		s.SyncDailyAwardNextRefTime = info.AwardNextRefTime
	}

	s.mkJadeAllInfo(p)
	s.mkStoreAllInfo(p)
	s.mkGachaAllInfo(p)
	s.mkBossInfo(p)
	s.mkTrialAllInfo(p)
	s.MkPhoneData(p.Account)

	if s.SyncSevenDayRankNeed {
		err, e, c := gamedata.CalcSevOpnRankEndTime(rank.GetSevenDayRankStartTime(shard))
		if err != nil {
			logs.Error("CalcSevOpnRankEndTime err %v", err)
		} else {
			s.SyncSevenDayRankEndTime = e
			s.SyncSevenDayRankCloseTime = c
			logs.Trace("SevOpnRankEndTime %d %d %d", rank.GetSevenDayRankStartTime(shard),
				s.SyncSevenDayRankEndTime,
				s.SyncSevenDayRankCloseTime)
		}
	}

	if s.SyncAccountSevenDayNeed {
		ct := gamedata.GetCommonDayBeginSec(p.Profile.CreateTime)
		s.SyncAccountSevenDayRefTime = make([]int64, 0, gamedata.GetAccount7DaySumDays())
		for i := 1; i <= gamedata.GetAccount7DaySumDays(); i++ {
			s.SyncAccountSevenDayRefTime = append(s.SyncAccountSevenDayRefTime,
				ct+int64(util.DaySec*i))
		}

		a7d := p.Profile.GetAccount7Day()
		a7d.UpdateGoods(now_t)
		s.SyncAccountSevenDayGoodNextRefTime = a7d.NextRefDailyTime
		s.SyncAccountSevenDayGoodInfo = make([][]byte, 0, len(a7d.Goods))
		servCount := p.getAccount7DayServGoodCount()
		for _, v := range a7d.Goods {
			c := Good7DayToClient{
				PromotionID:     v.PromotionID,
				LeftTimes:       v.LeftTimes,
				ServerLeftCount: servCount[v.PromotionID],
			}
			s.SyncAccountSevenDayGoodInfo = append(s.SyncAccountSevenDayGoodInfo,
				encode(c))
		}
	}

	// teampvp
	s.mkTeamPvpInfo(p)
	s.mkFirstPassRewardInfo(p)

	hero := p.Profile.GetHero()
	if hero.IsNeedSync() {
		s.SyncHeroStarLevel = hero.HeroStarLevel[:helper.AVATAR_NUM_CURR]
		s.SyncHeroLevel = hero.HeroLevel[:]
		s.SyncHeroExp = hero.HeroExp[:]
		s.SyncHeroStarPiece = hero.HeroStarPiece[:helper.AVATAR_NUM_CURR]
		//heroskill
		s.SyncHeroIdentiy1Len = make([]int, 0, len(hero.HeroSkills))
		s.SyncHeroIdentiy2len = make([]int, 0, len(hero.HeroSkills))
		s.SyncHeroIdentiy3len = make([]int, 0, len(hero.HeroSkills))
		s.SyncHeroIdentiy1n = make([]string, 0, 10)
		s.SyncHeroIdentiy2n = make([]string, 0, 10)
		s.SyncHeroIdentiy3n = make([]string, 0, 10)

		for i := 0; i < len(hero.HeroSkills) && i < helper.AVATAR_NUM_CURR; i++ {
			s.SyncHeroIdentiy1Len = append(s.SyncHeroIdentiy1Len, len(hero.HeroSkills[i].PassiveSkill))
			s.SyncHeroIdentiy2len = append(s.SyncHeroIdentiy2len, len(hero.HeroSkills[i].CounterSkill))
			s.SyncHeroIdentiy3len = append(s.SyncHeroIdentiy3len, len(hero.HeroSkills[i].TriggerSkill))
			for _, vlaue := range hero.HeroSkills[i].PassiveSkill {
				s.SyncHeroIdentiy1n = append(s.SyncHeroIdentiy1n, vlaue)
			}
			for _, vlaue := range hero.HeroSkills[i].CounterSkill {
				s.SyncHeroIdentiy2n = append(s.SyncHeroIdentiy2n, vlaue)
			}
			for _, vlaue := range hero.HeroSkills[i].TriggerSkill {
				s.SyncHeroIdentiy3n = append(s.SyncHeroIdentiy3n, vlaue)
			}

		}

		hero.SetHadSync()
	}
	if s.SyncHeroTalentNeed {
		tal := p.Profile.GetHeroTalent()
		s.SyncHeroTalentPointUpdateTick = tal.UpdateTalentPoint(now_t)
		s.SyncHeroTalentCount = gamedata.HeroTalentCount
		s.SyncHeroTalentLevel = make([]uint32, 0, helper.AVATAR_NUM_MAX)
		for i := 0; i < helper.AVATAR_NUM_MAX; i++ {
			heroTalent := tal.HeroTalentLevel[i]
			s.SyncHeroTalentLevel = append(s.SyncHeroTalentLevel,
				heroTalent[:gamedata.HeroTalentCount]...)
		}
		s.SyncHeroTalentPoint = tal.TalentPoint
	}

	// iscommoninfo
	s.SyncHeroSoulLvl = p.Profile.GetHeroSoul().HeroSoulLevel

	// first pay
	if s.SyncFirstPayRewardNeed {
		s.SyncFirstPayReward = p.Profile.FirstPayReward.FirstPayReward
	}

	// hit egg
	if s.SyncHitEggNeed {
		p.Profile.GetHitEgg().UpdateHitEgg(now_t)
		s.SyncHitEggNextUpdateTime = p.Profile.GetHitEgg().NextEggUpdateTime
		s.SyncHitEggCurIdx = p.Profile.GetHitEgg().CurIdx
		s.SyncHitEggShow = p.Profile.GetHitEgg().EggsShow
		s.SyncHitEggTodayGotHc = p.Profile.GetHitEgg().TodayGotHc
		s.SyncHitEggTodayGotHammer = p.Profile.GetHitEgg().TodayGotHammer
	}

	// grow fund
	if s.SyncGrowFundNeed {
		s.SyncGrowFundActivate = p.Profile.GetGrowFund().IsActivate
		s.SyncGrowFundHadBuy = p.Profile.GetGrowFund().Bought
	}

	s.mkHotInfo(p)

	// 我要名将
	if s.SyncWantGeneralInfoNeed {
		wantInfo := p.Profile.GetWantGeneralInfo()
		wantInfo.UpdateInfo(p.Account, p.Profile.GetProfileNowTime())
		s.SyncWantGeneralFreeResetCount = wantInfo.CanFreeResetCountCurr
		s.SyncWantGeneralFreeResetTotal = wantInfo.CanFreeResetCountTotal
		s.SyncWantGeneralCurrHcPlayCount = wantInfo.CurrHcResetCount
		s.SyncWantGeneralPlayCount = wantInfo.CanPlayCountCurr
		s.SyncWantGeneralPlayTotal = wantInfo.CanPlayCountTotal
		s.SyncWantGeneralRefreshTime = wantInfo.NextPlayRefreshTime
		s.SyncWantGeneralPlayResult = wantInfo.PlayResult[:]

		//SyncResp GetModule
		bs := time.Now().UnixNano()
		info := want_gen_best.GetModule(p.AccountID.ShardId).GetWantGenBest(p.Profile.GetProfileNowTime())
		metric_send(p.AccountID, "WantGeneralFirst", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
		s.SyncWantGeneralNo1Acid = info.Acid
		s.SyncWantGeneralNo1Name = info.Name
		s.SyncWantGeneralNo1Award = info.HeroPieceCount
	}
	// world boss
	p.ResetWBGotInfo(now_t, s)

	if s.SyncHeroTeamNeed {
		hts := p.Profile.GetHeroTeams()
		s.SyncHeroTeamC = make([]int, 0, len(hts.HeroTeams))
		s.SyncHeroTeam = make([]int, 0, len(hts.HeroTeams)*2)
		for _, ht := range hts.HeroTeams {
			if ht.Team != nil {
				s.SyncHeroTeamC = append(s.SyncHeroTeamC, len(ht.Team))
				s.SyncHeroTeam = append(s.SyncHeroTeam, ht.Team...)
			} else {
				s.SyncHeroTeamC = append(s.SyncHeroTeamC, 0)
			}
		}
	}

	if s.SyncBindMailRewardNeed {
		bindMailInfo := p.Profile.GetBindMailRewardInfo()
		s.SyncHasGetBindMailReward = bindMailInfo.BindMailRewardGet
		s.SyncHasGetBindEGReward = bindMailInfo.BindEGRewardGet
	}

	hgr := p.Profile.GetHeroGachaRaceInfo()
	hgr.CheckActivity(acid, int64(gamedata.GetHGRCurrValidActivityId()))
	if s.SyncHeroGachaRaceInfoNeed {
		s.SyncHeroGachaRaceActivityId = hgr.ActivityID
		s.SyncHeroGachaRaceRank = hgr.Rank
		s.SyncHeroGachaRaceScore = hgr.GetCurScore()
		chestInfo := hgr.GetChestInfo()
		s.SyncHeroGachaRaceChestInfo = make([]bool, 0, len(chestInfo))
		for _, i := range chestInfo {
			s.SyncHeroGachaRaceChestInfo = append(s.SyncHeroGachaRaceChestInfo, i)
		}

	}
	//招财猫
	if s.MoneyCatNeed {
		s.MoneyCatStatus = p.Profile.GetMoneyCatInfo().MoneyCatTime
	}

	//节日boss
	if s.SyncFestivalBossInfoNeed {
		s.FestivalShopRewardTime = p.Profile.GetFestivalBossInfo().GetFbShopRewardTime()
	}

	if s.SyncHeroDiffNeed {
		heroDiff := p.Profile.GetHeroDiff()
		needNew := heroDiff.UpdateTodayInfo(now_t)
		if needNew {
			heroDiff.UpdateTodayStage(hero_diff.GetModule(p.AccountID.ShardId).GetTodayStage(now_t))
			logs.Debug("update hero diff today stage, stageSeq is: %v", heroDiff.TodayStage)

		}
		s.SyncHeroDiff = encode(p.Profile.GetHeroDiff())
	}

	if s.SyncRedPacket7DaysNeed {
		rp := p.Profile.GetRedPacket7day()
		s.SyncRedPacket7Days = rp.GetInfo2Client()
		s.SyncRedPacketCreatTime = rp.GetRelativeTime(p.GetProfileNowTime())
		s.SyncRedPacket7DaysPoint = !gamedata.IsSameDayCommon(p.GetProfileNowTime(), p.Profile.LogoutTime)
		s.SyncRedPacket7DaysSumHc = rp.GetTotalHc()
	}

	if s.SyncWhiteGachaInfoNeed {
		wg := p.Profile.GetWhiteGachaInfo()
		s.SyncNextWhiteGachaFreeTime = wg.LastGachaTime + int64(util.DaySec)
		s.SyncWhiteGachaBless = wg.GachaBless
		s.SyncWhiteGachaNum = wg.GachaNum
		_actInfo := gamedata.GetHotDatas().Activity.GetActivityInfo(gamedata.ActWhiteGacha, p.Profile.ChannelQuickId)

		if _actInfo != nil {
			for _, v := range _actInfo {
				if now_t > v.StartTime && now_t < v.EndTime {
					s.SyncWhiteGachaMaxWish = int64(gamedata.GetHotDatas().Activity.GetActivityGachaSeting(v.ActivityId).GetWishMax())
				}
			}

		}
	}

	// 检查微信分享次数是否需要重置
	p.Profile.GetShareWeChatInfo().UpdateTimesAndRest(now_t)
	if s.SyncShareWeChatNeed {
		items := p.Profile.GetShareWeChatInfo().GetItems()
		s.SyncShareWeChat = make([][]byte, 0, len(items))
		for _, item := range items {
			s.SyncShareWeChat = append(s.SyncShareWeChat, encode(item))
		}
	}

	// gve

	s.GveStartTime, s.GveEndTime = gamedata.GetGVETime(time.Now().Unix())

	// 名将体验关
	if s.SyncExperienceLevelNeed {
		s.SyncExperienceLevelId = p.Profile.GetExperienceLevelInfo().GetExperiendeLevel()
	}

	if s.SyncOppoNeed {
		oppo := p.Profile.GetOppoRelated()
		if gamedata.IsSameDayCommon(now_t, oppo.LastSignTime) {
			s.SyncOppoTodaySigned = true
		}
		if gamedata.IsSameDayCommon(now_t, oppo.LastDailyQuestTime) {
			s.SyncOppoDailyQuestFineshed = true
		}
		s.SyncOppoSignDays = oppo.SignDays
	}

	if s.SyncHeroDestinyNeed {
		infoList := p.Profile.GetHeroDestiny().GetActivateDestiny()
		s.HeroDestinyActivated = make([]int, 0)
		s.HeroDestinyLevel = make([]int, 0)
		for _, info := range infoList {
			s.HeroDestinyActivated = append(s.HeroDestinyActivated, info.Id)
			s.HeroDestinyLevel = append(s.HeroDestinyLevel, info.Level)
		}
	}

	if s.SyncHeroStarMapNeed {
		astrology := p.Profile.GetAstrology()
		heros := astrology.GetHeros()
		s.HeroStarMap = make([][]byte, 0, len(heros))
		for _, hero := range heros {
			s.HeroStarMap = append(s.HeroStarMap, encode(buildNetAstrologyHero(hero)))
		}
	}

	if s.SyncMagicPetInfo {
		hero := p.Profile.GetHero()
		propData := hero.HeroMagicPets
		s.SyncHeroMagicPetsInfo = make([][]byte, 0, helper.AVATAR_NUM_CURR)
		for i, item := range propData {
			s.SyncHeroMagicPetsInfo = append(s.SyncHeroMagicPetsInfo, encode(p.getMagicPetInfo(i, item)))
		}
	}

	if s.SyncBattleArmyInfo {
		battle_armys_data := p.Profile.BattleArmys.GetBattleArmys()
		logs.Trace("p.Profile.BattleArmys.GetBattleArmys():%v", p.Profile.BattleArmys.GetBattleArmys())
		s.SyncAccountBattleArmysInfo = make([][]byte, 0, helper.BATTLE_ARMY_NUM_MAX*helper.BATTLE_ARMYLOC_NUM_MAX)
		for i, v := range battle_armys_data {
			for j, v1 := range v.GetBattleArmyLocs() {
				logs.Trace("[cyt]战阵详细信息:%v", BattleArmyInfo{BattleArmyID: int64(i*helper.BATTLE_ARMYLOC_NUM_MAX + j + 1),
					BattleArmyAvatarID: int64(p.Profile.BattleArmys.GetBattleArmys()[i].GetBattleArmyLocs()[j].AvatarID),
					BattleArmyLev:      int64(p.Profile.BattleArmys.GetBattleArmys()[i].GetBattleArmyLocs()[j].Lev)})
				s.SyncAccountBattleArmysInfo = append(s.SyncAccountBattleArmysInfo,
					encode(BattleArmyInfo{BattleArmyID: int64(i*helper.BATTLE_ARMYLOC_NUM_MAX + j + 1), BattleArmyAvatarID: int64(v1.AvatarID), BattleArmyLev: int64(v1.Lev)}))
			}
		}
		logs.Trace("传输战阵信息完成,共%v条数据", len(s.SyncAccountBattleArmysInfo))
	}

	s.mkGVGInfo(p)
	s.mkHeroSwingAllInfo(p)

	s.mkHeroCompanionInfo(p)
	s.mkLimitGoodBuy(p)

	s.mkFriendInfo(p)
	s.makeExclusiveWeaponInfo(p)
	s.mkWspvpInfo(p)
	s.mkUpdateHeroInfo(p)
	s.mkSurplusGachaInfo(p)
	//
	// GS更新, 因为所有涉及到数据变更的操作都会走同步,
	// 所以这里来检查最大GS是否变更
	//
	//p.CheckChangeMaxGS()

	// By Fanyang 临时代码  T4711 服务器实现收集玩家信息的机制
	// 要校验领取过7日登陆，战队达到等级
	// check7DayGiftAndCorpLv(p)

	for i := 0; i < len(s.SyncHeroGs); i++ {
		s.SyncHeroGs[i] = p.Profile.Data.HeroGs[i]
	}
	s.SyncCorpGs = p.Profile.Data.CorpCurrGS
	s.SyncHeroBaseGSSum = p.Profile.Data.HeroBaseGSSum_Max

	if s.SyncOffRecoverNeed {
		tempHasRewards := p.Profile.OfflineRecoverInfo.HasRewards()
		s.OffRecoverRedpoint = tempHasRewards // 红点显示 是根据是否有奖励来判断
		s.HasOfflineRecovers = tempHasRewards || p.Profile.OfflineRecoverInfo.IsClientShow(p.Profile.GetProfileNowTime())
	}

	// debug info
	s.DebugTimeOffSet = p.Profile.DebugAbsoluteTime
	s.SyncTodayLoginTimes = p.Profile.LoginTodayNum

}

func (s *SyncResp) OnChangeBag() {
	s.SyncFullItemsNeed = true
}

func (s *SyncResp) OnChangeUpdateItems(item_inner_type int, uId uint32, oldCount int64, reason string) {
	switch item_inner_type {
	case helper.Item_Inner_Type_Jade:
		s.SyncUpdateJadeNeed = true
		if s.jades_update_o == nil {
			s.jades_update_o = map[uint32]int64{}
		}
		if s.jades_update == nil {
			s.jades_update = []uint32{}
		}

		if _, ok := s.jades_update_o[uId]; !ok {
			s.jades_update_o[uId] = oldCount
		}
		s.jades_update = append(s.jades_update, uId)
		s.jades_update_reason = reason
	case helper.Item_Inner_Type_Fashion:
		s.SyncFashionBagUpdateNeed = true
		if s.fashion_update_o == nil {
			s.fashion_update_o = map[uint32]int64{}
		}
		if s.fashion_update == nil {
			s.fashion_update = []uint32{}
		}

		if _, ok := s.fashion_update_o[uId]; !ok {
			s.fashion_update_o[uId] = oldCount
		}
		s.fashion_update = append(s.fashion_update, uId)
		s.fashion_update_reason = reason
	default:
		s.SyncUpdateItemsNeed = true
		if s.items_update_o == nil {
			s.items_update_o = map[uint32]int64{}
		}
		if s.items_update == nil {
			s.items_update = []uint32{}
		}

		if _, ok := s.items_update_o[uId]; !ok {
			s.items_update_o[uId] = oldCount
		}
		s.items_update = append(s.items_update, uId)
		s.items_update_reason = reason
	}
}

func (s *SyncResp) OnChangeDelItems(
	item_inner_type int,
	uId uint32,
	itemId string,
	oldCount int64,
	reason string) {
	switch item_inner_type {
	case helper.Item_Inner_Type_Jade:
		s.SyncDelJadeNeed = true
		if s.jades_del == nil {
			s.jades_del = []uint32{}
		}
		if s.jades_del_o == nil {
			s.jades_del_o = map[string]int64{}
		}
		s.jades_del = append(s.jades_del, uId)
		s.jades_del_o[itemId] = oldCount
		s.jades_del_reason = reason
	case helper.Item_Inner_Type_Fashion:
		s.SyncFashionBagDelNeed = true
		if s.fashion_del == nil {
			s.fashion_del = []uint32{}
		}
		if s.fashion_del_o == nil {
			s.fashion_del_o = map[string]int64{}
		}
		s.fashion_del = append(s.fashion_del, uId)
		s.fashion_del_o[itemId] = oldCount
		s.fashion_del_reason = reason
	default:
		s.SyncDelItemsNeed = true
		if s.items_del == nil {
			s.items_del = []uint32{}
		}
		if s.items_del_o == nil {
			s.items_del_o = map[string]int64{}
		}
		s.items_del = append(s.items_del, uId)
		s.items_del_o[itemId] = oldCount
		s.items_del_reason = reason
	}
}

func (s *SyncResp) OnChangeSC() {
	s.sc_need_sync = true
}

func (s *SyncResp) OnChangeHC() {
	s.hc_need_sync = true
}

func (s *SyncResp) OnChangeAvatarExp() {
	s.exps_need_sync = true
}

func (s *SyncResp) OnChangeAvatarArousal() {
	s.arousal_all_sync = true
}

func (s *SyncResp) OnChangeCorpExp() {
	s.corp_lv_need_sync = true
	s.OnChangeQuestAll()
}

func (s *SyncResp) OnChangeEnergy() {
	s.energy_need_sync = true
}

func (s *SyncResp) OnChangeStage(sid string) {
	s.stage_update = sid
}

func (s *SyncResp) OnChangeStageAll() {
	s.stage_update_all = true
	s.OnChangeQuestAll()
}

func (s *SyncResp) OnChangeChapter(ch string) {
	s.chapter_update = ch
}

func (s *SyncResp) OnChangeChapterAll() {
	s.chapter_update_all = true
}

func (s *SyncResp) OnChangeEquip() {
	s.equip_need_sync = true
}

func (s *SyncResp) OnChangeAvatarEquip() {
	s.avatar_equip_need_sync = true
}

func (s *SyncResp) OnChangeQuestAll() {
	s.quest_need_all = true
}

func (s *SyncResp) OnChangeMonthlyGiftStateChange() {
	s.SyncGiftMonthlyNeed = true
}

func (s *SyncResp) OnChangeGiftStateChange() {
	s.SyncGiftsNeed = true
}

func (s *SyncResp) OnChangeStoreAllChange() {
	s.store_all_sync = true
}

func (s *SyncResp) OnChangeShopAllChange() {
	s.shop_all_sync = true
}

func (s *SyncResp) OnChangeShopChange(shopId uint32) {
	s.shop_sync = true
	s.shop_id = shopId
}

func (s *SyncResp) OnChangeGachaAllChange() {
	s.gacha_all_sync = true
}

func (s *SyncResp) OnChangeSkillAllChange() {
	s.skill_need_sync = true
}

func (s *SyncResp) OnChangeGeneralAllChange() {
	s.general_need_sync = true
}

func (s *SyncResp) OnChangeGeneralRelAllChange() {
	s.general_rel_need_sync = true
}

func (s *SyncResp) OnChangeGeneralTeamAllChange() {
	s.general_team_need_sync = true
}

func (s *SyncResp) OnChangeGeneralQuest() {
	s.SyncGeneralQuestNeed = true
}

func (s *SyncResp) OnChangeMail() {
	s.SyncMailNeed = true
}

func (s *SyncResp) OnChangeLastStage() {
	s.last_stage_update = true
}

func (s *SyncResp) OnChangeVIP() {
	s.vip_need_sync = true

	// XXX by Fanyang 所有受到VIP等级影响的数据都要更新
	// by qiaozhu 优化了商店以后可以删掉了
	logs.Debug("[SyncResp] OnChangeVIP")
	// s.OnChangeStoreAllChange()
}

func (s *SyncResp) OnChangeBuy() {
	s.buy_need_sync = true
}

func (s *SyncResp) OnChangeBoss() {
	s.boss_fight_need_sync = true
}

func (s *SyncResp) OnChangeGameMode(gameModeId uint32) {
	s.SyncGameModeNeed = true
	s.gameMode_update = append(s.gameMode_update, gameModeId)
	s.SyncGiftMonthlyNeed = true // 前端需要再同步gamemode的时候更新月卡红点
}

func (s *SyncResp) OnChangeAllGameMode() {
	s.SyncGameModeNeed = true
	s.gameMode_update_all = true
	s.SyncGiftMonthlyNeed = true // 前端需要再同步gamemode的时候更新月卡红点
}

func (s *SyncResp) OnChangePrivilegeBuy() {
	s.SyncPrivilegeBuyNeed = true
}

func (s *SyncResp) OnChangeBaseData() {
	s.base_data_need_sync = true
}

func (s *SyncResp) OnChangeNewHand() {
	s.SyncNewHandNeed = true
}

func (s *SyncResp) OnChangeNewHandIgnoreNeed() {
	s.SyncNewHandIgnoreNeed = true
}

func (s *SyncResp) OnChangeBossFightPoint() {
	s.boss_fight_point_Sync = true
}

func (s *SyncResp) OnChangeSimplePvp() {
	s.pvp_count_need_sync = true
}

func (s *SyncResp) OnChangeActivityByCond() {
	s.act_gift_by_cond_need_sync = true
}

func (s *SyncResp) OnChangeActivityByTime() {
	s.act_gift_by_time_need_sync = true
}

func (s *SyncResp) OnChangeIAPGoodInfo() {
	s.SyncIAPGoodInfoNeed = true
}

func (s *SyncResp) OnChangeUnlockAvatar() {
	s.unlock_avatar_need_sync = true
}

func (s *SyncResp) OnChangeDestinyGeneral() {
	s.destiny_need_sync = true
}

func (s *SyncResp) OnChangeAvatarJade() {
	s.avatar_jade_need_sync = true
}

func (s *SyncResp) OnChangeDestinyGenJade() {
	s.destinyGen_jade_need_sync = true
}

func (s *SyncResp) OnChangeTrial() {
	s.SyncTrialNeed = true
}

func (s *SyncResp) OnChangeJadeFull() {
	s.SyncFullJadeNeed = true
}

func (s *SyncResp) OnChangeJadeUpdate() {
	s.SyncUpdateJadeNeed = true
}

func (s *SyncResp) OnChangeJadeDel() {
	s.SyncDelJadeNeed = true
}

func (s *SyncResp) OnChangeRecover() {
	s.SyncRecoverNeed = true
}

func (s *SyncResp) OnChangeDailyAward() {
	s.SyncDailyAwardNeed = true
}

func (s *SyncResp) OnChangeFashionBag() {
	s.SyncFashionBagAllNeed = true
}

func (s *SyncResp) OnChangeSevenDayRank() {
	s.SyncSevenDayRankNeed = true
}

func (s *SyncResp) OnChangeAccount7Day() {
	s.SyncAccountSevenDayNeed = true
}

func (s *SyncResp) OnChangeFirstPayReward() {
	s.SyncFirstPayRewardNeed = true
}

func (s *SyncResp) OnChangeHitEgg() {
	s.SyncHitEggNeed = true
}

func (s *SyncResp) OnChangeGrowFund() {
	s.SyncGrowFundNeed = true
}

func (s *SyncResp) OnChangeMarketActivity() {
	s.SyncMarketActivityNeed = true
}

func (s *SyncResp) OnChangeHeroTalent() {
	s.SyncHeroTalentNeed = true
}

func (s *SyncResp) OnChangeHeroSoul() {
	s.SyncHeroSoulNeed = true
}

func (s *SyncResp) OnChangeWantGeneralInfo() {
	s.SyncWantGeneralInfoNeed = true
}

func (s *SyncResp) OnChangeHeroTeam() {
	s.SyncHeroTeamNeed = true
}

func (s *SyncResp) OnChangeHeroGachaRace() {
	s.SyncHeroGachaRaceInfoNeed = true
}

func (s *SyncResp) OnChangeShareWeChat() {
	s.SyncShareWeChatNeed = true
}

func (s *SyncResp) OnChangeHeroSwing() {
	s.SyncHeroSwingNeed = true
}

func metric_send(id db.Account, typ, value string) {
	name := fmt.Sprintf("sync.%d.%d.%s.%s", id.GameId, id.ShardId, typ, "time")
	metrics.SimpleSend(name, value)
}

func (s *SyncResp) OnChangerExpeditionInfo() {
	s.SyncExpeditionInfoNeed = true
}

func (s *SyncResp) OnChangeCompanion() {
	s.SyncHeroCompanionNeed = true
}

func (s *SyncResp) OnChangeGVG() {
	s.SyncGVGNeed = true
}

func (s *SyncResp) onChangeLimitShop() {
	s.SyncBuyLimitShopNeed = true
}

func (s *SyncResp) onChangeMoneyCat() {
	s.MoneyCatNeed = true
}

func (s *SyncResp) OnChangeFriendList() {
	s.update_friend_list_sync = true
}

func (s *SyncResp) OnChangeBlackList() {
	s.update_black_list_sync = true
}

func (s *SyncResp) onChangeFestivalBossInfo() {
	s.SyncFestivalBossInfoNeed = true
}

func (s *SyncResp) onChangeExclusiveWeaponInfo() {
	s.SyncExclusiveWeaponNeed = true
}

func (s *SyncResp) OnChangeHeroDiff() {
	s.SyncHeroDiffNeed = true
}

func (s *SyncResp) OnChangeRedPacket7Days() {
	s.SyncRedPacket7DaysNeed = true
}

func (s *SyncResp) OnChangeExperienceLevel() {
	s.SyncExperienceLevelNeed = true
}

func (s *SyncResp) OnChangeWhiteGacha() {
	s.SyncWhiteGachaInfoNeed = true
}

func (s *SyncResp) OnChangeWheel() {
	s.SyncWheelInfoNeed = true
}

func (s *SyncResp) OnChangeWSPVP() {
	s.SyncWSPVPNeed = true
}

func (s *SyncResp) OnChangeFaceBook() {
	s.SyncFaceBookNeed = true
}

func (s *SyncResp) OnChangeTwitter() {
	s.SyncTwitterNeed = true
}

func (s *SyncResp) OnChangeLine() {
	s.SyncLineNeed = true
}

func (s *SyncResp) OnChangeOppoRelated() {
	s.SyncOppoNeed = true
}

func (s *SyncResp) OnChangeHeroDestiny() {
	s.SyncHeroDestinyNeed = true
}

func (s *SyncResp) OnChangeUpdateHeroStarLevel(avatarId int) {
	s.ChangedHeroAvatar = avatarId
	s.UpdateHeroStarLevelNeed = true
}

func (s *SyncResp) OnChangeUpdateHeroStarPiece(avatarId int) {
	s.ChangedHeroAvatar = avatarId
	s.UpdateHeroStarPieceNeed = true
}

func (s *SyncResp) OnChangeUpdateHeroLevel(avatarId int) {
	s.ChangedHeroAvatar = avatarId
	s.UpdateHeroLevelNeed = true
}

func (s *SyncResp) OnChangeUpdateHeroExp(avatarId int) {
	s.ChangedHeroAvatar = avatarId
	s.UpdateHeroExpNeed = true
}

func (s *SyncResp) OnChangeUpdateHeroSkill(avatarId int) {
	s.ChangedHeroAvatar = avatarId
	s.UpdateHeroSkillsNeed = true
}

func (s *SyncResp) OnChangeUpdateHeroWing(avatarId int) {
	s.ChangedHeroAvatar = avatarId
	s.UpdateHeroWingsNeed = true
}

func (s *SyncResp) OnChangeIAPGift() {
	s.IsIAPGift = true
}

func (s *SyncResp) OnChangeUpdateHeroCompanion(avatarId int) {
	s.ChangedHeroAvatar = avatarId
	s.UpdateHeroCompanionNeed = true
}

func (s *SyncResp) OnChangeUpdateHeroExclusive(avatarId int) {
	s.ChangedHeroAvatar = avatarId
	s.UpdateHeroExclusiveNeed = true
}

func (s *SyncResp) OnChangeFriendGift() {
	s.SyncFriendGiftListNeed = true
}

func (s *SyncResp) OnChangeReceiveGiftList(refresh bool) {
	s.SyncReceiveGiftListNeed = true
}

func (s *SyncResp) OnChangeHeroStarMap() {
	s.SyncHeroStarMapNeed = true
}

func (s *SyncResp) OnChangeBindMailReward() {
	s.SyncBindMailRewardNeed = true
}

func (s *SyncResp) OnChangeMagicPetInfo() {
	s.SyncMagicPetInfo = true
}

func (s *SyncResp) OnChangeBattleArmyInfo() {
	s.SyncBattleArmyInfo = true
}

func (s *SyncResp) OnChangeOfflineRecover() {
	s.SyncOffRecoverNeed = true
}
