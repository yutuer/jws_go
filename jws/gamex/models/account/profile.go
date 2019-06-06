package account

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"time"

	"encoding/base64"
	"errors"

	"reflect"

	"vcs.taiyouxi.net/jws/gamex/models/astrology"
	"vcs.taiyouxi.net/jws/gamex/models/battlearmy"
	"vcs.taiyouxi.net/jws/gamex/models/bossfight"
	"vcs.taiyouxi.net/jws/gamex/models/buy"
	"vcs.taiyouxi.net/jws/gamex/models/clienttag"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/currency"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/fashion"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/item_history"
	"vcs.taiyouxi.net/jws/gamex/models/mail"
	"vcs.taiyouxi.net/jws/gamex/models/market_activity"
	"vcs.taiyouxi.net/jws/gamex/models/pay"
	"vcs.taiyouxi.net/jws/gamex/models/world_boss"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Profile should save to profile:{profileid}
type Profile struct {
	//profile:{profileid}
	dbkey        db.ProfileDBKey
	dirtiesCheck map[string]interface{}
	Ver          int64 `redis:"version"`

	Name        string `redis:"name"`
	Level       int    `redis:"lvl"`
	RenameCount int    `redis:"rename_count"`

	// 客户端版本，数据版本，设备等信息
	//PlayerGroup  string `redis:"group"`
	ClientInfo  ClientInfo `redis:"client_inf"`
	AccountName string     `redis:"account_name"` // 账户名
	DeviceId    string     `redis:"device_id"`
	MemSize     string     `redis:"MemSize"`
	PlatformId  string     `redis:"PlatformId"`
	DeviceToken string     `redis:"DeviceToken"`
	IDFA        string     `redis:"IDFA"`
	// 玩家当前角色
	CurrAvatar int `redis:"curr_avatar"`
	// 是否隐藏神翼
	IsHideSwing bool `redis:"show_swing"`

	// 玩家时间，这个时间在正常环境下应该是服务器时间，
	// 考虑到Debug时，我们需要修改当前时间，这个操作会引起一些问题
	// 这里加上一个时间值，用作各种判定的依据，求改其以实现Debug
	debug_time_start time.Time //设置时的服务器时间
	debug_from_time  time.Time //设置的起始时间

	time_start int64 //设置时的服务器时间
	from_time  int64 //设置的起始时间

	OnlineTimeCurrDay            int64 `redis:"onlinet"`
	OnlineTimeCurrDayLastSetTime int64 `redis:"onlinetl"`

	client_market_ver string // 客户端版本，用来更新dataver用的，不用存档

	Stage StageInfo `redis:"stage"`
	/*
		玩家体力值
	*/
	Energy PlayerEnergy `redis:"energy"`
	/*
		玩家军令值
	*/
	BossFightPoint BossFightPoint `redis:"bossfp"`
	/*
		玩家软通
	*/
	SC currency.SoftCurrency `redis:"sc"`

	/*
		玩家硬通
	*/
	HC currency.HardCurrency `redis:"hc"`

	/*
		玩家武将经验值
	*/
	AExp AvatarExp `redis:"avatarExp"`

	/*
		战队装备信息, 包括普通装备和公会官阶装备
		当前装备信息，因为装备位比较固定，所以用数组
		整个装备信息是一个装备Id的数组，数组idx表示slot，即装备栏位
		现阶段暂时定义装备栏位个数为10个
		所以curr_equips中slot 0-9表示0号角色的装备
	*/
	Equips Equips `redis:"equips"`

	/*
		武将装备信息, 目前只有时装在里面
		当前装备信息，因为装备位比较固定，所以用数组
		整个装备信息是一个装备Id的数组，数组idx表示slot，即装备栏位
		这里不分角色统一存储，通过slot区分各个角色
		现阶段暂时定义装备栏位个数为5个
		所以curr_equips中slot 0-5表示0号角色的装备
		slot 6-10表示1号角色的装备
	*/
	AEquips AvatarEquips `redis:"avatar_equips"`

	/*
		战队信息，包括玩家个人等级
	*/
	CorpInf Corp `redis:"corp"`

	/*
		统一的随机数生成器
	*/
	Rng util.Kiss64Rng `redis:"rng"`

	/*
		任务信息
	*/
	questInMem PlayerQuest
	Quest      playerQuestInDB `redis:"quest"`

	/*
		条件汇总信息，仅在内存中，初始化时根据其他系统创建
	*/
	conditons gamedata.PlayerCondition

	/*
		月签到信息
	*/
	GiftMonthly gamedata.ActivityGiftMonthly `redis:"gift_m"`

	/*
		签到信息
	*/
	Gifts gamedata.ActivityGiftDailys `redis:"gift_a"`

	/*
		Gacha
	*/
	Gacha PlayerGacha `redis:"gacha"`

	/*
		Skill
	*/
	Skill AvatarSkill `redis:"skill"`

	mail mail.PlayerMail `redis:"mail"`

	Data ProfileData `redis:"data"`

	RedeemCodes playerRedeemCodeTypHasToken `redis:"redeem"`

	/*
		VIP 信息
	*/
	Vip VIP `redis:"v"`

	/*
		购买信息
	*/
	Buy buy.PlayerBuy `redis:"buy"`

	/*
		世界Boss
	*/
	Boss bossfight.PlayerBoss `redis:"boss"`

	/*
		活动信息
	*/
	GameMode GameModesInfo `redis:"game_m"`

	/*
		特权购买
	*/
	PrivilegeBuy PrivilegeBuyInfo `redis:"prvlg_b"`

	/*

		新手引导步骤，gzip压缩并base64
	*/

	NewHandIgnore       bool `redis:"newhandignore"`
	newHandStep         string
	NewHandStep_b       string `redis:"newhand_b"`
	LastClientTimeEvent string `redis:"client_time_event"`

	/*
		天命
	*/
	Story PlayerStory `redis:"story"`

	/*
		洗练回退数据,不存数据库
	*/
	AbstractInfo PlayerAbstractCancelInfo `redis:"abstract"`

	/*
		Pvp
	*/
	SimplePvp PlayerSimplePvp `redis:"simpvp"`

	/*
		活动条件奖励
	*/
	ActGiftByConds PlayerActGiftByConds `redis:"act_giftcond"`
	ActGiftByTime  PlayerActGiftByTime  `redis:"act_gifttime"`

	/*
		龙玉
	*/
	JadeBag      PlayerJadeBagDB `redis:"jade_bag"`
	jadeBagInMem PlayerJadeBag

	EquipJades EquipmentJades `redis:"av_jades"`

	DGJades DestGeneralJades `redis:"dg_jades"`

	/*
		时装
	*/
	FashionBag fashion.PlayerFashionBagDB `redis:"fashion_bag"`
	fashionBag fashion.PlayerFashionBag

	/*
		战阵
	*/
	BattleArmys battlearmy.BattleArmys `redis:"battle_armys"`

	/*
		IAP Good info
	*/
	IAPGoodInfo pay.PayGoodInfos `redis:"iap_good_info"`

	/*
		ClientTag
	*/
	ClientTagInfo clienttag.ClientTag `redis:"client_tag"`

	/*
		IAP 信息
	*/
	IAPInfos pay.PlayerPayInfo `redis:"iap_info"`

	LoginTime     int64 `redis:"logintime"`
	LogoutTime    int64 `redis:"logouttime"`
	LoginTimes    int64 `redis:"logintimes"`
	LoginTodayNum int   `redis:"logintodaytimes"`
	Logindaynum   int   `redis:"logindaynum"` // bilog userinfo 用

	LastGetProtoTS int64 `redis:"lgp"`
	LastGetInfoTS  int64 `redis:"lgi"`

	/*
		pay支付，48小时内有效
	*/
	PayTime int64 `redis:"paytime"`

	/*
		first pay奖励记录
	*/
	FirstPayReward PlayerFirstPayReward `redis:"firstpay1"`

	/*
		game mode counts
	*/
	Counts counter.PlayerCounter `redis:"counts"`

	/*
		神将系统
	*/
	DestinyGenerals PlayerDestinyGeneral `redis:"destinys"`

	/*
		爬塔
	*/
	Trial PlayerTrial `redis:"trial"`

	/*
		兵临城下
	*/
	GatesEnemy playerGatesEnemyData `redis:"genemy"`

	/*
		工会boss
	*/
	guildBoss playerGuildBossData `redis:"gboss"`

	/*
		奖励追回
	*/
	Recover PlayerRecover `redis:"recover"`

	/*
		日常领奖
	*/
	DailyAwards PlayerDailyAwards `redis:"dailyaward"`

	/*
		手机绑定
	*/
	Phone PlayerPhoneData `redis:"phones"`

	/*
		切磋
	*/
	Gank PlayerGank `redis:"gank"`

	/*
		账号7天活动
	*/
	player7DayInMem Account7Day
	Player7Day      Account7DayInDB `redis:"acid7day"`

	/*
		玩家道具掉落历史
	*/
	PlayerItemHistory itemHistory.ItemHistory `redis:"history"`

	/*
		3v3竞技场
	*/
	TeamPvp PlayerTeamPvp `redis:"teampvp"`

	/*
		hero
	*/
	Hero PlayerHero `redis:"hero"`

	/*
		hero teams
	*/
	HeroTeams PlayerHeroTeams `redis:"heroteam"`

	/*
		heroTalent
	*/
	HeroTalent PlayerHeroTalent `redis:"herotalent"`

	/*
		hero soul
	*/
	HeroSoul PlayerHeroSoul `redis:"herosoul"`

	/*
		astrology
	*/
	Astrology *astrology.Astrology `redis:"astrology"`

	/*
		hit egg
	*/
	HitEgg PlayerHitEgg `redis:"hitegg"`

	/*
		title
	*/
	titleMem PlayerTitle
	Title    PlayerTitleInDB `redis:"title"`

	/*
		grow fund
	*/
	GrowFund PlayerGrowFund `redis:"growfund"`

	FirstPassRank PlayerFirstPassRankReward `redis:"firstpassr"`

	/*
		运营活动
	*/
	MarketActivitys market_activity.PlayerMarketActivitys `redis:"marketactivity"`
	/*
		单次最大吃包子数目
	*/
	EatBaoziInfo EatBaoziInfo `redis:"eatbaozi_info"`

	/*
		我要名将信息
	*/
	WantGeneralInfo WantFamousGeneralInfo `redis:"want_general_info`

	/*
		限时神将信息
	*/
	HeroGachaRaceInfo HeroGachaRaceInfo `redis:"hero_gacharace_info`

	/*
		微信分享信息
	*/
	ShareWeChatInfo ShareWeChatInfo `redis: "share_wechat_info"`

	/*
		招财猫招财次数
	*/
	MoneyCatInfo MoneyCatInfo `redis:"moneycat_info"`

	/*
		远征信息
	*/
	ExpeditionInfo ExpeditionInfo `redis:"expedition_info"`

	/*
		节日Boss信息
	*/
	FestivalBossInfo FestivalBossInfo `redis:"festival_info"`

	/*
		武将差异化，出奇制胜
	*/
	HeroDiff HeroDiff

	/*
		开服7日红包
	*/
	Redpacket7day RedPacket7Days

	/*
		白盒宝箱
	*/
	WhiteGachaInfo WhiteGachaInfo

	/*
		名将体验关卡
	*/
	ExperienceLevel ExperienceLevel

	/*
		无双争霸
	*/
	WSPVPPersonalInfo WSPVPPersonalInfo `redis:"wspvp_info"`

	/*
		client setting
	*/
	SystemLanguage string `redis:"sys_lang"` // 客户端系统语言

	/*
		OPPO 相关
	*/
	OppoRelated OppoRelated `redis:"oppo_related"`
	/*
		宿命相关
	*/
	HeroDestiny HeroDestiny `redis:"hero_destiny"`
	/*
		黑盒宝箱
	*/
	BlackGachaInfo BlackGachaInfo
	/*
		离线资源找回
	*/
	OfflineRecoverInfo OfflineRecoverInfo `redis:"offline_recover_info"`

	//facebook是否绑定
	IsFaceBook bool

	//twitter是否分享过
	IsTwitterShared bool `redis:"is_twitter_shared"`

	//line是否分析过
	IsLineShared bool `redis:"is_line_shared"`

	//存档创建时间，就是存档创建的时间，在newprofile里初始化，在load时在覆盖一次
	CreateTime int64 `redis:"createtime"`
	// 渠道信息
	ChannelQuickId string `redis:"channel_quick"`
	ChannelId      string `redis:"channel"`

	DebugAbsoluteTime int64 `redis:"debugAbsoluteTime"`

	// 为bilog用
	lastLogicLog       string
	lastClientEvent    string
	LastBIHourActiveTS int64 `redis:"bihour_ats"`

	// pay feed back
	GotPayFeedBack bool `redis:"payfeedback"`

	/*
		绑定邮箱领奖
	*/
	BindMailRewardInfo BindMailRewardInfo `redis:"bind_mail_rw"`
	/*
		世界boss
	*/
	WorldBossData world_boss.WorldBossData `redis:"world_boss"`
	/*
		武将多余碎片处理
	*/
	HeroSurplusInfo HeroSurplusInfo `redis:"hero_surplus_info"`
	/*
		组队boss宝箱信息
	*/
	TeamBossStorageInfo TeamBossStorageInfo `redis:"team_boss_storage_info"`
	/*
		组队boss队伍信息
	*/
	TeamBossTeamInfo TeamBossTeamInfo `redis:"team_boss_leave_info"`
	/*
		韩国礼包使用情况
	*/
	Koreapackget KoreaPackget `redis:"korea_package_info"`
	/*
		港澳台运营活动信息
	*/
	hmtActivityInfoInMem HmtPlayerActivityInfo
	HmtActivityInfo      HmtPlayerActivityInfoInDB `redis:"hmt_activity_info"`
	/*
		幸运轮盘
	*/
	LuckyWheelInfo LuckyWheel `redis:"lucky_wheel_info"`
}

