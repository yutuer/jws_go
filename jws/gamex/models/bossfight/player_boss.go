package bossfight

import (
	"errors"

	"math/rand"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// TODO By FanYang 一天打多少次需要配置
const BossFightCountDay = 3

type PlayerBoss struct {
	Bosses            [gamedata.MaxDegree]Boss `json:"boss"`
	MaxDegree         int                      `json:"mb"`
	MaxDegree2BossIdx int
}

func (p *PlayerBoss) OnAfterLogin(rd *rand.Rand) {
	// 兼容Boss数量改变的情况
	for i := 0; i < gamedata.GetBossFightCfgCount(); i++ {
		boss := gamedata.GetBoss(i, rd)
		if boss != nil {
			p.Bosses[i].FromData(boss)
		}
	}
}

func (p *PlayerBoss) BossFightDamage(acid string, boss_idx int, hp_del int64) (error, bool) {
	if boss_idx < 0 || boss_idx > len(p.Bosses) || p.Bosses[boss_idx].IsNil() {
		return errors.New("idxErr"), false
	}
	boss := &p.Bosses[boss_idx]
	logs.Trace("[%s]BossFightDamage %v,%d", acid, boss, hp_del)
	isSuccess := boss.MaxHp <= hp_del
	if isSuccess && int(boss.Degree) > p.MaxDegree {
		p.MaxDegree = int(boss.Degree)
		p.MaxDegree2BossIdx = boss_idx
	}
	return nil, isSuccess
}

func (p *PlayerBoss) GetBoss(boss_idx int) *Boss {
	if boss_idx < 0 || boss_idx > len(p.Bosses) || p.Bosses[boss_idx].IsNil() {
		return nil
	}
	return &p.Bosses[boss_idx]
}
