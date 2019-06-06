package main

import (
	"encoding/json"

	"fmt"

	"os"

	"bufio"

	"time"

	"strconv"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const url = "http://54.223.94.97:8081"
const serverID = "200:1001"

type commandParams struct {
	Params []string `json:"params" form:"params"`
	Key    string   `json:"key" form:"key"`
}

type commandInfo struct {
	id       string
	order_id string
	money    string
}

type payInfo struct {
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

type wrapper struct {
	AccountID string  `json:"accountid"`
	Info      payInfo `json:"info"`
}

func main() {
	defer logs.Close()
	commandInfo := getVirtualCommandInfoList("203iap.txt")
	//logs.Debug("commandInfo len: %d, commandInfo: %v", len(commandInfo), commandInfo)
	totolMoney := 0.0
	for _, item := range commandInfo {
		err := sendVirtualCommand(os.Args[1], item)
		if err != nil {
			logs.Error("send virtual command err: %", err)
		}
		money, err := strconv.ParseFloat(item.money, 64)
		if err != nil {
			logs.Error("parseFloat err by %v, for %v", err, item.money)
		}
		totolMoney += money
		time.Sleep(time.Second)
	}
	logs.Debug("total money: %f", totolMoney)
}

func getVirtualCommandInfoList(filepath string) []commandInfo {
	f, err := os.Open(filepath)
	if err != nil {
		logs.Error("read file err by %v", err)
		return nil
	}
	ret := make([]commandInfo, 0)
	reader := bufio.NewReader(f)
	for {
		lineData, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		wrapper := wrapper{}
		err = json.Unmarshal(lineData, &wrapper)
		if err != nil {
			break
		}
		ret = append(ret, commandInfo{id: wrapper.AccountID, order_id: wrapper.Info.OrderIdx, money: wrapper.Info.Money})
	}
	return ret
}

func sendVirtualCommand(key string, info commandInfo) error {
	param := commandParams{}
	param.Key = key
	param.Params = []string{info.order_id, info.money, "0"}
	data, err := json.Marshal(param)
	if err != nil {
		return err
	}
	//logs.Debug("serverID: %v", serverID)
	url := fmt.Sprintf("%s/api/v1/command/%s/%s/virtualtrueIAP", url, serverID, info.id)
	rsp, err := util.HttpPost(url, util.JsonPostTyp, data)
	logs.Debug("rsp: %v", string(rsp))
	return err
}