func NewProfile(account db.Account) Profile {
	now_t := time.Now().Unix()
	p := Profile{
		dbkey: db.ProfileDBKey{
			Account: account,
			Prefix:  "profile",
		},
		//Ver:        helper.CurrDBVersion,
		CreateTime: now_t,
	}

	return p
}

const (
	new_hand_max_size = 1024 * 5
)

func (p *Profile) SetNewHand(tid string) error {
	if len(tid) > new_hand_max_size {
		return errors.New(fmt.Sprintf("newhand size too large %d", tid))
	}
	// 压缩
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	defer w.Close()
	w.Write([]byte(tid))
	w.Flush()

	p.newHandStep = tid
	p.NewHandStep_b = base64.StdEncoding.EncodeToString(b.Bytes())
	return nil
}

func (p *Profile) GetNewHand() string {
	if p.newHandStep == "" {
		if len(p.NewHandStep_b) > 0 {
			bb, err := base64.StdEncoding.DecodeString(p.NewHandStep_b)
			if err != nil {
				logs.Error("sync new hand decode b64 err :", err)
			} else {
				var b bytes.Buffer
				b.Write(bb)
				r, _ := gzip.NewReader(&b)
				defer r.Close()
				unzipbb, _ := ioutil.ReadAll(r)
				p.newHandStep = string(unzipbb)
			}
		}
	}
	return p.newHandStep
}

