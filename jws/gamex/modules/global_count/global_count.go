package global_count

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	GlobalCount_Typ_Account7DayGood = "acid7day"
	GlobalCount_Typ_Record          = "record" // 玩家各种记录的刷新count

	GlobalCount_DB_Counter_Name = "GlobalCount"
)

const (
	SimplePvpRecord = iota

	RecordCount
)

func genGlobalCountModule(sid uint) *GlobalCountModule {
	return &GlobalCountModule{
		sid:             sid,
		globalCountInfo: make(map[string]*countkvs, 16),
	}
}

type GlobalCountModule struct {
	sid             uint
	globalCountInfo map[string]*countkvs
	w               worker
}

func (m *GlobalCountModule) AfterStart(g *gin.Engine) {
}

func (m *GlobalCountModule) BeforeStop() {
}

func (m *GlobalCountModule) Start() {
	m.w.start(m)
}

func (m *GlobalCountModule) Stop() {
	m.w.stop()
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (f *GlobalCountModule) CommandExec(cmd GlobalCountCmd) *GlobalCountRet {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	res_chan := make(chan GlobalCountRet, 1)
	cmd.resChan = res_chan
	errRet := &GlobalCountRet{}
	chann := f.w.cmd_chan
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("GlobalCountModule CommandExec chann full, cmd put timeout")
		return errRet
	}

	select {
	case res := <-res_chan:
		return &res
	case <-ctx.Done():
		logs.Error("GlobalCountModule CommandExec apply <-res_chan timeout")
		return errRet
	}
}

func (m *GlobalCountModule) dbLoad(gct string) (error, *countkvs) {
	hmc := m.globalCountInfo[gct]
	if hmc != nil {
		return nil, hmc
	}

	dbName := tableGlobalCount(gct, m.sid)

	_db := driver.GetDBConn()
	defer _db.Close()

	kvs := &countkvs{
		CountTyp: gct,
	}
	err := driver.RestoreFromHashDB(_db.RawConn(), dbName, kvs, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err, nil
	}
	if err == driver.RESTORE_ERR_Profile_No_Data {
		kvs.genInitkvs()
	}
	kvs.checkAddNewKvs()
	m.globalCountInfo[gct] = kvs
	return nil, kvs
}

func (m *GlobalCountModule) dbSave(kvs *countkvs) error {
	cb := redis.NewCmdBuffer()
	dbName := tableGlobalCount(kvs.CountTyp, m.sid)

	if err := driver.DumpToHashDBCmcBuffer(cb, dbName, kvs); err != nil {
		return fmt.Errorf("DumpToHashDBCmcBuffer err %v", err)
	}

	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		return fmt.Errorf("cant get redis conn")
	}

	if _, err := modules.DoCmdBufferWrapper(GlobalCount_DB_Counter_Name, db, cb, true); err != nil {
		return fmt.Errorf("DoCmdBuffer error %s", err.Error())
	}
	return nil
}

func (ckvs *countkvs) genInitkvs() {
	switch ckvs.CountTyp {
	case GlobalCount_Typ_Account7DayGood:
		mc := gamedata.GetAccount7DaySerGood()
		ckvs.Counts = make([]countkv, 0, len(mc))
		for k, c := range mc {
			ckvs.Counts = append(ckvs.Counts, countkv{
				Key: GlobalCountKey{
					IId: k,
				},
				Count: c,
			})
		}
	case GlobalCount_Typ_Record:
		ckvs.Counts = make([]countkv, 0, RecordCount)
		ckvs.Counts = append(ckvs.Counts, countkv{
			Key: GlobalCountKey{
				IId: SimplePvpRecord,
			},
			Count: 0,
		})
	default:
		logs.Error("global count genInitkvs CountTyp %s not def", ckvs.CountTyp)
	}
}

func (ckvs *countkvs) checkAddNewKvs() {
	switch ckvs.CountTyp {
	case GlobalCount_Typ_Account7DayGood:
		mc := gamedata.GetAccount7DaySerGood()
		oldkv := make(map[uint32]uint32, len(ckvs.Counts))
		for _, kv := range ckvs.Counts {
			oldkv[kv.Key.IId] = kv.Count
		}

		ckvs.Counts = make([]countkv, 0, len(mc))
		for k, c := range mc {
			oc, ok := oldkv[k]
			if ok {
				ckvs.Counts = append(ckvs.Counts, countkv{
					Key: GlobalCountKey{
						IId: k,
					},
					Count: oc,
				})
			} else {
				ckvs.Counts = append(ckvs.Counts, countkv{
					Key: GlobalCountKey{
						IId: k,
					},
					Count: c,
				})
			}
		}
	case GlobalCount_Typ_Record:

	default:
		logs.Error("global count checkAddNewKvs CountTyp %s not def", ckvs.CountTyp)
	}
}

func (ckvs *countkvs) tranKvs2Ret(ret *GlobalCountRet) {
	switch ckvs.CountTyp {
	case GlobalCount_Typ_Account7DayGood, GlobalCount_Typ_Record:
		ret.Counti2c = make(map[uint32]uint32, len(ckvs.Counts))
		for _, kv := range ckvs.Counts {
			ret.Counti2c[kv.Key.IId] = kv.Count
		}
	}
}

type countkv struct {
	Key   GlobalCountKey `json:"k"`
	Count uint32         `json:"c"`
}
type countkvs struct {
	CountTyp string    `json:"ct"`
	Counts   []countkv `json:"cs"`
}
