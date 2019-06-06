package game

import (
	"fmt"

	"strconv"

	"sync"

	"strings"

	"bufio"
	"encoding/json"
	"io"
	"os"

	"time"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	ver "vcs.taiyouxi.net/platform/planx/version"
)

//Config GameServer配置
//TODO: YZH game.Config 这里的很多配置都是Planx公有库下面根本用不到的

type Config struct {
	Listen        string `toml:"listen"`
	RedisDNSValid bool   `toml:"redis_dns_valid"`
	Redis         string `toml:"redis"`
	RedisAuth     string `toml:"redis_auth"`
	RedisDB       int    `toml:"redis_db"`
	ShardId       []uint `toml:"shard_id"`
	MergeRel      []uint `toml:"merge_rel"`

	ServiceIP         string `toml:"service_ip"`
	ServicePort       int    `toml:"service_port"`
	ServicePublicIP   string `toml:"service_public_ip"`
	ServicePublicPort int    `toml:"service_public_port"`

	Lang    string `toml:"lang"`
	RunMode string `toml:"run_mode"`

	MailDBDriver string `toml:"MailDBDriver"`
	MailMongoUrl string `toml:"MailMongoUrl"`
	MailDBName   string `toml:"MailDBName"`

	RedeemCodeMongoUrl string `toml:"RedeemCodeMongoUrl"`
	RedeemCodeDriver   string `toml:"RedeemCodeDriver"`

	PayDBDriver      string `toml:"PayDBDriver"`
	PayMongoUrl      string `toml:"PayMongoUrl"`
	PayAndroidDBName string `toml:"PayAndroidDBName"`
	PayIOSDBName     string `toml:"PayIOSDBName"`

	AWS_Region    string `toml:"AWS_Region"`
	AWS_AccessKey string `toml:"AWS_AccessKey"`
	AWS_SecretKey string `toml:"AWS_SecretKey"`

	AWS_InitialInterval int64   `toml:"AWS_InitialInterval"`
	AWS_Multiplier      float64 `toml:"AWS_Multiplier"`
	AWS_MaxElapsedTime  int64   `toml:"AWS_MaxElapsedTime"`

	TimeLocal            string   `toml:"time_local"`
	ServerLogicStartTime []string `toml:"server_start_time"`

	EtcdEndPoint []string `toml:"etcd_endpoint"`
	EtcdRoot     string   `toml:"etcd_root"`

	DiscoverRedis           string `toml:"discover_redis"`
	DiscoverRedisDB         uint32 `toml:"discover_redis_db"`
	DiscoverRedisAuth       string `toml:"discover_redis_Auth"`
	DiscoverKeyEtcdPath     string `toml:"discover_key_etcd_path"`
	DiscoverWatcherInterval uint32 `toml:"discover_watcher_interval"`
	CrossServiceConnPoolMax uint32 `toml:"crossservice_connpool_max"`

	BroadCastUrl   string `toml:"broadCast_url"`
	ChatAuthUrl    string `toml:"chatAuth_url"`
	SysNoticeValid int    `toml:"sysnotice_valid"`

	AuthNotifyNewRoleUrl string `toml:"auth_notifyNewRole_url"`

	LimitValid       bool   `toml:"limit_valid"`
	RateLimitUrlInfo string `toml:"ratelimit_url_info"`
	AdminIPs         []InternalIP_str

	DebugAccountTimeValid bool `toml:"debug_account_time_valid"`

	AntiCheatValid bool   `toml:"anticheat_valid"`
	BanUrl         string `toml:"ban_url"`
	BanTime        int64  `toml:"ban_time"`
	IsRegUrl       string `toml:"auth_isreg_url"`
	AntiCheat      []AntiCheat

	ListenPostAddr []string `toml:"listen_post_addr"`

	HotDataUrl string `toml:"hot_data_url"`
	HotS3Path  string `toml:"hot_s3_path"`

	RankReloadUrl string `toml:"rank_reload_url"`

	HourLogValid bool `toml:"hour_log_valid"`

	NewRedisPool bool `toml:"new_redisPool"`

	CloudDbDriver string `toml:"cloud_db_driver"`
	OSSEndpoint   string `toml:"oss_endpoint"`
	OSSAccessId   string `toml:"oss_access_id"`
	OSSAccessKey  string `toml:"oss_access_key"`

	// 不要再toml里出现，会覆写 TODO 从Config结构中删除，单独建立struct
	Gid             int                    // 本gamex所在的gid，从etcd抓取得来；若得到时小于等于0，表示无效（0为开发用）
	ShardShowState  string                 // 服务器显示状态
	DataVer         map[string]int32       // 全局数据版本，由modules/data_ver定时更新; 0无效
	BundleVer       map[string]int32       // 全局bundle版本，由modules/bundle_ver定时更新; 0无效
	DataMin         map[string]int32       // 全局数据版本，由modules/data_ver定时更新; 0无效
	BundleMin       map[string]int32       // 全局bundle版本，由modules/bundle_ver定时更新; 0无效
	ActValid        map[uint]*ActValidInfo // shard -> 活动开关
	HotDataVerC     int                    // 热更数据版本号, 大于0有效
	HotDataVerC_Suc string                 // 热更成功后的热更数据版本号
	LocalDebug      bool                   // 是否本地调试
}

