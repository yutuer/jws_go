package allinone

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"fmt"

	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
	"vcs.taiyouxi.net/platform/x/gm_tools/cmds"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/authdb"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/ftp"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/hero_gag"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/qiniu"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/store"
	gmConfig "vcs.taiyouxi.net/platform/x/gm_tools/config"
	"vcs.taiyouxi.net/platform/x/gm_tools/login"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/account"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/act_valid"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/ban_account"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/channel_url"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/data_ver"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/device"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/gag_account"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/iap"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/mail"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/player_account"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/profile_tool"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/rank_del"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/server_mng"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/server_show_state"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/sys_public"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/sys_roll_notice"
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
			cli.StringFlag{
				Name:  "yxiplimit, ip",
				Value: "yxiplimit.toml",
				Usage: "YXIpLimit Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

var CommonCfg gmConfig.CommonConfig
var IPLimit iplimitconfig.IPLimits

func Start(c *cli.Context) {
	pp := c.String("pprofport")
	if pp != "" {
		go func() {
			http.ListenAndServe("localhost:"+pp, nil)
		}()
	}

	cfgName := c.String("config")
	var common_cfg struct{ CommonCfg gmConfig.CommonConfig }
	cfgApp := config.NewConfigToml(cfgName, &common_cfg)
	CommonCfg = common_cfg.CommonCfg

	if cfgApp == nil {
		logs.Critical("CommonConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}
	CommonCfg.MergedShard = make([]string, 0, len(CommonCfg.RedisName))
	CommonCfg.MergedShard = append(CommonCfg.MergedShard, CommonCfg.RedisName...)
	CommonCfg.RedisAuth = make([]string, len(CommonCfg.RedisName))

	err := etcd.InitClient(CommonCfg.EtcdEndpoint)
	if err != nil {
		logs.Error("etcd InitClient %s", err.Error())
	}

	err = gmConfig.CfyByServers.LoadFromEtcd(CommonCfg.GameSerEtcdRoot, CommonCfg.GidFilter)
	if err != nil {
		logs.Critical("CommonConfig Read Error %s\n", err.Error())
		logs.Close()
		os.Exit(1)
	}

	//err = CommonCfg.LoadFromEtcd(CommonCfg.GmToolsCfg)
	//if err != nil {
	//	logs.Warn("CommonConfig Read Error %s\n", err.Error())
	//}

	if CommonCfg.Runmode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	gmConfig.CfyByServers.AddCfg(&CommonCfg)
	gmConfig.Cfg = CommonCfg

	//读取配置时区信息
	util.SetTimeLocal(gmConfig.Cfg.TimeLocal)

	var waitGroup util.WaitGroupWrapper

	logs.Info("Start Tool %v", gmConfig.Cfg)
	logs.Info("CommonCfg %v", CommonCfg)

	r := gin.Default()

	file_to_store := CommonCfg.DBStoreFile
	if file_to_store == "" {
		file_to_store = "./data.db"
	}
	err = store.InitStore(file_to_store)

	if err != nil {
		logs.Error("InitStore Err By %s", err.Error())
		return
	}

	profile_tool.SameServer(CommonCfg.RedisName,
		CommonCfg.MergedShard)
	profile_tool.InitDB()
	profile_tool.RouteProfileGet(r)
	profile_tool.RegCommands()

	mail.InitMail(CommonCfg)
	mail.RouteMailGet(r)
	mail.RegCommands()
	gm_command.CommandGet(r)

	/*************HERO GAG**********************/
	// 设置可以操作的限定IP
	ipCfgName := c.String("yxiplimit")
	config.NewConfigToml(ipCfgName, &IPLimit)
	hero_gag.SetCommonIPLimitCfg(IPLimit.IPLimit)
	hero_gag.InitHeroGag(r)
	/*******************************************/

	err = sys_roll_notice.Start()
	if err != nil {
		logs.Error("sys_roll_notice Start Err By %s", err.Error())
		return
	}
	sys_roll_notice.RegCommands()

	login.Init()
	login.RouteLogin(r)
	account.RegCommands()

	sys_public.Init()
	sys_public.RegCommands()

	ban_account.RegCommands()
	gag_account.RegCommands()

	channel_url.RegCommands()
	if err := channel_url.Init(); err != nil {
		logs.Error("init channel url err : %v", err)
	}

	iap.RegCommands()
	server_mng.RegCommands()
	server_show_state.RegCommands()
	act_valid.RegCommands()
	data_ver.RegCommands()
	rank_del.RegCommands()
	device.RegCommands()
	device.InitDb(CommonCfg)

	ftp.SetCfg(
		CommonCfg.FtpAddress,
		CommonCfg.FtpUser,
		CommonCfg.FtpPasswd,
		"")

	qiniu.SetCfg(
		CommonCfg.QiNiuAccKey,
		CommonCfg.QiNiuSecKey)

	authdb.Init(CommonCfg)
	player_account.RegCommands()

	r.Static("/webui/dist", "./public/dist")
	getIndexHtml(r)
	go r.Run(CommonCfg.Url) // listen and serve on 0.0.0.0:8080

	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()

	login.Save()
	sys_public.Save()
	sys_roll_notice.Stop()

	store.StopStore()
	logs.Close()
	time.Sleep(1 * time.Second)
}

func getIndexHtml(r *gin.Engine) {
	r.LoadHTMLFiles("./public/index.tmpl")
	fmt.Println("html template", r.HTMLRender)
	r.GET("/webui/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": gmConfig.Cfg.GidFilter[0],
			"zone":  gmConfig.Cfg.GidFilter[0],
		})
	})
}
