package rank

import (
	"encoding/json"
	"sort"

	"runtime"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/account/simple_info"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type CorpDataInRank struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CorpLv     int     `json:"corplv"`
	Gs         int     `json:"gs"`
	Score      int64   `json:"score"`
	RedisScore float64 `json:"redis_score"` // 更改score的计算方式后， 类型是double的
	scoreDeta  int64
	Info       helper.AccountSimpleInfo `json:"info"`
	// 额外数据
	ExtraData []byte `json:"extra_data"`
	Sid       string `json:"-"`
}

func (c *CorpDataInRank) getId() string {
	return c.ID
}

func (c *CorpDataInRank) setScore(score int64) {
	c.Score = score
}

func (c *CorpDataInRank) getScore() int64 {
	return c.Score
}

func (c *CorpDataInRank) SetExtraData(data interface{}) {
	byteData, err := json.Marshal(data)
	if err != nil {
		logs.Error("json marshal err by %v", err)
		return
	}
	c.ExtraData = byteData
}

func (c *CorpDataInRank) SetDataFromAccount(a *helper.AccountSimpleInfo, score int64) {
	c.Name = a.Name
	c.ID = a.AccountID
	c.CorpLv = int(a.CorpLv)
	c.Score = score
	c.Gs = a.CurrCorpGs
	c.Info = *a
}

// 新版本， 使用double精度的score
func (c *CorpDataInRank) SetDataFromAccountWithDouble(a *helper.AccountSimpleInfo, score float64) {
	c.Name = a.Name
	c.ID = a.AccountID
	c.CorpLv = int(a.CorpLv)
	c.RedisScore = score
	c.Score = int64(score)
	c.Gs = a.CurrCorpGs
	c.Info = *a
}

func (c *CorpDataInRank) setDataFromAccountDeta(a *helper.AccountSimpleInfo, score int64) {
	c.Name = a.Name
	c.ID = a.AccountID
	c.CorpLv = int(a.CorpLv)
	c.Gs = a.CurrCorpGs
	c.scoreDeta = score
	c.Info = *a
}

func s2a(top1 []CorpDataInRank) [RankTopSize]CorpDataInRank {
	res := [RankTopSize]CorpDataInRank{}
	for i := 0; i < len(top1) && i < RankTopSize; i++ {
		res[i] = top1[i]
	}
	return res
}

/*
func (c *CorpDataInRank) setDataFromAccount2Client(a *helper.Avatar2Client, score int64) {
	c.Name = a.Name
	c.ID = a.GetAcId()
	c.CorpLv = int(a.CorpLv)
	c.GS = a.GS
	c.Score = score
}

func (c *CorpDataInRank) setDataFromAccount2ClientDeta(a *helper.Avatar2Client, score int64) {
	c.Name = a.Name
	c.ID = a.GetAcId()
	c.CorpLv = int(a.CorpLv)
	c.scoreDeta = score
	c.GS = a.GS
}
*/
func (c *CorpDataInRank) setScoreDeta(score int64) {
	c.scoreDeta = score
}

type CorpTopN struct {
	TopN           [RankTopSize]CorpDataInRank `json:"topN"`
	MinScoreToTopN int64                       `json:"min"`
	ScorePow       int64                       `json:"pow"`
	ScoreSpeckCurr int64                       `json:"speck"`

	rankdb    *rankDB
	rank_name string
	db_name   string
}

func (c *CorpTopN) clean() {
	c.TopN = [RankTopSize]CorpDataInRank{}
	c.MinScoreToTopN = 0
	c.ScoreSpeckCurr = 0
}

func (c *CorpTopN) fromDB(data []byte) {
	err := json.Unmarshal(data, c)
	if err != nil {
		logs.Error("CorpDataInRank fromDB Err %s in %s", err.Error(), data)
	}
}

func (c *CorpTopN) toDB() ([]byte, error) {
	return json.Marshal(*c)
}

func (c *CorpTopN) setTopN(TopNFromOther [RankTopSize]CorpDataInRank) {
	c.TopN = TopNFromOther
	c.MinScoreToTopN = c.TopN[len(c.TopN)-1].getScore()
}

