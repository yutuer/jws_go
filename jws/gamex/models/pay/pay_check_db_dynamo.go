package pay

/*
import (
	"errors"
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/secure"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

type DBByDynamoDB struct {
	db *dynamodb.DynamoDB
}

func (d *DBByDynamoDB) Init(config DBConfig) error {
	logs.Info("DBByDynamoDB init")
	d.db = &dynamodb.DynamoDB{}
	err := d.db.Connect("", "", "", "")
	if err != nil {
		return err
	}

	err = d.db.InitTable()
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	d.db.CreateHashTable(cfg.Cfg.Dynamo_NameDevice, "Id", "S")

	err = d.db.InitTable()
	if err != nil {
		return err
	}
	return nil
}

func getFromAnys2String(typ, name string, data map[string]dynamodb.Any) (string, bool) {
	data_any, name_ok := data[name]
	if !name_ok {
		logs.Error("getFromAnys2String %s Err by no %s", typ, name)
		return "", false
	}

	data_str, ok := data_any.(string)
	if !ok {
		logs.Error("getFromAnys2String %s Err by %s no string", typ, name)
		return "", false
	}

	return data_str, true
}

func getFromAnys2int64(typ, name string, data map[string]dynamodb.Any) (int64, bool) {
	data_any, name_ok := data[name]
	if !name_ok {
		logs.Error("getFromAnys2int64 %s Err by no %s", typ, name)
		return 0, false
	}

	data_int64, ok := data_any.(int64)
	if !ok {
		logs.Error("getFromAnys2int64 %s Err by %s no string", typ, name)
		return 0, false
	}

	return data_int64, true
}

func (d *DBByDynamoDB) getTranIDInDB(tid string) string {
	return fmt.Sprintf("tran:%s", name)
}

func (d *DBByDynamoDB) IsTransIDExist(tid string) (int, error) {
	userNameKey := getTranIDInDB(tid)
	data, err := d.db.GetByHash(TranInfoDBName, userNameKey)
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
func (d *DBByDynamoDB) GetUnKey(name string) (db.UserID, error) {
	userNameKey := getTranIDInDB(tid)
	data, err := d.db.GetByHash(cfg.Cfg.Dynamo_NameName, userNameKey)
	if err != nil {
		return db.InvalidUserID, err
	}
	uids, ok := data.(string)
	if !ok {
		return db.InvalidUserID, XErrAuthUsernameNotFound
		//return db.InvalidUserID, errors.New(fmt.Sprintf("err:%s %s", data, name))
	}
	return db.UserIDFromStringOrNil(uids), nil

}

func (d *DBByDynamoDB) SetUnKey(name string, uid db.UserID) error {
	userNameKey := getTranIDInDB(tid)
	logs.Error("SetUnKey:", uid)
	// 注意db.UserID和int64在type比较时是不一样的
	return d.db.SetByHash(cfg.Cfg.Dynamo_NameName, userNameKey, uid.String())
}

func (d *DBByDynamoDB) GetUnInfo(uid db.UserID) (string, string, int64, int64, error) {
	uidkey := makeAuthUidKey(uid)

	res, err := d.db.GetByHashM(cfg.Cfg.Dynamo_NameUserInfo, uidkey)
	if err != nil {
		return "", "", 0, 0, err
	}
	fmt.Printf("data1 : %v \n", res)
	name_any, name_ok := res["name"]
	pass_any, pass_ok := res["pwd"]
	if !(name_ok && pass_ok) {
		return "", "", 0, 0, errors.New("UserInfo data Err")
	}
	name, name_ok := name_any.(string)
	pass, pass_ok := pass_any.(string)
	if name_ok && pass_ok {
		bantime, _ := getFromAnys2int64("GetUnInfo", "bantime", res)
		gagtime, _ := getFromAnys2int64("GetUnInfo", "gagtime", res)

		return name, pass, bantime, gagtime, nil

	} else {
		return "", "", 0, 0, errors.New("UserInfo data typ Err")
	}
}

func (d *DBByDynamoDB) UpdateUnInfo(uid db.UserID, deviceID, authToken string) error {
	uidkey := makeAuthUidKey(uid)
	data := make(map[string]dynamodb.Any, 4)

	data["authtoken"] = authToken
	if deviceID != "" {
		data["device"] = deviceID
	}
	data["lasttime"] = time.Now().Unix()
	return d.db.UpdateByHash(cfg.Cfg.Dynamo_NameUserInfo, uidkey, data)
}

func (d *DBByDynamoDB) UpdateBanUn(uid string, time_to_ban int64) error {
	uidkey := fmt.Sprintf("uid:%s", uid)
	data := make(map[string]dynamodb.Any, 4)

	data["bantime"] = time.Now().Unix() + time_to_ban
	return d.db.UpdateByHash(cfg.Cfg.Dynamo_NameUserInfo, uidkey, data)
}

func (d *DBByDynamoDB) UpdateGagUn(uid string, time_to_gag int64) error {
	uidkey := fmt.Sprintf("uid:%s", uid)
	data := make(map[string]dynamodb.Any, 4)

	data["gagtime"] = time.Now().Unix() + time_to_gag
	return d.db.UpdateByHash(cfg.Cfg.Dynamo_NameUserInfo, uidkey, data)
}

func (d *DBByDynamoDB) SetUnInfo(uid db.UserID,
	name, deviceID, passwd, email, authToken string) error {
	now_t := time.Now().Unix()
	dbpasswd := fmt.Sprintf("%x", secure.DefaultEncode.PasswordForDB(passwd))
	data := map[string]dynamodb.Any{
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
*/
