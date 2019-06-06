package discover

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//EventType ..
type EventType int32

//EventType
const (
	EventUnknown = EventType(iota)
	EventPut
	EventDel
)

//HandleWatchEvent ..
type HandleWatchEvent func(et EventType, path string, service *Service)

//StartWatcher ..
func StartWatcher(handles map[string]HandleWatchEvent, stop chan struct{}) error {
	cli, err := client.New(cfg)
	if nil != err {
		return err
	}

	api := client.NewKeysAPI(cli)

	//取当前已存在
	arrResp, err := api.Get(context.Background(), discoverCfg.Root, &client.GetOptions{Recursive: true})
	if nil != err {
		se, ok := err.(client.Error)
		if !ok || client.ErrorCodeKeyNotFound != se.Code {
			return err
		}
	}
	lastIndex := uint64(0)
	if nil != arrResp && nil != arrResp.Node {
		if false == arrResp.Node.Dir {
			return fmt.Errorf("Get exist node failed, pathRoot is not a directory, %+v", arrResp.Node)
		}
		lastIndex = arrResp.Index
		nodes := parseEndpointNodes(arrResp.Node)
		for _, node := range nodes {
			logs.Debug("Get Node %v", node.Key)
			service, err := decodeValue(node.Value)
			if nil != err {
				return err
			}
			for checkPath, handle := range handles {
				if strings.HasPrefix(string(node.Key), checkPath) {
					handle(EventPut, string(node.Key), service)
				}
			}
		}
	}

	//先开始接收
	watcher := api.Watcher(discoverCfg.Root, &client.WatcherOptions{Recursive: true, AfterIndex: lastIndex})
	watcherContext, cancel := context.WithCancel(context.Background())

	isClosed := false
	go func() {
		<-stop
		isClosed = true
		cancel()
	}()

	for {
		resp, err := watcher.Next(watcherContext)
		if nil != err {
			if isClosed {
				logs.Debug("watcher.Next isClosed")
				break
			}
			se, ok := err.(client.Error)
			if ok && client.ErrorCodeEventIndexCleared == se.Code {
				logs.Warn("Discover Watcher, get next event over index, redo, error: %+v", se)
				watcher = api.Watcher(discoverCfg.Root, &client.WatcherOptions{Recursive: true, AfterIndex: se.Index - 5})
				continue
			}
			logs.Error("Discover Watcher, get next event error, %v", err)
			continue
		}

		et := EventUnknown
		switch resp.Action {
		case "set", "update", "create":
			et = EventPut
		case "delete":
			et = EventDel
		}
		nodes := parseEndpointNodes(resp.Node)
		for _, node := range nodes {
			logs.Debug("Get Node %v", node.Key)
			var service *Service
			if EventPut == et {
				service, err = decodeValue(node.Value)
				if nil != err {
					logs.Error("Discover Watcher, decode node error, %v, ...%+v", err, node.Value)
					continue
				}
			}
			for checkPath, handle := range handles {
				logs.Debug("Get Node checkPath %v", checkPath)
				if strings.HasPrefix(string(node.Key), checkPath) {
					handle(et, string(node.Key), service)
				}
			}
		}
	}
	return nil
}

func parseEndpointNodes(n *client.Node) []*client.Node {
	if false == n.Dir {
		return []*client.Node{n}
	}
	ns := make([]*client.Node, 0)
	for _, node := range n.Nodes {
		ns = append(ns, parseEndpointNodes(node)...)
	}
	return ns
}
