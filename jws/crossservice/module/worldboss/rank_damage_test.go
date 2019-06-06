package worldboss

import (
	"math/rand"
	"testing"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func debugGetRankDamageMod() *RankDamageMod {
	rand.Seed(time.Now().UnixNano())
	return newResources(testGroupId, &WorldBoss{}).RankDamageMod
}

func DebugPushDamageData(rm *RankDamageMod, count uint32) {
	for i := uint32(0); i < count; i++ {
		rm.addPlayerDamage(uint32(i), uuid.NewV4().String(), int64(rand.Int63()))
	}
	//rm.doSort()
}

func debugGetRankDamageModWithData(dataAmount uint32) *RankDamageMod {
	rm := debugGetRankDamageMod()

	for i := uint32(0); i < dataAmount; i++ {
		rm.addPlayerDamage(uint32(i), uuid.NewV4().String(), int64(rand.Int63()))
	}

	return rm
}

// 测试添加数据性能
func BenchmarkAddData(b *testing.B) {
	ranker := debugGetRankDamageMod()

	for i := 0; i < b.N; i++ {
		ranker.addPlayerDamage(uint32(i), uuid.NewV4().String(), int64(rand.Int63()))
	}
}

// // 测试排序性能 （10000条数据）
func BenchmarkSortData(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ranker := debugGetRankDamageModWithData(uint32(10000))
		b.StartTimer() // 这里开始计时
		ranker.doSort()
	}
}

// 测试递增排序性能 （每次增加10条数据）
func BenchmarkSortIncreasedData(b *testing.B) {
	ranker := debugGetRankDamageMod()
	// 假设每次新增10个新数据
	newDataCount := uint32(10)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		DebugPushDamageData(ranker, newDataCount)
		b.StartTimer() // 这里开始计时
		ranker.doSort()
	}
}

func TestNewRankDamageMod(t *testing.T) {
	rm := newRankDamageMod(newResources(uint32(0), &WorldBoss{}))

	if rm.list == nil || rm.mapList == nil || rm.dirty == nil {
		t.Errorf("NewRankDamageMod error!")
	}
}

func TestGetRankByPos(t *testing.T) {
	ranker := debugGetRankDamageMod()
	acid := uuid.NewV4().String()

	ranker.addPlayerDamage(uint32(0), acid, 1)
	ranker.doSort()

	if ranker.getRankByPos(1).Acid != acid {
		t.Errorf("GetRankByPos error!")
	}

	vals := []uint32{100, 0}
	for _, val := range vals {
		r := ranker.getRankByPos(val)
		if r != nil {
			t.Errorf("GetRankByPos error!")
		}
	}
}

func TestGetMyRank(t *testing.T) {
	ranker := debugGetRankDamageMod()
	acid := uuid.NewV4().String()

	ranker.addPlayerDamage(uint32(0), acid, 1)
	ranker.doSort()

	dl := ranker.getMyRank(acid)
	if dl.Pos != 1 {
		t.Errorf("getmyRank error!")
	}

	DebugPushDamageData(ranker, uint32(10000))
	ranker.doSort()

	dl = ranker.getMyRank(acid)
	if dl.Pos != 10001 {
		// 如果能在int63里随机到1，我也服了
		t.Errorf("getmyRank error!")
	}
}

func TestGetRange(t *testing.T) {
	ranker := debugGetRankDamageModWithData(uint32(10001))

	dls := ranker.getRange(4, 10)
	if len(dls) != 7 {
		t.Errorf("getRange error!")
	}

	expected_pos := uint32(5)
	for _, dl := range dls {
		if dl.Pos != expected_pos {
			t.Errorf("expected: %d, actual: %d", expected_pos, dl.Pos)
		}
		expected_pos++
	}

	// 越界
	dls = ranker.getRange(10000, 20000)
	ranker.doSort()

	if len(dls) != 1 {
		t.Errorf("getRange error!")
	}

	dls = ranker.getRange(40000, 50000)
	if len(dls) != 0 {
		t.Errorf("getRange error!")
	}
}

func TestDataChanges(t *testing.T) {
	ranker := debugGetRankDamageMod()
	acid := uuid.NewV4().String()

	ranker.addPlayerDamage(uint32(0), acid, 1)
	ranker.doSort()

	dl := ranker.getMyRank(acid)
	if dl.Damage != 1 || dl.Pos != 1 {
		t.Errorf("addDamage error!")
	}

	// 伤害累积
	ranker.addPlayerDamage(uint32(1), acid, 100)
	ranker.doSort()

	dl = ranker.getMyRank(acid)
	if dl.Damage != 101 {
		t.Errorf("addDamage error!")
	}

	// 伤害累积到排名变更
	DebugPushDamageData(ranker, 1000)
	ranker.doSort()
	ranker.addPlayerDamage(uint32(3), acid, int64(ranker.getRankByPos(1).Damage))
	ranker.doSort()

	dl = ranker.getMyRank(acid)
	if dl.Pos != 1 {
		t.Errorf("Actual position: %d", dl.Pos)
	}
}

func TestDoSort(t *testing.T) {
	ranker := debugGetRankDamageModWithData(1000)
	ranker.doSort()

	dls := ranker.getAllRankDamage()

	// 测试是否正确按伤害排序
	prev_dmg := dls[0].Damage
	prev_pos := dls[0].Pos
	for _, dl := range dls {
		// (上次伤害 > 这次伤害) xor (上次排名 > 此次排名)
		if (prev_dmg < dl.Damage) != (prev_pos > dl.Pos) {
			t.Errorf("last: %d: %d, current:%d: %d", prev_pos, prev_dmg, dl.Pos, dl.Damage)
		} else {
			prev_dmg = dl.Damage
			prev_pos = dl.Pos
		}
	}
}
