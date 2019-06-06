package rank

import (
	"encoding/json"

	"runtime"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GuildDataInRank struct {
	ID            int64  `json:"id"`
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Lv            uint32 `json:"lv"`
	Score         int64  `json:"act"`
	ChiefName     string `json:"chief"`
	GuildMemCount int    `json:"memc"`
	GuildMemMax   int    `json:"memmax"`
	scoreDeta     int64
}

func (c *GuildDataInRank) getId() string {
	return c.UUID
}

func (c *GuildDataInRank) setScore(score int64) {
	c.Score = score
}

func (c *GuildDataInRank) getScore() int64 {
	return c.Score
}

func (c *GuildDataInRank) SetDataGuild(a *guild_info.GuildSimpleInfo, score int64) {
	c.Name = a.Name
	c.ID = a.GuildID
	c.UUID = a.GuildUUID
	c.Lv = a.Level
	c.Score = score
	c.GuildMemCount = a.MemNum
	c.GuildMemMax = a.MaxMemNum
	c.ChiefName = a.LeaderName
}

type GuildTopN struct {
	TopN           [RankTopSize]GuildDataInRank `json:"topN"`
	MinScoreToTopN int64                        `json:"min"`
	CurrPowValue   int64                        `json:"pv"` // 用于确定排序原则 同分先进居上

	rankdb    *rankDB
	rank_name string
	db_name   string
}

func (c *GuildTopN) clean() {
	logs.Trace("GuildTopN Before Clean %v", *c)
	c.TopN = [RankTopSize]GuildDataInRank{}
	c.MinScoreToTopN = 0
}

func (c *GuildTopN) fromDB(data []byte) {
	err := json.Unmarshal(data, c)
	if err != nil {
		logs.Error("GuildDataInRank fromDB Err %s in %s", err.Error(), data)
	}
}

func (c *GuildTopN) toDB() ([]byte, error) {
	return json.Marshal(*c)
}

func (c *GuildTopN) setTopN(TopNFromOther [RankTopSize]GuildDataInRank) {
	c.TopN = TopNFromOther
	c.MinScoreToTopN = c.TopN[len(c.TopN)-1].getScore()
}

// imp sort

// Len is part of sort.Interface.
func (s *GuildTopN) Len() int {
	return len(s.TopN)
}

// Swap is part of sort.Interface.
func (s *GuildTopN) Swap(i, j int) {
	s.TopN[i], s.TopN[j] = s.TopN[j], s.TopN[i]
}

// Less is part of sort.Interface.
func (s *GuildTopN) Less(i, j int) bool {
	return s.TopN[i].getScore() > s.TopN[j].getScore()
}

// imp sort end

func (c *GuildTopN) isTopN(score int64) bool {
	return score >= c.MinScoreToTopN
}

func (c *GuildTopN) isInTopN(acid string) bool {
	for _, a := range c.TopN {
		if a.UUID == acid {
			return true
		}
	}

	return false
}

func (c *GuildTopN) isHasInTopN(acid string) int {
	for i := 0; i < len(c.TopN); i++ {
		if acid == c.TopN[i].getId() {
			return i
		}
	}
	return -1
}

func (c *GuildTopN) Add(acid string, a GuildDataInRank) bool {
	if c.isTopN(a.getScore()) {
		c.Update()
		return true
	} else {
		return false
	}
}

func (c *GuildTopN) Update() {
	tail := len(c.TopN) - 1
	topNIDs, topNScores, err := c.rankdb.getTopWithScoreFromRedis(c.rank_name, c.db_name)

	if err != nil {
		logs.Error("getTopFromRedis Err By %s", err.Error())
		return
	}

	newTopN := [RankTopSize]GuildDataInRank{}
	_db := driver.GetDBConn()
	defer _db.Close()
	if _db.IsNil() {
		logs.Error("GuildTopN Update cant get redis conn")
		return
	}

	for idx, id := range topNIDs {

		if idx < 0 || idx >= len(newTopN) {
			logs.Trace("GuildTopN Err by no newTopN %d", idx)
			continue
		}
		if idx >= len(topNScores) {
			continue
		}

		guildData := guild_info.LoadGuildInfo(id, _db)
		if guildData == nil {
			logs.Trace("GuildTopN Err by no data %s", id)
			c.rankdb.del(c.rank_name, c.db_name, id)
			continue
		}
		newTopN[idx].SetDataGuild(guildData, topNScores[idx])
		runtime.Gosched()
	}

	c.TopN = newTopN
	c.MinScoreToTopN = c.TopN[tail].getScore()
}

//
//
//
