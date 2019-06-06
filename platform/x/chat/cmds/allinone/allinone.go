package chat

import (
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/funny/linkext"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
	"vcs.taiyouxi.net/platform/planx/version"
	"vcs.taiyouxi.net/platform/x/auth/limit"
	"vcs.taiyouxi.net/platform/x/chat/chatserver"
	"vcs.taiyouxi.net/platform/x/chat/cmds"
)

const (
	metricsConf = "metrics.toml"
)

func init() {
	cmds.Register(&cli.Command{
		Name:   "allinone",
		Usage:  "开启所有功能",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "chat.toml",
				Usage: "Chat toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

func Start(c *cli.Context) {
	logs.Info("chat tcp server for town with rooms version:%s", version.GetVersion())
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	cfgName := c.String("config")
	readConfig(cfgName)

	logs.Info("start server socket mode")
	listener, err := net.Listen(
		"tcp4",
		chatserver.CommonCfg.NetAddress)

	if err != nil {
		logs.Critical("server listen error, %s", err)
		logs.Close()
		os.Exit(1)
	}

	gin2Gamex := gin.Default()
	//	g.Static("/chat", "./public/") // 正式服上不能开

	ginToGamex(gin2Gamex)
	go func() {
		if err := gin2Gamex.Run(chatserver.CommonCfg.NetGamexAddress); err != nil {
			logs.Critical("ginToGamex run err : %v", err)
			logs.Close()
			os.Exit(2)
		}
	}()
	var finallst net.Listener
	var server *linkext.ServerExt
	var ct link.CodecType
	switch chatserver.CommonCfg.NetType {
	case "websocket":
		// client的websocket需要带上origin-》“/ws”
		logs.Info("start server websocket mode")
		logs.Warn(`client的websocket需要带上origin -> /ws`)

		gsl, g := linkext.NewGinWebSocketListen(nil, listener, "/ws")
		linkext.NewGinDebugUrl(g)
		go func() {
			http.Serve(listener, g)
		}()
		finallst = gsl
		//websocket has buffer io underlying implementation
		// ct = link.Async(64, link.RawBytes())
		//ct = linkext.RawBytes()
		// ct = linkext.Async(64, linkext.WebsocketType())
		ct = linkext.Async(128, linkext.WebsocketWithWriteDeadline())
	default:
		finallst = listener
		ct = link.Bufio(linkext.String())
		logs.Info("start server tcp mode")
	}

	server = linkext.NewServerExt(finallst, ct)
	server.AcceptHandle(func(s *link.Session, err error) error {
		if err != nil {
			logs.Error("Accept Err By %s", err.Error())
			return err
		}
		logs.Debug("Accept a Session!")
		player := chatserver.NewPlayer(s)
		go player.Start()
		return nil
	})

	server.StartServer()

	var waitGroup util.WaitGroupWrapper
	signalhandler.SignalKillHandler(server)
	waitGroup.Wrap(func() { server.Wait() })
	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()

}

func ginToGamex(g *gin.Engine) {
	gamex_api := g.Group("/", limit.InternalOnlyLimit())

	// 接收game服务器消息广播给所有玩家
	gamex_api.POST("broadcast", func(c *gin.Context) {
		var json BroadCastMsg
		err := c.BindJSON(&json)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
			return
		}
		go func() {
			chatserver.RouteBroadCast(json.Type, json.Shd, json.Msg, json.Acids)
		}()
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 接受game服务器消息，返回给game账号验证信息
	gamex_api.POST("auth", func(c *gin.Context) {
		acid := c.PostForm("acid")
		if acid == "" {
			c.String(http.StatusBadRequest, "")
			return
		}
		curAuh, _ := chatserver.GetAuthSource(acid)
		c.String(http.StatusOK, curAuh)
	})

	// 清理人数过少的房间
	gamex_api.GET("filterroom/:shard", func(c *gin.Context) {
		shard := c.Param("shard")
		chatserver.ClearRoomsTooFewPlayer(shard)
		c.String(http.StatusOK, "ok")
	})
}

type BroadCastMsg struct {
	Type  string   `form:"type" json:"type" binding:"required"`
	Shd   string   `form:"shd" json:"shd" binding:"required"`
	Msg   string   `form:"msg" json:"msg" binding:"required"`
	Acids []string `form:"acids" json:"acids"`
}

func readConfig(chatFile string) {
	var common_cfg struct{ CommonCfg chatserver.CommonConfig }
	cfgApp := config.NewConfigToml(chatFile, &common_cfg)
	chatserver.CommonCfg = common_cfg.CommonCfg
	if cfgApp == nil {
		logs.Critical("CommonConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	logs.Info("Common Config loaded %v.", chatserver.CommonCfg)

	// start etcd
	err := etcd.InitClient(chatserver.CommonCfg.EtcdEndPoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	if err := chatserver.CommonCfg.SidMergeInfo(); err != nil {
		logs.Critical("SidMergeInfo Error %s \n", err.Error())
		logs.Close()
		os.Exit(1)
	}

	var limit_cfg struct{ LimitConfig limit.LimitConfig }
	cfgLimit := config.NewConfigToml(chatFile, &limit_cfg)
	limit.SetLimitCfg(limit_cfg.LimitConfig)
	if cfgLimit == nil {
		logs.Critical("LimitConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	logs.Info("Limit Config loaded %v.", limit.LimitCfg)

	//metrics
	signalhandler.SignalKillFunc(func() { metrics.Stop() })
	metrics.Start(metricsConf)
}
