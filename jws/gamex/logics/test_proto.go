package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// TestProto : 测试生成协议
// 用来测试生成代码的协议

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgTestProto 测试生成协议请求消息定义
type reqMsgTestProto struct {
	Req
	AccountID string   `codec:"acid"`   // 膜拜的玩家ID
	ReqID2s   []string `codec:"reqid2"` // 请求2ID
}

// rspMsgTestProto 测试生成协议回复消息定义
type rspMsgTestProto struct {
	SyncRespWithRewards
	ResAccountID string  `codec:"resaccountid"` // 膜拜的玩家ID
	ReqID2s      []int64 `codec:"reqid2"`       // 请求2ID
}

// TestProto 测试生成协议: 用来测试生成代码的协议
func (p *Account) TestProto(r servers.Request) *servers.Response {
	req := new(reqMsgTestProto)
	rsp := new(rspMsgTestProto)

	initReqRsp(
		"Attr/TestProtoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	logs.Error("there is no Imp for TestProto")

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
