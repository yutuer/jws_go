package gates_enemy

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/gates_enemy/cmd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *GatesEnemyActivity) onEnterAct(accountID string,
	info helper.AccountSimpleInfo, okChann chan<- GatesEnemyRet) {
	fd, ok := r.memberData[accountID]
	if !ok {
		logs.Error("GatesEnemyActivity onEnterAct NoData %s", accountID)
		okChann <- GenGatesEnemyRet(RESWarnNoAct)
		return
	}
	fd.isInAct = true
	r.needPush = true
	okChann <- GenGatesEnemyRet(RESSuccess)
}

func (r *GatesEnemyActivity) onLeaveAct(accountID string, okChann chan<- GatesEnemyRet) {
	fd, ok := r.memberData[accountID]
	if !ok {
		logs.Error("GatesEnemyActivity onLeaveAct NoData %s", accountID)
		okChann <- GenGatesEnemyRet(RESWarnNoAct)
		return
	}
	fd.isInAct = false
	r.needPush = true
	okChann <- GenGatesEnemyRet(RESSuccess)
}

func (r *GatesEnemyActivity) onFightBegin(accountID string,
	info helper.AccountSimpleInfo,
	enemyTyp, enemyIDx int,
	okChann chan<- GatesEnemyRet) {
	if enemyTyp < 0 || enemyTyp >= len(r.enemyInfo) {
		okChann <- GenGatesEnemyRet(RESWarnEnemyIDErr)
		return
	}

	if enemyIDx < 0 || enemyIDx >= len(r.enemyInfo[enemyTyp]) {
		okChann <- GenGatesEnemyRet(RESWarnEnemyIDErr)
		return
	}

	fData, ok := r.memberData[accountID]
	if !ok || fData == nil {
		okChann <- GenGatesEnemyRet(RESWarnCanNotInAct)
		return
	}

	if r.enemyInfo[enemyTyp][enemyIDx] != 1 {
		okChann <- GenGatesEnemyRet(RESWarnEnemyHasFighting)
		return
	}

	fData.isFighting = true
	fData.currEnemyIdx = enemyIDx
	fData.currEnemyType = enemyTyp
	fData.startFightTime = time.Now().Unix()

	r.enemyInfo[enemyTyp][enemyIDx] = 0
	r.needPush = true

	for i := 0; i < len(r.members); i++ {
		if r.members[i].AccountID == accountID {
			r.members[i] = info
		}
	}

	okChann <- GenGatesEnemyRet(RESSuccess)
}

func (r *GatesEnemyActivity) onFightEnd(accountID string,
	info helper.AccountSimpleInfo,
	enemyTyp, enemyIDx int, isSuccess bool,
	okChann chan<- GatesEnemyRet) {

	if enemyTyp < 0 || enemyTyp >= len(r.enemyInfo) {
		okChann <- GenGatesEnemyRet(RESWarnEnemyIDErr)
		return
	}

	if enemyIDx < 0 || enemyIDx >= len(r.enemyInfo[enemyTyp]) {
		okChann <- GenGatesEnemyRet(RESWarnEnemyIDErr)
		return
	}

	fData, ok := r.memberData[accountID]
	if !ok || fData == nil {
		okChann <- GenGatesEnemyRet(RESWarnCanNotInAct)
		return
	}

	if !fData.isFighting {
		okChann <- GenGatesEnemyRet(RESWarnCanNotInAct)
		return
	}

	if enemyTyp != fData.currEnemyType || enemyIDx != fData.currEnemyIdx {
		okChann <- GenGatesEnemyRet(RESWarnEnemyIDErr)
		return
	}

	// TODO by FanYang 服务器暂时不处理时间

	fData.isFighting = false
	fData.currEnemyIdx = 0
	fData.currEnemyType = 0
	fData.startFightTime = 0

	// 击杀会获得杀戮值 以此激活Boss
	if isSuccess {
		data := gamedata.GetAllGEEnemyGroupCfg()
		logs.Trace("onFightEnd Give %v", data)
		if enemyTyp < len(data) {
			enemyID := data[enemyTyp].GetEGLevelID()
			lootData := gamedata.GetGEEnemyLootCfg(enemyID)
			if lootData == nil {
				okChann <- GenGatesEnemyRet(RESWarnEnemyIDErr)
				return
			}

			r.addKillValue(int(lootData.GetKillingValue()))
			pAdd := int(lootData.GetGEPoint())
			fData.currGEActivityPoint += pAdd
			r.gePointAll += pAdd
		}
	}
	r.needPush = true

	for i := 0; i < len(r.members); i++ {
		if r.members[i].AccountID == accountID {
			r.members[i] = info
		}
	}

	okChann <- GenGatesEnemyRet(RESSuccess)
}

