package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GVGMatchTest : GVGMatchTest
// GVG匹配测试
var matchTimes int64

// reqMsgGVGMatchTest GVGMatchTest请求消息定义
type reqMsgGVGMatchTest struct {
	Req
	RobotLevel int64 `codec:"robot_lv"` // 机器人等级
}

// rspMsgGVGMatchTest GVGMatchTest回复消息定义
type rspMsgGVGMatchTest struct {
	SyncResp
	IsSuccess int64  `codec:"is_success"` // 是否成功
	EnemyInfo []byte `codec:"enemy_info"`
	NowTime   int64  `codec:"now_time"`
}

// GVGMatchTest GVGMatchTest: GVG匹配测试
func (p *Account) GVGMatchTest(r servers.Request) *servers.Response {
	req := new(reqMsgGVGMatchTest)
	rsp := new(rspMsgGVGMatchTest)

	initReqRsp(
		"Attr/GVGMatchTestRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	droid := gamedata.GetDroidForGVG(uint32(req.RobotLevel))
	a := &helper.Avatar2Client{}
	account.FromAccountByDroid(a, droid, 1)
	logs.Debug("Get Account Equip Info: %d, %v, %v", a.EquipMatEnhMax, a.EquipMatEnh, a.EquipMatEnhLv)

	rsp.EnemyInfo = encode(a)

	matchTimes++
	logs.Debug("NowTime: %d", matchTimes)
	rsp.NowTime = matchTimes
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