func (p *Profile) AppendNewHand(tid string) error {
	return p.SetNewHand(fmt.Sprintf("%s&%s", p.GetNewHand(), tid))
}

func (p *Profile) SetNewHandIgnore(set bool) {
	p.NewHandIgnore = set
}

func (p *Profile) GetStage() *StageInfo {
	return &p.Stage
}

func (p *Profile) GetAstrology() *astrology.Astrology {
	return p.Astrology
}

func (p *Profile) DBName() string {
	return p.dbkey.String()
}

func (p *Profile) DBSave(cb redis.CmdBuffer, forceDirty bool) error {
	p.Quest = p.questInMem.ToDB()
	p.JadeBag = p.jadeBagInMem.ToDB()
	p.FashionBag = p.fashionBag.ToDB()
	p.Player7Day = p.player7DayInMem.ToDB()
	p.Title = p.titleMem.ToDB()
	p.HmtActivityInfo = p.hmtActivityInfoInMem.ToDB()

	key := p.DBName()

	if forceDirty {
		p.dirtiesCheck = nil
	}
	err, newDirtyCheck, chged := driver.DumpToHashDBCmcBufferCheckDirty(
		cb, key, p, p.dirtiesCheck)
	if err != nil {
		return err
	}
	if !game.Cfg.IsRunModeProd() {
		if !reflect.DeepEqual(p.dirtiesCheck, newDirtyCheck) {
			logs.Trace("Save Profile Data, NickName: %s, AccountName: %s  %v",
				p.Name, p.AccountName, chged)
		} else {
			logs.Trace("Save Profile Data Clean, NickName: %s, AccountName: %s", p.Name, p.AccountName)
		}
	}
	p.dirtiesCheck = newDirtyCheck
	return nil
}

