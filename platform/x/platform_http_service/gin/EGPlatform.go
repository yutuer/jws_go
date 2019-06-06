package gin

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"sort"
	"strconv"
	"strings"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/platform_http_service/config"
	platDb "vcs.taiyouxi.net/platform/x/platform_http_service/db"
)

// enjoy game
type EGPlatform struct {
	cfg config.PlatformConfig
}

func (e *EGPlatform) GetConfig() config.PlatformConfig {
	return e.cfg
}

func (e *EGPlatform) RegHandler(gin *gin.Engine) {
	gin.GET("getShardInfo", e.getShardHandler)
	gin.GET("getRoleInfo", e.getRoleInfo)
	gin.GET("getBuyInfo", e.getBuyInfo)
}

func (e *EGPlatform) getShardHandler(c *gin.Context) {
	time := c.Query("time")
	caller := c.Query("caller")
	sign := c.Query("sign")
	mySign := calcMd5Sign(fmt.Sprintf("caller=%s&time=%s%s", caller, time, e.GetConfig().Key))
	if sign != mySign {
		logs.Info("sign=%s, mysign=%s", sign, mySign)
		c.String(200, `{[]}`)
		return
	}

	shardList := getShardInfos()

	logs.Debug("get servers %v", shardList)
	respBytes, err := json.Marshal(shardList)
	if err != nil {
		logs.Error("parse server error, shardlist %v %v", shardList, err)
		c.String(200, `{[]}`)
		return
	}

	c.String(200, string(respBytes))
}

type ShardInfo struct {
	Id   string `json:"serverId"`
	Name string `json:"serverName"`
}

type ShardInfoSort []ShardInfo

func (s ShardInfoSort) Len() int {
	return len(s)
}

func (s ShardInfoSort) Less(i, j int) bool {
	return strings.Compare(s[i].Id, s[j].Id) < 0
}

func (s ShardInfoSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

var (
	shardInfoCache []ShardInfo
	lastUpdateTime int64
)

func getShardInfos() []ShardInfo {
	now := time.Now().Unix()
	if now-lastUpdateTime > 60 {
		lastUpdateTime = now
		var err error = nil
		shardInfoCache, err = getShardInfoFromEtcd(config.CommonConfig.EtcdCfg.EtcdRoot, config.CommonConfig.Gid)
		if err != nil {
			logs.Error("load server error, gameId = %d, %v", 0, err)
		}

	}
	sort.Sort(ShardInfoSort(shardInfoCache))
	return shardInfoCache
}

func getShardInfoFromEtcd(etcdRoot string, gid string) ([]ShardInfo, error) {
	// etcdRoot /a4k/0
	sids, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/%s", etcdRoot, gid))
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
		state, _ := kv[etcd.KeyState]
		if sn == "" || state != etcd.StateOnline {
			logs.Warn("load server list from etcd: sn is nil, %s", sid)
			continue
		}
		if dn == "" {
			dn = sidSplit[len(sidSplit)-1]
		} else {
			// 1.S13-qa1 => 1区 qa1
			namePrefix := strings.Split(dn, ".")[0]
			namePostfix := strings.Split(dn, "-")[1]
			dn = namePrefix + "區 " + namePostfix
		}
		serverId := sidSplit[len(sidSplit)-1]
		retServers = append(retServers, ShardInfo{
			Id:   fmt.Sprintf("%s:%s", config.CommonConfig.Gid, serverId),
			Name: dn,
		})
	}
	return retServers, nil
}

type RoleInfoReq struct {
	CpUid    int    `form:"cpUid"`
	ServerId string `form:"serverId"`
	Time     string `form:"time"`
	Sign     string `form:"sign"`
}

type RoleInfoResp struct {
	CpUid    int    `json:"cpUid"`
	RoleName string `json:"roleName"`
	RoleId   string `json:"roleId"`
}

func (e *EGPlatform) getRoleInfo(c *gin.Context) {
	reqData := &RoleInfoReq{}
	c.Bind(reqData)
	logs.Debug("req role info data, %v", reqData)

	mySign := calcMd5Sign(fmt.Sprintf("cpUid=%d&serverId=%s&time=%s%s",
		reqData.CpUid,
		reqData.ServerId,
		reqData.Time,
		e.GetConfig().Key))

	if mySign != reqData.Sign {
		logs.Info("sign=%s, mysign=%s", reqData.Sign, mySign)
		return
	}

	name1, uid1 := getByChannelId(reqData, "5001")
	name2, uid2 := getByChannelId(reqData, "5002")
	buyInfoArray := make([]RoleInfoResp, 0)
	if name1 != "" {
		buyInfoArray = append(buyInfoArray, RoleInfoResp{
			CpUid:    reqData.CpUid,
			RoleName: name1,
			RoleId:   fmt.Sprintf("%s:%s", reqData.ServerId, uid1),
		})
	}

	if name2 != "" {
		buyInfoArray = append(buyInfoArray, RoleInfoResp{
			CpUid:    reqData.CpUid,
			RoleName: name2,
			RoleId:   fmt.Sprintf("%s:%s", reqData.ServerId, uid2),
		})
	}

	respJson, err := json.Marshal(buyInfoArray)
	if err != nil {
		return
	}

	c.String(200, string(respJson))
}

