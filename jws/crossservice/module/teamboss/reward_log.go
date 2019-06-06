package teamboss

import (
	"encoding/json"
	"fmt"
	"sync"

	"vcs.taiyouxi.net/jws/crossservice/util/csdb"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const db_name = ""

type RewardLog struct {
	Rewards map[string]*Rewards `json:"rewards_log"`
	Group   uint32
	mutex   sync.RWMutex
}

type Rewards struct {
	GlobalRoomID string          `json:"gl_room_id"`
	hasReward    bool            `json:"has_rw"`
	HasRedBox    bool            `json:"has_rb"`
	Status       map[string]bool `json:"status"`
	Level        uint32          `json:"level"`
	CostID       string          `json:"cost_id"`
}

func (rl *RewardLog) receiveReward(globalRoomID string, acID string) *Rewards {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	if log := rl.Rewards[globalRoomID]; log != nil {
		logs.Debug("[TeamBoss] receiveReward: %v", *log)
		if s, ok := log.Status[acID]; ok {
			if !s {
				log.Status[acID] = true
				canRemove := true
				for _, item := range log.Status {
					if !item {
						canRemove = false
						break
					}
				}
				if canRemove {
					delete(rl.Rewards, globalRoomID)
				}
				return log
			}
		}
	}
	return nil
}

func (rl *RewardLog) setReward(globalRoomID string, log *Rewards) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	rl.Rewards[globalRoomID] = log
}

func (rl *RewardLog) dbName() string {
	return fmt.Sprintf("tb_reward_log:%d", rl.Group)
}

func (rl *RewardLog) save() error {
	conn := csdb.GetDBConn(rl.Group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("Get db conn failed")
	}
	value, err := json.Marshal(rl.Rewards)
	if err != nil {
		return fmt.Errorf("Json marshal err by: %v", err)
	}
	_, err = conn.Do(
		"SET", rl.dbName(), value)
	if nil != err {
		return fmt.Errorf("Redis do set round status SET failed, %v", err)
	}
	return nil
}

func (rl *RewardLog) load() error {
	conn := csdb.GetDBConn(rl.Group)
	defer conn.Close()
	if conn.IsNil() {
		return fmt.Errorf("Get db conn failed")
	}
	v, err := redis.Bytes(conn.Do(
		"GET", rl.dbName()))
	if nil != err && err != redis.ErrNil {
		return fmt.Errorf("Redis do set round status GET failed, %v", err)
	}
	if err == redis.ErrNil {
		return nil
	}
	rewards := make(map[string]*Rewards, 0)
	err = json.Unmarshal(v, &rewards)
	if err != nil {
		return fmt.Errorf("Json unmarshal err by: %v", err)
	}
	rl.Rewards = rewards
	return nil
}
