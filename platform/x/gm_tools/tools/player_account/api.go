package player_account

import (
	"encoding/json"

	"strings"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/authdb"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("getAccountByDeviceID", getAccountByDeviceID)
	gm_command.AddGmCommandHandle("getAccountByName", getAccountByName)
	gm_command.AddGmCommandHandle("createNamePassForUID", createNamePassForUID)
	gm_command.AddGmCommandHandle("createDeviceForUID", createDeviceForUID)
}

func getAccountByDeviceID(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("getAccountByDeviceID %v --> %v", server, params)

	db := authdb.GetDB()
	info, err, _ := db.GetDeviceInfo(params[0], false)

	if err != nil {
		return err
	}

	j, err := json.Marshal(*info)
	if err != nil {
		return err
	}

	c.SetData(string(j))
	return nil
}

func getAccountByName(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("getAccountByName %s %v --> %v", server, params)

	adb := authdb.GetDB()
	uid, err, _ := adb.GetUnKey(params[0], false)
	if err != nil {
		return err
	}

	name_in_db, pswd_in_db, ban_time, gag_time, ban_reason, err := adb.GetUnInfo(uid)

	if err != nil {
		return err
	}

	info := struct {
		UserId    db.UserID `json:"user_id"`
		NameInDB  string    `json:"name_in_db"`
		PswdInDB  string    `json:"pswd_in_db"`
		BanTime   int64     `json:"ban_time"`
		GagTime   int64     `json:"gag_time"`
		BanReason string    `json:"banreason"`
	}{
		uid,
		name_in_db,
		pswd_in_db,
		ban_time,
		gag_time,
		ban_reason,
	}

	j, err := json.Marshal(info)
	if err != nil {
		return err
	}

	c.SetData(string(j))
	return nil
}

func createNamePassForUID(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("CreateNamePassForUID  %v -%v-> %v", server, accountid, params)

	//
	// Name
	// Name表  un:{name} -> uid
	// UnInfo表 uid:{uid} -> name pass
	// models.AuthUserPass(strings.ToLower(string(namedb)), string(md5pwd))
	// "1" -> c355d4b7ab61ce6766783d2d738e36b1

	adb := authdb.GetDB()

	uid := db.UserIDFromStringOrNil(accountid)
	name := strings.ToLower(params[0])
	passMd5 := "c355d4b7ab61ce6766783d2d738e36b1" // 明文密码为1

	if err := adb.SetUnKey(name, uid); err != nil {
		return err
	}
	if err := adb.SetUnInfoPass(uid, name, "byGM@apple.com", passMd5, "byGM@taiyouxi.cn", ""); err != nil {
		return err
	}

	c.SetData("ok")
	return nil
}

func createDeviceForUID(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("CreateDeviceForUID %s %v --> %v", server, params)

	adb := authdb.GetDB()

	//
	// Device表

	uid := db.UserIDFromStringOrNil(accountid)
	deviceID := params[0]

	_, err := adb.SetDeviceInfo(deviceID, "", uid, "", "")
	if err != nil {
		return err
	}

	c.SetData("ok")
	return nil
}
