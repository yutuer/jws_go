package postService

import (
	"fmt"
	"strconv"
	"time"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func RegService(etcdRoot, serviceID, serviceUrl string, num int, ttl time.Duration) {
	r := fmt.Sprintf("%s/%s", etcdRoot, serviceID)
	err := etcd.Set(r+ServiceIDKey, serviceID, ttl)
	logs.Info("reg etcd key: %v, value: %v", r+ServiceIDKey, serviceID)
	if err != nil {
		logs.Error("reg etcd err by %v, key: %v, value: %v", err, r+ServiceIDKey, serviceID)
	}
	err = etcd.Set(r+ServiceNumKey, strconv.Itoa(num), ttl)
	if err != nil {
		logs.Error("reg etcd err by %v, key: %v, value: %v", err, r+ServiceNumKey, strconv.Itoa(num))
	}
	err = etcd.Set(r+ServiceUrlKey, serviceUrl, ttl)
	if err != nil {
		logs.Error("reg etcd err by %v, key: %v, value: %v", err, r+ServiceUrlKey, serviceUrl)
	}
}

func UnRegService(etcdRoot, serviceID string) {
	r := fmt.Sprintf("%s/%s", etcdRoot, serviceID)
	err := etcd.Delete(r + ServiceIDKey)
	logs.Info("unreg etcd key: %v", r+ServiceIDKey)
	if err != nil {
		logs.Error("unreg etcd err by %v, key: %v", err, r+ServiceIDKey)
	}
	err = etcd.Delete(r + ServiceNumKey)
	if err != nil {
		logs.Error("unreg etcd err by %v, key: %v", err, r+ServiceNumKey)
	}
	err = etcd.Delete(r + ServiceUrlKey)
	if err != nil {
		logs.Error("unreg etcd err by %v, key: %v", err, r+ServiceUrlKey)
	}
	err = etcd.DeleteDir(r)
	if err != nil {
		logs.Error("unreg etcd err by %v, key: %v", err, r)
	}
}

func RegServices(gveStartUrl, gveStopUrl string, tinc time.Duration) {
	for _, sid := range game.Cfg.MergeRel {
		serverID := fmt.Sprintf("%d", sid)
		RegService(
			GetGameStartNotifyServiceEtcdKey(game.Cfg.EtcdRoot),
			serverID,
			gveStartUrl,
			0,
			tinc*10)
		RegService(
			GetGameStopNotifyServiceEtcdKey(game.Cfg.EtcdRoot),
			serverID,
			gveStopUrl,
			0,
			tinc*10)
	}
}

func UnRegServices() {
	for _, sid := range game.Cfg.MergeRel {
		serverID := fmt.Sprintf("%d", sid)
		UnRegService(
			GetGameStartNotifyServiceEtcdKey(game.Cfg.EtcdRoot),
			serverID)
		UnRegService(
			GetGameStopNotifyServiceEtcdKey(game.Cfg.EtcdRoot),
			serverID)
	}
}

func RegTBServices(root, tbUrl string, group uint32, gid uint) {
	logs.Debug("tbUrl: %v", tbUrl)
	id := GetCSIDByGroupID(group, gid)
	RegService(
		GetGameStopNotifyServiceEtcdKey(root),
		id,
		tbUrl,
		0,
		0)
}

func RegGVGServices(root, gvgUrl string, sid uint) {
	serverID := fmt.Sprintf("%d", sid)
	RegService(
		GetGVGStopNotifyServiceEtcdKey(root),
		serverID,
		gvgUrl,
		0,
		0)
}

func GetCSIDByGroupID(groupID uint32, gid uint) string {
	return fmt.Sprintf("tb_%d_%d", gid, groupID)
}

func GetGamexIDByAcID(acID string) string {
	acc, _ := db.ParseAccount(acID)
	return strconv.Itoa(int(acc.ShardId))
}

func GetGamexIDBySID(SID int) string {
	return strconv.Itoa(SID)
}

