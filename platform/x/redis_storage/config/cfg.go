package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/mitchellh/mapstructure"

	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

var (
	CommonCfg  CommonConfig
	OnlandCfg  OnLandConfig
	RestoreCfg RestoreConfig
)

type CommonConfig struct {
	RunMode string `toml:"run_mode"`

	ShardId []int `toml:"shard_id"`
	Gid     uint  `toml:"gid"`

	Command_Addr string `toml:"Command_Addr"`
	Api_Addr     string `toml:"Api_Addr"`

	NeedStoreHead string `toml:NeedStoreHead`
	StoreMode     string `toml:StoreMode`

	AWS_Region    string `toml:"AWS_Region"`
	AWS_AccessKey string `toml:"AWS_AccessKey"`
	AWS_SecretKey string `toml:"AWS_SecretKey"`

	Format       string   `toml:"Onland_Format"`
	FormatSep    string   `toml:"Onland_Format_Separator"`
	EtcdEndpoint []string `toml:"etcd_endpoint"`
	EtcdRoot     string   `toml:"etcd_root"`
}

type OnLandConfig struct {
	Redis     string `toml:"Redis"`
	RedisAuth string `toml:"Redis_auth"`
	RedisDB   int    `toml:"Redis_db"`

	Workers_Max int `toml:"Workers_Max"`

	Backends []string `toml:"Backends"`
}

type RestoreConfig struct {
	Workers_Restore_Max int `toml:"Workers_Restore_Max"`
	Restore_Scan_Len    int `toml:"Restore_Scan_Len"`

	Restore_Data_from string `toml:"Restore_Data_from"`

	Redis_Restore     string `toml:"Redis_Restore"`
	RedisAuth_Restore string `toml:"Redis_Restore_auth"`
	RedisDB_Restore   int    `toml:"Redis_Restore_db"`

	SSDB_Ip      string
	SSDB_Port    int
	SSDB_Address string `toml:"SSDB_Address"`
}

var Cfg_Time string

func mkip(s string) (bool, string, int) {
	r := strings.Split(s, ":")
	if len(r) != 2 {
		logs.Error("Address To Ip Err by %s", s)
		return false, "", 0
	}

	i, err := strconv.ParseInt(r[1], 10, 64)
	if err != nil {
		logs.Error("Address To Ip Err by %s in %s", s, err.Error())
		return false, "", 0
	}
	return true, r[0], int(i)
}

func ReadConfig(c *cli.Context) {
	Cfg_Time = c.String("nowtime")

	cfgName := c.String("config")
	var common_cfg struct{ CommonCfg CommonConfig }
	cfgApp := config.NewConfigToml(cfgName, &common_cfg)
	CommonCfg = common_cfg.CommonCfg
	if cfgApp == nil {
		logs.Critical("CommonConfig Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	logs.Info("Common Config loaded %v.", CommonCfg)

	var onland_cfg struct{ OnlandCfg OnLandConfig }
	cfgApp = config.NewConfigToml(cfgName, &onland_cfg)
	OnlandCfg = onland_cfg.OnlandCfg
	if cfgApp == nil {
		logs.Critical("OnlandCfg Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	logs.Info("Onland Config loaded %v.", OnlandCfg)

	var restore_cfg struct{ RestoreCfg RestoreConfig }
	cfgApp = config.NewConfigToml(cfgName, &restore_cfg)
	RestoreCfg = restore_cfg.RestoreCfg
	if cfgApp == nil {
		logs.Critical("RestoreCfg Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	if RestoreCfg.SSDB_Address != "" {
		ok, ssdb_ip, ssdb_port := mkip(RestoreCfg.SSDB_Address)
		if ok {
			RestoreCfg.SSDB_Ip = ssdb_ip
			RestoreCfg.SSDB_Port = ssdb_port
		} else {
			logs.Error("Restore SSDB Address Error!")
			return
		}
	}

	logs.Info("Restore Config loaded %v.", RestoreCfg)

	//----------onland target config loading
	cfgApp = config.NewConfigToml(cfgName, cfgsmaps)
	if cfgApp == nil {
		logs.Critical("OnlandTargetCfg Read Error\n")
		logs.Close()
		os.Exit(1)
	}

	logs.Info("cfgs %v", cfgs)
	for k, _ := range cfgs {
		c := GetTarget(k)
		if c != nil {
			logs.Info("Onland Target %s is loaded. %v", k, c)
		}
	}

}

func init() {
	//Register("RestoreCfg", &RestoreConfig{})
	//Register("RestoreCfg", &RestoreConfig{})
	//Register("RestoreCfg", &RestoreConfig{})
}

type OnlandTargetCfg interface {
	HasConfigured() bool
	Setup(addFn func(storehelper.IStore))
}

var (
	cfgs          = make(map[string]interface{})
	cfgsmaps      = make(map[string]interface{})
	cfgsinterface = make(map[string]OnlandTargetCfg)
)

func GetTarget(key string) OnlandTargetCfg {
	if v, ok := cfgsinterface[key]; ok {
		c := cfgs[key]
		mapstructure.Decode(cfgsmaps[key], c)
		v.HasConfigured()
		return v
	}
	return nil
}

func Register(name string, c OnlandTargetCfg) {
	if c == nil {
		return
	}

	if _, ok := cfgs[name]; ok {
		log.Fatalln("cfg.go: Register called twice for adapter " + name)
	}
	cfgs[name] = c
	cfgsinterface[name] = c
}
