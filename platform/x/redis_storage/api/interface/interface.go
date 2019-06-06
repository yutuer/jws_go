package redisStorageApiInterface

import (
	"fmt"

	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
)

func GetServiceKeyID(shardID, redisAddr string, num int) string {
	return fmt.Sprintf("%s", shardID)
}

type WarmData struct {
	Key string `json:"key"`
}

type WarmApi struct {
	shardID int
	service postService.ServiceMng
}

func (w *WarmApi) Init(etcdRoot string, shardID int, endPoint []string) {
	w.shardID = shardID
	w.service.Init(
		fmt.Sprintf(
			postService.RedisStorageWarmServiceEtcdKey,
			etcdRoot),
		postService.NilFunc, endPoint)
}

func (w *WarmApi) Start() {
	w.service.Start()
}

func (w *WarmApi) Stop() {
	w.service.Stop()
}

func (w *WarmApi) SId() int {
	return w.shardID
}

func (w *WarmApi) WarmKey(serviceKey string, key string) error {
	_, err := w.service.PostById(serviceKey, WarmData{
		Key: key,
	})
	return err
}
