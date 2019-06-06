package data_ver

import (
	"fmt"

	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/profile_tool"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("getServerByVersion", getServerByVersion)
	gm_command.AddGmCommandHandle("getAllServer", getAllServer)
	gm_command.AddGmCommandHandle("setVersionHotC", setVersionHotC)
	gm_command.AddGmCommandHandle("getServerHotData", getServerHotData)
	gm_command.AddGmCommandHandle("signalServerHotC", signalServerHotC)
	gm_command.AddGmCommandHandle("signalAllServerHotC", signalAllServerHotC)
}

func getAllServer(c *gm_command.Context, server, accountid string, params []string) error {
	shardInfo := make([]string, 0, 10)
	logs.Info("getAllServer")
	for _, info := range profile_tool.GetAllServerCfg() {
		shardInfo = append(shardInfo, info.RedisName)
	}
	type res struct {
		Infos []string
	}
	sort.Sort(sort.StringSlice(shardInfo))
	var r res
	r.Infos = shardInfo
	b, _ := json.Marshal(&r)
	c.SetData(string(b))
	return nil
}

func getServerByVersion(c *gm_command.Context, server, accountid string, params []string) error {
	version := params[0]

	logs.Info("getServerByVersion %s", version)

	versions, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/%s/", config.Cfg.GameSerEtcdRoot, etcd.DirHotData))
	if err != nil {
		return err
	}
	filterGid := make(map[string]struct{}, 5)
	for _, gid := range config.Cfg.GidFilter {
		ss := strings.Split(gid, "/")
		filterGid[ss[len(ss)-1]] = struct{}{}
	}
	logs.Debug("getServerByVersion filterGid %v", filterGid)
	shardHotInfo := make([]string, 0, 64)
	for _, verpath := range versions {
		pg := strings.Split(verpath, "/")
		ver := pg[len(pg)-1]
		if ver != version {
			continue
		}
		gids, err := etcd.GetAllSubKeys(verpath + "/")
		if err != nil {
			continue
		}
		for _, gidPath := range gids {
			sids, err := etcd.GetAllSubKeys(gidPath + "/")
			if err != nil {
				continue
			}
			for _, sidPath := range sids {
				_sid := strings.Split(sidPath, "/")
				ishard, err := strconv.Atoi(_sid[len(_sid)-1])
				if err != nil {
					continue
				}
				kv, err := etcd.GetAllSubKeyValue(sidPath)
				if err != nil || len(kv) <= 0 {
					continue
				}
				gid, _ := kv[etcd.KeyHotDataGid]
				ser_n := fmt.Sprintf("%s:%d", gid, ishard)
				if config.Cfg.IsServerMerged(ser_n) {
					continue
				}
				if _, ok := filterGid[gid]; ok {
					shardHotInfo = append(shardHotInfo, ser_n)
				}
			}
		}
	}
	type res struct {
		Infos []string
	}
	sort.Sort(sort.StringSlice(shardHotInfo))
	var r res
	r.Infos = shardHotInfo
	b, _ := json.Marshal(&r)
	c.SetData(string(b))
	//c.SetData("{\"Infos\":[\"0:10\",\"0:11\",\"1:13\",\"1:14\",\"1:15\",\"1:16\",\"1:17\",\"1:18\",\"1:19\",\"1:20\",\"1:21\",\"1:22\",\"1:23\",\"1:24\",\"1:25\",\"1:26\",\"1:27\"]}")
	return nil
}

func setVersionHotC(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("setVersionHotC %v", params)

	if len(params) < 4 {
		return fmt.Errorf("param not enough")
	}

	version := params[0]
	dataC := params[1]
	shards := params[2:]

	for _, shard := range shards {
		ss := strings.Split(shard, ":")
		if len(ss) < 2 {
			continue
		}
		gid := ss[0]
		sid := ss[1]
		k := fmt.Sprintf("%s/%s/%s/%s/%s/%s",
			config.Cfg.GameSerEtcdRoot, etcd.DirHotData, version, gid,
			sid, etcd.KeyHotDataCurr)
		if err := etcd.Set(k, dataC, 0); err != nil {
			return err
		}
	}

	c.SetData("ok")
	return nil
}

