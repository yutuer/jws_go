package allinone

import (
	"os"

	"fmt"

	"net"

	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	ucfg "vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
	"vcs.taiyouxi.net/platform/x/auth/limit"
	"vcs.taiyouxi.net/platform/x/system_notice_server/cmds"
	noticeCfg "vcs.taiyouxi.net/platform/x/system_notice_server/config"
	"vcs.taiyouxi.net/platform/x/system_notice_server/logic"
	routers2 "vcs.taiyouxi.net/platform/x/system_notice_server/routers"
)

func init() {
	logs.Debug("allinone cmd loaded")
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
				Name:  "iplimit, ip",
				Value: "iplimit.toml",
				Usage: "IpLimit Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "superuid, uid",
				Value: "superuid.toml",
				Usage: "SuperUid Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "logiclog, ll",
				Value: "logiclog.xml",
				Usage: "log player logic logs, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

var CommonCfg noticeCfg.CommonConfig
var IPLimit iplimitconfig.IPLimits
var SuperUidCfg noticeCfg.SuperUidConfig

func Start(c *cli.Context) {
	var logiclogName string = c.String("logiclog")
	ucfg.NewConfig(logiclogName, true, logiclog.PLogicLogger.ReturnLoadLogger())
	defer logiclog.PLogicLogger.StopLogger()

	cfgName := c.String("config")

	cfgApp := initCommonCfg(cfgName)
	// iplimit
	cfgLimit := initIpLimit(cfgName, c)

	// super uid
	suidCfgName := c.String("superuid")
	cfgUid := config.NewConfigToml(suidCfgName, &SuperUidCfg)
	logs.Debug("SuperUid %v", SuperUidCfg)

	if cfgApp == nil || cfgLimit == nil || cfgUid == nil {
		logs.Critical("Config Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	limit.Init()

	if CommonCfg.IsRunModeProdAndTest() {
		gin.SetMode(gin.ReleaseMode)
	}

	var waitGroup util.WaitGroupWrapper

	logs.Debug("Start Notice Server %v", CommonCfg)

	//metrics
	signalhandler.SignalKillFunc(func() { metrics.Stop() })
	waitGroup.Wrap(func() {
		metrics.Start("metrics.toml")
	})

	cmds.InitSentry(CommonCfg.SentryDSN)

	r, exitfun := cmds.MakeGinEngine()
	defer exitfun()
	routers2.RegNotice(r)

	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)

	util.PProfStart()

	logic.InitS3(CommonCfg.Dynamo_region, CommonCfg.S3_Buckets_Notice, CommonCfg.Dynamo_accessKeyID, CommonCfg.Dynamo_secretAccessKey)
	logic.InitNotice()
	logic.NewWatcher(CommonCfg.EtcdEndPoint, fmt.Sprintf("%s/%s", CommonCfg.EtcdRoot, CommonCfg.WatchKey))
	logic.StartUpdateServer()

	noticeCfg.InitNoticeMetrics()

	go func() {
		if CommonCfg.EnableHttpTLS != "" {
			if err := r.RunTLS(CommonCfg.HttpsPort,
				CommonCfg.HttpCertFile,
				CommonCfg.HttpKeyFile); err != nil {
				logs.Close()
				os.Exit(1)
			}
		} else {
			if err := r.Run(CommonCfg.Httpport); err != nil { // listen and serve on 0.0.0.0:8080
				logs.Close()
				os.Exit(1)
			}
		}
	}()

	//
	scfg := servers.NewConnServerCfg{
		ListenTo:             CommonCfg.ListenTrySsl,
		NumberOfAcceptor:     CommonCfg.NAcceptor,
		NumberOfWaitingQueue: CommonCfg.NWaitingConn,
		SslCfg:               CommonCfg.SslCfg,
		SslCaCfg:             CommonCfg.SslCaCfg,
	}
	slcfg := servers.NewLimitConnServerCfg{
		NewConnServerCfg: scfg,
		MaxConn:          CommonCfg.MaxConn,
	}
	ser := servers.NewLimitConnServer(slcfg)
	handler := newHandleClient()
	signalhandler.SignalKillHandler(ser)
	signalhandler.SignalKillHandler(handler)
	waitGroup.Wrap(func() {
		handler.handleConn(ser.ConnServer.GetWaitingConnChan())
	})
	waitGroup.Wrap(func() { ser.Start() })

	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()

	logs.Close()
}

func initCommonCfg(cfgName string) *config.Config {
	var common_cfg struct{ CommonConfig noticeCfg.CommonConfig }
	cfgApp := config.NewConfigToml(cfgName, &common_cfg)
	CommonCfg = common_cfg.CommonConfig
	return cfgApp
}

func initIpLimit(cfgName string, c *cli.Context) *config.Config {
	iplimitconfig.LoadIPRangeConfig(cfgName, func(ilcfg []iplimitconfig.IPRangeConfig) {
		limit.SetIPLimitCfg(ilcfg)
	})
	var limit_cfg struct{ LimitConfig limit.LimitConfig }
	cfgLimit := config.NewConfigToml(cfgName, &limit_cfg)

	ipCfgName := c.String("iplimit")
	config.NewConfigToml(ipCfgName, &IPLimit)

	limit.SetIPLimitCfg(IPLimit.IPLimit)
	logs.Debug("IpLimit cfg %v", IPLimit)

	limit.SetLimitCfg(limit_cfg.LimitConfig)
	return cfgLimit
}

type handleClient struct {
	quit chan struct{}
}

func newHandleClient() *handleClient {
	return &handleClient{
		quit: make(chan struct{}),
	}
}
func (cc *handleClient) handleConn(conn_ch <-chan net.Conn) {
	for {
		select {
		case <-cc.quit:
			return
		case conn, ok := <-conn_ch:
			if ok {
				conn.Close()
			}
		}
	}
}

func (cc *handleClient) Stop() {
	close(cc.quit)
}