func (p *Profile) DBLoad(logInfo bool) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	key := p.DBName()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, p, false, logInfo)

	// RESTORE_ERR_Profile_No_Data 表示玩家第一次登陆游戏，没有存档，这不视为Bug
	// 外面的逻辑需要根据此判断是否是第一次登陆游戏
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}

	p.dirtiesCheck = driver.GenDirtyHash(p)

	p.OnAfterLogin()

	return err
}

// 登陆后一些初始化工作, 与afterAccountLogin不同, 是结构性的
// 在加载数据库之后立即执行
func (p *Profile) OnAfterLogin() {
	// 一些初始化工作
	p.questInMem.FromDB(&p.Quest)
	p.jadeBagInMem.FromDB(&p.JadeBag)
	p.fashionBag.FromDB(&p.FashionBag)
	p.player7DayInMem.FromDB(&p.Player7Day)
	p.titleMem.FromDB(&p.Title)
	p.hmtActivityInfoInMem.FromDB(&p.HmtActivityInfo)
	p.Energy.player = p
	p.BossFightPoint.player = p
	p.GetAvatarExp().SetAccount(p)
	p.GetCorp().player = p

	// Conditons只存在内存中，这里初始化
	p.conditons.Init()
	p.Gifts.Update(p.GetProfileNowTime())

	p.mail.OnAfterLogin(p.DBName())
	p.GetIAPInfo().Init()

	p.PlayerItemHistory.Init()

	if nil == p.Astrology {
		p.Astrology = astrology.NewAstrology()
	}
	p.Astrology.AfterLogin()
	p.GetHmtActivityInfo().IsLogin(p.GetProfileNowTime())

}

