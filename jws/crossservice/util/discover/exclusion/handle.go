package exclusion

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Handle ..
type Handle struct {
	Key   string
	Nodes []Node
	TTL   time.Duration
}

//Node ..
type Node struct {
	Key   string
	Value string
}

//NewHandle ..
func NewHandle(key string) *Handle {
	h := &Handle{
		Key: key,
	}
	return h
}

//SetTTL ..
func (h *Handle) SetTTL(t time.Duration) {
	h.TTL = t
}

//AddNode ..
func (h *Handle) AddNode(k, v string) bool {
	ret := false
	pushFunc := func(cli client.Client) error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRequest)
		api := client.NewKeysAPI(cli)
		path := makeExclusionPath(h.Key) + "/" + k
		_, err := api.Set(ctx, path, v, &client.SetOptions{TTL: h.TTL, PrevExist: client.PrevNoExist})
		if nil != err {
			se := err.(client.Error)
			if client.ErrorCodeNodeExist != se.Code {
				logs.Warn("Publish Handle AddNode error, node %v:%v, error: %v", path, v, err)
			} else {
				logs.Info("Publish Handle AddNode, Node %v:%v be Occupy by others", path, v)
			}
		} else {
			logs.Info("Publish Handle AddNode, Occupy Node %v:%v", path, v)
			h.Nodes = append(h.Nodes, Node{Key: k, Value: v})
			ret = true
		}
		cancel()
		return nil
	}
	if err := callWithClient(pushFunc); nil != err {
		logs.Warn("Publish Handle AddNode error, %v", err)
	}
	return ret
}

//Publish ..
func (h *Handle) Publish() {
	callWithClient(h.publish)
}

func (h *Handle) publish(cli client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRequest)
	api := client.NewKeysAPI(cli)
	for _, e := range h.Nodes {
		path := makeExclusionPath(h.Key) + "/" + e.Key
		_, err := api.Set(ctx, path, e.Value, &client.SetOptions{TTL: h.TTL})
		if nil != err {
			logs.Warn("Publish Handle Set error, Node %+v, error: %v", e, err)
			continue
		}
	}
	cancel()
	return nil
}

//UnPublish ..
func (h *Handle) UnPublish() {
	callWithClient(h.unPublish)
}

func (h *Handle) unPublish(cli client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRequest)
	api := client.NewKeysAPI(cli)
	for _, e := range h.Nodes {
		path := makeExclusionPath(h.Key) + "/" + e.Key
		_, err := api.Delete(ctx, path, &client.DeleteOptions{Recursive: true})
		if nil != err {
			logs.Warn("Publish Handle Delete error, Node %+v, error: %v", e, err)
			continue
		}
	}
	cancel()
	return nil
}
