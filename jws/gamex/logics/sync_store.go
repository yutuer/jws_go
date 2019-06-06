package logics

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/store"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type blankToClient struct {
	BlankID   uint32 `codec:"bid"`
	GoodIndex uint32 `codec:"gi"`
	Count     uint32 `codec:"c"`
	EquipData string `codec:"ed"`
}

func buildBlankToClient(blank *store.Blank) *blankToClient {
	client := &blankToClient{}

	client.BlankID = blank.BlankID
	client.GoodIndex = blank.GoodIndex
	client.Count = blank.Count
	if nil != blank.EquipData {
		client.EquipData = blank.EquipData.ToDataStr()
	} else {
		client.EquipData = ""
	}

	return client
}

type storeToClient struct {
	StoreID            uint32   `codec:"sid"`
	Blanks             [][]byte `codec:"bs"`
	LastRefreshTime    int64    `codec:"lt"`
	NextRefreshTime    int64    `codec:"nt"`
	ManualRefreshCount uint32   `codec:"mrc"`
}

func buildStoreToClient(store *store.Store, now int64) *storeToClient {
	client := &storeToClient{}

	client.StoreID = store.StoreID
	client.LastRefreshTime = store.LastRefreshTime
	client.ManualRefreshCount = store.ManualRefreshCount
	client.NextRefreshTime = gamedata.GetStoreNextAutoRefresh(store.StoreID, now)
	logs.Debug("[Store] SyncResp buildStoreToClient GetStoreNextAutoRefresh %d, now %d", client.NextRefreshTime, now)

	client.Blanks = make([][]byte, 0, len(store.Blanks))
	for _, b := range store.Blanks {
		client.Blanks = append(client.Blanks, encode(buildBlankToClient(b)))
	}

	return client
}

func buildStoreToClientWithBlankIDs(store *store.Store, ids []uint32, now int64) *storeToClient {
	client := &storeToClient{}

	client.StoreID = store.StoreID
	client.LastRefreshTime = store.LastRefreshTime
	client.ManualRefreshCount = store.ManualRefreshCount
	client.NextRefreshTime = gamedata.GetStoreNextAutoRefresh(store.StoreID, now)
	logs.Debug("[Store] SyncResp buildStoreToClient GetStoreNextAutoRefresh %d, now %d", client.NextRefreshTime, now)

	client.Blanks = make([][]byte, 0, len(store.Blanks))
	for _, bid := range ids {
		b := store.GetBlank(bid)
		client.Blanks = append(client.Blanks, encode(buildBlankToClient(b)))
	}

	return client
}

type syncStoreElem struct {
	storeID    uint32
	syncAll    bool
	syncBlanks []uint32
}

func (s *SyncResp) addSyncStore(storeID uint32) {
	if nil == s.syncStoreDelta {
		s.syncStoreDelta = map[uint32]syncStoreElem{}
	}
	s.syncStoreDelta[storeID] = syncStoreElem{
		storeID: storeID,
		syncAll: true,
	}
	s.syncStoreDeltaNeed = true
}
func (s *SyncResp) addSyncStoreBlank(storeID uint32, blankID uint32) {
	if nil == s.syncStoreDelta {
		s.syncStoreDelta = map[uint32]syncStoreElem{}
	}
	store, ok := s.syncStoreDelta[storeID]
	if false == ok {
		store = syncStoreElem{
			storeID:    storeID,
			syncAll:    false,
			syncBlanks: []uint32{},
		}
	}
	store.syncBlanks = append(store.syncBlanks, blankID)
	s.syncStoreDelta[storeID] = store
	s.syncStoreDeltaNeed = true
}

func (s *SyncResp) mkStoreAllInfo(p *Account) {
	// 玩家商店信息，注意不止一个商店 []byte 是 store
	// SyncStore [][]byte `codec:"stores_"`
	now := time.Now().Unix()
	if s.store_all_sync {
		now := p.Profile.GetProfileNowTime()

		corp := p.Profile.GetCorp()
		lv := corp.GetLvlInfo()
		p.StoreProfile.Update(p.AccountID.String(), now, lv, p.GetRand())

		stores := p.StoreProfile.GetStores()
		logs.Debug("[Store] SyncResp mkStoreAllInfo stores %+v", stores)
		s.SyncStore = make([][]byte, 0, len(stores))
		bytesLen := 0
		for _, store := range stores {
			logs.Debug("[Store] SyncResp mkStoreAllInfo store %+v", store)
			bs := encode(buildStoreToClient(store, now))
			s.SyncStore = append(s.SyncStore, bs)
			bytesLen += len(bs)
		}

		logs.Debug("[SyncStore] length of SyncStore bytes is [%d]", bytesLen)
	} else if true == s.syncStoreDeltaNeed && nil != s.syncStoreDelta {
		stores := p.StoreProfile.GetStores()
		s.SyncStore = make([][]byte, 0, len(stores))
		bytesLen := 0
		s.SyncStore = make([][]byte, 0)
		for _, elem := range s.syncStoreDelta {
			if true == elem.syncAll {
				store := p.StoreProfile.GetStore(elem.storeID)
				bs := encode(buildStoreToClient(store, now))
				s.SyncStore = append(s.SyncStore, bs)
				bytesLen += len(bs)
			} else if nil != elem.syncBlanks && 0 != len(elem.syncBlanks) {
				store := p.StoreProfile.GetStore(elem.storeID)
				bs := encode(buildStoreToClientWithBlankIDs(store, elem.syncBlanks, now))
				s.SyncStore = append(s.SyncStore, bs)
				bytesLen += len(bs)
			}
		}
		logs.Debug("[SyncStore] length of SyncStore bytes is [%d] (delta)", bytesLen)
	}
}
