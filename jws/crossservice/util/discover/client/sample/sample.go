package main

import (
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"

	"vcs.taiyouxi.net/platform/planx/util/uuid"

	"vcs.taiyouxi.net/jws/crossservice/util/discover"
	"vcs.taiyouxi.net/jws/crossservice/util/discover/client"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func main() {
	app := cli.NewApp()

	app.Name = "client"
	app.Usage = "a sample client for service discover"

	app.Action = action
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "i",
			Usage: "ip of discover service(etcd)",
			Value: "127.0.0.1",
		},
		cli.IntFlag{
			Name:  "p",
			Usage: "port of discover service(etcd)",
			Value: 2379,
		},
		cli.StringFlag{
			Name:  "project",
			Usage: "project name to watch",
			Value: "ProjectName",
		},
		cli.StringFlag{
			Name:  "service",
			Usage: "service name to watch",
			Value: "ServiceName",
		},
		cli.StringFlag{
			Name:  "ver",
			Usage: "version to watch, \"0.0.0\"",
			Value: "0.0.0",
		},
		cli.IntFlag{
			Name:  "hold",
			Usage: "hold time (s)",
			Value: 3,
		},
		cli.IntFlag{
			Name:  "update",
			Usage: "times to do update",
			Value: 0,
		},
	}

	app.Run(os.Args)
}

func action(c *cli.Context) {
	paramHost := c.GlobalString("i")
	paramPort := c.GlobalInt("p")
	paramProject := c.GlobalString("project")
	paramService := c.GlobalString("service")
	paramVersion := c.GlobalString("ver")
	paramHold := c.GlobalInt("hold")
	paramUpdate := c.GlobalInt("update")

	discover.InitEtcdServerCfg(discover.Config{
		Endpoints: []string{fmt.Sprintf("http://%s:%d", paramHost, paramPort)},
		Root:      "a4k/Discover/registry",
	})

	handle := client.NewClient()
	handle.SetProject(paramProject)
	handle.SetService(paramService)
	handle.SetVersion(paramVersion)
	handle.SetIP("www.taiyouxi.cn")
	handle.SetPort(9527)
	handle.SetIndex(uuid.NewV4().String())

	if err := handle.Reg(); nil != err {
		logs.Debug("Reg Failed, %v", err)
		return
	}

	<-time.After(time.Duration(paramHold) * time.Second)

	if 0 != paramUpdate {
		for i := 0; i < paramUpdate; i++ {
			handle.SetIP("www.taiyouxi.cn")
			handle.SetPort(9527 + i)
			if err := handle.Reg(); nil != err {
				logs.Debug("Reg Failed, %v", err)
				return
			}
		}
		<-time.After(time.Duration(paramHold) * time.Second)
	}

	if err := handle.UnReg(); nil != err {
		logs.Debug("UnReg Failed, %v", err)
		return
	}
	<-time.After(time.Second)
}
