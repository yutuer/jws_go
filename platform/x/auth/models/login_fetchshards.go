package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"strings"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"

	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"sort"

	"vcs.taiyouxi.net/platform/x/auth/config"
)

type ShardInfo struct {
	Sid       uint
	Gid       uint
	ShowState string
}
type ShardInfoForClient struct {
	Name        string `json:"name"`
	DisplayName string `json:"dn"`
	ShowState   string `json:"ss"`
	state       string
	MultiLang   string `json:"multi_lang"`
	Lang        string `json:"lang"`
}

func (info ShardInfoForClient) GetState() string {
	return info.state
}

type ShardHasRole struct {
	Shard     string `json:"shard"`
	RoleLevel string `json:"level"`
}

var (
	shardId_Info      map[string]ShardInfo
	gid_sid_shardId   map[string]string
	gid_shards_Info   map[int][]ShardInfoForClient
	shards_Info_mutex sync.RWMutex
	gid_version       map[string]string // gid对应的服务器列表版本， 版本不一致的时候才去更新服务器列表
)

const shards_Info_Update_Time_In = 60

func InitShardInfo() error {
	err := initGidVersion()
	if err != nil {
		return err
	}
	_, _, _, err = updateAllShardsByGids(config.Cfg.Gids)
	if err != nil {
		return err
	}
	startGetShardInfo()
	return nil
}

