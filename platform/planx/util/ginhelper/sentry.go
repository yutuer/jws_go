package ginhelper

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var _sentry *raven.Client

func InitSentry(DSN string) error {
	InitSentryTags(DSN, nil)
	return nil
}

func InitSentryTags(DSN string, tags map[string]string) error {
	s, err := raven.NewClient(DSN, tags)
	if err != nil {
		logs.Warn("Init Sentry failed with error %s", err.Error())
		_sentry = nil
		return err
	}
	_sentry = s
	return nil
}

func SentryGinLog() gin.HandlerFunc {
	return sentryRecovery(_sentry, false)
}

func sentryRecovery(client *raven.Client, onlyCrashes bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			flags := map[string]string{
				"endpoint": c.Request.RequestURI,
			}
			if rval := recover(); rval != nil {
				debug.PrintStack()
				rvalStr := fmt.Sprint(rval)
				packet := raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)))
				client.Capture(packet, flags)
				c.Writer.WriteHeader(http.StatusInternalServerError)
			}
			if !onlyCrashes {
				for _, item := range c.Errors {
					packet := raven.NewPacket(item.Error(), &raven.Message{item.Error(), []interface{}{item.Meta}})
					client.Capture(packet, flags)
				}
			}
		}()
		c.Next()
	}
}
