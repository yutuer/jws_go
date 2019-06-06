package csrob

//RewardBoxElem 发奖信息
type RewardBoxElem struct {
	Acid     string `json:"acid"`
	CarID    uint32 `json:"car_id"`
	EndStamp int64  `json:"end_stamp"`
}

//RewardWeek 每周结算奖励
type RewardWeek struct {
	Time    int64    `json:"time,omitempty"`
	Sid     uint     `json:"sid,omitempty"`
	GuildID string   `json:"guild_id,omitempty"`
	Members []string `json:"members,omitempty"`
	Count   uint32   `json:"count,omitempty"`
}
