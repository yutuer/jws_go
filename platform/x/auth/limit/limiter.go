package limit

import (
	//"net"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/mailgun/oxy/utils"
	"github.com/mailgun/timetools"
	//"vcs.taiyouxi.net/platform/planx/util/iputil"
	"vcs.taiyouxi.net/platform/planx/util/iputil"
	"vcs.taiyouxi.net/platform/planx/util/ratelimit"

	"github.com/gin-gonic/gin"
	"strings"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	internalUrlRegex *regexp.Regexp
)

func getMatchLimitedUrlRegex() (*regexp.Regexp, error) {
	if internalUrlRegex == nil {
		var err error
		internalUrlRegex, err = regexp.Compile(LimitCfg.LimitUrlRegex)
		if err != nil {
			return nil, err
		}
	}
	return internalUrlRegex, nil

}

func CheckIdentity(specHeader, specContent string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, err := getMatchLimitedUrlRegex()
		if err != nil {
			c.JSON(http.StatusUnauthorized, "IT1")
			c.Abort()
			return
		}
		//logs.Trace("****** yzh, %v", c.Request.URL.Path)
		ok := r.MatchString(c.Request.URL.Path)
		if ok {
			spec := c.Request.Header.Get(specHeader)
			if spec == "" || spec != specContent {
				// logs.Warn("[limit] Identity refuse by not had spec header : url %s ip %s specheader %s",
				//	c.Request.URL.Path, c.ClientIP(), spec)
				c.JSON(http.StatusUnauthorized, "-_-")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func InternalOnlyLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		//需要限定在内网
		clientIP := c.ClientIP()
		if isInternalIp(clientIP, internal_ips) {
			c.Next()
		} else {
			c.JSON(http.StatusForbidden, "limited ip")
			c.Abort()
		}
	}
}

func IsLimitedByRemoteIP(ip string) bool {
	ipmux.RLock()
	defer ipmux.RUnlock()
	ret := isLimitedIP(ip, remoteIPLimits[:])
	return ret
}

func isLimitedIP(ip string, ip_cfg []internalIP) bool {
	if IpOnly, _, err := net.SplitHostPort(ip); err != nil {
		return false
	} else {
		var limited bool
		ipOnly := net.ParseIP(IpOnly)
		for _, v := range ip_cfg {
			limited = limited || iputil.IpBetween(v.From, v.To, ipOnly)
			if limited {
				return true
			}
		}
	}
	return false
}

func isInternalIp(ip string, ip_cfg []internalIP) bool {
	IpOnly := ip
	if strings.Contains(ip, ":") {
		ipOnly, _, err := net.SplitHostPort(ip)
		if err != nil {
			return false
		}
		IpOnly = ipOnly
	}
	//检查是否是内部IP
	var isInternal bool
	isInternal = (IpOnly == "127.0.0.1")
	if isInternal {
		return true
	} else {
		return isLimitedIP(ip, ip_cfg)
	}
	return false

}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if LimitCfg.RateLimitValid { // 功能开启
			// url是否需要检查
			r, err := getMatchLimitedUrlRegex()
			if err != nil {
				c.JSON(http.StatusUnauthorized, "IT2")
				c.Abort()
				return
			}
			ok := r.MatchString(c.Request.URL.Path)
			if ok {
				//limited url
				if err := rateLimit.Consume(c.Request); err != nil {
					logs.Warn("[limit] ratelimit ip: %s", c.ClientIP())
					c.JSON(429, "Too Many Requests")
					c.Abort()
					return
				}
			} else {
				//url is not in limit, etc, /api
				//所有内部API都不做限制
			}
		}
		c.Next()
	}
}

var (
	rateLimit *ratelimit.TokenLimiter
)

func Init() {
	if LimitCfg.RateLimitValid {

		rates := ratelimit.NewRateSet()
		rates.Add(time.Second, LimitCfg.RateLimitAverage, LimitCfg.RateLimitBurst)

		extr, err := utils.NewExtractor("client.ip")
		if err != nil {
			panic("init ratelimite NewExtractor err: " + err.Error())
		}

		l, err := ratelimit.NewForHttp(
			nil,
			extr,
			rates,
			ratelimit.Clock(&timetools.RealTime{}))
		if err != nil {
			panic("init ratelimite new limiter err: " + err.Error())
		}
		rateLimit = l
	}
}
