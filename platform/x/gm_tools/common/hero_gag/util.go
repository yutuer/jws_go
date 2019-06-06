package hero_gag

import (
	"encoding/json"
	"fmt"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/authdb"
)

func GetAuthData(userID string) (string, error) {
	adb := authdb.GetDB()
	uid := db.UserIDFromStringOrNil(userID)

	name_in_db, pswd_in_db, ban_time, gag_time, ban_reason, err := adb.GetUnInfo(uid)

	if err != nil {
		return "", err
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
		return "", err
	}
	return string(j), err
}

func errRes(err string) string {
	return fmt.Sprintf("{\"err\":\"%s\"}", err)
}
