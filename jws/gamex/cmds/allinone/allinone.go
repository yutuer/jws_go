package allinone

import (
	"os"
	"time"

	"github.com/codegangsta/cli"

	_ "github.com/rakyll/gom/http"

	"vcs.taiyouxi.net/jws/crossservice/util/discover"
	"vcs.taiyouxi.net/jws/crossservice/util/discover/publish"
	"vcs.taiyouxi.net/jws/gamex/cmds"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/servers/gate"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logics"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/mail"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/jws/gamex/models/pay"
	"vcs.taiyouxi.net/jws/gamex/modules"
	_ "vcs.taiyouxi.net/jws/gamex/modules/city_fish"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	_ "vcs.taiyouxi.net/jws/gamex/modules/data_hot_update"
	_ "vcs.taiyouxi.net/jws/gamex/modules/data_ver"
	_ "vcs.taiyouxi.net/jws/gamex/modules/gates_enemy"
	_ "vcs.taiyouxi.net/jws/gamex/modules/global_count"
	"vcs.taiyouxi.net/jws/gamex/modules/global_mail"
	_ "vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	_ "vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/redeem_code"
	_ "vcs.taiyouxi.net/jws/gamex/modules/room"
	_ "vcs.taiyouxi.net/jws/gamex/modules/simple_pvp_rander"
	_ "vcs.taiyouxi.net/jws/gamex/modules/warm"
	_ "vcs.taiyouxi.net/jws/gamex/modules/worship"
	"vcs.taiyouxi.net/jws/gamex/modules/ws_pvp"
	"vcs.taiyouxi.net/jws/gamex/sdk/samsung"
	"vcs.taiyouxi.net/jws/gamex/sdk/vivo"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/chat"
	"vcs.taiyouxi.net/platform/planx/util/anticheatlog"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/security"
	samsungutil "vcs.taiyouxi.net/platform/x/api_gateway/util"

	_ "vcs.taiyouxi.net/jws/gamex/modules/crossservice"
)

func init() {
	logs.Trace("allinone cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "allinone",
		Usage:  "启动游戏AllInOne模式",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "gate.toml",
				Usage: "Gate Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "gidconfig, cg",
				Value: "gidinfo.toml",
				Usage: "Gid Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "record, r",
				Value: "packetlog.xml",
				Usage: "Gate recording mode config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "logiclog, ll",
				Value: "logiclog.xml",
				Usage: "log player logic logs, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "anticheatlog, hl",
				Value: "anticheatlog.xml",
				Usage: "log anticheat logs, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "multiplayer, mp",
				Value: "",
				Usage: "start allinone with multiplayer if multiplayer.toml exist, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "pprofport, pp",
				Value: "",
				Usage: "pprof port",
			},
			cli.StringFlag{
				Name:  "iplimit, ip",
				Value: "iplimit.toml",
				Usage: "IpLimit Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.BoolFlag{
				Name:  "debug, d",
				Usage: "debug mode",
			},
		},
	})
}

type GameConfig struct {
	GameConfig game.Config
}

type JWSConfig struct {
	JWSConfig uutil.JWSConfig
}

type RankConfig struct {
	RankConfig modules.Config
}

type VivoConfig struct {
	VivoConfig vivo.VivoConfig
}

type SamsungConfig struct {
	SamsungConfig samsung.SamsungConfig
}

type GidConfig struct {
	GidConfig game.GidConfig
}

type ChatConfig struct {
	ChatConfig chat.Config
}

var IPLimit iplimitconfig.IPLimits

