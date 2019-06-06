package config

import (
	"errors"
	// "fmt"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
)

type LoadCmd int

const (
	Nothing LoadCmd = 0
	Load            = 1 //The config file should be loaded as init loading. feature off -> on
	Reload          = 2 //should be loaded again. feature on -> off -> on, or on -> on
	Unload          = 3 //The config file disapear, so just turn off your feature!. feature on->off
)

//Config 提供Reload能力
//但是AppConfigPath只依赖于程序启动时NewConfigPath中发现的路径进行reload操作
type Config struct {
	AppConfigPath string
	Md5           string
	LoadFunc      func(string, LoadCmd)
}

var (
	AppPath string
)

func init() {

}

func NewConfig(cfgname string, reload bool, loadf func(string, LoadCmd)) *Config {
	var cfg Config
	cfg.AppConfigPath = NewConfigPath(cfgname)
	file, inerr := os.Open(cfg.AppConfigPath)
	logs.Info("NewConfig: %s", cfg.AppConfigPath)
	if inerr == nil {
		md5h := md5.New()
		io.Copy(md5h, file)
		cfg.Md5 = hex.EncodeToString(md5h.Sum(nil))
		if loadf != nil {
			cfg.LoadFunc = loadf
			loadf(cfg.AppConfigPath, Load)
			if reload {
				signalhandler.SignalReloadFunc(func() {
					cfg.Reload()
				})
			}
		}
		return &cfg
	}
	logs.Warn("NewConfig %s err %s", cfgname, inerr.Error())
	return nil
}

// NewConfigPath搜索
//1. 应用程序可执行文件做在路径
//  1.1 confd 目录：为了更容易支持etcd/confd方案
//  1.2 conf 目录：正常手动配置模式
//2. 当前work folder的路径下, confd, conf
func NewConfigPath(cfgname string) string {
	workPath, _ := os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	appConfigPath1 := filepath.Join(AppPath, "confd", cfgname)
	appConfigPath2 := filepath.Join(AppPath, "conf", cfgname)

	appConfigPath := appConfigPath1
	//if workPath != AppPath {
	if util.FileExists(appConfigPath1) {
		os.Chdir(AppPath)
	} else if util.FileExists(appConfigPath2) {
		os.Chdir(AppPath)
		appConfigPath = appConfigPath2
	} else {
		appConfigPath1 = filepath.Join(workPath, "confd", cfgname)
		appConfigPath2 = filepath.Join(workPath, "conf", cfgname)

		appConfigPath = appConfigPath1
		if util.FileExists(appConfigPath1) {
			appConfigPath = appConfigPath1
		} else if util.FileExists(appConfigPath2) {
			appConfigPath = appConfigPath2
		}
	}
	//}
	return appConfigPath
}

//config is the file name, ex: gate.toml
func NewConfigToml(cfgname string, v interface{}) *Config {
	return NewConfig(cfgname, false, func(lcfgname string, loadStatus LoadCmd) {
		if loadStatus == Load {
			if _, err := toml.DecodeFile(lcfgname, v); err != nil {
				logs.Critical("App config load failed. %s, %s\n", lcfgname, err.Error())
			} else {
				logs.Info("Config loaded: %s\n", lcfgname)
			}
		} else {
			logs.Warn("NewConfigToml does not support reload and unload.")
		}
	})
}

//Use for reload configuration
func (c *Config) Reload() {
	if c == nil {
		return
	}

	if c.LoadFunc == nil {
		return
	}
	loadf := c.LoadFunc
	if c.Md5 == "" {
		if _, err := os.Stat(c.AppConfigPath); err == nil {
			var newMd5 string
			file, inerr := os.Open(c.AppConfigPath)
			if inerr == nil {
				md5h := md5.New()
				io.Copy(md5h, file)
				newMd5 = hex.EncodeToString(md5h.Sum(nil))
				loadf(c.AppConfigPath, Load)
				c.Md5 = newMd5
				return
			} else {
				logs.Error("Config.Reload md5 open file failed, %v", inerr)
			}
		}
	} else {
		if _, err := os.Stat(c.AppConfigPath); os.IsNotExist(err) {
			loadf(c.AppConfigPath, Unload)
			c.Md5 = ""
			return
		} else {
			var newMd5 string
			file, inerr := os.Open(c.AppConfigPath)
			if inerr == nil {
				md5h := md5.New()
				io.Copy(md5h, file)
				newMd5 = hex.EncodeToString(md5h.Sum(nil))

				if newMd5 == c.Md5 {
					logs.Info("Nothing to load, Md5 is no changes, %s", c.AppConfigPath)
					return
				} else {
					logs.Info("Need to load, Md5 is changed, %s", c.AppConfigPath)
					loadf(c.AppConfigPath, Reload)
					c.Md5 = newMd5
					return
				}
			} else {
				logs.Error("Config.Reload md5 open file failed, %v", inerr)
			}

		}
	}
	return
}

// DebugLoadConfigToml 直接读取GOPATH路径下，指定文件名的toml文件
// 譬如 "src/vcs.taiyouxi.net/jws/multiplayer/match_server/conf/config.toml"
// 因为进行单元测试时，当前目录可能是任意位置, 并且有时需求读取修改后的配置文件
func DebugLoadConfigToml(cfgname string, v interface{}) error {
	var cfgFullName string
	goPathes := getSplitGoPath()

	for _, goPath := range goPathes {
		curPath := filepath.Join(goPath, cfgname)
		if stat, err := os.Stat(curPath); !os.IsNotExist(err) && !stat.IsDir() {
			cfgFullName = curPath
		}
	}

	if cfgFullName == "" {
		return errors.New("DebugLoadConfigToml FAILED! CANNOT find toml file!!!")
	}

	if _, err := toml.DecodeFile(cfgFullName, v); err != nil {
		return errors.New("DebugLoadConfigToml FAILED! CANNOT decode file!!!")
	}

	return nil
}

// GetSplitGoPath 读取环境GOPATH并转换为slice
func getSplitGoPath() []string {
	var goPathes []string
	goPath := os.Getenv("GOPATH")

	if runtime.GOOS == "windows" && strings.Contains(goPath, ";") {
		goPathes = strings.Split(goPath, ";")
	} else if runtime.GOOS != "windows" && strings.Contains(goPath, ":") {
		goPathes = strings.Split(goPath, ":")
	} else {
		goPathes = []string{goPath}
	}

	return goPathes
}
