package worldboss

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//define----

//DamageRankElemInfo ..
type DamageRankElemInfo struct {
	Pos    uint32
	Acid   string
	Name   string
	Damage uint64
	Team   []HeroInfoDetail
}

//DamageRankElem ..
type DamageRankElem struct {
	Acid   string `json:"acid,omitempty"`
	Sid    uint32 `json:"sid,omitempty"`
	Damage uint64 `json:"damage,omitempty"`
	Pos    uint32 `json:"pos,omitempty"`
}

func (e *DamageRankElem) clone() *DamageRankElem {
	return &DamageRankElem{
		Acid:   e.Acid,
		Sid:    e.Sid,
		Damage: e.Damage,
		Pos:    e.Pos,
	}
}

//Mod----

//..
const (
	BufferSnapLength = 10
)

//DamageRankElemInfoList ..
type DamageRankElemInfoList struct {
	Quick []DamageRankElemInfo
	List  []DamageRankElemInfo
}

//RankDamageMod ..
type RankDamageMod struct {
	res *resources

	// lockSnap sync.RWMutex
	snap      [BufferSnapLength]*DamageRankElemInfoList
	snapIndex int

	lockList sync.RWMutex
	list     []*DamageRankElem
	mapList  map[string]*DamageRankElem
	count    uint32

	outList    *[]*DamageRankElem
	outMapList *map[string]*DamageRankElem

	lockDirty sync.RWMutex
	dirty     map[string]uint64
	saveLast  time.Time
}

func newRankDamageMod(res *resources) *RankDamageMod {
	rm := &RankDamageMod{
		res: res,
	}
	for i := 0; i < BufferSnapLength; i++ {
		rm.snap[i] = &DamageRankElemInfoList{
			List:  make([]DamageRankElemInfo, 0, defaultRankQuickLength),
			Quick: make([]DamageRankElemInfo, 0, defaultRankQuickLength),
		}
	}

	rm.list = make([]*DamageRankElem, 0, gamedata.GetMaxRankLimit())
	rm.mapList = make(map[string]*DamageRankElem)
	rm.makeOutList()

	rm.dirty = make(map[string]uint64)

	rm.res.ticker.regTicker(1, 0, rm.doTick)
	rm.res.ticker.regTicker(180, 0, rm.trySaveRankToDB)

	return rm
}

func (rm *RankDamageMod) getTop() []DamageRankElemInfo {
	list := rm.snap[rm.snapIndex]
	return list.List
}

func (rm *RankDamageMod) getQuick() []DamageRankElemInfo {
	list := rm.snap[rm.snapIndex]
	return list.Quick
}

func (rm *RankDamageMod) getMyRank(acid string) *DamageRankElemInfo {
	me := rm.getRankByAcid(acid)
	return rm.elemToElemInfoSimple(me)
}

func (rm *RankDamageMod) getMyRankWithContext(acid string) (*DamageRankElemInfo, []DamageRankElemInfo) {
	me := rm.getRankByAcid(acid)
	if nil == me {
		bottom := rm.getRankBottom()
		cl := []DamageRankElemInfo{}
		if nil != bottom {
			cl = append(cl, *rm.elemToElemInfoSimple(bottom))
		}
		me := &DamageRankElem{Acid: acid, Pos: 0, Damage: 0}
		return rm.elemToElemInfoSimple(me), cl
	}
	if 0 == me.Pos {
		return rm.elemToElemInfoSimple(me), []DamageRankElemInfo{}
	}
	c1 := rm.getRankByPos(me.Pos - 1)
	return rm.elemToElemInfoSimple(me), []DamageRankElemInfo{*rm.elemToElemInfoSimple(c1)}
}

func (rm *RankDamageMod) addPlayerDamage(sid uint32, acid string, damage int64) {
	sum := rm.putinDamage(sid, acid, damage)
	rm.setDirty(acid, sum)
}

func (rm *RankDamageMod) putinDamage(sid uint32, acid string, damage int64) uint64 {
	rm.lockList.Lock()
	defer rm.lockList.Unlock()

	if nil == rm.mapList[acid] {
		if damage < 0 {
			damage = 0
		}
		e := &DamageRankElem{Acid: acid, Damage: uint64(damage), Sid: sid}
		rm.mapList[acid] = e
		rm.list = append(rm.list, e)
		rm.count++
	} else {
		rm.mapList[acid].Damage = uint64(int64(rm.mapList[acid].Damage) + damage)
	}

	return rm.mapList[acid].Damage
}

