package account

import (
	"reflect"
	"time"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/friend"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
)

type Friend struct {
	dbkey        db.ProfileDBKey
	dirtiesCheck map[string]interface{}
	CreateTime   int64 `redis:"createtime"`

	FriendList      []friend.FriendSimpleInfo `redis:"friend_list"`
	BlackList       []friend.FriendSimpleInfo `redis:"black_list"`
	FriendRecommend []friend.FriendSimpleInfo `redis:"friend_rcmd"`

	LastUpdateFriendListTime      int64 `redis:"last_u_f_list_t"`
	LastUpdateBlackListTime       int64 `redis:"last_u_b_list_t"`
	LastUpdateRecentPlayerTime    int64
	LastUpdateRecommendPlayerTime int64 `redis:"last_u_rp_list_t"`

	// 给好友送包子的记录
	RefreshGiftTime  int64    `redis:"ref_gift_t"`
	ReceiveGiftTimes int      `redis:"receive_guft_t"`
	GiftInfo         []string `redis:"gift_info"`
}

func NewFriend(account db.Account) Friend {
	now_t := time.Now().Unix()
	f := Friend{
		dbkey: db.ProfileDBKey{
			Account: account,
			Prefix:  "friend",
		},
		//Ver:        helper.CurrDBVersion,
		CreateTime:      now_t,
		FriendList:      make([]friend.FriendSimpleInfo, 0, 100),
		BlackList:       make([]friend.FriendSimpleInfo, 0, 100),
		FriendRecommend: make([]friend.FriendSimpleInfo, 0, friend.FriendCountPerReq),
		GiftInfo:        make([]string, 0, 100),
	}
	//// Debug for test
	//for i := 0; i < 40; i++ {
	//	f.FriendList = append(f.FriendList, friend.FriendSimpleInfo{})
	//	f.BlackList = append(f.BlackList, friend.FriendSimpleInfo{})
	//}
	//for i := 0; i < 50; i++ {
	//	f.FriendRecommend = append(f.FriendRecommend, friend.FriendSimpleInfo{})
	//}
	//logs.Debug("Debug Success!")
	return f
}

func (f *Friend) AddBlack(friendInfo friend.FriendSimpleInfo) bool {
	if !f.IsInBlackList(friendInfo.AcID) &&
		uint32(len(f.BlackList)) <= gamedata.GetFriendConfig().GetBlakeListNum() {
		f.BlackList = append(f.BlackList, friendInfo)
		return true
	}
	return false
}

func (f *Friend) RemoveBlack(acID string) bool {
	index := f.getFriendIndex(acID, f.BlackList)
	if index == -1 {
		logs.Warn("No Info in blackList by acid: %s", acID)
		return false
	}
	f.BlackList = append(f.BlackList[:index], f.BlackList[index+1:]...)
	return true
}

func (f *Friend) AddFriend(friendInfo friend.FriendSimpleInfo) bool {
	if !f.IsInList(friendInfo.AcID) &&
		uint32(len(f.FriendList)) <= gamedata.GetFriendConfig().GetFriendNum() {
		f.FriendList = append(f.FriendList, friendInfo)
		return true
	}
	return false
}

func (f *Friend) IsInFriendList(acID string) bool {
	for _, item := range f.FriendList {
		if item.AcID == acID {
			return true
		}
	}
	return false
}

func (f *Friend) IsInBlackList(acID string) bool {
	for _, item := range f.BlackList {
		if item.AcID == acID {
			return true
		}
	}
	return false
}

func (f *Friend) GetFriendIDByName(sid uint, name string) (string, error) {
	conn := driver.GetDBConn()
	defer conn.Close()
	return redis.String(conn.Do("HGET", driver.TableChangeName(sid), name))
}

func (f *Friend) RemoveFriend(acID string) {
	index := f.getFriendIndex(acID, f.FriendList)
	if index == -1 {
		logs.Warn("No Info in friendList by acid: %s", acID)
		return
	}
	f.FriendList = append(f.FriendList[:index], f.FriendList[index+1:]...)
}

func (f *Friend) getFriendIndex(acID string, list []friend.FriendSimpleInfo) int {
	for i, item := range list {
		if item.AcID == acID {
			return i
		}
	}
	return -1
}

