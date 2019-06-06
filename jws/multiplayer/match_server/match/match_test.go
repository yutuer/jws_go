package match

import (
	"testing"

	"vcs.taiyouxi.net/platform/planx/util/config"
	multConfig "vcs.taiyouxi.net/jws/multiplayer/match_server/config"
)

/*
type Match struct {
	playerWaittingTail    []PlayerInfo
	playerWaittingTop     []PlayerInfo
	playerWaittingTopIdx  int
	playerWaittingTailIdx int
	playerWaittingNum     int
}

func (m *Match) Enqueue(p PlayerInfo) {
	if m.playerWaittingTailIdx >= -1 {
		m.playerWaittingTail = append(m.playerWaittingTail, p)
	} else {
		topIdx := -(m.playerWaittingTailIdx + 1)
		m.playerWaittingTop[topIdx-1] = p
	}
	m.playerWaittingTailIdx++
	m.playerWaittingNum++
	logs.Trace("top %v", m.playerWaittingTop)
	logs.Trace("tai %v", m.playerWaittingTail)
}

func (m *Match) Dequeue() (PlayerInfo, bool) {
	res := PlayerInfo{}
	if m.playerWaittingTopIdx > 0 {
		res = m.playerWaittingTop[m.playerWaittingTopIdx]
		m.playerWaittingTop[m.playerWaittingTopIdx] = PlayerInfo{}
		m.playerWaittingTopIdx--
		m.playerWaittingNum--
		return res, true
	} else if m.playerWaittingTopIdx == 0 {
		if len(m.playerWaittingTail) > 0 {
			res = m.playerWaittingTail[0]
			m.playerWaittingTail[0] = PlayerInfo{}
			m.playerWaittingTopIdx = -2
			m.playerWaittingNum--
			return res, true
		} else {
			return res, false
		}
	} else {
		tailIdx := -(1 + m.playerWaittingTopIdx)
		if len(m.playerWaittingTail) > tailIdx {
			res = m.playerWaittingTail[tailIdx]
			m.playerWaittingTail[tailIdx] = PlayerInfo{}
			m.playerWaittingTopIdx = -tailIdx - 2
			m.playerWaittingNum--
			return res, true
		} else {
			return res, false
		}
	}
	logs.Trace("top %v", m.playerWaittingTop)
	logs.Trace("tai %v", m.playerWaittingTail)
	return res, false
}
*/

// 需要读config文件才能进行测试，手动指定config文件
func init() {
	configFileName := "src/vcs.taiyouxi.net/jws/multiplayer/match_server/conf/config.toml"

	var common_cfg struct{ CommonCfg multConfig.CommonConfig }
	config.DebugLoadConfigToml(configFileName, &common_cfg)
	multConfig.Cfg = common_cfg.CommonCfg
}

func show(t *testing.T, m *Match, res []MatchResult) {
	if res != nil {
		//t.Logf("Match %v", m.playerWaitting)
		t.Logf("MatchResult %v", res)
	} else {
		//t.Logf("Match %v", m.playerWaitting)
		t.Logf("MatchResult nil")
	}
}

func TestMatch(t *testing.T) {
	m := &Match{}
	m.Init(normalStartMatchCounter, 64)
	show(t, m, m.MatchAllGame())
}

func TestMatch1(t *testing.T) {
	m := &Match{}
	m.Init(normalStartMatchCounter, 64)
	m.AddWaittingPlayer("1", 2)
	show(t, m, m.MatchAllGame())
}

func TestMatch2(t *testing.T) {
	m := &Match{}
	m.Init(normalStartMatchCounter, 64)
	m.AddWaittingPlayer("1", 2)
	m.AddWaittingPlayer("2", 2)
	show(t, m, m.MatchAllGame())
}

func TestMatch3(t *testing.T) {
	m := &Match{}
	m.Init(normalStartMatchCounter, 64)
	m.AddWaittingPlayer("1", 2)
	m.AddWaittingPlayer("2", 2)
	m.AddWaittingPlayer("3", 2)
	m.AddWaittingPlayer("4", 2)
	show(t, m, m.MatchAllGame())
}

func TestMatch4(t *testing.T) {
	m := &Match{}
	m.Init(normalStartMatchCounter, 64)
	m.AddWaittingPlayer("1", 2)
	m.AddWaittingPlayer("2", 2)
	m.AddWaittingPlayer("3", 2)
	m.AddWaittingPlayer("4", 2)
	m.AddWaittingPlayer("5", 2)
	m.AddWaittingPlayer("6", 2)
	m.AddWaittingPlayer("7", 2)
	show(t, m, m.MatchAllGame())
}

func TestMatch5(t *testing.T) {
	m := &Match{}
	m.Init(normalStartMatchCounter, 64)
	m.AddWaittingPlayer("1", 2)
	m.AddWaittingPlayer("2", 2)
	m.AddWaittingPlayer("3", 2)
	m.AddWaittingPlayer("4", 2)
	m.AddWaittingPlayer("5", 2)
	m.AddWaittingPlayer("6", 2)
	m.AddWaittingPlayer("7", 2)
	m.AddWaittingPlayer("8", 2)
	show(t, m, m.MatchAllGame())
}

func TestMatch6(t *testing.T) {
	m := &Match{}
	m.Init(normalStartMatchCounter, 64)
	m.AddWaittingPlayer("1", 2)
	m.AddWaittingPlayer("2", 2)
	m.AddWaittingPlayer("3", 3)
	m.AddWaittingPlayer("4", 4)
	m.AddWaittingPlayer("5", 5)
	m.AddWaittingPlayer("6", 6)
	m.AddWaittingPlayer("7", 77)
	m.AddWaittingPlayer("8", 12)
	m.AddWaittingPlayer("9", 32)
	show(t, m, m.MatchAllGame())
}
