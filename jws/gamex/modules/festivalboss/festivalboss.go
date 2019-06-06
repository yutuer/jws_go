package festivalboss

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"

	"sync"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const Max_Festival_Boss_Info = 8

func genFestivalBossModule(sid uint) *FestivalBossModule {
	return &FestivalBossModule{
		shardId:      sid,
		FestivalBoss: make([]FestivalBossInfo, 0, Max_Festival_Boss_Info),
		save_chan:    make(chan saveCmd, 128),
	}
}

type FestivalBossModule struct {
	shardId       uint
	Boss_attacked int64
	FestivalBoss  []FestivalBossInfo
	mutx          sync.RWMutex
	save_chan     chan saveCmd
	waitter       util.WaitGroupWrapper
}

type FestivalBossInfo struct {
	Player_names string `json:"player_names"`
	Player_time  int64  `json:"player_time"`
}

type saveCmd struct {
	shardId       uint
	festivalBoss  []FestivalBossInfo
	boss_attacked int64
}

func (m *FestivalBossModule) Start() {
	m.dbLoad(m.shardId)
	m.waitter.Wrap(func() {
		for cmd := range m.save_chan {
			dbSave(cmd)
		}
	})
}
func (m *FestivalBossModule) AfterStart(g *gin.Engine) {
}

func (m *FestivalBossModule) BeforeStop() {
}

//func (m *FestivalBossModule) Start() {
//	m.dbLoad(m.shardId)
//}

func (m *FestivalBossModule) Stop() {
	close(m.save_chan)
	m.waitter.Wait()
}

func (m *FestivalBossModule) TryAddFestivalBossInfo(shardId uint, name string, ltime int64) bool {
	m.mutx.Lock()
	defer m.mutx.Unlock()
	if len(m.FestivalBoss) >= Max_Festival_Boss_Info {
		m.FestivalBoss = m.FestivalBoss[1:]
	}
	m.FestivalBoss = append(m.FestivalBoss, FestivalBossInfo{Player_names: name, Player_time: ltime})
	m.Boss_attacked += 1
	m._exec(saveCmd{
		shardId:       shardId,
		boss_attacked: m.Boss_attacked,
		festivalBoss:  m.FestivalBoss[:],
	})
	return true
}

func (m *FestivalBossModule) TrySetFestivalBoss2Zero() {
	m.mutx.RLock()
	if m.Boss_attacked == 0 {
		m.mutx.RUnlock()
		return
	}
	m.mutx.RUnlock()
	m.mutx.Lock()
	defer m.mutx.Unlock()
	m.FestivalBoss = m.FestivalBoss[:0]
	m.Boss_attacked = 0
	m._exec(saveCmd{
		shardId:       m.shardId,
		boss_attacked: 0,
		festivalBoss:  []FestivalBossInfo{},
	})
}

func (m *FestivalBossModule) dbLoad(shardId uint) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(),
		tableDestFestivalBoss(shardId), m, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return err
}

func dbSave(m saveCmd) error {
	cb := redis.NewCmdBuffer()

	if err := driver.DumpToHashDBCmcBuffer(cb,
		tableDestFestivalBoss(m.shardId), &FestivalBossModule{
			shardId:       m.shardId,
			Boss_attacked: m.boss_attacked,
			FestivalBoss:  m.festivalBoss,
		}); err != nil {
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

func (m *FestivalBossModule) GetFestivalBossInfo() ([]FestivalBossInfo, int64) {
	m.mutx.RLock()
	info := m.FestivalBoss[:]
	attack := m.Boss_attacked
	m.mutx.RUnlock()
	return info, attack
}

func (r *FestivalBossModule) _exec(cmd saveCmd) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	chann := r.save_chan
	select {
	case chann <- cmd:
	case <-ctx.Done():
		logs.Error("Festival CommandExec chann full, cmd put timeout")
	}
}

const (
	db_counter_key = "Festival_Boss_DB"
)
