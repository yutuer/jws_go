package builderjson

import (
	"fmt"

	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
)

type protoJson struct {
	Name    string           `json:"name"`
	Comment string           `json:"comment"`
	Path    string           `json:"path"`
	Title   string           `json:"title"`
	Cheat   bool             `json:"cheat"`
	Req     protoReqJson     `json:"req"`
	Rsp     protoRspJson     `json:"rsp"`
	Objects ProtoObjectArray `json:"objects"`
	//Objects []protoObjectJson `json:"objects"`
}

func (p *protoJson) ToDef() dsl.ProtoDef {
	return dsl.ProtoDef{
		Name:    p.Name,
		Comment: p.Comment,
		Path:    p.Path,
		Title:   p.Title,
		Cheat:   p.Cheat,
		Req:     p.Req.ToReqDef(),
		Rsp:     p.Rsp.ToRspDef(),
		Object:  p.Objects.ToObjectDefArray(),
	}
}

type protoJsonArray []protoJson

func (p protoJsonArray) ToDef() []dsl.ProtoDef {
	array := make([]dsl.ProtoDef, 0)
	fmt.Println(len(p))
	for _, pJson := range p {
		array = append(array, pJson.ToDef())
	}
	fmt.Println(len(array))
	return array
}
