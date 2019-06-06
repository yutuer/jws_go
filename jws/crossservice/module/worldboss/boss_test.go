package worldboss

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/crossservice/util/csdb"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"

	"github.com/stretchr/testify/assert"
)

// debugGetBossMod 每次调用时生成一个新的BossMod
func debugGetBossMod() *BossMod {
	return newBossMod(newResources(uint32(testGroupId), &WorldBoss{}))
}

func TestNewBossMod(t *testing.T) {
	bm := newBossMod(newResources(uint32(testGroupId), &WorldBoss{}))

	assert.NotNil(t, bm)
}

func BenchmarkGetCurrBossStatus(b *testing.B) {
	bm := debugGetBossMod()
	status := BossStatus{
		BossID:  "TestBoss001",
		SceneID: "TestScene001",
		Level:   22,
		HPMax:   32768,
		HPCurr:  16384,
		Seq:     11,
	}
	bm.currStatus = &status
	bm.makeOutStatus(time.Now())

	for i := 0; i < b.N; i++ {
		bm.getCurrBossStatus()
	}
}

func BenchmarkGetCommonStatus(b *testing.B) {
	bm := debugGetBossMod()
	bm.commonStatus.TotalDamage = uint64(666666)
	bm.makeOutStatus(time.Now())

	for i := 0; i < b.N; i++ {
		bm.getCommonStatus()
	}
}

func BenchmarkAttackBoss(b *testing.B) {
	bm := debugGetBossMod()
	bm.resetNewRoundBoss(time.Now())
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < b.N; i++ {
		bm.attackBoss(bm.currStatus.Level, uint64(rand.Int63n(1000000)))
	}
}

func TestGetCurrBossStatus(t *testing.T) {
	bm := newBossMod(newResources(uint32(testGroupId), &WorldBoss{}))

	status := BossStatus{
		BossID:  "TestBoss001",
		SceneID: "TestScene001",
		Level:   22,
		HPMax:   32768,
		HPCurr:  16384,
		Seq:     11,
	}

	bm.currStatus = &status
	bm.makeOutStatus(time.Now())

	bs := bm.getCurrBossStatus()
	assert.Equal(t, *bs, status)
}

func TestGetCommonStatus(t *testing.T) {
	bm := debugGetBossMod()
	bcs := bm.getCommonStatus()

	if bcs == nil || bcs.TotalDamage != 0 {
		t.Error("GetCommonStatus is inccrrect!")
	}
}

func TestAttackBoss(t *testing.T) {
	bm := debugGetBossMod()

	bm.currStatus.HPMax = uint64(10000)
	bm.currStatus.HPCurr = uint64(10000)

	damage := bm.attackBoss(bm.currStatus.Level, uint64(5000))
	assert.Equal(t, damage, uint64(5000))
	assert.Equal(t, bm.currStatus.HPCurr, uint64(5000))
	assert.Equal(t, bm.commonStatus.TotalDamage, uint64(5000))

	damage = bm.attackBoss(bm.currStatus.Level, uint64(7000))
	assert.Equal(t, damage, uint64(7000))
	assert.Equal(t, bm.commonStatus.TotalDamage, uint64(12000))

	damage = bm.attackBoss(uint32(testGroupId), uint64(12000))
	assert.Equal(t, damage, uint64(12000))
	assert.Equal(t, bm.commonStatus.TotalDamage, uint64(24000))
}

