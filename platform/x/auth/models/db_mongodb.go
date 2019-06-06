package models

import (
	"errors"
	"time"

	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/secure"
)

type MongoAuth struct {
	DevicdID    string `bson:"id,omitempty"`
	DisplanName string `bson:"dn,omitempty"`
	LastGidSid  string `bson:"gid_sid,omitempty"`
	UserID      string `bson:"user_id,omitempty"`
	ChannelID   string `bson:"channelid,omitempty"`
	Device      string `bson:"device,omitempty"`

	LastAuthTime int64  `bson:"lasttime,omitempty"`
	CreateTime   int64  `bson:"createtime,omitempty"`
	BanTime      int64  `bson:"bantime,omitempty"`
	BanReason    string `bson:"banreason,omitempty"`
	GagTime      int64  `bson:"gagtime,omitempty"`

	NameAuth    string `bson:"name,omitempty"`
	NameAuthPwd string `bson:"pwd,omitempty"`

	AuthToken  string `bson:"authtoken,omitempty"`
	Payfeeback string `bson:"payfeedback,omitempty"`
}

func (ma *MongoAuth) GetDeviceUserInfo() deviceUserInfo {
	return deviceUserInfo{
		UserId:     db.UserIDFromStringOrNil(ma.UserID),
		Display:    ma.DisplanName,
		LastTime:   ma.LastAuthTime,
		CreateTime: ma.CreateTime,
		Name:       ma.NameAuth,
		ChannelId:  ma.ChannelID,
	}
}

type MongoUserShardInfo struct {
	UserID  string `bson:"user_id,omitempty"`
	GidSid  string `bson:"gid_sid,omitempty"`
	HasRole string `bson:"has_role,omitempty"`
}

type MongoTokens struct {
	UserID    string    `bson:"user_id,omitempty"`
	AuthToken string    `bson:"authtoken,omitempty"`
	CreateAt  time.Time `bson:"create_at,omitempty"`
}

const (
	MongoCollectionDevices        = "Devices"
	MongoCollectionTokens         = "UserInfos"
	MongoCollectionUserShardInfos = "UserShardInfos"
)

