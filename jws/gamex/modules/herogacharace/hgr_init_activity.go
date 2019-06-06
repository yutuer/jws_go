package herogacharace

import (
	"errors"
	"time"

	"fmt"

	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
)

type HGRActivity struct {
	GroupID uint32
	//当前运行的活动ID
	ActivityId uint32
	//当前活动开始时间 Unix Timestamp
	StartTime int64
	//当先活动的结算时间 Unix Timestamp
	EndTime int64
}

func (a *HGRActivity) GetRedisExpireUnix() int64 {
	return a.EndTime + (a.EndTime-a.StartTime)*2
}

func (a *HGRActivity) GetRedisKey() string {
	return fmt.Sprintf("%s:%d:%d", REDIS_PREFIX, a.GroupID, a.ActivityId)
}

func getEtcdCfg() (map[string]RedisDBSetting, error) {
	var etcdMap map[string]RedisDBSetting
	if jsonvalue, err := etcd.Get(GetEtcdInfoKey()); err != nil {
		return nil, fmt.Errorf("HeroGachaRace.initRedis get key failed. %s", GetEtcdInfoKey())
	} else {

		if err := json.Unmarshal([]byte(jsonvalue), &etcdMap); err != nil {
			return nil, fmt.Errorf("HeroGachaRace.initRedis json.Unmarshal key failed. %s", GetEtcdInfoKey())
		}
	}
	return etcdMap, nil
}

func (hgr *HeroGachaRace) initRedis(hgra HGRActivity) error {
	if game.Cfg.IsRunModeDev() {
		//开发模式默认使用本地Redis配置
		hgr.redisCfg.AddrPort = game.Cfg.Redis
		hgr.redisCfg.Auth = game.Cfg.RedisAuth
		hgr.redisCfg.DB = game.Cfg.RedisDB
	} else {
		etcdMap, err := getEtcdCfg()
		if err != nil {
			return err
		}

		groupString := fmt.Sprintf("%d", hgra.GroupID)
		rcfg, ok := etcdMap[groupString]
		if !ok {
			rcfg, ok = etcdMap["default"]
			if !ok {
				return fmt.Errorf("HeroGachaRace.initRedis etcdmap key not found. %s", groupString)
			} else {
				hgr.redisCfg = rcfg
			}
		} else {
			hgr.redisCfg = rcfg
		}
	}

	//建立数据库链接, 尝试4次, 2 seconds 之内
	//TEST 720.254544ms
	//TEST 1.164560053s
	//TEST 1.875428374s
	err := backoff.Retry(func() error {
		conn, err := redis.Dial("tcp", hgr.redisCfg.AddrPort,
			redis.DialConnectTimeout(5*time.Second),
			redis.DialReadTimeout(5*time.Second),
			redis.DialWriteTimeout(5*time.Second),
			redis.DialPassword(hgr.redisCfg.Auth),
			redis.DialDatabase(hgr.redisCfg.DB),
		)
		if conn != nil {
			hgr.redis = conn
		}
		return err
	}, New2SecBackOff())

	if err != nil {
		logs.Critical("HeroGachaRace.initScores connect db failed.")
		return errors.New("HeroGachaRace.initScores connect db failed")
	} else {
		return nil
	}

	return errors.New("Already has activity")
}

func (hgr *HeroGachaRace) initScores(hgra HGRActivity) error {
	endDur := hgra.EndTime - time.Now().Unix()
	if endDur < 0 {
		return errors.New("Actvitiy end.")
	}

	if hgr.redis == nil {
		err := hgr.initRedis(hgra)
		if err != nil {
			return err
		}
	}

	if hgr.curActivity == nil {
		hgr.curActivity = &hgra
		hgr.pullAllScores(hgra)

		bt := time.Duration((endDur + util.MinSec) * int64(time.Second))
		logs.Debug("HeroGachaRace Balance endtime %v timer %v", endDur, bt)
		hgr.ResetChan = time.After(bt)
		balanceTime := endDur - int64(gamedata.GetHotDatas().HotLimitHeroGachaData.GetHGRConfig().GetPublicityTime())
		if balanceTime < 0 {
			logs.Error("Activity has balanced end")
			balanceTime = 0
		}

		// 提前某些时间发奖
		rt := time.Duration((balanceTime + util.MinSec) *
			int64(time.Second))
		hgr.BalanceChan = time.After(rt)
		hgr.CheatChan = make(chan struct{}, 1)
		hgr.stopChan = make(chan bool, 1)

		hgr.starterWg.Add(1)
		go func() {
			defer hgr.starterWg.Done()
			for {
				select {
				case <-hgr.CheatChan:
					logs.Debug("HeroGachaRace cheat balance %v", hgr.curActivity)
					hgr.balance()
					hgr.debugClear()
					hgr.Reset()
					return
				case <-hgr.BalanceChan:
					if balanceTime > 0 {
						logs.Debug("HeroGachaRace balance %v timer %v", hgr.curActivity, rt)
						hgr.balance()
					}
				case <-hgr.ResetChan:
					//请确保结算完成后在调用Reset
					//清空所有状态到空
					logs.Debug("HeroGachaRace Rest %v timer %v", hgr.curActivity, bt)
					hgr.Reset()
					return
				case <-hgr.stopChan:
					return
				}
			}
		}()
		logs.Debug("HeroGachaRace initScores %v", hgr.curActivity)
	}
	return nil
}

func (hgr *HeroGachaRace) CheatBalance() {
	if hgr.CheatChan != nil {
		hgr.CheatChan <- struct{}{}
	}
}
