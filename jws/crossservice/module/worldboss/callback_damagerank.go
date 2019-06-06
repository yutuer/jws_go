package worldboss

import (
	"fmt"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamDamageRank ..
type ParamDamageRank struct {
	Sid       uint32
	Batch     string
	Rank      []DamageRankElem
	BossLevel uint32
}

//CallbackDamageRank ..
type CallbackDamageRank struct {
	module.BaseMethod
}

func newCallbackDamageRank(m module.Module) *CallbackDamageRank {
	return &CallbackDamageRank{
		module.BaseMethod{Method: CallbackDamageRankID, Module: m},
	}
}

//NewParam ..
func (m *CallbackDamageRank) NewParam() module.Param {
	return &ParamDamageRank{}
}

func (c *callbackHolder) DamageRank() error {
	rankLimit := gamedata.GetMaxRankLimit()
	killLimit := gamedata.GetKillBossRewardRankLimit()
	if rankLimit < killLimit {
		rankLimit = killLimit
	}
	list := c.res.RankDamageMod.getRange(0, rankLimit-1)
	if 0 == len(list) {
		logs.Warn("[WorldBoss] callbackHolder DamageRank getRange empty")
		return nil
	}
	dis := map[uint32][]DamageRankElem{}
	for _, elem := range list {
		if nil == dis[elem.Sid] {
			dis[elem.Sid] = []DamageRankElem{}
		}
		dis[elem.Sid] = append(dis[elem.Sid], elem)
	}
	currBoss := c.res.BossMod.getCurrBossStatus()
	bossLevel := currBoss.Level
	if 0 != currBoss.HPCurr && currBoss.Seq == currBoss.Level {
		bossLevel--
	}
	for sid, rank := range dis {
		param := &ParamDamageRank{
			Sid:       sid,
			Batch:     c.res.ticker.roundStatus.BatchTag,
			Rank:      rank,
			BossLevel: bossLevel,
		}
		logs.Info("[WorldBoss] callbackHolder DamageRank, Shard [%d] Batch [%s] BossLevel [%d] Rank [%+v]", param.Sid, param.Batch, param.BossLevel, param.Rank)
		if err := c.res.module.Push(sid, ModuleID, CallbackDamageRankID, param); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] callbackHolder DamageRank, Push to Shard [%d] failed, %v ...Param %+v", sid, err, param))
		}
	}
	return nil
}
