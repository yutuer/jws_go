package main

import (
	"encoding/json"
	"flag"
	"math/rand"
	"time"

	"sync"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func main() {
	logs.Trace("use ./tester -c -t -m \"http://10.0.1.11:8888\"")
	countPreSecond := flag.Int("c", 10, "post count pre second")
	threadCount := flag.Int("t", 1, "thread count")
	url := flag.String("m", "http://127.0.0.1:8791", "match server address")
	flag.Parse()
	add := *url + helper.MatchPostUrlAddressV2
	wg := sync.WaitGroup{}
	wg.Add(*threadCount)
	c := sync.WaitGroup{}
	for i := 0; i < *threadCount; i++ {
		go func(id int) {
			defer wg.Done()
			logs.Trace("id %d start", id)
			t := time.Second
			tw := time.After(t)
			postCount := 0
			for {
				<-tw
				tw = time.After(t)
				data := helper.MatchValue{}
				for l := 0; l < *countPreSecond; l++ {
					data.IsHard = true
					data.CorpLv = uint32(rand.Int() % 40)
					data.AccountID = "0:10:" + uuid.NewV4().String()
					d, _ := json.Marshal(data)
					c.Add(1)
					go func() {
						defer c.Done()
						util.HttpPost(add,
							util.JsonPostTyp, d)
					}()
					postCount++
				}
				if postCount%(*countPreSecond*10) == 0 {
					logs.Trace("posted %d %d %v", id, postCount, c)
					logs.Flush()
				}
			}
		}(i)
	}

	wg.Wait()
	logs.Close()
}
