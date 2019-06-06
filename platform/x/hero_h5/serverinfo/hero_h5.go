package serverinfo

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"sync"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/auth/models"
	h5Config "vcs.taiyouxi.net/platform/x/hero_h5/config"
	"vcs.taiyouxi.net/platform/x/tiprotogen/log"
)

var (
	rsp   Rsp
	mutex sync.Mutex
)

type RspSerInfo struct {
	Name  string `json:"name"`
	Serid string `json:"serverid"`
}

type Rsp struct {
	Result         int
	IosServers     []RspSerInfo `json:"ios"`
	AndroidServers []RspSerInfo `json:"android"`
}

/*
gin的注册
*/
func RegisterAllServerInfo(e *gin.Engine) {
	log.Debug("begin to register")
	e.GET("/getServerInfo", ReturnAllServerInfo)
	log.Debug("register finish")

	if err := UpdateServerInfo(); err != nil {
		logs.Error("get shard info err %v", err)
	}
	logs.Debug("get all shardInfo")
	go func() {
		timeTicker := time.NewTicker(30 * time.Minute)
		for {
			<-timeTicker.C
			err := UpdateServerInfo()

			if err != nil {
				logs.Error("get shard info err %v", err)
			}
			logs.Debug("get all shardInfo")
		}
	}()
}

/*
返回服务器列表信息
*/
func ReturnAllServerInfo(c *gin.Context) {
	mutex.Lock()
	c.JSON(200, rsp)
	mutex.Unlock()
}

/*
更新服务器的列表信息
*/
func UpdateServerInfo() error {
	iosgids := h5Config.CommonConfig.IosGid
	andgids := h5Config.CommonConfig.AndGid
	data0 := make([]RspSerInfo, 0)
	data1 := make([]RspSerInfo, 0)
	logs.Debug("try to get info from etcd")
	for _, gid := range iosgids {
		info, err := models.GetAllShard(gid)
		if err != nil {
			rsp.Result = 0
			return err
		}
		logs.Debug("get shard info: %v from ios %d", info, gid)
		for _, item := range info {
			data0 = append(data0, RspSerInfo{
				Name:  item.DName,
				Serid: fmt.Sprintf("%d:%d", item.Gid, item.Sid),
			})
		}
	}

	for _, gid := range andgids {
		info, err := models.GetAllShard(gid)
		if err != nil {
			rsp.Result = 0
			return err
		}
		logs.Debug("get shard info: %v from andr %d", info, gid)
		for _, item := range info {
			data1 = append(data1, RspSerInfo{
				Name:  item.DName,
				Serid: fmt.Sprintf("%d:%d", item.Gid, item.Sid),
			})
		}
	}

	trsp := Rsp{
		Result:         1,
		IosServers:     data0,
		AndroidServers: data1,
	}
	mutex.Lock()
	rsp = trsp
	mutex.Unlock()
	return nil
}
