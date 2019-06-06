package dest_gen_first

import (
	"sync"

	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

func genDestGenFirstModule(sid uint) *DestGenFirstModule {
	return &DestGenFirstModule{
		shardId:    sid,
		FirstDests: make([]DestGenFirst, helper.MaxDestingGeneralCount),
	}
}

type DestGenFirstModule struct {
	shardId    uint
	FirstDests []DestGenFirst `json:"dest_gens"`
	mutx       sync.RWMutex
}

type DestGenFirst struct {
	FirstPlayerName      string `json:"fst_nm"`
	FirstPlayerAvatarId  int    `json:"fst_avid"`
	FirstPlayerTimeStamp int64  `json:"fst_st"`
}

func (m *DestGenFirstModule) AfterStart(g *gin.Engine) {
}

func (m *DestGenFirstModule) BeforeStop() {
}

func (m *DestGenFirstModule) Start() {
	m.dbLoad(m.shardId)
}

func (m *DestGenFirstModule) Stop() {
	m.dbSave(m.shardId)
}

func (m *DestGenFirstModule) TryAddFirstDestGen(destId int,
	name string, avatarId int) bool {
	m.mutx.RLock()
	if m.FirstDests[destId].FirstPlayerName != "" {
		m.mutx.RUnlock()
		return false
	}
	m.mutx.RUnlock()
	m.mutx.Lock()
	defer m.mutx.Unlock()
	if m.FirstDests[destId].FirstPlayerName != "" {
		return false
	}
	m.FirstDests[destId].FirstPlayerName = name
	m.FirstDests[destId].FirstPlayerAvatarId = avatarId
	m.FirstDests[destId].FirstPlayerTimeStamp = time.Now().Unix()
	// TODO 因为神兽的数量不是太多，所以这里没有把存db操作放到单独的goroutine里；可以优化
	m.dbSave(m.shardId)
	return true
}

func (m *DestGenFirstModule) GetFirstDestGen(destId int) DestGenFirst {
	m.mutx.RLock()
	info := m.FirstDests[destId]
	m.mutx.RUnlock()
	return info
}

func (m *DestGenFirstModule) dbLoad(shardId uint) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(),
		TableDestGenFirst(shardId), m, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return err
}

func (m *DestGenFirstModule) dbSave(shardId uint) error {
	cb := redis.NewCmdBuffer()

	if err := driver.DumpToHashDBCmcBuffer(cb,
		TableDestGenFirst(shardId), m); err != nil {
		return fmt.Errorf("DumpToHashDBCmcBuffer err %v", err)
	}

	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		return fmt.Errorf("cant get redis conn")
	}

	if _, err := modules.DoCmdBufferWrapper(
		db_counter_key, db, cb, true); err != nil {
		return fmt.Errorf("DoCmdBuffer error %s", err.Error())
	}
	return nil
}

const (
	db_counter_key = "DestGenFirst_DB"
)