func Start(c *cli.Context) {
	util.PProfStart()
	logics.DebugTest()
	game.InitPayFeedBack()

	var recordCfgName string = c.String("record")
	config.NewConfig(recordCfgName, true, client.LoadPacketLogger)
	defer client.StopPacketLogger()

	var logiclogName string = c.String("logiclog")
	config.NewConfig(logiclogName, true, logiclog.PLogicLogger.ReturnLoadLogger())
	defer logiclog.PLogicLogger.StopLogger()

	var anticheatlogName string = c.String("anticheatlog")
	config.NewConfig(anticheatlogName, true, anticheatlog.PAntiCheatLogger.ReturnLoadLogger())
	defer anticheatlog.PAntiCheatLogger.StopLogger()

	cfgName := c.String("config")

	//For Game server
	var gameConfig GameConfig
	cfgApp := config.NewConfigToml(cfgName, &gameConfig)
	game.Cfg = gameConfig.GameConfig
	if cfgApp == nil {
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	//For Jws
	var jwsConfig JWSConfig
	cfgJws := config.NewConfigToml(cfgName, &jwsConfig)
	uutil.JwsCfg = jwsConfig.JWSConfig
	if cfgJws == nil {
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	//For Sentry
	var sentryConfig logs.SentryCfg
	config.NewConfigToml(cfgName, &sentryConfig)
	logs.InitSentryTags(sentryConfig.SentryConfig.DSN,
		map[string]string{
			"service": "gamex",
			"subcmd":  "allinone",
			"sid":     fmt.Sprint(game.Cfg.ShardId),
			"mode":    game.Cfg.RunMode,
		},
	)

	logs.Info("Game Server Config loaded %v.", game.Cfg)
	util.SetTimeLocal(game.Cfg.TimeLocal)

	ipCfgName := c.String("iplimit")
	config.NewConfigToml(ipCfgName, &IPLimit)

	security.LoadConf(game.Cfg.LimitValid,
		game.Cfg.RateLimitUrlInfo,
		game.Cfg.AdminIPs)
	security.SetIPLimitCfg(IPLimit.IPLimit)
	security.Init()

	shards := make([]uint, 0, len(game.Cfg.ShardId))
	shard2StartTime := make(map[uint]string, len(game.Cfg.ServerLogicStartTime))
	for i, sid := range game.Cfg.ShardId {
		shards = append(shards, sid)
		shard2StartTime[sid] = game.Cfg.ServerLogicStartTime[i]
	}
	util.SetServerStartTimes(shard2StartTime)

	// For chat
	var chatConfig ChatConfig
	cfgChat := config.NewConfigToml(cfgName, &chatConfig)
	chat.Cfg = chatConfig.ChatConfig
	if cfgChat == nil {
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	chat.Cfg.Init()
	logs.Info("Chat Server Config loaded %v.", chat.Cfg)

	//For Rank
	var rankConfig RankConfig
	cfgRank := config.NewConfigToml(cfgName, &rankConfig)
	modules.Cfg = rankConfig.RankConfig
	if cfgRank == nil {
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	logs.Info("Rank Config loaded %v.", modules.Cfg)

	// For vivo
	var vivoConfig VivoConfig
	cfgVivo := config.NewConfigToml(cfgName, &vivoConfig)
	vivo.Cfg = vivoConfig.VivoConfig
	if cfgVivo == nil {
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	logs.Info("vivo Config loaded %v.", vivo.Cfg)

	// For samsung
	var samsungConfig SamsungConfig
	cfgSamsung := config.NewConfigToml(cfgName, &samsungConfig)
	samsung.Cfg = samsungConfig.SamsungConfig
	if cfgSamsung == nil {
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	logs.Info("samsung Config loaded %v.", samsung.Cfg)
	samsungutil.Init(samsung.Cfg.PrivateKey, samsung.Cfg.PublicKey)

	//For GidConfig
	var gidConfig GidConfig
	cfgGid := config.NewConfigToml(c.String("gidconfig"), &gidConfig)
	game.GidCfg = gidConfig.GidConfig
	if cfgGid == nil {
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	game.GidCfg.Init()
	logs.Info("Gid Config loaded %v.", game.GidCfg)

	mailcfg := mailhelper.MailConfig{
		DBDriver:     game.Cfg.MailDBDriver,
		AWSRegion:    game.Cfg.AWS_Region,
		DBName:       game.Cfg.MailDBName,
		AWSAccessKey: game.Cfg.AWS_AccessKey,
		AWSSecretKey: game.Cfg.AWS_SecretKey,
		MongoDBUrl:   game.Cfg.MailMongoUrl,
	}

	//For Mails
	global_mail.SetConfig(mailcfg)
	mail_sender.SetConfig(mailcfg)
	redeemCodeModule.SetConfig(
		game.Cfg.AWS_Region,
		game.Cfg.AWS_AccessKey,
		game.Cfg.AWS_SecretKey,
		game.Cfg.RedeemCodeMongoUrl,
		game.Cfg.RedeemCodeDriver,
		"RedeemCode")

	debug := c.Bool("debug")
	game.Cfg.LocalDebug = debug
	if debug {
		logs.Info("gamex run in debug!")
	}
	if !debug {
		// start etcd
		err := etcd.InitClient(gameConfig.GameConfig.EtcdEndPoint)
		if err != nil {
			logs.Error("etcd InitClient err %s", err.Error())
			logs.Critical("\n")
			logs.Close()
			os.Exit(1)
		}
		logs.Warn("etcd init success")

		if !game.Cfg.GetGid() {
			logs.Error("gamex GetGid fail")
			logs.Critical("\n")
			logs.Close()
			os.Exit(1)
		}

		if !game.Cfg.GetInfoFromEtcd() {
			logs.Error("gamex GetInfoFromEtcd fail")
			logs.Critical("\n")
			logs.Close()
			os.Exit(1)
		}

		// game cfg need to register to etcd
		if !game.Cfg.SyncInfo2Etcd() {
			logs.Error("gamex SyncCfg2Etcd fail")
			logs.Critical("\n")
			logs.Close()
			os.Exit(1)
		}
		// modules Cfg Sync2Etcd
		if !modules.Cfg.Sync2Etcd() {
			logs.Error("modules Sync2Etcd fail")
			logs.Critical("\n")
			logs.Close()
			os.Exit(1)
		}
		logs.Warn("etcd gamex SyncInfo2Etcd success")
	}
	// start s3
	if err := uutil.InitCloudDb(); err != nil {
		logs.Error("s3 Init err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	gamedata.LoadGameData("")
	account.InitPhoneHistory()

	driver.SetupRedis(
		game.Cfg.Redis,
		game.Cfg.RedisDB,
		game.Cfg.RedisAuth,
		gamedata.IsGameDevMode(),
	)

	modules.SetupRedis(
		modules.Cfg.Redis,
		modules.Cfg.RedisDB,
		modules.Cfg.RedisAuth,
		false)

	ws_pvp.InitRedis(gamedata.IsGameDevMode())
	csrob.InitRedis(gamedata.IsGameDevMode())

	//服务发现的注册地址
	discover.InitEtcdServerCfg(discover.Config{
		Endpoints: game.Cfg.EtcdEndPoint,
		Root:      game.Cfg.EtcdRoot + "/" + discover.DefaultPathRoot,
	})
	publish.InitConfig(publish.Config{
		Endpoints: game.Cfg.EtcdEndPoint,
		Root:      game.Cfg.EtcdRoot + "/" + publish.DefaultPathRoot,
	})

	err := mail.InitMail(mailcfg)
	if err != nil {
		logs.Error("mail.InitDynamoDB Err By %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	//err = account_info.InitAccountInfoDynamoDB(
	//	game.Cfg.AWS_Region,
	//	game.Cfg.AccountInfoDynamoDB,
	//	game.Cfg.AWS_AccessKey,
	//	game.Cfg.AWS_SecretKey)
	//if err != nil {
	//	logs.Error("accountInfo.InitDynamoDB Err By %s", err.Error())
	//	logs.Critical("\n")
	//	logs.Close()
	//	os.Exit(1)
	//}

	err = pay.InitPayDB(pay.PayDBConfig{
		PayDBDriver:      game.Cfg.PayDBDriver,
		PayMongoUrl:      game.Cfg.PayMongoUrl,
		PayAndroidDBName: game.Cfg.PayAndroidDBName,
		PayIOSDBName:     game.Cfg.PayIOSDBName,
		AWSRegion:        game.Cfg.AWS_Region,
		AWSAccessKey:     game.Cfg.AWS_AccessKey,
		AWSSecretKey:     game.Cfg.AWS_SecretKey,
	})
	if err != nil {
		logs.Error("pay.InitPayDB Err By %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}

	// gate cfg need to register to etcd
	GateSer := gate.NewGateServer(cfgName)
	if !debug {
		if !GateSer.SyncCfg2Etcd() {
			logs.Error("gate SyncCfg2Etcd fail")
			logs.Critical("\n")
			logs.Close()
			os.Exit(1)
		}
		logs.Warn("etcd gate SyncCfg2Etcd success")
	}

	modules.StartModule(shards)
	gameserver := NewChanGameServerManager()

	var waitGroup util.WaitGroupWrapper
	GamexStartMultiplay(c, waitGroup)
	//database save service
	for _, shardId := range shards {
		driver.InitDBService(shardId)
	}

	//metrics
	signalhandler.SignalKillFunc(func() { metrics.Stop() })
	waitGroup.Wrap(func() { metrics.Start("metrics.toml") })

	//tcp gate server start up
	signalhandler.SignalKillHandler(GateSer)
	waitGroup.Wrap(func() { GateSer.Start(gameserver) })

	//handle kill signal
	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })

	time.Sleep(100 * time.Millisecond)
	util.LeakTestStart()

	// //debug--start
	// time.Sleep(10 * time.Second)
	// fmt.Println(worldboss.GetInfo(10, "lalala"))
	// //debug--end

	waitGroup.Wait()
	// modules需要Stop做一些收尾工作
	modules.StopModule()

	//gameServer.Stop()
	for {
		time.Sleep(1 * time.Second)
		if game.IsAllAccountOffLine() {
			break
		}
		logs.Warn("Waiting account offline ...")
	}

	fmt.Println("Gamex Closed ... ")
}
