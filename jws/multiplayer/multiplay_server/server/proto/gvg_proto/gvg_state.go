package gvg_proto

import (
	"time"

	"vcs.taiyouxi.net/jws/multiplayer/helper"
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

const (
	WaitPlayerReadyTime      = 10
	WaitLostPlayerOnlineTime = 15 //30 seconds仍然未能进入ready状态,则认为这个玩家掉线,其他人开始战斗
)

type GVGPlayerState struct {
	AcID            string
	AvatarID        int32
	State           int
	Hp              map[int32]int32
	Pos             int
	LastHPDeltaTime int64
	LastPingTime    int64
	Latency         int32
	IsBackLeave     bool
	NotIdentical    bool
}

type GVGBossState struct {
	Hp     int
	Hatred []int
	Armor  int64
}

func (b *GVGBossState) Init(playerNum int) {
	b.Hatred = make([]int, playerNum, playerNum)
}

func (ps *GVGPlayerState) IsExit() bool {
	return ps.State == PlayerStateLost || ps.State == PlayerStateHasLeave || ps.State == PlayerStateKilled
}

func (b *GVGBossState) AddHatred(playerIdx int, addon int) {
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

func (b *GVGBossState) GetHatredMax() int {
	// 玩家并不多(3个)
	res := 0
	for i := 1; i < len(b.Hatred); i++ {
		if b.Hatred[i] > b.Hatred[res] {
			res = i
		}
	}
	return res
}

type GVGGameState struct {
	Player              []GVGPlayerState
	Boss                []GVGBossState
	EnemyWaveHP         [][]uint32
	State               int
	LastDamageType      int32
	ToAutoNextStateTime int64
	StartTime           int64
	EndTime             int64
	IsHard              bool
	GameClass           int
	GameScene           string
	RandomNum           int64
	Rng                 util.Kiss64Rng
	IsSuccess           bool
	Winner              string
	Param               map[string]string
}

//Init GVGGameState 初始化函数
// isHard 是否是困难模式
// st, et 开始时间st，结束时间et
// players accountid的列表
// bossSize 关卡中boss的数量
func (g *GVGGameState) Init(data *helper.GVGStartFightData, st, et int64, bossSize int, randNum int64) {
	g.State = GameStateWaitReady
	g.ToAutoNextStateTime = time.Now().Unix() + WaitLostPlayerOnlineTime

	g.RandomNum = randNum
	g.Player = make([]GVGPlayerState, 0, 2)
	p1 := GVGPlayerState{
		AcID:  data.Acid1,
		State: PlayerStateWaitReady,
		Hp:    make(map[int32]int32, 0),
	}
	for _, item := range data.Avatar1 {
		p1.Hp[int32(item)] = 1
	}
	g.Player = append(g.Player, p1)
	p2 := GVGPlayerState{
		AcID:  data.Acid2,
		State: PlayerStateWaitReady,
		Hp:    make(map[int32]int32, 0),
	}
	for _, item := range data.Avatar2 {
		p2.Hp[int32(item)] = 1
	}
	g.Player = append(g.Player, p2)

	g.Boss = make([]GVGBossState, bossSize, bossSize)
	g.Param = make(map[string]string, 0)
}

func (g *GVGGameState) InitPlayerHp(AcID string, hp int32, avatarID int32) {
	for i := 0; i < len(g.Player); i++ {
		if g.Player[i].AcID == AcID {
			g.Player[i].Hp[avatarID] = hp
		}
	}
}

func (g *GVGGameState) SetPlayerStat(AcID string, stat int) {
	logs.Info("set player: %v stat: %v", AcID, stat)
	for i := 0; i < len(g.Player); i++ {
		if g.Player[i].AcID == AcID {
			if g.Player[i].State == PlayerStateHasLeave && stat == PlayerStateLost {
				continue
			}
			g.Player[i].State = stat
		}
	}
}

func (g *GVGGameState) SetPlayerLatency(AcID string, latency int32) {
	for i := 0; i < len(g.Player); i++ {
		if g.Player[i].AcID == AcID {
			g.Player[i].Latency = latency
		}
	}
}

func (g *GVGGameState) GetPlayerIdx(AcID string) int {
	for i := 0; i < len(g.Player); i++ {
		if g.Player[i].AcID == AcID {
			return i
		}
	}
	return -1
}

func (g *GVGGameState) InitBossHp(idx, hp int) {
	if idx < 0 || idx >= len(g.Boss) {
		return
	}
	g.Boss[idx].Init(len(g.Player))
	g.Boss[idx].Hp = hp
}

//根据对应玩家的伤害值,计算仇恨
func (g *GVGGameState) AddHatred(bossIdx, playerIdx int, addon int) {
	if bossIdx < 0 || bossIdx >= len(g.Boss) {
		return
	}
	g.Boss[bossIdx].AddHatred(playerIdx, addon)
}

func (g *GVGGameState) GetHatredMax(bossIdx int) int {
	if bossIdx < 0 || bossIdx >= len(g.Boss) {
		return 0
	}
	return g.Boss[bossIdx].GetHatredMax()
}

func (g *GVGGameState) BossHpDeta(bossIdx int, hpDeta int) {
	if bossIdx < 0 || bossIdx >= len(g.Boss) {
		return
	}
	g.Boss[bossIdx].Hp += hpDeta
	if g.Boss[bossIdx].Hp <= 0 {
		g.Boss[bossIdx].Hp = 0
	}
}

func (g *GVGGameState) BossArmorDeta(bossIdx int, armorDeta int64) {
	if bossIdx < 0 || bossIdx >= len(g.Boss) {
		return
	}
	g.Boss[bossIdx].Armor += armorDeta
	if g.Boss[bossIdx].Armor <= 0 {
		g.Boss[bossIdx].Armor = 0
	}
}

func (g *GVGGameState) PlayerHpDeta(PlayerIdx int, hpDeta int, avatarID int32) {
	if PlayerIdx < 0 || PlayerIdx >= len(g.Player) {
		return
	}
	g.Player[PlayerIdx].Hp[avatarID] += int32(hpDeta)
	if g.Player[PlayerIdx].Hp[avatarID] <= 0 {
		g.Player[PlayerIdx].Hp[avatarID] = 0
	}
	allKill := true
	for _, v := range g.Player[PlayerIdx].Hp {
		if v > 0 {
			allKill = false
			break
		}
	}
	if allKill {
		g.Player[PlayerIdx].State = PlayerStateKilled
	}
}

func (g *GVGGameState) SetPlayerLeave(AcID string) {
	logs.Info("player leave: %v", AcID)
	g.SetPlayerStat(AcID, PlayerStateHasLeave)
}

func (g *GVGGameState) SetPlayerLost(AcID string) {
	logs.Info("player loss: %v", AcID)
	g.SetPlayerStat(AcID, PlayerStateLost)
}

func (g *GVGGameState) setAllPlayerStat(from, to int) {
	logs.Info("set all player stat for: %v to %v", from, to)
	for i := 0; i < len(g.Player); i++ {
		if g.Player[i].State == from {
			g.Player[i].State = to
		}
	}
}

func (g *GVGGameState) SomeOneKilled() string {
	for _, item := range g.Player {
		hp := int32(0)
		for _, v := range item.Hp {
			hp += v
		}
		if hp <= 0 || item.State == PlayerStateLost || item.State == PlayerStateHasLeave {
			return item.AcID
		}
	}
	return ""
}

func (g *GVGGameState) isAllPlayer(c func(p *GVGPlayerState) bool) bool {
	isAll := true
	for i := 0; i < len(g.Player); i++ {
		isAll = isAll && c(&g.Player[i])
	}
	return isAll
}

func (g *GVGGameState) UpdateGameState() bool {
	nowT := time.Now().Unix()
	isStateChange := false
	if g.State == GameStateWaitReady {
		// 1. 看看在线的玩家是不是都Ready了
		isAllReady := g.isAllPlayer(func(p *GVGPlayerState) bool {
			return p.State == PlayerStateReady ||
				p.State == PlayerStateLost ||
				p.State == PlayerStateHasLeave
		})
		if isAllReady {
			// 1.1 全都Ready 则进入等待掉线玩家上线状态
			g.State = GameStateWaitOnline
			g.ToAutoNextStateTime = nowT + WaitLostPlayerOnlineTime
			isStateChange = true
		}

		if nowT >= g.ToAutoNextStateTime {
			// 1.2 超时了 有些玩家就是不ready 视为掉线 进入等待掉线玩家上线状态
			g.State = GameStateWaitOnline
			g.ToAutoNextStateTime = nowT + WaitLostPlayerOnlineTime
			g.setAllPlayerStat(PlayerStateWaitReady, PlayerStateLost)
			isStateChange = true
		}
	} else if g.State == GameStateWaitOnline {
		// 2. 等待掉线玩家上线状态 最后再等等 能上就上
		isAllOnline := g.isAllPlayer(func(p *GVGPlayerState) bool {
			return p.State == PlayerStateReady
		})
		if isAllOnline || nowT >= g.ToAutoNextStateTime {
			// 2.1 全在线或者超时 不等了 开启游戏 进入游戏中状态
			logs.Info("begin game fighting")
			// TODO by ljz  read xlsx conf
			fightTime := int64(180)
			g.StartTime = nowT
			g.EndTime = nowT + fightTime
			g.State = GameStateFighting
			g.setAllPlayerStat(PlayerStateReady, PlayerStateFighting)
			isStateChange = true
		}
	} else if g.State == GameStateFighting {
		if player := g.SomeOneKilled(); player != "" {
			// 3.1 玩家全死了 输了
			logs.Info("PlayerKilled: %v, info: %v", player, *g)
			g.setAllPlayerStat(PlayerStateFighting, PlayerStateOver)
			g.JudgeWinner()
			g.State = GameStateOver
			isStateChange = true
		}
	}

	return isStateChange
}

func (r *GVGGameState) JudgeWinner() {
	winner := ""
	maxHP := int32(0)
	for _, item := range r.Player {
		hp := int32(0)
		for _, v := range item.Hp {
			hp += v
		}
		if hp > maxHP {
			maxHP = hp
			winner = item.AcID
		}
	}
	if winner == "" {
		logs.Warn("There are no available winner for info: %v", r)
		if len(r.Player) > 0 {
			winner = r.Player[0].AcID
		}
	}
	r.Winner = winner
}