//type Hot_Value int
const ActCount = 50

type AntiCheat struct {
	IsCheck  bool  `toml:"is_check"`
	IsRecord bool  `toml:"is_rec"`
	IsBan    bool  `toml:"is_ban"`
	ParamInt int64 `toml:"param_int"`
}

type InternalIP_str struct {
	From string `toml:"from"`
	To   string `toml:"to"`
}

var (
	//RunModeDev 开发模式
	RunModeDev = "dev"
	// RunModeTest 测试模式
	RunModeTest = "test"
	//RunModeProd 生产模式
	RunModeProd = "prod"

	hot_update_mutex sync.RWMutex
)

//IsRunModeProd 是否当前处于生产模式
func (c *Config) IsRunModeProd() bool {
	return RunModeProd == c.RunMode
}

func (c *Config) IsRunModeDev() bool {
	return RunModeDev == c.RunMode
}

var (
	yyb_merge_shard = map[uint]uint{
		4002: 4001,
		4004: 4003,
		4010: 4005,
		4016: 4006,
		4014: 4007,
		4009: 4008,
		4012: 4011,
		4015: 4013,
		4018: 4017,
	}
)

// 考虑合服情况下的shardid获取
func (c *Config) GetShardIdByMerge(shard uint) uint {
	if len(c.MergeRel) < 2 {
		// 临时修复 by zhangzhen
		rshard, ok := yyb_merge_shard[shard]
		if ok {
			return rshard
		}
		return shard
	}
	for _, sid := range c.MergeRel {
		if sid == shard {
			return c.MergeRel[0]
		}
	}
	return shard
}

//Cfg game server 配置
var Cfg Config

func (c *Config) GetHotActValidData(sid uint, actType int) bool {
	sid = c.GetShardIdByMerge(sid)
	hot_update_mutex.RLock()
	defer hot_update_mutex.RUnlock()
	actV, ok := c.ActValid[sid]
	if ok {
		if actType < len(actV.v) {
			return actV.v[actType]
		}
		return true
	}
	return false
}

func (c *Config) GetHotCfgData(ver string, sid uint) (dataVer, bundleVer int32, actValid []bool, dataMin int32, bundleMin int32) {
	sid = c.GetShardIdByMerge(sid)
	hot_update_mutex.RLock()
	defer hot_update_mutex.RUnlock()
	if c.DataVer != nil {
		dataVer = c.DataVer[ver]
	}
	if c.BundleVer != nil {
		bundleVer = c.BundleVer[ver]
	}
	if c.DataMin != nil {
		dataMin = c.DataMin[ver]
	}
	if c.BundleMin != nil {
		bundleMin = c.BundleMin[ver]
	}
	actV, ok := c.ActValid[sid]
	if ok {
		if len(actV.v) < ActCount {
			act_V := make([]bool, ActCount, ActCount)
			for i := 0; i < ActCount; i++ {
				act_V[i] = true
			}
			for i, data := range actV.v {
				act_V[i] = data
			}
			return dataVer, bundleVer, act_V, dataMin, bundleMin
		}
		actValid = actV.v
	}

	return dataVer, bundleVer, actValid, dataMin, bundleMin
}

