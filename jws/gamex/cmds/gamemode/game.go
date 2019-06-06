package gamemode

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/codegangsta/cli"

	"vcs.taiyouxi.net/jws/gamex/cmds"
	"vcs.taiyouxi.net/jws/gamex/logics"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/mail"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"

	"vcs.taiyouxi.net/jws/gamex/models/pay"
	"vcs.taiyouxi.net/jws/gamex/modules"
	_ "vcs.taiyouxi.net/jws/gamex/modules/data_ver"
	"vcs.taiyouxi.net/jws/gamex/modules/global_mail"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/modules/redeem_code"
	_ "vcs.taiyouxi.net/jws/gamex/modules/simple_pvp_rander"
	_ "vcs.taiyouxi.net/jws/gamex/modules/warm"
	"vcs.taiyouxi.net/platform/planx/servers/chat"
	"vcs.taiyouxi.net/platform/planx/util/anticheatlog"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/security"
)

const ()

func init() {
	logs.Trace("game cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "game",
		Usage:  "启动Game模式",
		Action: GameStart,
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
				Name:  "iplimit, ip",
				Value: "iplimit.toml",
				Usage: "IpLimit Configuration toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

type GameConfig struct {
	GameConfig game.Config
}

type RankConfig struct {
	RankConfig modules.Config
}

type GidConfig struct {
	GidConfig game.GidConfig
}

type ChatConfig struct {
	ChatConfig chat.Config
}

var IPLimit iplimitconfig.IPLimits

func GameStart(c *cli.Context) {
	rand.Seed(time.Now().UTC().UnixNano())
	fmt.Println("PlanX Game Server, version:", game.GetVersion())

	var logiclogName string = c.String("logiclog")
	config.NewConfig(logiclogName, true, logiclog.PLogicLogger.ReturnLoadLogger())
	defer logiclog.PLogicLogger.StopLogger()

	var anticheatlogName string = c.String("anticheatlog")
	config.NewConfig(anticheatlogName, true, anticheatlog.PAntiCheatLogger.ReturnLoadLogger())
	defer anticheatlog.PAntiCheatLogger.StopLogger()

	var cfgName string = c.String("config")

	var gameConfig GameConfig
	cfgApp := config.NewConfigToml(cfgName, &gameConfig)
	game.Cfg = gameConfig.GameConfig

	if cfgApp == nil {
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	logs.Info("Game Server Config loaded %v.", game.Cfg)

	//For Sentry
	var sentryConfig logs.SentryCfg
	config.NewConfigToml(cfgName, &sentryConfig)
	logs.InitSentryTags(sentryConfig.SentryConfig.DSN,
		map[string]string{
			"service": "gamex",
			"subcmd":  "gamemode",
			"sid":     fmt.Sprint(game.Cfg.ShardId),
			"mode":    game.Cfg.RunMode,
		},
	)

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

	mailCfg := mailhelper.MailConfig{
		DBDriver:     game.Cfg.MailDBDriver,
		AWSRegion:    game.Cfg.AWS_Region,
		DBName:       game.Cfg.MailDBName,
		AWSAccessKey: game.Cfg.AWS_AccessKey,
		AWSSecretKey: game.Cfg.AWS_SecretKey,
		MongoDBUrl:   game.Cfg.MailMongoUrl,
	}
	//For Mails
	global_mail.SetConfig(mailCfg)
	mail_sender.SetConfig(mailCfg)
	redeemCodeModule.SetConfig(
		game.Cfg.AWS_Region,
		game.Cfg.AWS_AccessKey,
		game.Cfg.AWS_SecretKey,
		game.Cfg.RedeemCodeMongoUrl,
		game.Cfg.RedeemCodeDriver,
		"RedeemCode")

	// etcd
	err := etcd.InitClient(gameConfig.GameConfig.EtcdEndPoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
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
	gamedata.LoadGameData("")
	driver.SetupRedis(
		game.Cfg.Redis,
		game.Cfg.RedisDB,
		game.Cfg.RedisAuth,
		gamedata.IsGameDevMode(),
	)
	modules.SetupRedis(modules.Cfg.Redis,
		modules.Cfg.RedisDB,
		modules.Cfg.RedisAuth,
		false)

	err = mail.InitMail(mailCfg)

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

	account.InitPhoneHistory()

	modules.StartModule(shards)

	server := game.NewTCPServer(game.Cfg.Listen)

	mux := logics.CreatePlayer

	var waitGroup util.WaitGroupWrapper
	waitGroup.Wrap(func() { server.Start(mux) })
	signalhandler.SignalKillHandler(server)

	for _, shardId := range shards {
		driver.InitDBService(shardId)
	}
	//metrics
	signalhandler.SignalKillFunc(func() { metrics.Stop() })
	waitGroup.Wrap(func() {
		metrics.Start("metrics.toml")
	})

	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()

	// modules需要Stop做一些收尾工作
	modules.StopModule()

	for {
		time.Sleep(1 * time.Second)
		if game.IsAllAccountOffLine() {
			break
		}
		logs.Warn("Waiting account offline ...")
	}
}
