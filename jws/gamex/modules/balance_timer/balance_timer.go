package balance

import (
	"errors"
	"sync"

	"time"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func newBalance(sid uint) *BalanceModule {
	return &BalanceModule{
		sid:                        sid,
		db_name:                    tableRankBalance(sid),
		rankNeedBalanceNotifyChans: make([]RankBalanceNotifyData, 0, 16),
		ticker: time.NewTicker(1 * time.Second),
	}
}

type BalanceTimerHandler interface {
	SetLast(last int64)
	GetLast() int64
}

type CommonBalanceTimeLast struct {
	L int64 `json:"l"`
}

func (c *CommonBalanceTimeLast) SetLast(last int64) {
	c.L = last
}

func (c *CommonBalanceTimeLast) GetLast() int64 {
	return c.L
}

type RankBalanceNotifyData struct {
	channel     chan<- bool
	timeBalance util.TimeToBalance
	b           CommonBalanceTimeLast
	name        string
}

type BalanceModule struct {
	sid                        uint
	db_name                    string
	rankNeedBalanceNotifyChans []RankBalanceNotifyData
	rankRWMutex                sync.RWMutex
	ticker                     *time.Ticker
}

func (m *BalanceModule) Start() {

}

func (m *BalanceModule) AfterStart(g *gin.Engine) {
	if err := m.fromDB(); err != nil {
		logs.Error("RankModule rankBalance.fromDB panic, err %v", err)
		logs.Error("RankModule award may lost !! ")
		panic(err)
	}
	go m.StartTimer()
}

func (m *BalanceModule) BeforeStop() {

}

func (m *BalanceModule) Stop() {

}

func (rb *BalanceModule) fromDB() error {
	for i, _ := range rb.rankNeedBalanceNotifyChans {
		r := &rb.rankNeedBalanceNotifyChans[i]
		lt, err := getBalanceLastTime(rb.db_name, r.name)
		if err != nil {
			return err
		}
		r.b.SetLast(lt)
		logs.Trace("BalanceModule load %s", r.name)
	}
	return nil
}

func (r *BalanceModule) DebugSetBalanceTimeOffset(t int64) {
	r.rankRWMutex.Lock()
	defer r.rankRWMutex.Unlock()
	for i := 0; i < len(r.rankNeedBalanceNotifyChans); i++ {
		rb := &r.rankNeedBalanceNotifyChans[i]
		rb.b.SetLast(t)
	}
}

func (r *BalanceModule) RegBalanceNotifyChan(name string, c chan<- bool, t util.TimeToBalance) {
	r.rankRWMutex.Lock()
	defer r.rankRWMutex.Unlock()
	r.rankNeedBalanceNotifyChans = append(r.rankNeedBalanceNotifyChans,
		RankBalanceNotifyData{
			name:        name,
			channel:     c,
			timeBalance: t,
		})
}

func (r *BalanceModule) BalanceNotifyAll() {
	r.rankRWMutex.RLock()
	defer r.rankRWMutex.RUnlock()
	for _, chan_2_balance := range r.rankNeedBalanceNotifyChans {
		logs.Warn("BalanceNotifyAll %s", chan_2_balance.name)
		chan_2_balance.channel <- true
	}
}

// 判断一个时间点是否是在今天的结算周期之内,
func (r *BalanceModule) isInNowDayRankBalanceTime(h BalanceTimerHandler,
	timeBalance util.TimeToBalance) (needBal, needSaveBalTime bool) {
	nowT := time.Now().Unix()
	last := h.GetLast()
	if last == 0 {
		h.SetLast(nowT)
		needSaveBalTime = true
	}
	last = h.GetLast()

	if !util.IsSameUnixByStartTime(last, nowT, timeBalance) {
		h.SetLast(nowT)
		return true, true
	}

	return false, needSaveBalTime
}

func (r *BalanceModule) StartTimer() {
	for {
		<-r.ticker.C
		func() {
			r.rankRWMutex.RLock()
			defer r.rankRWMutex.RUnlock()
			for i := 0; i < len(r.rankNeedBalanceNotifyChans); i++ {
				rb := &r.rankNeedBalanceNotifyChans[i]
				needBal, needSaveBal := r.isInNowDayRankBalanceTime(&rb.b, rb.timeBalance)
				if needSaveBal {
					// 将lastTime存db
					saveBalanceLastTime(r.db_name, rb.name, rb.b.GetLast())
				}
				if needBal {
					// 发结算信号
					logs.Warn("Balance %s For %v", rb.name, rb.timeBalance)
					rb.channel <- true
				}
			}
		}()
	}
}

func getBalanceLastTime(db_name, key string) (int64, error) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s", db_name)
		return 0, errors.New("GetDBConnNil")
	}
	lastTime, err := redis.Int64(conn.Do("HGET", db_name, key))
	if err != nil && err != redis.ErrNil {
		logs.Error("getBalanceLastTime err %v", err)
		return 0, err
	}
	return lastTime, nil
}

func saveBalanceLastTime(db_name, key string, t int64) error {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by %s", db_name)
		return errors.New("GetDBConnNil")
	}
	if _, err := conn.Do("HSET", db_name, key, t); err != nil {
		logs.Error("saveBalanceLastTime err %v", err)
		return err
	}
	return nil
}
