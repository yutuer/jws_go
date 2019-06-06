package security

import (
	"strconv"
	"strings"
	"time"

	"net"

	"sync"

	gm "github.com/rcrowley/go-metrics"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/iputil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/ratelimit"
)

type internalIP struct {
	From net.IP
	To   net.IP
}

var (
	limiter          *ratelimit.TokenStringLimiter
	isValid          bool
	cfgUrl           map[string]*ratelimit.RateSet
	ratelimitCounter gm.Counter

	adminIps []internalIP
	ipmux    sync.RWMutex
)

func LoadConf(valid bool, urlLimitInfo string, adminIps_str []game.InternalIP_str) {
	isValid = valid
	if !isValid {
		return
	}

	urlInfos := strings.Split(urlLimitInfo, ";")
	cfgUrl = make(map[string]*ratelimit.RateSet, len(urlInfos))
	for _, url := range urlInfos {
		infos := strings.Split(url, ",")
		if len(infos) < 3 {
			logs.Error("gamex ratelimit url cfg not enough: %s", url)
			continue
		}
		rates := ratelimit.NewRateSet()
		t, err := strconv.ParseInt(infos[1], 10, 64)
		if err != nil {
			logs.Error("gamex ratelimit url cfg time err: %s", url)
			continue
		}
		n, err := strconv.ParseInt(infos[2], 10, 64)
		if err != nil {
			logs.Error("gamex ratelimit url cfg count err: %s", url)
			continue
		}
		rates.Add(time.Second*time.Duration(t), n, n)
		cfgUrl[infos[0]] = rates
	}
	ratelimitCounter = metrics.NewCounter("gamex.ratelimit")
	logs.Trace("ratelimite load conf %v", cfgUrl)
}

func SetIPLimitCfg(iprc []iplimitconfig.IPRangeConfig) {
	if iprc == nil || len(iprc) == 0 {
		return
	}
	ipmux.Lock()
	defer ipmux.Unlock()
	adminIps = make([]internalIP, 0, len(iprc))
	for _, v := range iprc {
		adminIps = append(adminIps, internalIP{net.ParseIP(v.From), net.ParseIP(v.To)})
	}
	logs.Trace("admin_ips %v", adminIps)
}

func IsLimitedByRemoteIP(ip string) bool {
	ipmux.RLock()
	defer ipmux.RUnlock()
	ret := isLimitedIP(ip, adminIps[:])
	return ret
}

func isLimitedIP(ip string, ip_cfg []internalIP) bool {
	IpOnly := ip
	if strings.Contains(ip, ":") {
		_ip, _, err := net.SplitHostPort(ip)
		if err != nil {
			logs.Error("isAdminIp %s SplitHostPort err %v", ip, err)
			return false
		}
		IpOnly = _ip
	}

	var limited bool
	ipOnly := net.ParseIP(IpOnly)
	for _, v := range ip_cfg {
		limited = limited || iputil.IpBetween(v.From, v.To, ipOnly)
		if limited {
			return true
		}
	}

	return false
}

func Init() {
	if !isValid {
		return
	}

	rates := ratelimit.NewRateSet()
	rates.Add(time.Second, 1, 10)

	stringExtractRates := func(s string) (*ratelimit.RateSet, error) {
		rates, ok := cfgUrl[getUrl(s)]
		if ok {
			return rates, nil
		}
		// 找不到则不限制
		return nil, nil
	}

	l, err := ratelimit.NewForString(rates, ratelimit.StringExtractRates(ratelimit.StringRateExtractorFunc(stringExtractRates)))
	if err != nil {
		logs.Error("gamex init ratelimit err: %s", err.Error())
		return
	}
	limiter = l
}

func Consume(source string) error {
	if !isValid {
		return nil
	}
	if err := limiter.Consume(source); err != nil {
		ratelimitCounter.Inc(1)
		return err
	}
	return nil
}

func GenSource(acid string, url string) string {
	return acid + "@" + url
}

func LimitIpKick(ip string) (isKick bool) {
	if game.Cfg.Gid <= 0 {
		return false
	}
	show_state, err := strconv.Atoi(game.Cfg.ShardShowState)
	if err != nil ||
		(show_state != etcd.ShowState_Experience &&
			show_state != etcd.ShowState_Maintenance) {
		return false
	}
	// 限制ip
	if IsLimitedByRemoteIP(ip) {
		return false
	}
	return true
}

func getUrl(source string) string {
	ss := strings.Split(source, "@")
	return ss[1]
}
