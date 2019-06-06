package timongodb

/// MongoDB 的实现中在批量写入的过程中没有像DynamoDB一样使用任何重连机制。
/// DynamoDB的重连机制主要是为了应对一个服务的Capacity的临时不可用, 而MongoDB本身不是用Capacity来衡量的。

import (
	"fmt"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	. "vcs.taiyouxi.net/platform/planx/util/timail"
)

const (
	MAILMONGO_COLLECTION = "mails"
)

type DBByMongoDB struct {
	*mgo.Session
	DBName string
}

func (mdb *DBByMongoDB) getDB(user_id string) *mgo.Collection {
	var sid string
	if strings.HasPrefix(user_id, "all:") {
		sid = strings.Split(user_id, ":")[1]
	} else {
		uid := strings.Split(user_id, "profile:")[1]
		acId, err := db.ParseAccount(uid)
		if err != nil {
			logs.Error("DBByMongoDB.getDB get invalide userid:%s", user_id)
			return nil
		}
		sid = fmt.Sprintf("%d", game.Cfg.GetShardIdByMerge(acId.ShardId))
	}
	collectionName := fmt.Sprintf("%s_%s", MAILMONGO_COLLECTION, sid)
	if err := mdb.DB(mdb.DBName).C(collectionName).EnsureIndex(mgo.Index{
		Key:        []string{"uid", "t"},
		Unique:     true,
		Sparse:     true,
		Background: true,
		DropDups:   true,
	}); err != nil {
		logs.Critical("DBByMongoDB EnsureIndex err : %s", err.Error())
	}
	return mdb.DB(mdb.DBName).C(collectionName)
}

func (mdb *DBByMongoDB) GetMongo() *DBByMongoDB {
	return &DBByMongoDB{
		Session: mdb.Copy(),
		DBName:  mdb.DBName,
	}
}

//LoadAllMail 参考 func (d *DynamoDB) QueryMail, t,mail,begint,endt,isget,isread,acc2send,ctb2send,cte2send
func (mdb *DBByMongoDB) LoadAllMail(user_id string) ([]MailReward, error) {
	//XXX by YZH LoadAllMail&LoadAllMailByGM ProjectionExpression中的一些映射不同
	//应该是节省DynamoDB的Capacity, Mongodb不考虑
	mon := mdb.GetMongo()
	defer mon.Close()
	var res []MailRes
	err := mon.getDB(user_id).Find(bson.M{"uid": user_id, "t": bson.M{
		"$gte": 0,
		"$lte": MkMailId(0, 0),
	}}).Sort("uid", "-t").Batch(MAX_Mail_ONE_GET).All(&res)
	re := make([]MailReward, 0, len(res)+1)
	if err != nil {
		return re[:], err
	}

	for _, mail_from_db := range res {
		new_mail := MailReward{}
		new_mail.FormDB(&mail_from_db)
		re = append(re, new_mail)
	}

	return re, nil
}

// LoadAllMailByGM 只获取了如下字段: "t, mail, begint, endt, isget"
// 请参考 QueryAllMail
func (mdb *DBByMongoDB) LoadAllMailByGM(user_id string) ([]MailReward, error) {
	return mdb.LoadAllMail(user_id)
}

func (mdb *DBByMongoDB) SendMail(user_id string, mr MailReward) error {
	var m MailRes
	if err := mr.ToDB(&m); err != nil {
		return err
	}
	mon := mdb.GetMongo()
	defer mon.Close()
	m.Userid = user_id
	//bson.D{{"uid", m.Userid}, {"t", m.Id}}, bson.M{"$set": m}
	err := mon.getDB(user_id).Insert(m)
	return err
}

func (mdb *DBByMongoDB) SyncMail(user_id string, mails_add []MailReward, mails_del []int64) error {
	if len(mails_add) == 0 && len(mails_del) == 0 {
		return nil
	}
	mon := mdb.GetMongo()
	defer mon.Close()
	db := mon.getDB(user_id)
	var err1, err2 error
	if len(mails_del) > 0 {
		_, err1 = db.RemoveAll(bson.M{
			"uid": user_id,
			"t":   bson.M{"$in": mails_del},
		})
	}

	if len(mails_add) > 0 {

		for _, ma := range mails_add {
			var m MailRes
			m.Userid = user_id
			if err := ma.ToDB(&m); err != nil {
				logs.Error("DBByMongoDB.SyncMail error:%s, %v", err.Error(), ma)
				continue
			}
			_, err := db.Upsert(bson.D{{"uid", user_id}, {"t", m.Id}}, m)
			if err != nil {
				logs.Error("DBByMongoDB.SyncMail2 error:%s, %v", err.Error(), ma)
				continue
			}
		}

	}
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

// BatchWriteMails return error, and  failed MailKeys
func (mdb *DBByMongoDB) BatchWriteMails(user_id []string, mails_add []MailReward) (error, []MailKey) {
	sends, err := HelperBatchWriteMails(user_id, mails_add)
	if err != nil {
		return err, nil
	}

	mmail := HelperBatchDupDrop(sends)

	mon := mdb.GetMongo()
	defer mon.Close()
	db := mon.getDB(user_id[0])

	var lastErr error
	var failed []MailKey
	for k, v := range mmail {
		if err := db.Insert(v); err != nil {
			if mgo.IsDup(err) {
				continue
			} else {
				logs.Error("DBByMongoDB.BatchWriteMails failed with error: %s, %v", err.Error(), v)
				lastErr = err
				failed = append(failed, k)
			}
		}
	}

	return lastErr, failed
}

func (mdb *DBByMongoDB) MailExist(user_id string, idx int64) (bool, error) {
	mon := mdb.GetMongo()
	defer mon.Close()
	db := mon.getDB(user_id)
	n, err := db.Find(bson.D{{"uid", user_id}, {"t", idx}}).Count()
	if err == mgo.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (mdb *DBByMongoDB) Close() error {
	mdb.Session.Close()
	return nil
}
