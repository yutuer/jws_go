package redeemCodeModule

import (
	"time"

	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RedeemCodeMongoDB struct {
	*mgo.Session
	DBName string
}

func (rcm *RedeemCodeMongoDB) Open() error {
	if err := rcm.DB(rcm.DBName).C(rcm.DBName).EnsureIndex(mgo.Index{
		Key:        []string{"code"},
		Unique:     true,
		Sparse:     true,
		Background: true,
		DropDups:   true,
	}); err != nil {
		return err
	}
	return nil
}

func (rcm *RedeemCodeMongoDB) getDB() *mgo.Collection {
	return rcm.DB(rcm.DBName).C(rcm.DBName)
}

func (rcm *RedeemCodeMongoDB) GetRedeemCodeData(code string) (*RedeemCodeExchange, error) {
	rce := RedeemCodeExchange{}
	err := rcm.getDB().Find(bson.D{{"code", code}}).One(&rce)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(rce.Value), &rce); err != nil {
		return nil, err
	}
	logs.Debug("get redeem code exchange: %v", rce)
	return &rce, nil
}

func (rcm *RedeemCodeMongoDB) SetRedeemCodeUsed(acID string, code string) error {
	info, err := rcm.getDB().Upsert(bson.D{{"code", code}},
		bson.M{"$set": RedeemCodeExchange{
			Code:   code,
			DoneBy: acID,
			State:  RedeemCodeStateDone,
		}})
	if err != nil {
		return err
	}
	logs.Debug("change info: %v", info)
	return nil
}

func NewRedeemCodeMongoDB(url string, dbname string) *RedeemCodeMongoDB {
	session, err := mgo.DialWithTimeout(url, 20*time.Second)
	if err != nil {
		return nil
	}

	var mdb RedeemCodeMongoDB
	mdb.Session = session
	mdb.DBName = dbname
	session.SetMode(mgo.Monotonic, true)
	return &mdb
}
