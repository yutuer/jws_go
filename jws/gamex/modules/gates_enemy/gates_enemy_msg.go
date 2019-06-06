package gates_enemy

import (
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (g *GatesEnemyActivity) SendDataToPlayers() {
	currAccounts := make([]string, 0, len(g.members))
	for _, m := range g.members {
		currAccounts = append(currAccounts, m.AccountID)
	}

	msg := g.mkDataToPlayers()

	player_msg.SendToPlayers(currAccounts,
		player_msg.PlayerMsgGatesEnemyDataCode,
		*(msg.ToClient()))

	logs.Debug("GatesEnemyActivity SendDataToPlayers")
}

func (g *GatesEnemyActivity) mkDataToPlayers() *player_msg.GatesEnemyData {
	msg := &player_msg.GatesEnemyData{
		EnemyInfo:     g.enemyInfo,
		State:         g.state,
		StateOverTime: g.stateOverTime,
		KillPoint:     g.killPoint,
		BossMax:       g.bossMax,
		Point:         g.gePointAll,
		BuffCurLv:     g.buffCurLv,
		BuffMemAcid:   g.buffMemAcid,
		BuffMemName:   g.buffMemName,
	}

	msg.PlayerRank.Init()

	for i := 0; i < len(g.members); i++ {
		if g.members[i].AccountID != "" {
			m := g.members[i]
			var p int
			var stat int
			fd, ok := g.memberData[m.AccountID]
			if ok {
				p = fd.currGEActivityPoint
				stat = fd.GetState()
			}
			m.Other.Pi[0] = 0
			m.Other.Pi[1] = int64(stat)
			//logs.Warn("PlayerRank.OnPlayerSorce %v", p)
			msg.PlayerRank.OnPlayerSorceNoSort(&m, int64(p))
		}
	}
	msg.PlayerRank.Sort()
	return msg
}