func getServerHotData(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("getServerHotData")
	var versions []string
	if params[0] == "" {
		ver, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/%s/", config.Cfg.GameSerEtcdRoot, etcd.DirHotData))
		if err != nil {
			return err
		}
		versions = ver
	} else {
		versions = []string{fmt.Sprintf("%s/%s/%s", config.Cfg.GameSerEtcdRoot, etcd.DirHotData, params[0])}
	}
	filterGid := make(map[string]struct{}, 5)
	for _, gid := range config.Cfg.GidFilter {
		ss := strings.Split(gid, "/")
		filterGid[ss[len(ss)-1]] = struct{}{}
	}
	logs.Debug("getServerHotData filterGid %v", filterGid)
	shardHotInfo := make([]hotShardInfo, 0, 64)
	for _, verpath := range versions {
		pg := strings.Split(verpath, "/")
		ver := pg[len(pg)-1]

		gids, err := etcd.GetAllSubKeys(verpath + "/")
		if err != nil {
			continue
		}
		for _, gidPath := range gids {
			sids, err := etcd.GetAllSubKeys(gidPath + "/")
			if err != nil {
				continue
			}
			for _, sidPath := range sids {
				_sid := strings.Split(sidPath, "/")
				ishard, err := strconv.Atoi(_sid[len(_sid)-1])
				if err != nil {
					continue
				}
				kv, err := etcd.GetAllSubKeyValue(sidPath)
				if err != nil || len(kv) <= 0 {
					continue
				}
				baseHotBuild, _ := kv[etcd.KeyBaseDataBuild]
				hotBuild, _ := kv[etcd.KeyHotDataBuild]
				hotSeq, _ := kv[etcd.KeyHotDataSeq]
				gid, _ := kv[etcd.KeyHotDataGid]
				if _, ok := filterGid[gid]; !ok {
					continue
				}
				ser_n := fmt.Sprintf("%s:%d", gid, ishard)
				if config.Cfg.IsServerMerged(ser_n) {
					continue
				}
				VerSeq, _ := etcd.Get(fmt.Sprintf("%s/%s/%s/%s/%d/%s",
					config.Cfg.GameSerEtcdRoot, etcd.DirHotData, ver, gid,
					ishard, etcd.KeyHotDataCurr))
				shardHotInfo = append(shardHotInfo, hotShardInfo{
					Version:         ver,
					VerSeq:          VerSeq,
					GidSid:          fmt.Sprintf("%s:%d", gid, ishard),
					SidBaseHotBuild: baseHotBuild,
					SidHotBuild:     hotBuild,
					SidHotSeq:       hotSeq,
				})
			}
		}
	}
	type res struct {
		Infos []hotShardInfo
	}
	var r res
	r.Infos = shardHotInfo
	b, _ := json.Marshal(&r)
	c.SetData(string(b))
	return nil
}

func signalAllServerHotC(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("signalAllServerHotC %v", params)

	if len(params) < 3 {
		return fmt.Errorf("param not enough")
	}

	version := params[0]
	shards := params[1:]

	for _, shard := range shards {
		ss := strings.Split(shard, ":")
		if len(ss) < 2 {
			continue
		}
		serCfg := config.Cfg.GetServerCfgFromName(shard)
		if serCfg.ServerHotDataUrl == "" {
			logs.Error("server not found %v", shard)
			continue
		}
		gid := ss[0]
		sid := ss[1]
		now := time.Now().Unix()
		k := fmt.Sprintf("%s/%s/%s/%s/%s/%s",
			config.Cfg.GameSerEtcdRoot, etcd.DirHotData,
			version, gid, sid, etcd.KeyHotDataLastSignalTime)
		st, err := etcd.Get(k)
		if err == nil && st != "" {
			ist, err := strconv.Atoi(st)
			if err == nil {
				if now-int64(ist) < 10 {
					logs.Error("too quick ! %v", shard)
					continue
				}
			}
		}

		resp := httplib.Get(serCfg.ServerHotDataUrl).
			SetTimeout(5*time.Second, 5*time.Second)
		_, err = resp.String()
		if err != nil {
			logs.Error("signal shard err %v %v", shard, err)
			continue
		}

		etcd.Set(k, fmt.Sprintf("%d", now), 0)
	}
	c.SetData("ok")
	return nil
}

func signalServerHotC(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("signalServerHotC")

	gidsid := params[0]
	version := params[1]
	serCfg := config.Cfg.GetServerCfgFromName(gidsid)
	if serCfg.ServerHotDataUrl == "" {
		return fmt.Errorf("server not found")
	}
	ss := strings.Split(gidsid, ":")
	gid := ss[0]
	sid, err := strconv.Atoi(ss[1])
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	k := fmt.Sprintf("%s/%s/%s/%s/%d/%s", config.Cfg.GameSerEtcdRoot,
		etcd.DirHotData, version, gid, sid, etcd.KeyHotDataLastSignalTime)
	st, err := etcd.Get(k)
	if err == nil && st != "" {
		ist, err := strconv.Atoi(st)
		if err == nil {
			if now-int64(ist) < 10 {
				return fmt.Errorf("too quick !")
			}
		}
	}

	resp := httplib.Get(serCfg.ServerHotDataUrl).SetTimeout(5*time.Second, 5*time.Second)
	res, err := resp.String()
	if err != nil {
		return err
	}

	etcd.Set(k, fmt.Sprintf("%d", now), 0)
	c.SetData(res)
	return nil
}

type hotShardInfo struct {
	Version         string
	VerSeq          string
	GidSid          string
	SidBaseHotBuild string
	SidHotBuild     string
	SidHotSeq       string
}
