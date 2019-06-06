package http_util

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	e    *gin.Engine
	host string
	gid  uint32
)

func Init(_host string, _gid uint32) {
	e = gin.Default()
	host = _host
	gid = _gid
	e.NoRoute(defaultHandler)
}

func POST(path string, handle func(c *gin.Context)) {
	logs.Debug("[HTTPUTIL] Post for path: %v", fmt.Sprintf("/%d%s", gid, path))
	e.POST(fmt.Sprintf("/%d%s", gid, path), handle)
}

func GET(path string, handle func(c *gin.Context)) {
	logs.Debug("[HTTPUTIL] Get for path: %v", fmt.Sprintf("/%d%s", gid, path))
	e.GET(fmt.Sprintf("/%d%s", gid, path), handle)
}

func RegETCD(root string, path string, groupID uint32) {
	logs.Debug("[HTTPUTIL] RegETCD for path: %v", fmt.Sprintf("http://%s/%d%s", host, gid, path))
	postService.RegTBServices(root, fmt.Sprintf("http://%s/%d%s", host, gid, path), groupID, uint(gid))
}

func Run() {
	go e.Run(host)
}
