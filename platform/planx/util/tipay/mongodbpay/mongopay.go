package mongodbpay

/*
MongodDB使用_id默认作为主键索引。
需要为QueryByUid, 接口准备userid的索引。
*/

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"vcs.taiyouxi.net/platform/planx/util/tipay/pay"
)

const (
	PayCollection  = "Pay"
	PayUidIndex    = "uid"
	PayOrderMainId = "order_no"
)

type MongodbPay struct {
	*mgo.Session
	DBName string
}

func (mdb *MongodbPay) getMongo() *MongodbPay {
	m := &MongodbPay{
		Session: mdb.Copy(),
		DBName:  mdb.DBName,
	}
	return m
}

func (mdb *MongodbPay) getDB() *mgo.Collection {
	return mdb.DB(mdb.DBName).C(PayCollection)
}

func NewMongoDBPay(pc pay.PayDBConfig) (pay.PayDB, error) {
	session, err := mgo.DialWithTimeout(pc.MongoDBUrl, 20*time.Second)
	if err != nil {
		return nil, err
	}

	var mp MongodbPay
	mp.Session = session
	mp.DBName = pc.DBName
	session.SetMode(mgo.Monotonic, true)
	err = session.DB(mp.DBName).
		C(PayCollection).
		EnsureIndexKey(PayUidIndex)
	if err != nil {
		return nil, err
	}

	return &mp, nil
}

func (mp *MongodbPay) SetByHashM(hash interface{}, values map[string]interface{}) error {
	p := mp.getMongo()
	defer p.Close()
	//values["_id"] = values["PayOrderMainId"]
	//return p.getDB().Insert(values)
	//根据DynamoDB UpdateItem, PutItem语意修改
	_, err := p.getDB().Upsert(bson.D{{"_id", hash}}, bson.M{"$set": values})
	return err
}

func (mp *MongodbPay) SetByHashM_IfNoExist(hash interface{}, values map[string]interface{}) error {
	p := mp.getMongo()
	defer p.Close()
	values["_id"] = hash
	err := p.getDB().Insert(values)
	return err
}

func (mp *MongodbPay) GetByHashM(hash interface{}) (map[string]interface{}, error) {
	p := mp.getMongo()
	defer p.Close()
	v := make(map[string]interface{}, 10)
	err := p.getDB().FindId(hash).One(v)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return v, nil
}

func (mp *MongodbPay) UpdateByHash(hash interface{}, values map[string]interface{}) error {
	p := mp.getMongo()
	defer p.Close()
	_, err := p.getDB().Upsert(bson.D{{"_id", hash}}, bson.M{"$set": values})
	return err
}

func (mp *MongodbPay) IsExist(hash interface{}) (bool, error) {
	p := mp.getMongo()
	defer p.Close()
	n, err := p.getDB().FindId(hash).Count()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (mp *MongodbPay) QueryByUid(ucid string) ([]pay.GateWay_Pay_Info, error) {
	p := mp.getMongo()
	defer p.Close()
	var r []pay.GateWay_Pay_Info
	err := p.getDB().Find(bson.D{{PayUidIndex, ucid}}).All(&r)
	if err != nil {
		return nil, err
	}

	return r[:], nil
}

//TODO by YZH mongodb pay, close?
