package helper

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	RoomDefaultSource = "room_default_source"
)

const RoomPlayerMaxCount = 2

const (
	TBPlayerStateIdle  = 0
	TBPlayerStateReady = 1
)

const (
	TBRoomStateOpen = 0
	TBRoomStateLock = 1
	TBRoomStateFull = 2
	//TBRoomStateBusy = 4
)

const (
	BoxStatusLow     = 0
	BoxStatusAdvance = 1
)

const (
	TBRoomIdle  = 0
	TBRoomFight = 1
)

const (
	TBBoxStateAdvance = 1
)

const (
	RetCodeSuccess = iota
	RetCodeFail
	RetCodeRoomNotExist
	RetCodePlayerNotInRoom
	RetCodeOptInvalid
	RetCodeOptLimitPermission
	RetCodeStartFightFailed
	RetCodePositionOccupied
	RetCodeRoomPlayerFull
	RetCodeRoomInBattle
	RetCodeAlreadyTickRedBox
	RetCodeKickFightingRoom
	RetCodeRoomCantEntry
	RetCodeReadyFailed
	RetCodeDataError
)

type StartFightInfo struct {
	RoomID     string
	AcID       string
	BattleInfo []byte
}

type CreateRoomInfo struct {
	RoomLevel  uint32
	BossID     string
	TeamTypID  uint32
	SceneID    string
	RoomStatus int
	JoinInfo   PlayerJoinInfo
}

type PlayerJoinInfo struct {
	PlayerDetailInfo []byte // detail info
	AcID             string
	Sid              uint
	GS               int
	Avatar           int
	Name             string
	Level            int
	VIP              int
	StarLevel        int
	IsInvited        bool
}

type JoinRoomInfo struct {
	RoomID   string
	JoinInfo PlayerJoinInfo
}

type RoomListInfo struct {
	RoomLevel uint32
}

type SelectAvatar struct {
	AcID     string
	Wing     int
	Fashion  []string
	MagicPet int
}
type RetInfo struct {
}

type JoinRoomRetInfo struct {
	Code int
	Info RoomDetailInfo
}

type RoomDetailInfo struct {
	RoomID       string
	LeadID       string
	BossID       string
	TeamTypID    uint32
	SceneID      string
	SimpleInfo   []PlayerSimpleInfo
	RoomStatus   int
	BoxStatus    int
	Level        uint32
	PositionAcID [RoomPlayerMaxCount]string
}

type PlayerSimpleInfo struct {
	AcID            string
	Sid             uint
	GS              int
	Avatar          int
	BattleAvatar    int
	Name            string
	Wing            int
	Fashion         []string
	MagicPet        int
	ExclusiveWeapon string
	Status          int
	Level           int
	VIP             int
	StarLevel       int
	CompressGS      int
	InBattle        bool
}

type RoomListRetInfo struct {
	Code int
	List []RetRoomList
}

type RetRoomList struct {
	LeadName     string
	PlayerCount  int
	RoomStatus   int
	Level        uint32
	BattleAvatar []RoomListAvatar
	RoomID       string
	ServerID     uint
}

type RoomListAvatar struct {
	Avatar int
	StarLv int
}

type CreateRoomRetInfo struct {
	Code   int
	RoomID string
}

type StartFightRetInfo struct {
	Code         int
	ServerUrl    string
	GlobalRoomID string
}

type ReadyFightInfo struct {
	RoomID     string
	AcID       string
	BattleInfo []byte
	Status     int
}

type ReadyFightRetInfo struct {
	Code   int
	Status int
}

// if OptAcID == TgtAcID represent leave, otherwise represent kick
type LeaveRoomInfo struct {
	OptAcID string
	TgtAcID string
	RoomID  string
}

type LeaveRoomRetInfo struct {
	Code  int
	Param LeaveRoomParam
}

type LeaveRoomParam struct {
	RoomID    string
	LeaveTime int64
	TgtAcID   string
	IsRefresh bool
}

type ChangeAvatarInfo struct {
	RoomID          string
	AcID            string
	BattleAvatar    int
	Wing            int
	Fashion         []string
	MagicPet        int
	ExclusiveWeapon string
	Level           int
	StarLevel       int
	Position        int
	BattleInfo      []byte
	CompressGs      int
}

type ChangeAvatarRetInfo struct {
	Code int
}

type CostAdvanceInfo struct {
	RoomID    string
	BoxStatus int
	AcID      string
}

type CostAdvanceRetInfo struct {
	Code int
}

type ChangeRoomStatusInfo struct {
	RoomID     string
	RoomStatus int
	AcID       string
}

type ChangeRoomStatusRetInfo struct {
	Code       int
	RoomStatus int
}

type EndFightInfo struct {
	RoomID       string
	GlobalRoomID string
	AcID         string
}

type EndFightRetInfo struct {
	Code      int
	HasReward bool
	HasRedBox bool
	IsCost    bool
	Level     uint32
}

type GetPlayerDetailInfo struct {
	RoomID string
	AcID   string
	TgtID  string
}

type GetPlayerDetailRetInfo struct {
	Code   int
	Detail []byte
}

func RoomID2Global(roomid string, gid uint32, groupid uint32, nowT int64) string {
	return fmt.Sprintf("%d:%d:%s:%d", gid, groupid, roomid, nowT)
}

func Global2RoomID(globalRoomID string) string {
	return strings.Split(globalRoomID, ":")[2]
}

func ParseRoomID(roomID string) (uint32, int, error) {
	subs := strings.Split(roomID, "_")
	level, err := strconv.Atoi(subs[0])
	if err != nil {
		return 0, 0, err
	}
	id, err := strconv.Atoi(subs[1])
	if err != nil {
		return 0, 0, err
	}
	return uint32(level), int(id), nil
}
