package models

import (
	"errors"
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	cfg "vcs.taiyouxi.net/platform/x/auth/config"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/secure"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

type DBByDynamoDB struct {
	db *AuthDynamoDB
}

func (d *DBByDynamoDB) Init(config DBConfig) error {
	logs.Debug("DBByDynamoDB init")
	d.db = &AuthDynamoDB{DynamoDB: &dynamodb.DynamoDB{}}
	d.db.Connect(
		config.DynamoRegion,
		config.DynamoAccessKeyID,
		config.DynamoSecretAccessKey,
		config.DynamoSessionToken)

	// DynamoDB表操作之间需要等待一段时间
	// TODO 这里初始化DynamoDB的操作应该单独一个小工具

	err := d.db.InitTable()
	if err != nil {
		logs.Error("DBByDynamoDB Init err %s", err.Error())
		return err
	}

	//time.Sleep(1 * time.Second)
	d.db.CreateHashTable(cfg.Cfg.Dynamo_NameDevice, "Id", "S")
	//time.Sleep(1 * time.Second)
	d.db.CreateHashTable(cfg.Cfg.Dynamo_NameName, "Name", "S")
	//time.Sleep(1 * time.Second)
	d.db.CreateHashTable(cfg.Cfg.Dynamo_NameUserInfo, "UId", "S")

	d.db.CreateHashTable(cfg.Cfg.Dynamo_GM, "device", "S")
	//time.Sleep(1 * time.Second)
	//d.db.CreateHashTable(cfg.Cfg.Dynamo_NameAuthToken, "Token", "S")
	//time.Sleep(5 * time.Second)

	err = d.db.InitTable()
	if err != nil {
		logs.Error("DBByDynamoDB Init err %s", err.Error())
		return err
	}
	return nil
}

func getFromAnys2String(typ, name string, data map[string]interface{}) (string, bool) {
	data_any, name_ok := data[name]
	if !name_ok {
		logs.Warn("getFromAnys2String %s Err by no %s", typ, name)
		return "", false
	}

	data_str, ok := data_any.(string)
	if !ok {
		logs.Warn("getFromAnys2String %s Err by %s no string", typ, name)
		return "", false
	}

	return data_str, true
}

func getFromAnys2int64(typ, name string, data map[string]interface{}) (int64, bool) {
	data_any, name_ok := data[name]
	if !name_ok {
		logs.Warn("getFromAnys2int64 %s Err by no %s", typ, name)
		return 0, false
	}

	data_int64, ok := data_any.(int64)
	if !ok {
		logs.Warn("getFromAnys2int64 %s Err by %s no string", typ, name)
		return 0, false
	}

	return data_int64, true
}

// GetDeviceInfo
// 返回数据库级错误，或者XErrDBNotFound
func (d *DBByDynamoDB) GetDeviceInfo(deviceID string, checkGM bool) (
	info *deviceUserInfo, err error, isGM bool) {
	di := &deviceUserInfo{}
	// 首先去gm表查
	if checkGM {
		data_uid, err := d.db.GetByHash(cfg.Cfg.Dynamo_GM, deviceID)
		if err == nil {
			uids, ok := data_uid.(string)
			if ok {
				di.UserId = db.UserIDFromStringOrNil(uids)
				logs.Warn("GM login by %s %s", deviceID, di.UserId.String())
				return di, nil, true
			}
		}
	}

	data, err := d.db.GetByHashM(cfg.Cfg.Dynamo_NameDevice, deviceID)
	if err != nil {
		return nil, err, false
	}

	ok := true
	user_id, uok := getFromAnys2String("device", "user_id", data)
	ok, di.UserId = ok && uok, db.UserIDFromStringOrNil(user_id)

	dn, uok := getFromAnys2String("device", "dn", data)
	ok, di.Display = ok && uok, dn

	lasttime, uok := getFromAnys2int64("device", "lasttime", data)
	ok, di.LastTime = ok && uok, lasttime

	createtime, uok := getFromAnys2int64("device", "createtime", data)
	ok, di.CreateTime = ok && uok, createtime

	logs.Trace("GetDeviceInfo %v %v", ok, *di)

	if ok {
		// 可能这个没有
		di.Name, _ = getFromAnys2String("device", "name", data)
		di.ChannelId, _ = getFromAnys2String("device", "channelid", data)

		return di, nil, false
	} else {
		return nil, XErrDBNotFound, false
	}

}

func (d *DBByDynamoDB) UpdateDeviceInfo(deviceID string, info *deviceUserInfo) error {
	now_t := time.Now().Unix()
	data := map[string]interface{}{
		"lasttime": now_t,
	}
	return d.db.UpdateByHash(cfg.Cfg.Dynamo_NameDevice, deviceID, data)
}

func (d *DBByDynamoDB) SetDeviceInfo(deviceID, name string, uid db.UserID,
	channelID, device string) (*deviceUserInfo, error) {
	// 当需要用用户名登陆时，需要加入Name信息
	now_t := time.Now().Unix()
	data := map[string]interface{}{
		"user_id":    uid.String(),
		"dn":         "anonymous",
		"lasttime":   now_t,
		"createtime": now_t,
	}
	if name != "" {
		data["name"] = name
	}
	if channelID != "" {
		data["channelid"] = channelID
	}
	if device != "" {
		data["device"] = device
	}
	value := &deviceUserInfo{
		UserId:     uid,
		Display:    "anonymous",
		Name:       name,
		LastTime:   now_t,
		CreateTime: now_t,
		ChannelId:  channelID,
	}
	err := d.db.SetByHashM(cfg.Cfg.Dynamo_NameDevice, deviceID, data)
	return value, err
}

func (d *DBByDynamoDB) IsNameExist(name string) (int, error) {
	userNameKey := fmt.Sprintf("un:%s", name)
	data, err := d.db.GetByHash(cfg.Cfg.Dynamo_NameName, userNameKey)
	if err != nil {
		return -1, err
	}
	if data == nil {
		return 0, nil
	} else {
		return 1, nil
	}
}

//返回逻辑错误：XErrAuthUsernameNotFound
func (d *DBByDynamoDB) GetUnKey(name string, checkGM bool) (
	user db.UserID, error error, isGM bool) {
	// 首先去gm表查
	if checkGM {
		data, err := d.db.GetByHash(cfg.Cfg.Dynamo_GM, name)
		if err == nil {
			uids, ok := data.(string)
			if ok {
				return db.UserIDFromStringOrNil(uids), nil, true
			}
		}
	}

	userNameKey := fmt.Sprintf("un:%s", name)
	data, err := d.db.GetByHash(cfg.Cfg.Dynamo_NameName, userNameKey)
	if err != nil {
		return db.InvalidUserID, err, false
	}
	uids, ok := data.(string)
	if !ok {
		return db.InvalidUserID, XErrAuthUsernameNotFound, false
	}
	return db.UserIDFromStringOrNil(uids), nil, false
}

func (d *DBByDynamoDB) SetUnKey(name string, uid db.UserID) error {
	userNameKey := fmt.Sprintf("un:%s", name)
	logs.Error("SetUnKey:", uid)
	// 注意db.UserID和int64在type比较时是不一样的
	return d.db.SetByHash(cfg.Cfg.Dynamo_NameName, userNameKey, uid.String())
}

func (d *DBByDynamoDB) GetUnInfo(uid db.UserID) (string, string, int64, int64, string, error) {
	uidkey := makeAuthUidKey(uid)

	res, err := d.db.GetByHashM(cfg.Cfg.Dynamo_NameUserInfo, uidkey)
	if err != nil {
		return "", "", 0, 0, "", err
	}
	name_any, name_ok := res["name"]
	pass_any, pass_ok := res["pwd"]

	name := ""
	pass := ""
	if name_ok {
		name, _ = name_any.(string)
	}
	if pass_ok {
		pass, _ = pass_any.(string)
	}
	reason := ""
	reason_any, ok := res["banreason"]
	if ok {
		reason, _ = reason_any.(string)
	}
	bantime, _ := getFromAnys2int64("GetUnInfo", "bantime", res)
	gagtime, _ := getFromAnys2int64("GetUnInfo", "gagtime", res)
	return name, pass, bantime, gagtime, reason, nil
}

func (d *DBByDynamoDB) UpdateUnInfo(uid db.UserID, deviceID, authToken string) error {
	uidkey := makeAuthUidKey(uid)
	data := make(map[string]interface{}, 4)

	data["authtoken"] = authToken
	if deviceID != "" {
		data["device"] = deviceID
	}
	data["lasttime"] = time.Now().Unix()
	return d.db.UpdateByHash(cfg.Cfg.Dynamo_NameUserInfo, uidkey, data)
}

func (d *DBByDynamoDB) UpdateBanUn(uid string, time_to_ban int64, reason string) error {
	uidkey := fmt.Sprintf("uid:%s", uid)
	data := make(map[string]interface{}, 4)

	data["bantime"] = time.Now().Unix() + time_to_ban
	data["banreason"] = reason
	return d.db.UpdateByHash(cfg.Cfg.Dynamo_NameUserInfo, uidkey, data)
}

func (d *DBByDynamoDB) UpdateGagUn(uid string, time_to_gag int64) error {
	uidkey := fmt.Sprintf("uid:%s", uid)
	data := make(map[string]interface{}, 4)

	data["gagtime"] = time.Now().Unix() + time_to_gag
	return d.db.UpdateByHash(cfg.Cfg.Dynamo_NameUserInfo, uidkey, data)
}

func (d *DBByDynamoDB) SetLastLoginShard(uid string, gidsid string) error {
	uidkey := fmt.Sprintf("uid:%s", uid)
	data := make(map[string]interface{}, 4)

	data["gid_sid"] = gidsid
	return d.db.UpdateByHash(cfg.Cfg.Dynamo_NameUserInfo, uidkey, data)
}

func (d *DBByDynamoDB) SetPayFeedBackIfNot(uid string) (error, bool) {
	uidkey := fmt.Sprintf("uid:%s", uid)
	m, err := d.db.GetByHashM(cfg.Cfg.Dynamo_NameUserInfo, uidkey)
	if err != nil {
		return err, false
	}
	pfb, ok := getFromAnys2String("payfeedback", "payfeedback", m)
	if !ok || pfb != "got" {
		if err := d.db.UpdateByHash(cfg.Cfg.Dynamo_NameUserInfo, uidkey, map[string]interface{}{
			"payfeedback": "got",
		}); err != nil {
			return err, false
		}
		return nil, true
	}
	return nil, false
}

func (d *DBByDynamoDB) GetLastLoginShard(uid string) (string, error) {
	uidkey := fmt.Sprintf("uid:%s", uid)

	res, err := d.db.GetByHashM(cfg.Cfg.Dynamo_NameUserInfo, uidkey)
	if err != nil {
		return "", err
	}
	sidgid_any, sidgid_ok := res["gid_sid"]
	if !(sidgid_ok) {
		return "", nil
	}
	sidgid, sidgid_ok := sidgid_any.(string)
	if sidgid_ok {
		return sidgid, nil
	} else {
		return "", errors.New("UserInfo sidgid data typ Err")
	}
}

func (d *DBByDynamoDB) SetUnInfo(uid db.UserID,
	name, deviceID, passwd, email, authToken string) error {
	dbpasswd := fmt.Sprintf("%x", secure.DefaultEncode.PasswordForDB(passwd))
	return d.SetUnInfoPass(uid, name, deviceID, dbpasswd, email, authToken)
}

func (d *DBByDynamoDB) SetUnInfoPass(uid db.UserID,
	name, deviceID, dbpasswd, email, authToken string) error {
	now_t := time.Now().Unix()
	data := map[string]interface{}{
		"device":     deviceID,
		"pwd":        dbpasswd,
		"lasttime":   now_t,
		"createtime": now_t,
		"bantime":    0,
		"gagtime":    0,
	}
	if name != "" {
		data["name"] = name
	}
	if email != "" {
		data["email"] = email
	}
	uidkey := makeAuthUidKey(uid)
	return d.db.SetByHashM(cfg.Cfg.Dynamo_NameUserInfo, uidkey, data)
}

func (d *DBByDynamoDB) GetAuthToken(authToken string) (db.UserID, error) {
	_db := authRedisPool.Get()
	defer _db.Close()
	key := makeAuthTokenKey(authToken)
	r, err := redis.String(_db.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return db.InvalidUserID, XErrLoginAuthtokenNotFound
		} else {
			return db.InvalidUserID, err
		}
	}
	return db.UserIDFromStringOrNil(r), nil
}

func (d *DBByDynamoDB) SetAuthToken(authToken string, userID db.UserID, time_out int64) error {
	_db := authRedisPool.Get()
	defer _db.Close()
	key := makeAuthTokenKey(authToken)
	r, err := redis.String(_db.Do("SETEX", key, time_out, userID))
	if err != nil {
		return err
	}

	if r != "OK" {
		return fmt.Errorf("LoginRegAuthToken return is not OK in DB")
	}
	return nil
}

func (d *DBByDynamoDB) GetUserShardInfo(uid string) ([]AuthUserShardInfo, error) {
	info, err := d.db.QueryUserShardInfo(cfg.Cfg.Dynamo_UserShardInfo, uid)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (d *DBByDynamoDB) SetUserShardInfo(uid, gidSidStr string, hasRole string) error {
	return d.db.SetUserShardInfo(cfg.Cfg.Dynamo_UserShardInfo, uid, gidSidStr, hasRole)
}

func (d *DBByDynamoDB) GetDB() *AuthDynamoDB {
	return d.db
}
