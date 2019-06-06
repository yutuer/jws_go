package helper

import (
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

//DBProfile接口有两个函数
//
//存档数据结构的升级和修正代码也是通过这个函数去实现。
//
//## 只处理Struct <--> Hash的存储。
//配合DumpToHashDB&RestoreFromHashDB实现Redis的数据库存取。
//
//## 可以将数据结构转换成任何系列存储指令
//比如序列化所有数据为json，然后用GET/SET就能够实现了
//
//## 备注
//因为Redigo等Redis驱动只能处理如下成员变量 string, []byte, int, int64, float64, bool, nil
//所以如果成员变量是其他数据类型，则需要用如下方案解决。
// - 先转换成由基础类型成员组成的Struct
// - 转换成map
// - 在Struct中使用redis tag来实现冗余数据库值域
type DBProfile interface {
	DBName() string
	DBSave(redis.CmdBuffer) error
	DBLoad() error
}

// 统一的更新接口
type ISyncRsp interface {
	OnChangeBag()
	OnChangeUpdateItems(item_inner_type int, uId uint32, oldCount int64, reason string)
	OnChangeDelItems(item_inner_type int, uId uint32, itemId string, oldCount int64, reason string)
	OnChangeSC()
	OnChangeHC()
	OnChangeAvatarExp()
	OnChangeCorpExp()
	OnChangeEnergy()
	OnChangeBossFightPoint()
	OnChangeStage(sid string)
	OnChangeStageAll()
	OnChangeEquip()
	OnChangeAvatarEquip()
	OnChangeGeneralAllChange()
	OnChangeGameMode(gameModeId uint32)
	OnChangeVIP()
	OnChangeWheel()
	OnChangeBoss()
	OnChangeIAPGoodInfo()
	OnChangeQuestAll()
	OnChangeAvatarJade()
	OnChangeDestinyGenJade()
	OnChangeHitEgg()
	OnChangeHeroTalent()
	OnChangeIAPGift()
}

const OtherInfoForPlayerPLen = 3

type GuildSPRecord struct {
	GSP           uint32
	LastTimeStamp int64
}

func (r *GuildSPRecord) AddGuildSPRecord(gsp uint32, now_t int64) {
	r.GSP += gsp
	r.LastTimeStamp = now_t
}

type OtherInfoForPlayer struct {
	Pi [OtherInfoForPlayerPLen]int64
	Ps [OtherInfoForPlayerPLen]string
	// 公会科技树捐献
	GSTDay  GuildSPRecord
	GSTWeek GuildSPRecord
}

// 这个结构里不要在加东西了，尽量精简
type AccountSimpleInfo struct {
	Name            string                    `json:"name" codec:"name"`
	AccountID       string                    `json:"aid" codec:"aid"`
	CorpLv          uint32                    `json:"corplv" codec:"corplv"`
	Vip             uint32                    `json:"vip" codec:"vip"`
	GuildPosition   int                       `json:"position" codec:"position"` // 这个信息和其他信息不同是从工会向玩家存档更新, 注意其变化逻辑
	GuildName       string                    `json:"gname" codec:"gname"`
	GuildSp         int64                     `json:"sp" codec:"sp"`
	GuildBossCoin   int64                     `json:"gb" codec:"gb"`
	LastLoginTime   int64                     `json:"l_login" codec:"l_login"`
	CurrAvatar      int                       `json:"curr" codec:"curr"`
	CurrCorpGs      int                       `json:"currgs" codec:"currgs"`
	FashionEquips   [FashionPart_Count]string `json:"feqs" codec:"feqs"`
	WeaponStartLvl  uint32                    `json:"wsl" codec:"wsl"`     // 武器星级
	EqStartLvl      uint32                    `json:"eqsl" codec:"eqsl"`   // 其他装备最小星级
	InfoUpdateTime  int64                     `json:"time" codec:"time"`   // 整个info生成的时刻
	Contribution    [2]int64                  `json:"gc" codec:"gc"`       // 帮会贡献
	MaxTrialLv      int64                     `json:"mtl" codec:"mtl"`     // 最大爬塔等级
	AvatarStarLvl   [AVATAR_NUM_MAX]uint32    `json:"asl" codec:"asl"`     // 主将星级
	TeamPvpAvatar   [TeamPvpAvatarsCount]int  `json:"tpa" codec:"tps"`     // 3v3竞技场出站阵容
	TeamPvpAvatarLv [TeamPvpAvatarsCount]int  `json:"tpalv" codec:"tpslv"` // 3v3竞技场出站阵容星级
	TeamPvpGs       int                       `json:"tpags" codec:"tpsgs"` // 3v3竞技场出站阵容gs
	TitleOn         string                    `json:"ttlo" codec:"ttlo"`   // 当前头顶的称号
	TitleTimeOut    int64                     `json:"ttlto" codec:"ttlto"` // 称号过期时间，只影响有过期时间的称号
	GsHeroIds       []int                     `json:"gshero" codec:"gshero"`
	GsHeroGs        []int                     `json:"gsherogs" codec:"gsherogs"`
	GsHeroBaseGs    []int                     `json:"gsherobgs" codec:"gsherobgs"`
	online          bool                      // 是否在线
	Other           OtherInfoForPlayer        `json:"other" codec:"other"`
	Swing           int                       `json:"swing" codec:"swing"`
	MagicPetfigure  uint32                    `json:"magic_pet_figure" codec:"magic_pet_figure"`
	HeroDiffScore   [HeroDiff_Count]int       `json:"hero_diff_score codec:"hero_diff_score"`
	GVGScore        int                       `json:"gvg_score" codec:"gvg_score"`
	JadeLv          int64                     `json:"jade_lv" codec:"jade_lv"`
	DestinyLv       int64                     `json:"destiny_lv" codec:"destiny_lv"`
	EquipStarLv     int64                     `json:"equip_st_lv" codec:"equip_st_lv"`
	SwingStarLv     int64                     `json:"swing_star_lv" codec:"swing_star_lv"`
	HeroDestinyLv   int64                     `json:"hero_destiny_lv" codec:"hero_destiny_lv"`
	WuShuangGs      int64                     `json:"wu_shuang_gs" codec:"wu_shuang_gs"`
	Astrology       int64                     `json:"astrology" codec:"astrology"`
	ExclusivWeapon  int64                     `json:"exclusiv_weapon" codec:"exclusiv_weapon"`
	TopGsByCountry  [Country_Count]int64      `json:"top_gs_country" codec:"top_gs_country"` // 每个势力的最强3个武将总和
}

func (s *AccountSimpleInfo) GetOnline() bool {
	return s.online
}

func (s *AccountSimpleInfo) SetOnline(b bool) {
	s.online = b
}

type AvatarAttr_ struct {
	ATK             float32 `json:"atk"`        // 攻击力
	DEF             float32 `json:"def"`        // 防御力
	HP              float32 `json:"hp"`         // 生命值
	CritRate        float32 `json:"crit_r"`     // 暴击率
	ResilienceRate  float32 `json:"res_r"`      // 免暴率
	CritValue       float32 `json:"crit_v"`     // 暴击伤害
	ResilienceValue float32 `json:"res_v"`      // 免暴伤害
	Force           float32 `json:"force"`      // 武力
	Intellect       float32 `json:"intellect"`  // 智力
	Endurance       float32 `json:"endurance"`  // 统御
	HitRate         float32 `json:"hit_rate"`   // 命中率
	DodgeRate       float32 `json:"dodge_rate"` // 闪避率

	IceDamage       int32   `json:"ice_dmg"`     // 冰系攻击力
	IceDefense      int32   `json:"ice_def"`     // 冰系防御力
	IceBonus        float32 `json:"ice_bnus"`    // 冰系伤害加成
	IceResist       float32 `json:"ice_resist"`  // 冰系抗性
	FireDamage      int32   `json:"fire_dmg"`    // 火系攻击力
	FireDefense     int32   `json:"fire_def"`    // 火系防御力
	FireBonus       float32 `json:"fire_bnus"`   // 火系伤害加成
	FireResist      float32 `json:"fire_resist"` // 火系抗性
	LightingDamage  int32   `json:"lig_dmg"`     // 雷系攻击力
	LightingDefense int32   `json:"lig_def"`     // 雷系防御力
	LightingBonus   float32 `json:"lig_bnus"`    // 雷系伤害加成
	LightingResist  float32 `json:"lig_resist"`  // 雷系抗性
	PoisonDamage    int32   `json:"posn_dmg"`    // 毒系攻击力
	PoisonDefense   int32   `json:"posn_def"`    // 毒系防御力
	PoisonBonus     float32 `json:"posn_bnus"`   // 毒系伤害加成
	PoisonResist    float32 `json:"posn_resist"` // 毒系抗性

	gsAddon      uint32  `json:"gs_add"`   // Gs增加值
	gsFloatAddon float32 `json:"gs_f_add"` // Gs增加值
}

func (a *AvatarAttr_) GetGsAddon_() uint32 {
	return a.gsAddon
}
func (a *AvatarAttr_) AddGsAddon_(i uint32) {
	a.gsAddon += i
}

func (a *AvatarAttr_) GetGsFloatAddon_() float32 {
	return a.gsFloatAddon
}
func (a *AvatarAttr_) AddGsFloatAddon_(f float32) {
	a.gsFloatAddon = +f
}

type BagItemToClient struct {
	ID      uint32 `codec:"id"`
	TableID string `codec:"tableid"`
	ItemID  string `codec:"itemid"`
	Count   int64  `codec:"count"`
	Data    string `codec:"data"`
}

type FashionItem struct {
	ID              uint32 `codec:"id" json:"id"`
	TableID         string `codec:"tid" json:"tid"`
	ExpireTimeStamp int64  `codec:"ot" json:"ot"` // 过期时刻，99999永久
}

// 这里面有很多引用项, 从玩家身上构建出数据时,
// 注意不要保存这个结构,会引起不一致的问题
type Avatar2Client struct {
	Acid     string `codec:"acid"`
	AvatarId int    `codec:"aid"`
	CorpLv   uint32 `codec:"clv"`
	Name     string `codec:"name"`
	VipLv    uint32 `codec:"vip"`
	TitleOn  string `codec:"title_on"`

	Gs             int   `codec:"gs"`
	CorpGs         int   `codec:"corpgs"`
	SimplePvpScore int64 `codec:"pvp_score"`
	SimplePvpRank  int   `codec:"pvp_rank"`

	HeroStarLv []uint32 `codec:"hero"`
	HeroLv     []uint32 `codec:"heroLv"`
	HeroSoulLv uint32   `codec:"heroSoulLv"`

	Attr         []byte `codec:"attr"`
	GsHeroIds    []int  `codec:"gshero"`
	GsHeroGs     []int  `codec:"gsherogs"`
	GsHeroBaseGs []int  `codec:"gsherobgs"`

	// 下面可能会删除 TBD
	CorpXP         uint32   `codec:"cxp"`
	Arousals       []uint32 `codec:"arousals"`
	AvatarSkills   []uint32 `codec:"skills"`
	SkillPractices []uint32 `codec:"skillps"`
	AvatarLockeds  []int    `codec:"avatarlockeds"`

	Equips         [][]byte `codec:"equips"`
	equips         []BagItemToClient
	EquipUpgrade   []uint32 `codec:"equip_upgrade"`
	EquipStar      []uint32 `codec:"equip_star"`
	EquipMatEnhLv  []uint32 `codec:"equip_mat_enh"`
	EquipMatEnh    []bool   `codec:"eq_me_"`
	EquipMatEnhMax int      `codec:"eq_me_max_"`
	AvatarEquips   []uint32 `codec:"avatar_equips"` // 当前角色穿着的时装，客户端需要自己判断是否过期
	AllFashions    [][]byte `codec:"all_fashions"`  //
	allFashions    []FashionItem
	Title          []string `codec:"title"`

	Generals         []string `codec:"generals"` // TBD
	GeneralStars     []uint32 `codec:"genstar"`  // TBD
	GeneralRels      []string `codec:"genrels"`  // TBD
	GeneralRelLevels []uint32 `codec:"genrellv"` // TBD

	EquipJade       []string `codec:"equip_jade"`
	DestGeneralJade []string `codec:"dest_general_jade"`

	DestinyGeneralID        int   `codec:"dg"`
	DestinyGeneralLv        int   `codec:"dglv"`
	DestinyGeneralsID       []int `codec:"dgs"`
	DestinyGeneralsLv       []int `codec:"dgslv"`
	CurrDestinyGeneralSkill []int `codec:"dgss"`

	GuildUUID     string `codec:"guuid"`
	GuildName     string `codec:"gname"`
	GuildPos      int    `codec:"gpos"`
	GuildPost     string `codec:"post"`
	GuildPostTime int64  `codec:"postt"`

	PassiveSkillId []string `codec:"pskillid"`
	CounterSkillId []string `codec:"cskillid"`
	TriggerSkillId []string `codec:"tskillid"`
	HeroSwing      int      `codec:"swing"`
	MagicPetfigure uint32   `codec:"magic_pet_figure"`

	ShowCountry         int   `codec:"show_country"`
	HeroIdsByCountry    []int `codec:"ids_country"`
	HeroGsByCountry     []int `codec:"gs_country"`
	HeroBaseGsByCountry []int `codec:"base_gs_country"`
}

func (a *Avatar2Client) GetAcId() string {
	return a.Acid
}

func (a *Avatar2Client) SetAcId(acid string) {
	a.Acid = acid
}

func (a *Avatar2Client) GetEquips() []BagItemToClient {
	return a.equips[:]
}

func (a *Avatar2Client) AppendEquip(i BagItemToClient) {
	a.equips = append(a.equips, i)
}

func (a *Avatar2Client) GetAllFashions() []FashionItem {
	return a.allFashions[:]
}

func (a *Avatar2Client) AppendAllFashions(f FashionItem) {
	a.allFashions = append(a.allFashions, f)
}

type Avatar2ClientByJson struct {
	AcID           string   `json:"aid"`
	AvatarId       int      `json:"avatarid"`
	CorpLv         uint32   `json:"clv"`
	CorpXP         uint32   `json:"cxp"`
	Arousals       []uint32 `json:"arousals"`
	AvatarSkills   []uint32 `json:"skills"`
	SkillPractices []uint32 `json:"skillps"`
	Name           string   `json:"name"`
	VipLv          uint32   `json:"vip"`
	AvatarLockeds  []int    `json:"avatarlockeds"`
	HeroStarLv     []uint32 `json:"hero"`
	HeroLv         []uint32 `json:"heroLv"`

	HP   float32     `json:"hp"` // 生命值
	Attr AvatarAttr_ `json:"attr"`

	Gs             int   `json:"gs"`
	CorpGs         int   `json:"corpgs"`
	SimplePvpScore int64 `json:"pvp_score"`
	SimplePvpRank  int   `json:"pvp_rank"`

	Equips        []BagItemToClient `json:"equips"`
	EquipUpgrade  []uint32          `json:"equip_upgrade"`
	EquipStar     []uint32          `json:"equip_star"`
	EquipMatEnhLv []uint32          `json:"equip_mat_enh"`
	EquipMatEnh   [][]bool          `json:"eq_me_"`
	AvatarEquips  []uint32          `json:"avatar_equips"` // 当前角色穿着的时装，客户端需要自己判断是否过期
	AllFashions   []FashionItem     `json:"all_fashions"`
	Title         []string          `json:"title"`
	TitleOn       string            `json:"title_on"`

	Generals         []string `json:"generals"`
	GeneralStars     []uint32 `json:"genstar"`
	GeneralRels      []string `json:"genrels"`
	GeneralRelLevels []uint32 `json:"genrellv"`

	EquipJade       []string `json:"equip_jade"`
	DestGeneralJade []string `json:"dest_general_jade"`

	DestinyGeneralID        int   `json:"dg"`
	DestinyGeneralLv        int   `json:"dglv"`
	CurrDestinyGeneralSkill []int `json:"dgss"`

	GuildUUID     string `json:"guuid"`
	GuildName     string `json:"gname"`
	GuildPos      int    `json:"gpos"`
	GuildPost     string `json:"post"`
	GuildPostTime int64  `json:"postt"`

	PassiveSkillId []string `json:"pskillid"`
	CounterSkillId []string `json:"cskillid"`
	TriggerSkillId []string `json:"tskillid"`

	HeroSwing      int    `json:"swing"`
	MagicPetfigure uint32 `json:"magicpetfigure"`
}

func (a *Avatar2ClientByJson) GetAcId() string {
	return a.AcID
}

func (a *Avatar2ClientByJson) GetEquips() []BagItemToClient {
	return a.Equips[:]
}

func (a *Avatar2ClientByJson) GetAllFashions() []FashionItem {
	return a.AllFashions[:]
}

type GVGData struct {
	GvgAvatarState   [GVG_AVATAR_COUNT]AvatarState
	GvgDestinySkill  [DestinyGeneralSkillMax]int
	GvgEnemyAcID     string
	GvgEAvatarState  [GVG_AVATAR_COUNT]AvatarState
	GvgEDestinySkill [DestinyGeneralSkillMax]int
	GvgEnemyData     [GVG_AVATAR_COUNT]*Avatar2Client
	Url              string
	RoomID           string
}

type AvatarState struct {
	Avatar int
	HP     float32
	MP     float32
	WS     float32
}

const (
	Gs_Module_CorpLvl = iota
	Gs_Module_Equip
	Gs_Module_Equip_Evolution
	Gs_Module_Equip_Mat_Enhance
	Gs_Module_Equip_Trick
	Gs_Module_Equip_StarUp
	Gs_Module_Fashion
	Gs_Module_Arousal
	Gs_Module_DestinyGeneral
	Gs_Module_Jade
	Gs_Module_General
	Gs_Module_General_Rel
	Gs_Module_Hero
	Gs_Module_Title
	Gs_Module_Count
)
