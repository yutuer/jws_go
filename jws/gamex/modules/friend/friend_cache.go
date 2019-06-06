package friend

import (
	"fmt"
	"sync"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type FriendSimpleInfo struct {
	AcID        string `codec:"acid" json:"acid"`
	Name        string `codec:"name" json:"name"`
	GS          int    `codec:"gs" json:"gs"`
	Lv          int    `codec:"lv" json:"lv"`
	VIPLv       int    `codec:"vip_lv" json:"vip_lv"`
	Avatar      int    `codec:"avatar" json:"avatar"`
	LastActTime int64  `codec:"last_act_time" json:"last_act_time"`
}

func (fsi *FriendSimpleInfo) IsNil() bool {
	return fsi.AcID == "" || fsi.Name == ""
}

type friendCache struct {
	cache     [maxGSLevel]map[string]FriendSimpleInfo
	cacheMute sync.RWMutex
}

func (fc *friendCache) Init() {
	for i := 0; i < maxGSLevel; i++ {
		fc.cache[i] = make(map[string]FriendSimpleInfo, 1000)
	}
}

func (fc *friendCache) putInCache(friends []FriendSimpleInfo) {
	fc.cacheMute.Lock()
	defer fc.cacheMute.Unlock()
	nowT := time.Now().Unix()
	for _, f := range friends {
		if nowT-f.LastActTime < maxCacheTime || f.LastActTime == 0 {
			level := fc.selectCacheLevel(f.GS)
			fc.cache[level][f.AcID] = f
			//fc.cacheGauge[level].Update(int64(len(fc.cache[level])))
		}
	}
}

func (fc *friendCache) selectCacheLevel(gs int) int {
	return 0
}

func (fc *friendCache) findFriendInfo(id string) FriendSimpleInfo {
	fc.cacheMute.RLock()
	defer fc.cacheMute.RUnlock()
	for _, c := range fc.cache {
		v, ok := c[id]
		if ok {
			return v
		}
	}
	return FriendSimpleInfo{}
}

// 尝试从内存中读取信息，返回读到的info和未读到的id
func (fc *friendCache) queryFriendInfoAndRet(ids []string) ([]string, []FriendSimpleInfo) {
	fc.cacheMute.RLock()
	defer fc.cacheMute.RUnlock()
	retInfo := make([]FriendSimpleInfo, 0, len(ids))
	leaveID := make([]string, 0, 5)
	for _, id := range ids {
		info := FriendSimpleInfo{}
		for _, c := range fc.cache {
			v, ok := c[id]
			if ok {
				info = v
			}
		}
		if info.IsNil() {
			leaveID = append(leaveID, id)
		} else {
			retInfo = append(retInfo, info)
		}
	}
	return leaveID, retInfo
}

type ReceiveGiftInfo struct {
	AcID      string
	ReceiveAc []string
}

func (rbzi *ReceiveGiftInfo) RefreshReceiveInfo(removeItem []string) {
	markIndex := make(map[int]struct{}, 0)
	for i, item := range rbzi.ReceiveAc {
		for _, item2 := range removeItem {
			if item == item2 {
				markIndex[i] = struct{}{}
			}
		}
	}
	newReceive := []string{}
	for i, item := range rbzi.ReceiveAc {
		if _, ok := markIndex[i]; !ok {
			newReceive = append(newReceive, item)
		}
	}
	rbzi.ReceiveAc = newReceive
	maxC := int(gamedata.GetFriendConfig().GetStorageLimit())
	if len(rbzi.ReceiveAc) > int(maxC) {
		rbzi.ReceiveAc = rbzi.ReceiveAc[len(rbzi.ReceiveAc)-maxC:]
	}
	logs.Debug("refresh receive info: %v", rbzi.ReceiveAc)
}

func (rbzi *ReceiveGiftInfo) GetReceiveInfo2Client() []string {
	maxC := int(gamedata.GetFriendConfig().GetStorageLimit())
	if len(rbzi.ReceiveAc) > maxC {
		return append([]string{}, rbzi.ReceiveAc[len(rbzi.ReceiveAc)-maxC:]...)
	}
	return append([]string{}, rbzi.ReceiveAc[:]...)
}

func (rbzi *ReceiveGiftInfo) receiveGift(acID string) bool {
	var index int = -1
	for i, item := range rbzi.ReceiveAc {
		if item == acID {
			index = i
			break
		}
	}
	if index == -1 {
		logs.Error("no gift info for acid: %v tgt id: %v", rbzi.AcID, acID)
		return false
	}
	rbzi.ReceiveAc = append(rbzi.ReceiveAc[:index], rbzi.ReceiveAc[index+1:]...)
	return true
}

func (rbzi *ReceiveGiftInfo) putGift(acID string) {
	rbzi.ReceiveAc = append(rbzi.ReceiveAc, acID)
}

type GiftInfo struct {
	Info map[string]*ReceiveGiftInfo
}

func (bzi *GiftInfo) Init() {
	bzi.Info = make(map[string]*ReceiveGiftInfo, 1000)
}

func (bzi *GiftInfo) ReceiveGift(acID string, tgtAcID string) bool {
	if v, ok := bzi.Info[acID]; ok {
		return v.receiveGift(tgtAcID)
	} else {
		logs.Error("no gift info for acid: %v", acID)
		return false
	}
}

func (bzi *GiftInfo) PutGift(acID string, tgtAcID string) {
	if v, ok := bzi.Info[acID]; ok {
		v.putGift(tgtAcID)
	} else {
		info := ReceiveGiftInfo{
			AcID:      acID,
			ReceiveAc: []string{tgtAcID},
		}
		bzi.Info[acID] = &info
	}
}

func (bzi *GiftInfo) GetGiftInfo(acID string, refresh bool, removeItem []string) []string {
	if v, ok := bzi.Info[acID]; ok {
		if refresh {
			v.RefreshReceiveInfo(removeItem)
		}
		return v.GetReceiveInfo2Client()

	} else {
		logs.Info("no gift info for acid: %v:", acID)
	}
	return nil
}

func (bzi *GiftInfo) loadGiftInfo(sid uint) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(),
		TableFriend(sid), bzi, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return nil
}

func (bzi *GiftInfo) saveGiftInfo(sid uint) error {
	cb := redis.NewCmdBuffer()
	if err := driver.DumpToHashDBCmcBuffer(cb, TableFriend(sid), bzi); err != nil {
		return fmt.Errorf("DumpToHashDBCmcBuffer err %v", err)
	}

	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		return fmt.Errorf("cant get redis conn")
	}

	if _, err := modules.DoCmdBufferWrapper(
		db_counter_key, db, cb, true); err != nil {
		return fmt.Errorf("DoCmdBuffer error %s", err.Error())
	}
	return nil
}

type GiftCmd struct {
	Typ   int
	Param interface{}
	Ret   chan GiftRet
}

type GiftRet struct {
	Ret interface{}
}
