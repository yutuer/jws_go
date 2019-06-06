package multiplay_util

import (
	"math/rand"

	"sync"

	"fmt"

	csCfg "vcs.taiyouxi.net/jws/crossservice/config"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
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
	teamboss    postService.ServiceMng
	matchTokens string
}

func (n *NotifyMultiplayServer) Start(matchToken string) {
	n.matchTokens = matchToken
	n.teamboss.Init(
		postService.GetByGID(
			postService.TBServiceEtcdKey,
			csCfg.Cfg.EtcdRoot,
			fmt.Sprintf("%d", csCfg.Cfg.Gid),
			n.matchTokens,
		),
		func(s []postService.Service) int {
			if s == nil || len(s) == 0 {
				return -1
			} else {
				return rand.Intn(len(s))
			}
		},
		csCfg.Cfg.EtcdEndPoint)
	n.teamboss.Start()

}

func (n *NotifyMultiplayServer) Stop() {
	n.teamboss.Stop()
}

func (n *NotifyMultiplayServer) TeamBossNotify(data interface{}) ([]byte, error) {
	b, err := n.teamboss.PostBySelect(data)
	return b, err
}
