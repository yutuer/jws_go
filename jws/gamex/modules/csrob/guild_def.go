package csrob

type GuildInfo struct {
	GuildID   string `json:"guild_id,omitempty"`
	GuildName string `json:"-,omitempty"`
	BestGrade uint32 `json:"-,omitempty"`

	UpdateTime int64 `json:"update_time,omitempty"`
}

func genGuildInfo(guid string, name string) *GuildInfo {
	return &GuildInfo{
		GuildID:    guid,
		GuildName:  name,
		UpdateTime: 0,
	}
}

type GuildEnemy struct {
	GuildID   string `json:"guild_id,omitempty"`
	GuildName string `json:"-,omitempty"`
	Count     uint32 `json:"count,omitempty"`
	BestGrade uint32 `json:"-,omitempty"`
}

type GuildTeam struct {
	Acid       string     `json:"acid,omitempty"`
	Name       string     `json:"-,omitempty"`
	Hero       []HeroInfo `json:"hero,omitempty"`
	AutoAccept []uint32   `json:"-,omitempty"`
}

type GuildRobElem struct {
	Acid       string `json:"acid,omitempty"`        //出车的人
	CarID      uint32 `json:"car_id,omitempty"`      //车子ID
	StartStamp int64  `json:"start_stamp,omitempty"` //粮车开始时间戳
	EndStamp   int64  `json:"end_stamp,omitempty"`   //粮车结束时间戳
}

type GuildCommonStatus struct {
	RecommendRefreshTime int64 `json:"recommend_refresh_time,omitempty"` //上次刷新推荐列表的时间
	WeekRewardTime       int64 `json:"week_reward_time,omitempty"`       //上次发周奖励的时间
	// ClearRobRankTime     int64 `json:"clear_rob_rank_time,omitempty"`    //上次清除抢夺排行的时间
}

//GuildRankElem ..
type GuildRankElem struct {
	GuildID  string
	RobCount uint32
	RobTime  int64

	GuildName   string
	GuildMaster string
	Rank        uint32
}
