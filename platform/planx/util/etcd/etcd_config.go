package etcd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"strconv"

	"sync"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var etcdClient *client.Client = nil

var (
	ErrByNoInit = errors.New("ErrByNoInit")
)

//初始化etcd
func InitClient(ends []string) error {
	etcdClient = nil
	cfg := client.Config{
		Endpoints:               ends,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: 10 * time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		logs.Error("Init Etcd Client Err by %s", err.Error())
		return err
	} else {
		etcdClient = &c
	}

	return nil
}

func KeyExist(key string) bool {
	if etcdClient == nil {
		logs.Error("Etcd Client No Init Err")
		return false
	}
	kapi := client.NewKeysAPI(*etcdClient)
	_, err := kapi.Get(context.Background(), key, &client.GetOptions{})
	return err == nil
}

// "/a3k/1/10/GameConfig/server_start_time"
func Get(key string) (string, error) {
	if etcdClient == nil {
		logs.Error("Etcd Client No Init Err")
		return "", ErrByNoInit
	}
	kapi := client.NewKeysAPI(*etcdClient)
	resp, err := kapi.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		se, ok := err.(client.Error)
		if !ok || client.ErrorCodeKeyNotFound != se.Code {
			logs.Error("Get err:%v", err)
		}
		return "", err
	} else {
		return resp.Node.Value, nil
	}
}

func Gets(keys []string, g_err error) ([]string, error) {
	if g_err != nil {
		return []string{}, g_err
	}

	res := make([]string, 0, len(keys))
	for _, key := range keys {
		v, err := Get(key)
		if err != nil {
			return []string{}, err
		}
		res = append(res, v)
	}
	return res, nil
}

func GetAllSubKeyValue(key string) (map[string]string, error) {
	res := make(map[string]string)
	keys, err := GetAllSubKeys(key)
	if err != nil {
		return res, err
	}

	values, err := Gets(keys, err)
	if err != nil {
		return res, err
	}

	for idx, k := range keys {
		paths := strings.Split(k, "/")
		res[paths[len(paths)-1]] = values[idx]
	}

	logs.Trace("GetAllSubKeyValue %s --> %v", key, res)

	return res, nil

}

func GetAllSubKeys(key string) ([]string, error) {
	if etcdClient == nil {
		logs.Error("Etcd Client No Init Err")
		return nil, ErrByNoInit
	}

	kapi := client.NewKeysAPI(*etcdClient)
	resp, err := kapi.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		logs.Info("Etcd Client kapi.Get %s", err.Error())
		return nil, err
	} else {
		res := []string{}
		for _, node := range resp.Node.Nodes {
			if node != nil {
				res = append(res, node.Key)
			}
		}
		return res, nil
	}
}

func FindKeyFromNodes(nodes *client.Node, key string) *client.Node {
	for _, node := range nodes.Nodes {
		if node.Key == key {
			return node
		}
	}
	return nil
}

func GetAllSubKeysRcs(key string) (*client.Node, error) {
	if etcdClient == nil {
		logs.Error("Etcd Client No Init Err")
		return nil, ErrByNoInit
	}

	kapi := client.NewKeysAPI(*etcdClient)
	resp, err := kapi.Get(context.Background(), key, &client.GetOptions{Recursive: true})
	if err != nil {
		logs.Info("Etcd Client kapi.Get %s", err.Error())
		return nil, err
	} else {
		return resp.Node, nil
	}
}

func GetByServer(gid, sid, key string) (string, error) {
	if etcdClient == nil {
		logs.Error("Etcd Client No Init Err")
		return "", ErrByNoInit
	}

	kapi := client.NewKeysAPI(*etcdClient)
	url := ""
	if key[0] == '/' {
		url = fmt.Sprintf("/a3k/%s/%s%s", gid, sid, key)
	} else {
		url = fmt.Sprintf("/a3k/%s/%s/%s", gid, sid, key)
	}
	resp, err := kapi.Get(context.Background(), url, &client.GetOptions{})
	if err != nil {
		logs.Error("Etcd Client kapi.Get %s", err.Error())
		return "", err
	} else {
		return resp.Node.Value, nil
	}
}