func (c *Config) UpdateDataVer(ver string, data_ver int32) {
	ok := false
	var dv int32
	hot_update_mutex.RLock()
	if c.DataVer != nil {
		dv, ok = c.DataVer[ver]
	}
	hot_update_mutex.RUnlock()
	if !ok || dv != data_ver {
		hot_update_mutex.Lock()
		defer hot_update_mutex.Unlock()
		if c.DataVer == nil {
			c.DataVer = make(map[string]int32, 10)
		}
		dv, ok = c.DataVer[ver]
		if !ok || dv != data_ver {
			c.DataVer[ver] = data_ver
		}
	}
}

func (c *Config) UpdateBundleVer(ver string, bundle_ver int32) {
	ok := false
	var dv int32
	hot_update_mutex.RLock()
	if c.BundleVer != nil {
		dv, ok = c.BundleVer[ver]
	}
	hot_update_mutex.RUnlock()
	if !ok || dv != bundle_ver {
		hot_update_mutex.Lock()
		defer hot_update_mutex.Unlock()
		if c.BundleVer == nil {
			c.BundleVer = make(map[string]int32, 10)
		}
		dv, ok = c.BundleVer[ver]
		if !ok || dv != bundle_ver {
			c.BundleVer[ver] = bundle_ver
		}
	}
}

func (c *Config) UpdateDataMin(ver string, data_ver int32) {
	ok := false
	var dv int32
	hot_update_mutex.RLock()
	if c.DataMin != nil {
		dv, ok = c.DataMin[ver]
	}
	hot_update_mutex.RUnlock()
	if !ok || dv != data_ver {
		hot_update_mutex.Lock()
		defer hot_update_mutex.Unlock()
		if c.DataMin == nil {
			c.DataMin = make(map[string]int32, 10)
		}
		dv, ok = c.DataMin[ver]
		if !ok || dv != data_ver {
			c.DataMin[ver] = data_ver
		}
	}
}

func (c *Config) UpdateBundleMin(ver string, data_ver int32) {
	ok := false
	var dv int32
	hot_update_mutex.RLock()
	if c.BundleMin != nil {
		dv, ok = c.BundleMin[ver]
	}
	hot_update_mutex.RUnlock()
	if !ok || dv != data_ver {
		hot_update_mutex.Lock()
		defer hot_update_mutex.Unlock()
		if c.BundleMin == nil {
			c.BundleMin = make(map[string]int32, 10)
		}
		dv, ok = c.BundleMin[ver]
		if !ok || dv != data_ver {
			c.BundleMin[ver] = data_ver
		}
	}
}

func (c *Config) UpdateActValid(sid uint, act_valid string) {
	if act_valid == "" {
		return
	}
	of := strings.Split(act_valid, ",")
	av := make([]bool, len(of))
	for i, a := range of {
		ai, err := strconv.Atoi(a)
		if err != nil {
			logs.Error("UpdateActValid err %s", act_valid)
			return
		}
		if ai > 0 {
			av[i] = true
		} else {
			av[i] = false
		}
	}
	// 更新
	ok := false
	var old *ActValidInfo
	hot_update_mutex.RLock()
	if c.ActValid != nil {
		old, ok = c.ActValid[sid]
	}
	hot_update_mutex.RUnlock()
	if !ok || old.orig != act_valid {
		hot_update_mutex.Lock()
		defer hot_update_mutex.Unlock()
		if c.ActValid == nil {
			c.ActValid = make(map[uint]*ActValidInfo, ActCount)
		}
		c.ActValid[sid] = &ActValidInfo{act_valid, av}
	}
}

