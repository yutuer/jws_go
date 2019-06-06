package main

import (
	"math/rand"
	"testing"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func TestMakeData(t *testing.T) {
	conn, err := redis.Dial("tcp", "127.0.0.1:6379", redis.DialDatabase(5))
	if nil != err {
		t.Errorf("redis.Dial failed, %v", err)
		return
	}

	gs := []string{
		"1:11:",
		"2:12:",
	}
	count := 1000

	for _, kgs := range gs {
		//RankCorpHeroStar
		for i := 0; i < count; i++ {
			rs := float64(0)
			baseScore := rand.Int63() % 9000
			if 0 == rand.Int31()%3 {
				rs = float64(baseScore * PowerBase)
			} else {
				rs = float64(baseScore)
			}

			conn.Do("ZADD", kgs+"RankCorpHeroStar", rs, kgs+uuid.NewV4().String())
		}

		//RankCorpTrial
		for i := 0; i < count; i++ {
			rs := float64(0)
			baseScore := rand.Int63() % 9000
			if 0 == rand.Int31()%3 {
				rs = float64(baseScore * PowerBase)
			} else {
				rs = float64(baseScore)
			}

			conn.Do("ZADD", kgs+"RankCorpTrial", rs, kgs+uuid.NewV4().String())
		}

		//RankCorpHeroDiff:TU
		for i := 0; i < count; i++ {
			rs := float64(0)
			baseScore := rand.Int63() % 9000
			if 0 == rand.Int31()%3 {
				rs = float64(baseScore * PowerBase)
			} else {
				rs = float64(baseScore)
			}

			conn.Do("ZADD", kgs+"RankCorpHeroDiff:TU", rs, kgs+uuid.NewV4().String())
		}

		//RankCorpHeroDiff:ZHAN
		for i := 0; i < count; i++ {
			rs := float64(0)
			baseScore := rand.Int63() % 9000
			if 0 == rand.Int31()%3 {
				rs = float64(baseScore * PowerBase)
			} else {
				rs = float64(baseScore)
			}

			conn.Do("ZADD", kgs+"RankCorpHeroDiff:ZHAN", rs, kgs+uuid.NewV4().String())
		}

		//RankCorpHeroDiff:HU
		for i := 0; i < count; i++ {
			rs := float64(0)
			baseScore := rand.Int63() % 9000
			if 0 == rand.Int31()%3 {
				rs = float64(baseScore * PowerBase)
			} else {
				rs = float64(baseScore)
			}

			conn.Do("ZADD", kgs+"RankCorpHeroDiff:HU", rs, kgs+uuid.NewV4().String())
		}

		//RankCorpHeroDiff:SHI
		for i := 0; i < count; i++ {
			rs := float64(0)
			baseScore := rand.Int63() % 9000
			if 0 == rand.Int31()%3 {
				rs = float64(baseScore * PowerBase)
			} else {
				rs = float64(baseScore)
			}

			conn.Do("ZADD", kgs+"RankCorpHeroDiff:SHI", rs, kgs+uuid.NewV4().String())
		}
	}

}
