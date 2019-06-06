package info

import (
	"errors"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/errorcode"
)

const (
	RoomEventUpdate = iota + 1
	RoomEventChangeMaster
	RoomEventLeave
	RoomEventStartFight
	RoomEventKicked
	RoomEventDismiss
	RoomEventReadyForLevel
	RoomEventMakeReward
	RoomEventReJoin
	FinalDismiss
)

const (
	PlayerStatNoReady = iota
	PlayerStatReady
)

const (
	RoomStatWaitting = iota + 1
	//以上Room状态允许玩家加入
	RoomStatFightting
)

const (
	RoomSubLevelDeafult    = 0
	RoomSubLevelSelected   = 1
	RoomSubLevelDoneReward = 2
)

var (
	RoomErrCode_CTX_TIMEOUT = errorcode.New("Fenghuo Room Context timeout.", 200)
	RoomErrCode_UNKNOWN     = errorcode.New("Fenghuo Room Unknown.", 201)

	RoomWarnCode_ENTER_ROOM_NUM_WRONG     = errorcode.New("Fenghuo Enter Room but room num is wrong.", errCode.FenghuoRoomNumWarn)
	RoomWarnCode_ENTER_ROOM_ALREADYINROOM = errorcode.New("Fenghuo Enter Room but already in room.", errCode.FenghuoAlreadyInRoom)
	RoomWarnCode_ENTER_ROOM_FULL          = errorcode.New("Fenghuo Enter Room but room is full.", errCode.FenghuoRoomFull)
	RoomWarnCode_ENTER_ROOM_NOTJOINABLE   = errorcode.New("Fenghuo Enter Room but NotJoinable status, room status is fighting.", errCode.FenghuoNotJoinable)

	RoomWarnCode_MASTER_WAITING_OTHERS_READY = errorcode.New("Fenghuo Enter Room still waitint others ready.", errCode.FenghuoWaitOthersReady)
	RoomWarnCode_NOMONEY                     = errorcode.New("Fenghuo No Enough Money.", errCode.FenghuoNoMoney)

	RoomErr_CTX_TIMEOUT     = errors.New("Fenghuo Room Context timeout.")
	RoomErr_ShouldNotBeHere = errors.New("Fenghuo Room Should not be here.")

	RoomErr_Fight_NotSelectSubLevel = errors.New("Fenghuo Room player should select sublevel and fight firstly.")

	RoomErr_Fight_PlayerNotFound = errors.New("Fenghuo Room player id cant be found.")
	RoomErr_Fight_DupMakeReward  = errors.New("Fenghuo Room player has already picked rewards.")

	RoomErr_GetRoomInfo_NotFound = errors.New("Fenghuo Room Room Num not found.")
)