func (c *Config) GetGid() bool {
	_ss := make([]uint, 0, 4)
	_ss = append(_ss, c.ShardId...)
	_ss = append(_ss, c.MergeRel...)
	for _, shardId := range _ss {
		sid := shardId
		gid, err := etcd.FindServerGidBySid(c.EtcdRoot+"/", sid)
		if err != nil {
			logs.Error("etcd GetServerGidSid err %s or not found", err.Error())
			return false
		} else if gid < 0 {
			logs.Error("etcd GetServerGidSid sid %d not found gid", sid)
			return false
		} else if c.Gid > 0 && c.Gid != gid {
			logs.Error("etcd GetServerGidSid shard %s bind to different gid %d %d", c.ShardId, c.Gid, gid)
			return false
		} else {
			c.Gid = gid
		}
	}
	logs.Warn("GetGid %d", c.Gid)
	return true
}

// SyncInfo2Etcdy一些GM工具需要的信息,需要写到ETCD然后在GMtool里面抓取
//GMtools代码对应位置func (c *CfgByServer) LoadOneServerFromEtcd
func (c *Config) SyncInfo2Etcd() bool {
	_ss := make([]uint, 0, 4)
	_ss = append(_ss, c.ShardId...)
	_ss = append(_ss, c.MergeRel...)
	for i, shardId := range _ss {
		sid := shardId
		gid := c.Gid
		// get show state, 出错就是空字符串
		c.ShardShowState, _ = etcd.Get(fmt.Sprintf("%s/%d/%d/%s", c.EtcdRoot, gid, sid, etcd.KeyShowState))
		logs.Info("from etcd gid %d showstate %s", gid, c.ShardShowState)

		key_parent := fmt.Sprintf("%s/%d/%d/gm/", c.EtcdRoot, gid, sid)
		// broadCast_url
		if err := etcd.Set(key_parent+"broadCast_url", c.BroadCastUrl, 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"broadCast_url", err)
			return false
		}
		// MailDBName
		if err := etcd.Set(key_parent+"MailDBName", c.MailDBName, 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"MailDBName", err)
			return false
		}
		// MailDBDriver
		if err := etcd.Set(key_parent+"MailDBDriver", c.MailDBDriver, 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"MailDBDriver", err)
			return false
		}
		// MailMongoUrl
		if err := etcd.Set(key_parent+"MailMongoUrl", c.MailMongoUrl, 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"MailMongoUrl", err)
			return false
		}
		// redis
		if err := etcd.Set(key_parent+"redis", c.Redis, 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"redis", err)
			return false
		}
		// redis_db
		if err := etcd.Set(key_parent+"redis_db", fmt.Sprintf("%d", c.RedisDB), 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"redis_db", err)
			return false
		}
		// redis_db_pwd
		if err := etcd.Set(key_parent+"redis_db_auth", c.RedisAuth, 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"redis_db", err)
			return false
		}
		// run_mode
		if err := etcd.Set(key_parent+"run_mode", c.RunMode, 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"run_mode", err)
			return false
		}
		// merge shard
		if err := etcd.Set(key_parent+etcd.KeyMergedShard, fmt.Sprintf("%d", c.MergeRel[0]), 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+etcd.KeyMergedShard, err)
			return false
		}
		// KeyShardVersion
		if err := etcd.Set(key_parent+etcd.KeyShardVersion, ver.VerC(), 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+etcd.KeyShardVersion, err)
			return false
		}
		// auth_ip_addr
		auth_addr := c.BanUrl
		auth_addr = auth_addr[len("http://"):]
		ss := strings.Split(auth_addr, "/")
		auth_addr = ss[0]
		if err := etcd.Set(key_parent+"auth_ip_addr", auth_addr, 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"auth_ip_addr", err)
			return false
		}
		// hotdatapart
		listenPostAddr := c.ListenPostAddr[0]
		if i < len(c.ListenPostAddr) {
			listenPostAddr = c.ListenPostAddr[i]
		}
		if err := etcd.Set(key_parent+"hotdataurl",
			fmt.Sprintf("http://%s/%s", listenPostAddr, c.HotDataUrl), 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"hotdataurl", err)
			return false
		}
		if err := etcd.Set(key_parent+"rank_reload_url",
			fmt.Sprintf("http://%s/%s", listenPostAddr, c.RankReloadUrl), 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+"rank_reload_url", err)
			return false
		}
		// lauch time
		if err := etcd.Set(fmt.Sprintf("%s/%d/%d/", c.EtcdRoot, gid, sid)+
			etcd.KeyServerLaunchTime, fmt.Sprintf("%d", time.Now().Unix()), 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+etcd.KeyServerLaunchTime, err)
			return false
		}
		// server use time
		if err := etcd.Set(fmt.Sprintf("%s/%d/%d/", c.EtcdRoot, gid, sid)+
			etcd.KeyServerUsedStartTime, fmt.Sprintf("%d", ServerStartTime(c.ShardId[0])), 0); err != nil {
			logs.Error("set etcd key %s err %s", key_parent+etcd.KeyServerUsedStartTime, err)
			return false
		}
	}
	return true
}

