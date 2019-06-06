package builderjson

import "vcs.taiyouxi.net/platform/x/tiprotogen/def"

type protoReqJson struct {
	Base   string     `json:"base"`
	Params [][]string `json:"params"`
}

func (p *protoReqJson) ToReqDef() dsl.ProtoReq {
	return dsl.ProtoReq{
		Base:   p.Base,
		Params: toParamDef(p.Params),
	}
}
