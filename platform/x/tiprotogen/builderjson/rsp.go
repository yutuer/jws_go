package builderjson

import "vcs.taiyouxi.net/platform/x/tiprotogen/def"

type protoRspJson struct {
	Base   string     `json:"base"`
	Params [][]string `json:"params"`
}

func (p *protoRspJson) ToRspDef() dsl.ProtoRsp {
	return dsl.ProtoRsp{
		Base:   p.Base,
		Params: toParamDef(p.Params),
	}
}
