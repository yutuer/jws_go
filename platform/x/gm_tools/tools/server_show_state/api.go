package server_show_state

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"sort"

	"time"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("getAllGids", getAllGids)
	gm_command.AddGmCommandHandle("getShardShowStateByGid", getShardShowStateByGid)
	gm_command.AddGmCommandHandle("setShardsShowState", setShardsShowState)
	gm_command.AddGmCommandHandle("setShardsShowTeamAB", setShardsShowTeamAB)
}

func getAllGids(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("getAllGids")
	gids, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/", config.Cfg.GameSerEtcdRoot))
	if err != nil {
		return err
	}
	res_gids := make([]int, 0, 10)
	for _, gid := range gids {
		pg := strings.Split(gid, "/")
		igid, err := strconv.Atoi(pg[len(pg)-1])
		if err != nil {
			continue
		}
		res_gids = append(res_gids, igid)
	}
	b, _ := json.Marshal(&res_gids)
	c.SetData(string(b))
	return nil
}

func getShardShowStateByGid(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("getShardShowStateByGid")
	if len(params) <= 0 {
		return errors.New("getShardShowStateByGid param not enough")
	}
	gid := params[0]
	sids, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/%s/", config.Cfg.GameSerEtcdRoot, gid))
	if err != nil {
		return err
	}
	isids := make(sort.IntSlice, 0, len(sids))
	sid2ss := make(map[int]string, len(sids))
	teamAB := make(map[int]string, len(sids))
	for _, sid := range sids {
		ps := strings.Split(sid, "/")
		if len(ps) > 3 {
			logs.Warn("sid %s -- %s %s", sid, ps[2], ps[3])

			isid, err := strconv.Atoi(ps[3])
			if err != nil {
				continue
			}
			kv, err := etcd.GetAllSubKeyValue(sid)
			if err != nil {
				continue
			}
			isids = append(isids, isid)
			showState, _ := kv[etcd.KeyShowState]
			tAB, _ := kv[etcd.KeyTeamAB]
			sid2ss[isid] = showState
			if tAB == "" {
				tAB = "A"
			}
			teamAB[isid] = tAB
			//			res = append(res, fmt.Sprintf("%d %s", isid, showState))
		}
	}
	sort.Sort(isids)
	res := make([]string, 0, 10)
	for _, isid := range isids {
		showState := sid2ss[isid]
		tAB := teamAB[isid]
		res = append(res, fmt.Sprintf("%d %s (%s)", isid, showState, tAB))
	}
	b, _ := json.Marshal(&res)
	c.SetData(string(b))
	return nil
}

func setShardsShowState(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("setShardsShowState")
	logs.Debug("params: %v", params)

	if len(params) <= 2 {
		return errors.New("setShardsShowState param not enough")
	}
	gid := params[0]
	show_state, err := strconv.Atoi(params[1])
	if err != nil {
		return errors.New(fmt.Sprintf("setShardsShowState showState err %v", err))
	}
	if show_state >= etcd.ShowState_Count {
		return errors.New(fmt.Sprintf("setShardsShowState showState %v >= etcd.ShowState_Count ", show_state))
	}
	if show_state == etcd.ShowState_Experience ||
		show_state == etcd.ShowState_Maintenance {
		return errors.New(fmt.Sprintf("setShardsShowState showState illegal"))
	}
	// 获得所有sid
	sids, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/%s/", config.Cfg.GameSerEtcdRoot, gid))
	if err != nil {
		return err
	}
	m_sids := make(map[string]string, 10)
	for _, sid := range sids {
		ps := strings.Split(sid, "/")
		if len(ps) > 3 {
			logs.Warn("sid %s -- %s %s", sid, ps[2], ps[3])

			_, err := strconv.Atoi(ps[3])
			if err != nil {
				continue
			}
			m_sids[ps[3]] = ""
		}
	}
	// 设置showstate
	for i, param := range params {
		if i == 0 || i == 1 {
			continue
		}
		ss := strings.Split(param, " ")
		sid := ss[0]
		if _, ok := m_sids[sid]; !ok {
			logs.Error("setShardsShowState sid %s not found", sid)
			continue
		}
		key := fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyShowState)
		old_show_state_str, err := etcd.Get(key)
		if err == nil {
			old_show_state, err := strconv.Atoi(old_show_state_str)
			if err == nil {
				if old_show_state == etcd.ShowState_Experience ||
					old_show_state == etcd.ShowState_Maintenance {
					continue
				}
			}
		}
		if err := etcd.Set(key, fmt.Sprintf("%d", show_state), 0); err != nil {
			return err
		}
	}
	// 重新获取并刷新
	if err := getShardShowStateByGid(c, server, accountid, params); err != nil {
		return err
	}
	//
	key := fmt.Sprintf("%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, etcd.ServerListVersion)
	if err := etcd.Set(key, fmt.Sprintf("%d", time.Now().Unix()), 0); err != nil {
		return err
	}
	return nil
}

func setShardsShowTeamAB(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("setShardsShowTeamAB")

	if len(params) <= 2 {
		return errors.New("setShardsShowTeamAB param not enough")
	}
	gid := params[0]
	show_teamAB := params[1]
	if show_teamAB != "A" && show_teamAB != "B" && show_teamAB != "AB" {
		return errors.New(fmt.Sprintf("setShardsShowTeamAB showState illegal"))
	}
	// 获得所有sid
	sids, err := etcd.GetAllSubKeys(fmt.Sprintf("%s/%s/", config.Cfg.GameSerEtcdRoot, gid))
	if err != nil {
		return err
	}
	m_sids := make(map[string]string, 10)
	for _, sid := range sids {
		ps := strings.Split(sid, "/")
		if len(ps) > 3 {
			logs.Warn("sid %s -- %s %s", sid, ps[2], ps[3])

			_, err := strconv.Atoi(ps[3])
			if err != nil {
				continue
			}
			m_sids[ps[3]] = ""
		}
	}
	// 设置showteamAB
	for i, param := range params {
		if i == 0 || i == 1 {
			continue
		}
		ss := strings.Split(param, " ")
		sid := ss[0]
		if _, ok := m_sids[sid]; !ok {
			logs.Error("setShardsShowTeamAB sid %s not found", sid)
			continue
		}
		key := fmt.Sprintf("%s/%s/%s/%s", config.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyTeamAB)
		if err := etcd.Set(key, show_teamAB, 0); err != nil {
			return err
		}
	}
	// 重新获取并刷新
	if err := getShardShowStateByGid(c, server, accountid, params); err != nil {
		return err
	}
	return nil
}
