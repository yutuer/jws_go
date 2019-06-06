package act_valid

import (
	"fmt"
	"strconv"
	"strings"

	"errors"

	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	gmConfig "vcs.taiyouxi.net/platform/x/gm_tools/config"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("getActValidInfo", getActValidInfo)
	gm_command.AddGmCommandHandle("setActValidInfo", setActValidInfo)
}

const (
	ActCount = 50
)

func getActValidInfo(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("getShardsFromEtcd")

	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	logs.Info("[getActValidInfo]  %s --> %v", server, cfg)

	serv := strings.Split(cfg.ServerName, ":")
	if len(serv) < 2 {
		return fmt.Errorf("server err %s", serv)
	}
	gid := serv[0]
	sid := serv[1]
	key := fmt.Sprintf("%s/%s/%s/%s", gmConfig.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyActValid)
	res, err := etcd.Get(key)
	if err != nil {
		return err
	}

	acts := strings.Split(res, ",")

	if res == "" {
		var act_valued1 []string
		for i := 0; i < ActCount; i++ {
			act_valued1 = append(act_valued1, "1")
		}
		act_valueds := strings.Join(act_valued1, ",")
		etcd.Set(key, act_valueds, 0)

		c.SetData(act_valueds)
		return nil
	} else if len(acts) < ActCount {
		var act_valued []string
		for i := 0; i < ActCount; i++ {
			act_valued = append(act_valued, "1")
		}
		for i, act := range acts {
			_, err := strconv.Atoi(act)
			if err != nil {
				return fmt.Errorf("act_valid1 data err %s", res)
			}
			act_valued[i] = acts[i]
		}
		act_valueds := strings.Join(act_valued, ",")

		etcd.Set(key, act_valueds, 0)

		c.SetData(act_valueds)
		return nil
	}

	for _, act := range acts {
		_, err := strconv.Atoi(act)
		if err != nil {
			return fmt.Errorf("act_valid2 data err %s", res)
		}
	}
	c.SetData(res)
	return nil
}

func setActValidInfo(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("setActValidInfo")

	if len(params) < 1 {
		return errors.New("setActValidInfo params not enough")
	}
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	logs.Info("[getActValidInfo]  %s --> %v", server, cfg)

	serv := strings.Split(cfg.ServerName, ":")
	if len(serv) < 2 {
		return fmt.Errorf("server err %s", serv)
	}
	gid := serv[0]
	sid := serv[1]

	s := params[0]
	acts := strings.Split(s, ",")
	for _, act := range acts {
		_, err := strconv.Atoi(act)
		if err != nil {
			return fmt.Errorf("setActValidInfo act_valid data err %s", s)
		}
	}

	key := fmt.Sprintf("%s/%s/%s/%s", gmConfig.Cfg.GameSerEtcdRoot, gid, sid, etcd.KeyActValid)
	if err := etcd.Set(key, s, 0); err != nil {
		return err
	}

	c.SetData("ok")
	return nil
}
