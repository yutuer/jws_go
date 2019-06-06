package gen2golang

/*
package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// {{ProtoName}} : {{ProtoTitle}}
// {{ProtoInfo}}

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

*/

var headerTemplate string = `package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// %s : %s
// %s

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑`

var multiFileHeader string = `package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)`
var multiFileAnnotation string = `// %s : %s
// %s

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑
`
var handlerHeader string = `package logics`

/*
// reqMsgTestProto title: req msg
type reqMsgTestProto struct {
	Req
	AccountID string `codec:"accid"` // AccountID info
}
*/

var reqMsgBegin string = `// %s %s请求消息定义
type %s struct {
	%s`
var reqMsgEnd string = `}`

var reqCheatMsgBegin string = `// %s %s请求消息定义
	type %s struct {
		%s`
var reqCheatMsgEnd string = `}`

/*
// reqMsgTestProto title: rsp msg
type rspMsgTestProto struct {
	SyncRespWithRewards
	RespID int `codec:"respid"` // RespID info
}
*/

var rspMsgBegin string = `// %s %s回复消息定义
type %s struct {
	%s`
var rspMsgEnd string = `}`

var ObjectMsgBegin string = `// %s %s
type %s struct {
	%s`
var ObjectMsgEnd string = `}`

/*
// TestProto title: testProto info
func (p *Account) TestProto(r servers.Request) *servers.Response {
	req := new(reqMsgTestProto)
	rsp := new(rspMsgTestProto)

	initReqRsp(
		"Attr/TestProtoRsp",
		r.RawBytes,
		req, rsp, p)

	const (
		_ = iota
		ErrOne // ErrOne info
		ErrTwo // ErrTwo info
	)

	// logic imp begin
	logs.Error("there is no Imp for testProto")

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
*/

var funcBegin string = `// %s %s: %s
func (p *Account) %s(r servers.Request) *servers.Response {
	req := new(%s)
	rsp := new(%s)

	initReqRsp(
		"%sRsp",
		r.RawBytes,
		req, rsp, p)`
var funcEnd string = `	// logic imp begin
	warnCode := p.%sHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}`

// 	RespID int `codec:"respid"` // RespID info
var paramTemplate string = "	%s %s `codec:\"%s\"` // %s"

var pathRegBegin string = `package logics

import "vcs.taiyouxi.net/platform/planx/servers"

func handleAllGenFunc(r *servers.Mux, p *Account) {`

var pathRegFunc string = `	r.HandleFunc("%s/%sReq", p.%s)`

var pathRegEnd string = `}`

var msgHandler string = `// %s : %s
// %s
func (p *Account) %sHandler(req *%s, resp *%s) uint32 {
	return 0
}
`
