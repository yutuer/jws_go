package server_mng

import (
	"encoding/json"

	"strconv"
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"

	"errors"
	"fmt"

	"time"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/supervisord"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/act_valid"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("getShardsFromEtcd", getShardsFromEtcd)
	gm_command.AddGmCommandHandle("setShard2Etcd", setShard2Etcd)
	gm_command.AddGmCommandHandle("setShardStateEtcd", setShardStateEtcd)
	gm_command.AddGmCommandHandle("restartShard", restartShard)
}

func restartShard(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("restartShard %v", params)

	//gid := params[0]
	sid := params[1]
	machine := params[2]

	shard := fmt.Sprintf("gamex-shard%s", sid)
	url := fmt.Sprintf("http://%s:%s@%s:9001/RPC2",
		"4uG2ct6foow", "AMj9kHhR4kv4LkdXLeNxkXupDD", machine)
	s := supervisord.New(url, nil)
	s.StopProcess(shard, true)
	b, err := s.StartProcess(shard, true)
	if err != nil || !b {
		return fmt.Errorf("StartProcess err %s", err.Error())
	}
	info, err := s.GetProcessInfo("chat")
	if err != nil {
		return fmt.Errorf("GetProcessInfo err %s", err.Error())
	}
	if info.Statename != "RUNNING" {
		return fmt.Errorf("process state %s", info.Statename)
	}
	c.SetData("ok")
	return nil
}

func getShardsFromEtcd(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("getShardsFromEtcd")

	type res struct {
		Shards []etcd.ShardInfo
	}
	var r res

	shard2Info := make(map[int]etcd.ShardInfo, 10)
	logs.Debug(fmt.Sprintf("key is :%s/", config.Cfg.GameMachineEtcdRoot))
	// /a4k/skydns/net/gamex
	machines, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/", config.Cfg.GameMachineEtcdRoot))
	if err != nil {
		return err
	}
	type MachineHost struct {
		Host string `json:"host"`
	}
	for _, machine := range machines {
		pg := strings.Split(machine, "/")
		ishard, err := strconv.Atoi(pg[len(pg)-1])
		if err != nil {
			continue
		}
		kv, err := etcd.GetAllSubKeyValue(machine)
		if err != nil || len(kv) <= 0 {
			continue
		}
		machHost := MachineHost{}
		for _, v := range kv {
			if err := json.Unmarshal([]byte(v), &machHost); err != nil {
				logs.Error("getShardsFromEtcd machine err %v", err)
			}
			break
		}
		shard2Info[ishard] = etcd.ShardInfo{
			Sid:     uint(ishard),
			Machine: machHost.Host,
		}
	}

	// /a4k/shard
	var _shard2Info map[int]etcd.ShardInfo
	_shard2Info = make(map[int]etcd.ShardInfo, 10)
	for k, v := range shard2Info {
		_shard2Info[k] = v
	}
	//gids, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/", config.Cfg.GameSerEtcdRoot))
	//logs.Debug("###gids %v",gids)
	//if err != nil {
	//	return err
	//}
	gids := config.Cfg.GidFilter
	for _, root_gid := range gids {
		pg := strings.Split(root_gid, "/")
		igid, err := strconv.Atoi(pg[len(pg)-1])
		if err != nil {
			continue
		}
		sids, err := etcd.GetAllSubKeys(root_gid)
		if err != nil {
			return err
		}

		for _, root_sid := range sids {
			ps := strings.Split(root_sid, "/")
			if len(ps) > 3 {
				logs.Warn("sid %s -- %s %s", root_sid, ps[2], ps[3])

				isid, err := strconv.Atoi(ps[3])
				if err != nil {
					continue
				}
				logs.Debug("sid:%v", root_sid)
				kv, err := etcd.GetAllSubKeyValue(root_sid)
				if err != nil {
					continue
				}
				logs.Debug("kv:%v", kv)
				ip, _ := kv[etcd.KeyIp]
				sn, _ := kv[etcd.KeySName]
				dn, _ := kv[etcd.KeyDName]
				state, _ := kv[etcd.KeyState]
				showState, _ := kv[etcd.KeyShowState]
				serStartTime, _ := kv[etcd.KeyServerStartTime]
				serLaunchTime, _ := kv[etcd.KeyServerLaunchTime]
				//将启动时间转换为当地时区时间
				if serLaunchTime != "" {
					lanuchTime, err := util.UnixStringToTimeByLoacl(serLaunchTime)
					if err != nil {
						logs.Error("getShardsFromEtcd timechange error : %v", err)
					} else {
						serLaunchTime = lanuchTime.Format("2006/01/02 15:04")
					}
				}
				logs.Trace("serlanchTime:%v", serLaunchTime)
				serRealStartTime, _ := kv[etcd.KeyServerUsedStartTime]
				logs.Debug("before serRealStartTime:%v", serRealStartTime)
				//将开服时间转换为当地时区时间
				if serRealStartTime != "" {
					startTime, err := util.UnixStringToTimeByLoacl(serRealStartTime)
					if err != nil {
						logs.Error("getShardsFromEtcd timechange error : %v", err)
					} else {
						serRealStartTime = startTime.Format("2006/01/02 15:04")
					}
				}
				logs.Debug("last serRealStartTime:%v", serRealStartTime)

				serVersion, _ := etcd.Get(fmt.Sprintf("%s/gm/version", root_sid))
				logs.Debug("serVersion:%v", serVersion)

				multiLang, _ := kv[etcd.MultiLang]
				lang, _ := kv[etcd.Language]

				shard := etcd.ShardInfo{
					Gid:           uint(igid),
					Sid:           uint(isid),
					Ip:            ip,
					SName:         sn,
					DName:         dn,
					State:         state,
					ShowState:     showState,
					StartTime:     serStartTime,
					LaunchTime:    serLaunchTime,
					RealStartTime: serRealStartTime,
					Version:       serVersion,
					MultiLang:     multiLang,
					Language:      lang,
				}
				if sh, ok := _shard2Info[int(shard.Sid)]; ok {
					shard.Machine = sh.Machine
				}
				_shard2Info[int(shard.Sid)] = shard
			} else {
				logs.Error("sid format Err %s", root_sid)
			}
		}
	}

	r.Shards = make([]etcd.ShardInfo, 0, len(_shard2Info))
	for _, v := range _shard2Info {
		r.Shards = append(r.Shards, v)
	}
	b, _ := json.Marshal(&r)
	c.SetData(string(b))
	return nil
}