func (p *Profile) GetLastSetCurLogicLog(typ string) string {
	old := p.lastLogicLog
	if old == "" {
		old = "root"
	}
	p.lastLogicLog = typ
	return old
}
func (p *Profile) GetLastSetCurClientEvent(typ string) string {
	old := p.lastClientEvent
	if old == "" {
		old = "root"
	}
	p.lastClientEvent = typ
	return old
}

func (p *Profile) GetClientMarketVer() string {
	return p.client_market_ver
}

func (p *Profile) SetClientMarketVer(market_ver string) {
	p.client_market_ver = market_ver
}

func (p *Profile) GetEnergy() *PlayerEnergy {
	return &p.Energy
}

func (p *Profile) GetSC() *currency.SoftCurrency {
	return &p.SC
}

func (p *Profile) GetHC() *currency.HardCurrency {
	return &p.HC
}

func (p *Profile) GetAvatarExp() *AvatarExp {
	return &p.AExp
}

func (p *Profile) GetEquips() *Equips {
	return &p.Equips
}

func (p *Profile) GetAvatarEquips() *AvatarEquips {
	return &p.AEquips
}

func (p *Profile) GetCorp() *Corp {
	return &p.CorpInf
}

func (p *Profile) GetQuest() *PlayerQuest {
	return &p.questInMem
}

