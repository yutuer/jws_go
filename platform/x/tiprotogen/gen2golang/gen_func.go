package gen2golang

import (
	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

func (g *genner2golang) genFunc(
	buf *util.CodeGenData,
	def *dsl.ProtoDef) {
	buf.WriteLine(funcBegin,
		def.Name,
		def.Title,
		def.Comment,
		def.Name,
		def.GetReqMsgName(),
		def.GetRspMsgName(),
		def.GetPath())
	buf.WriteLine("")
	buf.WriteLine(funcEnd,
		def.Name)
}
