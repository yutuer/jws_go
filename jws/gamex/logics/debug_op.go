package logics

import (
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice/worldboss"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func init() {
	regDebugOpHandle(handlerDebugOp{"CrossServiceWorldBossGetInfo", DebugCrossServiceWorldBossGetInfo})
	regDebugOpHandle(handlerDebugOp{"CrossServiceWorldBossJoin", DebugCrossServiceWorldBossJoin})
	regDebugOpHandle(handlerDebugOp{"CrossServiceWorldBossAttack", DebugCrossServiceWorldBossAttack})
	regDebugOpHandle(handlerDebugOp{"CrossServiceWorldBossLeave", DebugCrossServiceWorldBossLeave})
	regDebugOpHandle(handlerDebugOp{"CrossServiceWorldBossGetRank", DebugCrossServiceWorldBossGetRank})
}

//DebugCrossServiceWorldBossGetInfo ..
func DebugCrossServiceWorldBossGetInfo(p *Account, req *RequestDebugOp) string {
	logs.Debug("DebugCrossServiceWorldBossGetInfo Begin")

	status, errcode, err := worldboss.GetInfo(p.AccountID.ShardId, p.AccountID.String())
	if nil != err {
		logs.Warn("DebugCrossServiceWorldBossGetInfo GetInfo Failed, ErrCode %d, error : %v", errcode, err)
		return "Fail"
	}

	logs.Debug("DebugCrossServiceWorldBossGetInfo End, RoomStatus %+v", status)
	return "OK"
}

//DebugCrossServiceWorldBossJoin ..
func DebugCrossServiceWorldBossJoin(p *Account, req *RequestDebugOp) string {
	logs.Debug("DebugCrossServiceWorldBossJoin Begin")

	playerInfo := &worldboss.PlayerInfo{
		Acid: p.AccountID.String(),
		Sid:  uint32(p.AccountID.ShardId),
		Name: p.Profile.Name,
		Vip:  p.Profile.GetVipLevel(),
	}
	status, errcode, err := worldboss.Join(p.AccountID.ShardId, p.AccountID.String(), playerInfo)
	if nil != err {
		logs.Warn("DebugCrossServiceWorldBossJoin GetInfo Failed, ErrCode %d, error : %v", errcode, err)
		return "Fail"
	}

	logs.Debug("DebugCrossServiceWorldBossJoin End, RoomStatus %+v", status)
	return "OK"
}

//DebugCrossServiceWorldBossAttack ..
func DebugCrossServiceWorldBossAttack(p *Account, req *RequestDebugOp) string {
	logs.Debug("DebugCrossServiceWorldBossAttack Begin")

	attackInfo := &worldboss.AttackInfo{
		Damage: uint64(req.P2),
		Level:  uint32(req.P1),
	}
	status, errcode, err := worldboss.Attack(p.AccountID.ShardId, p.AccountID.String(), attackInfo)
	if nil != err {
		logs.Warn("DebugCrossServiceWorldBossAttack GetInfo Failed, ErrCode %d, error : %v", errcode, err)
		return "Fail"
	}

	logs.Debug("DebugCrossServiceWorldBossAttack End, RoomStatus %+v", status)
	return "OK"
}

//DebugCrossServiceWorldBossLeave ..
func DebugCrossServiceWorldBossLeave(p *Account, req *RequestDebugOp) string {
	logs.Debug("DebugCrossServiceWorldBossLeave Begin")

	teamInfo := &worldboss.TeamInfoDetail{}
	status, errcode, err := worldboss.Leave(p.AccountID.ShardId, p.AccountID.String(), teamInfo, false)
	if nil != err {
		logs.Warn("DebugCrossServiceWorldBossLeave GetInfo Failed, ErrCode %d, error : %v", errcode, err)
		return "Fail"
	}

	logs.Debug("DebugCrossServiceWorldBossLeave End, RoomStatus %+v", status)
	return "OK"
}

//DebugCrossServiceWorldBossGetRank ..
func DebugCrossServiceWorldBossGetRank(p *Account, req *RequestDebugOp) string {
	logs.Debug("DebugCrossServiceWorldBossGetRank Begin")

	status, errcode, err := worldboss.GetRank(p.AccountID.ShardId, p.AccountID.String())
	if nil != err {
		logs.Warn("DebugCrossServiceWorldBossGetRank GetRank Failed, ErrCode %d, error : %v", errcode, err)
		return "Fail"
	}

	logs.Debug("DebugCrossServiceWorldBossGetRank End, RoomStatus %+v", status)
	return "OK"
}
