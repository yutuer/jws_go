package gamedata

import (
	"math/rand"
	"sort"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
// StoreMaxSize = 10
// BlankMaxSize = 40

// StorePoolSpecMax = 1

// poolWeightMax           = 10000
// storePoolRanderInitSize = 50
// VIPStoreIdx             = 6
)

//GoodData 对应配置表STOREGROUP中的一行
type GoodData struct {
	Index uint32

	Weight uint32

	GoodID    string
	GoodCount uint32

	PriceTyp   string
	PriceCount uint32

	IsTreasure uint32

	VIPLimit int
	Original int
	Discount int

	Give CostData
	Cost CostData
}

func newGoodData(data *ProtobufGen.STOREGROUP) *GoodData {
	gd := &GoodData{
		Index:      data.GetStoreGoodsID(),
		Weight:     data.GetWeight(),
		GoodID:     data.GetGoodsID(),
		GoodCount:  data.GetGoodsCount(),
		PriceTyp:   data.GetCoinItemID(),
		PriceCount: data.GetCoinItemCount(),
		IsTreasure: data.GetIsTreasure(),
		VIPLimit:   int(data.GetVIPLimit()),
		Original:   int(data.GetOriginalCount()),
		Discount:   int(data.GetDiscount()),
		Give:       CostData{},
		Cost:       CostData{},
	}

	gd.Give.AddItem(data.GetGoodsID(), data.GetGoodsCount())
	gd.Cost.AddItem(data.GetCoinItemID(), data.GetCoinItemCount())

	return gd
}

//GoodSet ..
type GoodSet struct {
	set map[uint32]*GoodData
}

func newGoodSet() *GoodSet {
	return &GoodSet{
		set: map[uint32]*GoodData{},
	}
}

func (s *GoodSet) addGood(good *GoodData) {
	s.set[good.Index] = good
}

func (s *GoodSet) checkGood(goodIndex uint32) *GoodData {
	return s.set[goodIndex]
}

//GoodPool 对应配置表 STOREGROUP 中的一个 GroupID
type GoodPool struct {
	PoolID uint32
	Goods  []uint32
	Rander *util.RandUintSetV2
}

//newGoodPool ..
func newGoodPool(id uint32) *GoodPool {
	g := &GoodPool{}
	g.PoolID = id
	g.Goods = make([]uint32, 0)
	g.Rander = util.NewRandUintSetV2()
	return g
}

func (g *GoodPool) addGood(good *GoodData) {
	g.Goods = append(g.Goods, good.Index)
	g.Rander.Add(good.Index, good.Weight)
}

//RandGood ..
func (g *GoodPool) RandGood(rd *rand.Rand) uint32 {
	return g.Rander.Rand(rd)
}

//GoodPoolSet ..
type GoodPoolSet struct {
	set map[uint32]*GoodPool
}

func newGoodPoolSet() *GoodPoolSet {
	return &GoodPoolSet{
		set: map[uint32]*GoodPool{},
	}
}

func (p *GoodPoolSet) getPool(poolID uint32) *GoodPool {
	pool, exist := p.set[poolID]
	if false == exist {
		pool = newGoodPool(poolID)
		p.set[poolID] = pool
	}

	return pool
}

func (p *GoodPoolSet) checkPool(poolID uint32) *GoodPool {
	return p.set[poolID]
}

//StoreBlankElem 对应配置表STOREBLANK中的一行
type StoreBlankElem struct {
	LvMin uint32
	LvMax uint32

	normalGroups []uint32
	normalRander *util.RandUintSetV2

	TreasureGroup         uint32
	TreasureRefreshNum    uint32
	TreasureRefreshSpace  uint32
	TreasureRefreshOffset uint32
}

func newStoreBlankElem(data *ProtobufGen.STOREBLANK) *StoreBlankElem {
	s := &StoreBlankElem{
		LvMin: data.GetLevelMin(),
		LvMax: data.GetLevelMax(),

		normalGroups: []uint32{},
		normalRander: util.NewRandUintSetV2(),

		TreasureGroup:         data.GetTreasureGroup(),
		TreasureRefreshNum:    data.GetRefreshNum(),
		TreasureRefreshSpace:  data.GetRefreshSpace(),
		TreasureRefreshOffset: data.GetRefreshOffset(),
	}

	for _, group := range data.GetNormalGroup_Table() {
		s.normalGroups = append(s.normalGroups, group.GetNormalGroup())
		s.normalRander.Add(group.GetNormalGroup(), group.GetGroupWeight())
	}
	return s
}

//RandNormal ..
func (s *StoreBlankElem) RandNormal(rd *rand.Rand) uint32 {
	return s.normalRander.Rand(rd)
}

//StoreBlank 对应配置表 STOREBLANK 中一个 StoreID 的一个 BlankID
type StoreBlank struct {
	BlankID uint32

	List []*StoreBlankElem
}

func newStoreBlank(id uint32) *StoreBlank {
	return &StoreBlank{
		BlankID: id,
		List:    []*StoreBlankElem{},
	}
}