func (rm *RankDamageMod) setDirty(acid string, damage uint64) {
	rm.lockDirty.Lock()
	defer rm.lockDirty.Unlock()
	rm.dirty[acid] = damage
}

func (rm *RankDamageMod) popDirty() map[string]uint64 {
	rm.lockDirty.Lock()
	defer rm.lockDirty.Unlock()
	out := rm.dirty
	rm.dirty = map[string]uint64{}

	return out
}

func (rm *RankDamageMod) makeSnap() {
	num := 100
	topList, count := rm.getSort(num)
	snapList := make([]DamageRankElemInfo, 0, count)
	for i := 0; i < count; i++ {
		elem := *rm.elemToElemInfoSimple(&topList[i])
		snapList = append(snapList, elem)
	}
	quickLength := defaultRankQuickLength
	if quickLength > count {
		quickLength = count
	}
	quickList := make([]DamageRankElemInfo, quickLength)
	copy(quickList, snapList[:quickLength])

	if rm.snapIndex+1 >= BufferSnapLength {
		rm.snapIndex = 0
	}
	rm.snap[rm.snapIndex] = &DamageRankElemInfoList{List: snapList, Quick: quickList}
}

//makeOutList ..
func (rm *RankDamageMod) makeOutList() {
	outList := make([]*DamageRankElem, len(rm.list))
	outMapList := make(map[string]*DamageRankElem, len(rm.mapList))

	for i := 0; i < len(rm.list); i++ {
		ne := rm.list[i].clone()
		outList[i] = ne
		outMapList[ne.Acid] = ne
	}

	rm.outList = &outList
	rm.outMapList = &outMapList
}

func (rm *RankDamageMod) getSort(n int) ([]DamageRankElem, int) {
	// rm.lockList.RLock()
	// defer rm.lockList.RUnlock()
	outlist := *(rm.outList)
	count := len(outlist)
	if count > n {
		count = n
	}
	ret := make([]DamageRankElem, 0, count)
	for i := 0; i < count; i++ {
		ret = append(ret, *outlist[i])
	}
	return ret, count
}

func (rm *RankDamageMod) getRankByAcid(acid string) *DamageRankElem {
	// rm.lockList.RLock()
	// defer rm.lockList.RUnlock()
	m := *rm.outMapList
	e := m[acid]
	if nil == e {
		return nil
	}
	re := &DamageRankElem{
		Pos:    e.Pos,
		Acid:   e.Acid,
		Damage: e.Damage,
	}
	return re
}

func (rm *RankDamageMod) getRankByPos(p uint32) *DamageRankElem {
	// rm.lockList.RLock()
	// defer rm.lockList.RUnlock()
	outlist := *rm.outList
	if p > uint32(len(outlist)) {
		return nil
	}
	if 0 == p {
		return nil
	}
	e := outlist[p-1]
	re := &DamageRankElem{
		Pos:    e.Pos,
		Acid:   e.Acid,
		Damage: e.Damage,
	}
	return re
}
func (rm *RankDamageMod) getRankBottom() *DamageRankElem {
	// rm.lockList.RLock()
	// defer rm.lockList.RUnlock()
	outlist := *rm.outList
	if 0 == len(outlist) {
		return nil
	}
	e := outlist[len(outlist)-1]
	re := &DamageRankElem{
		Pos:    e.Pos,
		Acid:   e.Acid,
		Damage: e.Damage,
	}
	return re
}

func (rm *RankDamageMod) getRange(begin, end uint32) []DamageRankElem {
	rm.doSort()
	// rm.lockList.RLock()
	// defer rm.lockList.RUnlock()

	outlist := *rm.outList
	count := uint32(len(outlist))
	if begin >= count {
		return []DamageRankElem{}
	}
	e := end
	if e >= count {
		e = count - 1
	}
	if e < begin {
		return []DamageRankElem{}
	}
	ret := make([]DamageRankElem, 0, e-begin+1)
	for i := begin; i <= e; i++ {
		ret = append(ret, *outlist[i])
	}

	return ret
}

func (rm *RankDamageMod) doSort() {
	rm.lockList.Lock()
	defer rm.lockList.Unlock()

	sort.Sort(sortListDamageRank(rm.list))
	for i, elem := range rm.list {
		elem.Pos = uint32(i + 1)
	}
	rm.makeOutList()
}

func (rm *RankDamageMod) doTick(now time.Time) {
	rm.doSort()
	rm.makeSnap()
}

