package config

import (
	"fmt"
	"sync"

	"strconv"
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type CommonConfig struct {
	Runmode         string `toml:"runmode"`
	InnerExperience bool   `toml:"inner_experience"` // 是否体验服
	AuthDBDriver    string `toml:"auth_db_driver"`
	//autorender      string `toml:"autorender"`
	//copyrequestbody string `toml:"copyrequestbody"`
	//EnableDocs      string `toml:"EnableDocs"`

	AuthRedisAddr  string `toml:"auth_redis_addr"`
	AuthRedisDB    string `toml:"auth_redis_db"`
	AuthRedisDBPwd string `toml:"auth_redis_db_pwd"`

	LoginRedisAddr  string `toml:"login_redis_addr"`
	LoginRedisDB    string `toml:"login_redis_db"`
	LoginRedisDBPwd string `toml:"login_redis_db_pwd"`

	//EnableHttpListen string `toml:"EnableHttpListen"`
	EnableHttpTLS string `toml:"EnableHttpTLS"`
	HttpsPort     string `toml:"HttpsPort"`
	HttpCertFile  string `toml:"HttpCertFile"`
	HttpKeyFile   string `toml:"HttpKeyFile"`

	Dynamo_endpoint        string `toml:"dynamo_endpoint"`
	Dynamo_region          string `toml:"dynamo_region"`
	Dynamo_accessKeyID     string `toml:"dynamo_accessKeyID"`
	Dynamo_secretAccessKey string `toml:"dynamo_secretAccessKey"`
	Dynamo_sessionToken    string `toml:"dynamo_sessionToken"`

	Dynamo_NameDevice   string `toml:"dynamo_db_Device"`
	Dynamo_NameName     string `toml:"dynamo_db_Name"`
	Dynamo_NameUserInfo string `toml:"dynamo_db_UserInfo"`
	Dynamo_GM           string `toml:"dynamo_db_GM"`
	//Dynamo_NameDeviceTotal string `toml:"dynamo_db_DeviceTotal"`
	//Dynamo_NameAuthToken string `toml:"dynamo_db_AuthToken"`
	Dynamo_UserShardInfo string `toml:"dynamo_db_UserShardInfo"`

	MongoDBUrl  string `toml:"mongodb_url"`
	MongoDBName string `toml:"mongodb_name"`

	VerUpdate_Valid         bool   `toml:"verupdate_valid"`
	S3_Buckets_VerUpdate    string `toml:"s3_buckets_verupdate"`
	VerUpdate_IsRemote      bool   `toml:"verupdate_is_remote"`
	VerUpdate_FileName      string `toml:"verupdate_filename"`
	VerUpdate_Meta_FileName string `toml:"verupdate_meta_filename"`

	LoginApiAddr     string `toml:"login_api_addr"`
	LoginKickApiAddr string `toml:"login_kick_api_addr"`
	LoginGagApiAddr  string `toml:"login_gag_api_addr"`

	passwd_b64map string `toml:"passwd_b64map"`
	passwd_salt   string `toml:"passwd_salt"`

	Httpport   string `toml:"httpport"`
	GMHttpport string `toml:"gmhttpport"`
	//EnableAdmin   string `toml:"EnableAdmin"`
	//AdminHttpAddr string `toml:"AdminHttpAddr"`
	//AdminHttpPort int    `toml:"AdminHttpPort"`

	SentryDSN string `toml:"SentryDSN"`

	EtcdEndPoint []string `toml:"etcd_endpoint"`
	EtcdRoot     string   `toml:"etcd_root"`
	Gids         []string `toml:"gids"`

	StoreDriver  string `toml:"store_driver"`
	OSSEndPoint  string `toml:"oss_endpoint"`
	OSSAccessId  string `toml:"oss_access_id"`
	OSSAccessKey string `toml:"oss_access_key"`
	OSSBucket    string `toml:"oss_bucket"`
}

type SdkHeroConfig struct {
	Url          string `toml:"url"`
	UrlToken     string `toml:"url_token"`
	UrlUserinfo  string `toml:"url_userinfo"`
	AppKey       string `toml:"appKey"`
	ProductId    string `toml:"productId"`
	ProjectId    string `toml:"projectId"`
	ClientSecret string `toml:"client_secret"`
	ServerId     string `toml:"serverId"`
}

type SdkQuickConfig struct {
	Url                       string `toml:"url"`
	AndroidProductCode        string `toml:"android_product_code"`
	AndroidProductCode_MuBao1 string `toml:"android_product_code_mubao1"`
	IOSProductCode            string `toml:"ios_product_code"`
}

type SdkVivoConfig struct {
	Url string `toml:"url"`
}

type SdkEnjoyConfig struct {
	Url     string `toml:"url"`
	TestUrl string `toml:"test_url"`
}
type SdkVNConfig struct {
	Url             string `toml:"url"`
	UrlToken        string `toml:"url_token"`
	UrlUserinfo     string `toml:"url_userinfo"`
	AppKey          string `toml:"appKey"`
	ProductId       string `toml:"productId"`
	ProjectId       string `toml:"projectId"`
	ClientSecret    string `toml:"client_secret"`
	ServerId        string `toml:"serverId"`
	IOSAppKey       string `toml:"ios_appKey"`
	IOSProductId    string `toml:"ios_productId"`
	IOSProjectId    string `toml:"ios_projectId"`
	IOSClientSecret string `toml:"ios_client_secret"`
	IOSServerId     string `toml:"ios_serverId"`
}
type Sdk6wavesConfig struct {
	Url string `toml:"url"`
	TestUrl string `toml:"test_url"`
}

type SuperUidConfig struct {
	Uid []string
}

type GonggaoInfo struct {
	WhitelistPwd      string
	MaintainStartTime int64
	MaintainEndTime   int64
}

var (
	Cfg          CommonConfig
	SuperUidsCfg SuperUidConfig
	HeroSdkCfg   SdkHeroConfig
	QuickSdkCfg  SdkQuickConfig
	VivoSdkCfg   SdkVivoConfig
	EnjoySdkCfg  SdkEnjoyConfig
	VNSdkCfg     SdkVNConfig
	WavesSdkCfg  Sdk6wavesConfig

	GonggaoInfos map[string]map[string]GonggaoInfo // gid->ver->info
	endPointMux  sync.RWMutex
)

const (
	Spec_Header         = "TYX-Request-Id"
	Spec_Header_Content = "f88a78a83463f1e03bbf0acb3883aefb"
)

func (c *CommonConfig) IsRunModeProdAndTest() bool {
	return c.Runmode == "prod" || c.Runmode == "test"
}

func (c *CommonConfig) IsRunModeLocal() bool {
	return c.Runmode == "local"
}

func (c *SuperUidConfig) IsSuperUid(uid string) bool {
	for _, _uid := range c.Uid {
		if uid == _uid {
			return true
		}
	}
	return false
}

func GetGonggao(gid, ver string) GonggaoInfo {
	endPointMux.RLock()
	defer endPointMux.RUnlock()
	vers, ok := GonggaoInfos[gid]
	if ok {
		return vers[ver]
	}
	return GonggaoInfo{}
}

func updateGonggao() {
	gids, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/%s/", Cfg.EtcdRoot, etcd.KeyEndPoint))
	if err != nil {
		logs.Error("updateGonggao getallgid err %s", err.Error())
		return
	}
	infos := make(map[string]map[string]GonggaoInfo, 4)
	for _, gid := range gids {
		pg := strings.Split(gid, "/")
		sgid := pg[len(pg)-1]
		_, err := strconv.Atoi(sgid)
		if err != nil {
			continue
		}

		vers, err := etcd.GetAllSubKeys(gid)
		if err != nil {
			logs.Error("updateGonggao getallsid err %s", err.Error())
			return
		}

		for _, ver := range vers {
			ps := strings.Split(ver, "/")
			if len(ps) > 4 {
				sver := ps[4]
				kv, err := etcd.GetAllSubKeyValue(ver)
				if err != nil {
					continue
				}
				whitepwd, _ := kv[etcd.KeyEndPoint_Whitelistpwd]
				met, _ := kv[etcd.KeyEndPoint_maintain_endtime]
				mst, _ := kv[etcd.KeyEndPoint_maintain_starttime]
				imet, err := strconv.ParseInt(met, 10, 64)
				if err != nil {
					logs.Error("updateGonggao maintain_endtime err %s %s %s %s",
						err.Error(), sgid, sver, met)
					imet = 0
				}
				imst, err := strconv.ParseInt(mst, 10, 64)
				if err != nil {
					logs.Error("updateGonggao maintain_starttime err %s %s %s %s",
						err.Error(), sgid, sver, mst)
					imst = 0
				}
				vers, ok := infos[sgid]
				if !ok {
					vers = make(map[string]GonggaoInfo, 4)
					infos[sgid] = vers
				}
				vers = infos[sgid]
				vers[sver] = GonggaoInfo{
					WhitelistPwd:      whitepwd,
					MaintainStartTime: imst,
					MaintainEndTime:   imet,
				}
				infos[sgid] = vers
			}
		}
	}

	endPointMux.Lock()
	GonggaoInfos = infos
	endPointMux.Unlock()
	logs.Info("updateGonggao %v", infos)
}
