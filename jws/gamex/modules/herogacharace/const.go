package herogacharace

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/errorcode"
)

const (
	MAXRANK           = 100
	REDIS_PREFIX      = "hgr"
	ETCD_SERVICE_NAME = "HeroGachaRace/dbs"
	RANK_REFRESH_TIME = 5
)

const (
	WARN_NotReady = errCode.HeroGachaRaceNotReady
	WARN_NotStart = errCode.HeroGachaRaceNotStart
	WARN_NoRank   = errCode.HeroGachaRaceNoRank
)

var (
	WARN_ACTIVITY_NOT_READY            = errorcode.New("HeroGachaRace curActivity is not ready.", WARN_NotReady)
	WARN_ACTIVITY_DB_FAILED            = errorcode.New("HeroGachaRace DB is not ready.", WARN_NotReady)
	WARN_ACTIVITY_NO_RANK              = errorcode.New("HeroGachaRace RANK is out of min.", WARN_NoRank)
	WARN_ACTIVITY_ACTIVITYISNOTSTARTED = errorcode.New("HeroGachaRace activity is not started.", WARN_NotStart)
)