func Set(key, value string, ttl time.Duration) error {
	if etcdClient == nil {
		logs.Error("Etcd Client No Init Err")
		return ErrByNoInit
	}

	kapi := client.NewKeysAPI(*etcdClient)
	_, err := kapi.Set(context.Background(), key, value, &client.SetOptions{TTL: ttl})
	if err != nil {
		logs.Error("Etcd Client kapi.Set %s", err.Error())
		return err
	}
	return nil
}

func Delete(key string) error {
	if etcdClient == nil {
		logs.Error("Etcd Client No Init Err")
		return ErrByNoInit
	}
	kapi := client.NewKeysAPI(*etcdClient)
	_, err := kapi.Delete(context.Background(), key, &client.DeleteOptions{})
	if err != nil {
		logs.Error("Etcd Client kapi.Delete %s", err.Error())
		return err
	}
	return nil
}

func DeleteDir(key string) error {
	if etcdClient == nil {
		logs.Error("Etcd Client No Init Err")
		return ErrByNoInit
	}
	kapi := client.NewKeysAPI(*etcdClient)
	_, err := kapi.Delete(context.Background(), key, &client.DeleteOptions{Dir: true})
	if err != nil {
		logs.Error("Etcd Client kapi.Delete %s", err.Error())
		return err
	}
	return nil
}

// "/a4k/"
func GetServerGidSid(root string) (map[int][]int, error) {
	gids, err := GetAllSubKeys(root)
	if err != nil {
		return nil, err
	}
	res := make(map[int][]int, len(gids))
	for _, gid := range gids {
		pg := strings.Split(gid, "/")
		igid, err := strconv.Atoi(pg[len(pg)-1])
		if err != nil {
			continue
		}

		sids, err := GetAllSubKeys(gid)
		if err != nil {
			return nil, err
		}

		ss, ok := res[igid]
		if !ok {
			ss = make([]int, 0, len(sids))
		}

		for _, sid := range sids {
			ps := strings.Split(sid, "/")
			if len(ps) > 3 {
				isid, err := strconv.Atoi(ps[3])
				if err != nil {
					continue
				}
				ss = append(ss, isid)
				res[igid] = ss
			}
		}
	}
	return res, nil
}

func FindServerGidBySid(root string, sid uint) (gid int, err error) {
	gid2sid, err := GetServerGidSid(root)
	if err != nil {
		return -1, err
	} else {
		for g, sids := range gid2sid {
			for _, s := range sids {
				if s == int(sid) {
					gid = g
					return gid, nil
				}
			}
		}
	}
	return -1, nil
}

//模块内部缓冲变量, 减少sid转换显示名称时对etcd的访问
var shareSid2DN map[uint]string
var shareSid2DNLock sync.RWMutex

func GetSidDisplayName(etcdRoot string, gid, sid uint) string {
	shareSid2DNLock.RLock()
	if shareSid2DN == nil {
		shareSid2DN = make(map[uint]string)
	}
	if sdn, ok := shareSid2DN[sid]; !ok {
		shareSid2DNLock.RUnlock()

		shareSid2DNLock.Lock()
		key := fmt.Sprintf("%s/%d/%d/%s", etcdRoot, gid, sid, KeyDName)
		dn, _ := Get(key)
		shareSid2DN[sid] = dn
		shareSid2DNLock.Unlock()
		return dn
	} else {
		shareSid2DNLock.RUnlock()
		return sdn
	}
}

func ParseDisplayShardName(sidStr string) string {
	indexDot := strings.Index(sidStr, ".")
	if indexDot == -1 || indexDot+1 >= len(sidStr) {
		return sidStr
	}
	return sidStr[indexDot+1:]
}

func ParseDisplayShardName2(etcdRoot string, gid uint, sid uint) string {
	displayName1 := GetSidDisplayName(etcdRoot, gid, sid)
	displayName2 := ParseDisplayShardName(displayName1)
	index := strings.Index(displayName2, "-")
	serverName := ""
	if index == -1 {
		serverName = displayName2
	} else {
		serverName = displayName2[:index]
	}
	return serverName
}
