package worldboss

import (
	"sync"
	"time"

	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//BossStatus ..
type BossStatus struct {
	BossID  string
	SceneID string
	Level   uint32
	HPMax   uint64
	HPCurr  uint64
	Seq     uint32
}

//BossCommonStatus ..
type BossCommonStatus struct {
	TotalDamage uint64
}

//AttackInfo ..
type AttackInfo struct {
	Damage uint64
	Level  uint32
}

//BossMod ..
type BossMod struct {
	res *resources

	lockerStatus sync.RWMutex
	currStatus   *BossStatus
	commonStatus *BossCommonStatus

	outCurrStatus   *BossStatus
	outCommonStatus *BossCommonStatus

	saveDirty bool
	saveLast  time.Time
}

func newBossMod(res *resources) *BossMod {
	bm := &BossMod{
		res: res,
	}
	bm.currStatus = &BossStatus{}
	bm.commonStatus = &BossCommonStatus{}

	bm.outCurrStatus = &BossStatus{}
	bm.outCommonStatus = &BossCommonStatus{}

	bm.res.ticker.regTicker(180, 0, bm.trySaveBossToDB)
	bm.res.ticker.regTicker(1, 0, bm.makeOutStatus)
	return bm
}

//getCurrBossStatus ..
func (bm *BossMod) getCurrBossStatus() *BossStatus {
	return bm.outCurrStatus
}

//getCommonStatus ..
func (bm *BossMod) getCommonStatus() *BossCommonStatus {
	return bm.outCommonStatus
}

//makeOutStatus ..
func (bm *BossMod) makeOutStatus(now time.Time) {
	bm.lockerStatus.Lock()
	defer bm.lockerStatus.Unlock()
	bm.outCurrStatus = &BossStatus{
		BossID:  bm.currStatus.BossID,
		SceneID: bm.currStatus.SceneID,
		Level:   bm.currStatus.Level,
		HPMax:   bm.currStatus.HPMax,
		HPCurr:  bm.currStatus.HPCurr,
		Seq:     bm.currStatus.Seq,
	}
	bm.outCommonStatus = &BossCommonStatus{
		TotalDamage: bm.commonStatus.TotalDamage,
	}
}

//attackBoss ..
func (bm *BossMod) attackBoss(lv uint32, damage uint64) uint64 {
	bm.lockerStatus.Lock()
	defer bm.lockerStatus.Unlock()

	bm.commonStatus.TotalDamage += damage

	if bm.currStatus.Level != lv {
		return damage
	}
	if bm.currStatus.HPCurr > damage {
		bm.currStatus.HPCurr -= damage
	} else {
		bm.currStatus.HPCurr = 0
		bm.currStatus = makeNextBoss(bm.currStatus)
		bm.res.ticker.roundStatus.LastNewBossTime = time.Now().Unix()
		bm.res.ticker.roundStatus.dirty = true
	}

	bm.saveDirty = true
	return damage
}

//loadBossFromDB ..
func (bm *BossMod) loadBossFromDB() error {
	now := time.Now()
	common, err := bm.res.BossDB.getBossCommonStatus(bm.res.ticker.roundStatus.BatchTag)
	if nil != err {
		return fmt.Errorf("getBossCommonStatus failed, %v", err)
	}
	if nil != common {
		bm.commonStatus.TotalDamage = common.TotalDamage
	}
	logs.Trace("[WorldBoss] BossMod loadBossFromDB, common %+v", bm.commonStatus)

	status, err := bm.res.BossDB.getBossStatus(bm.res.ticker.roundStatus.BatchTag)
	if nil != err {
		return fmt.Errorf("getBossStatus failed, %v", err)
	}

	if nil == status {
		bm.currStatus = makeNewBoss(now)
		logs.Trace("[WorldBoss] BossMod loadBossFromDB new, boss %+v", bm.currStatus)
		return nil
	}

	boss, scene := gamedata.GetTodayWBBoss(now)
	maxHP := gamedata.GetWBBossHP(status.Level)
	bm.currStatus.BossID = boss
	bm.currStatus.SceneID = scene
	bm.currStatus.Level = status.Level
	bm.currStatus.HPMax = uint64(maxHP)
	bm.currStatus.HPCurr = status.HPCurr
	bm.currStatus.Seq = status.Seq
	logs.Trace("[WorldBoss] BossMod loadBossFromDB, boss %+v", bm.currStatus)

	return nil
}

//trySaveBossToDB ..
func (bm *BossMod) trySaveBossToDB(now time.Time) {
	if true == bm.saveDirty || now.Sub(bm.saveLast) > defaultForceSaveInterval {
		if err := bm.saveBossToDB(); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] BossMod trySaveBossToDB, %v", err))
		}
		bm.saveDirty = false
		bm.saveLast = now
	}
}

func (bm *BossMod) saveBossToDB() error {
	currStatus := bm.getCurrBossStatus()
	if err := bm.res.BossDB.setBossStatus(*currStatus, bm.res.ticker.roundStatus.BatchTag); nil != err {
		return fmt.Errorf("saveBossToDB setBossStatus failed, %v", err)
	}
	logs.Trace("[WorldBoss] BossMod trySaveBossToDB, setBossStatus %+v", currStatus)
	commonStatus := bm.getCommonStatus()
	if err := bm.res.BossDB.setBossCommonStatus(*commonStatus, bm.res.ticker.roundStatus.BatchTag); nil != err {
		return fmt.Errorf("saveBossToDB setBossCommonStatus failed, %v", err)
	}
	logs.Trace("[WorldBoss] BossMod trySaveBossToDB, setBossCommonStatus %+v", commonStatus)
	return nil
}

//resetNewRoundBoss ..
func (bm *BossMod) resetNewRoundBoss(now time.Time) {
	bm.lockerStatus.Lock()
	defer bm.lockerStatus.Unlock()
	bm.currStatus = makeNewBoss(now)
	bm.commonStatus.TotalDamage = 0
	logs.Trace("[WorldBoss] BossMod resetNewRoundBoss, boss %+v", bm.currStatus)
}

//makeNewBoss ..
func makeNewBoss(now time.Time) *BossStatus {
	boss, scene := gamedata.GetTodayWBBoss(now)
	lv, hp := gamedata.GetNextWBBoss(0)
	return &BossStatus{
		BossID:  boss,
		SceneID: scene,
		Level:   lv,
		HPMax:   uint64(hp),
		HPCurr:  uint64(hp),
		Seq:     1,
	}
}

//makeNextBoss ..
func makeNextBoss(curr *BossStatus) *BossStatus {
	nlv, nhp := gamedata.GetNextWBBoss(curr.Level)
	return &BossStatus{
		BossID:  curr.BossID,
		SceneID: curr.SceneID,
		Level:   nlv,
		HPMax:   uint64(nhp),
		HPCurr:  uint64(nhp),
		Seq:     curr.Seq + 1,
	}
}
