package publish

import (
	"sync"
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Handle ..
type Handle struct {
	Parent string

	lockChildren sync.Mutex
	children     map[string]Elem

	TTL time.Duration
}

//Elem ..
type Elem struct {
	Key   string
	Value string
}

//NewHandle ..
func NewHandle(parent string) *Handle {
	h := &Handle{
		Parent:   parent,
		children: make(map[string]Elem),
	}
	return h
}

func (h *Handle) addElem(k, v string) {
	h.lockChildren.Lock()
	defer h.lockChildren.Unlock()
	h.children[k] = Elem{Key: k, Value: v}
}

func (h *Handle) delElem(k string) {
	h.lockChildren.Lock()
	defer h.lockChildren.Unlock()
	delete(h.children, k)
}

//SetTTL ..
func (h *Handle) SetTTL(t time.Duration) {
	h.TTL = t
}

//AddElem ..
func (h *Handle) AddElem(k, v string) {
	h.addElem(k, v)
}

//Publish ..
func (h *Handle) Publish() {
	callWithClient(h.publish)
}

func (h *Handle) publish(cli client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRequest)
	api := client.NewKeysAPI(cli)
	for _, e := range h.children {
		_, err := api.Set(ctx, makeElemPath(h.Parent, e), e.Value, &client.SetOptions{TTL: h.TTL})
		if nil != err {
			logs.Warn("Publish Handle Set error, Elem %+v, error: %v", e, err)
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
	_, err := api.Delete(ctx, makeParentPath(h.Parent), &client.DeleteOptions{Recursive: true})
	if nil != err {
		se := err.(client.Error)
		if client.ErrorCodeKeyNotFound != se.Code {
			logs.Warn("Publish Handle Delete error, Parent %+v, error: %v", h.Parent, err)
		}
	}
	cancel()
	return nil
}

//PushAdd ..
func (h *Handle) PushAdd(k, v string) {
	h.addElem(k, v)
	elem := Elem{Key: k, Value: v}
	pushFunc := func(cli client.Client) error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRequest)
		api := client.NewKeysAPI(cli)
		_, err := api.Set(ctx, makeElemPath(h.Parent, elem), elem.Value, &client.SetOptions{TTL: h.TTL})
		if nil != err {
			logs.Warn("Publish Handle Set error, Elem %+v, error: %v", elem, err)
		}
		cancel()
		return nil
	}
	callWithClient(pushFunc)
}

//PushDel ..
func (h *Handle) PushDel(k, v string) {
	h.delElem(k)
	elem := Elem{Key: k, Value: v}
	pushFunc := func(cli client.Client) error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRequest)
		api := client.NewKeysAPI(cli)
		_, err := api.Delete(ctx, makeElemPath(h.Parent, elem), nil)
		if nil != err {
			logs.Warn("Publish Handle Set error, Elem %+v, error: %v", elem, err)
		}
		cancel()
		return nil
	}
	callWithClient(pushFunc)
}