func (rm *RankDamageMod) getAllRankDamage() []DamageRankElem {
	rm.lockList.RLock()
	defer rm.lockList.RUnlock()
	ret := make([]DamageRankElem, len(rm.list))
	for i, e := range rm.list {
		ret[i] = *e
	}
	return ret
}

//loadRankFromDB ..
func (rm *RankDamageMod) loadRankFromDB() error {
	members, err := rm.res.RankDB.getAllRankMember(rm.res.ticker.roundStatus.BatchTag)
	if nil != err {
		return fmt.Errorf("RankDamageMod getAllRankMember failed, %v", err)
	}
	rm.lockList.Lock()
	defer rm.lockList.Unlock()

	rm.list = []*DamageRankElem{}
	rm.mapList = map[string]*DamageRankElem{}
	for acid, elem := range members {
		e := &DamageRankElem{
			Acid:   elem.Acid,
			Sid:    elem.Sid,
			Damage: elem.Damage,
			Pos:    elem.Pos,
		}
		rm.list = append(rm.list, e)
		rm.mapList[acid] = e
		rm.count++
	}
	logs.Info("[WorldBoss] RankDamageMod loadRankFromDB %+v", members)
	return nil
}

//trySaveRankToDB ..
func (rm *RankDamageMod) trySaveRankToDB(now time.Time) {
	dirty := rm.popDirty()
	if 0 != len(dirty) {
		dirtyRank := []DamageRankElem{}
		for acid := range dirty {
			e := rm.getRankByAcid(acid)
			if nil == e {
				continue
			}
			dirtyRank = append(dirtyRank, *e)
		}
		if err := rm.res.RankDB.pushRankMember(dirtyRank, rm.res.ticker.roundStatus.BatchTag); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] RankDamageMod trySaveRankToDB pushRankMember, %v", err))
		}
		logs.Trace("[WorldBoss] RankDamageMod trySaveBossToDB, save dirty %+v", dirty)
	}

	if now.Sub(rm.saveLast) > defaultForceSaveInterval {
		if err := rm.saveAllRank(); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] RankDamageMod trySaveRankToDB saveAllRank, %v", err))
		}
		rm.saveLast = now
		logs.Trace("[WorldBoss] RankDamageMod trySaveBossToDB, saveAllRank")
	}
}

//saveAllRank ..
func (rm *RankDamageMod) saveAllRank() error {
	all := rm.getAllRankDamage()
	if 0 == len(all) {
		return nil
	}
	logs.Trace("[WorldBoss] RankDamageMod saveAllRank, all %+v", all)
	return rm.res.RankDB.pushRankMember(all, rm.res.ticker.roundStatus.BatchTag)
}

//resetNewRound ..
func (rm *RankDamageMod) resetNewRound(now time.Time) {
	logs.Trace("[WorldBoss] RankDamageMod resetNewRound")

	for i := 0; i < BufferSnapLength; i++ {
		rm.snap[i] = &DamageRankElemInfoList{
			List: make([]DamageRankElemInfo, 0, defaultRankQuickLength),
		}
	}

	rm.lockList.Lock()
	rm.list = make([]*DamageRankElem, 0)
	rm.mapList = make(map[string]*DamageRankElem)
	rm.count = 0
	rm.lockList.Unlock()
}

//--util
func (rm *RankDamageMod) elemToElemInfoSimple(e *DamageRankElem) *DamageRankElemInfo {
	if nil == e {
		return &DamageRankElemInfo{}
	}
	info := &DamageRankElemInfo{}
	info.Acid = e.Acid
	info.Pos = e.Pos
	info.Damage = e.Damage
	if playerinfo := rm.res.PlayerMod.getPlayerInfo(e.Acid); nil != playerinfo {
		info.Name = playerinfo.Name
	} else {
		info.Name = defaultName
	}
	info.Team = []HeroInfoDetail{}
	if teaminfo := rm.res.PlayerMod.getTeamInfo(e.Acid); nil != teaminfo {
		info.Team = make([]HeroInfoDetail, len(teaminfo.Team))
		copy(info.Team, teaminfo.Team)
	}
	return info
}

//---sort

type sortListDamageRank []*DamageRankElem

func (s sortListDamageRank) Len() int {
	return len(s)
}
func (s sortListDamageRank) Less(i, j int) bool {
	return s[i].Damage > s[j].Damage
}
func (s sortListDamageRank) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