func (r *GatesEnemyActivity) onFightBossBegin(accountID string,
	info helper.AccountSimpleInfo,
	bossIDx int,
	okChann chan<- GatesEnemyRet) {
	logs.Trace("onFightBossBegin %s -> %d %v", accountID, bossIDx, info)

	if bossIDx < 0 || bossIDx > r.bossMax {
		okChann <- GenGatesEnemyRet(RESErrNoBoss)
		return
	}

	fData, ok := r.memberData[accountID]
	if !ok || fData == nil {
		okChann <- GenGatesEnemyRet(RESWarnCanNotInAct)
		return
	}

	fData.isFighting = true
	fData.currBossId = bossIDx
	fData.startFightTime = time.Now().Unix()

	for i := 0; i < len(r.members); i++ {
		if r.members[i].AccountID == accountID {
			r.members[i] = info
		}
	}

	okChann <- GenGatesEnemyRet(RESSuccess)

}

func (r *GatesEnemyActivity) onFightBossEnd(accountID string,
	info helper.AccountSimpleInfo,
	bossIDx int, isSuccess bool,
	okChann chan<- GatesEnemyRet) {
	logs.Trace("onFightBossEnd %s -> %d %v", accountID, bossIDx, info)

	if bossIDx < 0 || bossIDx > r.bossMax {
		okChann <- GenGatesEnemyRet(RESErrNoBoss)
		return
	}

	fData, ok := r.memberData[accountID]
	if !ok || fData == nil {
		okChann <- GenGatesEnemyRet(RESWarnCanNotInAct)
		return
	}

	if !fData.isFighting {
		okChann <- GenGatesEnemyRet(RESErrStateErr)
		return
	}

	if bossIDx != fData.currBossId {
		okChann <- GenGatesEnemyRet(RESErrStateErr)
		return
	}

	// TODO by FanYang 服务器暂时不处理时间

	// FIXME by FanYang 处理Boss打没打过 校验是不是BossID

	fData.isFighting = false
	fData.currBossId = 0

	if isSuccess {
		data := gamedata.GetAllGEEnemyGroupCfg()
		logs.Trace("onFightEnd Give %v", data)
		if bossIDx < len(data) {
			enemyID := data[bossIDx].GetEGLevelID()
			lootData := gamedata.GetGEEnemyLootCfg(enemyID)
			if lootData == nil {
				okChann <- GenGatesEnemyRet(RESWarnEnemyIDErr)
				return
			}

			r.addKillValue(int(lootData.GetKillingValue()))
			pAdd := int(lootData.GetGEPoint())
			fData.currGEActivityPoint += pAdd
			r.gePointAll += pAdd
		}
	}
	r.needPush = true

	for i := 0; i < len(r.members); i++ {
		if r.members[i].AccountID == accountID {
			r.members[i] = info
		}
	}

	okChann <- GenGatesEnemyRet(RESSuccess)

}

func (g *GatesEnemyActivity) addBuff(acid, name string,
	okChann chan<- GatesEnemyRet) {
	if g.buffCurLv >= helper.GateEnemyBuffMaxLv {
		okChann <- GenGatesEnemyRet(RESWarnBuffAlready)
		return
	}
	nlv := g.buffCurLv + 1
	if g.buffMemAcid[nlv] != "" {
		okChann <- GenGatesEnemyRet(RESWarnBuffAlready)
		return
	}
	g.buffMemAcid[nlv] = acid
	g.buffMemName[nlv] = name
	g.buffCurLv = nlv

	g.needPush = true
	okChann <- GatesEnemyRet{
		Code:        RESSuccess,
		RetStrParam: g.buffMemName[:],
	}
}

