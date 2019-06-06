package allinone

import (
	"net"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"

	//A visual interface to work with runtime profiling data from Go programs.
	//instead of pprof/http
	_ "github.com/rakyll/gom/http"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/cmds"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/config"

	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/notify"

	"fmt"

	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/gve"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/gvg"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/teamboss"
	"vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/funny/linkext"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
)

func init() {
	logs.Trace("allinone cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "allinone",
		Usage:  "开启所有功能",
		Action: StartAllInOne,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "config.toml",
				Usage: "Onland Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "logiclog, ll",
				Value: "logiclog.xml",
				Usage: "log player logic logs, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "pprofport, pp",
				Value: "",
				Usage: "pprof port",
			},
		},
	})
}

// StartAllInOne is  main function of subcommand "allinone"
func StartAllInOne(c *cli.Context) {
	util.PProfStart()
	logs.Info("start server")

	//debug port and http server start
	pp := c.String("pprofport")
	if pp != "" {
		go func() {
			http.ListenAndServe("localhost:"+pp, nil)
		}()
	}

	var logiclogName string = c.String("logiclog")
	config.NewConfig(logiclogName, true, logiclog.PLogicLogger.ReturnLoadLogger())
	defer logiclog.PLogicLogger.StopLogger()

	gamedata.LoadGameData("")
	var waitGroup util.WaitGroupWrapper
	cfgName := c.String("config")
	StartMultiplay(cfgName, waitGroup)
	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()
	logs.Close()
}

func StartMultiplay(multiplayConfig string, wg util.WaitGroupWrapper) error {
	if multiplayConfig == "" {
		return fmt.Errorf("StartMultiplay config parameter empty")
	}
	var MpCfg struct{ CommonCfg multConfig.CommonConfig }
	cfgApp := config.NewConfigToml(multiplayConfig, &MpCfg)
	if cfgApp == nil {
		logs.Critical("StartMultiplay Config Read Error\n")
		logs.Close()
		os.Exit(1)
	}
	multConfig.Cfg = MpCfg.CommonCfg
	multConfig.Cfg.PublicIP = util.GetPublicIP(multConfig.Cfg.PublicIP, multConfig.Cfg.Listen)
	logs.Info("StartMultiplay with matchtoken %s.", multConfig.Cfg.MatchToken)
	if multConfig.Cfg.MatchToken == "" {
		multConfig.Cfg.MatchToken = helper.MatchDefaultToken
	}
	logs.Info("StartMultiplay with public ip %s binded.", multConfig.Cfg.PublicIP)

	//NOTE: EtcEndpint的配置一般都是相同的
	etcd.InitClient(multConfig.Cfg.EtcdEndpoint)

	notify.GetNotify().Start()

	signalhandler.SignalKillHandler(notify.GetNotify())

	ginEngine := gin.New()
	gve.GVEGamesMgr.StartHttp(ginEngine)
	teamboss.TBGamesMgr.StartHttp(ginEngine)
	gvg.GVGGamesMgr.StartHttp(ginEngine)

	gve.GVEGamesMgr.Start(ginEngine)
	signalhandler.SignalKillHandler(&gve.GVEGamesMgr)

	teamboss.TBGamesMgr.Start(ginEngine)
	signalhandler.SignalKillHandler(&teamboss.TBGamesMgr)

	gvg.GVGGamesMgr.Start(ginEngine)
	signalhandler.SignalKillHandler(&gvg.GVGGamesMgr)

	wg.Wrap(func() { metrics.Start("metrics.toml") })
	signalhandler.SignalKillFunc(func() { metrics.Stop() })
	logs.Debug(" start all game mgr")
	go func() {
		logs.Debug("run gin engine")
		err := ginEngine.Run(multConfig.Cfg.ListenNotifyAddr)
		if err != nil {
			panic(err)
		}
	}()

	listener, err := net.Listen(
		"tcp4",
		multConfig.Cfg.Listen)

	if err != nil {
		logs.Critical("server listen error, %s", err)
		logs.Close()
		os.Exit(1)
	}

	//Gve
	gsl, g := linkext.NewGinWebSocketListen(nil, listener, "/ws")
	ct := linkext.Async(64, linkext.WebsocketWithWriteDeadline())
	server := linkext.NewServerExt(gsl, ct)
	server.AcceptHandle(func(s *link.Session, err error) error {
		if err != nil {
			return err
		}
		player := gve.NewPlayer(s)
		go player.Start()
		return nil
	})
	server.StartServer()
	signalhandler.SignalKillHandler(server)

	//Fenghuo
	//func() {
	//	gsl, _ := linkext.NewGinWebSocketListen(g, listener, "/wsfenghuo")
	//	ct := linkext.Async(64, linkext.WebsocketWithWriteDeadline())
	//	server := linkext.NewServerExt(gsl, ct)
	//	server.AcceptHandle(func(s *link.Session, err error) error {
	//		if err != nil {
	//			return err
	//		}
	//		player := fenghuo.NewFenghuoPlayer(s)
	//		go player.Start()
	//		return nil
	//	})
	//	server.StartServer()
	//	signalhandler.SignalKillHandler(server)
	//}()

	// teamboss
	func() {
		tb_gsl, _ := linkext.NewGinWebSocketListen(g, listener, "/teamboss")
		tb_ct := linkext.Async(256, linkext.WebsocketWithWriteDeadline())
		tb_server := linkext.NewServerExt(tb_gsl, tb_ct)
		tb_server.AcceptHandle(func(s *link.Session, err error) error {
			if err != nil {
				return err
			}
			player := teamboss.NewPlayer(s)
			go player.Start()
			return nil
		})
		tb_server.StartServer()
		signalhandler.SignalKillHandler(tb_server)
	}()

	// gvg
	func() {
		gvg_gsl, _ := linkext.NewGinWebSocketListen(g, listener, "/gvg")
		gvg_ct := linkext.Async(128, linkext.WebsocketWithWriteDeadline())
		gvg_server := linkext.NewServerExt(gvg_gsl, gvg_ct)
		gvg_server.AcceptHandle(func(s *link.Session, err error) error {
			if err != nil {
				return err
			}
			player := gvg.NewPlayer(s)
			go player.Start()
			return nil
		})
		gvg_server.StartServer()
		signalhandler.SignalKillHandler(gvg_server)
	}()

	go func() {
		http.Serve(listener, g)
	}()

	wg.Wrap(func() { server.Wait() })

	return nil
}
