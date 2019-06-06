package models

import (
	"errors"
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/secure"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
	"vcs.taiyouxi.net/platform/x/auth/config"
)

func makeAuthUidKey(uid db.UserID) string {
	return fmt.Sprintf("uid:%s", uid)
}

// TODO 增加deviceUserInfo field(ban,gag, ...)
type deviceUserInfo struct {
	UserId     db.UserID `json:"user_id"`
	Display    string    `json:"dn,omitempty"`
	Name       string    `json:"name,omitempty"`
	LastTime   int64     `json:"lasttime"`
	CreateTime int64     `json:"createtime"`
	ChannelId  string    `json:"channel_id"`
}

// checkDevice 数据查询，检查是否存在当前deviceID的信息
// 返回值：
// - (deviceInfo， nil) 查询到，正常结束
// - (xxxx, error) 查询过程失败，返回错误信息
func checkDevice(deviceID string) (*deviceUserInfo, error, bool) {
	deviceinfo, err, isGM := db_interface.GetDeviceInfo(deviceID, true)
	switch err {
	case XErrDBNotFound:
		return nil, XErrChkDeviceNotFound, isGM
	case nil:
		logs.Warn("Check Device:%v", deviceinfo)
		db_interface.UpdateDeviceInfo(deviceID, deviceinfo)
		return deviceinfo, nil, isGM
	default:
		return nil, err, isGM
	}
}

type UserInfoToClient struct {
	GagTime int64 `json:"g"`
}

// AuthUserPass ...
// 返回逻辑错误：XErrAuthUsernameNotFound，XErrAuthUserPasswordInCorrect
func AuthUserPass(name, passwd string) (string, db.UserID, *UserInfoToClient, int64, string, error) {
	uid, err, isGM := db_interface.GetUnKey(name, true)
	res_info := &UserInfoToClient{}
	if err != nil {
		return "", db.InvalidUserID, res_info, 0, "", err
	}

	var authToken string
	if !isGM {
		dbpasswd := fmt.Sprintf("%x", secure.DefaultEncode.PasswordForDB(passwd))

		name_in_db, pswd_in_db, ban_time, gag_time, ban_reason, _ := db_interface.GetUnInfo(uid)
		if name_in_db != name || pswd_in_db != dbpasswd {
			return "", db.InvalidUserID, res_info, 0, "", XErrAuthUserPasswordInCorrect
		}
		if ban_time > 0 && time.Now().Unix() <= ban_time {
			return "", db.InvalidUserID, res_info, ban_time, ban_reason, XErrByBan
		}
		authToken, err = AuthToken(uid, "")
		res_info.GagTime = gag_time
	} else {
		logs.Warn("GM login by %s %s", name, uid.String())
		authToken = uuid.NewV4().String()
	}

	return authToken, uid, res_info, 0, "", err
}

func UserBanInfo(uid db.UserID) (error, int64, string) {
	_, _, ban_time, _, ban_reason, _ := db_interface.GetUnInfo(uid)
	logs.Debug(" UserBanInfo user %v ban_time %v", uid, ban_time)
	if ban_time > 0 && time.Now().Unix() <= ban_time {
		return XErrByBan, ban_time, ban_reason
	}
	return nil, 0, ""
}

// AuthToken ...
// 返回DB级error
func AuthToken(uid db.UserID, deviceID string) (string, error) {
	authToken := uuid.NewV4().String()
	err := db_interface.UpdateUnInfo(uid, deviceID, authToken)
	if err != nil {
		logs.Error("AuthToken UpdateUnInfo err %s", err.Error())
		return "", err
	}
	return authToken, nil
}

func SetBanUnInfo(uid string, time_to_ban int64, reason string) error {
	return db_interface.UpdateBanUn(uid, time_to_ban, reason)
}

func SetGagUnInfo(uid string, time_to_ban int64) error {
	return db_interface.UpdateGagUn(uid, time_to_ban)
}

var (
	XErrUserExistButNeedPassword = errors.New("User account exists but need login with user/pass only")
	XErrUserNotExist             = errors.New("User account does not exist.")
)

