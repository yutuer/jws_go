package sdk_shop

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/httplib"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gift_sender/core"
	h5Config "vcs.taiyouxi.net/platform/x/h5_shop/config"
)

const ac_iota = "ac_iota"

var (
	legal_gid = []string{"0", "1", "200", "201", "202", "203"}
)

func init() {
	initErrorCode()
}

func ParseServerID(numberServerID int) (string, bool) {
	str := strconv.Itoa(numberServerID)
	for _, prefix := range legal_gid {
		if strings.HasPrefix(str, prefix) {
			sid := strings.TrimPrefix(str, prefix)
			serverID := prefix + ":" + sid
			logs.Debug("parse server id: %s", serverID)
			return serverID, true
		}
	}
	return "", false
}

func getAcIDFromNumberID(numberID int, gid, sid int) (string, error) {
	conn, err := core.GetAuthRedisConn(uint(gid), uint(sid))
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return "", err
	}
	acID, err := redis.String(conn.Do("HGET", ac_iota, numberID))
	if err != nil {
		return "", err
	}
	if strings.Split(acID, ":")[1] != strconv.Itoa(sid) {
		return "", fmt.Errorf("sid illegal for numberID: %d, gid: %d, sid: %d, acID: %s", acID)
	}
	return acID, nil
}

func GenH5ShopSign(data string, key string) string {
	data = data + "&" + key
	logs.Debug("DesSign: %s", data)
	_md5 := md5.New()
	_md5.Write([]byte(data))
	ret := _md5.Sum(nil)
	hexText := make([]byte, 32)
	hex.Encode(hexText, ret)
	return string(hexText)
}

func TryUnmarshal(bytes []byte, c *gin.Context) []byte {
	fmt.Printf("%s\n", bytes)
	bytLen := len(bytes)
	if bytLen > 52 {
		bytes[1], bytes[33] = bytes[33], bytes[1]
		bytes[10], bytes[42] = bytes[42], bytes[10]
		bytes[18], bytes[50] = bytes[50], bytes[18]
		bytes[19], bytes[51] = bytes[51], bytes[19]
	}
	bsOut, err := base64.StdEncoding.DecodeString(string(bytes))
	if err != nil {
		logs.Error("json: error decoding base64 binary '%s': %v", bsOut, err)
		c.String(200, genRet(RetCode_ErrorArg))
		return nil
	}
	fmt.Printf("TryUnmarshal %s", bsOut)
	return bsOut
}

type RspInfo struct {
	Ret     int    `json:"ret"`
	Diamond string `json:"diamond"`
}

type ProInfo struct {
	Roleid   int `json:"roleid"`
	Serverid int `json:"serverid"`
	DiamNum  int `json:"diamond_num"`
}

func CostDiamond(c *gin.Context) {
	param := make(map[string]string, 3)
	param["data"] = c.PostForm("data")
	//signOrg := c.PostForm("sign")
	param["timestamp"] = c.PostForm("timestamp")
	logs.Debug("receive params %v", param)
	Key := h5Config.CommonConfig.SecretKey
	logs.Debug(Key)

	if sign := GenH5ShopSign("data="+c.PostForm("data")+"&"+
		"timestamp="+c.PostForm("timestamp"), Key); sign != c.PostForm("sign") {
		logs.Debug("illegal sign: %s cmp right sign: %s", sign, c.PostForm("sign"))
		c.String(200, genRet(RetCode_InvalidSign))
		return
	}

	bsOut := TryUnmarshal([]byte(param["data"]), c)
	//进行解密
	logs.Debug("%s\n", bsOut)
	info := ProInfo{}

	err := json.Unmarshal(bsOut, &info)
	if err != nil {
		logs.Error("json: error CostDiamInfo decoding json binary '%s': %v", bsOut, err)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}

	logs.Debug("After all unmarshal:%v", info)

	serverID, ok := ParseServerID(info.Serverid)
	if !ok {
		logs.Error("parse server id error by serverNum: %d", serverID)
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}

	g_sid := strings.Split(serverID, ":")
	if g_sid[0] == "" {
		logs.Debug("putin wrong gid")
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	gid, err := strconv.Atoi(g_sid[0])
	if err != nil {
		logs.Error("putin wrong gid error %v", err)
	}
	if g_sid[1] == "" {
		logs.Debug("putin wrong sid")
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	sid, err := strconv.Atoi(g_sid[1])
	if err != nil {
		logs.Error("putin wrong sid error %v", err)
	}
	logs.Debug("%s/%d/%d/internalip", h5Config.CommonConfig.EtcdRoot, gid, sid)
	_redis_internalip, err := etcd.Get(fmt.Sprintf("%s/%d/%d/internalip",
		h5Config.CommonConfig.EtcdRoot,
		gid, sid))
	if err != nil {
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	redis_ip := strings.Split(_redis_internalip, ":")
	logs.Debug("%s/%d/%d/internalip:%s", h5Config.CommonConfig.EtcdRoot, gid, sid, _redis_internalip)
	_redis_listen_post_port, err := etcd.Get(fmt.Sprintf("%s/%d/%d/listen_post_port",
		h5Config.CommonConfig.EtcdRoot,
		gid, sid))
	if err != nil {
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	logs.Debug("%s/%d/%d/internalip:%s", h5Config.CommonConfig.EtcdRoot, gid, sid, "http://"+redis_ip[0]+":"+_redis_listen_post_port)

	roleId, err := getAcIDFromNumberID(info.Roleid, gid, sid)
	if err != nil {
		logs.Error("Get acid error: %v,  %d, ", info.Roleid, err)

	}

	resp := httplib.Post("http://"+redis_ip[0]+":"+_redis_listen_post_port+"/h5shop/CostDiamondGamex").SetTimeout(5*time.Second, 5*time.Second).
		Param("id", roleId).Param("serverid", serverID).Param("num", strconv.Itoa(info.DiamNum))
	res, err := resp.String()
	logs.Debug("Rsp Info from gamex %s", res)
	tfo := RspInfo{}
	json.Unmarshal([]byte(res), &tfo)
	if err != nil {
		logs.Debug("Get rep from gamex error %v", err)
	}
	num, err := strconv.Atoi(tfo.Diamond)
	if err != nil {
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	ret, err := json.Marshal(Rsp{
		RetCode_Success,
		GetMsg(RetCode_Success),
		tfo.Ret,
		num,
	})
	if err != nil {
		c.String(200, genRet(RetCode_ErrorArg))
		return
	}
	c.String(200, string(ret))
}
