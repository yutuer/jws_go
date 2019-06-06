package gin_core

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/config"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/encode"
)

const QQ_Wechat_Both = 3

const (
	Shard_Ret_OK = iota
	Shard_Ret_Ts_TimeOut
	Shard_Ret_Sig_NotMatch
)

const (
	Shard_Ret_Json_Error = iota + 100
	Shard_Ret_Get_Shard
)

var RetMsgMap = map[int]string{
	Shard_Ret_OK:           "拉取成功",
	Shard_Ret_Ts_TimeOut:   "请求参数timestamp已过期",
	Shard_Ret_Sig_NotMatch: "参数错误",
	Shard_Ret_Json_Error:   "生成json错误",
	Shard_Ret_Get_Shard:    "获取服务器列表错误",
}

type ShardReq struct {
	Ts    int64  `form:"timestamp"`
	AppId uint32 `form:"appid"`
	Area  string `form:"area"`
	Sig   string `form:"sig"`
}

type ShardResp struct {
	BaseResp
	ShardList []ShardInfo `json:"list"`
}

type ShardInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type serverSort []ShardInfo

func (s serverSort) Len() int {
	return len(s)
}

func (s serverSort) Less(i, j int) bool {
	// 正常的逻辑：按照sid升序排列
	return s[i].Id < s[j].Id
}

func (s serverSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func shardHandle(c *gin.Context) {
	reqData := &ShardReq{}
	c.Bind(reqData)

	if time.Now().Unix() > int64(reqData.Ts+delay_time) {
		logs.Warn("get shard info time out, %d, %v", time.Now().Unix(), reqData)
		c.String(200, getRespJson(Shard_Ret_Ts_TimeOut, nil))
		return
	}

	if match, theirs, mine := isSignRight(c.Request.URL.Query(), reqData.Sig); !match {
		logs.Warn("sig error, theirs = %s, mine = %s", theirs, mine)
		c.String(200, getRespJson(Shard_Ret_Sig_NotMatch, nil))
		return
	}

	shardList, err := LoadFromEtcd(config.CommonConfig.RechargeEtcdRoot)
	if err != nil {
		logs.Warn("load server error, gameId = %d, %v", 0, err)
		c.String(200, getRespJson(Shard_Ret_Get_Shard, nil))
		return
	}

	logs.Debug("get servers %v", shardList)

	c.String(200, getRespJson(Shard_Ret_OK, shardList))
}

func getRespJson(code int, shardList []ShardInfo) string {
	var resp interface{}
	if code != Shard_Ret_OK {
		resp = BaseResp{code, RetMsgMap[code]}
	} else {
		resp = ShardResp{
			BaseResp:  BaseResp{code, RetMsgMap[code]},
			ShardList: shardList,
		}
	}

	retJson, error := json.Marshal(resp)
	if error != nil {
		logs.Warn("marshal json error, %v, %v", resp, error)
		return fmt.Sprintf("{\"ret\":%d,\"msg\":%s}", Shard_Ret_Json_Error, RetMsgMap[Shard_Ret_Json_Error])
	}
	return string(retJson)
}

// 验证sig
func isSignRight(valueMap url.Values, oldSig string) (bool, string, string) {
	args := make(map[string]string, 10)
	for k, v := range valueMap {
		if k == "sig" {
			continue
		}
		args[k] = v[0]
	}
	mySig, err := encode.CalSig(args, config.CommonConfig.GiftUrl, config.CommonConfig.Appkey)
	if err != nil {
		logs.Warn("generate sig error, %v", err)
		return false, "", ""
	}
	return mySig == oldSig, oldSig, mySig
}

func LoadFromEtcd(etcdRoot string) ([]ShardInfo, error) {
	// etcdRoot /a4k/0
	sids, err := etcd.GetAllSubKeys(etcdRoot)
	if err != nil {
		return nil, err
	}

	logs.Debug("load server list from etcd: sids = %v", sids)

	retServers := make([]ShardInfo, 0, len(sids))

	for _, sid := range sids {
		// sid /a4k/0/10
		sidSplit := strings.Split(sid, "/")
		_, err := strconv.Atoi(sidSplit[len(sidSplit)-1])
		if err != nil {
			logs.Info("load server list from etcd: not a server %s", sid)
			continue
		}

		kv, err := etcd.GetAllSubKeyValue(sid)
		if err != nil {
			logs.Warn("load server list from etcd: get key value error, %v", err)
			continue
		}
		sn, _ := kv[etcd.KeySName]
		dn, _ := kv[etcd.KeyDName]
		if sn == "" {
			logs.Warn("load server list from etcd: sn is nil, %s", sid)
			continue
		}
		if dn == "" {
			dn = sidSplit[len(sidSplit)-1]
		}
		retServers = append(retServers, ShardInfo{
			Id:   sidSplit[len(sidSplit)-1],
			Name: dn,
			Type: QQ_Wechat_Both,
		})
	}
	sort.Sort(serverSort(retServers))
	return retServers, nil
}
