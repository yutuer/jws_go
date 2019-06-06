package worldboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamGetRank ..
type ParamGetRank struct {
	Sid  uint32
	Acid string
}

//RetGetRank ..
type RetGetRank struct {
	Rank   []DamageRankElemInfo
	MyRank DamageRankElemInfo
}

//MethodGetRank ..
type MethodGetRank struct {
	module.BaseMethod
}

func newMethodGetRank(m module.Module) *MethodGetRank {
	return &MethodGetRank{
		module.BaseMethod{Method: MethodGetRankID, Module: m},
	}
}

//NewParam ..
func (m *MethodGetRank) NewParam() module.Param {
	return &ParamGetRank{}
}

//NewRet ..
func (m *MethodGetRank) NewRet() module.Ret {
	return &RetGetRank{}
}

//Do ..
func (m *MethodGetRank) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamGetRank)
	logs.Trace("[WorldBoss] [%d][%s] MethodGetRank Do .. Param %+v", t.GroupID, t.HashSource, param)

	bm := m.ModuleAt().(*WorldBoss)
	rank := bm.res.RankDamageMod.getTop()
	myRank := bm.res.RankDamageMod.getMyRank(param.Acid)

	ret = &RetGetRank{
		Rank:   rank,
		MyRank: *myRank,
	}
	logs.Trace("[WorldBoss] [%d][%s] MethodGetRank Do .. Ret %+v", t.GroupID, t.HashSource, ret)
	errCode = message.ErrCodeOK
	return
}
