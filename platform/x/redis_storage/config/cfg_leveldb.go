package config

import "vcs.taiyouxi.net/platform/planx/util/storehelper"

type LevelDBOnlandConfig struct {
	DBPath string `mapstructure:"DBPath"`
}

func (c *LevelDBOnlandConfig) HasConfigured() bool {
	if c.DBPath == "" ||
		len(c.DBPath) == 0 {
		return false
	}
	return true
}

func init() {
	Register("LevelDB", &LevelDBOnlandConfig{})
}

func (c *LevelDBOnlandConfig) Setup(addFn func(storehelper.IStore)) {
	if !c.HasConfigured() {
		return
	}
	addFn(storehelper.NewStoreLevelDB(c.DBPath))
}
