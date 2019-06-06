package main

import (
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/examples/paydbdumper/godb2log"
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
	godb2log.BuildPayInfo = buildIosPayInfo

	app.Run(os.Args)
}

func buildIosPayInfo(dataArray []string) godb2log.PayInfo {

	//var unix int64
	//
	//loc, _ := time.LoadLocation("Asia/Shanghai")
	//const shortForm = "2006-01-02 15:04:05"
	//t, _ := time.ParseInLocation(shortForm, dataArray[godb2log.PayTime], loc)
	//unix = int64(t.Unix())

	return godb2log.PayInfo{
		SdkChannelUid:    dataArray[godb2log.ChannelUID],
		SdkOrderNo:       dataArray[godb2log.OrderNo],
		SdkPayTime:       dataArray[godb2log.PayTime],
		Money:            dataArray[godb2log.MoneyAmount],
		Success:          true,
		SdkStatus:        dataArray[godb2log.Status],
		SdkIsTest:        dataArray[godb2log.IsTest],
		SdkNote:          "",
		GameOrderNo:      dataArray[godb2log.Uid],
		GameExtrasParams: dataArray[godb2log.ExtraParams],
		OrderIdx:         dataArray[godb2log.GoodIndex],
		Tistatus:         dataArray[godb2log.TiStatus],
		//ClientPayTime:    fmt.Sprintf("%d", unix),
		ClientPayTime: dataArray[godb2log.ReceiveTimestamp],
		ProductId:     parseProductId(dataArray[godb2log.Uid]),
		ClientVer:     dataArray[godb2log.Version],
	}
}

func parseProductId(gameOrderNo string) string {
	strArray := strings.Split(gameOrderNo, ":")
	if len(strArray) < 1 {
		return ""
	} else {
		return strArray[0]
	}
}
