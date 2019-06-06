package server

var implementCode = `
package %s

import (
	"vcs.taiyouxi.net/comic/gamex/logics/protocol"
	"vcs.taiyouxi.net/comic/gamex/account"
)

func %sHandler(p *account.Account, req *protogen.%sReq, resp *protogen.%sResp) protogen.ErrCode {
	return 0
}

`

var registerReqCode = `
package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/comic/gamex/modules/playermsg"
)

func RegisterFunc(r *servers.Mux, p *Account) {
%s
}
`

var registerReqFunc = `	r.HandleFunc("Attr/%sReq", p.%s)
`

var registerPushCode = `func RegisterPushFunc(r *servers.Mux, p *Account) {
%s
}`

var registerPushFunc = `	r.HandleFunc(playermsg.MsgCode%s, p.Push%s)
`

var reqRspHeader = `
package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/comic/gamex/logics/protocol"
`

var reqRspFunc = `
func (p *Account) %s(r servers.Request) *servers.Response {
	req := &protogen.%sReq{}
	resp := &protogen.%sResp{
		Resp: &protogen.Resp{},
	}
	address := "Attr/%sRsp"

	initReqRsp(address, r.RawBytes, req, p)
	if warnCode := %s.%sHandler(p.Account, req, resp); warnCode != 0 {
		return rspWarn(address, req.Req.GetPassthroughID(), resp, resp.Resp, p, warnCode)
	}
	makeChange(p, resp.Resp)
	return rspSuccess(address, req.Req.GetPassthroughID(), resp, resp.Resp, p)
}
`

var pushCodeBuild string = `package push

import (
	"vcs.taiyouxi.net/comic/gamex/account"
	"vcs.taiyouxi.net/comic/gamex/logics/protocol"
	"time"
)

func Build%sPush(p *account.Account, sample *protogen.%sPushReq) *protogen.%sPush {
	// TODO
	return nil
}
`

var pushCodeHeader string = `package logics

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/comic/gamex/logics/notify"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/comic/gamex/logics/handlers/push"
	"vcs.taiyouxi.net/comic/gamex/logics/protocol"
)
`

var pushCodeFunc string = `
func (p *Account) Push%s(r servers.Request) *servers.Response {
	req := &protogen.%sPushReq{}
	if err := proto.Unmarshal(r.RawBytes, req); nil != err {
		logs.Error("%s Unmarshal Request Error:", err)
		return nil
	}
	notifyMsg := notify.NewMsgNotify()
	notifyMsg.SetAddr("Push/%sPush")
	notifyMsg.SetFuncMakeRsp(func(addr string) *servers.Response {
		pushMsg := push.Build%sPush(p.Account, req)
		bs, err := proto.Marshal(pushMsg)
		if nil != err {
			logs.Error("%s Marshal Response Error", err)
			return nil
		}
		rsp := &servers.Response{
			Code:     addr,
			RawBytes: bs,
		}
		return rsp
	})
	p.SendRespByPush(notifyMsg)
	return nil
}
`
var playerMsgCode string = `
package playermsg

//MsgCode Define
const (
%s
)
`

var playerMsgOne string = `	MsgCode%s         = "MSG/%s"
`