func (b *StoreBlank) addElem(elem *StoreBlankElem) {
	l := len(b.List)
	i := sort.Search(l, func(i int) bool { return b.List[i].LvMin > elem.LvMin })
	if i < l {
		oldList := b.List[:]
		b.List = make([]*StoreBlankElem, len(oldList)+1)
		copy(b.List[:i], oldList[:i])
		b.List[i] = elem
		copy(b.List[i+1:], oldList[i:])
	} else {
		b.List = append(b.List, elem)
	}
}

//StoreData 对应配置表 STOREBLANK 中一个 StoreID
type StoreData struct {
	StoreID uint32

	Blanks map[uint32]*StoreBlank

	AutoRefreshTime   []int64
	ManualRefreshCost []*StoreRefreshCost
}

func newStoreData(id uint32) *StoreData {
	return &StoreData{
		StoreID: id,
		Blanks:  map[uint32]*StoreBlank{},
	}
}

func (s *StoreData) getBlank(id uint32) *StoreBlank {
	blank, exist := s.Blanks[id]
	if false == exist {
		blank = newStoreBlank(id)
		s.Blanks[id] = blank
	}
	return blank
}

func (s *StoreData) addRefresh(cost *StoreRefreshCost) {
	l := len(s.ManualRefreshCost)
	i := sort.Search(l, func(i int) bool { return s.ManualRefreshCost[i].RefreshNum > cost.RefreshNum })
	if i < l {
		oldList := s.ManualRefreshCost[:]
		s.ManualRefreshCost = make([]*StoreRefreshCost, len(oldList)+1)
		copy(s.ManualRefreshCost[:i], oldList[:i])
		s.ManualRefreshCost[i] = cost
		copy(s.ManualRefreshCost[i+1:], oldList[i:])
	} else {
		s.ManualRefreshCost = append(s.ManualRefreshCost, cost)
	}
}

func (s *StoreData) addAutoRefresh(t int64) {
	l := len(s.AutoRefreshTime)
	i := sort.Search(l, func(i int) bool { return s.AutoRefreshTime[i] > t })
	if i < l {
		oldList := s.AutoRefreshTime[:]
		s.AutoRefreshTime = make([]int64, len(oldList)+1)
		copy(s.AutoRefreshTime[:i], oldList[:i])
		s.AutoRefreshTime[i] = t
		copy(s.AutoRefreshTime[i+1:], oldList[i:])
	} else {
		s.AutoRefreshTime = append(s.AutoRefreshTime, t)
	}
}

func (s *StoreData) checkBlank(blankID uint32) *StoreBlank {
	return s.Blanks[blankID]
}

//GetBlankIDs ..
func (s *StoreData) GetBlankIDs() []uint32 {
	list := []uint32{}

	for _, b := range s.Blanks {
		list = append(list, b.BlankID)
	}
	return list
}

func (s *StoreData) getRefreshCost(count uint32) (string, uint32) {
	l := len(s.ManualRefreshCost)
	i := sort.Search(l, func(i int) bool { return s.ManualRefreshCost[i].RefreshNum > count })
	if i >= l {
		return s.ManualRefreshCost[l-1].CostType, s.ManualRefreshCost[l-1].CostNum
	}
	return s.ManualRefreshCost[i].CostType, s.ManualRefreshCost[i].CostNum
}

//StoreDataSet ...
type StoreDataSet struct {
	Sets map[uint32]*StoreData
}

func newStoreDataSet() *StoreDataSet {
	s := &StoreDataSet{
		Sets: map[uint32]*StoreData{},
	}
	return s
}

func (s *StoreDataSet) getStore(id uint32) *StoreData {
	data, exist := s.Sets[id]
	if false == exist {
		data = newStoreData(id)
		s.Sets[id] = data
	}
	return data
}

func (s *StoreDataSet) checkStore(id uint32) *StoreData {
	return s.Sets[id]
}

//StoreRefreshCost ..
type StoreRefreshCost struct {
	RefreshNum uint32
	CostType   string
	CostNum    uint32
}

func newStoreRefreshCost(data *ProtobufGen.REFRESHPRICE) *StoreRefreshCost {
	cost := &StoreRefreshCost{
		RefreshNum: data.GetRefreshNum(),
		CostType:   data.GetRefreshCoin(),
		CostNum:    data.GetRefreshPrice(),
	}
	return cost
}

//StoreConfigPack ..
type StoreConfigPack struct {
	stores    *StoreDataSet
	goods     *GoodSet
	goodpools *GoodPoolSet
}

var gStoreConfig *StoreConfigPack

//getStoreConfigInstance ..
func getStoreConfigInstance() *StoreConfigPack {
	if nil == gStoreConfig {
		gStoreConfig = &StoreConfigPack{
			stores:    newStoreDataSet(),
			goods:     newGoodSet(),
			goodpools: newGoodPoolSet(),
		}
	}

	return gStoreConfig
}

func loadStoreDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	blankDataList := &ProtobufGen.STOREBLANK_ARRAY{}
	err = proto.Unmarshal(buffer, blankDataList)
	errcheck(err)

	stores := getStoreConfigInstance().stores
	for _, blankData := range blankDataList.GetItems() {
		store := stores.getStore(blankData.GetStoreID())
		blank := store.getBlank(blankData.GetBlankID())
		blank.addElem(newStoreBlankElem(blankData))
	}
}

func loadStorePoolDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.STOREGROUP_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	pools := getStoreConfigInstance().goodpools
	goods := getStoreConfigInstance().goods
	for _, elem := range dataList.GetItems() {
		good := newGoodData(elem)
		goods.addGood(good)

		pool := pools.getPool(elem.GetGroupID())
		pool.addGood(good)
	}
}

func loadStoreRefreshCostDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.REFRESHPRICE_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	stores := getStoreConfigInstance().stores
	for _, costData := range dataList.GetItems() {
		store := stores.getStore(costData.GetStoreID())
		store.addRefresh(newStoreRefreshCost(costData))
	}
}

func loadStoreAutoRefreshDataConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.REFRESHTIME_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	stores := getStoreConfigInstance().stores
	for _, autoData := range dataList.GetItems() {
		store := stores.getStore(autoData.GetStoreID())
		for _, t := range autoData.GetStoreRule_Template() {
			store.addAutoRefresh(util.DailyRealTimeFromString(t.GetRefreshTime()))
		}
		logs.Debug("[Store] loadStoreAutoRefreshDataConfig store %v", store)
	}

	logs.Debug("[Store] loadStoreAutoRefreshDataConfig stores %v", stores)
}

//CheckStoreNeedAutoRefresh ..
func CheckStoreNeedAutoRefresh(storeID uint32, last, now int64) bool {
	todayBegin := util.DailyBeginUnix(now)
	daily := now - todayBegin
	lastDaily := last - todayBegin

	store := getStoreConfigInstance().stores.getStore(storeID)

	logs.Debug("[Store] CheckStoreNeedAutoRefresh now %d last %d daily %d lastDaily %d, auto:%+v", now, last, daily, lastDaily, store.AutoRefreshTime)

	for _, auto := range store.AutoRefreshTime {
		if lastDaily < auto && daily >= auto {
			return true
		}
	}

	return false
}

//GetStoreNextAutoRefresh ..
func GetStoreNextAutoRefresh(storeID uint32, now int64) int64 {
	todayBegin := util.DailyBeginUnix(now)
	daily := now - todayBegin

	store := getStoreConfigInstance().stores.getStore(storeID)

	for _, auto := range store.AutoRefreshTime {
		if auto > daily {
			return todayBegin + auto
		}
	}

	if 0 == len(store.AutoRefreshTime) {
		return todayBegin + util.DaySec + 5*util.HourSec
	}
	return todayBegin + store.AutoRefreshTime[0] + util.DaySec
}

//CheckStoreSameDay ..
func CheckStoreSameDay(last, now int64) bool {
	return util.DailyBeginUnix(last+DayOffset) == util.DailyBeginUnix(now+DayOffset)
}

//GetStoreIDs ..
func GetStoreIDs() []uint32 {
	list := []uint32{}
	for _, s := range getStoreConfigInstance().stores.Sets {
		list = append(list, s.StoreID)
	}
	return list
}

//GetStoreCfg ..
func GetStoreCfg(storeID uint32) *StoreData {
	return getStoreConfigInstance().stores.checkStore(storeID)
}

//GetStoreBlankCfg ..
func GetStoreBlankCfg(storeID uint32, blankID uint32, lv uint32) *StoreBlankElem {
	store := getStoreConfigInstance().stores.checkStore(storeID)
	if nil == store {
		return nil
	}

	blank := store.checkBlank(blankID)
	if nil == blank {
		return nil
	}

	for _, elem := range blank.List {
		if elem.LvMin <= lv && elem.LvMax >= lv {
			return elem
		}
	}

	return nil
}

//GetStoreGoodGroup ..
func GetStoreGoodGroup(groupID uint32) *GoodPool {
	return getStoreConfigInstance().goodpools.checkPool(groupID)
}

//GetStoreGoodCfg ..
func GetStoreGoodCfg(goodIndex uint32) *GoodData {
	return getStoreConfigInstance().goods.checkGood(goodIndex)
}

//GetStoreManualRefreshCost ..
func GetStoreManualRefreshCost(storeID uint32, count uint32) (string, uint32) {
	store := getStoreConfigInstance().stores.checkStore(storeID)
	if nil == store {
		return "", 0
	}

	return store.getRefreshCost(count)
}

//GetStoreManualRefreshLimit ...
func GetStoreManualRefreshLimit(storeID uint32, vip uint32) uint32 {
	vipCfg := GetVIPCfg(int(vip))
	if nil == vipCfg {
		return 0
	}

	return vipCfg.StoreRefreshLimitTable[storeID]
}
