package logic

import (
	"time"

	"golang.org/x/net/context"

	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type EtcdWatcher struct {
	KeysAPI client.KeysAPI
}

func NewWatcher(endpoints []string, watchKey string) *EtcdWatcher {
	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	etcdClient, err := client.New(cfg)
	if err != nil {
		logs.Error("Error: cannot connec to etcd:", err)
		return nil
	}

	master := &EtcdWatcher{
		KeysAPI: client.NewKeysAPI(etcdClient),
	}
	go master.WatchWorkers(watchKey)
	return master
}

const (
	WatchAction_Update = "update"
	WatchAction_Set    = "set"
)

// https://coreos.com/etcd/docs/latest/v2/api.html
func (m *EtcdWatcher) WatchWorkers(watchKey string) {
	api := m.KeysAPI
	watcher := api.Watcher(watchKey, &client.WatcherOptions{})
	for {
		res, err := watcher.Next(context.Background())
		if err != nil {
			logs.Warn("Error watch workers:", err)
			// Watch from cleared event index
			watcher = api.Watcher(watchKey, &client.WatcherOptions{})
			continue
		}
		logs.Info("watch worker %v", res)
		if res.Action == WatchAction_Update || res.Action == WatchAction_Set {
			updateNotice()
		}
		// watcher.Next 是顺序监听，如果短时间内有连续的改变 我们只需要最后一次变化 不需要针对每次变化做出响应
		if res.Index > res.Node.ModifiedIndex {
			logs.Info("new watch, ModifiedIndex=%d, X-Etcd-Index%d", res.Node.ModifiedIndex, res.Index)
			// 使用X-Etcd-Index是为了保证能够监听到最后一次变化
			watcher = api.Watcher(watchKey, &client.WatcherOptions{AfterIndex: res.Index - 1})
		}
	}
}
