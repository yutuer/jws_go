package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/codegangsta/cli"

	"vcs.taiyouxi.net/jws/crossservice/util/discover"
	"vcs.taiyouxi.net/jws/crossservice/util/discover/watcher"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func main() {
	app := cli.NewApp()

	app.Name = "watcher"
	app.Usage = "a sample watcher for service discover"

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
			Value: "",
		},
		cli.StringFlag{
			Name:  "ver",
			Usage: "version to watch, \"0.0.0\"",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "vermin",
			Usage: "min version to watch",
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
	paramVersionMin := c.GlobalBool("vermin")

	filters := []watcher.Filter{}
	if "" != paramService {
		filters = append(filters, watcher.FilterService{Service: paramService})
	}
	if "" != paramVersion {
		if paramVersionMin {
			filters = append(filters, watcher.FilterVersionMin{VersionMin: paramVersion})
		} else {
			filters = append(filters, watcher.FilterVersion{Version: paramVersion})
		}
	}

	discover.InitEtcdServerCfg(discover.Config{
		Endpoints: []string{fmt.Sprintf("http://%s:%d", paramHost, paramPort)},
		Root:      "/a4k/Discover/registry",
	})

	handle := watcher.NewWatcher()
	handle.SetProject(paramProject)
	handle.SetFilter(filters)

	handle.OnServiceAdd = handleAdd
	handle.OnServiceUpdate = handleUpdate
	handle.OnServiceDel = handleDel

	handle.Start()

	signalChan := make(chan os.Signal, 10)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	handle.Close()
	<-time.After(time.Second)
}

func handleAdd(s discover.Service) {
	logs.Debug("Add Service %+v", s)
}

func handleUpdate(s discover.Service) {
	logs.Debug("Update Service %+v", s)
}

func handleDel(s discover.Service) {
	logs.Debug("Del Service %+v", s)
}
