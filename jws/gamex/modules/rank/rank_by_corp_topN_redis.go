package rank

import (
	"encoding/json"

	"runtime"

	"vcs.taiyouxi.net/jws/gamex/models/account/simple_info"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type CorpTopNRedis struct {
	TopN           [RankTopSize]CorpDataInRank `json:"topN"`
	MinScoreToTopN int64                       `json:"min"`
	CurrPowValue   int64                       `json:"pv"` // 用于确定排序原则 同分先进居上

	acid2PosScoreCache map[string]PairPosScore
	rankdb             *rankDB
	rank_name          string
	db_name            string
}

func (c *CorpTopNRedis) clean() {
	c.TopN = [RankTopSize]CorpDataInRank{}
	c.MinScoreToTopN = 0
}

func (c *CorpTopNRedis) fromDB(data []byte) {
	err := json.Unmarshal(data, c)
	if err != nil {
		logs.Error("CorpDataInRank fromDB Err %s in %s", err.Error(), data)
	}
}

func (c *CorpTopNRedis) toDB() ([]byte, error) {
	return json.Marshal(*c)
}

func (c *CorpTopNRedis) setTopN(TopNFromOther [RankTopSize]CorpDataInRank) {
	c.TopN = TopNFromOther
	c.MinScoreToTopN = c.TopN[len(c.TopN)-1].getScore()
}

func (c *CorpTopNRedis) setAcid2PosScoreCache(a2p map[string]PairPosScore) {
	c.acid2PosScoreCache = a2p
}

// imp sort

// Len is part of sort.Interface.
func (s *CorpTopNRedis) Len() int {
	return len(s.TopN)
}

// Swap is part of sort.Interface.
func (s *CorpTopNRedis) Swap(i, j int) {
	s.TopN[i], s.TopN[j] = s.TopN[j], s.TopN[i]
}

// Less is part of sort.Interface.
func (s *CorpTopNRedis) Less(i, j int) bool {
	return s.TopN[i].getScore() > s.TopN[j].getScore()
}

// imp sort end

func (c *CorpTopNRedis) isTopN(score int64) bool {
	if c.TopN[len(c.TopN)-1].Name == "" {
		return true
	}
	return score >= c.MinScoreToTopN
}

func (c *CorpTopNRedis) isHasInTopN(acid string) int {
	for i := 0; i < len(c.TopN); i++ {
		if acid == c.TopN[i].getId() {
			return i
		}
	}
	return -1
}

func (c *CorpTopNRedis) Rename(acid string, name string) {
	for i := 0; i < len(c.TopN); i++ {
		if acid == c.TopN[i].getId() {
			c.TopN[i].Name = name
			return
		}
	}
}

func (c *CorpTopNRedis) Add(acid string, a CorpDataInRank) bool {
	//logs.Trace("CorpTopNRedis Add %s %v", acid, a)
	if c.isTopN(a.getScore()) {
		//logs.Trace("CorpTopNRedis True Add %s %v", acid, a)
		tail := len(c.TopN) - 1

		//logs.Trace("getTopWithScoreFromRedis %v", *c)
		topNIDs, topNScores, err := c.rankdb.getTopWithScoreFromRedis(c.rank_name, c.db_name)

		if err != nil {
			logs.Error("getTopFromRedis Err By %s", err.Error())
			return false
		}

		//logs.Trace("CorpTopNRedis getTopWithScoreFromRedis %v %v", topNIDs, topNScores)

		newTopN := [RankTopSize]CorpDataInRank{}

		for idx, id := range topNIDs {
			if idx < 0 || idx >= len(newTopN) {
				//logs.Error("ParseAccount %d", idx)
				continue
			}
			if idx >= len(topNScores) {
				//logs.Error("ParseAccount %d", idx)
				continue
			}
			accountDBID, err := db.ParseAccount(id)
			if err != nil {
				logs.Error("ParseAccount %s Err By %s", accountDBID, err.Error())
				continue
			}
			accountData, err := simple_info.LoadAccountSimpleInfoProfile(accountDBID)
			if err != nil {
				logs.Error("LoadAccount %s Err By %s", accountDBID, err.Error())
				continue
			}
			newTopN[idx].SetDataFromAccount(accountData, topNScores[idx])
		}

		//logs.Trace("CorpTopNRedis setTopN %v", newTopN)

		c.TopN = newTopN
		c.MinScoreToTopN = c.TopN[tail].getScore()
		return true
	} else {
		return false
	}
}

func (c *CorpTopNRedis) Update() {
	tail := len(c.TopN) - 1

	//logs.Trace("getTopWithScoreFromRedis %v", *c)
	topNIDs, topNScores, err := c.rankdb.getTopWithScoreFromRedis(c.rank_name, c.db_name)

	if err != nil {
		logs.Error("getTopFromRedis Err By %s", err.Error())
		return
	}

	//logs.Trace("CorpTopNRedis getTopWithScoreFromRedis %v %v", topNIDs, topNScores)

	newTopN := [RankTopSize]CorpDataInRank{}

	for idx, id := range topNIDs {
		if idx < 0 || idx >= len(newTopN) {
			//logs.Error("ParseAccount %d", idx)
			continue
		}
		if idx >= len(topNScores) {
			//logs.Error("ParseAccount %d", idx)
			continue
		}
		accountDBID, err := db.ParseAccount(id)
		if err != nil {
			logs.Error("ParseAccount %s Err By %s", accountDBID, err.Error())
			continue
		}
		accountData, err := simple_info.LoadAccountSimpleInfoProfile(accountDBID)
		if err != nil {
			logs.Trace("LoadAccount %s Err By %s", accountDBID, err.Error())
			continue
		}
		newTopN[idx].SetDataFromAccount(accountData, topNScores[idx])
		runtime.Gosched()
	}

	//logs.Trace("CorpTopNRedis setTopN %v", newTopN)

	c.TopN = newTopN
	c.MinScoreToTopN = c.TopN[tail].getScore()
}
