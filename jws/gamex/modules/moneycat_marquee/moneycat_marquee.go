package moneycat_marquee

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

const Max_MoneyCat_Marquee = 10

func genDestGenFirstModule(sid uint) *MoneyCatModule {
	return &MoneyCatModule{
		shardId: sid,
		Marquee: make([]MarqueeInfo, 0, Max_MoneyCat_Marquee),
	}
}

type MoneyCatModule struct {
	shardId    uint
	ActivityId uint32
	Marquee    []MarqueeInfo `json:"dest_gens"`
	mutx       sync.RWMutex
}

type MarqueeInfo struct {
	Player_names string `json:"player_names"`
	Player_GetHc int64  `json:"player_get_hc"`
}

func (m *MoneyCatModule) AfterStart(g *gin.Engine) {
}

func (m *MoneyCatModule) BeforeStop() {
}

func (m *MoneyCatModule) Start() {
	m.dbLoad(m.shardId)
}

func (m *MoneyCatModule) Stop() {
	m.dbSave(m.shardId)
}

func (m *MoneyCatModule) TryAddMoneyCatInfo(activityType int, channelId string, nt int64, name string, gethc int64) bool {
	m.mutx.Lock()
	defer m.mutx.Unlock()
	m._TrySetMoneyCatInfo2Zero(activityType, channelId, nt)

	if len(m.Marquee) == Max_MoneyCat_Marquee {
		m.Marquee = m.Marquee[1:]
	}
	m.Marquee = append(m.Marquee, MarqueeInfo{Player_names: name, Player_GetHc: gethc})
	m.dbSave(m.shardId)
	return true
}

func (m *MoneyCatModule) TrySetMoneyCatInfo2Zero(activityType int, channelId string, nt int64) {
	m.mutx.Lock()
	defer m.mutx.Unlock()

	m._TrySetMoneyCatInfo2Zero(activityType, channelId, nt)

}

func (m *MoneyCatModule) dbLoad(shardId uint) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(),
		tableDestMoneyCat(shardId), m, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return err
}

func (m *MoneyCatModule) dbSave(shardId uint) error {
	cb := redis.NewCmdBuffer()

	if err := driver.DumpToHashDBCmcBuffer(cb,
		tableDestMoneyCat(shardId), m); err != nil {
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

func (m *MoneyCatModule) GetMoneyCatInfo() []MarqueeInfo {
	m.mutx.RLock()
	info := m.Marquee[:]
	m.mutx.RUnlock()
	return info
}

func (m *MoneyCatModule) _TrySetMoneyCatInfo2Zero(activityType int, channelId string, nt int64) {
	actId := gamedata.GetHotDatas().Activity.GetActActivity(nt, activityType, channelId)
	if m.ActivityId != actId {
		m.ActivityId = actId
		if len(m.Marquee) > 0 {
			m.Marquee = m.Marquee[:0]
			m.dbSave(m.shardId)
		}
	}

}

const (
	db_counter_key = "DestGenFirst_DB"
)
