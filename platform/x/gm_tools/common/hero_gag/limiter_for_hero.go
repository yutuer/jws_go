package hero_gag

import (
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"vcs.taiyouxi.net/platform/planx/util/iputil"

	"sync"
	"vcs.taiyouxi.net/platform/planx/util/iplimitconfig"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type internalIP struct {
	From net.IP
	To   net.IP
}

var (
	commonlimitedIPs []internalIP
	ipmux            sync.RWMutex
)

const InternalIP = "127.0.0.1"

func CommonIPLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if isCommonLimitedIP(clientIP) {
			c.Next()
		} else {
			c.JSON(http.StatusForbidden, "limited ip")
			c.Abort()
		}
	}
}

func isCommonLimitedIP(clientIP string) bool {
	if IpOnly, _, err := net.SplitHostPort(clientIP); err != nil {
		return false
	} else {
		var isInternal bool
		isInternal = (IpOnly == InternalIP)
		if isInternal {
			return true
		} else {
			var limited bool
			ipOnly := net.ParseIP(IpOnly)
			for _, v := range commonlimitedIPs {
				limited = limited || iputil.IpBetween(v.From, v.To, ipOnly)
				if limited {
					return true
				}
			}
		}
		return false
	}
}

func SetCommonIPLimitCfg(iprc []iplimitconfig.IPRangeConfig) {
	if iprc == nil || len(iprc) == 0 {
		return
	}
	ipmux.Lock()
	defer ipmux.Unlock()
	commonlimitedIPs = make([]internalIP, 0, len(iprc))
	for _, v := range iprc {
		commonlimitedIPs = append(commonlimitedIPs, internalIP{net.ParseIP(v.From), net.ParseIP(v.To)})
	}
	logs.Info("admin_ips %v", commonlimitedIPs)
}
