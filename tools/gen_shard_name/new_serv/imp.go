package new_serv

import (
	"encoding/csv"
	"os"
	"strconv"

	"fmt"
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/tools/gen_shard_name/shard_init"
)

func Imp() {
	file, err := os.Open("conf/shards.csv")
	if err != nil {
		logs.Error("open csv err %s", err.Error())
		return
	}
	defer file.Close()

	r2 := csv.NewReader(file)
	ss, err := r2.ReadAll()
	if err != nil {
		logs.Error("read csv err %s", err.Error())
		return
	}

	shards := make(map[int]shardInfo, len(ss))
	for i, r := range ss {
		if i != 0 {
			sid, err := strconv.Atoi(r[0])
			if err != nil {
				logs.Error("new server shard id is not number %s", r)
				return
			}
			gid, err := strconv.Atoi(r[1])
			if err != nil {
				logs.Error("new server gid is not number %s", r)
				return
			}
			order, err := strconv.Atoi(r[2])
			if err != nil {
				logs.Error("new server order is not number %s", r)
				return
			}
			sname := shard_init.MkShardName(sid, gid)
			shards[sid] = shardInfo{sid, gid, order, sname}
		}
	}

	logs.Debug("csv shards %v", shards)

	if strings.ToLower(Cfg.Shards[0]) == all_shard {
		for _, info := range shards {
			snKey := fmt.Sprintf("%s/%d/%d/%s", Cfg.EtcdRoot, info.gid, info.sid, etcd.KeySName)
			if err := etcd.Set(snKey, info.sn, 0); err != nil {
				logs.Error("new server set etcd err %s", err.Error())
				return
			}
			logs.Debug("set %s", snKey)
			keyOrder := fmt.Sprintf("%s/%d/%d/%s", Cfg.EtcdRoot, info.gid, info.sid, etcd.KeyOrder)
			if err := etcd.Set(keyOrder, fmt.Sprintf("%d", info.order), 0); err != nil {
				logs.Error("new server set etcd err %s", err.Error())
				return
			}
			logs.Debug("set %s", keyOrder)
		}
	} else {
		for _, ssid := range Cfg.Shards {
			sid, _ := strconv.Atoi(ssid)
			info, ok := shards[sid]
			if !ok {
				logs.Error("new server shard %d not found in csv", sid)
				return
			}
			snKey := fmt.Sprintf("%s/%d/%d/%s", Cfg.EtcdRoot, info.gid, info.sid, etcd.KeySName)
			if err := etcd.Set(snKey, info.sn, 0); err != nil {
				logs.Error("new server set etcd err %s", err.Error())
				return
			}
			logs.Debug("set %s", snKey)
			keyOrder := fmt.Sprintf("%s/%d/%d/%s", Cfg.EtcdRoot, info.gid, info.sid, etcd.KeyOrder)
			if err := etcd.Set(keyOrder, fmt.Sprintf("%d", info.order), 0); err != nil {
				logs.Error("new server set etcd err %s", err.Error())
				return
			}
			logs.Debug("set %s", keyOrder)
		}
	}
}

type shardInfo struct {
	sid   int
	gid   int
	order int
	sn    string
}
