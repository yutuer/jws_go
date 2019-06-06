package metrics

import (
	"strings"

	"net"
	"strconv"

	"errors"

	"github.com/BurntSushi/toml"
	gm "github.com/rcrowley/go-metrics"
	"vcs.taiyouxi.net/platform/planx/metrics/simplegraphite"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	_cfg                     Config
	_stop_graphite           chan struct{}
	_goStatsPerSeond         gm.Registry
	_goCurstomStatsPerSecond gm.Registry
	_ip                      string
	_simpleGraphite          *simplegraphite.Graphite
)

func init() {
	_goStatsPerSeond = gm.NewRegistry()
	_goCurstomStatsPerSecond = gm.NewRegistry()
	ip, err := util.GetExternalIP()
	if err != nil {
		panic(err)
	}
	_ip = strings.Replace(ip, ".", "_", -1)
	_simpleGraphite = simplegraphite.NewGraphiteNop("127.0.0.1", 2003)
}

func GetIPToken() string {
	return _ip
}

//func GetRegister() gm.Registry {
//return _goStatsPerSeond
//}

func NewCounter(name string) gm.Counter {
	return gm.NewRegisteredCounter(name, _goStatsPerSeond)
}

func NewMeter(name string) gm.Meter {
	return gm.NewRegisteredMeter(name, _goStatsPerSeond)
}

func NewGauge(name string) gm.Gauge {
	return gm.NewRegisteredGauge(name, _goStatsPerSeond)
}

func NewCustomCounter(name string) gm.Counter {
	return gm.NewRegisteredCounter(name, _goCurstomStatsPerSecond)
}

func NewCustomMeter(name string) gm.Meter {
	return gm.NewRegisteredMeter(name, _goCurstomStatsPerSecond)
}

func NewCustomGauge(name string) gm.Gauge {
	return gm.NewRegisteredGauge(name, _goCurstomStatsPerSecond)
}

func Start(cfgname string) {
	config.NewConfig(cfgname, true, func(lcfgname string, cmd config.LoadCmd) {
		switch cmd {
		case config.Load, config.Reload:
			var c Config
			if _, err := toml.DecodeFile(lcfgname, &c); err != nil {
				logs.Critical("App config load failed. %s, %s\n", lcfgname, err.Error())
			} else {
				logs.Info("Config loaded: %s\n", lcfgname)
				if cmd == config.Reload {
					Reload(c)
					_cfg = c
					return
				}
				_cfg = c
				logs.Info("Config _cfg: %+v", _cfg)
			}
		case config.Unload:
			close(_stop_graphite)
			if _simpleGraphite != nil {
				_simpleGraphite.Disconnect()
			}
			_simpleGraphite = simplegraphite.NewGraphiteNop("127.0.0.1", 2003)
			return
		}
	})

	logs.Info("metric started.")
	//统计状态数据应该可以通过信号Reload
	_stop_graphite = make(chan struct{})
	go startGolangStats(_stop_graphite)
	if _simpleGraphite != nil {
		_simpleGraphite.Disconnect()
	}
	if sg, err := NewSimpleGraphie(_cfg); err == nil {
		_simpleGraphite = sg
	} else {
		logs.Error("NewSimpleGraphie err: %s", err.Error())
	}
}

func NewSimpleGraphie(cfg Config) (*simplegraphite.Graphite, error) {
	host, port, err := net.SplitHostPort(_cfg.GraphiteHost)
	if err != nil {
		return nil, err
	}
	if pport, err := strconv.Atoi(port); err != nil {
		return nil, err
	} else {
		if gg, err := simplegraphite.NewGraphiteUDP(host, pport); err != nil {
			return nil, err
		} else {
			return gg, nil
		}
	}
	return nil, errors.New("NewSimpleGraphie should not happpen.")
}

func Reload(cfg Config) {
	close(_stop_graphite)
	_cfg = cfg
	_stop_graphite = make(chan struct{})
	go startGolangStats(_stop_graphite)

	if _simpleGraphite != nil {
		_simpleGraphite.Disconnect()
	}
	if sg, err := NewSimpleGraphie(_cfg); err == nil {
		_simpleGraphite = sg
	} else {
		logs.Error("NewSimpleGraphie err: %s", err.Error())
	}
}

func Stop() {
	close(_stop_graphite)
}

func SimpleSend(stat string, value string) error {
	if !_cfg.SimpleGraphite {
		return nil
	}
	if _simpleGraphite != nil {
		return _simpleGraphite.SimpleSend(stat, value)
	}
	return nil
}
