package csrob

import (
	"time"
)

//RedisPoolSize ..
const RedisPoolSize = 5

//返回值定义
const (
	RetInvalid = iota
	RetOK

	RetLocked
	RetCountLimit
	RetTimeout
	RetHasHelper
	RetCannotAgain
)

//各种范围/区间定义
const (
	maxLoadAppealList = 60

	scaleSaveRecordsNum = 5

	guildRecommendNum = 100

	delayRankDo   = 10 * time.Second
	delayRewardDo = 3 * time.Second

	intervalCommonTicker       = 1 * time.Minute
	intervalCommonTickerSecond = 1 * time.Second
)
