package guild_boss

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (a *ActivityState) OnDamage(boss *BossState, damage int64) (float64, uint32) {
	nHp := boss.Hp - damage
	if nHp < 0 {
		nHp = 0
	}

	var gbRealCount uint32
	if damage > 0 && nHp == 0 {
		cfg := gamedata.GetGuildBossDataByLv(uint32(a.BossDegree))
		bossCfg := cfg.GetBossData_Table()[boss.Idx]
		logs.Debug("guild boss degree %d, %d", a.BossDegree, bossCfg.GetDropID())
		gbRealCount = a.OnGuildBossDied(bossCfg.GetGuildBossName(), bossCfg.GetDropID())
	}

	res := float64(damage) / float64(boss.TotalHp)
	boss.Hp = nHp

	return res, gbRealCount
}

func (a *ActivityState) OnGuildBossDied(bossName string, dropID string) uint32 {
	rewards := gamedata.GetGuildBossDroplist(dropID)
	return a.GetGuildHandler().OnGuildBossDied(bossName, rewards.Ids, rewards.Counts)
}
