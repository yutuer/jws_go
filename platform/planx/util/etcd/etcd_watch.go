package etcd

import (
	"time"

	"golang.org/x/net/context"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	WatchAction_Update           = "update"
	WatchAction_Set              = "set"
	WatchAction_Del              = "delete"
	WatchAction_Create           = "create"
	WatchAction_CompareAndSwap   = "compareAndSwap"
	WatchAction_CompareAndDelete = "compareAndDelete"
)

type EtcdWatcher struct {
	KeysAPI client.KeysAPI
}
type Hander func(res *client.Response)

//,quitChan chan bool
func NewWatcher(endpoints []string, watchKey string, res chan *client.Response) *EtcdWatcher {
	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	fmt.Println("New a watcher")
	etcdClient, err := client.New(cfg)
	if err != nil {
		logs.Error("Error: cannot connec to etcd:", err)
		return nil
	}

	master := &EtcdWatcher{
		KeysAPI: client.NewKeysAPI(etcdClient),
	}
	go master.WatchWorkers(watchKey, res)
	return master
}

// https://coreos.com/etcd/docs/latest/v2/api.html
func (m *EtcdWatcher) WatchWorkers(watchKey string, resChan chan *client.Response) {
	api := m.KeysAPI
	watcher := api.Watcher(watchKey, &client.WatcherOptions{Recursive: true})
	for {
		res, err := watcher.Next(context.Background())
		if err != nil {
			se, ok := err.(client.Error)
			if !ok {
				logs.Error("%v", err)
			}
			logs.Warn("Error watch workers:", err)
			switch se.Code {
			// Watch from cleared event index
			case client.ErrorCodeEventIndexCleared:
				watcher = api.Watcher(watchKey, &client.WatcherOptions{Recursive: true, AfterIndex: se.Index - 1})
				continue
			}
		}
		resChan <- res
	}
}
