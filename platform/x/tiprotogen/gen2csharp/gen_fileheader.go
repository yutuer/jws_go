package gen2csharp

import (
	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

func (g *genner2Csharp) genFileHeader(
	buf *util.CodeGenData,
	def *dsl.ProtoDef) {
	buf.WriteLine(headerTemplate,
		def.Name,
		def.Title,
		def.Comment)
}
