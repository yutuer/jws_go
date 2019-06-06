package game

import (
	"sync"
	"time"

	"github.com/siddontang/go/timingwheel"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	dbMux   sync.RWMutex
	dbReady map[string]bool
	dbTimer *timingwheel.TimingWheel
)

func init() {
	dbReady = make(map[string]bool)
	dbTimer = timingwheel.NewTimingWheel(100*time.Millisecond, 10*60)
}

func markDBStatusInUse(accountId string) {
	dbMux.Lock()
	defer dbMux.Unlock()
	logs.Debug("DBStatus InUse:%s", accountId)
	dbReady[accountId] = true
}

func markDBStatusRelease(accountId string) {
	dbMux.Lock()
	defer dbMux.Unlock()
	logs.Debug("DBStatus Release:%s", accountId)
	delete(dbReady, accountId)
}

func isDBStatusReady(accountId string) bool {
	dbMux.RLock()
	defer dbMux.RUnlock()
	_, ok := dbReady[accountId]
	if ok {
		logs.Debug("DBStatus Check failed:%s", accountId)
		return false
	}
	logs.Debug("DBStatus Check ok:%s", accountId)
	return true
}

func IsAllAccountOffLine() bool {
	dbMux.RLock()
	defer dbMux.RUnlock()
	if len(dbReady) > 0 {
		logs.Warn("account online %v", dbReady)
	}
	return len(dbReady) <= 0
}
