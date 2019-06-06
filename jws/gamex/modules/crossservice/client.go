package crossservice

import (
	"fmt"

	csClient "vcs.taiyouxi.net/jws/crossservice/client"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type client struct {
	res *resources
	cs  *csClient.Client

	isClosed bool
}

func newClient(res *resources) *client {
	c := &client{}
	c.res = res
	c.cs = csClient.NewClient(uint32(game.Cfg.Gid), res.shardList)
	return c
}

func (c *client) start() error {
	c.cs.AddGroupIDs(c.res.groupIDs)
	if 0 != game.Cfg.CrossServiceConnPoolMax {
		c.cs.SetConnPoolMax(game.Cfg.CrossServiceConnPoolMax)
	}
	if err := c.cs.Start(); nil != err {
		return fmt.Errorf("[CrossService] Start Client, Start CrossService Client, Failed, %v", err)
	}

	go c.pullHolding()

	return nil
}

func (c *client) stop() {
	c.isClosed = true
	c.cs.Stop()
}

func (c *client) pullHolding() {
	defer logs.PanicCatcherWithInfo("CrossService Client, pullHolding run")

	logs.Warn("[CrossService] Client, start pullHolding")
	for !c.isClosed {
		req, param, err := c.cs.Pull()
		if nil != err {
			logs.Warn("[CrossService] Client, pullHolding break, err: %v", err)
			break
		}
		if nil == req {
			continue
		}
		callback := getCallBackHandle(req.Module, req.Method)
		if nil == callback {
			logs.Warn("[CrossService] Client, unknown callback %s.%s", req.Module, req.Method)
			continue
		}
		func() {
			defer logs.PanicCatcherWithInfo("CrossService Client, callback err")
			callback(param)
		}()
	}
}
