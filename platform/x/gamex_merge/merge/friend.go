package merge

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/modules/friend"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

func updateAccountFriend(acid string, resDB redispool.RedisPoolConn, cb redis.CmdBuffer) error {
	accountID, err := db.ParseAccount(acid)
	if err != nil {
		logs.Error("updateAccountFriend ParseAccount %s err %s", acid, err.Error())
		return err
	}

	// profile
	a := &account.Account{
		AccountID: accountID,
		Friend:    account.NewFriend(accountID),
	}
	f := &a.Friend

	key := f.DBName()

	err = driver.RestoreFromHashDB(resDB.RawConn(), key, f, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		logs.Error("updateAccountFriend loadDB err %s %s", key, err.Error())
		return err
	}

	nFriends := make([]friend.FriendSimpleInfo, 0, 100)
	for _, friend := range f.FriendList {
		if isAcidDel(friend.AcID) {
			continue
		}
		nname, ok := account_name_old2new[friend.Name]
		if ok {
			friend.Name = nname
		}
		nFriends = append(nFriends, friend)
	}
	nBlacks := make([]friend.FriendSimpleInfo, 0, 100)
	for _, black := range f.BlackList {
		if isAcidDel(black.AcID) {
			continue
		}
		nname, ok := account_name_old2new[black.Name]
		if ok {
			black.Name = nname
		}
		nBlacks = append(nBlacks, black)
	}

	f.FriendList = nFriends
	f.BlackList = nBlacks
	f.FriendRecommend = make([]friend.FriendSimpleInfo, 0, 1)
	f.LastUpdateFriendListTime = 0

	err = driver.DumpToHashDBCmcBuffer(cb, key, f)
	if err != nil {
		logs.Error("updateAccountFriend saveDB err %s", err.Error())
		return err
	}

	return nil
}