func TestResetNewRoundBoss(t *testing.T) {
	bm := debugGetBossMod()
	bm.resetNewRoundBoss(time.Now())

	assert.Equal(t, bm.currStatus.Seq, uint32(1))

	// 一周7天
	timeMon := time.Date(2017, 6, 12, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 7; i++ {
		bm.resetNewRoundBoss(timeMon)
		assert.Equal(t, bm.currStatus.Seq, uint32(1))
		timeMon.AddDate(0, 0, 1)
	}
}

func TestMakeNextBoss(t *testing.T) {
	bm := debugGetBossMod()
	bm.resetNewRoundBoss(time.Now())

	maxBossLvl := gamedata.GetWBConfig().GetMaxBossLevel()

	// 最后一次应该已经达到上限
	for i := uint32(0); i < maxBossLvl; i++ {
		currStat := bm.currStatus
		bm.currStatus = makeNextBoss(bm.currStatus)
		assert.True(t, bm.currStatus.HPMax >= currStat.HPMax)
		assert.True(t, bm.currStatus.Level >= currStat.Level)
		assert.True(t, bm.currStatus.Seq >= currStat.Seq)
	}

	bm.currStatus = makeNextBoss(bm.currStatus)
	// Level不越界，Seq正常，开始单独一次，结束单独一次
	assert.Equal(t, bm.currStatus.Level, maxBossLvl)
	assert.Equal(t, bm.currStatus.Seq, maxBossLvl+uint32(2))
}

func TestBossDamageToNextLevel(t *testing.T) {
	bm := debugGetBossMod()
	bm.resetNewRoundBoss(time.Now())

	currentLevel := bm.currStatus.Level
	currentHP := bm.currStatus.HPCurr

	// 目前表里配的都是整数，万一哪天调成不是整数就只能自己手动写数据了
	theHP := currentHP / uint64(2)

	// 正好打到0刷新下一个Boss
	inputDamages := []uint64{theHP, theHP, 666666}
	expectedTotalDamages := []uint64{theHP, currentHP, currentHP + 666666}
	expectedBossSeq := []uint32{1, 2, 2}

	for i, inputDamage := range inputDamages {
		actualDamage := bm.attackBoss(currentLevel, inputDamage)
		assert.Equal(t, actualDamage, inputDamage)
		assert.Equal(t, bm.commonStatus.TotalDamage, expectedTotalDamages[i])
		assert.Equal(t, bm.currStatus.Seq, expectedBossSeq[i])
	}
}

func TestBossDamageOverNextLevel(t *testing.T) {
	bm := debugGetBossMod()
	bm.resetNewRoundBoss(time.Now())

	currentLevel := bm.currStatus.Level
	currentHP := bm.currStatus.HPCurr

	// 超过剩下的HP
	inputDamages := []uint64{currentHP + 666666, 1000000}
	expectedTotalDamages := []uint64{currentHP + 666666, currentHP + 1666666}
	expectedBossSeq := []uint32{2, 2}

	for i, inputDamage := range inputDamages {
		actualDamage := bm.attackBoss(currentLevel, inputDamage)
		assert.Equal(t, actualDamage, inputDamage)
		assert.Equal(t, bm.commonStatus.TotalDamage, expectedTotalDamages[i])
		assert.Equal(t, bm.currStatus.Seq, expectedBossSeq[i])
	}

	// 刷新前伤害不影响刷新后Boss血量
	assert.Equal(t, bm.currStatus.HPMax, bm.currStatus.HPCurr)
}

// 下面是dbm和db相关的测试
func debugDelBossDB(groupId uint32, tag string) {
	keyName := fmt.Sprintf("worldboss:%d:boss:%s", groupId, tag)
	conn := csdb.GetDBConn(groupId)
	_, e := conn.Do("DEL", keyName)
	if e != nil {
		fmt.Println("debugDelBossDB failed: %s", e.Error())
	}
}

func (bm *BossMod) debugSetBossStatusDebugPreset1() {
	status := BossStatus{
		BossID:  "TestBossLv50",
		SceneID: "TestScene001",
		Level:   50,
		HPMax:   666666,
		HPCurr:  233333,
		Seq:     66,
	}

	bm.currStatus = &status
	bm.outCurrStatus = &status
	bm.commonStatus.TotalDamage = uint64(65536)
	bm.outCommonStatus.TotalDamage = uint64(65536)
}

func TestNewBoss(t *testing.T) {
	bdb := newBossDB(debugGetResource())

	assert.NotEmpty(t, bdb.group)
}

func TestSetBossStatus(t *testing.T) {
	bdb := newBossDB(debugGetResource())
	bm := newBossMod(debugGetResource())
	bm.debugSetBossStatusDebugPreset1()

	// 正常set
	err0 := bdb.setBossStatus(*bm.outCurrStatus, bm.res.ticker.roundStatus.BatchTag)
	assert.Nil(t, err0)

	// Redis 不可用
	unavailableDB := newBossDB(newResources(unavailableGroupId, &WorldBoss{}))
	err1 := unavailableDB.setBossStatus(*bm.outCurrStatus, bm.res.ticker.roundStatus.BatchTag)
	assert.NotNil(t, err1)
	assert.True(t, strings.Contains(err1.Error(), "refused"))

	// Redis 不存在
	notExistDB := newBossDB(newResources(notExistGroupId, &WorldBoss{}))
	err2 := notExistDB.setBossStatus(*bm.outCurrStatus, bm.res.ticker.roundStatus.BatchTag)
	assert.NotNil(t, err2)
	assert.Contains(t, err2.Error(), "GetDBConn")
}

func TestBossSaveBossToDB(t *testing.T) {
	bm := newBossMod(debugGetResource())
	bm.debugSetBossStatusDebugPreset1()
	err := bm.saveBossToDB()
	defer debugDelBossDB(testGroupId, bm.res.ticker.roundStatus.BatchTag)
	assert.Nil(t, err)

	// 正常save
	conn := csdb.GetDBConn(testGroupId)
	keyName := fmt.Sprintf("worldboss:%d:boss:%s", testGroupId, bm.res.ticker.roundStatus.BatchTag)
	code, e := conn.Do("EXISTS", keyName)
	assert.Nil(t, e)
	assert.Equal(t, int64(1), code)
}

func TestBossLoadBossFromDB(t *testing.T) {
	bm := newBossMod(debugGetResource())
	bm.debugSetBossStatusDebugPreset1()
	bm.saveBossToDB()
	defer debugDelBossDB(testGroupId, bm.res.ticker.roundStatus.BatchTag)

	// 清空并确定
	bm.resetNewRoundBoss(time.Now())
	bm.makeOutStatus(time.Now())
	assert.Equal(t, uint32(1), bm.currStatus.Seq)
	assert.Equal(t, uint32(1), bm.outCurrStatus.Seq)

	// 正常load
	err := bm.loadBossFromDB()
	assert.Nil(t, err)
	assert.Equal(t, uint32(50), bm.currStatus.Level)
	assert.Equal(t, uint64(233333), bm.currStatus.HPCurr)
	assert.Equal(t, uint32(66), bm.currStatus.Seq)
	assert.Equal(t, uint64(65536), bm.commonStatus.TotalDamage)

	// 如果BatchTag不存在，返回一个新Boss
	preBatchTag := bm.res.ticker.roundStatus.BatchTag
	bm.res.ticker.roundStatus.BatchTag = "Opps!"
	err2 := bm.loadBossFromDB()
	assert.Nil(t, err2)
	assert.Equal(t, uint32(1), bm.currStatus.Seq)
	bm.res.ticker.roundStatus.BatchTag = preBatchTag
}
