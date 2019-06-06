package global_info

import (
	"encoding/json"
	"sync"

	"fmt"
	"strings"

	"strconv"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

const (
	typ_Level = "lvl"
	typ_Trial = "trial"
	typ_Boss  = "boss"
)

type RecInfo struct {
	Aid  string `json:"aid"`
	Name string `json:"name"`
}

type GlobalLevelFinishInfo struct {
	curLevelFinish map[key]RecInfo
	LastSaved      map[string]RecInfo `json:"levelId2Info"`
	dbKey          string

	cmdChan  chan cmd_info
	saveChan chan cmd_info
	waitter  sync.WaitGroup
}

type key struct {
	typ      string
	level_id string
}

func (k key) String() string {
	return fmt.Sprintf("%s:%s", k.typ, k.level_id)
}
func genKey(k string) key {
	ss := strings.Split(k, ":")
	if len(ss) < 2 {
		return key{}
	}
	return key{ss[0], ss[1]}
}

type cmd_info struct {
	k    key
	aid  string
	name string
}

func (info *GlobalLevelFinishInfo) start(sid uint) {
	info.dbKey = tableGlobalLevelFinish(sid)
	info.curLevelFinish = make(map[key]RecInfo, 512)
	info.LastSaved = make(map[string]RecInfo, 512)
	info.cmdChan = make(chan cmd_info, 1024)
	info.saveChan = make(chan cmd_info, 1024)
	info.loadDB()

	// 请求查询
	info.waitter.Add(1)
	go func() {
		defer info.waitter.Done()
		for {
			cmd, ok := <-info.cmdChan
			if !ok {
				close(info.saveChan)
				logs.Warn("modules global_level_finish req_go close")
				return
			}

			// 先查缓存
			if _, ok = info.curLevelFinish[cmd.k]; ok {
				continue
			}
			// 缓存没有先存到缓存
			info.curLevelFinish[cmd.k] = RecInfo{cmd.aid, cmd.name}
			// 存到数据库
			info.saveChan <- cmd
			switch cmd.k.typ {
			case typ_Level:
				onFirstLevelFinish(cmd.k.level_id, cmd.aid, cmd.name)
			case typ_Trial:
				onFirstTrialFinish(cmd.k.level_id, cmd.aid, cmd.name)
			case typ_Boss:
				onFirstBossFinish(cmd.k.level_id, cmd.aid, cmd.name)
			}

		}
	}()
	// 存储
	info.waitter.Add(1)
	go func() {
		defer info.waitter.Done()
		for {
			cmd, ok := <-info.saveChan
			if !ok {
				logs.Warn("modules global_level_finish save_go close")
				return
			}
			info.LastSaved[cmd.k.String()] = RecInfo{cmd.aid, cmd.name}
			// 存db
			info.saveDB()
		}
	}()
}

func (info *GlobalLevelFinishInfo) stop() {
	close(info.cmdChan)
	info.waitter.Wait()
	info.saveDB()
}

func (info *GlobalLevelFinishInfo) loadDB() bool {
	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("modules globalinfo levelfinish loaddb GetDBConn nil")
		return false
	}

	has, err := redis.Bool(_do(db, "EXISTS", info.dbKey))
	if err != nil {
		logs.Error("modules globalinfo levelfinish loaddb redis.Bool err %v", err)
		return false
	}

	if has {
		bb, err := redis.Bytes(_do(db, "GET", info.dbKey))
		if err != nil {
			logs.Error("modules globalinfo levelfinish loaddb redis.Bytes err %v", err)
			return false
		}

		err = json.Unmarshal(bb, info)
		if err != nil {
			logs.Error("modules globalinfo levelfinish loaddb Err %s in %s", err.Error(), bb)
			return false
		}

		for k, v := range info.LastSaved {
			_k := genKey(k)
			if _k.typ == "" {
				logs.Error("GlobalLevelFinishInfo loadDB genKey err %s %v", k, v)
				continue
			}
			info.curLevelFinish[_k] = v
		}
	}
	return true
}

func (info *GlobalLevelFinishInfo) saveDB() bool {
	bb, err := json.Marshal(*info)
	if err != nil {
		logs.Error("modules globalinfo levelfinish savedb marshal err %s", err.Error())
		return false
	}

	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("modules globalinfo levelfinish GetDBConn nil")
		return false
	}
	if _, err := _do(db, "SET", info.dbKey, string(bb)); err != nil {
		logs.Error("modules globalinfo levelfinish savedb error %s", err.Error())
		return false
	}

	return true
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (info *GlobalLevelFinishInfo) levelFinishReq(typ, levelId, acid, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case info.cmdChan <- cmd_info{key{typ, levelId}, acid, name}:
	case <-ctx.Done():
		logs.Error("levelFinishReq  put timeout")
	}
}

func onFirstLevelFinish(levelId, acid, name string) {
	// 跑马灯
	logs.Trace("sysnotice firstlevelfinish %s %s", name, levelId)
	account, _ := db.ParseAccount(acid)
	sysnotice.NewSysRollNotice(account.ServerString(), gamedata.SN_FirstFinishLevel).
		AddParam(sysnotice.ParamType_RollName, name).
		AddParam(sysnotice.ParamType_LevelId, levelId).Send()
}

func onFirstTrialFinish(levelId, acid, name string) {
	lvlId, err := strconv.Atoi(levelId)
	if err != nil {
		logs.Error("onFirstTrialFinish levelId err %s", levelId)
		return
	}
	logs.Trace("sysnotice onFirstTrialFinish %s %s", name, levelId)
	account, _ := db.ParseAccount(acid)
	noticId := gamedata.Trial2SysNotice(uint32(lvlId))
	if noticId <= 0 {
		logs.Error("onFirstTrialFinish Trial2SysNotice err %s", levelId)
		return
	}
	sysnotice.NewSysRollNotice(account.ServerString(), int32(noticId)).
		AddParam(sysnotice.ParamType_RollName, name).
		AddParam(sysnotice.ParamType_Trial_LevelId, levelId).Send()
}

func onFirstBossFinish(levelId, acid, name string) {
	// 跑马灯
	logs.Trace("sysnotice onFirstBossFinish %s %s", name, levelId)
	account, _ := db.ParseAccount(acid)
	sysnotice.NewSysRollNotice(account.ServerString(), gamedata.IDS_KILLHERO).
		AddParam(sysnotice.ParamType_BossId, levelId).
		AddParam(sysnotice.ParamType_RollName, name).Send()
}

func _do(db redispool.RedisPoolConn, commandName string, args ...interface{}) (reply interface{}, err error) {
	return modules.DoWraper("Global_Info_DB", db, commandName, args...)
}
