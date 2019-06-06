package abot

import (
	"os"
	//"sync"

	//"github.com/astaxie/beego/httplib"
	"github.com/codegangsta/cli"

	"vcs.taiyouxi.net/botx/bot"
	"vcs.taiyouxi.net/botx/cmds"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func init() {

	logs.Trace("Bot cmd loaded")
	cmds.Register(&cli.Command{
		Name:   "bot",
		Usage:  "run a bot",
		Action: BotStart,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "rpc, r",
				Value: "127.0.0.1:8668",
				Usage: "gate server rpc ip:port",
			},
			cli.StringFlag{
				Name:  "server",
				Value: "127.0.0.1:8667",
				Usage: "gate server ip:port",
			},
			cli.Float64Flag{
				Name:  "speed, s",
				Value: 1.0,
				Usage: "2.0 2xtimes faster, 0 fixed 100ms",
			},
		},
	})
}

//单个机器人模式下，数据是流式读取, 解析成LogEntry然后发送给机器人。
//在单机模拟很多文件模式的情况下，需要一次性读取然后关闭FD，然后理由io.Reader进行重复利用。
//这样就能够大量节省实验机器的FD消耗。

func BotStart(c *cli.Context) {
	mybot := bot.NewPlayerBot(
		"0:0:87d5f092-7bb7-44a5-9869-b42fd9bf5899",
		c.String("server"),
		c.String("rpc"),
		c.Float64("speed"),
	)
	if mybot == nil {
		return
	}
	defer logs.Flush()

	var wg util.WaitGroupWrapper
	wg.Wrap(func() { mybot.Run(os.Stdin, nil, nil) })
	wg.Wait()
}
