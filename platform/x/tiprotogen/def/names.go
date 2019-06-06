package dsl

import "strings"

func (d *ProtoDef) GetReqMsgName() string {
	return "reqMsg" + d.Name
}

func (d *ProtoDef) GetRspMsgName() string {
	return "rspMsg" + d.Name
}

func (d *ProtoDef) GetPath() string {
	return d.Path + "/" + d.Name
}

func (d *ProtoDef) GetFileName() string {
	return strings.ToLower(d.Name)
}

func (d *ProtoDef) GetNetObjectMsgNameOfParse(index int) string {
	return d.Object[index].Name
}

func (d *ProtoDef) GetNetObjectMsgNameOfNet(index int) string {
	return "Net" + d.Object[index].Name
}
