package sPvpRander

import (
	"errors"
	"math/rand"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type simplePvpRander struct {
	pools [][]string
}

func isSelf(selfID, idInAvatarRank string) bool {
	logs.Trace("isSelf %s %s", selfID, idInAvatarRank)
	return selfID == idInAvatarRank
}

func (s *simplePvpRander) randEnemy(selfID string, rank, count int) ([]string, error) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("randEnemy panic, Err %v", err)
		}
	}()
	if s.pools == nil || len(s.pools) < 1 {
		return nil, errors.New("PoolNoInit")
	}

	ok, poolIdx, _, _ := gamedata.GetPvpPoolData(rank)

	if !ok || poolIdx >= len(s.pools) || poolIdx < 0 {
		poolIdx = len(s.pools) - 1
	}

	for poolIdx >= 0 {
		pool := s.pools[poolIdx]
		res := make([]string, 0, count)
		poolLen := len(pool)
		rIdxStart := rand.Intn(poolLen)

		for i, j := 0, 0; i < poolLen && j < count; i++ {
			id := pool[(i+rIdxStart)%poolLen]
			if id != "" && !isSelf(selfID, id) {
				res = append(res, id)
				j++
			}
		}

		if len(res) >= count {
			return res[:], nil
		} else {
			poolIdx -= 1
		}

	}

	return nil, errors.New("NoCount")
}
