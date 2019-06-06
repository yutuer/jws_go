package gates_enemy_cmd

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

const (
	GatesEnemyCommandMsgTypNull     = iota
	GatesEnemyCommandMsgTypStartAct // 公会开启活动
	GatesEnemyCommandMsgTypStartActASync
	GatesEnemyCommandMsgTypReadyAct       //
	GatesEnemyCommandMsgTypGuildChange    // 公会成员变更通知
	GatesEnemyCommandMsgTypGetAct         // 玩家请求Act的channel
	GatesEnemyCommandMsgTypPlayerLogin    // 玩家通知上线
	GatesEnemyCommandMsgTypPlayerLogout   // 玩家通知下线
	GatesEnemyCommandMsgTypStopAct        // 系统通知停止活动
	GatesEnemyCommandMsgTypFightBegin     // 玩家打杂兵
	GatesEnemyCommandMsgTypFightEnd       // 玩家打杂兵
	GatesEnemyCommandMsgTypFightBossBegin // 玩家打Boss
	GatesEnemyCommandMsgTypFightBossEnd   // 玩家打Boss
	GatesEnemyCommandMsgTypEnterAct       // 玩家进入活动场景
	GatesEnemyCommandMsgTypLeaveAct       // 玩家离开活动场景
	GatesEnemyCommandMsgTypGetActChan
	GatesEnemyCommandMsgTypAddBuff
	GatesEnemyCommandMsgTypDebugOp
)

type GatesEnemyCommandMsg struct {
	Type      int
	GuildID   string
	AccountID string
	EnemyTyp  int
	EnemyIdx  int
	BossIdx   int
	EndTime   int64
	IsSuccess bool
	ResChann  chan<- chan<- GatesEnemyCommandMsg
	OkChann   chan<- GatesEnemyRet
	Members   []helper.AccountSimpleInfo
}

type GatesEnemyRet struct {
	Code        uint32
	RetStrParam []string
}

func GenGatesEnemyRet(code uint32) GatesEnemyRet {
	return GatesEnemyRet{
		Code: code,
	}
}

const (
	RESSuccess              = 0
	RESWarnNoAct            = errCode.GuildGateEnemyRESWarnNoAct
	RESWarnCanNotInAct      = errCode.GuildGateEnemyRESWarnCanNotInAct
	RESWarnEnemyIDErr       = errCode.GuildGateEnemyRESWarnEnemyIDErr
	RESWarnEnemyHasFighting = errCode.GuildGateEnemyRESWarnEnemyHasFighting
	RESErrNoBoss            = errCode.GuildGateEnemyRESErrNoBoss
	RESErrStateErr          = errCode.GuildGateEnemyRESErrStateErr
	RESTimeOut              = errCode.GuildGateEnemyRESTimeOut
	RESWarnBuffAlready      = errCode.GateEnemyBuffAlready
)
