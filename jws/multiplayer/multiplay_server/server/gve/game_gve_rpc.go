package gve

import (
	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/common"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/gve_proto"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *GVEGame) PlayerLoss(acID string) {
	msg := GVEGameCommandMsg{}
	msg.req_lossPlayer = new(LossPlayerReq)
	msg.req_lossPlayer.AcID = acID
	r.PushCommand(&msg)
	return
}

func (r *GVEGame) SendReqToGame(reqPacket msgprocessor.IPacket) []byte {
	if r == nil {
		logs.Warn("GVEGame.SendReqToGame r == nil %v", reqPacket)
		return gve_proto.GenErrorRspPacket(reqPacket, common.MsgResCodeReqPacketErr)
	}

	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		logs.Error("GVEGame.SendReqToGame reqPacket.Data fail %v", reqPacket)
		return gve_proto.GenErrorRspPacket(reqPacket, common.MsgResCodeReqPacketErr)
	}

	resChan := make(chan GVEGameCommandResMsg, 1)
	msg := GVEGameCommandMsg{}
	msg.number = reqPacket.Number()
	msg.ResChann = resChan

	switch reqPacket.DataType() {
	case DatasEnterMultiplayGameReq:
		req := new(EnterMultiplayGameReq)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_enterMG = req
		rsp := r.PushCommandWithRsp(&msg)
		return gve_proto.GenEnterMultiplayGameRsp(reqPacket, &rsp.Stat, &rsp.Datas)

	case DatasGetGameDatasReq:
		req := new(GetGameDatasReq)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_getData = req
		rsp := r.PushCommandWithRsp(&msg)
		return gve_proto.GenGetGameDataRsp(reqPacket, &rsp.Stat, &rsp.Datas)

	case DatasGetGameStateReq:
		req := new(GetGameStateReq)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_getStat = req
		rsp := r.PushCommandWithRsp(&msg)
		return gve_proto.GenGetGameStatRsp(reqPacket, &rsp.Stat, &rsp.Datas)
	case DatasGetGameRewardsReq:
		req := new(GetGameRewardsReq)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_getReward = req
		rsp := r.PushCommandWithRsp(&msg)
		return gve_proto.GenGetGameRwardRsp(reqPacket, rsp.Idx, &rsp.Datas)
	}

	return nil

}

//1. [RPC]进入同步战斗服务器
func (r *GVEGame) enterRoom(number int64, req *EnterMultiplayGameReq) int {
	r.SetNeedPushGameState()
	return 0
}

//
////7. [RPC]获取当前战斗状态
//func (r *GVEGame) getGameState(number int64, req *GetGameStateReq) int{
//	return 0
//}
//
////8. [RPC]获取战斗数据
//func (r *GVEGame) getGameDatas(number int64, req *GetGameDatasReq) int {
//	return 0
//}
