package friend

import (
	"time"

	"vcs.taiyouxi.net/platform/planx/util"
)

const UpdateFriendListInterval = 2 * util.MinSec

const UpdateBlackListInterval = 30 * util.MinSec

const UpdateRecentPlayerInterval = 2 * util.MinSec

const UpdateRecommendPlayerInterval = 10 * util.MinSec

const FriendCountPerRet = 10

const FriendCountPerReq = 50

const targetDBKey = "0:10:friends" //ZSET

const maxCachePlayer = 10000

const maxGSLevel = 1

const countPerScan = 100

const scanPollTime = 2 * util.MinSec

const maxTmpInfo = 1000

const selectGSCloseNum = 60

const maxCacheTime = 10 * util.DaySec

const countPerInit = 100

const db_counter_key = "Friend_DB"

const save_interval = 5 * time.Minute

const (
	_ = iota
	GiftCmd_Receive
	GiftCmd_Give
	GiftCmd_GetInfo
)
