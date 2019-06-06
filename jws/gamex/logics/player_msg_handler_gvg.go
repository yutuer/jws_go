package logics

import (
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) OnPlayerMsgGVGStart(r servers.Request) *servers.Response {
	req := player_msg.PlayerMsgGVGStart{}

	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerMsgGVGStart %s %v", p.AccountID.String(), req)
	p.Tmp.SetGVGData(req.PlayerInfo, req.DestinySkill, req.EnemyAcID, req.EnemyPlayerInfo, req.EDestinySkill,
		req.EnemyData, req.RoomID, req.URL)
	return nil
}
