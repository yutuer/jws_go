package city_broadcast

import (
	"net/http"
	"time"

	"fmt"
	"github.com/astaxie/beego/httplib"
	"golang.org/x/net/context"
	"strings"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/youtube/vitess/pools"
)

const (
	CBC_Typ_SysNotice = "SysRollNotice"
	CBC_Typ_Fish      = "FishInfo"
	CBC_Typ_Gve       = "Gve"
	CBC_Typ_FengHuo   = "FengHuo"
	CBC_Typ_GuildRoom = "GuildRoom"
)
const (
	connectTimeout   = 10 * time.Second
	readWriteTimeout = 10 * time.Second
)

var Pool CityBroadCastPool

type SendChanResource struct {
	t struct{}
}

func (res SendChanResource) Close() {}

type CityBroadCastPool struct {
	pool *pools.ResourcePool
}

func init() {
	Pool = CityBroadCastPool{
		pool: pools.NewResourcePool(
			func() (pools.Resource, error) {
				return SendChanResource{}, nil
			},
			100,         //Capacity
			1000,        //MaxCapacity
			time.Minute, //idleTimeout
		)}
}

type Msg struct {
	Typ   string   `json:"type"`
	Shd   string   `json:"shd"`
	Msg   string   `json:"msg"`
	Acids []string `json:"acids"`
}

func cityBroadCastSend(typ, shardId, message string, acids []string) {
	msg := Msg{typ, shardId, message, acids}
	// str, _ := json.Marshal(msg)
	info := strings.Split(shardId, ":")
	if len(info) < 2 {
		logs.Error("[cyt]wrong serve,need gid:sid,but only gid or sid")
		return
	}
	targeturl, err := etcd.Get(fmt.Sprintf("%s/%s/%s/gm/broadCast_url", game.Cfg.EtcdRoot, info[0], info[1]))
	if err != nil {
		logs.Error("[cyt]cannot find broatCast_url err %v", err)
		return
	}
	logs.Debug("[cyt]targeturl is :%v", targeturl)
	req := httplib.Post(targeturl).
		SetTimeout(connectTimeout, readWriteTimeout)
	req, err = req.JSONBody(msg)
	if err != nil {
		logs.Error("send CityBroadCastPool err %v", err)
		return
	}

	var res map[string]interface{}
	resp, err := req.Response()
	if err != nil {
		logs.Error("rev CityBroadCastPool resp err %v", err)
		return
	}
	defer resp.Body.Close()

	errCode := resp.StatusCode
	if errCode != http.StatusOK {
		logs.Error("rev CityBroadCastPool errcode %v", errCode)
		return
	}
	err = req.ToJSON(&res)
	if err != nil || res["status"] != "ok" {
		logs.Error("rev CityBroadCastPool err %v status %v", err, res)
		return
	}

	logs.Trace("City BroadCast send to shard %s %s %s %v",
		shardId, typ, message, acids)
}

func (pool *CityBroadCastPool) UseRes2Send(typ, shardId, msg string, acids []string) {
	ctx, cancel := context.WithTimeout(context.Background(),
		210*time.Millisecond)
	defer cancel()

	res, err := pool.pool.Get(ctx)
	switch err {
	case nil:
	//normal
	case pools.ErrTimeout:
		logs.Error("CityBroadCastPool get Pool timeout, err %s %s %s %s", ctx.Err(), typ, shardId, msg)
		return
	case pools.ErrClosed:
		logs.Error("CityBroadCastPool get Pool mistake? err:%s %s %s %s", err, typ, shardId, msg)
		return
	default:
		logs.Error("CityBroadCastPool get Pool error: %s %s %s %s", err.Error(), typ, shardId, msg)
		return
	}

	go func() {
		cityBroadCastSend(typ, shardId, msg, acids)
		pool.pool.Put(res)
	}()
}
