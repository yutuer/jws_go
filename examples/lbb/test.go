package main

import (
	"fmt"
	"crypto/md5"
	//"encoding/base64"
	"net/url"
	"encoding/hex"
	"encoding/json"
	"time"
)

var a = []string{"cp_uid","server_id","role_id","reward_list","reward_name","description","time"}

//data := "cp_uid=233"
func CalSign(data string) (string,error) {
	n,_ := url.QueryUnescape(data)
	//n, err := base64.URLEncoding.DecodeString(data)
	//if err != nil {
	//	fmt.Println(err)
	//	return "", err
	//}
	fmt.Println(n)
	dataKey := n + "a5ced306175ff1deaff676da872c05c5"

	fmt.Println(dataKey)

	_md5 := md5.New()
	_md5.Write([]byte(dataKey))
	ret := _md5.Sum(nil)

	fmt.Println("####")
	fmt.Println(hex.EncodeToString(ret))
	return string(ret),nil
}

type GiftInfomation struct {
	CpUid       int        `json:"cp_uid"`
	ServerId    int        `json:"server_id"`
	RoleId      string     `json:"role_id"`
	ItemInfo    []ItemInfo `json:"item_info"`
	RewardName  string     `json:"reward_name"`
	Description string     `json:"description"`
	Time        string     `json:"time"`
}

type ItemInfo struct {
	PropID  string `json:"propid"`
	PropNum int    `json:"pronum"`
}


func CalJson(teststring string)  {
	fmt.Println(teststring)
	info := GiftInfomation{
		CpUid:       1,
		ServerId:    2,
		RoleId:      "roleId",
		RewardName:  "rewardName",
		Description: "description",
		Time:        "time",
	}

	var dat map[string]interface{}
	json.Unmarshal([]byte(teststring), &dat)
	info.ItemInfo = make([]ItemInfo,0)
	for v,x := range dat{
		info.ItemInfo = append(info.ItemInfo,ItemInfo{
			PropID:v,
			PropNum:int(x.(float64)),
		})
	}
	fmt.Println(info)
}

func formatDate(nowt int64) string {
	tm := time.Unix(nowt, 0)
	year, month, day := tm.Date()
	fmt.Println(fmt.Sprintf("%d-%d-%d", year, month, day))
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}


func main() {
	//CalSign("server_id=1:16&role_id=1:16:91f4b475-616c-41c2-8830-09b00c84681a&reward_list={"VI_HC":100}&reward_name=SDKTEST&description=SDKTESTTEST&time=1000000000")
	//CalJson("{\"VI_HC\":56}")
	formatDate(1504108800)
}