func (c *Config) GetInfoFromEtcd() bool {
	shard2StartTime := make(map[uint]string, len(c.ShardId))
	for _, shardId := range c.ShardId {
		sid := shardId
		key := fmt.Sprintf("%s/%d/%d/%s", c.EtcdRoot, c.Gid, sid, etcd.KeyServerStartTime)
		startTime, _ := etcd.Get(key)
		if startTime != "" {
			shard2StartTime[uint(sid)] = startTime
		} else {
			logs.Error("GetInfoFromEtcd fail key %s", key)
		}
	}
	util.SetServerStartTimes(shard2StartTime)
	return true
}

func (c *Config) IsHotDataValid() bool {
	return c.HotDataVerC > 0
}

func (c *Config) GetHotDataVerC() string {
	return fmt.Sprintf("%d", c.HotDataVerC)
}

func (c *Config) HotDataS3Bucket() string {
	return c.HotS3Path
}

var (
	GidCfg      GidConfig
	Gid2Channel map[int]string
)

type GidInfo struct {
	Gid     int
	Channel string
}

type GidConfig struct {
	GidInfo []GidInfo
}

func (g *GidConfig) Init() {
	Gid2Channel = make(map[int]string, len(g.GidInfo))
	for _, info := range g.GidInfo {
		Gid2Channel[info.Gid] = info.Channel
	}
}

type ActValidInfo struct {
	orig string
	v    []bool
}

type VivoConfig struct {
}

// pay feed back 充值反钻
type PayFeedBackRec struct {
	Uid         string
	AccountName string
	Money       float64
	FirstPay    float64
	SecondPay   float64
}

const PayFeedBackFile = "payfeedback"

var (
	payFeedBackMap map[string]PayFeedBackRec
)

func InitPayFeedBack() {
	fp := config.NewConfigPath(PayFeedBackFile)
	if !util.FileExists(fp) {
		return
	}
	file, err := os.OpenFile(fp, os.O_RDONLY, 0777)
	if err != nil {
		logs.Error("InitPayFeedBack OpenFile err %s", err.Error())
		return
	}
	payFeedBackMap = make(map[string]PayFeedBackRec, 64)
	bf := bufio.NewReaderSize(file, 1024)
	for {
		line, err := bf.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				logs.Error("InitPayFeedBack ReadString err %s", err.Error())
			}
			break
		}
		r := &PayFeedBackRec{}
		if err := json.Unmarshal([]byte(line), r); err != nil {
			logs.Error("InitPayFeedBack json.Unmarshal %s", err.Error())
			continue
		}
		payFeedBackMap[r.Uid] = *r
	}
	logs.Debug("InitPayFeedBack success %v", payFeedBackMap)
}

func GetPayFeedBackByUid(uid string) (bool, float64, float64, float64) {
	if payFeedBackMap != nil {
		if v, ok := payFeedBackMap[uid]; ok {
			return true, v.FirstPay, v.SecondPay, v.Money
		}
	}
	return false, 0.0, 0.0, 0.0
}