type DBByMongoDB struct {
	*mgo.Session
	DBName string
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

func (mdb *DBByMongoDB) Init(config DBConfig) error {
	logs.Info("DBByMongoDB Init %v", config)
	session, err := mgo.DialWithTimeout(config.MongoDBUrl, 20*time.Second)
	if err != nil {
		return err
	}
	mdb.Session = session
	mdb.DBName = config.MongoDBName

	session.SetMode(mgo.Monotonic, true)
	index := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c := mdb.GetDB().C(MongoCollectionDevices)
	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index2 := mgo.Index{
		Key:        []string{"user_id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c2 := mdb.GetDB().C(MongoCollectionDevices)
	if err := c2.EnsureIndex(index2); err != nil {
		return err
	}

	//Tokens
	{
		index := mgo.Index{
			Key:         []string{"create_at"},
			ExpireAfter: time.Second * AUTHTOKEN_TIMEOUT,
		}
		ct := mdb.GetDB().C(MongoCollectionTokens)
		if err := ct.EnsureIndex(index); err != nil {
			return err
		}

		index2 := mgo.Index{
			Key:        []string{"user_id"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		}
		ct2 := mdb.GetDB().C(MongoCollectionTokens)
		if err := ct2.EnsureIndex(index2); err != nil {
			return err
		}
	}

	indexName := mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	cName := mdb.GetDB().C(MongoCollectionDevices)
	if err := cName.EnsureIndex(indexName); err != nil {
		return err
	}

	index3 := mgo.Index{
		Key:        []string{"user_id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c3 := mdb.GetDB().C(MongoCollectionUserShardInfos)
	if err := c3.EnsureIndex(index3); err != nil {
		return err
	}
	return nil
}

func (mdb *DBByMongoDB) UpdateDeviceCollection(selector interface{}, update interface{}) error {

	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionDevices)

	query := c.Find(selector)
	if n, err := query.Count(); err != nil {
		return err
	} else if n != 1 {
		return fmt.Errorf("DBByMongoDB.UpdateDeviceCollection Find%v, %d account.", selector, n)
	}
	var ma MongoAuth
	if err := query.One(&ma); err != nil {
		return err
	}
	if err := c.Update(selector, update); err != nil {
		return err
	}

	return nil

}

// GetDeviceInfo 根据DeviceID查找是否存在已注册信息
func (mdb *DBByMongoDB) GetDeviceInfo(deviceID string, checkGM bool) (*deviceUserInfo, error, bool) {
	s := mdb.GetMongo()
	defer s.Close()

	c := s.GetDB().C(MongoCollectionDevices)
	var ma MongoAuth
	err := c.Find(bson.D{{"id", deviceID}}).One(&ma)
	switch err {
	case mgo.ErrNotFound:
		logs.Trace("DBByMongoDB.GetDeviceInfo not found devicdid")
		return nil, XErrDBNotFound, false
	case nil:
		di := ma.GetDeviceUserInfo()
		logs.Trace("DBByMongoDB.GetDeviceInfo %v", di)
		return &di, nil, false
	default:
		return nil, err, false

	}
	return nil, nil, false
}

// UpdateDeviceInfo 来更新之后的访问时间
func (mdb *DBByMongoDB) UpdateDeviceInfo(deviceID string, info *deviceUserInfo) error {
	now_t := time.Now().Unix()
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionDevices)
	_, err := c.Upsert(bson.D{{"id", deviceID}}, bson.M{"$set": bson.D{{"lasttime", now_t}}})

	return err
}

// SetDeviceInfo
func (mdb *DBByMongoDB) SetDeviceInfo(deviceID, name string, uid db.UserID,
	channelID, device string) (*deviceUserInfo, error) {
	now_t := time.Now().Unix()
	ma := MongoAuth{
		UserID:       uid.String(),
		DisplanName:  "anonymous",
		LastAuthTime: now_t,
		CreateTime:   now_t,
		NameAuth:     name,
		ChannelID:    channelID,
		Device:       device,
	}

	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionDevices)
	_, err := c.Upsert(bson.D{{"id", deviceID}}, bson.M{"$set": ma})
	switch err {
	case nil:
		dui := ma.GetDeviceUserInfo()
		logs.Trace("DBByMongoDB.SetDeviceInfo %v", ma.GetDeviceUserInfo())
		return &dui, nil
	default:
		return nil, err

	}
	return nil, nil
}

//IsNameExist 判断玩家是否使用了用户名模式
func (mdb *DBByMongoDB) IsNameExist(name string) (int, error) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionDevices)
	n, err := c.Find(bson.D{{"name", name}}).Count()
	if err != nil {
		return -1, nil
	}
	switch err {
	case mgo.ErrNotFound:
		return 0, nil
	case nil:
		//do nothing
	default:
		return -1, err
	}

	if n == 0 {
		return 0, nil
	}

	return 1, nil
}

//返回逻辑错误：XErrAuthUsernameNotFound
func (mdb *DBByMongoDB) GetUnKey(name string, checkGM bool) (db.UserID, error, bool) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionDevices)
	var ma MongoAuth
	err := c.Find(bson.D{{"name", name}}).One(&ma)
	switch err {
	case mgo.ErrNotFound:
		return db.InvalidUserID, XErrAuthUsernameNotFound, false
	case nil:
		return db.UserIDFromStringOrNil(ma.UserID), nil, false
	default:
		return db.InvalidUserID, err, false
	}
	return db.InvalidUserID, nil, false
}

func (mdb *DBByMongoDB) SetUnKey(name string, uid db.UserID) error {
	return mdb.UpdateDeviceCollection(bson.D{{"user_id", uid.String()}},
		bson.M{"$set": bson.D{{"name", name}}})
}

// GetUnInfo
// return value: name, password, bantime, gagtime, error
func (mdb *DBByMongoDB) GetUnInfo(uid db.UserID) (string, string, int64, int64, string, error) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionDevices)
	var ma MongoAuth
	err := c.Find(bson.D{{"user_id", uid.String()}}).One(&ma)
	if err != nil {
		return "", "", 0, 0, "", nil
	}

	return ma.NameAuth, ma.NameAuthPwd, ma.BanTime, ma.GagTime, ma.BanReason, nil
}

//UpdateUnInfo 目前用在更新AuthToken
//XXX by YZH UserInfoTable
func (mdb *DBByMongoDB) UpdateUnInfo(uid db.UserID, deviceID, authToken string) error {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionDevices)
	if deviceID != "" {
		c.Upsert(bson.D{{"id", deviceID}},
			bson.M{"$set": MongoAuth{
				AuthToken: authToken,
			}})
	} else if uid.IsValid() {
		c.Upsert(bson.D{{"id", deviceID}},
			bson.M{"$set": MongoAuth{
				AuthToken: authToken,
			}})
	}
	return nil
}

func (mdb *DBByMongoDB) UpdateBanUn(uid string, time_to_ban int64, reason string) error {
	bantime := time.Now().Unix() + time_to_ban
	return mdb.UpdateDeviceCollection(bson.D{{"user_id", uid}},
		bson.M{"$set": MongoAuth{BanTime: bantime, BanReason: reason}})
}

