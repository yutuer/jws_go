package csrob

import (
	"time"
)

//Record类型
const (
	RecordRob = iota
	RecordBeRob
	RecordDoneDriving
	RecordDoneHelp
)

//PlayerParam ..
type PlayerParam struct {
	Acid          string
	GuildID       string
	Name          string
	GuildPosition int
	Vip           uint32

	FormationNew      []int
	FormationTeamFunc func([]int) []HeroInfo
}

//PlayerInfo 玩家基础信息
type PlayerInfo struct {
	Acid    string `json:"acid,omitempty"`
	GuildID string `json:"guild_id,omitempty"`

	GradeRefresh PlayerGradeRefresh `json:"grade_refresh,omitempty"`

	CurrFormation []int `json:"curr_formation,omitempty"`

	Count PlayerCount `json:"count,omitempty"`

	NextCarID uint32 `json:"next_car_id,omitempty"`

	CarList []PlayerCarListElem `json:"car_list,omitempty"`

	LastRob CacheForMarquee `json:"for_marquee,omitempty"`
	// LastHelpEnd int64           `json:"last_help_end,omitempty"`

	UpdateTime int64 `json:"update_time,omitempty"`
}

func genPlayerInfo(param *PlayerParam) *PlayerInfo {
	return &PlayerInfo{
		Acid:    param.Acid,
		GuildID: param.GuildID,

		GradeRefresh:  PlayerGradeRefresh{},
		CurrFormation: []int{},

		NextCarID: 1,

		CarList: []PlayerCarListElem{},

		LastRob: CacheForMarquee{
			CacheTime: 0,
			Goods:     map[string]uint32{},
		},

		UpdateTime: 0,
	}
}

type PlayerCount struct {
	Build uint32 `json:"build"`
	Help  uint32 `json:"help"`
	Rob   uint32 `json:"rob"`
}

type PlayerRecord struct {
	Type       int               `json:"type"` //日志类型
	Timestamp  int64             `json:"timestamp"`
	DriverID   string            `json:"driver_id"`
	DriverName string            `json:"-"`
	RobberID   string            `json:"robber_id"`
	RobberName string            `json:"-"`
	HelperID   string            `json:"helper_id"`
	HelperName string            `json:"-"`
	Grade      uint32            `json:"grade"`
	Goods      map[string]uint32 `json:"goods"`
}

type PlayerAppeal struct {
	Acid       string   `json:"acid"`
	Name       string   `json:"-"`
	CarID      uint32   `json:"car_id"`
	Grade      uint32   `json:"grade"`       //品质
	AppealTime int64    `json:"appeal_time"` //请求时间
	EndStamp   int64    `json:"end_stamp"`   //粮车结束时间戳
	HasHelper  bool     `json:"-"`           //已有护送
	HelperIsMe bool     `json:"-"`           //就是我护送的
	Robbers    []string `json:"-"`           //抢劫人
}

type PlayerEnemy struct {
	Acid    string     `json:"acid"`
	Name    string     `json:"-"`
	Count   uint32     `json:"count"`
	CurrCar *PlayerRob `json:"-"`
}

type PlayerRob struct {
	CarID uint32
	Info  PlayerRobInfo

	Robbing bool
	Robbers []string

	Helper *PlayerRobHelper
	Reward *PlayerRobReward

	Acid      string
	Name      string
	GuildID   string
	GuildName string
	GuildPos  int

	AlreadyAppeal []string
}

type PlayerRobInfo struct {
	CarID      uint32     `json:"car_id"`
	Grade      uint32     `json:"grade"`       //品质
	Team       []HeroInfo `json:"team"`        //护送队伍
	StartStamp int64      `json:"start_stamp"` //粮车结束时间戳
	EndStamp   int64      `json:"end_stamp"`   //粮车结束时间戳
}

type PlayerCarListElem struct {
	PlayerRobInfo
	AlreadySendHelp []string
}

type PlayerRobHelper struct {
	Acid string     `json:"acid"`
	Name string     `json:"-"`
	Team []HeroInfo `json:"team"`
}

type PlayerRobReward struct {
	Time   int64             `json:"time"`    //发奖时间
	BeDark bool              `json:"be_dark"` //被暗格
	Goods  map[string]uint32 `json:"goods"`   //发的奖励
}

type HeroInfo struct {
	Idx       int    `json:"idx"`        // id
	Attr      []byte `json:"attr"`       // 战斗属性
	StarLevel int    `json:"star_level"` // 星级
	Gs        int64  `json:"gs"`         // 战力

	AvatarSkills   []int64  `json:"skills"`         // 武将技能等级
	AvatarFashion  []string `json:"avatar_equips"`  // 武将时装
	HeroSwing      int      `json:"swing"`          // 翅膀
	MagicPetfigure uint32   `json:"magicpetfigure"` // 灵宠外观
	PassiveSkillId []string `json:"pskillid"`       // 被动技能
	CounterSkillId []string `json:"cskillid"`
	TriggerSkillId []string `json:"tskillid"`
	DesSkills      []int64  `json:"dgss"` // 神兽技能
}

type PlayerRewardInfo struct {
	UpdateTime  int64             `json:"update_time"`  //上次数据刷新时间
	DropHistory map[string]uint32 `json:"drop_history"` //掉落历史, 用于暗格
}

type CacheForMarquee struct {
	CacheTime int64 `json:"ct"`

	CarID      uint32 `json:"car_id"`
	Driver     string `json:"driver"`
	DriverName string `json:"-"`
	Robber     string `json:"robber,omitempty"`
	RobberName string `json:"-"`

	HasHelper  bool   `json:"has_helper"`
	Helper     string `json:"helper"`
	HelperName string `json:"-"`

	Grade uint32            `json:"grade"`
	Goods map[string]uint32 `json:"goods"`

	Sent bool `json:"sent"`
}

func (c *CacheForMarquee) setCacheForMarquee(robber string, data *PlayerRob, goods map[string]uint32) {
	c.Sent = false
	c.CacheTime = time.Now().Unix()

	c.CarID = data.CarID
	c.Driver = data.Acid
	c.Robber = robber
	if nil != data.Helper {
		c.HasHelper = true
		c.Helper = data.Helper.Acid
	} else {
		c.HasHelper = false
	}

	c.Grade = data.Info.Grade
	c.Goods = goods
}

//PlayerGradeRefresh ..
type PlayerGradeRefresh struct {
	CurrGrade uint32 `json:"curr_grade,omitempty"`
	// CurrGradeRefresh   uint32 `json:"curr_grade_refresh,omitempty"`
	CarSumGradeRefresh uint32 `json:"car_sum_grade_refresh,omitempty"`
	CarSumGradeCost    uint32 `json:"car_sum_grade_cost,omitempty"`
	LastBuildTime      int64  `json:"last_build_time,omitempty"`
}

func (p *PlayerGradeRefresh) reset() {
	p.CurrGrade = 0
	// p.CurrGradeRefresh = 0
	p.CarSumGradeCost = 0
	p.CarSumGradeRefresh = 0
}

//PlayerStatus ..
type PlayerStatus struct {
	VIP               uint32   `redis:"vip"`                 //玩家VIP等级
	AcceptAppealCount uint32   `redis:"accept_appeal_count"` //已接受求援的数量
	AutoAcceptBottom  []uint32 `redis:"-"`                   //自动接受求援的下限

	LastUpdate int64 `redis:"last_update"` //上次更新的时间
}
