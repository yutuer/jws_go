package builderjson

import "vcs.taiyouxi.net/platform/x/tiprotogen/def"

type protoObjectJson struct {
	Name   string     `json:"name"`
	Params [][]string `json:"params"`
}

func (p *protoObjectJson) ToObjectDef() dsl.ProtoObject {
	return dsl.ProtoObject{
		Name:   p.Name,
		Params: toParamDef(p.Params),
	}
}

type ProtoObjectArray []protoObjectJson

func (p ProtoObjectArray) ToObjectDefArray() []dsl.ProtoObject {
	array := make([]dsl.ProtoObject, 0)
	for _, pObject := range p {
		array = append(array, pObject.ToObjectDef())
	}
	return array
}
