package helper

import (
	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type MatchGameInfo struct {
	AcIDs  []string
	IsHard bool
}

type GameStartInfo struct {
	AcIDs      string
	MServerUrl string
	GameID     string
	Secret     string
	FightCount int
}

type GameStopInfo struct {
	AcIDs       string
	GameID      string
	IsSuccess   bool
	IsHasReward bool
}

type MatchValue struct {
	AccountID string `json:"id"`
	IsHard    bool   `json:"h"`
	CorpLv    uint32 `json:"l"`
}

type FenghuoValue struct {
	AcIDs      []string                      `json:"accounts,omitempty"`
	Shutdown   bool                          `json:"shutdown,omitempty"`
	RoomID     string                        `json:"id,omitempty"`
	AvatarInfo []*helper.Avatar2ClientByJson `json:"avatarinfo,omitempty"`
}

type FenghuoCreateInfo struct {
	RoomID    string `json:"ID"`
	WebsktUrl string `json:"url"`
	CancelUrl string `json:"cancelurl"`
}

type TeamBossData struct {
}

type TeamBossInfo struct {
	AcIDs []string
}

type TBStartFightData struct {
	RoomID    string
	GroupID   uint32
	Data      [][]byte
	Info      []*helper.Avatar2ClientByJson
	AcID      []string
	GID       uint32
	SceneID   string
	Level     uint32
	BossID    string
	TeamTypID uint32
	BoxStatus int
	CostID    string
}

type GVGStartFightData struct {
	Acid1   string
	Acid2   string
	Avatar1 []int
	Avatar2 []int
	Sid     uint
	// TODO by ljz add field
}

type GVGStartFigntRetData struct {
	RoomID    string `json:"ID"`
	WebsktUrl string `json:"url"`
}

func (tbsfd *TBStartFightData) Init() error {
	tbsfd.Info = make([]*helper.Avatar2ClientByJson, len(tbsfd.Data))
	for i, item := range tbsfd.Data {
		logs.Warn("get fight data info: %v", string(item))
		ret := &helper.Avatar2ClientByJson{}
		err := json.Unmarshal(item, ret)
		if err != nil {
			return err
		}
		tbsfd.Info[i] = ret
	}
	return nil
}

type TeamBossStopInfo struct {
	Winner    string
	GroupID   uint32
	GameID    string
	IsSuccess bool
	AcIDs     []string
	Level     uint32
	BoxStatus int
	CostID    string
}

const (
	_ = iota
	Invalid
	Normal
)

type GVGStopInfo struct {
	RoomID string
	Winner string
	AcIDs  []string
	Status int
}

type TeamBossCreateinfo struct {
	GlobalRoomID string `json:"ID"`
	WebsktUrl    string `json:"url"`
}