func TryToCheckDeviceInfo(DeviceID string) (*deviceUserInfo, error, bool) {

	di, err, isGM := checkDevice(DeviceID)
	if err != nil {
		switch err {
		case XErrChkDeviceNotFound:
			//可以创建新帐号
			return nil, XErrUserNotExist, isGM
		default:
			return nil, err, isGM
		}
	}
	//err == nil && di != nil
	if di != nil {
		//找到匿名帐号
		uid := di.UserId
		if uid.IsValid() {
			if di.Name != "" {
				//要求强制用用户名密码登录
				return di, XErrUserExistButNeedPassword, isGM
			}
			//可以匿名登录
			return di, nil, isGM
		}
	}

	return nil, fmt.Errorf("TryToCheckDeviceInfo got unknown errors."), isGM
}

// AddDevice 创建新UUID的然后返回uid
// error 返回数据库级和语言类错误
func AddDevice(DeviceID, Name, channelId, device string) (*deviceUserInfo, error) {
	//num, err := db_interface.IncrDeviceTotal()
	//if err != nil {
	//return nil, err
	//}

	//uid := db.UserID(num + 1000)
	uid := db.NewUserID()
	value, err := db_interface.SetDeviceInfo(DeviceID, Name, uid, channelId, device)
	if err != nil {
		return nil, err
	}
	config.AddAuthRegisterCount()
	return value, nil
}

var (
	XErrRegPwdUserNameUsed = errors.New("User name has been used.")
	XErrRegPwdUserHasName  = errors.New("You've already have user name registered.")
)

// RegisterWithPassword 注册用户email，当前user_id同用户名密码绑定
// 返回逻辑错误：XErrRegPwdUserNameUsed, XErrRegPwdUserHasName， 数据库级错误
func RegisterWithPassword(name, passwd, email, deviceID string) (db.UserID, error) {
	// 检测注册用户名, 已经有人使用
	isexist, err := db_interface.IsNameExist(name)
	if err != nil {
		return db.InvalidUserID, err //db error
	}
	if isexist == 1 {
		return db.InvalidUserID, XErrRegPwdUserNameUsed
	}
	var creatNewAccount bool
	var bindAccount bool
	di, err, _ := checkDevice(deviceID)
	if err != nil {
		switch err {
		case XErrChkDeviceNotFound:
			creatNewAccount = true
			bindAccount = true
			err = nil
		default:
			//数据库级错误
			logs.Error("Check Device Error:", err.Error())
		}
	} else if di != nil {
		//设备ID存在，检查是否已有用户名
		if di.UserId.IsValid() && di.Name != "" {
			err = XErrRegPwdUserHasName
		} else {
			//bind user name
			bindAccount = true
		}
	}

	if err != nil {
		return db.InvalidUserID, err
	}
	if creatNewAccount && bindAccount {
		//create new
		//检查是否已经存在
		var errn error
		di, errn = AddDevice(deviceID, name, "", "")
		if errn != nil {
			//返回数据库级和语言类错误
			logs.Error("Check Device Error:", errn.Error())
			return db.InvalidUserID, errn
		}
	} else if bindAccount {
		//bind
	} else {
		return db.InvalidUserID, err
	}

	uid := di.UserId

	//更新deviceid -> user_id的信息，方便处理强制有密码用户用密码登录
	/*
		info.Name = name

		deviceInfo, err := json.Marshal(info)
		if err != nil {
			return -1, err
		}
		_, err = db.Do("SET", deviceID, deviceInfo)
		if err != nil {
			return -2, err
		}
	*/

	//构建用户名索引，用户用户密码登录检索用
	err = db_interface.SetUnKey(name, uid)
	if err != nil {
		return db.InvalidUserID, err
	}

	//构建注册用户数据
	err = db_interface.SetUnInfo(uid, name, deviceID, passwd, email, "")
	if err != nil {
		return db.InvalidUserID, err
	}

	return uid, nil
}
