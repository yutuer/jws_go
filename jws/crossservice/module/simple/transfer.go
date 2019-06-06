package simple

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamSimpleTransfer ..
type ParamSimpleTransfer struct {
	ShardID uint32
	Payload []byte
}

//RetSimpleTransfer ..
type RetSimpleTransfer struct {
}

//MethodSimpleTransfer ..
type MethodSimpleTransfer struct {
	module.BaseMethod
}

func newMethodSimpleTransfer(m module.Module) *MethodSimpleTransfer {
	return &MethodSimpleTransfer{
		module.BaseMethod{Method: MethodSimpleTransferID, Module: m},
	}
}

//NewParam ..
func (m *MethodSimpleTransfer) NewParam() module.Param {
	return &ParamSimpleTransfer{}
}

//NewRet ..
func (m *MethodSimpleTransfer) NewRet() module.Ret {
	return &RetSimpleTransfer{}
}

//Do ..
func (m *MethodSimpleTransfer) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamSimpleTransfer)

	logs.Info("[%d][%s] MethodSimpleTransfer Do ShardID:[%d]", t.GroupID, t.HashSource, param.ShardID)

	// 不太清楚中间两个string应该是什么，总之先让其能Build通过
	err := m.Module.Push(param.ShardID, "", "", param.Payload)
	if nil != err {
		logs.Warn("MethodSimpleTransfer, Push Message to Shard %d failed, %v", param.ShardID, err)
	}

	ret = nil
	errCode = message.ErrCodeOK
	return
}
