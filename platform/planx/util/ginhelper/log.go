package ginhelper

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/timesking/seelog"
	"os"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// use http://goaccess.io/ to analysis log
// how to make output is nginx common log format
//    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
//											'$status $body_bytes_sent $request_time '
//					                      '"$http_user_agent"' comment;
func NginxLoggerWithWriter() gin.HandlerFunc {
	return nginxLoggerWithWriter(nil)
}

func NgixLoggerToFile(cfgName string) (gin.HandlerFunc, func()) {
	logger, err := seelog.LoggerFromParamConfigAsFile(cfgName, nil)
	if err != nil {
		logs.Critical("NgixLoggerToFile config load error %s", err.Error())
		logs.Close()
		os.Exit(1)
		return nil, nil
	}
	return nginxLoggerWithWriter(logger), func() {
		if logger != nil {
			logger.Flush()
			logger.Close()
			logger = nil
		}
	}
}

func nginxLoggerWithWriter(logger seelog.LoggerInterface) gin.HandlerFunc {

	myLogger := func(format string, v ...interface{}) {
		if logger != nil {
			logger.Infof(format, v...)
		} else {
			logs.Info(format, v...)
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		clientIPonly := strings.Split(clientIP, ":")[0]
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()
		if comment == "" {
			comment = "-"
		}
		clientProto := c.Request.Proto

		myLogger(`%s - - [%s] "%s %s %s" %d %d "%s" "%s"`,
			clientIPonly,                             //$remote_addr - $remote_user
			end.Format("02/Jan/2006:15:04:05 -0700"), //time_local
			method, path, clientProto, //"$request"
			//TODO body_bytes_sent is 0 here, maybe need it in future
			// request time is seconds with millionsec precision, but golang float multiply is confuzing.
			// so, I just use ns precision here!
			//statusCode, 0, (0.000000001)*(float64(latency)), //$status $body_bytes_sent $request_time
			statusCode, latency, //$status $body_bytes_sent $request_time
			c.Request.UserAgent(),
			comment, //comment
		)
	}
}

//func LoggerWithWriter() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		// Start timer
//		start := time.Now()
//		path := c.Request.URL.Path
//
//		// Process request
//		c.Next()
//
//		// Stop timer
//		end := time.Now()
//		latency := end.Sub(start)
//
//		clientIP := c.ClientIP()
//		method := c.Request.Method
//		statusCode := c.Writer.Status()
//		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()
//
//		logs.Info("%v | %3d | %s | %d | %s | %s | %s",
//			end.Format("2006/01/02 - 15:04:05"),
//			statusCode, method, latency,
//			clientIP, path, comment,
//		)
//	}
//}
