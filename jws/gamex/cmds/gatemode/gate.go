package gatemode

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"os"

	"vcs.taiyouxi.net/jws/gamex/cmds"
	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers/gate"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
)

const ()

func init() {
	logs.Trace("gate cmd loaded")
	cmds.Register(&cli.Command{
		Name: "gate",
		//ShortName: "",
		Usage:  "启动Gate模式",
		Action: GateStart,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "gate.toml",
				Usage: "Gate Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "record, r",
				Value: "packetlog.xml",
				Usage: "Gate recording mode config, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

func GateStart(c *cli.Context) {
	fmt.Println("PlanX Gate Server, version:", gate.GetVersion())
	//config loading

	var recordCfgName string = c.String("record")
	config.NewConfig(recordCfgName, true, client.LoadPacketLogger)
	defer client.StopPacketLogger()

	var cfgName string = c.String("config")

	//For Sentry
	var sentryConfig logs.SentryCfg
	config.NewConfigToml(cfgName, &sentryConfig)
	logs.InitSentryTags(sentryConfig.SentryConfig.DSN,
		map[string]string{
			"service": "gamex",
			"subcmd":  "gatemode",
		},
	)

	tcpgame := NewTCPGameServerManager()

	// gate cfg need to register to etcd
	Gate := gate.NewGateServer(cfgName)
	if !Gate.SyncCfg2Etcd() {
		logs.Error("gate SyncCfg2Etcd fail")
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	var waitGroup util.WaitGroupWrapper

	//metrics
	signalhandler.SignalKillFunc(func() { metrics.Stop() })
	waitGroup.Wrap(func() {
		metrics.Start("metrics.toml")
	})

	//tcp gate server start up
	signalhandler.SignalKillHandler(Gate)
	waitGroup.Wrap(func() { Gate.Start(tcpgame) })

	//handle kill signal
	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	time.Sleep(1 * time.Second)
	waitGroup.Wait()

}
