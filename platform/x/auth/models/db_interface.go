package models

import (
	"errors"

	"vcs.taiyouxi.net/platform/planx/servers/db"
)

type DBInterface interface {
	Init(config DBConfig) error
	GetDeviceInfo(deviceID string, checkGM bool) (*deviceUserInfo, error, bool)
	UpdateDeviceInfo(deviceID string, info *deviceUserInfo) error
	SetDeviceInfo(deviceID, name string, uid db.UserID, channelID, device string) (*deviceUserInfo, error)
	IsNameExist(name string) (int, error)
	GetUnKey(name string, checkGM bool) (db.UserID, error, bool)
	SetUnKey(name string, uid db.UserID) error
	GetUnInfo(uid db.UserID) (string, string, int64, int64, string, error)
	UpdateUnInfo(uid db.UserID, deviceID, authToken string) error
	UpdateBanUn(uid string, time_to_ban int64, reason string) error
	UpdateGagUn(uid string, time_to_gag int64) error
	SetUnInfo(uid db.UserID, name, deviceID, passwd, email, authToken string) error
	SetUnInfoPass(uid db.UserID, name, deviceID, passwd, email, authToken string) error

	//IncrDeviceTotal() (db.UserID, error)
	GetAuthToken(authToken string) (db.UserID, error)
	SetAuthToken(authToken string, userID db.UserID, time_out int64) error

	GetUserShardInfo(uid string) ([]AuthUserShardInfo, error)
	SetUserShardInfo(uid, gidSidStr string, hasRole string) error

	SetLastLoginShard(uid string, gidsid string) error
	GetLastLoginShard(uid string) (string, error)

	SetPayFeedBackIfNot(uid string) (error, bool)
}

//XXX 曾经这里有Redis的使用接口，目前只使用DynamoDB
var db_interface DBInterface = &DBByDynamoDB{}

var (
	XErrDBNotFound = errors.New("DB Not Found")
)
