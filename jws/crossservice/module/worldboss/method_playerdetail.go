package worldboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamPlayerDetail ..
type ParamPlayerDetail struct {
	Sid       uint32
	Acid      string
	CheckAcid string
}

//RetPlayerDetail ..
type RetPlayerDetail struct {
	PlayerInfo PlayerInfo
	Team       TeamInfoDetail
}

//MethodPlayerDetail ..
type MethodPlayerDetail struct {
	module.BaseMethod
}

func newMethodPlayerDetail(m module.Module) *MethodPlayerDetail {
	return &MethodPlayerDetail{
		module.BaseMethod{Method: MethodPlayerDetailID, Module: m},
	}
}

//NewParam ..
func (m *MethodPlayerDetail) NewParam() module.Param {
	return &ParamPlayerDetail{}
}

//NewRet ..
func (m *MethodPlayerDetail) NewRet() module.Ret {
	return &RetPlayerDetail{}
}

//Do ..
func (m *MethodPlayerDetail) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamPlayerDetail)
	logs.Trace("[WorldBoss] [%d][%s] MethodPlayerDetail Do .. Param %+v", t.GroupID, t.HashSource, param)

	bm := m.ModuleAt().(*WorldBoss)
	playerInfo := bm.res.PlayerMod.getPlayerInfo(param.CheckAcid)
	teamInfo := bm.res.PlayerMod.getTeamInfo(param.CheckAcid)
	detail := &RetPlayerDetail{}
	if nil != playerInfo {
		detail.PlayerInfo = *playerInfo
	}
	if nil != teamInfo {
		detail.Team = *teamInfo
	}
	ret = detail
	logs.Trace("[WorldBoss] [%d][%s] MethodPlayerDetail Do .. Ret %+v", t.GroupID, t.HashSource, ret)
	errCode = message.ErrCodeOK
	return
}
