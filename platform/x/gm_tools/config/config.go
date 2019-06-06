package config

import (
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type CommonConfig struct {
	Runmode string `toml:"runmode"`
	Url     string `toml:"Url"`

	MailDBDriver string `toml:"MailDBDriver"`
	MailMongoUrl string `toml:"MailMongoUrl"`
	MailDBName   string `toml:"MailDBName"`

	AWS_Region    string `toml:"AWS_Region"`
	AWS_AccessKey string `toml:"AWS_AccessKey"`
	AWS_SecretKey string `toml:"AWS_SecretKey"`

	S3_Buckets_VerUpdate    string `toml:"s3_buckets_verupdate"`
	VerUpdate_FileName      string `toml:"verupdate_filename"`
	VerUpdate_Meta_FileName string `toml:"verupdate_meta_filename"`
	S3GongGao_Region        string `toml:"s3_gonggao_region"`
	S3_Buckets_GongGao      string `toml:"s3_buckets_gonggao"`

	//AWS_InitialInterval int64   `toml:"AWS_InitialInterval"`
	//AWS_Multiplier      float64 `toml:"AWS_Multiplier"`
	//AWS_MaxElapsedTime  int64   `toml:"AWS_MaxElapsedTime"`

	AuthDBDriver         string `toml:"auth_db_driver"`
	AuthMongoDBUrl       string `toml:"auth_mongodb_url"`
	AuthMongoDBName      string `toml:"auth_mongodb_name"`
	AuthDynamoDBDevice   string `toml:"auth_dynamo_db_Device"`
	AuthDynamoDBName     string `toml:"auth_dynamo_db_Name"`
	AuthDynamoDBUserInfo string `toml:"auth_dynamo_db_UserInfo"`

	PayDBDriver      string `toml:"PayDBDriver"`
	PayMongoUrl      string `toml:"PayMongoUrl"`
	PayAndroidDBName string `toml:"PayAndroidxDBName"`
	PayIOSDBName     string `toml:"PayIOSDBName"`

	DBStoreFile string `toml:"DBStoreFile"`

	EtcdEndpoint []string `toml:"etcd_endpoint"`
	//GmToolsCfg          string   `toml:"GmToolsCfg"`
	GameSerEtcdRoot     string   `toml:"GameSerEtcdRoot"`
	GidFilter           []string `toml:"GidFilter"`
	GameMachineEtcdRoot string   `toml:"GameMachineEtcdRoot"`

	RedisName        []string `toml:"RedisName"`
	RedisAddress     []string `toml:"RedisAddress"`
	RedisDB          []int    `toml:"RedisDB"`
	RedisRankAddress []string `toml:"RedisRankAddress"`
	RedisRankDB      []int    `toml:"RedisRankDB"`
	RedisAuth        []string `toml:"RedisAuth"`
	ServerName       []string `toml:"ServerName"`
	MailDBNames      []string `toml:"MailDBNames"`
	MailDBDrivers    []string `toml:"MailDBDrivers"`
	MailMongoUrls    []string `toml:"MailMongoUrls"`
	SysRollNoticeUrl []string `toml:"SysRollNoticeUrl"`
	AuthApi          []string `toml:"AuthApi"`
	AuthGagApi       []string `toml:"AuthGagApi"`
	ServerHotDataUrl []string `toml:"ServerHotDataUrl"`
	RankReloadUrl    []string `toml:"RankReloadUrl"`
	MergedShard      []string `toml:"MergedShard"`
	ElasticUrl       string   `toml:"ElasticUrl"`
	ElasticIndex     string   `toml:"ElasticIndex"`

	FtpAddress string `toml:"FtpAddress"`
	FtpUser    string `toml:"FtpUser"`
	FtpPasswd  string `toml:"FtpPasswd"`

	QiNiuAccKey string `toml:"QiNiuAccKey"`
	QiNiuSecKey string `toml:"QiNiuSecKey"`

	KS3Domain    string `toml:"ks3_domain"`
	KS3Bucket    string `toml:"ks3_bucket"`
	KS3AccessKey string `toml:"ks3_access_key"`
	KS3SecretKey string `toml:"ks3_secret_key"`

	StoreDriver      string `toml:"store_driver"`
	OSSEndpoint      string `toml:"oss_endpoint"`
	OSSAccessId      string `toml:"oss_access_id"`
	OSSAccessKey     string `toml:"oss_access_key"`
	OSSChannelBucket string `toml:"oss_channle_bucket"`
	OSSPublicBucket  string `toml:"oss_public_bukcet"`

	NoticeWatchKey string `toml:"notice_watch_key"`

	TimeLocal string `toml:TimeLocal`
}

type ServerCfg struct {
	RedisName        string `json:"name"`
	RedisAddress     string `json:"redisAdd"`
	RedisDB          int    `json:"redisNum"`
	RedisRankAddress string `json:"redisRankAdd"`
	RedisRankDB      int    `json:"redisRankNum"`
	RedisAuth        string `json:"redisAuth"`
	ServerName       string `json:"serverName"`
	MailDBName       string `json:"mailDB"`
	MailMongoUrl     string `json:"MongoDBUrl"`
	MailDBDriver     string `json:"MongoDBDriver"`
	SysRollNoticeUrl string `json:"sysRollNoticeUrl"`
	AuthApi          string `json:"authapi"`
	AuthGagApi       string `json:"authgagapi"`
	ServerHotDataUrl string `json:"ServerHotDataUrl"`
	RankReloadUrl    string `json:"RankReloadUrl"`
	MergedShard      string `json:"MergedShard"`
}

func (c *CommonConfig) AddCfg(name string, scfg *ServerCfg) {
	c.RedisName = append(c.RedisName, scfg.RedisName)
	c.RedisAddress = append(c.RedisAddress, scfg.RedisAddress)
	c.RedisDB = append(c.RedisDB, scfg.RedisDB)
	c.RedisRankAddress = append(c.RedisRankAddress, scfg.RedisRankAddress)
	c.RedisRankDB = append(c.RedisRankDB, scfg.RedisRankDB)
	c.RedisAuth = append(c.RedisAuth, scfg.RedisAuth)
	c.ServerName = append(c.ServerName, scfg.ServerName)
	c.MailDBNames = append(c.MailDBNames, scfg.MailDBName)
	c.MailDBDrivers = append(c.MailDBDrivers, scfg.MailDBDriver)
	c.MailMongoUrls = append(c.MailMongoUrls, scfg.MailMongoUrl)
	c.SysRollNoticeUrl = append(c.SysRollNoticeUrl, scfg.SysRollNoticeUrl)
	c.AuthApi = append(c.AuthApi, scfg.AuthApi)
	c.AuthGagApi = append(c.AuthGagApi, scfg.AuthGagApi)
	c.ServerHotDataUrl = append(c.ServerHotDataUrl, scfg.ServerHotDataUrl)
	c.RankReloadUrl = append(c.RankReloadUrl, scfg.RankReloadUrl)
	c.MergedShard = append(c.MergedShard, scfg.MergedShard)
}

func (c *CommonConfig) IsServerMerged(server string) bool {
	cfg := c.GetServerCfgFromName(server)
	return cfg.MergedShard != cfg.RedisName
}

func (c *CommonConfig) GetServerCfgFromName(name string) ServerCfg {
	logs.Debug("debug server cfg , %v, %v", c.ServerName, c.RankReloadUrl)
	for idx, n := range c.ServerName {
		if n == name {
			rankReloadUrl := ""
			if idx < len(c.RankReloadUrl) {
				rankReloadUrl = c.RankReloadUrl[idx]
			}
			return ServerCfg{
				RedisName:        c.RedisName[idx],
				RedisAddress:     c.RedisAddress[idx],
				RedisDB:          c.RedisDB[idx],
				RedisRankAddress: c.RedisRankAddress[idx],
				RedisRankDB:      c.RedisRankDB[idx],
				RedisAuth:        c.RedisAuth[idx],
				ServerName:       c.ServerName[idx],
				MailDBName:       c.MailDBNames[idx],
				MailDBDriver:     c.MailDBDrivers[idx],
				MailMongoUrl:     c.MailMongoUrls[idx],
				SysRollNoticeUrl: c.SysRollNoticeUrl[idx],
				AuthApi:          c.AuthApi[idx],
				AuthGagApi:       c.AuthGagApi[idx],
				ServerHotDataUrl: c.ServerHotDataUrl[idx],
				RankReloadUrl:    rankReloadUrl,
				MergedShard:      c.MergedShard[idx],
			}
		}
	}
	return ServerCfg{}
}

func (c *CommonConfig) GetGidAllServer(ser string) []ServerCfg {
	res := make([]ServerCfg, 0, 10)
	ser_ss := strings.Split(ser, ":")
	for idx, n := range c.ServerName {
		cfg_ss := strings.Split(n, ":")
		if cfg_ss[0] == ser_ss[0] {
			res = append(res, ServerCfg{
				RedisName:        c.RedisName[idx],
				RedisAddress:     c.RedisAddress[idx],
				RedisDB:          c.RedisDB[idx],
				RedisRankAddress: c.RedisRankAddress[idx],
				RedisRankDB:      c.RedisDB[idx],
				ServerName:       c.ServerName[idx],
				MailDBName:       c.MailDBNames[idx],
				MailDBDriver:     c.MailDBDrivers[idx],
				MailMongoUrl:     c.MailMongoUrls[idx],
				SysRollNoticeUrl: c.SysRollNoticeUrl[idx],
				AuthApi:          c.AuthApi[idx],
				AuthGagApi:       c.AuthGagApi[idx],
				ServerHotDataUrl: c.ServerHotDataUrl[idx],
			})
		}
	}
	return res
}

var Cfg CommonConfig

type ServerCfgSlice []ServerCfg

func (pq ServerCfgSlice) Len() int { return len(pq) }

func (pq ServerCfgSlice) Less(i, j int) bool {
	return pq[i].ServerName < pq[j].ServerName
}

func (pq ServerCfgSlice) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