func getByChannelId(reqData *RoleInfoReq, channelId string) (name string, uid string) {
	uid = getUidByCpUid(strconv.Itoa(reqData.CpUid), channelId)

	logs.Debug("uid %s", uid)
	if uid == "" {
		logs.Warn("uid is nil, %d %s", reqData.CpUid, channelId)
		return "", ""
	}

	acid := fmt.Sprintf("%s:%s", reqData.ServerId, uid)
	logs.Debug("acid %s", acid)

	name = getRoleName(acid)
	if name == "" {
		logs.Warn("role is nil, %d %s", reqData.CpUid, channelId)
		return "", ""
	}
	return
}

func getUidByCpUid(cpUid string, channelId string) string {
	deviceId := buildDeviceId(cpUid, channelId)
	logs.Debug("deviceId %s", deviceId)
	return getDeviceInfo(deviceId)
}

func buildDeviceId(sdkUid, channelId string) string {
	return fmt.Sprintf("%s@%s@%s", sdkUid, channelId, "enjoy.com")
}

func getDeviceInfo(deviceId string) string {
	data, err := platDb.DyDb.GetDb().GetByHashM(config.CommonConfig.DynamoConfig.AWS_DeviceName, deviceId)
	if err != nil {
		logs.Error("get device info err %v", err)
		return ""
	}
	uid, ok := data["user_id"]
	if !ok {
		return ""
	}
	return uid.(string)
}

func getRoleName(acid string) string {
	account, err := db.ParseAccount(acid)
	if err != nil {
		logs.Error("parse account err, %v", err)
		return ""
	}

	redisUrl, redisDb, redisAuth, err := getRedisInfo(account.GameId, account.ShardId)
	if err != nil {
		logs.Error("get redis info err, %v", err)
		return ""
	}
	logs.Debug("get redis info , %s, %s, %s", redisUrl, redisDb, redisAuth)

	driver.SetupRedisByCap(
		redisUrl,
		redisDb,
		redisAuth,
		1, // 这个字段底层函数并不会使用
	)

	defer driver.ShutdownRedis()
	return getRoleNameFromRedis(acid)
}

func getRedisInfo(gameId, shardId uint) (url string, db int, auth string, err error) {
	etcdRoot := config.CommonConfig.EtcdCfg.EtcdRoot
	url, err = etcd.Get(fmt.Sprintf("%s/%d/%d/redis", etcdRoot, gameId, shardId))
	if err != nil {
		return
	}

	_redis_db, err := etcd.Get(fmt.Sprintf("%s/%d/%d/redis_db", etcdRoot, gameId, shardId))
	if err != nil {
		return
	}
	db, err = strconv.Atoi(_redis_db)
	if err != nil {
		return
	}
	auth, err = etcd.Get(fmt.Sprintf("%s/%d/defaults/redis_auth", etcdRoot, gameId))
	if err != nil {
		return
	}
	len_str := len(auth)
	logs.Debug("redis_auth src: %s, %d", auth, len_str)
	// etcd 转义,方便配置,抛弃etcd中redis_auth所配置的空值的引号
	if len_str >= 2 && strings.HasPrefix(auth, `"`) && strings.HasSuffix(auth, `"`) {
		auth = auth[1 : len(auth)-1]
	}
	return
}

func getRoleNameFromRedis(acid string) string {
	tableName := fmt.Sprintf("profile:%s", acid)
	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("Save Error:WSPVP DB Save, cant get redis conn")
		return ""
	}
	name, err := redis.String(db.Do("HGET", tableName, "name"))
	if err != nil {
		logs.Error("get name err, %v", err)
	}

	return name
}

type BuyInfoReq struct {
	CpUid    int    `form:"cpUid"`
	Client   string `form:"client"`
	ServerId string `form:"serverId"`
	RoleId   string `form:"roleId"`
	GoodId   string `form:"goodId"`
	Time     string `form:"time"`
	Sign     string `form:"sign"`
}

type BuyInfoResp struct {
	PayDescription string `json:"payDescription"`
}

func (e *EGPlatform) getBuyInfo(c *gin.Context) {
	reqData := &BuyInfoReq{}
	c.Bind(reqData)
	logs.Debug("req buy info data, %v", reqData)

	mySign := calcMd5Sign(fmt.Sprintf("client=%s&cpUid=%d&goodId=%s&roleId=%s&serverId=%s&time=%s%s",
		reqData.Client,
		reqData.CpUid,
		reqData.GoodId,
		reqData.RoleId,
		reqData.ServerId,
		reqData.Time,
		e.GetConfig().Key))

	if mySign != reqData.Sign {
		logs.Info("sign=%s, mysign=%s", reqData.Sign, mySign)
		return
	}
	payDescription := getPayDescription(reqData.RoleId, reqData.GoodId, reqData.Time)
	if payDescription == "" {
		logs.Warn("pay description is nil, %s", reqData.GoodId)
		c.String(500, "{}")
		return
	}
	c.String(200, fmt.Sprintf(`{"payDescription":"%s"}`, payDescription))
}

func getPayDescription(acid, goodId string, time string) string {
	iapIndex := gamedata.GetIAPIdxByID(goodId)
	if iapIndex == 0 {
		logs.Warn("not find iap %s", goodId)
		return ""
	}
	iapOrder := fmt.Sprintf("%s:%d:%d", acid, 0, iapIndex)
	marketVersion := config.CommonConfig.MarketVersion
	return fmt.Sprintf("%s|%s|%s|%s", time, goodId, marketVersion, iapOrder)
}

func calcMd5Sign(source string) string {
	logs.Debug("souce string %s", source)
	_md5 := md5.New()
	_md5.Write([]byte(source))
	ret := _md5.Sum(nil)
	hexText := make([]byte, 32)
	hex.Encode(hexText, ret)
	return string(hexText)
}
