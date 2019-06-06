package want_gen_best

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func genWantGenBestModule(sid uint) *WantGenBestModule {
	return &WantGenBestModule{
		shardId:   sid,
		save_chan: make(chan WantGenBestInfo, 16),
	}
}

type WantGenBestInfo struct {
	NextRefTime    int64  `json:"nrts"`
	HeroPieceCount uint32 `json:"heropc"`
	Acid           string `json:"acid"`
	Name           string `json:"name"`
}
type WantGenBestModule struct {
	shardId   uint
	Info      WantGenBestInfo
	mutx      sync.RWMutex
	save_chan chan WantGenBestInfo
	waitter   util.WaitGroupWrapper
}

func (m *WantGenBestModule) AfterStart(g *gin.Engine) {
}

func (m *WantGenBestModule) BeforeStop() {
}

func (m *WantGenBestModule) Start() {
	m.dbLoad(m.shardId)
	m.waitter.Wrap(func() {
		for fr := range m.save_chan {
			if err := m.dbSave(&fr); err != nil {
				logs.Error("GlobalCountModule save err: %s", err.Error())
			}
		}
	})
}

func (m *WantGenBestModule) Stop() {
	close(m.save_chan)
	m.waitter.Wait()
	m.dbSave(&m.Info)
}

func (m *WantGenBestModule) CheckAndReplace(now_t int64, n uint32,
	acid, name string) bool {
	m.mutx.RLock()
	if now_t < m.Info.NextRefTime && n <= m.Info.HeroPieceCount {
		m.mutx.RUnlock()
		return false
	}
	m.mutx.RUnlock()
	m.mutx.Lock()
	defer m.mutx.Unlock()
	if now_t < m.Info.NextRefTime && n <= m.Info.HeroPieceCount {
		return false
	}
	if now_t >= m.Info.NextRefTime {
		m.Info.NextRefTime = util.GetNextDailyTime(
			gamedata.GetCommonDayBeginSec(now_t), now_t)
		m.Info.HeroPieceCount = n
		m.Info.Acid = acid
		m.Info.Name = name
	} else {
		if n > m.Info.HeroPieceCount {
			m.Info.HeroPieceCount = n
			m.Info.Acid = acid
			m.Info.Name = name
		}
	}
	m.save_chan <- m.Info
	return true
}

func (m *WantGenBestModule) GetWantGenBest(now_t int64) WantGenBestInfo {
	m.mutx.RLock()
	defer m.mutx.RUnlock()
	if now_t >= m.Info.NextRefTime {
		return WantGenBestInfo{}
	}
	return m.Info
}

func (m *WantGenBestModule) dbLoad(shardId uint) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(),
		tableWantGenBest(shardId), &m.Info, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return err
}

func (m *WantGenBestModule) dbSave(info *WantGenBestInfo) error {
	cb := redis.NewCmdBuffer()

	if err := driver.DumpToHashDBCmcBuffer(cb,
		tableWantGenBest(m.shardId), info); err != nil {
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
	db_counter_key = "WantGenBest_DB"
)
