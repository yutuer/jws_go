package csrob

import (
	"sync"
	"time"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type ticker struct {
	inQueue chan *tickReward
	rewards []*tickReward

	inQueueCommand chan *tickCommand
	commands       []*tickCommand

	tickerList []*tickerObj

	close chan struct{}
	sync.WaitGroup

	res *resources
}

type tickReward struct {
	Acid     string
	CarID    uint32
	EndStamp int64
	Rewarded bool
}

type tickCommand struct {
	notify   func()
	endStamp int64
}

func newTicker() *ticker {
	return &ticker{
		inQueue:        make(chan *tickReward, 1024),
		rewards:        make([]*tickReward, 0),
		inQueueCommand: make(chan *tickCommand, 1024),
		commands:       make([]*tickCommand, 0),
		close:          make(chan struct{}, 1),
		tickerList:     make([]*tickerObj, 0),
	}
}
func (t *ticker) Start() {
	go func() {
		t.Add(1)
		t.doTicker()
		t.Done()
	}()
	logs.Debug("[CSRob] ticker Start")
}

func (t *ticker) Stop() {
	close(t.close)
	t.Wait()
	logs.Debug("[CSRob] ticker Stop")
}

func (t *ticker) doTicker() {
	after := time.NewTicker(delayRewardDo)
	ct := time.NewTicker(intervalCommonTickerSecond)
	bClose := false
	for !bClose {
		select {
		case <-t.close:
			bClose = true
		case r := <-t.inQueue:
			func() {
				defer logs.PanicCatcherWithInfo("[CSRob] ticker doTicker pushRewardToBox Panic")
				logs.Trace("[CSRob] doTicker receive reward")
				t.rewards = append(t.rewards, r)
				t.res.RewardDB.pushRewardToBox(&RewardBoxElem{Acid: r.Acid, CarID: r.CarID, EndStamp: r.EndStamp})
			}()
		case c := <-t.inQueueCommand:
			logs.Trace("[CSRob] doTicker receive command")
			t.commands = append(t.commands, c)
		case <-after.C:
			func() {
				defer logs.PanicCatcherWithInfo("[CSRob] ticker doTicker doReward Panic")
				//logs.Trace("[CSRob] doTicker doReward")
				now := time.Now().Unix()
				t.doReward(now)
				t.doCommand(now)
			}()
		case now := <-ct.C:
			func() {
				defer logs.PanicCatcherWithInfo("[CSRob] ticker doTicker common ticker Panic")
				t.doTickerList(now)
			}()
		}
		//logs.Trace("[CSRob] doTicker ----------------")
	}

	logs.Trace("[CSRob] doTicker close")
}

func (t *ticker) doReward(now int64) {
	index := 0
	for index < len(t.rewards) {
		it := t.rewards[index]
		if true == it.Rewarded {
			// panic(makeError("doReward touch a rewarded tick"))
		}
		if it.EndStamp < now {
			if err := t.res.RewardDB.removeRewardFromBox(&RewardBoxElem{Acid: it.Acid, CarID: it.CarID, EndStamp: it.EndStamp}); nil != err {
				logs.Error("%v", err)
			} else {
				//发奖
				logs.Trace("[CSRob] doReward send reward {%v}", t.rewards[index])
				t.rewards[index].Rewarded = true
				t.rewards = append(t.rewards[:index], t.rewards[index+1:]...)

				t.res.CommandMod.notifyReward(it.Acid, it.CarID)
				continue
			}
		}
		index++
	}
}

func (t *ticker) doCommand(now int64) {
	index := 0
	for index < len(t.commands) {
		it := t.commands[index]
		if it.endStamp < now {
			//发奖
			logs.Trace("[CSRob] doCommand {%v}", t.commands[index])
			t.commands[index].notify()
			t.commands = append(t.commands[:index], t.commands[index+1:]...)
			continue
		}
		index++
	}
}

func (t *ticker) regReward(acid string, carID uint32, end int64) {
	reward := &tickReward{
		Acid:     acid,
		CarID:    carID,
		EndStamp: end,
		Rewarded: false,
	}

	logs.Trace("[CSRob] regReward {%v}", reward)
	t.inQueue <- reward
}

func (t *ticker) regCommand(notify func(), end int64) {
	tc := &tickCommand{
		notify:   notify,
		endStamp: end,
	}

	logs.Trace("[CSRob] regCommand {%v}", tc)
	t.inQueueCommand <- tc
}

func (t *ticker) reloadRewardList() {
	list, err := t.res.RewardDB.getAllRewardFromBox()
	if nil != err {
		logs.Error(fmt.Sprint(err))
		return
	}

	for _, reward := range list {
		t.rewards = append(t.rewards, &tickReward{Acid: reward.Acid, CarID: reward.CarID, EndStamp: reward.EndStamp, Rewarded: false})
	}
}

type tickerObj struct {
	maxInterval  int
	currInterval int
	action       func(time.Time)
}

func (t *ticker) regTickerToList(max, curr int, a func(time.Time)) {
	t.tickerList = append(t.tickerList, &tickerObj{
		maxInterval:  max,
		currInterval: curr,
		action:       a,
	})
}

func (t *ticker) doTickerList(now time.Time) {
	for _, elem := range t.tickerList {
		elem.currInterval++
		if elem.currInterval < elem.maxInterval {
			continue
		}
		elem.currInterval = 0
		elem.action(now)
	}
}