func (p *Profile) GetCondition() *gamedata.PlayerCondition {
	return &p.conditons
}

func (p *Profile) GetGiftMonthly() *gamedata.ActivityGiftMonthly {
	return &p.GiftMonthly
}

func (p *Profile) GetGifts() *gamedata.ActivityGiftDailys {
	return &p.Gifts
}

func (p *Profile) GetGacha(idx int) *GachaState {
	return p.Gacha.GetGachaStat(idx)
}

func (p *Profile) GetCurrAvatar() int {
	return p.CurrAvatar
}

func (p *Profile) GetAvatarSkill() *AvatarSkill {
	return &p.Skill
}

func (p *Profile) GetMails() *mail.PlayerMail {
	return &p.mail
}

func (p *Profile) GetData() *ProfileData {
	return &p.Data
}

func (p *Profile) GetVip() *VIP {
	return &p.Vip
}

func (p *Profile) GetBuy() *buy.PlayerBuy {
	return &p.Buy
}

func (p *Profile) GetVipLevel() uint32 {
	return p.GetVip().V
}

func (p *Profile) GetMyVipCfg() *gamedata.VIPConfig {
	return gamedata.GetVIPCfg(int(p.GetVip().V))
}

func (p *Profile) GetBoss() *bossfight.PlayerBoss {
	return &p.Boss
}

func (p *Profile) GetGameMode() *GameModesInfo {
	return &p.GameMode
}

func (p *Profile) GetPrivilegeBuy() *PrivilegeBuyInfo {
	return &p.PrivilegeBuy
}

func (p *Profile) GetStory() *PlayerStory {
	return &p.Story
}

func (p *Profile) GetBossFightPoint() *BossFightPoint {
	return &p.BossFightPoint
}

func (p *Profile) GetAbstractCancelInfo() *PlayerAbstractCancelInfo {
	return &p.AbstractInfo
}

func (p *Profile) GetSimplePvp() *PlayerSimplePvp {
	return &p.SimplePvp
}

func (p *Profile) GetRedeemCode() *playerRedeemCodeTypHasToken {
	return &p.RedeemCodes
}

func (p *Profile) GetActGiftByCond() *PlayerActGiftByConds {
	return &p.ActGiftByConds
}

func (p *Profile) GetActGiftByTime() *PlayerActGiftByTime {
	return &p.ActGiftByTime
}

func (p *Profile) GetIAPGoodInfo() *pay.PayGoodInfos {
	return &p.IAPGoodInfo
}

func (p *Profile) GetIAPInfo() *pay.PlayerPayInfo {
	return &p.IAPInfos
}

func (p *Profile) GetClientTagInfo() *clienttag.ClientTag {
	return &p.ClientTagInfo
}

func (p *Profile) GetCounts() *counter.PlayerCounter {
	return &p.Counts
}

func (p *Profile) GetEquipJades() *EquipmentJades {
	return &p.EquipJades
}

func (p *Profile) GetDestGeneralJades() *DestGeneralJades {
	return &p.DGJades
}

func (p *Profile) GetDestinyGeneral() *PlayerDestinyGeneral {
	return &p.DestinyGenerals
}

func (p *Profile) GetPlayerTrial() *PlayerTrial {
	return &p.Trial
}

func (p *Profile) GetGatesEnemy() *playerGatesEnemyData {
	return &p.GatesEnemy
}

func (p *Profile) GetJadeBag() *PlayerJadeBag {
	return &p.jadeBagInMem
}

func (p *Profile) GetFashionBag() *fashion.PlayerFashionBag {
	return &p.fashionBag
}

func (p *Profile) GetRecover() *PlayerRecover {
	return &p.Recover
}

func (p *Profile) GetDailyAwards() *PlayerDailyAwards {
	return &p.DailyAwards
}

func (p *Profile) GetPhone() *PlayerPhoneData {
	return &p.Phone
}

func (p *Profile) GetGank() *PlayerGank {
	return &p.Gank
}

func (p *Profile) GetAccount7Day() *Account7Day {
	return &p.player7DayInMem
}

