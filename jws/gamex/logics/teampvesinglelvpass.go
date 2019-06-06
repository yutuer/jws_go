package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// TeamPVESingleLvPass : 烽火燎原单人模式小关卡结算
// 烽火燎原单人模式小关卡结算,服务器会返回奖励,如果是最后一个小关卡会返回最终奖励

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgTeamPVESingleLvPass 烽火燎原单人模式小关卡结算请求消息定义
type reqMsgTeamPVESingleLvPass struct {
	Req
	LittleLvId int64 `codec:"p1_"` // 小关卡的id(0-7)
	IsSuccess  int64 `codec:"p2_"` // 是否成功,0为没有,1为全部通过
}

// rspMsgTeamPVESingleLvPass 烽火燎原单人模式小关卡结算回复消息定义
type rspMsgTeamPVESingleLvPass struct {
	SyncRespWithRewards
	IsAllPass int64 `codec:"p1_"` // 是否全部通过,0为没有,1为全部通过
}

// TeamPVESingleLvPass 烽火燎原单人模式小关卡结算: 烽火燎原单人模式小关卡结算,服务器会返回奖励,如果是最后一个小关卡会返回最终奖励
func (p *Account) TeamPVESingleLvPass(r servers.Request) *servers.Response {
	req := new(reqMsgTeamPVESingleLvPass)
	rsp := new(rspMsgTeamPVESingleLvPass)

	initReqRsp(
		"Attr/TeamPVESingleLvPassRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	logs.Error("there is no Imp for TeamPVESingleLvPass")

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
