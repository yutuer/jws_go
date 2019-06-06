package godb2log

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"os"
	"strconv"
	"time"
	"vcs.taiyouxi.net/platform/planx/servers/db"
)

const (
	AccountName = iota
	Channel
	ChannelUID
	ExtraParams
	GoodIndex
	GoodName
	HcBuy
	HcGive
	IsTest
	Mobile
	MoneyAmount
	OrderNo
	PayTime
	PayTimestamp
	Platform
	ProductID
	ReceiveTimestamp
	RoleName
	SN
	Status
	TiStatus
	Uid
	Version
)

type logStruct struct {
	Level     string  `json:"Level"`
	Logtime   int64   `json:"logtime"`
	Timestamp string  `json:"@timestamp"`
	Utc8      string  `json:"utc8"`
	Type_name string  `json:"type_name"`
	Gid       uint    `json:"gid"`
	Sid       uint    `json:"sid"`
	Userid    string  `json:"userid"`
	Accountid string  `json:"accountid"`
	Avatar    string  `json:"avatar"`
	Channel   string  `json:"channel"`
	Info      PayInfo `json:"info"`
	Extra     string  `json:"extra"`
	//Log_time  string  `json:"log_time"`
}

type PayInfo struct {
	SdkChannelUid    string
	SdkOrderNo       string
	SdkPayTime       string
	Money            string
	Success          bool
	SdkStatus        string
	SdkIsTest        string
	SdkNote          string
	GameOrderNo      string
	GameExtrasParams string
	OrderIdx         string
	Tistatus         string
	ClientPayTime    string
	ProductId        string
	ClientVer        string
}

func StartParse(c *cli.Context) {
	timeByBefore, channelParam := readCmdParam(c)
	reader := csv.NewReader(os.Stdin)
	parseCsvData(reader, timeByBefore, channelParam)
}

func readCmdParam(c *cli.Context) (int64, string) {
	timeParam := c.String("time")
	channelParam := c.String("channel")
	var timeByBefore int64
	var err error
	if timeParam != "" {
		timeByBefore, err = strconv.ParseInt(timeParam, 10, 64)
		if err != nil {
			PrintlnError("parse time error %v", err)
			return 0, ""
		}
	}
	return timeByBefore, channelParam
}

type FuncDoParse func(csvData []string)

var DoParse FuncDoParse = print2Json

func parseCsvData(reader *csv.Reader, beforeTime int64, channelParam string) {
	_, err := reader.Read()
	if err != nil {
		PrintlnError("read line error, %v", err)
		return
	}

	for {
		csvData, err := reader.Read()
		if err == io.EOF {
			return
		}
		if err != nil {
			PrintlnError("read line error, %v", err)
			continue
		}
		if beforeTime != 0 {
			logTime, err := strconv.ParseInt(csvData[SN], 10, 64)
			if err != nil {
				PrintlnError("parse SN error, %s, %v", csvData[SN], err)
				continue
			}
			if logTime > beforeTime {
				continue // 只统计指定时间之前的
			}
		}
		channel := csvData[Channel]
		if channel == "" && channelParam != "" {
			csvData[Channel] = channelParam // 重新设置channel
		}
		DoParse(csvData)
	}
}

func print2Json(csvData []string) {
	log, ret := parseLine2Json(csvData)
	if ret {
		retJson, err := json.Marshal(log)
		if err != nil {
			PrintlnError("%v", err)
		} else {
			fmt.Println(string(retJson))
		}
	}
}

type FuncBuildPayInfo func([]string) PayInfo

var BuildPayInfo FuncBuildPayInfo

func parseLine2Json(dataArray []string) (logStruct, bool) {
	logTime, err := strconv.ParseInt(dataArray[SN], 10, 64)
	if err != nil {
		PrintlnError("parse receiveTimestamp error, %s, %v", dataArray[ReceiveTimestamp], err)
		return logStruct{}, false
	}
	nt := time.Unix(0, logTime)
	mst := nt.Format("2006-01-02T15:04:05.000Z")
	utc8 := nt.UTC().Add(8 * time.Hour)
	utc8st := utc8.Format("2006-01-02 15:04:05")

	dbAccount, err := db.ParseAccount(dataArray[Uid])
	if err != nil {
		PrintlnError("parse account id error, %s, %v", dataArray[Uid], err)
		return logStruct{}, false
	}

	logPay := logStruct{
		Level:     "Error",
		Logtime:   logTime,
		Timestamp: mst,
		Utc8:      utc8st,
		Type_name: dataArray[Platform],
		Gid:       dbAccount.GameId,
		Sid:       dbAccount.ShardId,
		Userid:    dbAccount.UserId.UUID.String(),
		Accountid: dataArray[Uid],
		Avatar:    "0",
		Channel:   dataArray[Channel],
		Info:      BuildPayInfo(dataArray),
		Extra:     "[BI]",
	}
	return logPay, true
}

func PrintlnError(format string, a ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintf(format, a))
	os.Stderr.WriteString("\n")
}

func CalcHc(hcBuy, hcGive string) string {
	hcBuyInt, err := strconv.ParseInt(hcBuy, 10, 32)
	if err != nil {
		hcBuyInt = 0
	}
	hcGiveInt, err := strconv.ParseInt(hcGive, 10, 32)
	if err != nil {
		hcGiveInt = 0
	}
	return fmt.Sprintf("%d", hcBuyInt+hcGiveInt)
}

func FmtPayTime(payTimeStamp string) string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	paySt, err := strconv.ParseInt(payTimeStamp, 10, 64)
	if err != nil {
		PrintlnError("convert paytimestamp err", payTimeStamp)
		return ""
	}
	return time.Unix(paySt, 0).In(loc).Format("2006-01-02 15:04:05")
}
