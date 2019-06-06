package worldboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamLeave ..
type ParamLeave struct {
	Sid  uint32
	Acid string
	Team TeamInfoDetail

	BadCheat bool
}

//RetLeave ..
type RetLeave struct {
	Boss     BossStatus
	MyPos    uint32
	MyDamage uint64
}

//MethodLeave ..
type MethodLeave struct {
	module.BaseMethod
}

func newMethodLeave(m module.Module) *MethodLeave {
	return &MethodLeave{
		module.BaseMethod{Method: MethodLeaveID, Module: m},
	}
}

//NewParam ..
func (m *MethodLeave) NewParam() module.Param {
	return &ParamLeave{}
}

//NewRet ..
func (m *MethodLeave) NewRet() module.Ret {
	return &RetLeave{}
}

//Do ..
func (m *MethodLeave) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamLeave)
	logs.Trace("[WorldBoss] [%d][%s] MethodLeave Do .. Param %+v", t.GroupID, t.HashSource, param)

	bm := m.ModuleAt().(*WorldBoss)

	heros := make([]HeroInfoDetail, len(param.Team.Team))
	copy(heros, param.Team.Team)
	info := param.Team.Copy()
	info.DamageInLife = bm.res.PlayerMod.getDamageInLife(param.Acid)
	bm.res.PlayerMod.clearDamageInLife(param.Acid)

	if true == param.BadCheat {
		logs.Warn("[WorldBoss] [%d][%s] MethodLeave, BadCheat, Acid [%s], DamageInLife [%d]", t.GroupID, t.HashSource, param.Acid, info.DamageInLife)
		bm.res.RankDamageMod.addPlayerDamage(param.Sid, param.Acid, -int64(info.DamageInLife))
		info.DamageInLife = 0
	}

	bm.res.PlayerMod.updateTeamInfo(param.Acid, info)
	if 0 != info.DamageInLife {
		bm.res.FormationRankMod.addPlayerFormation(param.Sid, param.Acid, info.DamageInLife, info.Team, info.BuffLevel)
	}

	boss := bm.res.BossMod.getCurrBossStatus()
	myRank := bm.res.RankDamageMod.getMyRank(param.Acid)

	ret = &RetLeave{
		Boss:     *boss,
		MyPos:    myRank.Pos,
		MyDamage: myRank.Damage,
	}
	logs.Trace("[WorldBoss] [%d][%s] MethodLeave Do .. Ret %+v", t.GroupID, t.HashSource, ret)
	errCode = message.ErrCodeOK
	return
}
