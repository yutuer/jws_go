package etcd

import (
	"context"
	"sync"
	"testing"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	WatchAction_Update = "update"
	WatchAction_Set    = "set"
)

//func DoWatch(res *client.Response){
//	if res.Action == WatchAction_Set || res.Action == WatchAction_Update{
//		fmt.Println("xxx have been changed ",res.Node.Value)
//	}
//}

func TestEtcdWatch(t *testing.T) {
	InitClient([]string{"http://127.0.0.1:2379/"})
	logs.Debug("begin watch")
	res := make(chan *client.Response, 10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	NewWatcher([]string{"http://127.0.0.1:2379/"}, "/a4k", res)
	var warp sync.WaitGroup
	warp.Add(1)
	for {
		select {
		case ret := <-res:
			logs.Debug("xxx have been changed ", ret.Node.Value)
			logs.Debug(ret.Node.Value)
		case <-ctx.Done():
			return
		}
	}
}
