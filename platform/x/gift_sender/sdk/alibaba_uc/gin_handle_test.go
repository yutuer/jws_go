package alibaba_uc

import (
	"encoding/json"
	"fmt"
	"testing"

	"time"

	"os"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/auth/models"
)

func TestAesDecode(t *testing.T) {
	defer logs.Close()

	ret, err := aesDecode("jSzmFUujwtSqfZo7ab8BW2o4sDFrEhPHXBqGnzQg2HsFqJdYZj00QkF5mlA4b+fw8w"+
		"Sy9fnMcMkB8J79D2TX7qpY7xrb26hI2LMYmoQKfQY=",
		"1234567890123456")
	if err != nil {
		fmt.Println(err)
	}
	logs.Debug("aes decode data len: %v", len(ret))
	fmt.Println(ret)
	type GetRoleShardReqData2 struct {
		Platform  int    `json:"platform"`
		AccountID string `json:"accountId"`
		GameID    int    `json:"gameId"`
	}

	reqData := GetRoleShardReqData2{}
	err = json.Unmarshal(ret, &reqData)
	if err != nil {
		logs.Error("json unmarshal err by %v", err)
	}
	logs.Debug("unmarshal json: %v", reqData)

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		logs.Error("json marshal err by %v", err)
	}
	logs.Debug("encode rsp, marshal json data: %v", jsonData)
	encodeData, err := aesEncode(jsonData, "1234567890123456")
	if err != nil {
		logs.Error("json unmarshal err by %v", err)
	}
	logs.Debug("encode data: %v", encodeData)
	sign, err := GenSign("a13cPwhP3rK0JRdJsOPUxWlxHCN/gNP6Nh3wL1vguxwCB0YNiV804+5I0kh7jPxM"+
		"xnm8k19BSK+36qYeBzHm5w==", "202cb962234w4ers2aa", "changyou")
	if err != nil {
		logs.Error("gen sign err by %v", err)
	}
	logs.Debug("md5: %v", sign)
	nowT2 := fmt.Sprintf("%d", time.Now().AddDate(0, 0, 4).Unix())
	nowT := fmt.Sprintf("%d", time.Now().Unix())
	flag := isSameWeekDay(nowT, nowT2)
	logs.Debug("flag: %v", flag)

	flag2 := isSameDay2("2017-06-14", "2017-06-15")
	logs.Debug("flag2: %v", flag2)
	flag3 := isSameDay2("2017-06-14", "2017-06-14")
	logs.Debug("flag2: %v", flag3)
	err := etcd.InitClient(config.CommonConfig.EtcdEndPoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	info, err := models.GetAllShard(1)
	if err != nil {
		logs.Error("get all shard eror")
	}
	logs.Debug("get all shard info: %v", info)
	a := make([]ShardInfo, 0)
	for _, item := range info {
		a = append(a, ShardInfo{
			ServerName: item.DName,
			ServerID:   fmt.Sprintf("%d:%d", item.Gid, item.Sid),
		})
	}
	logs.Info("shard info: %v", a)
}
