package main

import (
	"time"

	"sync"

	"github.com/rcrowley/go-metrics"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/match_server/config"
	"vcs.taiyouxi.net/jws/multiplayer/match_server/match"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func main() {
	multConfig.Cfg.MatchTicks = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	multConfig.Cfg.NewEnterMatchLv = 33
	multConfig.Cfg.MatchLvs = []uint32{10, 20, 30, 40, 50, 60, 40, 50, 60}

	wait := sync.WaitGroup{}
	wait.Add(2)
	tickTime := time.Duration(10 * 1000 / 50)
	logs.Trace("t %v", tickTime)
	t := tickTime * time.Millisecond
	logs.Trace("t %v", t)

	go func() {

		defer wait.Done()
		matcher := match.Match{}
		matcher.Init(metrics.NewCounter(), 1024)
		var i int64 = 0
		const dd = 2 * 31 * 24 * 3600 * 100
		matcher.AddWaittingPlayer("1", 22)
		matcher.AddWaittingPlayer("2", 55)
		matcher.AddWaittingPlayer("3", 66)
		matcher.AddWaittingPlayer("4", 99)
		matcher.AddWaittingPlayer("5", 99)
		matcher.AddWaittingPlayer("6", 99)
		matcher.AddWaittingPlayer("7", 99)
		matcher.AddWaittingPlayer("8", 199)
		tick := time.After(tickTime * time.Millisecond)
		for ; i < dd; i++ {
			defer func() {
				if err := recover(); err != nil {
					logs.Error("GVEMatch Panic, Err %v", err)
				}
			}()
			select {
			case <-tick:
				if i%36000 == 0 {
					logs.Trace("v %d", i)
				}
				func() {
					matcher.MatchNewGame()
					matcher.GetWaitter().OnTick()
					matcher.MatchAllGame()
				}()
				tick = time.After(tickTime * time.Millisecond)
			}
		}
	}()
	go func() {
		defer wait.Done()
		matcher := match.Match{}
		matcher.Init(metrics.NewCounter(), 1024)
		var i int64 = 0
		const dd = 2 * 31 * 24 * 3600 * 100
		matcher.AddWaittingPlayer("1", 22)
		matcher.AddWaittingPlayer("2", 55)
		matcher.AddWaittingPlayer("3", 66)
		matcher.AddWaittingPlayer("4", 99)
		matcher.AddWaittingPlayer("5", 99)
		matcher.AddWaittingPlayer("6", 99)
		matcher.AddWaittingPlayer("7", 99)
		matcher.AddWaittingPlayer("8", 199)
		tick := time.After(tickTime * time.Millisecond)
		for ; i < dd; i++ {
			defer func() {
				if err := recover(); err != nil {
					logs.Error("GVEMatch Panic, Err %v", err)
				}
			}()
			select {
			case <-tick:
				if i%36000 == 0 {
					logs.Trace("v %d", i)
				}
				func() {
					matcher.MatchNewGame()
					matcher.GetWaitter().OnTick()
					matcher.MatchAllGame()
				}()
				tick = time.After(tickTime * time.Millisecond)
			}
		}
	}()
	wait.Wait()
	logs.Close()
}
