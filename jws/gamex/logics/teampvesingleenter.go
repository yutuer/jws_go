package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// TeamPVESingleEnter : 进入烽火燎原单人模式
// 进入单人的烽火燎原模式是发送的请求,服务器会返回各个小关卡的奖励

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgTeamPVESingleEnter 进入烽火燎原单人模式请求消息定义
type reqMsgTeamPVESingleEnter struct {
	Req
	LevelId int64 `codec:"p1_"` // 进入的战役难度
}

// rspMsgTeamPVESingleEnter 进入烽火燎原单人模式回复消息定义
type rspMsgTeamPVESingleEnter struct {
	SyncResp
	LittleLvs []string `codec:"p1_"` // 各个小关卡的ID
}

// TeamPVESingleEnter 进入烽火燎原单人模式: 进入单人的烽火燎原模式是发送的请求,服务器会返回各个小关卡的奖励
func (p *Account) TeamPVESingleEnter(r servers.Request) *servers.Response {
	req := new(reqMsgTeamPVESingleEnter)
	rsp := new(rspMsgTeamPVESingleEnter)

	initReqRsp(
		"Attr/TeamPVESingleEnterRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	logs.Error("there is no Imp for TeamPVESingleEnter")

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
