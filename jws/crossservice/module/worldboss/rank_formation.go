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

//FormationRankElemInfo ..
type FormationRankElemInfo struct {
	Pos       uint32
	Sid       uint32
	Acid      string
	Damage    uint64
	Team      []HeroInfoDetail
	BuffLevel uint32

	Name string
}

//FormationRankElem ..
type FormationRankElem struct {
	Pos       uint32           `json:"pos,omitempty"`
	Sid       uint32           `json:"sid,omitempty"`
	Acid      string           `json:"acid,omitempty"`
	Damage    uint64           `json:"damage,omitempty"`
	Team      []HeroInfoDetail `json:"team,omitempty"`
	BuffLevel uint32           `json:"bf_lv,omitempty"`
}

//Mod----

//FormationRankMod ..
type FormationRankMod struct {
	res *resources

	lockSnap sync.RWMutex
	snap     *[]FormationRankElemInfo

	lockList sync.RWMutex
	list     []*FormationRankElem
	mapList  map[string]*FormationRankElem
	count    uint32

	lockDirty sync.RWMutex
	dirty     map[string]uint64
	saveLast  time.Time
}

func newFormationRankMod(res *resources) *FormationRankMod {
	rm := &FormationRankMod{
		res: res,
	}
	rm.snap = &[]FormationRankElemInfo{}
	rm.list = make([]*FormationRankElem, 0, gamedata.GetMaxRankLimit())
	rm.mapList = make(map[string]*FormationRankElem)

	rm.dirty = make(map[string]uint64)

	rm.res.ticker.regTicker(29, 0, rm.doTick)
	rm.res.ticker.regTicker(181, 0, rm.trySaveRankToDB)

	return rm
}

func (rm *FormationRankMod) addPlayerFormation(sid uint32, acid string, damage uint64, team []HeroInfoDetail, buffLevel uint32) {
	rm.putinFormation(sid, acid, damage, team, buffLevel)
	rm.pushDirty(acid, damage)
}

func (rm *FormationRankMod) putinFormation(sid uint32, acid string, damage uint64, team []HeroInfoDetail, buffLevel uint32) uint64 {
	rm.lockList.Lock()
	defer rm.lockList.Unlock()

	if nil == rm.mapList[acid] {
		e := &FormationRankElem{Acid: acid, Damage: damage, Sid: sid, Team: team, BuffLevel: buffLevel}
		rm.mapList[acid] = e
		rm.list = append(rm.list, e)
		rm.count++
	} else {
		if rm.mapList[acid].Damage < damage {
			rm.mapList[acid].Damage = damage
			rm.mapList[acid].Team = team
			rm.mapList[acid].BuffLevel = buffLevel
		}
	}

	return rm.mapList[acid].Damage
}

func (rm *FormationRankMod) popDirty() map[string]uint64 {
	rm.lockDirty.Lock()
	defer rm.lockDirty.Unlock()
	out := rm.dirty
	rm.dirty = map[string]uint64{}

	return out
}

func (rm *FormationRankMod) pushDirty(acid string, damage uint64) {
	rm.lockDirty.Lock()
	defer rm.lockDirty.Unlock()
	rm.dirty[acid] = damage
}

func (rm *FormationRankMod) getTop() []FormationRankElemInfo {
	rm.lockSnap.RLock()
	defer rm.lockSnap.RUnlock()
	return *rm.snap
}

func (rm *FormationRankMod) makeSnap() {
	num := 100
	topList, count := rm.getSort(num)
	snapList := make([]FormationRankElemInfo, 0, count)
	for i := 0; i < count; i++ {
		elem := *rm.elemToElemInfoSimple(&topList[i])
		snapList = append(snapList, elem)
	}

	rm.lockSnap.Lock()
	defer rm.lockSnap.Unlock()
	rm.snap = &snapList
}

func (rm *FormationRankMod) getRankInfoByAcid(acid string) *FormationRankElemInfo {
	elem := rm.getRankByAcid(acid)
	return rm.elemToElemInfoSimple(elem)
}

func (rm *FormationRankMod) getRankByAcid(acid string) *FormationRankElem {
	rm.lockList.RLock()
	defer rm.lockList.RUnlock()
	e := rm.mapList[acid]
	if nil == e {
		return nil
	}
	team := make([]HeroInfoDetail, len(e.Team))
	copy(team, e.Team)
	re := &FormationRankElem{
		Pos:       e.Pos,
		Sid:       e.Sid,
		Acid:      e.Acid,
		Damage:    e.Damage,
		Team:      team,
		BuffLevel: e.BuffLevel,
	}
	return re
}

func (rm *FormationRankMod) getAllRank() []FormationRankElem {
	rm.lockList.RLock()
	defer rm.lockList.RUnlock()
	ret := make([]FormationRankElem, len(rm.list))
	for i, e := range rm.list {
		ret[i] = *e
	}
	return ret
}

