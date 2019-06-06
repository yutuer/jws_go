package match

import (
	"sort"

	"math/rand"
	"time"

	gm "github.com/rcrowley/go-metrics"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/match_server/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type waitter struct {
	AcID        string
	CorpLv      uint32
	EnterTime   int64
	MaxWaitTime int64
}

type matchResults struct {
	w []MatchResult
}

func (m *matchResults) Init(cap int) {
	m.w = make([]MatchResult, 0, cap)
}

func (m *matchResults) AddResult(ws []waitter) {
	nm := MatchResult{NumMatched: len(ws)}
	for i, v := range ws {
		nm.AcIDs[i] = v.AcID
	}
	m.w = append(m.w, nm)
}

type waitterArray struct {
	w []waitter
}

func (w *waitterArray) Init(cap int) {
	w.w = make([]waitter, 0, cap)
}

func (w *waitterArray) Append(nw waitter) {
	w.w = append(w.w, nw)
}

func (w *waitterArray) AppendNew(acID string, corpLv uint32) {
	min_t, max_t := multConfig.Cfg.WaitRobotTimeMin, multConfig.Cfg.WaitRobotTimeMax
	rand.Seed(time.Now().Unix())
	true_t := rand.Intn(int(max_t-min_t)) + int(min_t)
	w.w = append(w.w, waitter{
		AcID:        acID,
		CorpLv:      corpLv,
		EnterTime:   time.Now().Unix(),
		MaxWaitTime: int64(true_t),
	})
}

func (w *waitterArray) Clean() {
	w.w = w.w[0:0]
}

func (w *waitterArray) Add(o *waitterArray) {
	for i := 0; i < len(o.w); i++ {
		w.w = append(w.w, o.w[i])
	}
}

func (w *waitterArray) Len() int {
	return len(w.w)
}

func (w *waitterArray) Less(i, j int) bool {
	return w.w[i].CorpLv < w.w[j].CorpLv
}

func (w *waitterArray) Swap(i, j int) {
	t := w.w[i]
	w.w[i] = w.w[j]
	w.w[j] = t
}

func (w *waitterArray) fastClean(needClean func(string) bool) {
	data := w.w
	n := len(data)
	i := 0
loop:
	for i < n {
		r := data[i]
		if needClean(r.AcID) {
			data[i] = data[n-1]
			n--
			continue loop
		}
		i++
	}
	w.w = data[0:n]
}

func (w *waitterArray) cleanTimeOut(now_time int64) []waitter {
	selected := make([]waitter, 0, 10)
	data := w.w
	n := len(data)
	i := 0
	for i < n {
		r := data[i]
		if now_time-r.EnterTime > r.MaxWaitTime {
			data[i] = data[n-1]
			n--
			selected = append(selected, r)
			continue
		}
		i++
	}
	w.w = data[0:n]
	return selected
}

const MAXWaitTicks = 50

const (
	matchInfoCountNumAll = iota
	matchInfoCountMatchNumAll
	matchInfoCountCount
)

type waitStatus struct {
	on     bool
	cancel bool
}
type waitters struct {
	waitterInTick       []*waitterArray
	waitterMoreThenTick []*waitterArray
	playerWaittingMap   map[string]waitStatus

	waitterTmp *waitterArray
	result     matchResults

	matched        map[string]struct{}
	matchInfo      [matchInfoCountCount]int
	timeoutCounter gm.Counter
}

func (w *waitters) logMatchInfo() {
	logs.Trace("logMatchInfo %v/%v %v",
		w.matchInfo[matchInfoCountMatchNumAll],
		w.matchInfo[matchInfoCountNumAll],
		float32(100)-100*float32(w.matchInfo[matchInfoCountMatchNumAll])/float32(w.matchInfo[matchInfoCountNumAll]))
}

func (w *waitters) Init(timeoutCounter gm.Counter, cap int) {
	MaxTick := multConfig.Cfg.TickMax
	w.waitterInTick = make([]*waitterArray, MaxTick)
	w.waitterMoreThenTick = make([]*waitterArray, MaxTick)
	for i := 0; i < multConfig.Cfg.TickMax; i++ {
		w.waitterInTick[i] = new(waitterArray)
		w.waitterMoreThenTick[i] = new(waitterArray)
		w.waitterInTick[i].Init(cap)
		w.waitterMoreThenTick[i].Init(cap)
	}
	w.waitterTmp = new(waitterArray)
	w.waitterTmp.Init(cap)
	w.result.Init(4096)
	w.playerWaittingMap = make(map[string]waitStatus, 10240)
	w.timeoutCounter = timeoutCounter
}

func (w *waitters) OnTick() {
	MaxTick := multConfig.Cfg.TickMax
	w.cleanCancelWaiting()
	timeOuts := w.waitterInTick[MaxTick-1] // 这个timeout里面也包括了已经匹配过的人
	timeOutMap := make(map[string]waitter, timeOuts.Len())
	for i := 0; i < timeOuts.Len(); i++ {
		timeOutMap[timeOuts.w[i].AcID] = timeOuts.w[i]
		delete(w.playerWaittingMap, timeOuts.w[i].AcID)
	}

	timeoutNum := w.waitterMoreThenTick[MaxTick-1].Len() -
		w.waitterMoreThenTick[MaxTick-2].Len()

	tail := w.waitterMoreThenTick[MaxTick-1]
	for i := MaxTick - 2; i >= 0; i-- {
		w.waitterMoreThenTick[i+1] = w.waitterMoreThenTick[i]
	}
	w.waitterMoreThenTick[0] = tail
	w.waitterMoreThenTick[0].Clean()
	w.waitterMoreThenTick[0].Add(w.waitterMoreThenTick[1])

	for i := 0; i < len(w.waitterMoreThenTick); i++ {
		w.waitterTmp.Clean()
		wmt := w.waitterMoreThenTick[i]
		for j := 0; j < wmt.Len(); j++ {
			_, ok := timeOutMap[wmt.w[j].AcID]
			if !ok {
				w.waitterTmp.Append(wmt.w[j])
			}
		}
		w.waitterMoreThenTick[i] = w.waitterTmp
		w.waitterTmp = wmt
	}

	for i := MaxTick - 2; i >= 0; i-- {
		w.waitterInTick[i+1] = w.waitterInTick[i]

	}
	w.timeoutCounter.Inc(int64(timeoutNum))
	timeOuts.Clean()
	w.waitterInTick[0] = timeOuts
}

func (w *waitters) match(ws *waitterArray, Lv uint32, maxMatch int) {
	sort.Sort(ws)
	for i := 0; i < ws.Len(); {
		step := i + maxMatch
		lastIdx := step - 1
		if (lastIdx < ws.Len()) &&
			(ws.w[lastIdx].CorpLv-ws.w[i].CorpLv <= Lv) {

			w.result.AddResult(ws.w[i:step])
			for jj := i; jj < step; jj++ {
				w.matched[ws.w[jj].AcID] = struct{}{}
				delete(w.playerWaittingMap, ws.w[jj].AcID)
			}
			w.matchInfo[matchInfoCountMatchNumAll] += maxMatch
			i += maxMatch

		} else {

			// 没有匹配中
			w.waitterTmp.Append(ws.w[i])
			i++

		}
	}
}

func (w *waitters) matchNewEnter(Lv uint32) {
	ws := w.waitterInTick[0]
	w.waitterTmp.Clean()
	w.matched = make(map[string]struct{})
	sort.Sort(ws)

	w.match(ws, Lv, MatchPlayerNum)

	w.waitterInTick[0] = w.waitterTmp
	ws.Clean()
	w.waitterTmp = ws

	wmt := w.waitterMoreThenTick[0]
	for j := 0; j < wmt.Len(); j++ {
		_, ok := w.matched[wmt.w[j].AcID]
		if !ok {
			w.waitterTmp.Append(wmt.w[j])
		}
	}
	w.waitterMoreThenTick[0] = w.waitterTmp
	w.waitterTmp = wmt
	w.waitterTmp.Clean()
}

func (w *waitters) matchWaitTicketMoreThan(ticket int, Lv uint32) {
	if ticket >= len(w.waitterMoreThenTick) {
		return
	}
	w.matched = make(map[string]struct{})
	ws := w.waitterMoreThenTick[ticket]
	w.waitterTmp.Clean()

	if ticket < 4 {
		w.match(ws, Lv, MatchPlayerNum)
	} else {
		w.match(ws, Lv, helper.MatchMinPlayerNum)
	}

	w.waitterMoreThenTick[ticket] = w.waitterTmp
	ws.Clean()
	w.waitterTmp = ws

	for i := 0; i < len(w.waitterMoreThenTick); i++ {
		wmt := w.waitterMoreThenTick[i]
		w.waitterTmp.Clean()
		for j := 0; j < wmt.Len(); j++ {
			_, ok := w.matched[wmt.w[j].AcID]
			if !ok {
				w.waitterTmp.Append(wmt.w[j])
			}
		}
		w.waitterMoreThenTick[i] = w.waitterTmp
		w.waitterTmp = wmt
		w.waitterTmp.Clean()
	}

	for i := 0; i < len(w.waitterInTick); i++ {
		wmt := w.waitterInTick[i]
		w.waitterTmp.Clean()
		for j := 0; j < wmt.Len(); j++ {
			_, ok := w.matched[wmt.w[j].AcID]
			if !ok {
				w.waitterTmp.Append(wmt.w[j])
			}
		}
		w.waitterInTick[i] = w.waitterTmp
		w.waitterTmp = wmt
		w.waitterTmp.Clean()
	}

}

func (w *waitters) CancelWaitter(acID string) bool {
	v, ok := w.playerWaittingMap[acID]
	if !ok {
		return false
	}
	v.cancel = true
	w.playerWaittingMap[acID] = v
	return true
}

func (w *waitters) AddWaitter(acID string, corpLv uint32) bool {
	v, ok := w.playerWaittingMap[acID]
	if ok {
		if v.cancel {
			//waiting cancel
			v.cancel = false
			w.playerWaittingMap[acID] = v
			return true
		} else {
			return false
		}
	}
	w.playerWaittingMap[acID] = waitStatus{on: true, cancel: false}
	w.waitterInTick[0].AppendNew(acID, corpLv)
	w.waitterMoreThenTick[0].AppendNew(acID, corpLv)
	w.matchInfo[matchInfoCountNumAll]++
	return true
}

func (w *waitters) logSelf(i int) {
	logs.Error("Start %v", i)
	logs.Trace("match %v", w.result.w)
	logs.Flush()
	logs.Warn("InTick")
	for i := 0; i < len(w.waitterInTick); i++ {
		logs.Trace("InTick %2d %v", i, w.waitterInTick[i].w)
		logs.Flush()
	}
	logs.Warn("More")
	for i := 0; i < len(w.waitterMoreThenTick); i++ {
		logs.Trace("More %2d %v", i, w.waitterMoreThenTick[i].w)
		logs.Flush()
	}
	logs.Error("End %v", i)
	logs.Flush()
}

func (w *waitters) TestMatch() {
	w.matchWaitTicketMoreThan(1, 3)
	w.matchWaitTicketMoreThan(3, 5)
	w.matchWaitTicketMoreThan(5, 7)
	w.matchWaitTicketMoreThan(10, 10)
	w.matchWaitTicketMoreThan(15, 15)
	//w.matchWaitTicketMoreThan(20, 20)
	//w.matchWaitTicketMoreThan(24, 35)
}

func (w *waitters) cleanCancelWaiting() {
	//TODO: YZH cleanCancelWaiting performace might be issue
	needClean := func(acId string) bool {
		if v, ok := w.playerWaittingMap[acId]; ok {
			if v.cancel {
				return true
			}
		}
		return false
	}
	for _, v := range w.waitterInTick {
		v.fastClean(needClean)
	}
	for _, v := range w.waitterMoreThenTick {
		v.fastClean(needClean)
	}

	//https://golang.org/doc/effective_go.html#for
	//it is safe deleting key in map for-range
	for k, v := range w.playerWaittingMap {
		if v.cancel {
			delete(w.playerWaittingMap, k)
		}
	}
}

func (w *waitters) Match(ticks []int, lvNew uint32, lvs []uint32) []MatchResult {
	w.cleanCancelWaiting()
	w.result.w = w.result.w[0:0]
	// 将等待时间过长的人从队列中剔除, 上层将为其匹配机器人
	//w.SelectWaitTimeMuch()
	if len(w.result.w) <= 0 && len(w.playerWaittingMap) < helper.MatchMinPlayerNum {
		//人数太少不需要Match
		return w.result.w[:]
	}

	w.matchNewEnter(lvNew)
	for idx, t := range ticks {
		w.matchWaitTicketMoreThan(t, lvs[idx])
	}
	return w.result.w[:]
}

func (w *waitters) NewMatch(lvNew uint32) []MatchResult {
	w.cleanCancelWaiting()
	w.result.w = w.result.w[0:0]
	// 将等待时间过长的人从队列中剔除, 上层将为其匹配机器人
	//w.SelectWaitTimeMuch()
	if len(w.result.w) <= 0 && len(w.playerWaittingMap) < helper.MatchMinPlayerNum {
		//人数太少不需要Match
		return w.result.w[:]
	}
	w.matchNewEnter(lvNew)
	return w.result.w[:]
}

//func (w *waitters) SelectWaitTimeMuch() {
//	now_time := time.Now().Unix()
//	selected := make([]waitter, 0, 100)
//	for _, v := range w.waitterInTick {
//		selected = append(selected, v.cleanTimeOut(now_time)[:]...)
//	}
//	for _, v := range w.waitterMoreThenTick {
//		v.cleanTimeOut(now_time)
//	}
//
//	//https://golang.org/doc/effective_go.html#for
//	//it is safe deleting key in map for-range
//
//	// 删除等待过长的玩家
//	for _, v := range selected {
//		_, ok := w.playerWaittingMap[v.AcID]
//		if ok {
//			delete(w.playerWaittingMap, v.AcID)
//		}
//	}
//	for k, v := range w.playerWaittingMap {
//		if v.cancel {
//			delete(w.playerWaittingMap, k)
//		}
//	}
//
//	// 等待时间足够,交给上层赋予与机器人交战的权限
//	for _, item := range selected {
//		w.result.AddResult([]waitter{item})
//	}
//
//}
