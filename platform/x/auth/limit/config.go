package limit

import (
	"net"

	"sync"

	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type internalIP_str struct {
	From string `toml:"from"`
	To   string `toml:"to"`
}

type internalIP struct {
	From net.IP
	To   net.IP
}

type LimitConfig struct {
	LimitUrlRegex    string `toml:"limit_url_regex"`
	InternalIPs      []internalIP_str
	RateLimitValid   bool  `toml:"rate_limit_valid"`
	RateLimitAverage int64 `toml:"rate_limit_Average"`
	RateLimitBurst   int64 `toml:"rate_limit_Burst"`
}

var (
	LimitCfg     LimitConfig
	internal_ips []internalIP // 内网ip段

	ipmux          sync.RWMutex
	remoteIPLimits []internalIP // 管理员ip段，用于外部玩家不能进但内部人员可以进游戏时使用
)

func SetIPLimitCfg(iprc []iplimitconfig.IPRangeConfig) {
	if iprc == nil || len(iprc) == 0 {
		return
	}
	ipmux.Lock()
	defer ipmux.Unlock()
	remoteIPLimits = make([]internalIP, 0, len(iprc))
	for _, v := range iprc {
		remoteIPLimits = append(remoteIPLimits, internalIP{net.ParseIP(v.From), net.ParseIP(v.To)})
	}
	logs.Trace("admin_ips %v", remoteIPLimits)
}

func SetLimitCfg(limitCfg LimitConfig) {
	LimitCfg = limitCfg
	internal_ips = make([]internalIP, 0, len(LimitCfg.InternalIPs))
	for _, v := range LimitCfg.InternalIPs {
		internal_ips = append(internal_ips, internalIP{net.ParseIP(v.From), net.ParseIP(v.To)})
	}

	logs.Trace("LimitCfg %v netips %v", LimitCfg, internal_ips)
}
