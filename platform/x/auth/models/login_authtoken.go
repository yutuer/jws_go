package models

import "vcs.taiyouxi.net/platform/planx/servers/db"

// LoginRegAuthToken
func LoginRegAuthToken(authToken string, userID db.UserID) error {
	var time_out int64 = AUTHTOKEN_TIMEOUT
	return db_interface.SetAuthToken(authToken, userID, time_out)
}

func LoginVerifyAuthToken(authToken string) (db.UserID, error) {
	return db_interface.GetAuthToken(authToken)
}
