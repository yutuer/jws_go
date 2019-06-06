package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"strconv"
	"vcs.taiyouxi.net/examples/paydbdumper/godb2log"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/version"
)

func main() {
	app := cli.NewApp()

	app.Version = version.GetVersion()
	app.Name = "paydb2log"
	app.Usage = "palyerdb2log"
	app.Author = "LBB"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "channel",
			Value: "",
			Usage: "指定所有channel为空的订单全采用该值",
		},
		cli.StringFlag{
			Name:  "time",
			Value: "",
			Usage: "筛选该时间之前的订单, ns",
		},
	}

	app.Action = godb2log.StartParse
	godb2log.DoParse = parse2bdc

	app.Run(os.Args)
}

//
func parse2bdc(csvData []string) {
	paySt := calcPayTimeInt(csvData[godb2log.PayTimestamp])
	if paySt < 1480608000 || paySt > 1480694400 {
		return
	}
	dbAccount, err := db.ParseAccount(csvData[godb2log.Uid])
	if err != nil {
		godb2log.PrintlnError("parse account id error, %s, %v", csvData[godb2log.Uid], err)
		return
	}
	deviceId := dbAccount.UserId
	accountNameId := fmt.Sprintf("%d:%s", dbAccount.GameId, dbAccount.UserId)
	channel := csvData[godb2log.Channel]
	if channel == "" {
		channel = "110134101106.0"
	}
	hc := godb2log.CalcHc(csvData[godb2log.HcBuy], csvData[godb2log.HcGive])
	orderTime := godb2log.FmtPayTime(csvData[godb2log.PayTimestamp])
	level := "0"
	sid := fmt.Sprintf("%s%04s%06d", "11", "134", dbAccount.ShardId) //
	bdcStr := fmt.Sprintf("%s$$%s$$%s$$%s$$%s$$%s$$%s$$%s$$%s$$%s$$%s$$%s$$%s",
		deviceId,
		accountNameId,
		csvData[godb2log.AccountName],
		channel,
		dbAccount.String(),
		csvData[godb2log.RoleName],
		level,
		hc,
		csvData[godb2log.MoneyAmount],
		csvData[godb2log.GoodIndex],
		csvData[godb2log.OrderNo],
		orderTime,
		sid)
	fmt.Println(bdcStr)
}

func calcPayTimeInt(timeStr string) int64 {
	st, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return 0
	} else {
		return st
	}
}