//只改变服务器状态
func setShardStateEtcd(c *gm_command.Context, server, accountid string, params []string) error {
	gid := params[0]
	sid := params[1]
	state := params[2]
	show_state, err := strconv.Atoi(state)
	if err != nil || show_state >= etcd.ShowState_Count {
		logs.Error("setShardStateEtcd err:%v", err)
		return errors.New(fmt.Sprintf("shard_show_state param %s err", state))
	}
	key := fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyShowState)
	if res, err := etcd.Get(key); err != nil || res != state {
		if err := etcd.Set(key, state, 0); err != nil {
			logs.Error("setShardStateEtcd err:%v", err)
			return err
		}
	}
	c.SetData("ok")
	return nil
}

func setShard2Etcd(c *gm_command.Context, server, accountid string, params []string) error {
	if len(params) < 8 {
		return errors.New("setShard2Etcd params not enough")
	}
	gid := params[0]
	sid := params[1]
	dn := params[2]
	state := params[3]
	//	ip := params[4]
	sn := params[5]
	ss := params[6]
	serStartTime := params[7]
	multiLang := params[8]
	language := params[9]

	//	key := fmt.Sprintf("%s/%s/%s/ip", config.Cfg.GameSerEtcdRoot, gid, sid)
	//	if res, err := etcd.Get(key); err != nil || res != ip {
	//		if err != nil {
	//			return err
	//		}
	//		return errors.New(fmt.Sprintf("setShard2Etcd  %s  ip chg %s %s", key, ip, res))
	//	}

	key := fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeySName)
	if res, err := etcd.Get(key); err != nil || res != sn {
		if err != nil {
			return err
		}
		return errors.New(fmt.Sprintf("setShard2Etcd  %s  ip chg %s %s", key, sn, res))
	}

	if dn == "" {
		return errors.New("setShard2Etcd params dn is empty")
	}
	key = fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyDName)
	if res, err := etcd.Get(key); err != nil || res != dn {
		if err := etcd.Set(key, dn, 0); err != nil {
			return err
		}
	}

	if state == etcd.StateOnline {
		key = fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyState)
		if res, err := etcd.Get(key); err != nil || res != state {
			if err := etcd.Set(key, state, 0); err != nil {
				return err
			}
		}
	}

	show_state, err := strconv.Atoi(ss)
	if err != nil || show_state >= etcd.ShowState_Count {
		return errors.New(fmt.Sprintf("shard_show_state param %s err", ss))
	}
	key = fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyShowState)
	if res, err := etcd.Get(key); err != nil || res != ss {
		if err := etcd.Set(key, ss, 0); err != nil {
			return err
		}
	}

	if serStartTime != "" {
		if !strings.Contains(serStartTime, " 00:00") {
			serStartTime += " 00:00"
		}
		_, err = time.Parse("2006/1/2 15:04", serStartTime)
		if err != nil {
			return errors.New(fmt.Sprintf("server start time param %s err", serStartTime))
		}
		key = fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyServerStartTime)
		if err := etcd.Set(key, serStartTime, 0); err != nil {
			return err
		}
	}

	key = fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyActValid)
	act_v, err := etcd.Get(key)
	if err != nil || act_v == "" {
		act_s := make([]string, act_valid.ActCount)
		for i := 0; i < act_valid.ActCount; i++ {
			act_s[i] = "1"
		}
		//临时更改活动etcd配置
		act_s[0] = "0"

		ss := strings.Join(act_s, ",")
		etcd.Set(key, ss, 0)
	}

	key = fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.MultiLang)
	if res, err := etcd.Get(key); err != nil || res != multiLang {
		if err := etcd.Set(key, multiLang, 0); err != nil {
			return err
		}
	}

	key = fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.Language)
	if res, err := etcd.Get(key); err != nil || res != language {
		if err := etcd.Set(key, language, 0); err != nil {
			return err
		}
	}

	// 更改版本号
	key = fmt.Sprintf("%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.ServerListVersion)
	if err := etcd.Set(key, fmt.Sprintf("%d", time.Now().Unix()), 0); err != nil {
		return err
	}

	c.SetData("ok")
	return nil
}
