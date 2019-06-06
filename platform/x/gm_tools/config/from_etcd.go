package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type CfgByServer struct {
	cfgs map[string]*ServerCfg
}

var CfyByServers CfgByServer

// sid /a3k/1/default.
// sid /a3k/1/10
func (c *CfgByServer) LoadFromEtcd(etcdRoot string, gidfilter []string) error {
	gids, err := etcd.GetAllSubKeys(etcdRoot + "/")
	if err != nil {
		return err
	}
	c.cfgs = make(map[string]*ServerCfg)
	logs.Warn("GetAllSubKeys gids %v", gids)
	for _, gid := range gids {
		for _, ggid := range gidfilter {
			if gid != ggid {
				continue
			}

			pg := strings.Split(gid, "/")
			_, err := strconv.Atoi(pg[len(pg)-1])
			if err != nil {
				logs.Warn("gid no num %s", gid)
				continue
			}

			sids, err := etcd.GetAllSubKeys(gid)
			if err != nil {
				return err
			}

			for _, sid := range sids {
				ps := strings.Split(sid, "/")
				if len(ps) > 3 {
					logs.Warn("sid %s -- %s %s", sid, ps[2], ps[3])
					id := fmt.Sprintf("%s:%s", ps[2], ps[3])
					gid := fmt.Sprintf("%s", ps[2])
					_, err := strconv.Atoi(ps[3])
					if err != nil {
						logs.Warn("sid no num %s", ps[3])
						continue
					}

					cf, err := c.LoadOneServerFromEtcd(id, gid, sid, "", ServerCfg{})
					if err != nil {
						logs.Error("server cfg err %s", err.Error())
					} else {
						logs.Warn("server cfg %s --> %v", ps, cf)
						c.cfgs[id] = cf
					}
				} else {
					logs.Error("sid format Err %s", sid)
				}
			}
		}
	}
	return nil
}

func (c *CfgByServer) GetAllFilterServerName() []string {
	logs.Debug("config: %v", *c)
	ret := make([]string, 0, 10)
	for key, _ := range c.cfgs {
		ret = append(ret, key)
	}
	return ret
}

func (c *CfgByServer) AddCfg(common *CommonConfig) {
	for name, cfg := range c.cfgs {
		common.AddCfg(name, cfg)
	}

}

//LoadOneServerFromEtcd
//ServerCfg 对应game/config.go
//func (c *Config) SyncInfo2Etcd() bool {
func (c *CfgByServer) LoadOneServerFromEtcd(name, gid, key, private_ip string, s ServerCfg) (*ServerCfg, error) {
	gckeyValues, err := etcd.GetAllSubKeyValue(key + "/gm")
	if err != nil {
		return nil, err
	}
	res := s

	ok := true

	// 当s为空时读的是default,这里不要求有private_id
	if private_ip == "" {
		private_ip, ok = gckeyValues["private_ip"]
		if !ok {
			logs.Error("No private_ip by %s %s", name, key)
			return nil, errors.New("NoPrivateIP")
		}
	}
	logs.Warn("private_ip %s", private_ip)

	res.RedisName = name
	redis, ok := gckeyValues["redis"]
	if ok {
		res.RedisAddress = redis
	}
	redis_db, ok := gckeyValues["redis_db"]
	if ok {
		res.RedisDB, err = strconv.Atoi(redis_db)
		if err != nil {
			return nil, err
		}
	}

	redis_rank, ok := gckeyValues["redis_rank"]
	if ok {
		res.RedisRankAddress = redis_rank
	}
	redis_rank_db, ok := gckeyValues["redis_rank_db"]
	if ok {
		res.RedisRankDB, err = strconv.Atoi(redis_rank_db)
		if err != nil {
			return nil, err
		}
	}

	redis_db_auth, ok := gckeyValues["redis_db_auth"]
	if ok {
		res.RedisAuth = redis_db_auth
	}

	mail, ok := gckeyValues["MailDBName"]
	if ok {
		res.MailDBName = mail
	}

	maildriver, ok := gckeyValues["MailDBDriver"]
	if ok {
		res.MailDBDriver = maildriver
	}

	mailmongourl, ok := gckeyValues["MailMongoUrl"]
	if ok {
		res.MailMongoUrl = mailmongourl
	}

	broadCast_url, ok := gckeyValues["broadCast_url"]
	if ok {
		broadCast_url = strings.Replace(broadCast_url, "127.0.0.1", private_ip, -1)
		res.SysRollNoticeUrl = broadCast_url
		logs.Warn("private_ip %s %s", private_ip, broadCast_url)
	}

	auth_url, ok := gckeyValues["auth_ip_addr"]
	if ok {
		logs.Warn("private_ip %s", private_ip)
		//auth_url = strings.Replace(auth_url, "127.0.0.1", private_ip, -1)
		if auth_url[len(auth_url)-1] == '/' {
			res.AuthApi = fmt.Sprintf("http://%sauth/v1/api/user/ban/", auth_url)
			res.AuthGagApi = fmt.Sprintf("http://%sauth/v1/api/user/gag/", auth_url)
		} else {
			res.AuthApi = fmt.Sprintf("http://%s/auth/v1/api/user/ban/", auth_url)
			res.AuthGagApi = fmt.Sprintf("http://%s/auth/v1/api/user/gag/", auth_url)
		}
	}
	res.ServerName = name

	hotdataurl, ok := gckeyValues["hotdataurl"]
	if ok {
		res.ServerHotDataUrl = hotdataurl
	}
	rankreloadurl, ok := gckeyValues["rank_reload_url"]
	if ok {
		res.RankReloadUrl = rankreloadurl
	}
	mergedshard, ok := gckeyValues["mergedshard"]
	if ok {
		res.MergedShard = fmt.Sprintf("%s:%s", gid, mergedshard)
	}
	return &res, nil
}

//
//func (c *CommonConfig) LoadFromEtcd(cfg_addr string) error {
//	cfgs, err := etcd.GetAllSubKeyValue(cfg_addr)
//
//	if err != nil {
//		return err
//	}
//
//	run_mode, ok := cfgs["run_mode"]
//	if ok {
//		c.Runmode = run_mode
//	}
//
//	url, ok := cfgs["url"]
//	if ok {
//		c.Url = url
//	}
//
//	mail_db, ok := cfgs["mail_db"]
//	if ok {
//		c.MailDBName = mail_db
//	}
//
//	aws_region, ok := cfgs["aws_region"]
//	if ok {
//		c.AWS_Region = aws_region
//	}
//
//	aws_initial_interval, ok := cfgs["aws_initial_interval"]
//	if ok {
//		i, err := strconv.ParseInt(aws_initial_interval, 10, 64)
//		if err != nil {
//			return err
//		}
//		c.AWS_InitialInterval = i
//	}
//
//	aws_multiplier, ok := cfgs["aws_multiplier"]
//	if ok {
//		f, err := strconv.ParseFloat(aws_multiplier, 64)
//		if err != nil {
//			return err
//		}
//		c.AWS_Multiplier = f
//	}
//
//	aws_max_elapsed_time, ok := cfgs["aws_max_elapsed_time"]
//	if ok {
//		i, err := strconv.ParseInt(aws_max_elapsed_time, 10, 64)
//		if err != nil {
//			return err
//		}
//		c.AWS_MaxElapsedTime = i
//	}
//
//	db_store_file, ok := cfgs["db_store_file"]
//	if ok {
//		c.DBStoreFile = db_store_file
//	}
//
//	return nil
//}
