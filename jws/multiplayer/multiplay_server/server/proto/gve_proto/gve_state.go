package gve_proto

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 玩家状态 1-掉线 2-已退出 3-未准备 4-已准备 5-已死亡 6-战斗中
const (
	PlayerStateNull = iota
	PlayerStateLost
	PlayerStateHasLeave
	PlayerStateWaitReady
	PlayerStateReady
	PlayerStateKilled
	PlayerStateFighting
	PlayerStateOver
)

// 游戏状态 1 准备开始 2 等待准备开始的玩家在线 3 正在战斗 4 已结束
const (
	GameStateNull = iota
	GameStateWaitReady
	GameStateWaitOnline
	GameStateFighting
	GameStateOver
)

type GVEPlayerState struct {
	AcID  string
	State int
	Hp    int
}

type GVEBossState struct {
	Hp     int
	Hatred []int
}

func (b *GVEBossState) Init(playerNum int) {
	b.Hatred = make([]int, playerNum, playerNum)
}

func (b *GVEBossState) AddHatred(playerIdx int, addon int) {
	if playerIdx < 0 || playerIdx >= len(b.Hatred) {
		return
	}
	if addon < 0 {
		addon = -addon
	} else {
		addon = 0
	}

	b.Hatred[playerIdx] += addon
}

func (b *GVEBossState) GetHatredMax() int {
	// 玩家并不多(3个)
	res := 0
	for i := 1; i < len(b.Hatred); i++ {
		if b.Hatred[i] > b.Hatred[res] {
			res = i
		}
	}
	return res
}

type GVEGameState struct {
	Player              []GVEPlayerState
	Boss                []GVEBossState
	State               int
	ToAutoNextStateTime int64
	StartTime           int64
	EndTime             int64
	IsHard              bool
	GameClass           int
	GameScene           string
	RandomNum           int64
	Rng                 util.Kiss64Rng
	IsSuccess           bool
}

//Init GVEGameState 初始化函数
// isHard 是否是困难模式
// st, et 开始时间st，结束时间et
// players accountid的列表
// bossSize 关卡中boss的数量
func (g *GVEGameState) Init(isHard bool, st, et int64, players []string, bossSize int, randNum int64) {
	g.State = GameStateWaitReady
	g.ToAutoNextStateTime = time.Now().Unix() + int64(gamedata.BoxCfg.MaxWaitTime)
	g.StartTime = st
	g.EndTime = et
	g.IsHard = isHard
	g.RandomNum = randNum
	g.Player = make([]GVEPlayerState, 0, len(players))
	for i := 0; i < len(players); i++ {
		g.Player = append(g.Player, GVEPlayerState{
			AcID:  players[i],
			State: PlayerStateWaitReady,
			Hp:    0,
		})
	}
	g.Boss = make([]GVEBossState, bossSize, bossSize)
}

func (g *GVEGameState) InitPlayerHp(AcID string, hp int) {
	for i := 0; i < len(g.Player); i++ {
		if g.Player[i].AcID == AcID {
			g.Player[i].Hp = hp
		}
	}
}

func (g *GVEGameState) SetPlayerStat(AcID string, stat int) {
	for i := 0; i < len(g.Player); i++ {
		if g.Player[i].AcID == AcID {
			g.Player[i].State = stat
		}
	}
}

func (g *GVEGameState) GetPlayerIdx(AcID string) int {
	for i := 0; i < len(g.Player); i++ {
		if g.Player[i].AcID == AcID {
			return i
		}
	}
	return -1
}

func (g *GVEGameState) InitBossHp(idx, hp int) {
	if idx < 0 || idx >= len(g.Boss) {
		return
	}
	g.Boss[idx].Init(len(g.Player))
	g.Boss[idx].Hp = hp
}

//根据对应玩家的伤害值,计算仇恨
func (g *GVEGameState) AddHatred(bossIdx, playerIdx int, addon int) {
	if bossIdx < 0 || bossIdx >= len(g.Boss) {
		return
	}
	g.Boss[bossIdx].AddHatred(playerIdx, addon)
}

func (g *GVEGameState) GetHatredMax(bossIdx int) int {
	if bossIdx < 0 || bossIdx >= len(g.Boss) {
		return 0
	}
	return g.Boss[bossIdx].GetHatredMax()
}

func (g *GVEGameState) BossHpDeta(bossIdx int, hpDeta int) {
	if bossIdx < 0 || bossIdx >= len(g.Boss) {
		return
	}
	g.Boss[bossIdx].Hp += hpDeta
	if g.Boss[bossIdx].Hp <= 0 {
		g.Boss[bossIdx].Hp = 0
	}
}

