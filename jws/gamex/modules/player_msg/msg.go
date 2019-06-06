package player_msg

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/info"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/common/guild_player_rank"
)

const (
	PlayerMsgGatesEnemyDataCode       = "MSG/GatesEnemy"
	PlayerMsgGuildInfoSyncCode        = "MSG/PlayerGuildInfoSync"
	PlayerMsgGuildBossSyncCode        = "MSG/PlayerGuildBossSync"
	PlayerMsgGuildApplyInfoSyncCode   = "MSG/PlayerGuildApplySync"
	PlayerMsgGuildScienceInfoSyncCode = "MSG/PlayerGuildScienceSync"
	PlayerMsgGuildRedPacketSyncCode   = "MSG/PlayerGuildRedPacketSync"
	PlayerMsgRedPoint                 = "MSG/PlayerRedPoint"
	PlayerMsgGank                     = "MSG/PlayerGank"
	PlayerMsgGVEGameStartCode         = "MSG/GVEGameStart"
	PlayerMsgGVEGameStopCode          = "MSG/GVEGameStop"
	PlayerMsgTeamPvpRankChgCode       = "MSG/TeamPvpRankChg"
	PlayerMsgTitleCode                = "MSG/TitleChg"
	PlayerMsgRooms                    = "MSG/Rooms"
	PlayerMsgRoomEvent                = "MSG/RoomEv"
	PlayerMsgGVGStartCode             = "MSG/GVGStart"
	PlayerMsgOnLoginCode              = "MSG/OnLogin"

	PlayerMsgCSRobSetFormation  = "MSG/CSRobSetFormation"
	PlayerMsgCSRobRefreshSelf   = "MSG/CSRobRefreshSelf"
	PlayerMsgCSRobAddPlayerRank = "MSG/CSRobAddPlayerRank"

	PlayerMsgTeamBossKicked = "MSG/TBKicked"
	PlayerMsgTeamStartFight = "MSG/TBStartFight"
	PlayerMsgRefreshRoom    = "MSG/TBRefreshRoom"

	Playerh5MsgCostDiamond = "MSG/h5CostDiamond"
)

const (
	PlayerMsgTypNull = iota
	PlayerMsgTypGuildChange
	PlayerMsgTypGatesEnemyPush
)

type PlayerCommonMsg struct {
	Typ     int
	Num     []int
	Text    []string
	ResChan chan PlayerCommonMsg
}

type PlayerMsgGatesEnemyData struct {
	EnemyInfo      [][]byte                           `codec:"ei"`
	State          int                                `codec:"s"`
	StateOverTime  int64                              `codec:"st"`
	KillPoint      int                                `codec:"kp"`
	Point          int                                `codec:"gp"`
	BossMax        int                                `codec:"b"`
	AcID           []string                           `codec:"acIDs_"`
	AvatarID       []int                              `codec:"geraIDs_"`
	Names          []string                           `codec:"gernames_"`
	Points         []int                              `codec:"gerpoints_"`
	Fashion        [][helper.FashionPart_Count]string `codec:"gerfashion_"`
	WeaponStartLvl []uint32                           `codec:"gerwstar_"`
	EqStartLvl     []uint32                           `codec:"gerestar_"`
	TitleOn        []string                           `codec:"gertitle_"`
	GetRewardTime  []int64                            `codec:"getrtime_"`
	MemStats       []int                              `codec:"gerstats_"`
	Swing          []int                              `codec:"geswing"`
	MagicPetfigure []uint32                           `codec:"geMagicpet"`
	BuffCurLv      uint32                             `codec:"bufflv_"`
	BuffMemAcid    [helper.GateEnemyBuffCount]string  `codec:"buffmemacid_"`
	BuffMemName    [helper.GateEnemyBuffCount]string  `codec:"buffmemname_"`
}

type GatesEnemyData struct {
	EnemyInfo     [][]byte                               `json:"ei"`
	State         int                                    `json:"s"`
	StateOverTime int64                                  `json:"st"`
	KillPoint     int                                    `json:"kp"`
	Point         int                                    `json:"gp"`
	BossMax       int                                    `json:"b"`
	PlayerRank    guild_player_rank.PlayerSimpleInfoRank `json:"rank"`
	BuffCurLv     uint32                                 `json:"bufflv_"`
	BuffMemAcid   [helper.GateEnemyBuffCount]string      `json:"buffmemacid_"`
	BuffMemName   [helper.GateEnemyBuffCount]string      `json:"buffmemname_"`
}

