package worldboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamGetInfo ..
type ParamGetInfo struct {
	Sid  uint32
	Acid string
}

//RetGetInfo ..
type RetGetInfo struct {
	Boss     BossStatus
	MyPos    uint32
	MyDamage uint64
}

//MethodGetInfo ..
type MethodGetInfo struct {
	module.BaseMethod
}

func newMethodGetInfo(m module.Module) *MethodGetInfo {
	return &MethodGetInfo{
		module.BaseMethod{Method: MethodGetInfoID, Module: m},
	}
}

//NewParam ..
func (m *MethodGetInfo) NewParam() module.Param {
	return &ParamGetInfo{}
}

//NewRet ..
func (m *MethodGetInfo) NewRet() module.Ret {
	return &RetGetInfo{}
}

//Do ..
func (m *MethodGetInfo) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamGetInfo)
	logs.Trace("[WorldBoss] [%d][%s] MethodGetInfo Do .. Param %+v", t.GroupID, t.HashSource, param)

	bm := m.ModuleAt().(*WorldBoss)
	boss := bm.res.BossMod.getCurrBossStatus()
	myRank := bm.res.RankDamageMod.getMyRank(param.Acid)

	ret = &RetGetInfo{
		Boss:     *boss,
		MyPos:    myRank.Pos,
		MyDamage: myRank.Damage,
	}
	logs.Trace("[WorldBoss] [%d][%s] MethodGetInfo Do .. Ret %+v", t.GroupID, t.HashSource, ret)
	errCode = message.ErrCodeOK
	return
}
