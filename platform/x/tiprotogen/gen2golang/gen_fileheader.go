package gen2golang

import (
	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

func (g *genner2golang) genFileHeader(
	buf *util.CodeGenData,
	def *dsl.ProtoDef) {
	buf.WriteLine(headerTemplate,
		def.Name,
		def.Title,
		def.Comment)
}
