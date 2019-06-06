package chat

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/gate"
)

type Config struct {
	PublicIp string `toml:"publicip"`
	Listen   string `toml:"listen"`

	// 不要再toml里出现，会覆写
	ChatAddr string
}

var Cfg Config

func (c *Config) Init() {
	ip := gate.GetPublicIP(c.PublicIp, c.Listen)
	c.ChatAddr = fmt.Sprintf("ws://%s/ws", ip)
}
