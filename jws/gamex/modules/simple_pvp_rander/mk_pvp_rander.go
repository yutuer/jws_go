package sPvpRander

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (s *simplePvpRander) Make(sid uint) error {
	n := gamedata.GetPvpPoolMaxRank()

	topN, err := rank.GetModule(sid).RankSimplePvp.GetTopNAccountID(n)
	if err != nil || topN == nil {
		return err
	}

	s.buildPools()

	lastPoolIdx := -1 // 用来计算这一次的是这个pool中的第几个
	poolCount := 0    // 这一次的是这个pool中的第几个
	for pos := 1; pos <= len(topN); pos++ {
		ok, poolIdx, poolSize, poolRankIdx := gamedata.GetPvpPoolData(pos)
		if !ok {
			//logs.Warn("rank Idx %d no find data!", pos)
			continue
		}

		if lastPoolIdx != poolIdx {
			poolCount = 0
			lastPoolIdx = poolIdx
			newPool := make([]string, poolSize, poolSize)
			s.pools = append(s.pools, newPool[:])
		} else {
			poolCount++
		}
		accIdxInPool := poolRankIdx[poolCount]
		s.pools[poolIdx][accIdxInPool] = topN[pos-1]
	}

	logs.Trace("Pools %v", s.pools)

	return nil
}

func (s *simplePvpRander) buildPools() {
	poolSizes := gamedata.GetPvpPoolSizeData()
	s.pools = make([][]string, 0, len(poolSizes))
}
