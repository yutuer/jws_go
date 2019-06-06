package logiclog

import (
	"fmt"

	"strconv"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	LogicTag_Login                = "Login"
	LogicTag_Create               = "CreateProfile"
	LogicTag_CreateRole           = "CreateRole"
	LogicTag_Logout               = "Logout"
	LogicTag_GiveItem             = "GiveItem"
	LogicTag_CostItem             = "CostItem"
	LogicTag_GiveItemUseSelf      = "GiveItemUseSelf"
	LogicTag_CostItemUseSelf      = "CostItemUseSelf"
	LogicTag_GiveCurrency         = "GiveCurrency"
	LogicTag_CostCurrency         = "CostCurrency"
	LogicTag_CorpExpChg           = "CorpExpChg"
	LogicTag_CorpLevelChg         = "CorpLevelChg"
	LogicTag_QuestFinish          = "QuestFinish"
	LogicTag_StageFinish          = "StageFinish"
	LogicTag_StoreBuy             = "StoreBuy"
	LogicTag_ShopBuy              = "ShopBuy"
	LogicTag_Tutorial             = "Tutorial"
	LogicTag_Gacha                = "Gacha"
	LogicTag_GeneralAddNum        = "GeneralAddNum"
	LogicTag_GeneralStarLvlUp     = "GeneralStarLvlUp"
	LogicTag_GeneralRelLvlUp      = "GeneralRelLvlUp"
	LogicTag_GeneralQuestRec      = "GeneralQuestReceive"
	LogicTag_GeneralQuestFinish   = "GeneralQuestFinish"
	LogicTag_Pvp                  = "Pvp"
	LogicTag_IAPTry               = "IAPTry"
	LogicTag_IAP                  = "IAP"
	LogicTag_GuildCreate          = "GuildCreate"
	LogicTag_GuildAddMem          = "GuildAddMem"
	LogicTag_GuildDelMem          = "GuildDelMem"
	LogicTag_GuildDismiss         = "GuildDismiss"
	LogicTag_GuildPosChg          = "GuildPosChg"
	LogicTag_GuildGateEnemyOver   = "GuildGateEnemyOver"
	LogicTag_GuildBossFight       = "GuildBossFight"
	LogicTag_GuildBoss            = "GuildBoss"
	LogicTag_AddGuildInventory    = "AddGuildInventory"
	LogicTag_AssignGuildInventory = "AssignGuildInventory"
	LogicTag_GuildLvUp            = "GuildLvUp"
	LogicTag_RedeemCode           = "RedeemCode"
	LogicTag_TrialLvlFinish       = "TrialLvlFinish"
	LogicTag_TrialReset           = "TrialReset"
	LogicTag_TrialSweep           = "TrialSweep"
	LogicTag_Phone                = "Phone"
	LogicTag_PveBoss              = "PveBoss"
	LogicTag_EquipAbstract        = "EquipAbstract"
	LogicTag_EquipAbstractCancel  = "EquipAbstractCancel"
	LogicTag_EquipStarUp          = "EquipStarLvlUp"
	LogicTag_EquipMatEnhAdd       = "EquipMatEnhAdd"
	LogicTag_EquipMatEnhLvlUp     = "EquipMatEnhLvlUp"
	LogicTag_Fish                 = "Fish"
	LogicTag_Gank                 = "Gank"
	LogicTag_TeamPvp              = "TeamPvp"
	LogicTag_GveStartMatch        = "GevStartMatch"
	LogicTag_GveCancelMatch       = "GevCancelMatch"
	LogicTag_GevGameStart         = "GevGameStart"
	LogicTag_GevGameStop          = "GevGameStop"
	LogicTag_HeroStarUp           = "HeroStarUp"
	LogicTag_HeroAddPiece         = "HeroAddPiece"
	LogicTag_HeroUnlock           = "HeroUnlock"
	LogicTag_Account7Point        = "Account7Point"
	LogicTag_DailyPoint           = "DailyPoint"
	LogicTag_GuildRank            = "GuildRank"
	LogicTag_HitEgg               = "HitEgg"
	LogicTag_PayFeedBack          = "PayFeedBack"
	LogicTag_FirstPay             = "FirstPay"

	LogicTag_Act_DestinyGeneral = "ActDestinyGeneral"
	LogicTag_HeroSoulLvUp       = "HeroSoulUlvUp"
	LogicTag_TalentLvUp         = "TalentLvUp"
	LogicTag_TitleChange        = "TitleChange"

	LogicTar_ExpeditionEvent     = "ExpeditionEvent"
	LogicTar_ExpeditionRest      = "ExpeditionRest"
	LogicTag_ClientEvent         = "ClientEvent"
	LogicTag_ClientTimeEvent     = "ClientTimeEvent"
	LogicTag_ShareWeChat         = "ShareWeChat"
	LogicTar_GachaRank           = "GachaRank"
	LogicTar_PassiveSkillAdd     = "PassiveEvent"
	LogicTag_HeroWingStarLvUp    = "HeroWingStarLevelUp"
	LogicTag_HeroWingLvUp        = "HeroWingLevelUp"
	LogicTag_HeroWingReset       = "HeroWingReset"
	LogicTag_HeroWingAct         = "HeroWingAct"
	LogicTag_HeroWingGetProp     = "HeroWingGetProp"
	LogicTag_HeroWingCostProp    = "HeroWingCostProp"
	LogicTag_GvGstartFight       = "GvGstartFight"
	LogicTag_GvGFinishFight      = "GvGFinishFight"
	LogicTag_GvGGuildFinish      = "GvgGuildFinish"
	LogicTag_GvGGuildInfo        = "GvGGuildInfo"
	LogicTag_GvGGuildScoreGM     = "GvGGuildScoreInfoGM"
	LogicTag_FestivalBoss        = "FestivalBossFight"
	LogicTag_HeroDiffStart       = "HeroDiffStart"
	LogicTag_HeroDiffFinish      = "HeroDiffFinish"
	LogicTag_ActivateGloryWeapon = "ActivateGloryWeapon"
	LogicTag_EnvolveGloryWeapon  = "EnvolveGloryWeapon"
	LogicTag_NicknameChg         = "NicknameChg"
	LogicTag_GuildWorship        = "GuildWorship"
	LogicTag_SendMail            = "SendMail"
	LogicTag_WspvpChallenge      = "WspvpChallenge"

	LogicTag_CSRobBuildCar      = "CSRobBuildCar"
	LogicTag_CSRobSendHelp      = "CSRobSendHelp"
	LogicTag_CSRobReceiveHelp   = "CSRobReceiveHelp"
	LogicTag_CSRobRobResult     = "CSRobRobResult"
	LogicTag_CSRobSetAutoAccept = "CSRobSetAutoAccept"
	LogicTag_HeroDestiny        = "DestinyHero"

	LogicTag_AstrologyInto    = "AstrologyInto"
	LogicTag_AstrologyUpgrade = "AstrologyUpgrade"
	LogicTag_AstrologyDestroy = "AstrologyDestroy"
	LogicTag_AstrologyAugur   = "AstrologyAugur"
	LogicTag_BeginWorldBoss   = "BeginWorldBoss"
	LogicTag_EndWorldBoss     = "EndWorldBoss"
	LogicTag_HotActivityAward = "HotActivityAward"

	LogicTag_HeroMagicPetLev               = "HeroMagicPetLev"
	LogicTag_HeroMagicPetStar              = "HeroMagicPetStar"
	LogicTag_HeroMagicPetChangeTalent      = "HeroMagicPetChangeTalent"
	LogicTag_HeroMagicPetChangeTalentSaved = "LogicTag_HeroMagicPetChangeTalentSaved"

	BITag = "[BI]"
)

type LogicInfo_MailInfo struct {
	IdsID       uint32
	MailRewards []LogicInfo_ItemC
	Reason      string
}

type LogicInfo_ProfileInfo struct {
	ChannelId string `json:"channelId"`
	Group     string `json:"group"`
}

type LogicInfo_Login struct {
	AccountName  string
	HcBuy        int64
	DeviceId     string
	MemSize      string
	ProfileName  string
	ProfileInfo  LogicInfo_ProfileInfo
	Ip           string
	IsReg        bool
	ClientVer    string
	MachineType  string
	PhoneNum     string
	LoginTimes   int64
	LoginType    string
	IDFA         string
	BundleUpdate string
	DataUpdate   string
}

type LogicInfo_CreateProfile struct {
	Name string
}

type LogicInfo_GiveItem struct {
	Reason     string
	Items      map[string]int64
	Avatar     int
	CorpLvl    uint32
	Channel    string
	Ip         string
	VIP        uint32
	Money      uint32
	Platform   string
	Type       string
	Value      int64
	BefValue   int64
	AfterValue int64
}

type LogicInfo_CostItem struct {
	Reason     string
	Items      map[string]int64
	VIP        uint32
	Type       string
	Value      int64
	BefValue   int64
	AfterValue int64
}
type LogicInfo_GiveCurrency struct {
	Reason   string
	Type     string
	Value    int64
	BefValue int64
	AftValue int64
	Channel  string
	CorpLvl  uint32
	Avatar   int
	Platform string
	Ip       string
	VIP      uint32
	HCcost   uint32
	ItemType string
	Name     string
}
type LogicInfo_CostCurrency struct {
	Reason   string
	Type     string
	Value    int64
	BefValue int64
	AftValue int64
	VIP      uint32
}
type LogicInfo_CorpExpChg struct {
	Reason   string
	BefValue uint32
	AftValue uint32
}
type LogicInfo_CorpLevelChg struct {
	Reason   string
	BefLevel uint32
	BefExp   uint32
	AftLevel uint32
	AftExp   uint32
}
type LogicInfo_QuestFinish struct {
	QuestId uint32
	Rewards []string
}
type LogicInfo_StageFinish struct {
	StageId       string
	IsWin         int
	Star          int32
	Times         int
	IsSweep       bool
	CostTime      int64
	CorpLvl       uint32
	GS            int
	SkillGenerals [helper.DestinyGeneralSkillMax]int
}

type LogicInfo_Pvp struct {
	AccountID     string
	MyGs          int
	MyChgScore    int
	MyBefScore    int
	MyAftScore    int
	MyChgPos      int
	MyBefPos      int
	MyAftPos      int
	IsWin         int
	EnemyId       string
	EnemyAvatar   int
	EnemyGs       int
	EnemyChgScore int
	EnemyBefScore int
	EnemyAftScore int
	EnemyChgPos   int
	EnemyBefPos   int
	EnemyAftPos   int
	CostTime      int64
}
type LogicInfo_StoreBuy struct {
	StoreType string
	ItemId    string
	ItemCount uint32
	CoinType  string
	CoinCount uint32
}
type LogicInfo_Tutorial struct {
	Step string
}
type LogicInfo_Gacha struct {
	GachaType string
	CoinType  string
	CoinCount uint32
	Items     map[string]uint32
}

