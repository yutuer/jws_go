package worldboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamGetFormationRank ..
type ParamGetFormationRank struct {
	Sid  uint32
	Acid string
}

//RetGetFormationRank ..
type RetGetFormationRank struct {
	Rank   []FormationRankElemInfo
	MyRank FormationRankElemInfo
}

//MethodGetFormationRank ..
type MethodGetFormationRank struct {
	module.BaseMethod
}

func newMethodGetFormationRank(m module.Module) *MethodGetFormationRank {
	return &MethodGetFormationRank{
		module.BaseMethod{Method: MethodGetFormationRankID, Module: m},
	}
}

//NewParam ..
func (m *MethodGetFormationRank) NewParam() module.Param {
	return &ParamGetFormationRank{}
}

//NewRet ..
func (m *MethodGetFormationRank) NewRet() module.Ret {
	return &RetGetFormationRank{}
}

//Do ..
func (m *MethodGetFormationRank) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamGetFormationRank)
	logs.Trace("[WorldBoss] [%d][%s] MethodGetFormationRank Do .. Param %+v", t.GroupID, t.HashSource, param)

	bm := m.ModuleAt().(*WorldBoss)
	rank := bm.res.FormationRankMod.getTop()
	myRank := bm.res.FormationRankMod.getRankInfoByAcid(param.Acid)

	ret = &RetGetFormationRank{
		Rank:   rank,
		MyRank: *myRank,
	}
	logs.Trace("[WorldBoss] [%d][%s] MethodGetFormationRank Do .. Ret %+v", t.GroupID, t.HashSource, ret)
	errCode = message.ErrCodeOK
	return
}
