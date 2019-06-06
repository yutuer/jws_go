package simple

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamSimpleSync ..
type ParamSimpleSync struct {
	In uint32
}

//RetSimpleSync ..
type RetSimpleSync struct {
	Out uint32
}

//MethodSimpleSync ..
type MethodSimpleSync struct {
	module.BaseMethod
}

func newMethodSimpleSync(m module.Module) *MethodSimpleSync {
	return &MethodSimpleSync{
		module.BaseMethod{Method: MethodSimpleSyncID, Module: m},
	}
}

//NewParam ..
func (m *MethodSimpleSync) NewParam() module.Param {
	return &ParamSimpleSync{}
}

//NewRet ..
func (m *MethodSimpleSync) NewRet() module.Ret {
	return &RetSimpleSync{}
}

//Do ..
func (m *MethodSimpleSync) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamSimpleSync)

	logs.Info("[%d][%s] MethodSimpleAsync Do In:[%d]", t.GroupID, t.HashSource, param.In)

	ret = &RetSimpleSync{
		Out: param.In * param.In,
	}
	errCode = message.ErrCodeOK
	return
}
