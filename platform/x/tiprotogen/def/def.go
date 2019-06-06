package dsl

type ProtoDefs struct {
	Def []ProtoDef
}

type ProtoDef struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Path    string `json:"path"`
	Title   string `json:"title"`
	Cheat   bool
	Req     ProtoReq      `json:"req"`
	Rsp     ProtoRsp      `json:"rsp"`
	Object  []ProtoObject `json:"object"`
	//Objects []ProtoObject `json:"objects"`
}

type ProtoReq struct {
	Base   string       `json:"base"`
	Params []ProtoParam `json:"params"`
}

type ProtoRsp struct {
	Base   string       `json:"base"`
	Params []ProtoParam `json:"params"`
}

type ProtoParam struct {
	Name      string `json:"name"`
	Comment   string `json:"comment"`
	Type      string `json:"type"`
	ShortName string `json:"shortName"`
	Transfer  bool   `json:"-"`
}

type ProtoObject struct {
	Name   string       `json:"Name"`
	Params []ProtoParam `json:"params"`
}
