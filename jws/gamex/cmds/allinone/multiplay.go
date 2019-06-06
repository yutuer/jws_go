package allinone

import (
	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/cmds/allinone"
	"vcs.taiyouxi.net/platform/planx/util"
)

func GamexStartMultiplay(c *cli.Context, wg util.WaitGroupWrapper) {
	multiplayConfig := c.String("multiplayer")
	if multiplayConfig == "" {
		return
	}

	allinone.StartMultiplay(multiplayConfig, wg)
}
