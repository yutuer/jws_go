package gm_command

import (
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/login"
)

func GetAccountByKey(key string) *login.Account {
	logs.Trace("GetAccountByKey %s", key)
	logs.Trace("GetAccountByKey %v", login.AccountManager.Accounts)
	return login.AccountManager.GetByKey(key)
}

func CheckGrant(key, command_typ string) bool {
	logs.Trace("GetAccountByKey %s", key)
	acc := login.AccountManager.GetByKey(key)
	if acc == nil {
		return false
	}

	t, ok := grantByTyp[command_typ]
	if !ok {
		return false
	} else {
		return acc.IsGrant(t, 1)
	}
}