func (f *Friend) UpdateRecommendPlayer(selfID string, players []friend.FriendSimpleInfo) {
	if f.FriendRecommend == nil {
		f.FriendRecommend = make([]friend.FriendSimpleInfo, 0, friend.FriendCountPerReq)
	}
	f.FriendRecommend = f.FriendRecommend[:0]
	for _, item := range players {
		if !item.IsNil() && item.AcID != selfID {
			f.FriendRecommend = append(f.FriendRecommend, item)
		}
	}
}

func (f *Friend) NeedUpdateRecommendPlayer(nowT int64) bool {
	return nowT-f.LastUpdateRecommendPlayerTime > friend.UpdateRecommendPlayerInterval || len(f.FriendRecommend) <= 0
}

func (f *Friend) GetRecommendPlayer() []friend.FriendSimpleInfo {
	ret := make([]friend.FriendSimpleInfo, 0, friend.FriendCountPerReq)
	for _, item := range f.FriendRecommend {
		if !f.IsInList(item.AcID) {
			ret = append(ret, item)
		}
	}
	return ret
}

func (f *Friend) IsInList(acID string) bool {
	return f.IsInFriendList(acID) || f.IsInBlackList(acID)
}

func (f *Friend) GetFriendList() []friend.FriendSimpleInfo {
	return f.FriendList
}

func (f *Friend) SetFriendList(friendInfo []friend.FriendSimpleInfo) {
	for _, info := range friendInfo {
		index := f.getFriendIndex(info.AcID, f.FriendList)
		if index != -1 {
			f.FriendList[index] = info
		}
	}
}

func (f *Friend) SetBlackList(friendInfo []friend.FriendSimpleInfo) {
	for _, info := range friendInfo {
		index := f.getFriendIndex(info.AcID, f.BlackList)
		if index != -1 {
			f.BlackList[index] = info
		}
	}
}

func (f *Friend) GetBlackList() []friend.FriendSimpleInfo {
	return f.BlackList
}

func (f *Friend) GetBlackAcID() []string {
	ret := []string{}
	for _, info := range f.BlackList {
		ret = append(ret, info.AcID)
	}
	return ret
}

func (f *Friend) RefreshGiftTimes(nowT int64) {
	if !gamedata.IsSameDayFriendGift(nowT, f.RefreshGiftTime) {
		f.RefreshGiftTime = nowT
		f.GiftInfo = make([]string, 0)
		f.ReceiveGiftTimes = 0
	}
}

func (f *Friend) MarkGift2Friend(acID string) {
	f.GiftInfo = append(f.GiftInfo, acID)
}

func (f *Friend) CanGiveGift(acID string) bool {
	for _, item := range f.GiftInfo {
		if item == acID {
			return false
		}
	}
	return true
}

func (f *Friend) GetReceiveGiftTimes() int {
	return f.ReceiveGiftTimes
}

func (f *Friend) AddReceiveGiftTimes(value int) {
	f.ReceiveGiftTimes += value
}

func (f *Friend) CanReceiveGift() bool {
	return f.ReceiveGiftTimes < int(gamedata.GetFriendConfig().GetGetBaoziTimes())
}

func (f *Friend) GetFriendGiftInfo() []string {
	return append([]string{}, f.GiftInfo[:]...)
}

func (f *Friend) DBSave(cb redis.CmdBuffer, forceDirty bool) error {
	key := f.DBName()

	if forceDirty {
		f.dirtiesCheck = nil
	}
	err, newDirtyCheck, chged := driver.DumpToHashDBCmcBufferCheckDirty(
		cb, key, f, f.dirtiesCheck)
	if err != nil {
		return err
	}
	if !game.Cfg.IsRunModeProd() {
		if !reflect.DeepEqual(f.dirtiesCheck, newDirtyCheck) {
			logs.Trace("Save Friend %s %v", f.dbkey.Account.String(), chged)
		} else {
			logs.Trace("Save Friend clean %s", f.dbkey.Account.String())
		}
	}
	f.dirtiesCheck = newDirtyCheck
	return nil
}

func (f *Friend) DBLoad(logInfo bool) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	key := f.DBName()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, f, false, logInfo)

	// RESTORE_ERR_Profile_No_Data 表示玩家第一次登陆游戏，没有存档，这不视为Bug
	// 外面的逻辑需要根据此判断是否是第一次登陆游戏
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return nil
	}
	f.dirtiesCheck = driver.GenDirtyHash(f)
	return nil
}

func (p *Friend) DBName() string {
	return p.dbkey.String()
}
