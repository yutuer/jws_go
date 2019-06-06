package store

import (
	"math/rand"
	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/mn_selector"
)

//Blank ...
type Blank struct {
	StoreID   uint32                      `json:"-"`
	BlankID   uint32                      `json:"b,omitempty"`
	GoodIndex uint32                      `json:"g,omitempty"`
	Count     uint32                      `json:"c,omitempty"`
	EquipData *gamedata.BagItemData       `json:"e,omitempty"`
	MN        *mnSelector.MNSelectorState `json:"mn,omitempty"`
}

func newBlank(id uint32) *Blank {
	return &Blank{
		BlankID:   id,
		GoodIndex: 0,
		Count:     0,
		EquipData: nil,
		MN:        nil,
	}
}

func (b *Blank) refresh(storeID uint32, acid string, lv uint32, rd *rand.Rand) {
	cfg := gamedata.GetStoreBlankCfg(storeID, b.BlankID, lv)

	if true == b.checkSpecial(cfg, rd) {
		b.refreshGood(cfg.TreasureGroup, acid, rd, cfg)
	} else {
		b.refreshGood(cfg.RandNormal(rd), acid, rd, cfg)
	}
}

func (b *Blank) refreshGood(groupID uint32, acid string, rd *rand.Rand, cfg *gamedata.StoreBlankElem) {
	group := gamedata.GetStoreGoodGroup(groupID)
	if nil == group {
		logs.Warn("[Store] Refresh store [%d] blank [%d], but group [%d] is not exist in config", b.StoreID, b.BlankID, groupID)
		return
	}

	goodIndex := group.RandGood(rd)
	goodCfg := gamedata.GetStoreGoodCfg(goodIndex)

	b.GoodIndex = goodIndex
	b.Count = 0

	dp := gamedata.MakeItemData(acid, rd, goodCfg.GoodID)
	if dp != nil && false == dp.IsNil() {
		b.EquipData = dp
	}

	return
}

func (b *Blank) newMN(cfg *gamedata.StoreBlankElem, rd *rand.Rand) *mnSelector.MNSelectorState {
	mn := &mnSelector.MNSelectorState{}
	space := cfg.TreasureRefreshSpace + cfg.TreasureRefreshOffset - (rd.Uint32() % (2*cfg.TreasureRefreshOffset + 1))
	mn.Init(int64(cfg.TreasureRefreshNum), int64(space))

	return mn
}

func (b *Blank) checkSpecial(cfg *gamedata.StoreBlankElem, rd *rand.Rand) bool {
	if 0 == cfg.TreasureGroup {
		return false
	}
	if nil == b.MN || b.MN.IsNowNeedNewTurn() {
		b.MN = b.newMN(cfg, rd)
	}
	return b.MN.Selector(rd)
}

//Store ...
type Store struct {
	StoreID    uint32   `json:"id,omitempty"`
	Blanks     []*Blank `json:"b,omitempty"`
	BlankCount uint32   `json:"bc,omitempty"`

	LastRefreshTime    int64  `json:"lt,omitempty"`
	ManualRefreshCount uint32 `json:"mt,omitempty"`
}

func newStore(id uint32) *Store {
	return &Store{
		StoreID:            id,
		Blanks:             []*Blank{},
		BlankCount:         0,
		LastRefreshTime:    0,
		ManualRefreshCount: 0,
	}
}

func (s *Store) afterLogin() {
	for _, b := range s.Blanks {
		b.StoreID = s.StoreID
	}
}

//GetBlank ..
func (s *Store) GetBlank(id uint32) *Blank {
	for _, b := range s.Blanks {
		if b.BlankID == id {
			return b
		}
	}
	return nil
}

//addBlank ..
func (s *Store) addBlank(id uint32) *Blank {
	b := newBlank(id)
	b.StoreID = s.StoreID

	l := len(s.Blanks)
	i := sort.Search(l, func(i int) bool { return s.Blanks[i].BlankID > id })
	if i < l {
		oldList := s.Blanks[:]
		s.Blanks = make([]*Blank, len(oldList)+1)
		copy(s.Blanks[:i], oldList[:i])
		s.Blanks[i] = b
		copy(s.Blanks[i+1:], oldList[i:])
	} else {
		s.Blanks = append(s.Blanks, b)
	}

	return b
}