func (g *GVEGameState) PlayerHpDeta(PlayerIdx int, hpDeta int) {
	if PlayerIdx < 0 || PlayerIdx >= len(g.Player) {
		return
	}
	logs.Debug("player hp deta: %v, %v", PlayerIdx, g.Player[PlayerIdx])
	g.Player[PlayerIdx].Hp += hpDeta
	if g.Player[PlayerIdx].Hp <= 0 {
		g.Player[PlayerIdx].Hp = 0
		g.Player[PlayerIdx].State = PlayerStateKilled
	}

}

func (g *GVEGameState) SetPlayerLeave(AcID string) {
	g.SetPlayerStat(AcID, PlayerStateHasLeave)
}

func (g *GVEGameState) SetPlayerLost(AcID string) {
	g.SetPlayerStat(AcID, PlayerStateLost)
}

//func (g *GVEGameState) SetPlayerOnline(AcID string) {
//	toState := PlayerStateNull
//	switch g.State {
//	case GameStateWaitReady:
//		toState = PlayerStateWaitReady
//	case GameStateWaitOnline:
//		toState = PlayerStateReady
//	case GameStateFighting:
//		toState = PlayerStateFighting
//	}
//	if toState != PlayerStateNull {
//		g.SetPlayerStat(AcID, toState)
//	}
//}

func (g *GVEGameState) setAllPlayerStat(from, to int) {
	for i := 0; i < len(g.Player); i++ {
		if g.Player[i].State == from {
			g.Player[i].State = to
		}
	}
}

func (g *GVEGameState) isAllPlayer(c func(p *GVEPlayerState) bool) bool {
	isAll := true
	for i := 0; i < len(g.Player); i++ {
		isAll = isAll && c(&g.Player[i])
	}
	return isAll
}

func (g *GVEGameState) UpdateGameState() bool {
	nowT := time.Now().Unix()
	isStateChange := false
	if g.State == GameStateWaitReady {
		// 1. 看看在线的玩家是不是都Ready了
		isAllReady := g.isAllPlayer(func(p *GVEPlayerState) bool {
			return p.State == PlayerStateReady ||
				p.State == PlayerStateLost ||
				p.State == PlayerStateHasLeave
		})
		if isAllReady {
			// 1.1 全都Ready 则进入等待掉线玩家上线状态
			g.State = GameStateWaitOnline
			g.ToAutoNextStateTime = nowT + int64(gamedata.BoxCfg.MaxWaitTime)
			isStateChange = true
		}

		if nowT >= g.ToAutoNextStateTime {
			// 1.2 超时了 有些玩家就是不ready 视为掉线 进入等待掉线玩家上线状态
			g.State = GameStateWaitOnline
			g.ToAutoNextStateTime = nowT + int64(gamedata.BoxCfg.MaxWaitTime)
			g.setAllPlayerStat(PlayerStateWaitReady, PlayerStateLost)
			isStateChange = true
		}
	} else if g.State == GameStateWaitOnline {
		// 2. 等待掉线玩家上线状态 最后再等等 能上就上
		isAllOnline := g.isAllPlayer(func(p *GVEPlayerState) bool {
			return p.State == PlayerStateReady
		})
		if isAllOnline || nowT >= g.ToAutoNextStateTime {
			// 2.1 全在线或者超时 不等了 开启游戏 进入游戏中状态
			g.State = GameStateFighting
			g.setAllPlayerStat(PlayerStateReady, PlayerStateFighting)
			isStateChange = true
		}
	} else if g.State == GameStateFighting {
		// 3. 游戏中
		isAllPlayerKilled := g.isAllPlayer(func(p *GVEPlayerState) bool {
			return p.Hp <= 0 || p.State == PlayerStateHasLeave || p.State == PlayerStateLost
		})
		// ROBOT
		if len(g.Player) == 1 {
			isAllPlayerKilled = g.Player[0].State == PlayerStateHasLeave || g.Player[0].State == PlayerStateLost
		}
		if isAllPlayerKilled {
			// 3.1 玩家全死了 输了
			logs.Trace("isAllPlayerKilled %v", *g)
			g.setAllPlayerStat(PlayerStateFighting, PlayerStateOver)
			g.IsSuccess = false
			g.State = GameStateOver
			isStateChange = true
		}

		isAllBossKilled := true
		for i := 0; i < len(g.Boss); i++ {
			isAllBossKilled = isAllBossKilled && (g.Boss[i].Hp <= 0)
		}
		if isAllBossKilled {
			// 3.2 Boss全死了 赢了
			logs.Trace("isAllBossKilled %v", *g)
			g.setAllPlayerStat(PlayerStateFighting, PlayerStateOver)
			g.IsSuccess = true
			g.State = GameStateOver
			isStateChange = true
		}
	}

	return isStateChange
}
