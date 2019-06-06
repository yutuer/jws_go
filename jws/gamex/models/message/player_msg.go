package message

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//
// 向玩家发送离线信息 现在用于发送玩家pvp战斗记录
//

type PlayerMsg struct {
	Typ    int      `json:"t"`
	Params []string `json:"p"`
}

// msgs:XXX:1:10:UUID
func LoadPlayerMsgs(accountID, msgType string, maxCount int) ([]PlayerMsg, error) {
	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("LoadPlayerMsgs GetDBConn nil")
		return []PlayerMsg{}, fmt.Errorf("GetDBConn nil")
	}

	redisKey := fmt.Sprintf("msgs:%s:%s", msgType, accountID)

	res, err := redis.Strings(db.Do("LRANGE", redisKey, 0, maxCount-1))
	if err != nil {
		return []PlayerMsg{}, err
	}

	if len(res) == 0 {
		return []PlayerMsg{}, nil
	}

	msgs := make([]PlayerMsg, 0, len(res))
	for i := 0; i < len(res); i++ {
		newMsg := PlayerMsg{}
		err := json.Unmarshal([]byte(res[i]), &newMsg)
		if err != nil {
			logs.SentryLogicCritical(accountID,
				"LoadPlayerMsgs Unmarshal Err By %s", err.Error())
			continue
		}
		msgs = append(msgs, newMsg)
	}

	return msgs[:], nil
}

func SendPlayerMsgs(accountID, msgType string, maxCount int, msg PlayerMsg) {
	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("SendPlayerMsgs GetDBConn nil")
		return
	}

	redisKey := fmt.Sprintf("msgs:%s:%s", msgType, accountID)

	msgStr, err := json.Marshal(msg)

	if err != nil {
		logs.SentryLogicCritical(accountID, "SendPlayerMsgs Marshal Err By %s", err.Error())
		return
	}

	count, err := redis.Int(db.Do("LPUSH", redisKey, msgStr))
	if err != nil {
		logs.SentryLogicCritical(accountID, "SendPlayerMsgs Err By %s", err.Error())
		return
	}

	if count > maxCount {
		db.Do("LTRIM", redisKey, 0, maxCount-1)
	}
}

func RemPlayerMsg(accountID, msgType string, msg PlayerMsg) bool {
	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("RemPlayerMsg GetDBConn nil")
		return false
	}

	redisKey := fmt.Sprintf("msgs:%s:%s", msgType, accountID)
	msgStr, err := json.Marshal(msg)
	if err != nil {
		logs.SentryLogicCritical(accountID, "RemPlayerMsg Marshal Err By %s", err.Error())
		return false
	}

	remCount, err := redis.Int(db.Do("LREM", redisKey, 1, msgStr))
	if err != nil {
		logs.SentryLogicCritical(accountID, "SendPlayerMsgs Err By %s", err.Error())
		return false
	}
	return remCount > 0
}

func RemGankMsgById(accountID, msgType string, eId int, eAcid string, isSameMsg func(int, string, string) bool) {
	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		logs.Error("RemPlayerMsg GetDBConn nil")
		return
	}

	redisKey := fmt.Sprintf("msgs:%s:%s", msgType, accountID)
	records, err := redis.Strings(db.Do("LRANGE", redisKey, 0, -1))
	if err != nil {
		logs.SentryLogicCritical(accountID, "RemPlayerMsgById Err %s", err.Error())
		return
	}
	toRem := ""
	for _, rec := range records {
		msg := &PlayerMsg{}
		err = json.Unmarshal([]byte(rec), msg)
		if err != nil {
			logs.SentryLogicCritical(accountID, "remplayermsgbyid msg json err %s", err.Error())
			return
		}
		if len(msg.Params) == 0 {
			logs.Warn(accountID, "rem player msg by id err")
			return
		}
		if isSameMsg(eId, eAcid, msg.Params[0]) {
			toRem = rec
			break
		}
	}
	if toRem == "" {
		logs.Warn(accountID, "not found toRem msg, %d, %v", eId, records)
		return
	}

	_, err = redis.Int(db.Do("LREM", redisKey, 1, toRem))
	if err != nil {
		logs.SentryLogicCritical(accountID, "SendPlayerMsgs Err By %s", err.Error())
	}
	logs.Debug("remove player msg success, %s, %d, %s", accountID, eId, eAcid)
}