func (r *GatesEnemyActivity) onDebugOp(accountID, guildID string,
	p1, p2, p3 int64,
	okChann chan<- GatesEnemyRet) {
	logs.SentryLogicCritical(accountID, "GatesEnemyActivity onDebugOp %s %d %d %d",
		guildID, p1, p2, p3)

	/*
	  p1  p2 p3
	   1  杀戮值  增加杀戮值
	*/

	switch p1 {
	case 1:
		r.addKillValue(int(p2))
		break
	case 2:
		// 能够调个人排行榜和公会排行榜的刷新时间（需要所有公会内的人一起刷新，以前那个调时间只能自己生效的cheat不能满足需求）
		if p2 >= 1 || p2 < 24*3600 {
			r.stateOverTime = time.Now().Unix() + p2
			r.needPush = true
		}

	}

	okChann <- GenGatesEnemyRet(RESSuccess)
}

type MemberInGatesEnemyActivityData struct {
	isFighting          bool
	isInAct             bool
	currBossId          int
	currEnemyType       int
	currEnemyIdx        int
	startFightTime      int64
	currKillPoint       int
	currGEActivityPoint int
	boss                []byte
}

func (m *MemberInGatesEnemyActivityData) GetState() int {
	if !m.isInAct {
		return GatesEnemyMemStateExit
	}

	if m.isFighting {
		return GatesEnemyMemStateFight
	} else {
		return GatesEnemyMemStateWait
	}

	return GatesEnemyMemStateNull
}

func (g *GatesEnemyActivity) updateMemberInGatesEnemyActivityData() {
	newMemberData := make(map[string]*MemberInGatesEnemyActivityData, len(g.members))
	for i := 0; i < len(g.members); i++ {
		id := g.members[i].AccountID
		d, ok := g.memberData[id]
		if !ok {
			newMemberData[id] = &MemberInGatesEnemyActivityData{}
		} else {
			newMemberData[id] = d
		}
	}
	g.memberData = newMemberData
}

func (g *GatesEnemyActivity) initEnemys() {
	nowT := time.Now().Unix()
	data := gamedata.GetAllGEEnemyGroupCfg()
	for _, e := range data {
		if e != nil && e.GetRenovateTime() > 0 {
			limit := int(e.GetNumberLimit())
			es := make([]byte, limit, limit)
			for i := 0; i < len(es); i++ {
				es[i] = 1
			}
			g.enemyInfo = append(g.enemyInfo, es)
		}
	}
	g.enemyLastUpdateTime = make([]int64, len(g.enemyInfo), len(g.enemyInfo))
	for i := 0; i < len(g.enemyLastUpdateTime); i++ {
		g.enemyLastUpdateTime[i] = nowT
	}
	g.needPush = true
}

func (g *GatesEnemyActivity) updateGatesEnemy(nowT int64) {
	if g.state != GatesEnemyActivityStateStarted {
		return
	}
	data := gamedata.GetAllGEEnemyGroupCfg()
	for i := 0; i < len(g.enemyInfo); i++ {
		inc := data[i].GetRenovateTime()

		if inc > 0 {
			isAllOne := true
			for j := 0; j < len(g.enemyInfo[i]); j++ {
				if g.enemyInfo[i][j] == 0 {
					isAllOne = false
				}
			}

			if isAllOne {
				g.enemyLastUpdateTime[i] = nowT
			}

			if nowT-g.enemyLastUpdateTime[i] >= int64(inc) {
				g.enemyLastUpdateTime[i] = nowT
				for j := 0; j < len(g.enemyInfo[i]); j++ {
					if g.enemyInfo[i][j] == 0 {
						g.enemyInfo[i][j] = 1
						g.needPush = true
						break
					}
				}
			}
		}
	}
}

func (g *GatesEnemyActivity) addKillValue(v int) {
	g.killPoint += v
	data := gamedata.GetAllGEEnemyGroupCfg()
	for i := g.bossMax + 1; i < len(data); i++ {
		actCondition := int(data[i].GetActiveCondition())
		if g.killPoint >= actCondition {
			g.bossMax = i
			g.killPoint -= actCondition
		} else {
			return
		}
	}
}
