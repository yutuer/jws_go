package modules

const (
	Module_DataVer       = "data_ver"
	Module_DataHotUpdate = "datahotupdate"
	Module_HourLog       = "hour_log"
	Module_PlayerMsg     = "player_msg"
	Module_MailSender    = "mail_sender"
	Module_GlobalMail    = "global_mail"
	Module_GlobalCount   = "global_count"
	Module_GlobalInfo    = "global_info"
	Module_CrossService  = "crossservice"
	//Module_Warm                = "warm"
	Module_TitleRank           = "titlerank"
	Module_Balance             = "balance"
	Module_Rank                = "rank"
	Module_RedeemCode          = "redeem_code"
	Module_CityFish            = "cityfish"
	Module_DestingGeneralFirst = "desting_general_first"
	Module_FestivalBoss        = "FestivalBoss"
	Module_Friend              = "friend"
	Module_GateEnemy           = "gates_enemy"
	Module_Guild               = "guild"
	Module_Gve                 = "gve"
	Module_GvG                 = "gvg"
	Module_HeroGachaRace       = "herogacherace"
	Module_MoneyCat            = "moneycat"
	Module_Room                = "roomsmng"
	Module_SPvpRander          = "simple_pvp_rander"
	Module_TeamPvp             = "teampvp"
	Module_WantGenBest         = "want_general_best"
	Module_Worship             = "worship"
	Module_HeroDiff            = "hero_diff"
	Module_WsPvp               = "ws_pvp"
	Module_MarketActivity      = "market_activity"
	Module_CSRob               = "csrob"
)

/*
	modules启动按照server_modules_seq正序，关闭反顺序
*/
var (
	server_modules_seq = []string{
		Module_DataVer,
		Module_DataHotUpdate,
		Module_HourLog,
		Module_PlayerMsg,
		Module_MailSender,
		Module_GlobalMail,
		Module_GlobalCount,
		Module_GlobalInfo,
		//Module_Warm,

		Module_CrossService,

		Module_TitleRank,
		Module_Balance,
		Module_Rank,

		Module_RedeemCode,
		Module_CityFish,
		Module_DestingGeneralFirst,
		Module_FestivalBoss,
		Module_Friend,
		Module_GateEnemy,
		Module_Guild,
		Module_Gve,
		Module_GvG,
		Module_HeroGachaRace,
		Module_MoneyCat,
		Module_Room,
		Module_SPvpRander,
		Module_TeamPvp,
		Module_WantGenBest,
		Module_Worship,
		Module_HeroDiff,
		Module_WsPvp,
		Module_MarketActivity,

		Module_CSRob,
	}
)