func (s *Store) refresh(acid string, lv uint32, rd *rand.Rand) {
	storeCfg := gamedata.GetStoreCfg(s.StoreID)
	if nil == storeCfg {
		logs.Warn("[Store] Refresh store [%d], but it is not in store config", s.StoreID)
		return
	}
	blankIDs := storeCfg.GetBlankIDs()
	for _, bid := range blankIDs {
		blank := s.GetBlank(bid)
		if nil == blank {
			blank = s.addBlank(bid)
		}

		blank.refresh(s.StoreID, acid, lv, rd)
	}

	//打乱它们
	if len(blankIDs) == len(s.Blanks) {
		logs.Debug("[Store] Refresh store [%d], Blanks before perm %+v", s.StoreID, s.Blanks)
		permIndex := rd.Perm(len(blankIDs))
		for i := 0; i < len(blankIDs); i++ {
			logs.Debug("[Store] Refresh store [%d], BlankID change %d -> %d", s.StoreID, s.Blanks[i].BlankID, blankIDs[permIndex[i]])
			s.Blanks[i].BlankID = blankIDs[permIndex[i]]
		}
		for i := 0; i < len(s.Blanks); i++ {
			for j := i + 1; j < len(s.Blanks); j++ {
				if s.Blanks[i].BlankID > s.Blanks[j].BlankID {
					s.Blanks[i], s.Blanks[j] = s.Blanks[j], s.Blanks[i]
				}
			}
		}
		logs.Debug("[Store] Refresh store [%d], perm %+v", s.StoreID, permIndex)
		logs.Debug("[Store] Refresh store [%d], Blanks after perm %+v", s.StoreID, s.Blanks)
	} else {
		logs.Warn("[Store] Refresh store [%d], but its blankIDs length(%d) is not equal blanks length(%d)", s.StoreID, len(blankIDs), len(s.Blanks))
	}
}

//Update ...
func (s *Store) update(acid string, now int64, lv uint32, rd *rand.Rand) (chg bool) {
	chg = false
	last := s.LastRefreshTime

	//检查是否清空当前的计数
	if false == gamedata.CheckStoreSameDay(last, now) {
		logs.Debug("[Store] Store update do change day")
		s.ManualRefreshCount = 0
		s.LastRefreshTime = now
		chg = true
	}

	//检查是否要自动刷新
	if true == gamedata.CheckStoreNeedAutoRefresh(s.StoreID, last, now) {
		s.refresh(acid, lv, rd)
		logs.Debug("[Store] Store update do refresh, after: %+v", s)
		// logs.Warn("[Store] Store [%s] [%d] update do refresh", acid, s.StoreID)
		s.LastRefreshTime = now
		chg = true
	}

	return
}

//ManualRefresh ..
func (s *Store) ManualRefresh(acid string, now int64, lv uint32, rd *rand.Rand) {
	s.ManualRefreshCount++
	s.LastRefreshTime = now
	s.refresh(acid, lv, rd)
}

//Market ..
type Market struct {
	Stores []*Store          `json:"stores,omitempty"`
	InMap  map[uint32]*Store `json:"-"`
}

func newMarket() *Market {
	m := &Market{
		Stores: []*Store{},
		InMap:  map[uint32]*Store{},
	}
	return m
}

func (m *Market) afterLogin() {
	m.InMap = map[uint32]*Store{}
	for _, store := range m.Stores {
		m.InMap[store.StoreID] = store
		store.afterLogin()
	}
}

func (m *Market) getStore(id uint32) *Store {
	return m.InMap[id]
}

func (m *Market) addStore(id uint32) *Store {
	store, exist := m.InMap[id]
	if true == exist {
		return store
	}

	store = newStore(id)
	l := len(m.Stores)
	i := sort.Search(l, func(j int) bool { return m.Stores[j].StoreID > store.StoreID })
	if i < l {
		oldList := m.Stores[:]
		m.Stores = make([]*Store, len(oldList)+1)
		copy(m.Stores[:i], oldList[:i])
		m.Stores[i] = store
		copy(m.Stores[i+1:], oldList[i:])
	} else {
		m.Stores = append(m.Stores, store)
	}
	m.InMap[store.StoreID] = store

	logs.Debug("[Store] Market addStore stores %+v", m.Stores)
	return store
}

func (m *Market) update(acid string, now int64, lv uint32, rd *rand.Rand) map[uint32]bool {
	chg := map[uint32]bool{}

	storeIDs := gamedata.GetStoreIDs()
	logs.Debug("[Store] Market update GetStoreIDs %v", storeIDs)
	for _, id := range storeIDs {
		store := m.getStore(id)
		if nil == store {
			store = m.addStore(id)
			logs.Debug("[Store] Market update addStore %+v", store)
		}
		if true == store.update(acid, now, lv, rd) {
			chg[id] = true
		}
	}

	return chg
}