func (p *Profile) GetItemHistory() *itemHistory.ItemHistory {
	return &p.PlayerItemHistory
}

func (p *Profile) GetTeamPvp() *PlayerTeamPvp {
	return &p.TeamPvp
}

func (p *Profile) GetHero() *PlayerHero {
	return &p.Hero
}

func (p *Profile) GetHeroTeams() *PlayerHeroTeams {
	return &p.HeroTeams
}

func (p *Profile) GetHitEgg() *PlayerHitEgg {
	return &p.HitEgg
}

func (p *Profile) GetTitle() *PlayerTitle {
	return &p.titleMem
}

func (p *Profile) GetGrowFund() *PlayerGrowFund {
	return &p.GrowFund
}

func (p *Profile) GetFirstPassRank() *PlayerFirstPassRankReward {
	return &p.FirstPassRank
}

func (p *Profile) GetMarketActivitys() *market_activity.PlayerMarketActivitys {
	return &p.MarketActivitys
}

func (p *Profile) GetHeroTalent() *PlayerHeroTalent {
	return &p.HeroTalent
}

func (p *Profile) GetHeroSoul() *PlayerHeroSoul {
	return &p.HeroSoul
}

func (p *Profile) GetEatBaozi() *EatBaoziInfo {
	return &p.EatBaoziInfo
}

func (p *Profile) GetWantGeneralInfo() *WantFamousGeneralInfo {
	return &p.WantGeneralInfo
}

func (p *Profile) GetHeroGachaRaceInfo() *HeroGachaRaceInfo {
	return &p.HeroGachaRaceInfo
}

func (p *Profile) GetShareWeChatInfo() *ShareWeChatInfo {
	return &p.ShareWeChatInfo
}

func (p *Profile) GetMoneyCatInfo() *MoneyCatInfo {
	return &p.MoneyCatInfo
}

func (p *Profile) GetExpeditionInfo() *ExpeditionInfo {
	return &p.ExpeditionInfo
}

func (p *Profile) GetFestivalBossInfo() *FestivalBossInfo {
	return &p.FestivalBossInfo
}

func (p *Profile) GetGuildBossInfo() *playerGuildBossData {
	return &p.guildBoss
}

func (p *Profile) GetHeroDiff() *HeroDiff {
	return &p.HeroDiff
}

func (p *Profile) GetRedPacket7day() *RedPacket7Days {
	return &p.Redpacket7day
}

func (p *Profile) GetWhiteGachaInfo() *WhiteGachaInfo {
	return &p.WhiteGachaInfo
}

func (p *Profile) GetWheelGachaInfo() *LuckyWheel {
	return &p.LuckyWheelInfo
}

func (p *Profile) GetExperienceLevelInfo() *ExperienceLevel {
	return &p.ExperienceLevel
}

func (p *Profile) GetWSPVPInfo() *WSPVPPersonalInfo {
	return &p.WSPVPPersonalInfo
}

func (p *Profile) GetOppoRelated() *OppoRelated {
	return &p.OppoRelated
}

func (p *Profile) GetHeroDestiny() *HeroDestiny {
	return &p.HeroDestiny
}

func (p *Profile) GetBindMailRewardInfo() *BindMailRewardInfo {
	return &p.BindMailRewardInfo
}

func (p *Profile) GetWorldBossData() *world_boss.WorldBossData {
	return &p.WorldBossData
}

func (p *Profile) GetHeroSurplusInfo() *HeroSurplusInfo {
	return &p.HeroSurplusInfo
}

func (p *Profile) GetTeamBossStorageInfo() *TeamBossStorageInfo {
	return &p.TeamBossStorageInfo
}

func (p *Profile) GetTeamBossTeamInfo() *TeamBossTeamInfo {
	return &p.TeamBossTeamInfo
}

func (p *Profile) GetHmtActivityInfo() *HmtPlayerActivityInfo {
	return &p.hmtActivityInfoInMem
}

func (p *Profile) GetBattleArmyInfo() *battlearmy.BattleArmys {
	return &p.BattleArmys
}