func (g *GatesEnemyData) ToClient() *PlayerMsgGatesEnemyData {
	res := new(PlayerMsgGatesEnemyData)
	res.EnemyInfo = g.EnemyInfo
	res.State = g.State
	res.StateOverTime = g.StateOverTime
	res.KillPoint = g.KillPoint
	res.Point = g.Point
	res.BossMax = g.BossMax

	memLen := g.PlayerRank.Len()
	res.AcID = make([]string, 0, memLen)
	res.AvatarID = make([]int, 0, memLen)
	res.Names = make([]string, 0, memLen)
	res.Points = make([]int, 0, memLen)
	res.Fashion = make([][helper.FashionPart_Count]string, 0, memLen)
	res.WeaponStartLvl = make([]uint32, 0, memLen)
	res.EqStartLvl = make([]uint32, 0, memLen)
	res.TitleOn = make([]string, 0, memLen)
	res.GetRewardTime = make([]int64, 0, memLen)
	res.MemStats = make([]int, 0, memLen)
	res.Swing = make([]int, 0, memLen)
	res.MagicPetfigure = make([]uint32, 0, memLen)
	for i := 0; i < memLen; i++ {
		r := g.PlayerRank.GetSimpleInfo(i)
		res.AcID = append(res.AcID, r.AccountID)
		res.AvatarID = append(res.AvatarID, r.CurrAvatar)
		res.Names = append(res.Names, r.Name)
		res.Points = append(res.Points, int(g.PlayerRank.Rank.GetSorce(i)))
		res.Fashion = append(res.Fashion, r.FashionEquips)
		res.WeaponStartLvl = append(res.WeaponStartLvl, r.WeaponStartLvl)
		res.EqStartLvl = append(res.EqStartLvl, r.EqStartLvl)
		res.TitleOn = append(res.TitleOn, r.TitleOn)
		res.GetRewardTime = append(res.GetRewardTime, r.Other.Pi[0])
		res.MemStats = append(res.MemStats, int(r.Other.Pi[1]))
		res.Swing = append(res.Swing, int(r.Swing))
		res.MagicPetfigure = append(res.MagicPetfigure, uint32(r.MagicPetfigure))
	}
	res.BuffMemAcid = g.BuffMemAcid
	res.BuffMemName = g.BuffMemName
	res.BuffCurLv = g.BuffCurLv
	return res
}

func (p *PlayerMsgGatesEnemyData) IsNil() bool {
	return p.State == 0 &&
		p.StateOverTime == 0 &&
		p.KillPoint == 0 &&
		p.Point == 0 &&
		p.AcID == nil
}

func (p *PlayerMsgGatesEnemyData) GetBuffCount(acid string) int {
	var res int
	for _, _acid := range p.BuffMemAcid {
		if _acid == acid {
			res++
		}
	}
	return res
}

func (p *GatesEnemyData) GetGetRewardTime(acID string) int64 {
	rank := p.PlayerRank.GetRank(acID)

	if rank < 0 {
		return 0
	}

	sorce := p.PlayerRank.Rank.GetSorce(rank)
	if sorce <= 0 {
		return 0
	}

	simpleInfo := p.PlayerRank.GetSimpleInfo(p.PlayerRank.GetRank(acID))
	if simpleInfo == nil {
		return 0
	}

	return simpleInfo.Other.Pi[0]
}

func (p *GatesEnemyData) GetRewardCount() (int, int) {
	//allReward := p.PlayerRank.Len()
	allReward := 0
	gettedReward := 0

	for i := 0; i < p.PlayerRank.Len(); i++ {
		simpleInfo := p.PlayerRank.GetSimpleInfo(i)
		if simpleInfo != nil && simpleInfo.Other.Pi[0] > 0 {
			gettedReward++
		}
		if p.HasReward(simpleInfo.AccountID) {
			allReward++
		}
	}

	return gettedReward, allReward
}

func (p *GatesEnemyData) HasReward(acID string) bool {
	rank := p.PlayerRank.GetRank(acID)

	if rank < 0 {
		return false
	}

	sorce := p.PlayerRank.Rank.GetSorce(rank)
	if sorce <= 0 {
		return false
	}

	return true
}

type PlayerGuildInfoUpdate struct {
	GuildUUID     string   `codec:"id"`
	GuildName     string   `codec:"name"`
	GuildPosition int      `codec:"pos"`
	GuildLv       int      `codec:"lv"`
	LeaveTime     int64    `codec:"leat"`
	NextJoinTime  int64    `codec:"njt"`
	AssignID      []string `codec:"assign_id"`
	AssignTimes   []int64  `codec:"assign_times"`
}

type PlayerGuildBossUpdate struct {
	GuildUUID string              `codec:"id"`
	Info      info.ActBoss2Client `codec:"info"`
}

type PlayerGuildApplyUpdate struct {
	HasApplyCanApprove bool `codec:"haca"`
}

type PlayerRedPoint struct {
	RedPointId int `codec:"rdp"`
}

type PlayerMsgGVEGameStart struct {
	GameID        string `codec:"id"`
	GameServerUrl string `codec:"url"`
	GameSecret    string `codec:"s"`
	IsBot         bool   `codec:"is_bot"`
}

type PlayerMsgGVEGameStop struct {
	GameID      string `codec:"id"`
	IsHasReward bool   `codec:"r"`
	IsSuccess   bool   `codec:"s"`
}

type PlayerMsgTeamPvpRankChg struct {
	Rank int `codec:"r"`
}

type PlayerGank struct {
	LogTS int64 `codec:"ts"`
}

type DefaultMsg struct {
}

type ActivePlayerTitle struct {
	Title string `codec:"tl"`
}

type RoomsSyncInfo struct {
	RoomNew [][]byte `codec:"n"`
	RoomDel []int    `codec:"d"`
}

type RoomEventNotify struct {
	Type int    `codec:"t"`
	Room []byte `codec:"r"`
}

type PlayerMsgGVGStart struct {
	PlayerInfo      [helper.GVG_AVATAR_COUNT]helper.AvatarState
	DestinySkill    [helper.DestinyGeneralSkillMax]int
	EnemyAcID       string
	EnemyPlayerInfo [helper.GVG_AVATAR_COUNT]helper.AvatarState
	EDestinySkill   [helper.DestinyGeneralSkillMax]int
	EnemyData       [helper.GVG_AVATAR_COUNT]*helper.Avatar2Client
	URL             string
	RoomID          string
}

type PlayerCSRobSetFormation struct {
}

type PlayerCSRobRefreshSelf struct {
}

type PlayerCSRobAddPlayerRank struct {
}

type PlayerCostDiamondInfo struct {
	Roleid   string `codec:"roleid"`
	Serverid string `codec:"serverid"`
	DiamNum  string `json:"diamond_num"`
}
