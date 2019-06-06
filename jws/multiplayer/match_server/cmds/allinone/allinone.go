package allinone

import (
	"net/http"
	"os"

	"github.com/codegangsta/cli"

	_ "github.com/rakyll/gom/http"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/multiplayer/match_server/cmds"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/match_server/config"
	"vcs.taiyouxi.net/jws/multiplayer/match_server/match"
	"vcs.taiyouxi.net/jws/multiplayer/match_server/notify"
	matchServer "vcs.taiyouxi.net/jws/multiplayer/match_server/server"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
)

func init() {
	logs.Trace("allinone cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "allinone",
		Usage:  "开启所有功能",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "config.toml",
				Usage: "Onland Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "pprofport, pp",
				Value: "",
				Usage: "pprof port",
			},
		},
	})
}

func Start(c *cli.Context) {
	util.PProfStart()
	logs.Info("start server websocket mode")
	var waitGroup util.WaitGroupWrapper

	pp := c.String("pprofport")
	if pp != "" {
		go func() {
			http.ListenAndServe("localhost:"+pp, nil)
		}()
	}

	cfgName := c.String("config")
	var common_cfg struct{ CommonCfg multConfig.CommonConfig }
	cfgApp := config.NewConfigToml(cfgName, &common_cfg)

	if cfgApp == nil {
		logs.Critical("CommonConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	multConfig.Cfg = common_cfg.CommonCfg
	etcd.InitClient(multConfig.Cfg.EtcdEndpoint)

	g := gin.Default()
	matchServer.MatchServer(g)
	go g.Run(multConfig.Cfg.Url)

	//match.GetGVEMatch().Start(helper.MatchDefaultToken)

	//metrics
	signalhandler.SignalKillFunc(func() { metrics.Stop() })

	waitGroup.Wrap(func() {})
	waitGroup.Wrap(func() { metrics.Start("metrics.toml") })
	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()

	//match.GetGVEMatch().Stop()
	match.GVEMatchV2_Stop()
	notify.StopNotifyies()

	logs.Close()
}
