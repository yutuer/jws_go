package worldboss

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamJoin ..
type ParamJoin struct {
	Sid    uint32
	Acid   string
	Player PlayerInfo
}

//RetJoin ..
type RetJoin struct {
	Boss        BossStatus
	MyPos       uint32
	MyDamage    uint64
	TotalDamage uint64
	Rank        []DamageRankElemInfo
}

//MethodJoin ..
type MethodJoin struct {
	module.BaseMethod
}

func newMethodJoin(m module.Module) *MethodJoin {
	return &MethodJoin{
		module.BaseMethod{Method: MethodJoinID, Module: m},
	}
}

//NewParam ..
func (m *MethodJoin) NewParam() module.Param {
	return &ParamJoin{}
}

//NewRet ..
func (m *MethodJoin) NewRet() module.Ret {
	return &RetJoin{}
}

//Do ..
func (m *MethodJoin) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamJoin)
	logs.Trace("[WorldBoss] [%d][%s] MethodJoin Do .. Param %+v", t.GroupID, t.HashSource, param)

	bm := m.ModuleAt().(*WorldBoss)

	info := param.Player
	bm.res.PlayerMod.updatePlayerInfo(&info)
	bm.res.PlayerMod.clearDamageInLife(param.Player.Acid)

	boss := bm.res.BossMod.getCurrBossStatus()
	common := bm.res.BossMod.getCommonStatus()
	topQuick := bm.res.RankDamageMod.getQuick()
	myRank, myContext := bm.res.RankDamageMod.getMyRankWithContext(param.Acid)
	rankList := topQuick
	if 0 != myRank.Pos && myRank.Pos > uint32(len(topQuick))+1 {
		rankList = append(rankList, *myRank)
		rankList = append(rankList, myContext...)
	}

	ret = &RetJoin{
		Boss:        *boss,
		MyPos:       myRank.Pos,
		MyDamage:    myRank.Damage,
		TotalDamage: common.TotalDamage,
		Rank:        rankList,
	}
	logs.Trace("[WorldBoss] [%d][%s] MethodJoin Do .. Ret %+v", t.GroupID, t.HashSource, ret)
	errCode = message.ErrCodeOK
	return
}
