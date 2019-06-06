package postService

import "fmt"

const (
	MultiplayServiceEtcdKey        = "%s/%s/%s/multiplayServices"
	FenghuoServiceEtcdKey          = "%s/%s/%s/fenghuoServices"
	TBServiceEtcdKey               = "%s/%s/%s/tb_services"
	gameStartNotifyServiceEtcdKey  = "%s/gameStartNotifyServices"
	gameStopNotifyServiceEtcdKey   = "%s/gameStopNotifyServices"
	RedisStorageWarmServiceEtcdKey = "%s/warmServices"
	GVGServiceEtcdKey              = "%s/%s/%s/gvg_services"
	GVGStopNotifyEtcdKey           = "%s/gvgStopNotifyEtcdKey"
)

func GetByGID(key, etcdRoot, gid, servicetoken string) string {
	return fmt.Sprintf(key, etcdRoot, gid, servicetoken)
}

func GetGameStartNotifyServiceEtcdKey(etcdRoot string) string {
	return fmt.Sprintf(
		gameStartNotifyServiceEtcdKey,
		etcdRoot)
}

func GetGameStopNotifyServiceEtcdKey(etcdRoot string) string {
	return fmt.Sprintf(
		gameStopNotifyServiceEtcdKey,
		etcdRoot)
}

func GetGVGStopNotifyServiceEtcdKey(etcdRoot string) string {
	return fmt.Sprintf(
		GVGStopNotifyEtcdKey,
		etcdRoot)
}
