package simpledynamic

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamGetStatus ..
type ParamGetStatus struct {
}

//RetGetStatus ..
type RetGetStatus struct {
	Status uint32
}

//MethodGetStatus ..
type MethodGetStatus struct {
	module.BaseMethod
}

func newMethodGetStatus(m module.Module) *MethodGetStatus {
	return &MethodGetStatus{
		module.BaseMethod{Method: MethodGetStatusID, Module: m},
	}
}

//NewParam ..
func (m *MethodGetStatus) NewParam() module.Param {
	return &ParamGetStatus{}
}

//NewRet ..
func (m *MethodGetStatus) NewRet() module.Ret {
	return &RetGetStatus{}
}

//Do ..
func (m *MethodGetStatus) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	logs.Info("[%d][%s] MethodSimpleAsync Do", t.GroupID, t.HashSource)

	ret = &RetGetStatus{
		Status: m.Module.(*SimpleDynamic).status,
	}
	errCode = message.ErrCodeOK
	return
}