func (rm *FormationRankMod) getSort(n int) ([]FormationRankElem, int) {
	count := int(rm.count)
	if 0 == count {
		return nil, 0
	}
	if count > n {
		count = n
	}
	ret := make([]FormationRankElem, 0, count)
	rm.lockList.RLock()
	defer rm.lockList.RUnlock()
	for i := 0; i < count; i++ {
		ret = append(ret, *rm.list[i])
	}
	return ret, count
}

func (rm *FormationRankMod) doSort() {
	rm.lockList.Lock()
	defer rm.lockList.Unlock()

	sort.Sort(sortListFormationRank(rm.list))
	for i, elem := range rm.list {
		elem.Pos = uint32(i + 1)
	}
}

func (rm *FormationRankMod) doTick(now time.Time) {
	rm.doSort()
	rm.makeSnap()
}

//--restore

//loadRankFromDB ..
func (rm *FormationRankMod) loadRankFromDB() error {
	members, err := rm.res.RankDB.getAllFormationRankMember(rm.res.ticker.roundStatus.BatchTag)
	if nil != err {
		return fmt.Errorf("FormationRankMod getAllRankMember failed, %v", err)
	}
	rm.lockList.Lock()
	defer rm.lockList.Unlock()

	rm.list = []*FormationRankElem{}
	rm.mapList = map[string]*FormationRankElem{}
	for acid, elem := range members {
		e := &FormationRankElem{
			Pos:    elem.Pos,
			Sid:    elem.Sid,
			Damage: elem.Damage,
			Acid:   elem.Acid,
		}
		e.Team = make([]HeroInfoDetail, len(elem.Team))
		copy(e.Team, elem.Team)
		rm.list = append(rm.list, e)
		rm.mapList[acid] = e
		rm.count++
	}
	logs.Info("[WorldBoss] FormationRankMod loadRankFromDB %+v", members)
	return nil
}

//trySaveRankToDB ..
func (rm *FormationRankMod) trySaveRankToDB(now time.Time) {
	dirty := rm.popDirty()
	if 0 != len(dirty) {
		dirtyRank := []FormationRankElem{}
		for acid := range dirty {
			dirtyRank = append(dirtyRank, *rm.getRankByAcid(acid))
		}
		if err := rm.res.RankDB.pushFormationRankMember(dirtyRank, rm.res.ticker.roundStatus.BatchTag); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] FormationRankMod trySaveRankToDB pushRankMember, %v", err))
		}
		logs.Trace("[WorldBoss] FormationRankMod trySaveBossToDB, save dirty %+v", dirty)
	}

	if now.Sub(rm.saveLast) > defaultForceSaveInterval {
		if err := rm.saveAllRank(); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] FormationRankMod trySaveRankToDB saveAllRank, %v", err))
		}
		rm.saveLast = now
		logs.Trace("[WorldBoss] FormationRankMod trySaveBossToDB, saveAllRank")
	}
}

//saveAllRank ..
func (rm *FormationRankMod) saveAllRank() error {
	all := rm.getAllRank()
	if 0 == len(all) {
		return nil
	}
	logs.Trace("[WorldBoss] FormationRankMod saveAllRank, all %+v", all)
	return rm.res.RankDB.pushFormationRankMember(all, rm.res.ticker.roundStatus.BatchTag)
}

//resetNewRound ..
func (rm *FormationRankMod) resetNewRound(now time.Time) {
	logs.Trace("[WorldBoss] FormationRankMod resetNewRound")

	rm.lockSnap.Lock()
	rm.snap = &[]FormationRankElemInfo{}
	rm.lockSnap.Unlock()

	rm.lockList.Lock()
	rm.list = make([]*FormationRankElem, 0)
	rm.mapList = make(map[string]*FormationRankElem)
	rm.count = 0
	rm.lockList.Unlock()
}

//--util
func (rm *FormationRankMod) elemToElemInfoSimple(e *FormationRankElem) *FormationRankElemInfo {
	if nil == e {
		return &FormationRankElemInfo{}
	}
	info := &FormationRankElemInfo{}
	info.Acid = e.Acid
	info.Pos = e.Pos
	info.Damage = e.Damage
	if playerinfo := rm.res.PlayerMod.getPlayerInfo(e.Acid); nil != playerinfo {
		info.Name = playerinfo.Name
	} else {
		info.Name = defaultName
	}
	info.Team = make([]HeroInfoDetail, len(e.Team))
	copy(info.Team, e.Team)
	info.BuffLevel = e.BuffLevel
	return info
}

//---sort

type sortListFormationRank []*FormationRankElem

func (s sortListFormationRank) Len() int {
	return len(s)
}
func (s sortListFormationRank) Less(i, j int) bool {
	return s[i].Damage > s[j].Damage
}
func (s sortListFormationRank) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
