package gen2csharp

import (
	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

func (g *genner2Csharp) genReqMsg(
	buf *util.CodeGenData,
	def *dsl.ProtoDef) {

	buf.WriteLine(reqMsgBegin,
		def.GetReqMsgName(),
		def.Title,
		def.GetReqMsgName(),
		g.GetReqBase(def))

	g.genReqParams(buf, def.Req.Params)

	buf.WriteLine(reqMsgEnd,
		def.GetReqMsgName())
}

func (g *genner2Csharp) genRspMsg(
	buf *util.CodeGenData,
	def *dsl.ProtoDef) {
	//
	buf.WriteLine(rspMsgBegin,
		def.GetRspMsgName(),
		def.Title,
		def.GetRspMsgName(),
		g.GetRspBase(def))

	g.genRspParams(buf, def.Rsp.Params)

	buf.WriteLine(msgDataDefBegin)
	g.genParamsDataDef(buf, def.Rsp.Params)
	buf.WriteLine(msgDataDefEnd)

	buf.WriteLine(msgDataGetBegin, "msgData", "msgData")
	g.genParamsDataGet(buf, def.Rsp.Params)
	buf.WriteLine(msgDataGetEnd)

	buf.WriteLine(rspMsgEnd,
		def.GetRspMsgName())
}

func (g *genner2Csharp) genNetObject(buf *util.CodeGenData, def *dsl.ProtoDef) {
	if len(def.Object) <= 0 {
		return
	}
	for i, _ := range def.Object {
		g.genNetObjectOfNet(buf, def, i)
		g.genNetObjectOfParse(buf, def, i)
	}
}

func (g *genner2Csharp) genNetObjectOfNet(buf *util.CodeGenData, def *dsl.ProtoDef, index int) {
	buf.WriteLine(objectMsgBegin,
		def.GetNetObjectMsgNameOfNet(index),
		def.Title,
		def.GetNetObjectMsgNameOfNet(index)+": AbstractRspMsg",
		"")

	g.genRspParams(buf, def.Object[index].Params)

	buf.WriteLine(msgDataGetBegin, def.GetNetObjectMsgNameOfParse(index), def.GetNetObjectMsgNameOfParse(index))
	g.genParamsDataGet(buf, def.Object[index].Params)
	buf.WriteLine(msgDataGetEnd)

	buf.WriteLine(objectMsgEnd)

}

func (g *genner2Csharp) genNetObjectOfParse(buf *util.CodeGenData, def *dsl.ProtoDef, index int) {
	buf.WriteLine(objectMsgBegin,
		def.GetNetObjectMsgNameOfParse(index),
		def.Title,
		def.GetNetObjectMsgNameOfParse(index),
		"")

	g.genParamsDataDefNoTab(buf, def.Object[index].Params)

	buf.WriteLine(objectMsgEnd)
}
