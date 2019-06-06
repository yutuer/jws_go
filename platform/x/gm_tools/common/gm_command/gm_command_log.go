package gm_command

import (
	"encoding/json"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type CommandRecord struct {
	AccountId  string   `json:"account"`
	ServerName string   `json:"server"`
	Command    string   `json:"command"`
	Params     []string `json:"params"`
	GM         string   `json:"gm"`
}

func NewCommand(gm, command string, server, accountid string, params []string) CommandRecord {
	return CommandRecord{
		AccountId:  accountid,
		ServerName: server,
		Command:    command,
		Params:     params,
		GM:         gm,
	}
}

func LogCommand(gm, command string, server, accountid string, params []string) {
	c := NewCommand(gm, command, server, accountid, params)
	j, err := json.Marshal(c)
	if err != nil {
		logs.Error("[LogCommand] Marshal Err by %s", err.Error())
		return
	}

	logs.Info("[LogCommand]:%s", string(j))
}
