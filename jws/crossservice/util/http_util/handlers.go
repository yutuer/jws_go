package http_util

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func defaultHandler(c *gin.Context) {
	logs.Debug("Got default GIN handle")

	routes := e.Routes()

	str := ""
	for _, r := range routes {
		subs := strings.Split(r.Handler, "/")
		str += fmt.Sprintf("-> [%s] %s -> %s\n", r.Method, r.Path, subs[len(subs)-1])
	}

	c.String(200, str)
}
