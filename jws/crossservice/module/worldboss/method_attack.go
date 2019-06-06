package worldboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamAttack struct {
	Sid    uint32
	Acid   string
	Attack AttackInfo
}

//RetAttack ..
type RetAttack struct {
	Boss        BossStatus
	MyPos       uint32
	MyDamage    uint64
	DamageRound uint64
	TotalDamage uint64
	Rank        []DamageRankElemInfo
}

//MethodAttack ..
type MethodAttack struct {
	module.BaseMethod
}

func newMethodAttack(m module.Module) *MethodAttack {
	return &MethodAttack{
		module.BaseMethod{Method: MethodAttackID, Module: m},
	}
}

//NewParam ..
func (m *MethodAttack) NewParam() module.Param {
	return &ParamAttack{}
}

//NewRet ..
func (m *MethodAttack) NewRet() module.Ret {
	return &RetAttack{}
}

//Do ..
func (m *MethodAttack) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamAttack)
	logs.Trace("[WorldBoss] [%d][%s] MethodAttack Do .. Param %+v", t.GroupID, t.HashSource, param)

	bm := m.ModuleAt().(*WorldBoss)

	if 0 != param.Attack.Damage {
		validDamage := bm.res.BossMod.attackBoss(param.Attack.Level, param.Attack.Damage)
		bm.res.RankDamageMod.addPlayerDamage(param.Sid, param.Acid, int64(validDamage))
		bm.res.PlayerMod.addDamageInLife(param.Acid, validDamage)
	}

	boss := bm.res.BossMod.getCurrBossStatus()
	common := bm.res.BossMod.getCommonStatus()
	topQuick := bm.res.RankDamageMod.getQuick()
	myRank, myContext := bm.res.RankDamageMod.getMyRankWithContext(param.Acid)
	rankList := topQuick
	if 0 != myRank.Pos && myRank.Pos > uint32(len(topQuick))+1 {
		rankList = append(rankList, myContext...)
	}
	damageInLife := bm.res.PlayerMod.getDamageInLife(param.Acid)

	ret = &RetAttack{
		Boss:        *boss,
		MyPos:       myRank.Pos,
		MyDamage:    myRank.Damage,
		DamageRound: damageInLife,
		TotalDamage: common.TotalDamage,
		Rank:        rankList,
	}
	logs.Trace("[WorldBoss] [%d][%s] MethodAttack Do .. Ret %+v", t.GroupID, t.HashSource, ret)
	errCode = message.ErrCodeOK
	return
}