type LogicInfo_GeneralAddNum struct {
	GeneralId string
	Count     uint32
	Reason    string
}

type LogicInfo_GeneralStar struct {
	GeneralId string
	Star_Aft  uint32
	Reason    string
	GS        int
}

type LogicInfo_GeneralRel struct {
	Relation string
	Level    uint32
	Reason   string
	GS       int
}

type LogicInfo_GeneralQuestRec struct {
	Quest      string
	GeneralIds []string
}

type LogicInfo_GeneralQuestFinish struct {
	Quest      string
	GeneralIds []string
	Reward     []string
	UseHc      int
}

type LogicInfo_IAP struct {
	Platform    string `json:"Platform,omitempty"`
	ChannelId   string `json:"ChannelId,omitempty"`
	AccountName string
	CorpLvl     uint32
	Name        string
	Avatar      int
	GoodIdx     uint32
	GameOrderId string `json:"GameOrderId,omitempty"`
	GoodName    string `json:"GoodName,omitempty"`
	Money       uint32 `json:"Money,omitempty"`
	Order       string `json:"Order,omitempty"`
	PayTime     string `json:"PayTime,omitempty"`
	HcBuy       uint32 `json:"HcBuy,omitempty"`
	HCGive      uint32 `json:"HCGive,omitempty"`
	Ip          string
	VIP         uint32
	Moneysum    uint32
	HasHcBuy    int64
	HasHcGive   int64
	HasHcCp     int64
	Channel     string
}

type LogicInfo_Guild struct {
	GuildUUID string
	GuildID   int64
	Name      string
	Level     uint32
	MemNum    int
	Mems      []LogicInfo_GuildMem

	Acid   string // 某事件的当事人
	BefPos string `json:"BefPos,omitempty"` // 只有改变职位有效，之前职位
	AftPos string `json:"AftPos,omitempty"` // 只有改变职位有效，之后职位
}
type LogicInfo_GuildLv struct {
	GuildUUID   string
	PrevGuildLv uint32
	CurGuildLv  uint32
}
type LogicInfo_GuildMem struct {
	Name          string
	AccountID     string
	GuildPosition string
}

type LogicInfo_GuildGateEnemy struct {
	GuildUUID    string
	GuildID      int64
	Name         string
	MemJoinCount int
	Point        int
}

type LogicInfo_GuildBossFight struct {
	GuildUUID    string
	DamageHpRate string
	LeftHpRate   string
}

type LogicInfo_RedeemCode struct {
	Name                 string
	RedeemCode           string
	RedeemCode_BatchId   int64
	RedeemCode_IsNoLimit bool
}

type LogicInfo_LogOut struct {
	OnLineTime int64
}

type LogicInfo_Phone struct {
	Name  string
	Phone string
}

type LogicInfo_PveBoss struct {
	PlayerGs   int
	IsWin      int
	CostTime   int64
	BossName   string
	BossDegree uint32
	BossGs     uint32
}

type LogicInfo_EquipAbstract struct {
	AimEquip    string
	MatEquip    string
	AimOldTrick []string
	AimNewTrick []string
	FineIron    int64
	GS          int
}

type LogicInfo_EquipAbstractCancel struct {
	AimEquip    string
	AimOldTrick []string
	AimNewTrick []string
	GS          int
}

type LogicInfo_EquipStarUp struct {
	Typ       string
	Slot      int
	BefStar   uint32
	AftStar   uint32
	ScCost    uint32
	MoneyCost uint32
	HCCost    uint32
}

type LogicInfo_EquipMatEnhAdd struct {
	Slot   int
	Lvl    uint32
	BefMat []bool
	AftMat []bool
	GS     int
}

type LogicInfo_EquipMatEnhLvlUp struct {
	Slot   int
	BefLvl uint32
	AftLvl uint32
	GS     int
}

type LogicInfo_Fish struct {
	AwardIds []uint32
	IsTen    bool
	IsHc     bool
}

type LogicInfo_Gank struct {
	Enemy       string
	IsWin       int
	IsRevenge   bool
	IsSysNotice bool
}

type LogicInfo_TeamPvp struct {
	AttackerAcid         string
	AttackerCorpLvl      uint32
	AttackerGs           int
	AttackerRankChg      int
	AttackerIsWin        int
	AttackerAvatars      []int
	AttackerAvatarStar   []uint32
	AttackerWinRate      string
	AttackerRnd          string
	BeAttackerAcid       string
	BeAttackerCorpLvl    uint32
	BeAttackerGs         int
	BeAttackerRankChg    int
	BeAttackerIsWin      int
	BeAttackerAvatars    []int
	BeAttackerAvatarStar []uint32
}

type LogicInfo_GveMatch struct {
	Server string
}

type LogicInfo_GveMatchCancel struct {
	Server         string
	CancelCostTime int64
}

type LogicInfo_GveGame struct {
	GameId   string
	IsHard   bool
	IsDouble bool
	IsUseHc  bool
	BossId   []string
	CostTime int64
	GS       int
}

type LogicInfo_GveGame1 struct {
	GameId   string
	IsWin    int
	IsHard   bool
	IsDouble bool
	IsUseHc  bool
	BossId   []string
	GS       int
}

type LogicInfo_HeroUnLock struct {
	AvatarID int
	AftPiece uint32
}

type LogicInfo_HeroStarUp struct {
	AvatarID int
	BefStar  uint32
	AftStar  uint32
	GS       int
	AftPiece uint32
}

type LogicInfo_HeroAddPiece struct {
	AvatarID int
	Piece    uint32
	BefPiece uint32
	AftPiece uint32
	Reason   string
}

type LogicInfo_FirstPay struct {
	GoodIdx            uint32
	FarthestLevel      int32
	FarthestEliteLevel int32
	FarthestHellLevel  int32
}

type LogicInfo_PointInfo struct {
	Point    int
	BefPoint int
	AftPoint int
	Reason   string
}

type LogicInfo_GuildRank struct {
	GuildUUid []string
}

type LogicInfo_HitEgg struct {
	Tier   int
	IsSpec bool
	LootID string
	Weight uint32
}

type LogicInfo_PayFeedBack struct {
	FirstMoney  int
	FirstBack   int
	SecondMoney int
	SecondBack  int
}

type LogicInfo_ItemC struct {
	Item  string
	Count uint32
}

type LogicInfo_GuildAddInventory struct {
	AddLoot []LogicInfo_ItemC
	Reason  string
}

type LogicInfo_GuildAssignInventory struct {
	AssignItem string
	Gotter     string
	Item       []LogicInfo_ItemC
}

type LogicInfo_GuildBoss struct {
	BossId     string
	LeftHpRate string
}

type LogicInfo_GuildBossInfo struct {
	GuildUUID    string
	Name         string
	Degree       int
	Boss         []LogicInfo_GuildBoss
	BigBoss      LogicInfo_GuildBoss
	JoinCount    int
	JoinMemCount int
}

type LogicInfo_DestinyGeneralAct struct {
	Id int
	GS int
}

type LogicInfo_GuildHeroSoul struct {
	PrevSoulLv uint32
	CurSoulLv  uint32
	GS         int
}

type LogicInfo_TalentLvUp struct {
	TalentId uint32
	PrevLv   uint32
	curLv    uint32
	GS       int
	IsAct    bool // 是否穿戴, true为穿戴, false为摘除
}

type LogicInfo_TitleChange struct {
	OldTitleId string
	NewTitleId string
	GS         int
}

type LogicInfo_ShareWeChat struct {
	AccountID        string
	Level            int
	Type_ShareWeChat int
	HeroID           int
	OwnedHeroCount   int
	AllHeroCount     int
	Time             int64
	VIP              uint32
}

type GachaRankInfo struct {
	AccountID string
	Rank      string
	Score     string
}
type LogicInfo_GachaRank struct {
	GachaRank []GachaRankInfo
}

type LogicInfo_PassiveSkillAdd struct {
	SkillId string
	GS      int
}

type LogicInfo_Expedition struct {
	FightResult       int
	Level             int
	GS                int
	HeroId            []int64
	IsExpeditionSweep int
}

type LogicInfo_HeroWing struct {
	ActedWing []int
	NewGet    int
	StarLv    int
	Level     int
	Recycle   int
	Reason    string
	Typ       string
	Item      []string
	Count     []int
}

type LogicInfo_HeroSwingAct struct {
	ActWing  int
	AvatarID int
}

type LogicInfo_HeroSwingLvUp struct {
	AvatarID int

	Level int
}
type LogicInfo_HeroSwingStarLvUp struct {
	AvatarID  int
	StarLevel int
}
type LogicInfo_HeroSwingRecycle struct {
	AvatarID int
	GetItem  []string
	GetCount []uint32
}
type LogicInfo_HeroSwingGetItem struct {
	GetItem  []string
	GetCount []uint32
	Reason   string
}
type LogicInfo_HeroSwingCostItem struct {
	AvatarID  int
	CostItem  []string
	CostCount []uint32
}

type LogicInfo_FestivalBossFight struct {
	Corplvl uint32
	GS      int
	Time    int64
	Iskill  int
}

type LogicInfo_GuildWorship struct {
	AccountID        string
	CorpLvl          uint32
	VIP              uint32
	WorshipAccountId string
	WorshipCorpLvl   uint32
	WorshipVipLvl    uint32
}
type GetLastSetCurLogType func(string) string

type ActiveExclusiveWeapon struct {
	VIP    uint32
	Avatar int
	GS     int
}

type EvolveExclusiveWeapon struct {
	VIP     uint32
	Avatar  int
	GS      int
	BeforeQ int
	AfterQ  int
}

type BeginWorldBoss struct {
	AccountID string
	AvatarID  []int
	BuffLevel int
	GS        int
	VIP       int
}

type EndWorldBoss struct {
	AccountID string
	AvatarID  []int
	GS        int
	VIP       int
}

type PlayerChangeName struct {
	BeforeName string
	AfterName  string
}

type WspvpChallenge struct {
	Attacker   WspvpPlayer
	BeAttacker WspvpPlayer
}

