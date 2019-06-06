package notify

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/modules/gve_notify/post_data"
	multiHelper "vcs.taiyouxi.net/jws/multiplayer/helper"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/config"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type NotifyToLogicServer struct {
	url            string
	startService   postService.ServiceMng
	stopService    postService.ServiceMng
	gvgStopService postService.ServiceMng
}

func (n *NotifyToLogicServer) Start() {
	n.url = fmt.Sprintf("ws://%s/ws", multConfig.Cfg.PublicIP)
	n.startService.Init(
		postService.GetGameStartNotifyServiceEtcdKey(multConfig.Cfg.EtcdRoot),
		postService.NilFunc, multConfig.Cfg.EtcdEndpoint)
	n.startService.Start()
	n.stopService.Init(
		postService.GetGameStopNotifyServiceEtcdKey(multConfig.Cfg.EtcdRoot),
		postService.NilFunc, multConfig.Cfg.EtcdEndpoint)
	n.stopService.Start()

	n.gvgStopService.Init(
		postService.GetGVGStopNotifyServiceEtcdKey(multConfig.Cfg.EtcdRoot),
		postService.NilFunc, multConfig.Cfg.EtcdEndpoint)
	n.gvgStopService.Start()
}

func (n *NotifyToLogicServer) Stop() {
	n.startService.Stop()
	n.stopService.Stop()
}

func (n *NotifyToLogicServer) GameStart(acID, gameID string, count int, serviceID string) (*post_data.StartGVEPostResData, error) {
	resData, err := n.startService.PostById(serviceID,
		multiHelper.GameStartInfo{
			AcIDs:      acID,
			GameID:     gameID,
			MServerUrl: n.url,
			Secret:     "123456",
			FightCount: count,
		})
	if err != nil {
		logs.Error("NotifyToLogicServer.NotifyToLogicServer PostById err %s", err.Error())
		return nil, err
	}
	res := &post_data.StartGVEPostResData{}
	err = json.Unmarshal(resData, res)
	if err != nil {
		logs.Error("NotifyToLogicServer.NotifyToLogicServer json.Unmarshal err %s", err.Error())
	}
	return res, err
}

func (n *NotifyToLogicServer) GameStop(data interface{}, serviceID string) error {
	_, err := n.stopService.PostById(serviceID, data)
	return err
}

func (n *NotifyToLogicServer) GVGGameStop(data interface{}, serviceID string) error {
	_, err := n.gvgStopService.PostById(serviceID, data)
	return err
}

var notify NotifyToLogicServer

func GetNotify() *NotifyToLogicServer {
	return &notify
}
