package match

import (
	"errors"
	"sync"
	"time"

	gm "github.com/rcrowley/go-metrics"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/match_server/config"
	"vcs.taiyouxi.net/jws/multiplayer/match_server/notify"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const MatchPlayerNum = helper.MatchPlayerNum

var (
	normalStartMatchCounter gm.Counter
	hardStartMatchCounter   gm.Counter
	normalMatchedCounter    gm.Counter
	hardMatchedCounter      gm.Counter
	normalTimeoutCounter    gm.Counter
	hardTimeoutCounter      gm.Counter
	normalTickCounter       gm.Counter
)

func init() {
	normalStartMatchCounter = metrics.NewCounter("multiplay.match.startNormal")
	hardStartMatchCounter = metrics.NewCounter("multiplay.match.startHard")
	normalMatchedCounter = metrics.NewCounter("multiplay.match.matchedNormal")
	hardMatchedCounter = metrics.NewCounter("multiplay.match.matchedHard")
	normalTimeoutCounter = metrics.NewCounter("multiplay.match.timeoutNormal")
	hardTimeoutCounter = metrics.NewCounter("multiplay.match.timeourHard")
	normalTickCounter = metrics.NewCounter("multiplay.match.tickNormal")
}

var (
	ErrTimeout = errors.New("WarmKeyUrlErrTimeout")
)

type PlayerInfo struct {
	AcID          string
	WaitStartTime int64
}

type MatchResult struct {
	AcIDs      [MatchPlayerNum]string
	NumMatched int
}

func (mr MatchResult) GetResult() []string {
	return mr.AcIDs[:mr.NumMatched]
}

type Match struct {
	//playerWaitting    []PlayerInfo

	//playLeave         [MatchPlayerNum]PlayerInfo

	wait waitters
}

func (m *Match) Init(c gm.Counter, cap int) {
	//m.playerWaitting = make([]PlayerInfo, 0, cap*MatchPlayerNum)
	//m.playerWaittingMap = make(map[string]bool, cap*MatchPlayerNum)
	m.wait.Init(c, cap*MatchPlayerNum)
}

func (m *Match) GetWaitter() *waitters {
	return &m.wait
}

func (m *Match) MatchAllGame() []MatchResult {
	/*
		num := len(m.playerWaitting) / MatchPlayerNum
		if num > 0 {
			res := make([]MatchResult, num, num)
			for i := 0; i < num; i++ {
				all := i * MatchPlayerNum
				for j := all; j < all+MatchPlayerNum; j++ {
					res[i].AcIDs[j-all] = m.playerWaitting[j].AcID
					delete(m.playerWaittingMap, m.playerWaitting[j].AcID)
				}
			}
			for i := num * MatchPlayerNum; i < len(m.playerWaitting); i++ {
				m.playLeave[i-num*MatchPlayerNum] = m.playerWaitting[i]
			}
			leaveNum := len(m.playerWaitting) - num*MatchPlayerNum
			m.playerWaitting = m.playerWaitting[:leaveNum]
			for i := 0; i < len(m.playerWaitting); i++ {
				m.playerWaitting[i] = m.playLeave[i]
			}
			return res[:]
		} else {
			return nil
		}
	*/
	return m.wait.Match(
		multConfig.Cfg.MatchTicks,
		multConfig.Cfg.NewEnterMatchLv,
		multConfig.Cfg.MatchLvs)
}

func (m *Match) MatchNewGame() []MatchResult {
	return m.wait.NewMatch(
		multConfig.Cfg.NewEnterMatchLv)
}

func (m *Match) CancelWaiting(acID string) bool {
	return m.wait.CancelWaitter(acID)
}

func (m *Match) AddWaittingPlayer(acID string, corpLv uint32) bool {
	/*
		_, ok := m.playerWaittingMap[acID]
		if ok {
			return
		}
		m.playerWaitting = append(m.playerWaitting, PlayerInfo{
			AcID:          acID,
			WaitStartTime: time.Now().Unix(),
		})
		m.playerWaittingMap[acID] = true
	*/
	return m.wait.AddWaitter(acID, corpLv)
}

type gVEMag struct {
	AcID     string
	IsHard   bool
	IsCancel bool
	CorpLv   uint32
}

type GVEMatch struct {
	matchToken  string
	normalMatch Match
	highMatch   Match

	wait       sync.WaitGroup
	cmdChannel chan gVEMag
	quitChan   chan bool
}

func (g *GVEMatch) Start(matchToken string) {
	g.matchToken = matchToken

	g.normalMatch.Init(normalTimeoutCounter, 2048)
	g.highMatch.Init(hardTimeoutCounter, 2048)

	g.quitChan = make(chan bool, 1)
	g.cmdChannel = make(chan gVEMag, 4096)

	tickTime := time.Duration(multConfig.Cfg.MatchTimeoutSecond * 1000 / multConfig.Cfg.TickMax)

	g.wait.Add(1)
	go func() {
		defer g.wait.Done()
		timerChan := time.After(tickTime * time.Millisecond)
		timerMatchFreq := time.After(time.Second)
		for {
			select {
			case command, _ := <-g.cmdChannel:
				g.wait.Add(1)
				func() {
					defer g.wait.Done()

					if command.IsCancel {
						g.cancelPlayer(command.AcID, command.IsHard)
					} else {
						g.addPlayer(command.AcID,
							command.CorpLv,
							command.IsHard)
						g.matchNew()
					}
				}()
			case <-timerMatchFreq:
				g.wait.Add(1)
				func() {
					defer g.wait.Done()
					g.match()
				}()
				timerMatchFreq = time.After(time.Second)
			case <-timerChan:
				g.wait.Add(1)
				func() {
					defer g.wait.Done()
					g.normalMatch.wait.OnTick()
					g.highMatch.wait.OnTick()
					normalTickCounter.Inc(1)
				}()
				timerChan = time.After(tickTime * time.Millisecond)
			case <-g.quitChan:
				return
			}
		}
	}()
}

func (g *GVEMatch) cancelPlayer(acID string, isHard bool) {
	if isHard {
		g.highMatch.CancelWaiting(acID)
	} else {
		g.normalMatch.CancelWaiting(acID)
	}
}

func (g *GVEMatch) addPlayer(acID string, corpLv uint32, isHard bool) {
	if isHard {
		if g.highMatch.AddWaittingPlayer(acID, corpLv) {
			hardStartMatchCounter.Inc(1)
		}
	} else {
		if g.normalMatch.AddWaittingPlayer(acID, corpLv) {
			normalStartMatchCounter.Inc(1)
		}
	}
}

func (g *GVEMatch) match() {
	mh := g.highMatch.MatchAllGame()
	mn := g.normalMatch.MatchAllGame()
	if len(mn) > 0 || len(mh) > 0 {
		logs.Info("Match %s normal %v", g.matchToken, mn)
		logs.Info("Match %s hard   %v", g.matchToken, mh)
	}
	normalMatchedCounter.Inc(int64(len(mn) * MatchPlayerNum))
	hardMatchedCounter.Inc(int64(len(mh) * MatchPlayerNum))
	for _, m := range mn {
		notify.GetNotify(g.matchToken).Notify(helper.MatchGameInfo{
			AcIDs:  m.GetResult(),
			IsHard: false,
		})
	}
	for _, m := range mh {
		notify.GetNotify(g.matchToken).Notify(helper.MatchGameInfo{
			AcIDs:  m.GetResult(),
			IsHard: true,
		})
	}
}

func (g *GVEMatch) matchNew() {
	mh := g.highMatch.MatchNewGame()
	mn := g.normalMatch.MatchNewGame()
	if len(mn) > 0 || len(mh) > 0 {
		logs.Info("MatchNew %s normal %v", g.matchToken, mn)
		logs.Info("MatchNew %s hard   %v", g.matchToken, mh)
	}
	normalMatchedCounter.Inc(int64(len(mn) * MatchPlayerNum))
	hardMatchedCounter.Inc(int64(len(mh) * MatchPlayerNum))
	for _, m := range mn {
		notify.GetNotify(g.matchToken).Notify(helper.MatchGameInfo{
			AcIDs:  m.GetResult(),
			IsHard: false,
		})
	}
	for _, m := range mh {
		notify.GetNotify(g.matchToken).Notify(helper.MatchGameInfo{
			AcIDs:  m.GetResult(),
			IsHard: true,
		})
	}
}

func (g *GVEMatch) AddPlayer(acID string, corpLv uint32, isHard bool, isCancel bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case g.cmdChannel <- gVEMag{
		AcID:     acID,
		IsHard:   isHard,
		IsCancel: isCancel,
		CorpLv:   corpLv,
	}:
	case <-ctx.Done():
		return ErrTimeout
	}

	return nil
}

func (g *GVEMatch) Stop() {
	g.quitChan <- true
	g.wait.Wait()
}

var gveMatch GVEMatch

func GetGVEMatch() *GVEMatch {
	return &gveMatch
}
