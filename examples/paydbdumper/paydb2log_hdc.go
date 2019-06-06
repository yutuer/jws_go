package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"strconv"
	"strings"
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
	godb2log.DoParse = parse2hdc

	app.Run(os.Args)
}

//
func parse2hdc(csvData []string) {
	paySt := calcPayTimeInt1(csvData[godb2log.PayTimestamp])
	if paySt < 1480608000 || paySt > 1480694400 {
		return
	}

	dbAccount, err := db.ParseAccount(csvData[godb2log.Uid])
	if err != nil {
		godb2log.PrintlnError("parse account id error, %s, %v", csvData[godb2log.Uid], err)
		return
	}
	channel := csvData[godb2log.Channel]
	if channel == "" {
		channel = "110134101106.0"
	}
	platform := convertPlatform(csvData[godb2log.Platform])
	avatar := "0"
	money := csvData[godb2log.MoneyAmount]
	corpLevel := "0"
	vipLevel := "0"
	ip := "0.0.0.0"
	payTime := fmtPayTimeForHdc(csvData[godb2log.PayTimestamp])
	gameOrderId := fmt.Sprintf("%s:%d:%s", dbAccount.String(), 0, csvData[godb2log.GoodIndex])
	logTime := fmtPayTimeForHdc(csvData[godb2log.SN][:10])
	bdcStr := fmt.Sprintf("102\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t1\t%s\t%s\t%s\tCNY\t%s\t%s\t0"+
		"\t%s\t%s\t%s\t%s\t%s\t%s\tplayercharger",
		channel,
		dbAccount.ShardId,
		platform,
		dbAccount.String(),
		avatar,
		csvData[godb2log.OrderNo],
		gameOrderId, // gameorderId
		logTime,     // time1
		payTime,
		money+"00",
		csvData[godb2log.HcBuy],
		csvData[godb2log.HcGive],
		csvData[godb2log.GoodIndex],
		csvData[godb2log.RoleName],
		corpLevel,
		vipLevel,
		"0",
		ip)
	fmt.Println(bdcStr)
}

func convertPlatform(platform string) string {
	switch platform {
	case "ios":
		return "0"
	case "android":
		return "1"
	}
	return "3"
}

func fmtPayTimeForHdc(payTimeStamp string) string {
	fmtTime := godb2log.FmtPayTime(payTimeStamp)
	fmtTime = strings.Replace(fmtTime, "-", "", -1)
	fmtTime = strings.Replace(fmtTime, " ", "", -1)
	fmtTime = strings.Replace(fmtTime, ":", "", -1)
	return fmtTime
}

func calcPayTimeInt1(timeStr string) int64 {
	st, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return 0
	} else {
		return st
	}
}