func initGidVersion() error {
	gid_version = make(map[string]string)
	nowTime := time.Now().Unix()
	for _, gid := range config.Cfg.Gids {
		versionKey := getServerVersionKey(gid)
		isExist := etcd.KeyExist(versionKey)
		if !isExist {
			versionValue := fmt.Sprintf("%d", nowTime)
			err := etcd.Set(versionKey, versionValue, 0)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func startGetShardInfo() {
	go func() {
		loopTimer := time.After(time.Second * shards_Info_Update_Time_In)
		for {
			select {
			case <-loopTimer:
				loopTimer = time.After(time.Second * shards_Info_Update_Time_In)
				func() {
					defer logs.PanicCatcherWithInfo("update all shard info err")
					changedGids := GetCanUpdateGids(config.Cfg.Gids)
					if len(changedGids) == 0 {
						return
					}
					_, _, _, err := updateAllShardsByGids(changedGids)
					if err != nil {
						logs.Error("UpdateAllShardInfo Err by %s", err.Error())
					}
				}()
			}
		}
	}()
}

// FetchShardsInfo 通过game id获取当前所有分服的明细
func FetchShardsInfo(gid int) ([]ShardInfoForClient, error) {
	shards_Info_mutex.RLock()
	defer shards_Info_mutex.RUnlock()
	sinfo, sok := gid_shards_Info[gid]
	if !sok {
		return []ShardInfoForClient{}, errors.New("NoInfo")
	}
	return sinfo[:], nil
}

// GetShardID 通过shardname获取shard id(sid)
func GetShardID(name string) (sid, gid uint, showState string, err error) {
	shards_Info_mutex.RLock()
	info, ok := shardId_Info[name]
	shards_Info_mutex.RUnlock()

	if !ok {
		return 0, 0, "", errors.New(fmt.Sprintf("[GetShardID] shard %s NoInfo", name))
	}
	return info.Sid, info.Gid, info.ShowState, nil
}

func GetShardName(sid, gid uint) (shardName string, err error) {
	gs_str := gid_sid_str(gid, sid)
	shards_Info_mutex.RLock()
	sh, ok := gid_sid_shardId[gs_str]
	shards_Info_mutex.RUnlock()
	if !ok {
		return "", errors.New(fmt.Sprintf("[GetShardName] sid %d gid %d NoInfo", sid, gid))
	}
	return sh, nil
}

func GetUserShardHasRole(uid string) ([]ShardHasRole, error) {
	res, err := db_interface.GetUserShardInfo(uid)
	if err != nil {
		return nil, err
	}
	logs.Debug("[GetUserShardHasRole] get uid %s from db info: %v", uid, res)
	shs := make([]ShardHasRole, 0, len(res))
	for _, r := range res {
		if r.HasRole != "" {
			gid, sid, err := parse_gid_sid_str(r.GidSid)
			if err != nil {
				logs.Warn("[GetUserShardHasRole] parse_gid_sid_str err %s", err.Error())
				continue
			}
			shardName, err := GetShardName(sid, gid)
			if err != nil {
				logs.Warn("[GetUserShardHasRole] GetShardName err %s", err.Error())
				continue
			}
			shs = append(shs, ShardHasRole{shardName, r.HasRole})
		}
	}
	logs.Debug("[GetUserShardHasRole] GetRoles %v", shs)
	return shs, nil
}

func SetUserShardHasRole(gid, sid int, uid string, hasNewRole string) error {
	return db_interface.SetUserShardInfo(uid, gid_sid_str(uint(gid), uint(sid)), hasNewRole)
}

func GetUserLastShard(uid string) (string, error) {
	gidsid, err := db_interface.GetLastLoginShard(uid)
	if err != nil {
		return "", err
	}
	gid, sid, err := parse_gid_sid_str(gidsid)
	if err != nil {
		logs.Warn("[GetUserLastShard] parse_gid_sid_str err %s", err.Error())
		return "", err
	}
	shardName, err := GetShardName(sid, gid)
	if err != nil {
		logs.Warn("[GetUserLastShard] GetShardName err %s", err.Error())
		return "", err
	}
	return shardName, nil
}

func SetUserLastShard(uid string, gid, sid int) error {
	return db_interface.SetLastLoginShard(uid, gid_sid_str(uint(gid), uint(sid)))
}

func SetUserPayFeedBackIfNot(uid string) (error, bool) {
	return db_interface.SetPayFeedBackIfNot(uid)
}

func gid_sid_str(gid, sid uint) string {
	return fmt.Sprintf("%d:%d", gid, sid)
}

func parse_gid_sid_str(gid_sid_str string) (gid, sid uint, err error) {
	ss := strings.Split(gid_sid_str, ":")
	if len(ss) < 2 {
		return 0, 0, errors.New(fmt.Sprintf("[parse_gid_sid_str] gid_sid err: %s", gid_sid_str))
	}
	g, err := strconv.Atoi(ss[0])
	if err != nil {
		return 0, 0, errors.New(fmt.Sprintf("[parse_gid_sid_str] err: %v", err))
	}
	s, err := strconv.Atoi(ss[1])
	if err != nil {
		return 0, 0, errors.New(fmt.Sprintf("[parse_gid_sid_str] err: %v", err))
	}
	return uint(g), uint(s), nil
}

func GetAllShard(gid uint) ([]etcd.ShardInfo, error) {
	shards_Info_mutex.Lock()
	defer shards_Info_mutex.Unlock()

	var shards shardSlice

	shards, err := getShardsFromEtcd()
	if err != nil {
		return nil, err
	}
	sort.Sort(shards)
	ret := make([]etcd.ShardInfo, 0)
	for _, shard := range shards {
		if shard.Gid == gid {
			ret = append(ret, shard)
		}
	}
	return ret, nil
}

func getShardsFromRedis() ([]etcd.ShardInfo, error) {
	re := make([]etcd.ShardInfo, 0, 32)
	re_m := make(map[string]etcd.ShardInfo, 32)
	db := loginRedisPool.Get()
	defer db.Close()

	shards_info, err := redis.StringMap(db.Do("HGETALL", "cfg:shards"))
	if err != nil {
		return re, err
	}
	client_shards_info, err := redis.StringMap(db.Do("HGETALL", "cfg:client_shards"))
	if err != nil {
		return re, err
	}

	if len(shards_info) == 0 || len(client_shards_info) == 0 {
		logs.Error("getShardsFromRedis Err by No Shards!")
		return re, errors.New("NoShards")
	}

	//	logs.Trace("getShardsFromRedis shards_info %v", shards_info)
	//	logs.Trace("getShardsFromRedis client_shards_info %v", client_shards_info)

	for key, json_data := range shards_info {
		var data struct {
			ID, GID uint
		}
		if err := json.Unmarshal([]byte(json_data), &data); err != nil {
			return re, err
		}
		re_m[key] = etcd.ShardInfo{
			Sid:   data.ID,
			Gid:   data.GID,
			SName: key,
			State: etcd.StateOnline,
		}
	}

	for key, json_data := range client_shards_info {
		var shards []ShardInfoForClient
		if err := json.Unmarshal([]byte(json_data), &shards); err != nil {
			return re, fmt.Errorf("err: %s, str:%s", err.Error(), json_data)
		}
		_, err := strconv.Atoi(key)
		if err != nil {
			return re, fmt.Errorf("key err: %s, str:%s", err.Error(), key)
		}
		for _, shard := range shards {
			_info, ok := re_m[shard.Name]
			if ok {
				_info.DName = shard.DisplayName
				_info.ShowState = shard.ShowState
				re_m[shard.Name] = _info
			} else {
				logs.Error("getShardsFromRedis shard name %s not found ", shard.Name)
			}
		}
	}
	for _, v := range re_m {
		if _shard_valid(v) {
			re = append(re, v)
		}
	}
	return re, nil
}

func getShardsFromEtcd() ([]etcd.ShardInfo, error) {
	re := make([]etcd.ShardInfo, 0, 32)
	gids, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/", config.Cfg.EtcdRoot))
	if err != nil {
		return re, err
	}
	for _, gid := range gids {
		pg := strings.Split(gid, "/")
		igid, err := strconv.Atoi(pg[len(pg)-1])
		if err != nil {
			continue
		}

		sids, err := etcd.GetAllSubKeys(gid)
		if err != nil {
			return re, err
		}

		for _, sid := range sids {
			ps := strings.Split(sid, "/")
			if len(ps) > 3 {
				isid, err := strconv.Atoi(ps[3])
				if err != nil {
					continue
				}

				kv, err := etcd.GetAllSubKeyValue(sid)

				ip, _ := kv[etcd.KeyIp]
				sn, _ := kv[etcd.KeySName]
				dn, _ := kv[etcd.KeyDName]
				state, _ := kv[etcd.KeyState]
				showState, _ := kv[etcd.KeyShowState]
				multiLang, _ := kv[etcd.MultiLang]
				lang, _ := kv[etcd.Language]

				if sn == "" {
					continue
				}
				// 在服务器曾经成功启动过并操作过上线后，此服务器就可以在服务器列表中出现
				// etcd.StateOnline 这个检查由外面做，若是超级账号都能看到
				// || state != etcd.StateOnline

				if dn == "" {
					dn = fmt.Sprintf("%d", isid)
				}
				shard := etcd.ShardInfo{
					Gid:       uint(igid),
					Sid:       uint(isid),
					Ip:        ip,
					SName:     sn,
					DName:     dn,
					State:     state,
					ShowState: showState,
					MultiLang: multiLang,
					Language:  lang,
				}
				if _shard_valid(shard) {
					re = append(re, shard)
				}
			} else {
				logs.Error("sid format Err %s", sid)
			}
		}
	}
	return re, nil
}

func getShardsFromEtcdByGids(gids []string) ([]etcd.ShardInfo, error) {
	re := make([]etcd.ShardInfo, 0, 32)

	for _, gid := range gids {
		intGid, err := strconv.Atoi(gid)
		if err != nil {
			return nil, err
		}

		etcdGid := fmt.Sprintf("%s/%s", config.Cfg.EtcdRoot, gid)
		sids, err := etcd.GetAllSubKeys(etcdGid)
		if err != nil {
			return re, err
		}

		for _, sid := range sids {
			ps := strings.Split(sid, "/")
			if len(ps) > 3 {
				isid, err := strconv.Atoi(ps[3])
				if err != nil {
					continue
				}

				kv, err := etcd.GetAllSubKeyValue(sid)

				ip, _ := kv[etcd.KeyIp]
				sn, _ := kv[etcd.KeySName]
				dn, _ := kv[etcd.KeyDName]
				state, _ := kv[etcd.KeyState]
				showState, _ := kv[etcd.KeyShowState]
				multiLang, _ := kv[etcd.MultiLang]
				lang, _ := kv[etcd.Language]

				if sn == "" {
					continue
				}
				// 在服务器曾经成功启动过并操作过上线后，此服务器就可以在服务器列表中出现
				// etcd.StateOnline 这个检查由外面做，若是超级账号都能看到
				// || state != etcd.StateOnline

				if dn == "" {
					dn = fmt.Sprintf("%d", isid)
				}
				shard := etcd.ShardInfo{
					Gid:       uint(intGid),
					Sid:       uint(isid),
					Ip:        ip,
					SName:     sn,
					DName:     dn,
					State:     state,
					ShowState: showState,
					MultiLang: multiLang,
					Language:  lang,
				}
				if _shard_valid(shard) {
					re = append(re, shard)
				}
			} else {
				logs.Error("sid format Err %s", sid)
			}
		}
	}
	return re, nil
}

func _shard_valid(s etcd.ShardInfo) bool {
	show_state, err := strconv.Atoi(s.ShowState)
	if err != nil {
		show_state = etcd.ShowState_Count // 没配showstate的话，就给个不存在的值
	}
	if !config.Cfg.InnerExperience { // auth 不是体验服
		if show_state == etcd.ShowState_Experience {
			return false
		}
	}
	return true
}

type shardSlice []etcd.ShardInfo

func (s shardSlice) Len() int {
	return len(s)
}

func (s shardSlice) Less(i, j int) bool {
	// 正常的逻辑：按照sid升序排列
	return s[i].Sid < s[j].Sid
}

func (s shardSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func GetCanUpdateGids(gids []string) []string {
	versionsFromEtcd := GetUpdatedGidVersion(gids)
	logs.Info("update server list from etcd %v", versionsFromEtcd)
	logs.Info("local server list %v", gid_version)
	ret := make([]string, 0)
	for i, ver := range versionsFromEtcd {
		gid := gids[i]
		localVersion := gid_version[gid]
		if localVersion != ver {
			ret = append(ret, gid)
		}
	}
	return ret
}

// 获取指定的GID的服务器列表版本 按顺序返回版本号
func GetUpdatedGidVersion(gids []string) []string {
	ret := make([]string, 0)
	for _, gid := range gids {
		etcdVersionKey := getServerVersionKey(gid)
		version, err := etcd.Get(etcdVersionKey)
		if err != nil {
			logs.Error("get server version from etcd err", err)
			return nil
		}
		ret = append(ret, version)
	}
	return ret
}

func getServerVersionKey(gid string) string {
	return fmt.Sprintf("%s/%s/%s", config.Cfg.EtcdRoot, gid, etcd.ServerListVersion)
}

func updateAllShardsByGids(gids []string) (_gid_shards map[int][]ShardInfoForClient,
	_shardId_Info map[string]ShardInfo,
	_gid_sid_shardId map[string]string,
	err error) {
	_gid_shards = make(map[int][]ShardInfoForClient, 16)
	_shardId_Info = make(map[string]ShardInfo, 16)
	_gid_sid_shardId = make(map[string]string, 16)

	shards, err := getShards(gids)
	if err != nil {
		return nil, nil, nil, err
	}
	logs.Info("getAllShards gids=%v, shards=%v", gids, shards)

	for _, shard := range shards {
		info, ok := _gid_shards[int(shard.Gid)]
		if !ok {
			info = make([]ShardInfoForClient, 0, 32)
			_gid_shards[int(shard.Gid)] = info
		}
		info = _gid_shards[int(shard.Gid)]
		info = append(info, ShardInfoForClient{
			Name:        shard.SName,
			DisplayName: shard.DName,
			ShowState:   shard.ShowState,
			state:       shard.State,
			Lang:        shard.Language,
			MultiLang:   shard.MultiLang,
		})
		_gid_shards[int(shard.Gid)] = info

		_shardId_Info[shard.SName] = ShardInfo{
			Sid:       shard.Sid,
			Gid:       shard.Gid,
			ShowState: shard.ShowState,
		}
		_gid_sid_shardId[gid_sid_str(shard.Gid, shard.Sid)] = shard.SName
	}

	newVersions := GetUpdatedGidVersion(gids)

	updateShardInfo(_gid_shards, _shardId_Info, _gid_sid_shardId, gids, newVersions)
	logs.Debug("new shards info = %v, \n shardId = %v, \n nameInfo = %v", gid_shards_Info, gid_sid_shardId, shardId_Info)
	return _gid_shards, _shardId_Info, _gid_sid_shardId, nil
}

func updateShardInfo(_gid_shards map[int][]ShardInfoForClient, _shardId_Info map[string]ShardInfo,
	_gid_sid_shardId map[string]string, gids, newVersions []string) {
	shards_Info_mutex.Lock()
	defer shards_Info_mutex.Unlock()

	if shardId_Info == nil {
		shardId_Info = _shardId_Info
	} else {
		for key, value := range _shardId_Info {
			shardId_Info[key] = value
		}
	}

	if gid_sid_shardId == nil {
		gid_sid_shardId = _gid_sid_shardId
	} else {
		for key, value := range _gid_sid_shardId {
			gid_sid_shardId[key] = value
		}
	}

	if gid_shards_Info == nil {
		gid_shards_Info = _gid_shards
	} else {
		for key, value := range _gid_shards {
			gid_shards_Info[key] = value
		}
	}

	for i := range gids {
		gid_version[gids[i]] = newVersions[i]
	}
}

func getShards(gids []string) (shardSlice, error) {
	var shards shardSlice
	var err error
	if !config.Cfg.IsRunModeLocal() {
		// from etcd
		shards, err = getShardsFromEtcdByGids(gids)
		if err != nil {
			return nil, err
		}
	} else {
		// from redis
		shards, err = getShardsFromRedis()
		if err != nil {
			return nil, err
		}
	}
	sort.Sort(shards)
	return shards, nil
}
