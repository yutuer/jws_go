package gvg

import (
	"math/rand"

	"sync"

	"fmt"

	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var notifys map[string]*NotifyMultiplayServer
var mux sync.Mutex

func init() {
	notifys = make(map[string]*NotifyMultiplayServer, 10)
}

func GetNotify(matchTokens string) *NotifyMultiplayServer {
	mux.Lock()
	defer mux.Unlock()
	if n, ok := notifys[matchTokens]; ok {
		return n
	} else {
		nms := new(NotifyMultiplayServer)
		nms.Start(matchTokens)
		notifys[matchTokens] = nms
		return nms
	}
	return nil
}

func StopNotifyies() {
	for _, v := range notifys {
		v.Stop()
	}
}

type NotifyMultiplayServer struct {
	gvg         postService.ServiceMng
	matchTokens string
}

func (n *NotifyMultiplayServer) Start(matchToken string) {
	n.matchTokens = matchToken
	n.gvg.Init(
		postService.GetByGID(
			postService.GVGServiceEtcdKey,
			game.Cfg.EtcdRoot,
			fmt.Sprintf("%d", game.Cfg.Gid),
			n.matchTokens,
		),
		func(s []postService.Service) int {
			if s == nil || len(s) == 0 {
				return -1
			} else {
				return rand.Intn(len(s))
			}
		},
		game.Cfg.EtcdEndPoint)
	n.gvg.Start()

}

func (n *NotifyMultiplayServer) Stop() {
	n.gvg.Stop()
}

func (n *NotifyMultiplayServer) GVGNotify(data interface{}) ([]byte, error) {
	b, err := n.gvg.PostBySelect(data)
	return b, err
}
