package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
)

type OppoRelated struct {
	LastSignTime       int64 `json:"last_sign_t"`
	LastDailyQuestTime int64 `json:"last_daily_quest_t"`
	SignDays           int   `json:"sign_days"`
	LastLoginTime      int64 `json:"last_login_t"`
}

func (or *OppoRelated) UpdateTime(nowT int64, isOppo bool) {
	if !gamedata.IsSameDayCommon(nowT, or.LastLoginTime) {
		if gamedata.IsSameDayCommon(nowT, or.LastLoginTime+util.DaySec) {
			if isOppo {
				or.SignDays += 1
			}
		} else {
			if isOppo {
				or.SignDays = 1
			} else {
				or.SignDays = 0
			}
		}
	}
	if isOppo {
		or.LastLoginTime = nowT
	}
}
