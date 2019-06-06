package rank

import (
	"fmt"
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	rankbycop := RankByCorp{}
	rankbycop.topN.ScorePow = 100000
	for i := 0; i < 10; i++ {
		rankbycop.topN.TopN[i] = CorpDataInRank{
			Name:  fmt.Sprintf("name:%d", i),
			Score: int64(100000 + i),
		}
	}
	for _, ranker := range rankbycop.topN.TopN[:10] {
		fmt.Println(ranker.Name, ranker.Score, ranker.RedisScore)
	}
	rankbycop.topN.Add("profile:5", CorpDataInRank{
		Name:  "name:5",
		Score: 1,
	})

	rankbycop.topN.Add("profile:4", CorpDataInRank{
		Name:  "name:4",
		Score: 1,
	})
	time.Sleep(4 * time.Second)
	rankbycop.topN.Add("profile:3", CorpDataInRank{
		Name:  "name:3",
		Score: 1,
	})
	for _, ranker := range rankbycop.topN.TopN[:10] {
		fmt.Println(ranker.Name, ranker.Score, ranker.RedisScore)
	}
}
