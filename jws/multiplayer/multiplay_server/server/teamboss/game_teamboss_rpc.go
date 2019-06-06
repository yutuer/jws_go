package teamboss

import (
	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/teamboss_proto"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *TBGame) PlayerLoss(acID string) {
	msg := TBGameCommandMsg{}
	msg.req_lossPlayer = new(LossPlayerReq)
	msg.req_lossPlayer.AcID = acID
	r.PushCommand(&msg)
	return
}

func (r *TBGame) SendReqToGame(reqPacket msgprocessor.IPacket) []byte {
	unionTable := new(flatbuffers.Table)
	if !reqPacket.Data(unionTable) {
		logs.Error("GVEGame.SendReqToGame reqPacket.Data fail %v", reqPacket)
		return teamboss_proto.GenErrorRspPacket(reqPacket, MsgResCodeReqPacketErr)
	}

	resChan := make(chan TBGameCommandResMsg, 1)
	msg := TBGameCommandMsg{}
	msg.number = reqPacket.Number()
	msg.ResChann = resChan

	switch reqPacket.DataType() {
	case multiplayMsg.DatasEnterMultiplayGameReq:
		req := new(multiplayMsg.EnterMultiplayGameReq)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_enterMG = req
		rsp := r.PushCommandWithRsp(&msg)
		if rsp == nil {
			return nil
		}
		return teamboss_proto.GenEnterMultiplayGameRsp(reqPacket, &rsp.Stat, &rsp.Datas)

	case multiplayMsg.DatasGetGameDatasReq:
		req := new(multiplayMsg.GetGameDatasReq)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_getData = req
		rsp := r.PushCommandWithRsp(&msg)
		if rsp == nil {
			return nil
		}
		return teamboss_proto.GenGetGameDataRsp(reqPacket, &rsp.Stat, &rsp.Datas)

	case multiplayMsg.DatasGetGameStateReq:
		req := new(multiplayMsg.GetGameStateReq)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_getStat = req
		if r != nil {
			rsp := r.PushCommandWithRsp(&msg)
			if rsp == nil {
				return nil
			}
			return teamboss_proto.GenGetGameStatRsp(reqPacket, &rsp.Stat, &rsp.Datas)
		} else {
			return teamboss_proto.GenGetGameStatRsp(reqPacket, &teamboss_proto.TBGameState{}, &teamboss_proto.TBGameDatas{})
		}
	case multiplayMsg.DatasGetGameRewardsReq:
		req := new(multiplayMsg.GetGameRewardsReq)
		req.Init(unionTable.Bytes, unionTable.Pos)
		msg.req_getReward = req
		rsp := r.PushCommandWithRsp(&msg)
		if rsp == nil {
			return nil
		}
		return teamboss_proto.GenGetGameRwardRsp(reqPacket, rsp.Idx, &rsp.Datas)

	}

	return nil

}

//1. [RPC]进入同步战斗服务器
func (r *TBGame) enterRoom(number int64, req *multiplayMsg.EnterMultiplayGameReq) int {
	r.SetNeedPushGameState()
	return 0
}
