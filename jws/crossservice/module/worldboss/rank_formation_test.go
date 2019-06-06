package worldboss

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func DebugGetFormationRankMod() *FormationRankMod {
	rand.Seed(time.Now().UnixNano())
	return newResources(uint32(0), &WorldBoss{}).FormationRankMod
}

func DebugGetHeroInfoDetail() (string, []HeroInfoDetail) {
	acid := uuid.NewV4().String()
	heroInfoDetail := make([]HeroInfoDetail, 16)

	return acid, heroInfoDetail
}

func DebugPushFormationRankData(frm *FormationRankMod, count int) {
	for i := 0; i < count; i++ {
		acid, heroInfoDetail := DebugGetHeroInfoDetail()
		frm.addPlayerFormation(uint32(i), acid, rand.Uint64(), heroInfoDetail, 1)
	}
	//frm.doSort()
}

// 排序10000固定数据
func BenchmarkDoSortWithFixedAmount(b *testing.B) {
	frm := DebugGetFormationRankMod()

	for i := 0; i < b.N; i++ {
		frm.resetNewRound(time.Now())
		DebugPushFormationRankData(frm, 10000)
		frm.doSort()
	}
}

// 递增10数据
func BenchmarkDoSortWithIncreasing(b *testing.B) {
	frm := DebugGetFormationRankMod()

	for i := 0; i < b.N; i++ {
		DebugPushFormationRankData(frm, 10)
		frm.doSort()
	}
}

func TestNewFormationRankMod(t *testing.T) {
	frm := newFormationRankMod(newResources(uint32(0), &WorldBoss{}))

	if frm == nil {
		t.Error("newFormationRankMod failed!")
	}
}

func TestPutInFormation(t *testing.T) {
	frm := DebugGetFormationRankMod()
	acid, heroInfoDetail := DebugGetHeroInfoDetail()

	// 新增
	damage := frm.putinFormation(uint32(0), acid, uint64(65536), heroInfoDetail, 1)

	if frm.count != 1 || frm.mapList[acid] == nil || len(frm.list) != 1 {
		t.Error("putInFormation failed: putin new")
	}

	if damage != uint64(65536) || frm.mapList[acid].Damage != damage {
		t.Error("putInFormation failed: return value")
	}

	// 更新，伤害低于之前
	damage = frm.putinFormation(uint32(1), acid, uint64(65535), heroInfoDetail, 0)

	if frm.count != 1 || frm.mapList[acid] == nil || len(frm.list) != 1 {
		t.Error("putInFormation failed: update")
	}

	if damage != uint64(65536) || frm.mapList[acid].Damage != damage {
		t.Error("putInFormation failed: update")
	}

	// 更新，伤害高于之前
	damage = frm.putinFormation(uint32(2), acid, uint64(131072), heroInfoDetail, 2)

	if frm.count != 1 || frm.mapList[acid] == nil || len(frm.list) != 1 {
		t.Error("putInFormation failed: update")
	}

	if damage != uint64(131072) || frm.mapList[acid].Damage != damage {
		t.Error("putInFormation failed: update")
	}
}

func TestPushDirty(t *testing.T) {
	frm := DebugGetFormationRankMod()
	acid := uuid.NewV4().String()
	damage := uint64(8888)

	frm.pushDirty(acid, damage)
	if frm.dirty[acid] != damage {
		t.Error("pushDirty failed!")
	}
}

func TestPopDirty(t *testing.T) {
	frm := DebugGetFormationRankMod()
	count := 10
	srcDirties := make(map[string]uint64, count)

	for i := 0; i < count; i++ {
		acid := uuid.NewV4().String()
		damage := rand.Uint64()
		srcDirties[acid] = damage
		frm.pushDirty(acid, damage)
	}

	dirties := frm.popDirty()

	if !reflect.DeepEqual(srcDirties, dirties) {
		t.Error("popDirty failed: content")
	}
}

func TestRankFormationDoSort(t *testing.T) {
	frm := DebugGetFormationRankMod()
	DebugPushFormationRankData(frm, 1000)

	frm.doSort()

	// 排序正确性验证
	curPos := frm.list[0].Pos
	curDam := frm.list[0].Damage

	for _, e := range frm.list {
		if e.Pos < curPos || e.Damage > curDam {
			t.Error("doSort failed!")
		}
		curPos = e.Pos
		curDam = e.Damage
	}
}

func TestMakeSnap(t *testing.T) {
	frm := DebugGetFormationRankMod()

	// 小于100条数据
	DebugPushFormationRankData(frm, 65)

	frm.doSort()
	frm.makeSnap()

	if len(*frm.snap) != 65 {
		t.Error("makeSnap failed: Less than 100")
	}

	// 多于100条数据
	DebugPushFormationRankData(frm, 233)

	frm.doSort()
	frm.makeSnap()

	if len(*frm.snap) != 100 {
		t.Error("makeSnap failed: More than 100")
	}
}

func TestGetRankByAcid(t *testing.T) {
	frm := DebugGetFormationRankMod()
	acid := uuid.NewV4().String()
	DebugPushFormationRankData(frm, 1000)

	// 无数据
	if frm.getRankByAcid(acid) != nil {
		t.Error("getRankByAcid failed: no data")
	}

	// 有数据
	frm.list[0].Damage = math.MaxUint64
	acid = frm.list[0].Acid

	frm.doSort()

	if frm.getRankByAcid(acid).Pos != 1 {
		t.Error("getRankByAcid failed: exists")
	}
}

func TestGetSort(t *testing.T) {
	frm := DebugGetFormationRankMod()
	DebugPushFormationRankData(frm, 100)

	// 不超过总数
	if frmList, count := frm.getSort(20); len(frmList) != 20 || count != 20 {
		t.Error("getSort failed: less than total")
	}

	// 等于总数
	if frmList, count := frm.getSort(100); len(frmList) != 100 || count != 100 {
		t.Error("getSort failed: equal to total")
	}

	// 超过总数
	if frmList, count := frm.getSort(200); len(frmList) != 100 || count != 100 {
		t.Error("getSort failed: more than total")
	}
}

func TestGetAllRank(t *testing.T) {
	frm := DebugGetFormationRankMod()
	DebugPushFormationRankData(frm, 100)
	frm.doSort()

	allRankList := frm.getAllRank()
	for i, e := range frm.list {
		if !reflect.DeepEqual(*e, allRankList[i]) {
			t.Error("getAllRank failed!")
		}
	}
}

func TestResetNewRound(t *testing.T) {
	frm := DebugGetFormationRankMod()
	DebugPushFormationRankData(frm, 100)

	frm.resetNewRound(time.Now())

	if len(frm.list) != 0 || len(frm.mapList) != 0 ||
		len(*frm.snap) != 0 || frm.count != 0 {
		t.Error("resetNewRound failed!")
	}
}

func TestElemToElemInfoSimple(t *testing.T) {
	frm := DebugGetFormationRankMod()
	DebugPushFormationRankData(frm, 1)

	frm.makeSnap()

	if !reflect.DeepEqual(*frm.elemToElemInfoSimple(frm.list[0]), (*frm.snap)[0]) {
		t.Error("elemToElemInfoSimple failed!")
	}
}
