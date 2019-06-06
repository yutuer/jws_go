package gen2golang

import (
	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/log"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

type genner2golang struct {
}

func New() *genner2golang {
	return new(genner2golang)
}

func (g *genner2golang) Gen(def *dsl.ProtoDef) []byte {
	log.Trace("gen 2 golang %v", def)

	buf := util.NewCodeGenData()

	g.genFileHeader(buf, def)
	buf.WriteLine("")
	g.genReqMsg(buf, def)
	buf.WriteLine("")
	g.genRspMsg(buf, def)
	buf.WriteLine("")
	g.genFunc(buf, def)
	buf.WriteLine("")
	g.genNetObject(buf, def)
	buf.WriteLine("")
	return buf.Bytes()
}

func (g *genner2golang) GenMultiFileHeader() []byte {

	buf := util.NewCodeGenData()
	buf.WriteLine(multiFileHeader)
	buf.WriteLine("")

	return buf.Bytes()
}

func (g *genner2golang) GenHandlerFileHeader() []byte {

	buf := util.NewCodeGenData()
	buf.WriteLine(handlerHeader)
	buf.WriteLine("")

	return buf.Bytes()
}

func (g *genner2golang) GenMultiFileBody(def *dsl.ProtoDef) []byte {
	log.Trace("gen 2 golang %v", def)

	buf := util.NewCodeGenData()
	buf.WriteLine(multiFileAnnotation,
		def.Name,
		def.Title,
		def.Comment)
	buf.WriteLine("")
	g.genReqMsg(buf, def)
	buf.WriteLine("")
	g.genRspMsg(buf, def)
	buf.WriteLine("")
	g.genFunc(buf, def)
	buf.WriteLine("")
	g.genNetObject(buf, def)
	buf.WriteLine("")
	return buf.Bytes()
}

func GenPathFile(defs []*dsl.ProtoDef) []byte {
	buf := util.NewCodeGenData()
	buf.WriteLine(pathRegBegin)
	for _, def := range defs {
		buf.WriteLine(pathRegFunc,
			def.Path,
			def.Name,
			def.Name)
	}
	buf.WriteLine(pathRegEnd)

	return buf.Bytes()
}

func AppendRegPathFile(defs []*dsl.ProtoDef) []byte {
	if len(defs) == 0 {
		return nil
	}
	buf := util.NewCodeGenData()
	buf.WriteLine("        //%s", defs[0].Comment)
	for _, def := range defs {
		buf.WriteLine(pathRegFunc,
			def.Path,
			def.Name,
			def.Name)
	}
	return buf.Bytes()
}

func (g *genner2golang) GenHandler(def *dsl.ProtoDef) []byte {
	log.Trace("gen 2 golang %v", def)

	buf := util.NewCodeGenData()
	buf.WriteLine(msgHandler,
		def.Name,
		def.Title,
		def.Comment,
		def.Name,
		def.GetReqMsgName(),
		def.GetRspMsgName())
	return buf.Bytes()
}
