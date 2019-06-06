package gve_notify

import (
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (m *module) onGVEStart(c *cmd) {
	k := c.AccountID + c.GameID + c.GameServerUrl
	m.dataWaitting[k] = c.ResChan
	player_msg.Send(c.AccountID, player_msg.PlayerMsgGVEGameStartCode,
		player_msg.PlayerMsgGVEGameStart{
			GameID:        c.GameID,
			GameSecret:    c.GameSecret,
			GameServerUrl: c.GameServerUrl,
			IsBot:         c.IsBot,
		})
}

func (m *module) onGVEStop(c *cmd) {
	player_msg.Send(c.AccountID, player_msg.PlayerMsgGVEGameStopCode,
		player_msg.PlayerMsgGVEGameStop{
			GameID:      c.GameID,
			IsHasReward: c.IsHasReward,
			IsSuccess:   c.IsSuccess,
		})
}

func (m *module) onAccountData(c *cmd) {
	k := c.AccountID + c.GameID + c.GameServerUrl
	waitting, ok := m.dataWaitting[k]
	if !ok {
		logs.Trace("dataWaitting no find %v", *c)
		return
	}

	waitting <- *c
	delete(m.dataWaitting, k)
	return
}
