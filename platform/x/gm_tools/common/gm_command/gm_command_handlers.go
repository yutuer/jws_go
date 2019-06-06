package gm_command

import (
	"errors"
)

type gmCommandHander func(c *Context, server, accountid string, params []string) error

var _commands map[string]gmCommandHander

func AddGmCommandHandle(command string, h gmCommandHander) {
	if _commands == nil {
		_commands = make(map[string]gmCommandHander, 128)
	}
	_commands[command] = h
}

func OnCommand(command string, server, accountid string, params []string) (string, error) {
	h, ok := _commands[command]
	if !ok || h == nil {
		return "", errors.New("ERR_NO_HANDLE")
	}

	LogCommand("admin", command, server, accountid, params)

	c := NewContext()

	err := h(c, server, accountid, params)
	data := c.GetData()

	return data, err
}

func OnCommandByRecord(command *CommandRecord) (string, error) {
	return OnCommand(
		command.Command,
		command.ServerName,
		command.AccountId,
		command.Params,
	)
}
