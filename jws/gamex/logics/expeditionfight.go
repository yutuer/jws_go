package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ExpeditionFight : 远征对战协议
// 用来传输所有的敌人对战信息

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgExpeditionFight 远征对战协议请求消息定义
type reqMsgExpeditionFight struct {
	Req
	EnmyID int64 `codec:"enmyid"` // 第几个敌人
}

// rspMsgExpeditionFight 远征对战协议回复消息定义
type rspMsgExpeditionFight struct {
	SyncResp
	Enmy [][]byte `codec:"enmy"`
}

//一共9个宝箱,传过来的宝箱范围应在1~9之间
const maxNum = 9
const minNum = 1

// ExpeditionFight 远征对战协议: 用来传输所有的敌人对战信息
func (p *Account) ExpeditionFight(r servers.Request) *servers.Response {
	req := new(reqMsgExpeditionFight)
	rsp := new(rspMsgExpeditionFight)

	initReqRsp(
		"Attr/ExpeditionFightRsp",
		r.RawBytes,
		req, rsp, p)

	if req.EnmyID < minNum || req.EnmyID > maxNum {
		logs.Error("ExpeditionFight param err %d", req.EnmyID)
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	for _, info := range p.Profile.GetExpeditionInfo().ExpeditionEnmyDetail[req.EnmyID-1].Enemies {
		if info.Acid == "" {
			continue
		}
		rsp.Enmy = append(rsp.Enmy, encode(info))

	}
	rsp.OnChangerExpeditionInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
