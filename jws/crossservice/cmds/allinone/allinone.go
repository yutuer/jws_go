package allinone

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/uuid"

	"github.com/codegangsta/cli"
	_ "github.com/rakyll/gom/http"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"

	"vcs.taiyouxi.net/jws/crossservice/cmds"
	csCfg "vcs.taiyouxi.net/jws/crossservice/config"
	"vcs.taiyouxi.net/jws/crossservice/metrics"
	"vcs.taiyouxi.net/jws/crossservice/server"
	dbFromEtcd "vcs.taiyouxi.net/jws/crossservice/util/csdb/frometcd"
	"vcs.taiyouxi.net/jws/crossservice/util/discover"
	"vcs.taiyouxi.net/jws/crossservice/util/discover/exclusion"
	"vcs.taiyouxi.net/jws/crossservice/util/discover/publish"
	"vcs.taiyouxi.net/jws/crossservice/util/http_util"
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
	logs.Info("CrossService start server")
	var waitGroup util.WaitGroupWrapper

	pp := c.String("pprofport")
	if pp != "" {
		go func() {
			http.ListenAndServe("localhost:"+pp, nil)
		}()
	}

	cfgName := c.String("config")
	var common_cfg struct{ CommonCfg csCfg.CommonConfig }
	cfgApp := config.NewConfigToml(cfgName, &common_cfg)

	if cfgApp == nil {
		logs.Critical("CommonConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	csCfg.Cfg = common_cfg.CommonCfg
	logs.Info("CommonCfg: %+v", csCfg.Cfg)

	//For Sentry
	logs.InitSentryTags(csCfg.Cfg.DSN,
		map[string]string{
			"service": "crossservice",
			"subcmd":  "allinone",
			"gid":     fmt.Sprint(csCfg.Cfg.Gid),
			"mode":    csCfg.Cfg.RunMode,
		},
	)

	// start etcd
	err := etcd.InitClient(csCfg.Cfg.EtcdEndPoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	logs.Warn("etcd init success")

	gamedata.LoadGameData("")

	csCfg.SetIndex(uuid.NewV4().String())

	//服务发现的注册地址
	discover.InitEtcdServerCfg(discover.Config{
		Endpoints: csCfg.Cfg.EtcdEndPoint,
		Root:      csCfg.Cfg.EtcdRoot + "/" + discover.DefaultPathRoot,
	})
	publish.InitConfig(publish.Config{
		Endpoints: csCfg.Cfg.EtcdEndPoint,
		Root:      csCfg.Cfg.EtcdRoot + "/" + publish.DefaultPathRoot,
	})
	exclusion.InitConfig(exclusion.Config{
		Endpoints: csCfg.Cfg.EtcdEndPoint,
		Root:      csCfg.Cfg.EtcdRoot + "/" + exclusion.DefaultPathRoot,
	})

	http_util.Init(fmt.Sprintf("%s:%d", csCfg.Cfg.PublicIP, csCfg.Cfg.InternalHTTPPort), csCfg.Cfg.Gid)
	regGinHandles()

	groupIDs := occupyGroup()
	defer releaseGroup()
	logs.Info("OccupyGroup: %+v", groupIDs)
	csCfg.Cfg.GroupIDs = groupIDs

	//DB注册
	dbFromEtcd.InitRedis(csCfg.IsDevMode())

	server := server.NewServer(csCfg.Cfg.Gid, csCfg.Cfg.PublicIP, csCfg.Cfg.PublicPort)
	server.AddGroupIDs(groupIDs)

	waitGroup.Wrap(func() { metrics.Start("metrics.toml") })
	cs_metrics.Reg()

	http_util.Run()

	if 0 != len(csCfg.Cfg.IPFilter) {
		server.SetIPFilter(csCfg.Cfg.IPFilter)
	}

	<-time.After(1 * time.Second)

	if err := server.Start(); nil != err {
		logs.Critical("Server Start Error, %v\n", err)
		logs.Close()
		os.Exit(1)
		return
	}
	logs.Info("Server start listen: %v", server.LocalAddr().String())

	//metrics
	signalhandler.SignalKillFunc(func() { metrics.Stop() })

	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()

	server.Stop()
	logs.Info("Server Stop")

	logs.Close()
}