// imp sort

// Len is part of sort.Interface.
func (s *CorpTopN) Len() int {
	return len(s.TopN)
}

// Swap is part of sort.Interface.
func (s *CorpTopN) Swap(i, j int) {
	s.TopN[i], s.TopN[j] = s.TopN[j], s.TopN[i]
}

// Less is part of sort.Interface.  降序
func (s *CorpTopN) Less(i, j int) bool {
	var scoreI, scoreJ float64
	if s.TopN[i].RedisScore != 0 {
		scoreI = s.TopN[i].RedisScore
	} else {
		scoreI = float64(s.TopN[i].Score)
	}
	if s.TopN[j].RedisScore != 0 {
		scoreJ = s.TopN[j].RedisScore
	} else {
		scoreJ = float64(s.TopN[j].Score)
	}
	return scoreI > scoreJ
}

// imp sort end

func (c *CorpTopN) isTopN(score int64) bool {
	s := score
	if c.ScorePow > 0 {
		s = s * c.ScorePow
	}
	return s >= c.MinScoreToTopN
}

func (c *CorpTopN) isHasInTopN(acid string) int {
	for i := 0; i < len(c.TopN); i++ {
		if acid == c.TopN[i].getId() {
			return i
		}
	}
	return -1
}

//  改函数涉及到score算法的变更
func (c *CorpTopN) Add(acid string, a CorpDataInRank) (bool, float64) {
	//logs.Trace("CorpTopN Add %s %v", acid, a)
	if c.isTopN(a.getScore()) { // 这里假设score只能上升不能下降
		tail := len(c.TopN) - 1

		a.RedisScore, a.Score = genScoreFromBase(a.getScore(), c.ScorePow)

		if idx := c.isHasInTopN(acid); idx >= 0 {
			c.TopN[idx] = a
		} else {
			c.TopN[tail] = a
		}

		sort.Sort(c) // 按分值降序排
		c.MinScoreToTopN = c.TopN[tail].getScore()
		return true, a.RedisScore
	} else {
		a.RedisScore, a.Score = genScoreFromBase(a.getScore(), c.ScorePow)
		return false, a.RedisScore
	}
}

// 由基础值生成排序分数  当scorePow = 1e5时 精度 3秒
func genScoreFromBase(baseScore int64, scorePow int64) (float64, int64) {
	now := time.Now().Unix()
	return float64(baseScore*scorePow) + 1e9*float64(scorePow)/float64(now), baseScore * scorePow
}

// 从老版本的分数转换成新版分数
func genScoreFromOld(scoreParam int64, baseScore int64) (float64, int64) {
	now := time.Now().Unix()
	return float64(scoreParam-scoreParam%baseScore) + 1e9*float64(baseScore)/float64(now), baseScore * scoreParam
}

func (c *CorpTopN) Reload() {
	tail := len(c.TopN) - 1

	topNIDs, topNScores, err := c.rankdb.getTopWithFloatScoreFromRedis(c.rank_name, c.db_name)

	if err != nil && err != redis.ErrNil {
		logs.Error("getTopFromRedis Err By %s", err.Error())
		return
	}

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
		newTopN[idx].SetDataFromAccountWithDouble(accountData, topNScores[idx])
		runtime.Gosched()
	}
	for i, data := range newTopN {
		for _, data2 := range c.TopN {
			if data.ID == data2.ID {
				newTopN[i].ExtraData = data2.ExtraData
			}
		}
	}
	c.TopN = newTopN
	c.MinScoreToTopN = c.TopN[tail].getScore()
	logs.Debug("load top n ", c.TopN)
}

func (c *CorpTopN) updateInfoWithoutScore(acid string, data CorpDataInRank) bool {
	for i, ranker := range c.TopN {
		if ranker.ID == acid {
			c.TopN[i].Name = data.Name
			c.TopN[i].Gs = data.Gs
			c.TopN[i].CorpLv = data.CorpLv
			c.TopN[i].Info = data.Info
			return true
		}
	}
	return false
}
