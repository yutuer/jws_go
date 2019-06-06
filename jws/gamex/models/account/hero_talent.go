package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type PlayerHeroTalent struct {
	TalentPoint    uint32 `json:"tp"`
	TPUpdateLastTS int64  `json:"tp_l_up_ts"`

	HeroTalentLevel [helper.AVATAR_NUM_MAX][helper.MaxTalentCount]uint32 `json:"tp_lv"`
}

func (t *PlayerHeroTalent) OnAfterLogin(now_t int64) {
	if t.TPUpdateLastTS <= 0 {
		t.TPUpdateLastTS = now_t
		t.TalentPoint = gamedata.GetHeroCommonConfig().GetHSPointLimit()
	}
}

func (t *PlayerHeroTalent) UpdateTalentPoint(now_t int64) int64 {
	ccfg := gamedata.GetHeroCommonConfig()
	if t.TalentPoint >= ccfg.GetHSPointLimit() ||
		t.TPUpdateLastTS > now_t {
		return 0
	}
	dt := now_t - t.TPUpdateLastTS
	t.TalentPoint += uint32(dt / int64(ccfg.GetHSPointTime()))
	leftT := dt % int64(ccfg.GetHSPointTime())
	t.TPUpdateLastTS = now_t - leftT
	if t.TalentPoint >= ccfg.GetHSPointLimit() {
		t.TalentPoint = ccfg.GetHSPointLimit()
		return 0
	}
	return leftT
}

func (t *PlayerHeroTalent) UseTalentPoint(now_t int64) {
	if t.TalentPoint <= 0 {
		return
	}
	if t.TalentPoint >= gamedata.GetHeroCommonConfig().GetHSPointLimit() {
		t.TPUpdateLastTS = now_t
	}
	t.TalentPoint--
}

func (t *PlayerHeroTalent) BuyTalentPoint() {
	t.TalentPoint = gamedata.GetHeroCommonConfig().GetHSPointLimit()
}

func (t *PlayerHeroTalent) DebugAddTP(i uint32, now_t int64) {
	t.UpdateTalentPoint(now_t)
	ccfg := gamedata.GetHeroCommonConfig()
	t.TalentPoint += i
	if t.TalentPoint > ccfg.GetHSPointLimit() {
		t.TalentPoint = ccfg.GetHSPointLimit()
	}
}

func (t *PlayerHeroTalent) ActTalentByStar(avatarId int, star uint32) {
	talentIds := gamedata.GetHeroStarUnlockTalent(star)
	for _, tid := range talentIds {
		if t.HeroTalentLevel[avatarId][tid] <= 0 {
			t.HeroTalentLevel[avatarId][tid] = 1
			logs.Debug("ActTalentByStar avatarId %d star %d talent %d",
				avatarId, star, tid)
		}
	}
}