func (mdb *DBByMongoDB) UpdateGagUn(uid string, time_to_gag int64) error {
	gagtime := time.Now().Unix() + time_to_gag
	return mdb.UpdateDeviceCollection(bson.D{{"user_id", uid}},
		bson.M{"$set": MongoAuth{GagTime: gagtime}})

}

func (mdb *DBByMongoDB) SetUnInfo(uid db.UserID, name, deviceID, passwd, email, authToken string) error {
	dbpasswd := fmt.Sprintf("%x", secure.DefaultEncode.PasswordForDB(passwd))
	return mdb.SetUnInfoPass(uid, name, deviceID, dbpasswd, email, authToken)
}

//SetUnInfoPass GMTools中，用来乾坤大挪移。用来把一个accountid映射到一个有用户名密码的帐号下。方便查找问题。
func (mdb *DBByMongoDB) SetUnInfoPass(uid db.UserID, name, deviceID, dbpasswd, email, authToken string) error {
	now_t := time.Now().Unix()
	//FIXME by YZH UpdateDeviceCollection 使用了双索引,两个独立索引,能起到加速索引作用吗?
	//FIXME by YZH 这里应该是insert而不应该是Update

	return mdb.UpdateDeviceCollection(bson.D{{"id", deviceID}, {"user_id", uid.String()}},
		bson.M{"$set": MongoAuth{
			NameAuth:     name,
			NameAuthPwd:  dbpasswd,
			LastAuthTime: now_t,
			//CreateTime: //now_t, 因为MongoDB使用一个Collection就能够处理原来的多个表,因此这个创建时间不应该更新了。
			BanTime: 0,
			GagTime: 0,
		}})
}

func (mdb *DBByMongoDB) GetAuthToken(authToken string) (db.UserID, error) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionTokens)
	var mt MongoTokens
	err := c.Find(MongoTokens{AuthToken: authToken}).One(&mt)
	switch err {
	case mgo.ErrNotFound:
		return db.InvalidUserID, XErrLoginAuthtokenNotFound
	case nil:
		//do nothing
	default:
		return db.InvalidUserID, err
	}
	return db.UserIDFromStringOrNil(mt.UserID), nil
}

func (mdb *DBByMongoDB) SetAuthToken(authToken string, userID db.UserID, time_out int64) error {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionTokens)
	if _, err := c.Upsert(MongoTokens{UserID: userID.String()}, MongoTokens{
		UserID:    userID.String(),
		AuthToken: authToken,
		CreateAt:  time.Now(),
	}); err != nil {
		return err
	}
	return nil
}

func (mdb *DBByMongoDB) GetUserShardInfo(uid string) ([]AuthUserShardInfo, error) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionUserShardInfos)
	var result []AuthUserShardInfo
	err := c.Find(MongoUserShardInfo{UserID: uid}).All(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (mdb *DBByMongoDB) SetUserShardInfo(uid, gidSidStr string, hasRole string) error {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionUserShardInfos)

	_, err := c.Upsert(MongoUserShardInfo{UserID: uid},
		bson.M{"$set": MongoUserShardInfo{
			UserID:  uid,
			GidSid:  gidSidStr,
			HasRole: hasRole,
		}})
	return err
}

func (mdb *DBByMongoDB) SetLastLoginShard(uid string, gidsid string) error {
	return mdb.UpdateDeviceCollection(bson.D{{"user_id", uid}},
		bson.M{"$set": bson.D{{"gid_sid", gidsid}}})

}

func (mdb *DBByMongoDB) GetLastLoginShard(uid string) (string, error) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionDevices)

	query := c.Find(bson.D{{"user_id", uid}})
	if n, err := query.Count(); err != nil {
		return "", err
	} else if n != 1 {
		return "", errors.New("DBByMongoDB.GetLastLoginShard Find more than one account.")
	}
	var ma MongoAuth
	if err := query.One(&ma); err != nil {
		return "", err
	}

	if ma.LastGidSid == "" {
		return "", nil
	}

	return ma.LastGidSid, nil
}

func (mdb *DBByMongoDB) SetPayFeedBackIfNot(uid string) (error, bool) {
	s := mdb.GetMongo()
	defer s.Close()
	c := s.GetDB().C(MongoCollectionDevices)

	query := c.Find(bson.D{{"user_id", uid}})
	if n, err := query.Count(); err != nil {
		return err, false
	} else if n != 1 {
		return errors.New("DBByMongoDB.GetLastLoginShard Find more than one account."), false
	}
	var ma MongoAuth
	if err := query.One(&ma); err != nil {
		return err, false
	}

	if ma.Payfeeback != "got" {
		if err := c.Update(bson.D{{"user_id", uid}},
			bson.M{"$set": bson.D{{"payfeedback", "got"}}}); err != nil {
			return err, false
		}
		return nil, true
	}
	return nil, false
}
