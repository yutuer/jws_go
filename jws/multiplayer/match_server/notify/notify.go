package notify

import (
	"math/rand"

	"sync"

	"vcs.taiyouxi.net/jws/multiplayer/helper"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/match_server/config"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
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
	service     postService.ServiceMng
	matchTokens string
}

func (n *NotifyMultiplayServer) Start(matchToken string) {
	n.matchTokens = helper.FmtMatchToken(matchToken)

	n.service.Init(
		postService.GetByGID(
			postService.MultiplayServiceEtcdKey,
			multConfig.Cfg.EtcdRoot,
			multConfig.Cfg.GID,
			n.matchTokens,
		),
		func(s []postService.Service) int {
			if s == nil || len(s) == 0 {
				return -1
			} else {
				return rand.Intn(len(s))
			}
		},
		multConfig.Cfg.EtcdEndpoint)
	n.service.Start()

}

func (n *NotifyMultiplayServer) Stop() {

	n.service.Stop()

}

func (n *NotifyMultiplayServer) Notify(matchRes helper.MatchGameInfo) {
	go func() {
		//YZH 因为GvE三英讨逆的这个匹配是单向推送的，这里如果Post出问题是会影响所有人的。因此go 出去
		_, err := n.service.PostBySelect(matchRes)
		if err != nil {
			logs.Error("NotifyMultiplayServer Notify failed: %s", err.Error())
		}
	}()
}
