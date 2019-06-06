package simple

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamSimpleAsync ..
type ParamSimpleAsync struct {
	Hello string
}

//RetSimpleAsync ..
type RetSimpleAsync struct {
}

//MethodSimpleAsync ..
type MethodSimpleAsync struct {
	module.BaseMethod
}

func newMethodSimpleAsync(m module.Module) *MethodSimpleAsync {
	return &MethodSimpleAsync{
		module.BaseMethod{Method: MethodSimpleAsyncID, Module: m},
	}
}

//NewParam ..
func (m *MethodSimpleAsync) NewParam() module.Param {
	return &ParamSimpleAsync{}
}

//NewRet ..
func (m *MethodSimpleAsync) NewRet() module.Ret {
	return &RetSimpleAsync{}
}

//Do ..
func (m *MethodSimpleAsync) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamSimpleAsync)

	logs.Info("[%d][%s] MethodSimpleAsync Do Hello:[%s]", t.GroupID, t.HashSource, param.Hello)

	ret = nil
	errCode = message.ErrCodeOK
	return
}
