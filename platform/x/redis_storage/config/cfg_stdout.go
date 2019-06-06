package config

import "vcs.taiyouxi.net/platform/planx/util/storehelper"

type StdoutOnlandConfig struct {
}

func (c *StdoutOnlandConfig) HasConfigured() bool {
	return true
}

func init() {
	Register("Stdout", &StdoutOnlandConfig{})
}

func (c *StdoutOnlandConfig) Setup(addFn func(storehelper.IStore)) {
	if !c.HasConfigured() {
		return
	}
	addFn(storehelper.NewStoreStdout())
}
