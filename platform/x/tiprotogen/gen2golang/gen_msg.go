package gen2golang

import (
	"fmt"

	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

func (g *genner2golang) genReqMsg(
	buf *util.CodeGenData,
	def *dsl.ProtoDef) {

	//
	buf.WriteLine(reqMsgBegin,
		def.GetReqMsgName(),
		def.Title,
		def.GetReqMsgName(),
		g.GetReqBase(def))
	for i := 0; i < len(def.Req.Params); i++ {
		def.Req.Params[i].Transfer = true
	}
	g.genParams(buf, def.Req.Params)

	buf.WriteLine(reqMsgEnd)
}

func (g *genner2golang) genRspMsg(
	buf *util.CodeGenData,
	def *dsl.ProtoDef) {
	//
	buf.WriteLine(rspMsgBegin,
		def.GetRspMsgName(),
		def.Title,
		def.GetRspMsgName(),
		g.GetRspBase(def))
	for i := 0; i < len(def.Rsp.Params); i++ {
		def.Rsp.Params[i].Transfer = true
	}
	g.genParams(buf, def.Rsp.Params)

	buf.WriteLine(rspMsgEnd)
}

func (g *genner2golang) genNetObject(buf *util.CodeGenData, def *dsl.ProtoDef) {
	if len(def.Object) <= 0 {
		return
	}
	fmt.Println("test:", def)
	for i, _ := range def.Object {
		buf.WriteLine(ObjectMsgBegin,
			def.GetNetObjectMsgNameOfParse(i),
			def.Title,
			def.GetNetObjectMsgNameOfParse(i),
			"")

		g.genParams(buf, def.Object[i].Params)
		buf.WriteLine(ObjectMsgEnd)
	}
}
