package pay

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	ucfg "vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/ginhelper"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
	"vcs.taiyouxi.net/platform/x/api_gateway/cmds"
	"vcs.taiyouxi.net/platform/x/api_gateway/config"
	"vcs.taiyouxi.net/platform/x/api_gateway/pay"
)

func init() {
	logs.Trace("androidpay cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "pay",
		Usage:  "开启所有功能",
		Action: Start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Value: "pay.toml",
				Usage: "pay toml config, in {CWD}/conf/ or {AppPath}/conf",
			},
			cli.StringFlag{
				Name:  "logiclog, ll",
				Value: "logiclog.xml",
				Usage: "log player logic logs, in {CWD}/conf/ or {AppPath}/conf",
			},
		},
	})
}

func Start(c *cli.Context) {
	config.ReadConfig(c)

	var waitGroup util.WaitGroupWrapper
	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })

	//加载数据,其实只加载了iapbase.data, iapmain.data
	//TODO: 如何只加载这两个数据表?, 如何热更?从DynamoDB里面加载配置?

	gamedata.LoadGameData("")

	// gin for sdk
	gSdk := cmds.MakeGinEngine()

	var gh gin.HandlerFunc
	var exitfun func()
	ucfg.NewConfig("accesslog.xml", false, func(lcfgname string, cmd ucfg.LoadCmd) {
		gh, exitfun = ginhelper.NgixLoggerToFile(lcfgname)
	})
	defer exitfun()

	var logiclogName string = c.String("logiclog")
	ucfg.NewConfig(logiclogName, true, logiclog.PLogicLogger.ReturnLoadLogger())
	defer logiclog.PLogicLogger.StopLogger()

	gSdk.Use(gh)

	// quick sdk
	err := pay.RegSdk(gSdk,
		func(run func()) { waitGroup.Wrap(run) },
		func(stop func()) { signalhandler.SignalKillFunc(stop) },
	)
	if err != nil {
		logs.Error("RegQuickSdk failed %s", err.Error())
		logs.Close()
		os.Exit(1)
	}

	go func() {
		err := gSdk.Run(config.PayCfg.SdkHttpPort)
		if err != nil {
			logs.Error("RegSdk failed %s", err.Error())
			logs.Close()
			os.Exit(1)
		}
	}()

	waitGroup.Wait()
	logs.Close()
}
