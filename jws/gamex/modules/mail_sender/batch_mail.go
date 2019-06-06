package mail_sender

import (
	"fmt"

	"encoding/json"

	"time"

	"math/rand"

	"strings"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

/*
	主要用来批量发邮件，并支持失败重发和停服重启继续发的功能
	redis作为cache，用个list保存要发的邮件
	定时把redis的list里的邮件，拿出来，批量发到dynamodb，失败在放回redis的list
*/

type mailForBatch struct {
	UserId     string
	Mail       timail.MailReward
	IsActivity bool
	Typ        int64
}

// isAct: 为true，将优先发送，为特殊活动准备，慎用
func AddMailBatchSend(shardId uint, userId string, mail timail.MailReward, isAct bool, typ int64) {
	if userId == "" {
		logs.Error("AddMailBatchSend userId %s", userId)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	select {
	case GetModule(shardId).command_batch_chan <- mailForBatch{
		UserId:     "profile:" + userId,
		Mail:       mail,
		IsActivity: isAct,
		Typ:        typ,
	}:
	case <-ctx.Done():
		logs.Error("[AddMailBatchSend] chann full, mail put timeout")
	}
}

func mail2Cache(shardId uint, mail mailForBatch, isAct bool) error {
	v, err := json.Marshal(mail)
	if err != nil {
		return fmt.Errorf("mail2Cache json.Marshal err: %s", err.Error())
	}

	_db := driver.GetDBConn()
	defer _db.Close()

	if isAct {
		_, err = _do(_db, "LPUSH", cacheTableName(shardId), string(v))
		if err != nil {
			return fmt.Errorf("mail2Cache LPUSH err %s", err.Error())
		}
	} else {
		_, err = _do(_db, "RPUSH", cacheTableName(shardId), string(v))
		if err != nil {
			return fmt.Errorf("mail2Cache RPUSH err %s", err.Error())
		}
	}
	return nil
}

func cache2DB(shardId uint) error {
	tableName := cacheTableName(shardId)

	_db := driver.GetDBConn()
	defer _db.Close()

	mailCount, err := redis.Int(_do(_db, "LLEN", tableName))
	if err != nil && err != redis.ErrNil {
		return fmt.Errorf("cache2DB LLEN err: %s", err.Error())
	}
	ts := _batchTimes(mailCount)
	for _, at := range ts {
		time.Sleep(time.Millisecond * time.Duration(rand.Int63n(5000)))
		ms, err := redis.Strings(_do(_db, "LRANGE", tableName, 0, at-1))
		if err != nil && err != redis.ErrNil {
			return fmt.Errorf("cache2DB LRANGE err: %s", err.Error())
		}
		if ms == nil || len(ms) <= 0 {
			break
		}

		logs.Debug("MailBatchCache2DB oper mails %d", len(ms))
		mails := make(map[timail.MailKey]string, len(ms))

		// 解析所有mail
		userIds := make([]string, 0, len(ms))
		mailRewards := make([]timail.MailReward, 0, len(ms))
		for _, v := range ms {
			mail := &mailForBatch{}
			if err := json.Unmarshal([]byte(v), mail); err != nil {
				return fmt.Errorf("cache2DB json.Unmarshal err: %s", err.Error())
			}
			k := timail.MailKey{mail.UserId, mail.Mail.Idx}
			if _, ok := mails[k]; ok {
				logs.Error("cache2DB mail duplicate %v", mail)
				continue
			}
			mails[k] = v
			if checkUserid(mail.UserId) {
				userIds = append(userIds, mail.UserId)
				mailRewards = append(mailRewards, mail.Mail)
			}
		}
		// 存dynamodb
		var fails []timail.MailKey
		if len(userIds) > 0 {
			err, fails = mail_db.BatchWriteMails(userIds, mailRewards)
			// logic log
			logicLogBatchMail(userIds, mailRewards, fails)
			if err != nil {
				return fmt.Errorf("cache2DB BatchWriteMails err: %s", err.Error())
			}
		}
		ret, err := _do(_db, "LTRIM", tableName, at, -1)
		if err != nil || ret != "OK" {
			return fmt.Errorf("cache2DB LTRIM err %v %s", err, ret)
		}
		if len(fails) > 0 {
			logs.Warn("MailBatchCache2DB UnprocessedItems len %d", len(fails))
			cb := redis.NewCmdBuffer()
			for _, k := range fails {
				if v, ok := mails[k]; ok {
					err = cb.Send("RPUSH", tableName, string(v))
					if err != nil {
						continue
					}
				}
			}
			if _, err := modules.DoCmdBufferWrapper(Mail_DB_Counter_Name, _db, cb, true); err != nil {
				return fmt.Errorf("cache2DB DoCmdBuffer err: %s", err.Error())
			}
		}
	}
	return nil
}
func logicLogBatchMail(ids []string, mailRewards []timail.MailReward, key []timail.MailKey) {
	for i, item := range mailRewards {
		mailKey := timail.MailKey{
			Idx: item.Idx,
			Uid: ids[i],
		}
		success := true
		for _, k := range key {
			if k == mailKey {
				success = false
				break
			}
		}
		if success {
			logiclog.LogSendMail(ids[i], item.IdsID, item.Reason, item.ItemId, item.Count)
		}
	}
}

func cacheTableName(shardId uint) string {
	return fmt.Sprintf("%s:%d", mail_batch_table, shardId)
}

func _do(db redispool.RedisPoolConn, commandName string, args ...interface{}) (reply interface{}, err error) {
	return modules.DoWraper(Mail_DB_Counter_Name, db, commandName, args...)
}

func _batchTimes(mailCount int) []int {
	batchCount := 4 // 每批次数量
	t := mailCount / batchCount
	res := make([]int, 0, t+1)
	for i := 0; i < t; i++ {
		res = append(res, batchCount)
	}
	l := mailCount % batchCount
	if l > 0 {
		res = append(res, l)
	}
	return res
}

func checkUserid(user_id string) bool {
	if strings.HasPrefix(user_id, "all:") {
		return len(strings.Split(user_id, ":")) >= 1
	} else {
		ss := strings.Split(user_id, "profile:")
		if len(ss) <= 1 {
			return false
		}
		uid := ss[1]
		_, err := db.ParseAccount(uid)
		if err != nil {
			logs.Error("checkUserid get invalide userid:%s", user_id)
			return false
		}
		return true
	}
}