type WspvpPlayer struct {
	AccountID  string
	AvatarStar []int
	AvatarID   []int
	AvatarGs   []int
	BeforeRank int
	AfterRank  int
	CorpLvl    int
	IsWin      bool
}

type HeroDestiny struct {
	AccountID string
	CorpLvl   uint32
	VIP       uint32
	GS        int
	DestinyId int
}

//时间$$FINISH_ACTIVITY$$区服ID$$账号ID$$角色ID$$主任务ID$$奖励$$任务类型$$子任务id
// 运营活动奖励
type LogicInfo_HotActivityAward struct {
	Sid           uint32
	AccountId     string
	Activityid    uint32
	SubActivityId uint32
	AwardItem     map[string]uint32
	ActivityType  uint32
}

// CSRob 劫营夺粮系列 -------
type LogicInfo_CSRobBuildCar struct {
	AccountID string
	VIP       uint32
	Grade     uint32
	IsSkip    bool
	Num       uint32
}

func LogCSRobBuildCar(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, grade uint32, skip bool, num uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[LogCSRobBuildCar][%s] vip %d grade %d skip %v num %d", accountId, vip, grade, skip, num)

	r := LogicInfo_CSRobBuildCar{
		AccountID: accountId,
		VIP:       vip,
		Grade:     grade,
		IsSkip:    skip,
		Num:       num,
	}

	TypeInfo := LogicTag_CSRobBuildCar
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

type LogicInfo_CSRobSendHelp struct {
	AccountID       string
	VIP             uint32
	Grade           uint32
	TargetAccountID string
	TargetVip       uint32
}

func LogCSRobSendHelp(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, grade uint32, target string, targetVip uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[LogCSRobSendHelp][%s] vip %d grade %d target %s vip %d", accountId, vip, grade, target, targetVip)

	r := LogicInfo_CSRobSendHelp{
		AccountID:       accountId,
		VIP:             vip,
		Grade:           grade,
		TargetAccountID: target,
		TargetVip:       targetVip,
	}

	TypeInfo := LogicTag_CSRobSendHelp
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

type LogicInfo_CSRobReceiveHelp struct {
	AccountID       string
	VIP             uint32
	Grade           uint32
	SenderAccountID string
	SenderVip       uint32
}

func LogCSRobReceiveHelp(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, grade uint32, sender string, senderVip uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[LogCSRobReceiveHelp][%s] vip %d grade %d sender %s vip %d", accountId, vip, grade, sender, senderVip)

	r := LogicInfo_CSRobReceiveHelp{
		AccountID:       accountId,
		VIP:             vip,
		Grade:           grade,
		SenderAccountID: sender,
		SenderVip:       senderVip,
	}

	TypeInfo := LogicTag_CSRobReceiveHelp
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

type LogicInfo_CSRobRobResult struct {
	AccountID       string
	VIP             uint32
	GS              int
	TargetAccountID string
	TargetVip       uint32
	TargetGs        int
	CarGrade        uint32
	IsWin           int
}

func LogCSRobRobResult(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, gs int, target string, targetVip uint32, targetGs int, carGrade uint32, isWin bool,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[LogCSRobRobResult][%s] vip %d gs %d target %s vip %d gs %d, isWin %v", accountId, vip, gs, target, targetVip, targetGs, isWin)

	win := 0
	if false == isWin {
		win = 1
	}
	r := LogicInfo_CSRobRobResult{
		AccountID:       accountId,
		VIP:             vip,
		GS:              gs,
		TargetAccountID: target,
		TargetVip:       targetVip,
		TargetGs:        targetGs,
		CarGrade:        carGrade,
		IsWin:           win,
	}

	TypeInfo := LogicTag_CSRobRobResult
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

type LogicInfo_CSRobSetAutoAccept struct {
	AccountID         string
	VIP               uint32
	AutoAcceptSetting string
}

func LogCSRobSetAutoAccept(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, autoAcceptSetting string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[LogCSRobSetAutoAccept][%s] vip %d auto %s", accountId, vip, autoAcceptSetting)

	r := LogicInfo_CSRobSetAutoAccept{
		AccountID:         accountId,
		VIP:               vip,
		AutoAcceptSetting: autoAcceptSetting,
	}

	TypeInfo := LogicTag_CSRobSetAutoAccept
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

// 星图系列-------------------------
type LogicInfo_AstrologyInto struct {
	AccountID string
	CorpLvl   int
	VIP       uint32
	AvatarID  int
	HoleID    uint32
	Grade     uint32
	ItemId    string
	BeforeGs  int
	AfterGs   int
}

func LogAstrologyInto(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, hero uint32, hole uint32, grade uint32, soulID string, beforeGs int, afterGs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[LogAstrologyInto][%s] lv %d vip %d hero %d hole %d grade %d soulID %s beforeGs %d afterGs %d",
		accountId, corpLvl, vip, hero, hole, grade, soulID, beforeGs, afterGs)

	r := LogicInfo_AstrologyInto{
		AccountID: accountId,
		CorpLvl:   int(corpLvl),
		VIP:       vip,
		AvatarID:  int(hero),
		HoleID:    hole,
		Grade:     grade,
		ItemId:    soulID,
		BeforeGs:  beforeGs,
		AfterGs:   afterGs,
	}

	TypeInfo := LogicTag_AstrologyInto
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

type LogicInfo_AstrologyUpgrade struct {
	AccountID     string
	CorpLvl       int
	VIP           uint32
	AvatarID      int
	HoleID        uint32
	Grade         uint32
	ItemId        string
	BeforeUpgrade uint32
	AfterUpgrade  uint32
	BeforeGs      int
	AfterGs       int
}

func LogAstrologyUpgrade(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, hero uint32, hole uint32, grade uint32, soulID string, beforeUpgrade uint32, afterUpgrade uint32, beforeGs int, afterGs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[LogAstrologyUpgrade][%s] lv %d vip %d hero %d hole %d grade %d soulID %s beforeGs %d afterGs %d beforeUpgrade %d afterUpgrade %d",
		accountId, corpLvl, vip, hero, hole, grade, soulID, beforeUpgrade, afterUpgrade, beforeGs, afterGs)

	r := LogicInfo_AstrologyUpgrade{
		AccountID:     accountId,
		CorpLvl:       int(corpLvl),
		VIP:           vip,
		AvatarID:      int(hero),
		HoleID:        hole,
		Grade:         grade,
		ItemId:        soulID,
		BeforeUpgrade: beforeUpgrade,
		AfterUpgrade:  afterUpgrade,
		BeforeGs:      beforeGs,
		AfterGs:       afterGs,
	}

	TypeInfo := LogicTag_AstrologyUpgrade
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

type LogicInfo_AstrologyDestroy struct {
	AccountID     string
	CorpLvl       int
	VIP           uint32
	AvatarID      int
	HoleID        uint32
	Grade         uint32
	Items         map[string]int64
	BeforeUpgrade uint32
}

func LogAstrologyDestroy(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, hero uint32, hole uint32, grade uint32, souls map[string]int64, beforeUpgrade uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[LogAstrologyDestroy][%s] lv %d vip %d hero %d hole %d grade %d souls %+v beforeGs %d",
		accountId, corpLvl, vip, hero, hole, grade, souls, beforeUpgrade)

	r := LogicInfo_AstrologyDestroy{
		AccountID:     accountId,
		CorpLvl:       int(corpLvl),
		VIP:           vip,
		AvatarID:      int(hero),
		HoleID:        hole,
		Grade:         grade,
		Items:         souls,
		BeforeUpgrade: beforeUpgrade,
	}

	TypeInfo := LogicTag_AstrologyDestroy
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

type LogicInfo_AstrologyAugur struct {
	AccountID      string
	CorpLvl        int
	VIP            uint32
	AugurLevel     uint32
	NextAugurLevel uint32
	Items          map[string]int64
}

func LogAstrologyAugur(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, augurLevel uint32, nextAugurLevel uint32, souls map[string]int64,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[LogAstrologyAugur][%s] lv %d vip %d augurLevel %d nextAugurLevel %d souls %+v",
		accountId, corpLvl, vip, augurLevel, nextAugurLevel, souls)

	r := LogicInfo_AstrologyAugur{
		AccountID:      accountId,
		CorpLvl:        int(corpLvl),
		VIP:            vip,
		AugurLevel:     augurLevel,
		NextAugurLevel: nextAugurLevel,
		Items:          souls,
	}

	TypeInfo := LogicTag_AstrologyAugur
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

// 英雄log-----------------------------------------------------------------------------------------------------------
func LogLogin(accountId, accountName string, avatar int,
	hc int64, deviceId, memSize, profileName, ip, bundleUpdate, dataUpdate string,
	isReg bool, clientVer, machineType, phoneNum string, profileInfo LogicInfo_ProfileInfo, idfa string,
	loginTimes int64, loginType string, corpLvl uint32, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Login][%s] %s %s", accountId, accountName, info)

	r := LogicInfo_Login{
		AccountName:  accountName,
		HcBuy:        hc,
		DeviceId:     deviceId,
		MemSize:      memSize,
		ProfileName:  profileName,
		ProfileInfo:  profileInfo,
		IDFA:         idfa,
		Ip:           ip,
		IsReg:        isReg,
		ClientVer:    clientVer,
		MachineType:  machineType,
		PhoneNum:     phoneNum,
		LoginTimes:   loginTimes,
		LoginType:    loginType,
		BundleUpdate: bundleUpdate,
		DataUpdate:   dataUpdate,
	}
	TypeInfo := LogicTag_Login
	logiclog.Error(accountId, avatar, corpLvl, profileInfo.ChannelId, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogCreateProfile(accountId string, avatar int, name string, corpLvl uint32, channel string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[CreateProfile][%s] %s %s", accountId, name, info)

	r := LogicInfo_CreateProfile{
		Name: name,
	}

	TypeInfo := LogicTag_Create
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogCreateRole(accountId string, avatar int, name string, corpLvl uint32, channel string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[CreateRole][%s] %s %s", accountId, name, info)

	r := LogicInfo_CreateProfile{
		Name: name,
	}

	TypeInfo := LogicTag_CreateRole
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}
func LogLogout(accountId string, avatar int, onlineTime int64, corpLvl uint32, channel string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Logout][%s] %s", accountId, info)

	r := LogicInfo_LogOut{
		OnLineTime: onlineTime,
	}

	TypeInfo := LogicTag_Logout
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogGiveItem(accountId string, avatar int, corpLvl uint32, channel string,
	reason string, items map[string]int64, vipLvl uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[%s][GiveItem][%s] %v %s ", reason, accountId, items, info)

	r := LogicInfo_GiveItem{
		Reason:  reason,
		Items:   items,
		Avatar:  avatar,
		CorpLvl: corpLvl,
		Channel: channel,
		VIP:     vipLvl,
	}

	TypeInfo := LogicTag_GiveItem
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogGiveItemUseSelf(accountId string, avatar int, corpLvl uint32, channel string,
	reason string, typ string, oldV int64, chgV int64, afV int64, vipLvl uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[%s][GiveItem][%s] %s ", reason, accountId, info)

	r := LogicInfo_GiveItem{
		Reason:     reason,
		Avatar:     avatar,
		CorpLvl:    corpLvl,
		Channel:    channel,
		VIP:        vipLvl,
		Type:       typ,
		Value:      chgV,
		BefValue:   oldV,
		AfterValue: afV,
	}

	TypeInfo := LogicTag_GiveItemUseSelf
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogCostItem(accountId string, avatar int, corpLvl uint32, channel string,
	reason string, items map[string]int64, vipLvl uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[%s][CostItem][%s] %v %s ", reason, accountId, items, info)

	r := LogicInfo_CostItem{
		Reason: reason,
		Items:  items,
		VIP:    vipLvl,
	}

	TypeInfo := LogicTag_CostItem
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogCostItemUseSelf(accountId string, avatar int, corpLvl uint32, channel string,
	reason string, typ string, oldV int64, chgV int64, afV int64, vipLvl uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[%s][CostItem][%s] %v %s ", reason, accountId, info)

	r := LogicInfo_CostItem{
		Reason:     reason,
		VIP:        vipLvl,
		Type:       typ,
		Value:      chgV,
		BefValue:   oldV,
		AfterValue: afV,
	}

	TypeInfo := LogicTag_CostItemUseSelf
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogGiveCurrency(accountId string, avatar int, corpLvl uint32, channel string,
	reason string, typ string, oldV int64, chgV int64, platform string, ip string,
	vipLvl uint32, hccost uint32, itmeType string, name string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[%s][GiveCurrency][%s] type %s oldV %d chgV %d %s", reason, accountId, typ, oldV, chgV, info)
	var platformnew string
	if platform == "ios" {
		platformnew = "0"

	}
	if platform == "android" {
		platformnew = "1"
	}
	if platform == "" {
		platformnew = "3"
	}
	r := LogicInfo_GiveCurrency{
		Reason:   reason,
		Type:     typ,
		Value:    chgV,
		BefValue: oldV,
		AftValue: oldV + chgV,
		Channel:  channel,
		CorpLvl:  corpLvl,
		Avatar:   avatar,
		Platform: platformnew,
		Ip:       ip,
		VIP:      vipLvl,
		HCcost:   hccost,
		ItemType: itmeType,
		Name:     name,
	}

	TypeInfo := LogicTag_GiveCurrency
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogCostCurrency(accountId string, avatar int, corpLvl uint32, channel string,
	reason string, typ string, oldV int64, chgV int64, vip uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[%s][CostCurrency][%s] type %s oldV %d chgV %d %s", reason, accountId, typ, oldV, chgV, info)

	r := LogicInfo_CostCurrency{
		Reason:   reason,
		Type:     typ,
		Value:    chgV,
		BefValue: oldV,
		AftValue: oldV - chgV,
		VIP:      vip,
	}

	TypeInfo := LogicTag_CostCurrency
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogCorpExpChg(accountId string, avatar int, corpLvl uint32, channel string,
	reason string, oldV, chgV uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[%s][CorpExpChg][%s]  oldV %d chgV %d %s", reason, accountId, oldV, chgV, info)

	r := LogicInfo_CorpExpChg{
		Reason:   reason,
		BefValue: oldV,
		AftValue: oldV + chgV,
	}

	TypeInfo := LogicTag_CorpExpChg
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogCorpLevelChg(accountId string, avatar int, corpLvl uint32, channel string,
	reason string, befLevel, befExp, AftLevel, AftExp uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[%s][CorpLevelChg][%s]  befLevel %d befExp %d AftLevel %d AftExp %d %s", reason, accountId,
		befLevel, befExp, AftLevel, AftExp, info)

	r := LogicInfo_CorpLevelChg{
		Reason:   reason,
		BefLevel: befLevel,
		BefExp:   befExp,
		AftLevel: AftLevel,
		AftExp:   AftExp,
	}

	TypeInfo := LogicTag_CorpLevelChg
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogQuestFinish(accountId string, avatar int, corpLvl uint32, channel string,
	questId uint32, items []string, itemcounts []uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[QuestFinish][%s]  quest %d %s", accountId, questId, info)

	reward := make([]string, 0, 10)
	for i, item := range items {
		reward = append(reward, fmt.Sprintf("%s,%d", item, itemcounts[i]))
	}
	r := LogicInfo_QuestFinish{
		QuestId: questId,
		Rewards: reward,
	}

	TypeInfo := LogicTag_QuestFinish
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogStageFinish(accountId string, avatar int, stageId string, isWin bool,
	star int32, times int, isSweep bool, costTime int64,
	corpLvl uint32, channel string, gs int, skillGeneral [helper.DestinyGeneralSkillMax]int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[StageFinish][%s]  stage %s isWin %v times %d %s", accountId, stageId, isWin, times, info)

	iwin := 0
	if !isWin {
		iwin = 1
	}

	r := LogicInfo_StageFinish{
		StageId:       stageId,
		IsWin:         iwin,
		Star:          star,
		Times:         times,
		IsSweep:       isSweep,
		CostTime:      costTime,
		CorpLvl:       corpLvl,
		GS:            gs,
		SkillGenerals: skillGeneral,
	}

	TypeInfo := LogicTag_StageFinish
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogPvPFinish(accountId string, avatar int, corpLvl uint32, channel string,
	gs, befSocre, aftSocre, befPos, aftPos int,
	enemyAcid string, eavatar, egs, ebefSocre, eaftSocre, ebefPos, eaftPos int,
	isWin bool, costTime int64, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Pvp][%s] Pvp %s", accountId, info)

	iwin := 0
	if !isWin {
		iwin = 1
	}
	r := LogicInfo_Pvp{
		AccountID:     accountId,
		MyGs:          gs,
		MyChgScore:    aftSocre - befSocre,
		MyBefScore:    befSocre,
		MyAftScore:    aftSocre,
		MyChgPos:      aftPos - befPos,
		MyBefPos:      befPos,
		MyAftPos:      aftPos,
		IsWin:         iwin,
		EnemyId:       enemyAcid,
		EnemyAvatar:   eavatar,
		EnemyGs:       egs,
		EnemyChgScore: eaftSocre - ebefSocre,
		EnemyBefScore: ebefSocre,
		EnemyAftScore: eaftSocre,
		EnemyChgPos:   eaftPos - ebefPos,
		EnemyBefPos:   ebefPos,
		EnemyAftPos:   eaftPos,
		CostTime:      costTime,
	}
	TypeInfo := LogicTag_Pvp
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogStoreBuy(accountId string, avatar int, corpLvl uint32, channel string,
	storeType string, itemId string, itemCount uint32, coinType string, coinCount uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[StoreBuy][%s]  store %s itemId %s itemCount %d coinType %s coinCount %d %s", accountId, storeType,
		itemId, itemCount, coinType, coinCount, info)

	r := LogicInfo_StoreBuy{
		StoreType: storeType,
		ItemId:    itemId,
		ItemCount: itemCount,
		CoinType:  coinType,
		CoinCount: coinCount,
	}

	TypeInfo := LogicTag_StoreBuy
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogShopBuy(accountId string, avatar int, corpLvl uint32, channel string,
	storeType string, itemId string, itemCount uint32, coinType string, coinCount uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogShopBuy][%s]  shop %s itemId %s itemCount %d coinType %s coinCount %d %s", accountId, storeType,
		itemId, itemCount, coinType, coinCount, info)

	r := LogicInfo_StoreBuy{
		StoreType: storeType,
		ItemId:    itemId,
		ItemCount: itemCount,
		CoinType:  coinType,
		CoinCount: coinCount,
	}

	TypeInfo := LogicTag_ShopBuy
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogTutorial(accountId string, avatar int, corpLvl uint32, channel string,
	step string, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Tutorial][%s]  step %d %s", accountId, step, info)

	r := LogicInfo_Tutorial{
		Step: step,
	}

	TypeInfo := LogicTag_Tutorial
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogGacha(accountId string, avatar int, corpLvl uint32, channel string,
	gachaType string, coinType string, coinCount uint32, items map[string]uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Gacha][%s]  gachaType %s coinType %s coinCount %d items %v %s", accountId,
		gachaType, coinType, coinCount, items, info)

	r := LogicInfo_Gacha{
		GachaType: gachaType,
		CoinType:  coinType,
		CoinCount: coinCount,
		Items:     items,
	}

	TypeInfo := LogicTag_Gacha
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogGeneralAddNum(accountId string, avatar int, corpLvl uint32, channel string,
	generalId string, count uint32, reason string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[General-AddNum][%s]  general %s count %d reason %s %s", accountId, generalId, count, reason, info)

	r := LogicInfo_GeneralAddNum{
		GeneralId: generalId,
		Count:     count,
		Reason:    reason,
	}
	typeInfo := LogicTag_GeneralAddNum
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogGeneralStarUp(accountId string, avatar int, corpLvl uint32, channel string,
	generalId string, star uint32, reason string, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[General-Star][%s]  general %s star %d reason %s %s", accountId, generalId, star, reason, info)

	r := LogicInfo_GeneralStar{
		GeneralId: generalId,
		Star_Aft:  star,
		Reason:    reason,
		GS:        gs,
	}

	typeInfo := LogicTag_GeneralStarLvlUp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogGeneralRelLevelUp(accountId string, avatar int, corpLvl uint32, channel string,
	relation string, level uint32, reason string, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[General-Rel-Lvl][%s]  relation %s lvl %d reason %s %s", accountId, relation, level, reason, info)

	r := LogicInfo_GeneralRel{
		Relation: relation,
		Level:    level,
		Reason:   reason,
		GS:       gs,
	}

	typeInfo := LogicTag_GeneralRelLvlUp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogGeneralQuestReceive(accountId string, avatar int, corpLvl uint32, channel string,
	quest string, GeneralIds []string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[GeneralQuestReceive][%s] %s", accountId, info)

	r := LogicInfo_GeneralQuestRec{
		Quest:      quest,
		GeneralIds: GeneralIds,
	}

	typeInfo := LogicTag_GeneralQuestRec
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogGeneralQuestFinish(accountId string, avatar int, corpLvl uint32, channel string,
	quest string, GeneralIds []string,
	reward map[string]uint32, costHc bool,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogGeneralQuestFinish][%s] %s", accountId, info)

	usehc := 1
	if costHc {
		usehc = 0
	}

	rc := make([]string, 0, len(reward))
	for r, c := range reward {
		rc = append(rc, fmt.Sprintf("%s %d", r, c))
	}
	r := LogicInfo_GeneralQuestFinish{
		Quest:      quest,
		GeneralIds: GeneralIds,
		Reward:     rc,
		UseHc:      usehc,
	}

	typeInfo := LogicTag_GeneralQuestFinish
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogRedeemCode(accountId string, avatar int, corpLvl uint32, channel string,
	name string, redeemCode string, batchId int64, isNoLimit bool,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[RedeemCode][%s] name %s code %s %s", accountId, name, redeemCode, info)

	r := LogicInfo_RedeemCode{
		Name:                 name,
		RedeemCode:           redeemCode,
		RedeemCode_BatchId:   batchId,
		RedeemCode_IsNoLimit: isNoLimit,
	}

	typeInfo := LogicTag_RedeemCode
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogPhone(accountId string, avatar int, corpLvl uint32, channel string,
	name, phone string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Phone][%s] %s", accountId, info)

	r := LogicInfo_Phone{
		Name:  name,
		Phone: phone,
	}

	typeInfo := LogicTag_Phone
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

// 名将试炼
func LogPveBoss(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, isSuccess bool, costTime int64,
	boss string, bossDegree uint32, bossGs uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[PveBoss][%s] %s", accountId, info)

	iwin := 0
	if !isSuccess {
		iwin = 1
	}

	r := LogicInfo_PveBoss{
		PlayerGs:   gs,
		IsWin:      iwin,
		CostTime:   costTime,
		BossName:   boss,
		BossDegree: bossDegree,
		BossGs:     bossGs,
	}
	typeInfo := LogicTag_PveBoss
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

// 洗练
func LogEquipAbstract(accountId string, avatar int, corpLvl uint32, channel string,
	aimEquip, matEquip string,
	aimEquipTrickBef []string, aimEquipTrickAft []string, ironCount int64, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[EquipAbstract][%s] %s", accountId, info)

	r := LogicInfo_EquipAbstract{
		AimEquip:    aimEquip,
		MatEquip:    matEquip,
		AimOldTrick: aimEquipTrickBef,
		AimNewTrick: aimEquipTrickAft,
		FineIron:    ironCount,
		GS:          gs,
	}

	typeInfo := LogicTag_EquipAbstract
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

// fish
func LogFish(accountId string, avatar int, corpLvl uint32, channel string,
	awardId []uint32, isTen bool, isHc bool,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Fish][%s] %s", accountId, info)

	r := LogicInfo_Fish{
		AwardIds: awardId,
		IsTen:    isTen,
		IsHc:     isHc,
	}

	typeInfo := LogicTag_Fish
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

// gank
func LogGank(accountId string, avatar int, corpLvl uint32, channel string,
	enemyId string, isWin, isRevenge, isSysNotice bool,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Gank][%s] %s", accountId, info)

	var win int
	if !isWin {
		win = 1
	}
	r := LogicInfo_Gank{
		Enemy:       enemyId,
		IsWin:       win,
		IsRevenge:   isRevenge,
		IsSysNotice: isSysNotice,
	}

	typeInfo := LogicTag_Gank
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogEquipAbstractCancel(accountId string, avatar int, corpLvl uint32, channel string,
	aimEquip string,
	aimEquipTrickBef []string, aimEquipTrickAft []string, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[EquipAbstractCancel][%s] %s", accountId, info)

	r := LogicInfo_EquipAbstractCancel{
		AimEquip:    aimEquip,
		AimOldTrick: aimEquipTrickBef,
		AimNewTrick: aimEquipTrickAft,
		GS:          gs,
	}

	typeInfo := LogicTag_EquipAbstractCancel
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogEquipStarUp(accountId string, avatar int, corpLvl uint32, channel string,
	typ string, slot int, befStar, aftStar uint32, sc, money, hc uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogEquipStarUp][%s] %s", accountId, info)

	r := LogicInfo_EquipStarUp{
		Typ:       typ,
		Slot:      slot,
		BefStar:   befStar,
		AftStar:   aftStar,
		ScCost:    sc,
		MoneyCost: money,
		HCCost:    hc,
	}

	typeInfo := LogicTag_EquipStarUp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogEquipMatEnhAdd(accountId string, avatar int, corpLvl uint32, channel string,
	slot int, lvl uint32, befMat []bool, aftMat []bool, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogEquipMatEnhAdd][%s] %s", accountId, info)

	r := LogicInfo_EquipMatEnhAdd{
		Slot:   slot,
		Lvl:    lvl,
		BefMat: befMat,
		AftMat: aftMat,
		GS:     gs,
	}

	typeInfo := LogicTag_EquipMatEnhAdd
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogEquipMatEnhLvlUp(accountId string, avatar int, corpLvl uint32, channel string,
	slot int, befLvl, aftLvl uint32, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogEquipMatEnhLvlUp][%s] %s", accountId, info)

	r := LogicInfo_EquipMatEnhLvlUp{
		Slot:   slot,
		BefLvl: befLvl,
		AftLvl: aftLvl,
		GS:     gs,
	}

	typeInfo := LogicTag_EquipMatEnhLvlUp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

// 主将
func LogHeroUnlock(accountId string, avatar int, corpLvl uint32, channel string,
	avatarId int, aftPiece uint32, fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogHeroUnlock][%s] %s", accountId, info)

	r := LogicInfo_HeroUnLock{
		AvatarID: avatarId,
		AftPiece: aftPiece,
	}

	typeInfo := LogicTag_HeroUnlock
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogHeroStarUp(accountId string, avatar int, corpLvl uint32, channel string,
	avatarId int, befStar, aftStar uint32, gs int, aftPiece uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogHeroStarUp][%s] %s", accountId, info)

	r := LogicInfo_HeroStarUp{
		AvatarID: avatarId,
		BefStar:  befStar,
		AftStar:  aftStar,
		GS:       gs,
		AftPiece: aftPiece,
	}

	typeInfo := LogicTag_HeroStarUp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogHeroAddPiece(accountId string, avatar int, corpLvl uint32, channel string,
	avatarId int, piece, befPiece, aftPiece uint32, reason string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogHeroAddPiece][%s] %s", accountId, info)

	r := LogicInfo_HeroAddPiece{
		AvatarID: avatarId,
		Piece:    piece,
		BefPiece: befPiece,
		AftPiece: aftPiece,
		Reason:   reason,
	}

	typeInfo := LogicTag_HeroAddPiece
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogFirstPay(accountId string, avatar int, corpLvl uint32, channel string,
	goodIdx uint32, farthestLvl, farthestEliteLevel, farthestHellLevel int32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogFirstPay][%s] %s", accountId, info)

	r := LogicInfo_FirstPay{
		GoodIdx:            goodIdx,
		FarthestLevel:      farthestLvl,
		FarthestEliteLevel: farthestEliteLevel,
		FarthestHellLevel:  farthestHellLevel,
	}

	typeInfo := LogicTag_FirstPay
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogPoint(accountId string, avatar int, corpLvl uint32, channel string,
	isDaily bool, point, bef, aft int, reason string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogPoint][%s] %s", accountId, info)

	r := LogicInfo_PointInfo{
		Point:    point,
		BefPoint: bef,
		AftPoint: aft,
		Reason:   reason,
	}

	typeInfo := LogicTag_Account7Point
	if isDaily {
		typeInfo = LogicTag_DailyPoint
	}
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogGuildRank(topN []string) {
	uuids := make([]string, 0, 20)
	for i := 0; i < 20 && i < len(topN); i++ {
		uuids = append(uuids, topN[i])
	}
	r := LogicInfo_GuildRank{
		GuildUUid: uuids,
	}
	typeInfo := LogicTag_GuildRank
	logiclog.Error("", 0, 0, "", typeInfo, r, "", "")
}

// team pvp
func LogTeamPvp(accountId string, avatar int, corpLvl uint32, channel string,
	r LogicInfo_TeamPvp, isWin bool,
	myHeroStar [helper.AVATAR_NUM_MAX]uint32, enemyHeroStar [helper.AVATAR_NUM_MAX]uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogTeamPvp][%v] %s", r, info)

	win := 0
	viceWin := 1
	if !isWin {
		win = 1
		viceWin = 0
	}
	r.AttackerIsWin = win
	r.BeAttackerIsWin = viceWin
	myAvatarsStar := make([]uint32, len(r.AttackerAvatars))
	for i, id := range r.AttackerAvatars {
		myAvatarsStar[i] = myHeroStar[id]
	}
	r.AttackerAvatarStar = myAvatarsStar

	enemyAvatarsStar := make([]uint32, len(r.BeAttackerAvatars))
	for i, id := range r.BeAttackerAvatars {
		enemyAvatarsStar[i] = enemyHeroStar[id]
	}
	r.BeAttackerAvatarStar = enemyAvatarsStar

	typeInfo := LogicTag_TeamPvp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogHitEgg(accountId string, avatar int, corpLvl uint32, channel string,
	tier int, isSpec bool, loot string, weight uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	r := LogicInfo_HitEgg{
		Tier:   tier,
		IsSpec: isSpec,
		LootID: loot,
		Weight: weight,
	}
	typeInfo := LogicTag_HitEgg
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogDestinyGeneralAct(accountId string, avatar int, corpLvl uint32, channel string,
	id int, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[DestinyGeneralAct][%s] Got %d DestinyGeneral %s", accountId, id, info)
	r := LogicInfo_DestinyGeneralAct{
		Id: id,
		GS: gs,
	}
	typeInfo := LogicTag_Act_DestinyGeneral
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogHeroSoulLvUp(accountId string, avatar int, corpLvl uint32, channel string,
	prevSoulLv uint32, curSoulLv uint32, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroSoulLvUp][%s] HeroSoul LevelUp %d->%d %s", accountId, prevSoulLv, curSoulLv, info)
	r := LogicInfo_GuildHeroSoul{
		PrevSoulLv: prevSoulLv,
		CurSoulLv:  curSoulLv,
		GS:         gs,
	}
	typeInfo := LogicTag_HeroSoulLvUp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogTalentLvUp(accountId string, avatar int, corpLvl uint32, channel string,
	talentId uint32, prevTalentLv uint32, curTalentLv uint32, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[TalentLvUp][%s] Talent %d LevelUp %d->%d %s", accountId, talentId, prevTalentLv, curTalentLv, info)
	r := LogicInfo_TalentLvUp{
		TalentId: talentId,
		PrevLv:   prevTalentLv,
		curLv:    curTalentLv,
		GS:       gs,
	}
	typeInfo := LogicTag_TalentLvUp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogTitleChange(accountId string, avatar int, corpLvl uint32, channel string,
	oldId string, newId string, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[TitleActivate][%s] Lose %s Title and got %s Title %s", accountId, oldId, newId, info)
	r := LogicInfo_TitleChange{
		OldTitleId: oldId,
		NewTitleId: newId,
		GS:         gs,
	}
	typeInfo := LogicTag_TitleChange
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

// gve
func LogGveMatch(accountId string, avatar int, corpLvl uint32, channel string,
	server string, cancel bool, cancelCostTime int64,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogGveMatch] cancel %v %s", cancel, info)

	if cancel {
		typeInfo := LogicTag_GveCancelMatch
		r := LogicInfo_GveMatchCancel{
			Server:         server,
			CancelCostTime: cancelCostTime,
		}
		logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
	} else {
		r := LogicInfo_GveMatch{
			Server: server,
		}

		typeInfo := LogicTag_GveStartMatch
		logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
	}
}

func LogGveGameStart(accountId string, avatar int, corpLvl uint32, channel string,
	GameID string, BossId []string, GameIsHard, GameIsDouble, GameIsUseHc bool,
	costTime int64, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogGveGameStart] %s", info)

	r := LogicInfo_GveGame{
		GameId:   GameID,
		IsHard:   GameIsHard,
		IsDouble: GameIsDouble,
		IsUseHc:  GameIsUseHc,
		BossId:   BossId,
		CostTime: costTime,
		GS:       gs,
	}

	typeInfo := LogicTag_GevGameStart
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogGveGameStop(accountId string, avatar int, corpLvl uint32, channel string,
	GameID string, BossId []string, isWin, GameIsHard, GameIsDouble, GameIsUseHc bool, gs int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogGveGameStop] %s", info)

	iwin := 0
	if !isWin {
		iwin = 1
	}
	r := LogicInfo_GveGame1{
		GameId:   GameID,
		IsWin:    iwin,
		IsHard:   GameIsHard,
		IsDouble: GameIsDouble,
		IsUseHc:  GameIsUseHc,
		BossId:   BossId,
		GS:       gs,
	}

	typeInfo := LogicTag_GevGameStop
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

// 公会
func LogGuildOper(accountId string, channel string, guildOper string, guild *LogicInfo_Guild, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Guild-Oper][%s]  %s %s", guild.GuildUUID, guildOper, info)

	logiclog.ErrorForGuild(accountId, channel, guild.GuildUUID, guildOper, *guild, format, params...)
}

func LogGuildGEOver(guildUuid string, guildId int64, guildName string, joinCount, point int, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Guild-GateEnemy][%s]  GateEnemy over joinCount %d point %d %s", guildUuid, joinCount, point, info)

	r := LogicInfo_GuildGateEnemy{
		GuildUUID:    guildUuid,
		GuildID:      guildId,
		Name:         guildName,
		MemJoinCount: joinCount,
		Point:        point,
	}

	logiclog.ErrorForGuild("", "", guildUuid, LogicTag_GuildGateEnemyOver, r, format, params...)
}

func LogGuildLvUp(guildUuid string, prevGuildLv, curGuildLv uint32, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[Guild-LvUp][%s] Guild LevelUp %d->%d %s", guildUuid, prevGuildLv, curGuildLv, info)

	r := LogicInfo_GuildLv{
		PrevGuildLv: prevGuildLv,
		CurGuildLv:  curGuildLv,
		GuildUUID:   guildUuid,
	}
	logiclog.ErrorForGuild("", "", guildUuid, LogicTag_GuildLvUp, r, format, params...)
}

// guild boss
func LogGuildBossFight(accountId string, avatar int, corpLvl uint32, channel string,
	guild string, damageRate float64, leftRate float64,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogGveGameStart] %s", info)

	r := LogicInfo_GuildBossFight{
		GuildUUID:    guild,
		DamageHpRate: fmt.Sprintf("%.2f", damageRate),
		LeftHpRate:   fmt.Sprintf("%.2f", leftRate),
	}

	typeInfo := LogicTag_GuildBossFight
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogGuildBoss(guildUuid, name string, degree int, boss []LogicInfo_GuildBoss,
	bigBoss LogicInfo_GuildBoss,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[LogGuildBoss][%s]  %s", guildUuid, info)

	r := LogicInfo_GuildBossInfo{
		GuildUUID: guildUuid,
		Name:      name,
		Degree:    degree,
		Boss:      boss,
		BigBoss:   bigBoss,
	}

	logiclog.ErrorForGuild("", "", guildUuid, LogicTag_GuildBoss, r, format, params...)
}

// guild inventory
func LogAddGuildInventory(guildUuid string, loots []LogicInfo_ItemC, reason string,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[AddGuildInventory][%s]  %s", guildUuid, info)

	r := LogicInfo_GuildAddInventory{
		AddLoot: loots,
		Reason:  reason,
	}
	logiclog.ErrorForGuild("", "", guildUuid, LogicTag_AddGuildInventory, r, format, params...)
}

func LogAssignGuildInventory(guildUuid string, lootId, acid string,
	item []string, count []uint32, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[AssignGuildInventory][%s]  %s", guildUuid, info)

	_item := make([]LogicInfo_ItemC, len(item))
	for i, itm := range item {
		_item[i] = LogicInfo_ItemC{itm, count[i]}
	}
	r := LogicInfo_GuildAssignInventory{
		AssignItem: lootId,
		Gotter:     acid,
		Item:       _item,
	}
	logiclog.ErrorForGuild("", "", guildUuid, LogicTag_AssignGuildInventory, r, format, params...)
}

const (
	Android_Platform = "android"
	IOS_Platform     = "ios"
)

func LogIAPTry(accountId, accountName, name string, avatar int, corpLvl uint32, channel string,
	goodIdx uint32, gameOrderId string,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[IAPTry][%s] %d %s %s ", accountId, goodIdx, gameOrderId, info)

	r := LogicInfo_IAP{
		AccountName: accountName,
		Name:        name,
		GoodIdx:     goodIdx,
		GameOrderId: gameOrderId,
	}
	typeInfo := LogicTag_IAPTry
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogIAP(accountId, accountName, name string, avatar int, corpLvl uint32, channel string,
	goodIdx uint32, gameOrderId, goodName, order string, money uint32, platform, channelId, payTime string,
	hcBuy, hcGive uint32, ip string, vipLvl uint32, moneysum uint32, hasHcBuy int64, hasHcGive int64,
	hasHcCp int64, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[IAP][%s]  %s %d %s %s %d %d", accountId, platform, money, order, info, hcBuy, hcGive)
	var platformnew string
	money1 := money * 100
	moneysum1 := moneysum * 100
	iPayTime, err := strconv.ParseInt(payTime, 10, 64)
	if err != nil {
		logs.Error("logiclog LogIAP err %v", err)
		return
	}
	payTime2 := time.Unix(iPayTime, 0).In(util.ServerTimeLocal)
	payTime3 := payTime2.Format("20060102150405")
	if platform == "ios" {
		platformnew = "0"

	}
	if platform == "android" {
		platformnew = "1"
	}
	if platform == "" {
		platformnew = "3"
	}
	r := LogicInfo_IAP{
		Platform:    platformnew,
		ChannelId:   channelId,
		AccountName: accountName,
		Name:        name,
		GoodIdx:     goodIdx,
		GameOrderId: gameOrderId,
		GoodName:    goodName,
		Money:       money1,
		Order:       order,
		PayTime:     payTime3,
		HcBuy:       hcBuy,
		HCGive:      hcGive,
		CorpLvl:     corpLvl,
		Avatar:      avatar,
		Ip:          ip,
		VIP:         vipLvl,
		Moneysum:    moneysum1,
		HasHcBuy:    hasHcBuy,
		HasHcGive:   hasHcGive,
		HasHcCp:     hasHcCp,
		Channel:     channel,
	}
	typeInfo := LogicTag_IAP
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogPayFeedBack(accountId, accountName string, avatar int, corpLvl uint32, channel string,
	firstMoney, firstBack, secondMoney, secondBack int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[PayFeedBack][%s] %s %d %d %d %d %s", accountId, accountName,
		firstMoney, firstBack, secondMoney, secondBack, info)

	r := LogicInfo_PayFeedBack{
		FirstMoney:  firstMoney,
		FirstBack:   firstBack,
		SecondMoney: secondMoney,
		SecondBack:  secondBack,
	}
	typeInfo := LogicTag_PayFeedBack
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

// other log-----------------------------------------------------------------------------------------------------------
func LogReturnTown_c(accountId string, avatar int, corpLvl uint32, channel string,
	format string, params ...interface{}) {
	info := fmt.Sprintf(format, params...)
	logs.Trace("[ReturnTown][%s] %s", accountId, info)

	r := struct {
		Event string
	}{
		Event: "ReturnTown",
	}

	TypeInfo := LogicTag_ClientEvent
	logiclog.Info(accountId, avatar, corpLvl, channel, TypeInfo, r, format, params...)
}

func LogStage_c(accountId string, avatar int, corpLvl uint32, channel string,
	stageId string, eventTyp string, mark int64, gs int, format string, params ...interface{}) {
	info := fmt.Sprintf(format, params...)
	TypeInfo := LogicTag_ClientEvent
	logs.Trace("[%s][%s] %s %s", TypeInfo, accountId, stageId, info)

	r := struct {
		Event   string
		StageId string
		Mark    int64
		GS      int
	}{
		StageId: stageId,
		Mark:    mark,
		GS:      gs,
	}
	r.Event = eventTyp

	logiclog.Info(accountId, avatar, corpLvl, channel, TypeInfo, r, format, params...)
}

func LogBoss_c(accountId string, avatar int, corpLvl uint32, channel string,
	bossId string, isEnter bool, mark int64, format string, params ...interface{}) {
	info := fmt.Sprintf(format, params...)
	TypeInfo := LogicTag_ClientEvent
	logs.Trace("[%s][%s] %s %s", TypeInfo, accountId, bossId, info)

	r := struct {
		Event  string
		BossId string
		Mark   int64
	}{
		BossId: bossId,
		Mark:   mark,
	}

	if isEnter {
		r.Event = "EnterBoss"
	} else {
		r.Event = "LeaveBoss"
	}

	logiclog.Info(accountId, avatar, corpLvl, channel, TypeInfo, r, format, params...)
}

func LogClientData_c(accountId string, avatarId int, corpLvl uint32, channel string,
	timeEvent string, time int64, avatars []int, chgAvC uint32, deadAvatars []int,
	fgs GetLastSetCurLogType, subfgs GetLastSetCurLogType, format string, params ...interface{}) {
	info := fmt.Sprintf(format, params...)
	logs.Trace("[ClientData][%s] %s", accountId, info)

	_av := make([]string, 3)
	for i, a := range avatars {
		_av[i] = fmt.Sprintf("%d", a)
	}
	_dav := make([]string, 3)
	for i, a := range deadAvatars {
		_dav[i] = fmt.Sprintf("%d", a)
	}
	TypeInfo := LogicTag_ClientTimeEvent
	r := struct {
		TimeEvent      string
		Time           int64
		Avatar1        string
		Avatar2        string
		Avatar3        string
		ChgAvatarCount uint32
		DeadAvatar1    string
		DeadAvatar2    string
		DeadAvatar3    string
		LastTimeEvent  string
	}{
		TimeEvent:      timeEvent,
		Time:           time,
		Avatar1:        _av[0],
		Avatar2:        _av[1],
		Avatar3:        _av[2],
		ChgAvatarCount: chgAvC,
		DeadAvatar1:    _dav[0],
		DeadAvatar2:    _dav[1],
		DeadAvatar3:    _dav[2],
		LastTimeEvent:  subfgs(timeEvent),
	}

	logiclog.Debug(accountId, avatarId, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogClientFenghuoData_c(accountId string, avatarId int64, corpLvl uint32, channel string,
	EventName string,
	SelfUid string,
	SelfGs int64,
	OtherUid string,
	OtherGs int64,
	Difficult string,
	RoomWaitTime int64,
	Rate string,
	IsWin int64,
	BattleTotalTime int64,
	fgs GetLastSetCurLogType, subfgs GetLastSetCurLogType, format string, params ...interface{}) {

	info := fmt.Sprintf(format, params...)
	logs.Trace("[ClientData][%s] %s", accountId, info)

	TypeInfo := LogicTag_ClientTimeEvent

	r := struct {
		TimeEvent       string
		SelfUid         string
		SelfGs          int64
		OtherUid        string
		OtherGs         int64
		Difficult       string
		RoomWaitTime    int64
		Rate            string
		IsWin           int64
		BattleTotalTime int64
	}{
		TimeEvent:       EventName,
		SelfUid:         SelfUid,
		SelfGs:          SelfGs,
		OtherUid:        OtherUid,
		OtherGs:         OtherGs,
		Difficult:       Difficult,
		RoomWaitTime:    RoomWaitTime,
		Rate:            Rate,
		IsWin:           IsWin,
		BattleTotalTime: BattleTotalTime,
	}
	logiclog.Debug(accountId, int(avatarId), corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

// 微信分享
func LogShareWeChat(accountID string, avatarID int, corpLvl uint32, channel string,
	level int, shareType int, heroID int, ownedHeroCount int, allHeroCount int, time int64, vipLvl uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[ShareWeChat][%s]  type: %d count: %d/%d, %s", accountID, shareType, ownedHeroCount, allHeroCount, info)

	r := LogicInfo_ShareWeChat{
		AccountID:        accountID,
		Level:            level,
		Type_ShareWeChat: shareType,
		HeroID:           heroID,
		OwnedHeroCount:   ownedHeroCount,
		AllHeroCount:     allHeroCount,
		Time:             time,
		VIP:              vipLvl,
	}

	TypeInfo := LogicTag_ShareWeChat
	logiclog.Error(accountID, avatarID, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

//限时神将排名
func LogGachaRank(ids, ranks, scores []string) {
	accountids := make([]GachaRankInfo, 0, 100)
	for i := 0; i < 20 && i < len(ids); i++ {
		accountids = append(accountids, GachaRankInfo{ids[i], ranks[i], scores[i]})
	}
	r := LogicInfo_GachaRank{
		GachaRank: accountids,
	}
	typeInfo := LogicTar_GachaRank
	logiclog.Error("", 0, 0, "", typeInfo, r, "", "")
}

func LogPassiveSkillAdd(accountId string, avatar int, corpLvl uint32, channel string,
	skillid string, gs int, fgs GetLastSetCurLogType,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[PassiveSkillAdd] awatarid: %d skillid %s  corplvl %d gs %d , %s", avatar, skillid, corpLvl, gs, info)
	r := LogicInfo_PassiveSkillAdd{
		SkillId: skillid,
		GS:      gs,
	}
	typeInfo := LogicTar_PassiveSkillAdd
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogExpedition(accountId string, avatar int, corpLvl uint32, channel string,
	fightresult int, step int, gs int, heroid []int64, isSweep int, fgs GetLastSetCurLogType,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[ExpeditionEvent] awatarid: %d FightResult %s  corplvl %d step %d gs %d , %s", avatar, fightresult, corpLvl, step, gs, info)
	r := LogicInfo_Expedition{
		FightResult:       fightresult,
		Level:             step,
		GS:                gs,
		HeroId:            heroid,
		IsExpeditionSweep: isSweep,
	}
	typeInfo := LogicTar_ExpeditionEvent
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogExpeditionRest(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, fgs GetLastSetCurLogType,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[ExpeditionRest] awatarid: %d  gs %d , %s", avatar, corpLvl, gs, info)
	r := LogicInfo_Expedition{
		GS: gs,
	}
	typeInfo := LogicTar_ExpeditionRest
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogHeroWingAct(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, fgs GetLastSetCurLogType, actWing int, AvatarID int,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroWingChange] avatarid: %s  corplvl %d gs %d , %s", avatar, corpLvl, gs, info)
	r := LogicInfo_HeroSwingAct{
		ActWing:  actWing,
		AvatarID: AvatarID,
	}
	typeInfo := LogicTag_HeroWingAct
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogHeroWingLvUp(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, fgs GetLastSetCurLogType, level int, AvatarID int, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroWingChange] avatarid: %s  corplvl %d gs %d , %s", avatar, corpLvl, gs, info)
	r := LogicInfo_HeroSwingLvUp{
		Level:    level,
		AvatarID: AvatarID,
	}
	typeInfo := LogicTag_HeroWingLvUp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}
func LogHeroWingStarLvUp(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, fgs GetLastSetCurLogType, starLv int, AvatarID int,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroWingChange] avatarid: %s  corplvl %d gs %d , %s", avatar, corpLvl, gs, info)
	r := LogicInfo_HeroSwingStarLvUp{
		StarLevel: starLv,
		AvatarID:  AvatarID,
	}
	typeInfo := LogicTag_HeroWingStarLvUp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}
func LogHeroWingReset(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, fgs GetLastSetCurLogType, avatarID int, item []string, count []uint32,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroWingChange] avatarid: %s  corplvl %d gs %d , %s", avatar, corpLvl, gs, info)
	r := LogicInfo_HeroSwingRecycle{
		GetItem:  item,
		GetCount: count,
		AvatarID: avatarID,
	}
	typeInfo := LogicTag_HeroWingReset
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}
func LogHeroWingGetProp(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, fgs GetLastSetCurLogType, item []string, count []uint32, reason string,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroWingChange] avatarid: %s  corplvl %d gs %d , %s", avatar, corpLvl, gs, info)
	r := LogicInfo_HeroSwingGetItem{
		GetItem:  item,
		GetCount: count,
		Reason:   reason,
	}
	typeInfo := LogicTag_HeroWingGetProp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}
func LogHeroWingCostProp(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, fgs GetLastSetCurLogType, item []string, count []uint32, AvatarID int,
	format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroWingChange] avatarid: %s  corplvl %d gs %d , %s", avatar, corpLvl, gs, info)
	r := LogicInfo_HeroSwingCostItem{
		CostItem:  item,
		CostCount: count,
		AvatarID:  AvatarID,
	}
	typeInfo := LogicTag_HeroWingCostProp
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

const (
	MAT_HWLevelStone = "MAT_HWLevelStone"
	MAT_StarStone    = "MAT_StarStone"
	MAT_StarStone_1  = "MAT_StarStone_1"
	MAT_StarStone_2  = "MAT_StarStone_2"
	MAT_StarStone_3  = "MAT_StarStone_3"
	MAT_StarStone_4  = "MAT_StarStone_4"
	MAT_StarStone_5  = "MAT_StarStone_5"
)

func LogIsWingProp(id string) bool {
	return id == MAT_HWLevelStone ||
		id == MAT_StarStone ||
		id == MAT_StarStone_1 ||
		id == MAT_StarStone_2 ||
		id == MAT_StarStone_3 ||
		id == MAT_StarStone_4 ||
		id == MAT_StarStone_5
}

func LogFestivalBossFight(accountId string, avatar int, corpLvl uint32, channel string,
	gs int, time int64, iskill int, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[FestivalBoss] avatarid: %s  corplvl %d gs %d , %s", avatar, corpLvl, gs, info)
	r := LogicInfo_FestivalBossFight{
		Corplvl: corpLvl,
		GS:      gs,
		Time:    time,
		Iskill:  iskill,
	}
	typeInfo := LogicTag_FestivalBoss
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

type BIBaseInfo struct {
	AccountID string
	Avatar    int
	Channel   string
	CorpLvl   uint32
	Fgs       GetLastSetCurLogType
}

func LogCommonInfo(baseInfo BIBaseInfo, content interface{}, typeInfo string, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("%s avatarid: %s %v , %s", typeInfo, baseInfo.Avatar, content, info)
	logiclog.Error(baseInfo.AccountID, baseInfo.Avatar, baseInfo.CorpLvl, baseInfo.Channel,
		typeInfo, content, baseInfo.Fgs(typeInfo), format, params...)
}

func LogGuildWorship(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, worshipAccountid string, worshipCorpLvl uint32, worshipVip uint32, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[GuildWorship] avatarid: %s  corplvl %d vip %d , %s", avatar, corpLvl, vip, info)
	r := LogicInfo_GuildWorship{
		AccountID:        accountId,
		CorpLvl:          corpLvl,
		VIP:              vip,
		WorshipAccountId: worshipAccountid,
		WorshipCorpLvl:   worshipCorpLvl,
		WorshipVipLvl:    worshipVip,
	}
	typeInfo := LogicTag_GuildWorship
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogSendMail(userID string, IdsID uint32, Reason string, ItemId []string, Count []uint32) {
	if len(ItemId) != len(Count) {
		return
	}
	info := LogicInfo_MailInfo{
		IdsID:       IdsID,
		Reason:      Reason,
		MailRewards: make([]LogicInfo_ItemC, 0),
	}
	for j, v := range ItemId {
		info.MailRewards = append(info.MailRewards, LogicInfo_ItemC{
			Item:  v,
			Count: Count[j],
		})
	}
	LogCommonInfo(BIBaseInfo{AccountID: userID, Fgs: func(string) string {
		return ""
	}}, info, LogicTag_SendMail, "")
}

func LogHeroDestiny(accountId string, avatar int, corpLvl uint32, channel string,
	vip uint32, gs int, destinyId int, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroDestiny] avatarid: %s  corplvl %d vip %d , %s", avatar, corpLvl, vip, info)
	r := HeroDestiny{
		AccountID: accountId,
		CorpLvl:   corpLvl,
		VIP:       vip,
		GS:        gs,
		DestinyId: destinyId,
	}
	typeInfo := LogicTag_HeroDestiny
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

func LogHgrHotActivity(accountId string, avatar int, corpLvl uint32, channel string,
	sid uint32, activityId uint32, subActivityId uint32,
	awardItem map[string]uint32, ActivityType uint32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	logs.Trace("[HotActivityAward][%s] ActivityTime %d ActivityId %d SubActivityId %s Award %v ActivityType %s", accountId,
		activityId, subActivityId, awardItem, ActivityType)

	r := LogicInfo_HotActivityAward{
		Sid:           sid,
		AccountId:     accountId,
		Activityid:    activityId,
		SubActivityId: subActivityId,
		AwardItem:     awardItem,
		ActivityType:  ActivityType,
	}

	TypeInfo := LogicTag_HotActivityAward
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

type LogicInfo_HeroMagicPetLevUp struct {
	//主将id、灵兽id、前后等级、升级后战力、VIP
	AvatarID       int
	PetID          int
	BeforePetLevUp int
	EndPetLevUp    int
	AfterLevUpGs   int
	VIP            int
}

func LogHeroMagicPetLevUp(accountId string, avatar int, corpLvl uint32, channel string,
	petID, beforePetLevUp, endPetLevUp, afterLevUpGs, vip int, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroMagicPetLevUp] avatarid: %d,  petid: %d, beforePetLevUp: %d, endPetLevUp: %d, afterLevUpGs: %d, vip: %d,%s", avatar, petID, beforePetLevUp, endPetLevUp, afterLevUpGs, vip, info)
	r := LogicInfo_HeroMagicPetLevUp{
		AvatarID:       avatar,
		PetID:          petID,
		BeforePetLevUp: beforePetLevUp,
		EndPetLevUp:    endPetLevUp,
		AfterLevUpGs:   afterLevUpGs,
		VIP:            vip,
	}
	typeInfo := LogicTag_HeroMagicPetLev
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

type LogicInfo_HeroMagicPetStarUp struct {
	//主将id、灵兽id、前后星级、升星后战力、VIP、是否使用保星符、是否成功（成功：0、失败：1）
	AvatarID        int
	PetID           int
	BeforePetStarUp int
	EndPetStarUp    int
	AfterStarUpGs   int
	VIP             int
	Special         bool
	Success         bool
}

func LogHeroMagicPetStarUp(accountId string, avatar int, corpLvl uint32, channel string,
	petID, beforePetStarUp, endPetStarUp, afterStarUpGs, vip int, special, success bool, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroMagicPetStarUp] avatarid: %d,  petid: %d, beforePetStarUp: %d, endPetStarUp: %d, afterStarUpGs: %d, vip: %d, special: %t, success: %t,%s",
		avatar, petID, beforePetStarUp, endPetStarUp, afterStarUpGs, vip, special, success, info)
	r := LogicInfo_HeroMagicPetStarUp{
		AvatarID:        avatar,
		PetID:           petID,
		BeforePetStarUp: beforePetStarUp,
		EndPetStarUp:    endPetStarUp,
		AfterStarUpGs:   afterStarUpGs,
		VIP:             vip,
		Special:         special,
		Success:         success,
	}
	typeInfo := LogicTag_HeroMagicPetStar
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

type TalentInterface interface {
	GetType() uint32
	GetValue() uint32
}

type Talent struct {
	Type  uint32 //类型
	Value uint32 //数值
}

func (t Talent) GetType() uint32 {
	return t.Type
}
func (t Talent) GetValue() uint32 {
	return t.Value
}

type LogicInfo_HeroMagicPetTalent struct {
	//主将id、灵兽id、前后属性变化（综合资质&单条资质属性类型和数值）、VIP、是否使用高级洗练符、当前累计洗炼次数（分两类：使用高级洗练符、没有使用高级洗练符）
	AvatarID              int
	PetID                 int
	BeforePetCompreTalent int
	EndPetCompreTalent    int
	BeforePetTalents      []Talent
	EndPetTalents         []Talent
	VIP                   int
	Special               bool
	SpecialCountnums      int
	NormalCountnums       int
}

//暂时保存在这里，reView之后考虑使用
//TODO by cyt----删除
func LogHeroMagicPetTalentOper(accountId string, avatar int, corpLvl uint32, channel string, magicpet *LogicInfo_HeroMagicPetTalent,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[HeroMagicPetTalent] avatarid: %d, petid: %d, beforePetCompreTalent: %d, "+
		"endPetCompreTalent: %d, beforePetTalents: %v, endPetTalents: %v, vip: %d, special: %t, specialCountnums: %d, normalCountnums: %d,%s",
		avatar, magicpet.PetID, magicpet.BeforePetCompreTalent, magicpet.EndPetCompreTalent, magicpet.BeforePetTalents, magicpet.EndPetTalents,
		magicpet.VIP, magicpet.Special, magicpet.SpecialCountnums, magicpet.NormalCountnums, info)
	typeInfo := LogicTag_HeroMagicPetChangeTalent
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, magicpet, fgs(typeInfo), format, params...)
}

func LogHeroMagicPetTalent(accountId string, avatar int, corpLvl uint32, channel string,
	petID, beforePetCompreTalent, endPetCompreTalent int, beforePetTalents, endPetTalents []TalentInterface, vip int, special bool,
	specialCountnums, normalCountnums int, fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)

	beforeTalents := make([]Talent, 0, len(beforePetTalents))
	endTalents := make([]Talent, 0, len(endPetTalents))
	for _, v := range beforePetTalents {
		beforeTalents = append(beforeTalents, Talent{v.GetType(), v.GetValue()})
	}
	for _, v := range endPetTalents {
		endTalents = append(endTalents, Talent{v.GetType(), v.GetValue()})
	}

	logs.Trace("[HeroMagicPetTalent] avatarid: %d, petid: %d, beforePetCompreTalent: %d, "+
		"endPetCompreTalent: %d, beforePetTalents: %v, endPetTalents: %v, vip: %d, special: %t, specialCountnums: %d, normalCountnums: %d,%s",
		avatar, petID, beforePetCompreTalent, endPetCompreTalent, beforeTalents, endTalents, vip, special, specialCountnums, normalCountnums, info)

	r := LogicInfo_HeroMagicPetTalent{
		AvatarID: avatar,
		PetID:    petID,
		BeforePetCompreTalent: beforePetCompreTalent,
		EndPetCompreTalent:    endPetCompreTalent,
		BeforePetTalents:      beforeTalents,
		EndPetTalents:         endTalents,
		VIP:                   vip,
		Special:               special,
		SpecialCountnums:      specialCountnums,
		NormalCountnums:       normalCountnums,
	}
	typeInfo := LogicTag_HeroMagicPetChangeTalent
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}

type LogicInfo_HeroMagicPetTalentSaved struct {
	//主将id、灵兽id、前后属性变化（综合资质&单条资质属性类型和数值）、洗练后战力、VIP、当前累计洗炼次数（分两类：使用高级洗练符、没有使用高级洗练符）
	AvatarID              int
	PetID                 int
	BeforePetCompreTalent int
	EndPetCompreTalent    int
	BeforePetTalents      []Talent
	EndPetTalents         []Talent
	AfterTalentGs         int
	VIP                   int
	SpecialCountnums      int
	NormalCountnums       int
}

func LogHeroMagicPetTalentSaved(accountId string, avatar int, corpLvl uint32, channel string,
	petID, beforePetCompreTalent, endPetCompreTalent int, beforePetTalents, endPetTalents []TalentInterface, afterTalentGs, vip, specialCountnums, normalCountnums int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)

	beforeTalents := make([]Talent, 0, len(beforePetTalents))
	endTalents := make([]Talent, 0, len(endPetTalents))
	for _, v := range beforePetTalents {
		beforeTalents = append(beforeTalents, Talent{v.GetType(), v.GetValue()})
	}
	for _, v := range endPetTalents {
		endTalents = append(endTalents, Talent{v.GetType(), v.GetValue()})
	}
	logs.Trace("[HeroMagicPetTalentSaved] avatarid: %d, petid: %d, beforePetCompreTalent: %d, "+
		"endPetCompreTalent: %d, beforePetTalents: %v, endPetTalents: %v, afterTalentGs: %d, vip: %d, specialCountnums: %d, normalCountnums: %d,%s",
		avatar, petID, beforePetCompreTalent, endPetCompreTalent, beforeTalents, endTalents, afterTalentGs, vip, specialCountnums, normalCountnums, info)
	r := LogicInfo_HeroMagicPetTalentSaved{
		AvatarID: avatar,
		PetID:    petID,
		BeforePetCompreTalent: beforePetCompreTalent,
		EndPetCompreTalent:    endPetCompreTalent,
		BeforePetTalents:      beforeTalents,
		EndPetTalents:         endTalents,
		AfterTalentGs:         afterTalentGs,
		VIP:                   vip,
		SpecialCountnums:      specialCountnums,
		NormalCountnums:       normalCountnums,
	}
	typeInfo := LogicTag_HeroMagicPetChangeTalentSaved
	logiclog.Error(accountId, avatar, corpLvl, channel, typeInfo, r, fgs(typeInfo), format, params...)
}
