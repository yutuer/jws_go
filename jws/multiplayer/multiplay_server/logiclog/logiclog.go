package logiclog

import (
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	LogicTag_GameOver           = "PveBossGameOver"
	LogicTag_GameOverWithPlayer = "PveBossGameOverWithPlayer"

	LogicTag_TBossBattleStart = "TBossBattleStart"
	LogicTag_TBossBattleEnd   = "TBossBattleEnd"

	BITag = "[BI]"
)

type LogicInfo_Player struct {
	PlayerId string
	Avatar   int
	Gs       int
	Hp       int
	HpOrg    int
	HpRate   float32
	Lvl      uint32
	IsUseHc  bool
	IsDouble bool
}

type LogicInfo_GameOver struct {
	GameId       string
	Duration     int64
	IsWin        int
	IsHard       bool
	BossId       []string
	PlayerLvlAvg uint32
}

type LogicInfo_GameOverWithPlayer struct {
	GameId       string
	Duration     int64
	IsWin        int
	IsHard       bool
	BossId       []string
	BossHp       []int
	BossOrgHp    []int
	BossHpRate   []float32
	PlayerLvlAvg uint32
	Player1      LogicInfo_Player
	Player2      LogicInfo_Player
	Player3      LogicInfo_Player
}

type LogicInfo_TBossBattleStart struct {
	AccounID      []string
	AvatarID      []int
	BattleID      string
	CompressGS    []int
	BossID        string
	VIP           []uint32
	IsTickRedBox  bool
	WhoTickRedBox string
}

type LogicInfo_TBossBattleEnd struct {
	AccounID     []string
	AvatarID     []int
	BattleID     string
	CompressGS   []int
	BossID       string
	VIP          []uint32
	BattleTime   []int64
	PlayerLeftHP []int
	BossLeftHP   int
	IsWin        bool
	PlayerState  []int
	IsBackLeave  []bool
	Difficulty   uint32
}

func GameOver(gameId string, duration int64, isWin bool, isHard bool,
	bossId []string, playerLvlAvg uint32) {

	format := BITag
	logs.Trace("GameOver %s", format)

	var win int
	if !isWin {
		win = 1
	}
	r := LogicInfo_GameOver{
		GameId:       gameId,
		Duration:     duration,
		IsWin:        win,
		IsHard:       isHard,
		BossId:       bossId,
		PlayerLvlAvg: playerLvlAvg,
	}
	TypeInfo := LogicTag_GameOver
	logiclog.MultiInfo(TypeInfo, r, format)
}

func GameOverWithPlayer(gameId string, duration int64, isWin bool, isHard bool,
	bossId []string, bossHp []int, bossOrgHp []int, bossHpRate []float32, playerLvlAvg uint32, player1, player2, player3 LogicInfo_Player) {

	format := BITag
	logs.Trace("GameOverWithPlayer %s", format)

	var win int
	if !isWin {
		win = 1
	}
	r := LogicInfo_GameOverWithPlayer{
		GameId:       gameId,
		Duration:     duration,
		IsWin:        win,
		IsHard:       isHard,
		BossId:       bossId,
		BossHp:       bossHp,
		BossOrgHp:    bossOrgHp,
		BossHpRate:   bossHpRate,
		PlayerLvlAvg: playerLvlAvg,
		Player1:      player1,
		Player2:      player2,
		Player3:      player3,
	}
	TypeInfo := LogicTag_GameOverWithPlayer
	logiclog.MultiInfo(TypeInfo, r, format)
}
