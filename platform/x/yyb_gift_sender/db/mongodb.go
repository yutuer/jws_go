package db

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/yyb_gift_sender/config"
)

type DBByMongoDB struct {
	*mgo.Session
	DBName string
}

func (mdb *DBByMongoDB) checkGSCondition(sdkUid string, acId string, giftId, cond uint32) (int, bool) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(YYBGiftTable)
	info := c.Find(YYBGift{SdkUid: sdkUid, GiftID: int(giftId)})
	count, err := info.Count()
	if err != nil {
		logs.Error("DBByMongoDB Find err %v", err)
		return config.RetCode_SendFailed, false
	}
	if count > 0 {
		logs.Error("Already sent the gift")
		return config.RetCode_HadSend, false
	}
	account := loadProfile(acId)
	if account == nil {
		logs.Error("LoadProfile err %v", err)
		return config.RetCode_SendFailed, false
	}
	if account.Profile.GetData().CorpCurrGS < int(cond) {
		logs.Error("Hero GS err")
		return config.RetCode_NoCondition, false
	}

	err = c.Insert(bson.M{"id": sdkUid, "gift_id": int(giftId), "times": 1, "get_time": time.Now().Unix()})
	if err != nil {
		logs.Error("DBByMongoDB Insert err %v", err)
		return config.RetCode_SendFailed, false
	}
	return 0, true
}

func (mdb *DBByMongoDB) checkLevelCondition(sdkUid string, acId string, giftId, cond uint32) (int, bool) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(YYBGiftTable)
	info := c.Find(YYBGift{SdkUid: sdkUid, GiftID: int(giftId)})
	count, err := info.Count()
	if err != nil {
		logs.Error("DBByMongoDB Find err %v", err)
		return config.RetCode_SendFailed, false
	}
	if count > 0 {
		logs.Error("Already sent the gift")
		return config.RetCode_HadSend, false
	}
	account := loadProfile(acId)
	if account == nil {
		logs.Error("LoadProfile err %v", err)
		return config.RetCode_SendFailed, false
	}

	if account.Profile.GetCorp().Level < cond {
		logs.Error("Hero Level err")
		return config.RetCode_NoCondition, false
	}
	err = c.Insert(bson.M{"id": sdkUid, "gift_id": int(giftId), "times": 1, "get_time": time.Now().Unix()})
	if err != nil {
		logs.Error("DBByMongoDB Insert err %v", err)
		return config.RetCode_SendFailed, false
	}
	return 0, true
}

func (mdb *DBByMongoDB) checkLoginCondition(sdkUid string, acId string, giftId uint32) (int, bool) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(YYBGiftTable)
	info := c.Find(YYBGift{SdkUid: sdkUid, GiftID: int(giftId)})
	_, err := info.Count()
	if err != nil {
		logs.Error("DBByMongoDB Find err %v", err)
		return config.RetCode_SendFailed, false
	}

	var giftInfo YYBGift

	account := loadProfile(acId)
	if account == nil {
		logs.Error("LoadProfile err %v", err)
		return config.RetCode_SendFailed, false
	}
	lastLoginTime := account.Profile.LoginTime

	err = info.One(&giftInfo)
	if err != nil && err != mgo.ErrNotFound {
		logs.Error("DBByMongoDB One %v", err)
		return config.RetCode_SendFailed, false
	}
	now_t := time.Now()
	if err == mgo.ErrNotFound && now_t.Day() == time.Unix(lastLoginTime, 0).Day() {
		err = c.Insert(bson.M{"id": sdkUid, "gift_id": int(giftId), "times": 1, "get_time": time.Now().Unix()})
		if err != nil {
			logs.Error("DBByMongoDB Insert err %v", err)
			return config.RetCode_SendFailed, false
		}
	} else {
		if now_t.Day() != time.Unix(giftInfo.GetTime, 0).Day() && now_t.Day() == time.Unix(lastLoginTime, 0).Day() {
			_, err := c.Upsert(YYBGift{SdkUid: sdkUid, GiftID: int(giftId)},
				bson.M{"$set": YYBGift{Times: giftInfo.Times + 1, GetTime: now_t.Unix()}})
			if err != nil {
				logs.Error("DBByMongoDB Upsert err %v", err)
				return config.RetCode_SendFailed, false
			}
		} else if now_t.Day() == time.Unix(giftInfo.GetTime, 0).Day() && now_t.Day() == time.Unix(lastLoginTime, 0).Day() {
			logs.Error("Get already err")
			return config.RetCode_HadSend, false
		} else {
			logs.Error("DBByMongoDB Upsert err")
			return config.RetCode_NoCondition, false
		}
	}

	return 0, true
}

func (mdb *DBByMongoDB) GetDB() *mgo.Database {
	return mdb.DB(mdb.DBName)
}

func (mdb *DBByMongoDB) GetMongo() *DBByMongoDB {
	return &DBByMongoDB{
		Session: mdb.Copy(),
		DBName:  mdb.DBName,
	}
}

func (mdb *DBByMongoDB) Init(cfg config.Config) error {
	logs.Info("DBByMongoDB Init %v", cfg)
	session, err := mgo.DialWithTimeout(cfg.GiftMongoDBUrl, 20*time.Second)
	if err != nil {
		return err
	}

	mdb.Session = session
	mdb.DBName = cfg.GiftDBName

	session.SetMode(mgo.Monotonic, true)

	if err := mdb.DB(mdb.DBName).C(YYBGiftTable).EnsureIndex(
		mgo.Index{
			Key:        []string{"id", "gift_id"},
			Unique:     true,
			Sparse:     true,
			Background: true,
			DropDups:   true,
		}); err != nil {
		logs.Critical("DBByMongoDB EnsureIndex err : %s", err.Error())
		return err
	}

	return nil
}

func (mdb *DBByMongoDB) checkSignCondition(sdkUid string, giftId uint32) (int, bool) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(YYBGiftTable)
	info := c.Find(YYBGift{SdkUid: sdkUid, GiftID: int(giftId)})
	_, err := info.Count()
	if err != nil {
		logs.Error("DBByMongoDB Find err %v", err)
		return config.RetCode_SendFailed, false
	}

	var giftInfo YYBGift
	err = info.One(&giftInfo)
	if err != nil && err != mgo.ErrNotFound {
		logs.Error("DBByMongoDB One %v", err)
		return config.RetCode_SendFailed, false
	}
	now_t := time.Now().Unix()
	if err == mgo.ErrNotFound {
		err = c.Insert(bson.M{"id": sdkUid, "gift_id": int(giftId), "times": 1, "get_time": time.Now().Unix()})
		if err != nil {
			logs.Error("DBByMongoDB Insert err %v", err)
			return config.RetCode_SendFailed, false
		}
	} else {
		if time.Now().Day() != time.Unix(giftInfo.GetTime, 0).Day() {
			_, err := c.Upsert(YYBGift{SdkUid: sdkUid, GiftID: int(giftId)},
				bson.M{"$set": YYBGift{Times: giftInfo.Times + 1, GetTime: now_t}})
			if err != nil {
				logs.Error("DBByMongoDB Upsert err %v", err)
				return config.RetCode_SendFailed, false
			}
		} else {
			logs.Error("Get already err %v", err)
			return config.RetCode_HadSend, false
		}
	}

	return 0, true
}
