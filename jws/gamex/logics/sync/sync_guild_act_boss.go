package syncData

import "vcs.taiyouxi.net/jws/gamex/modules/guild/activity/info"

type SyncGuildActBoss struct {
	info.ActBoss2Client
}

func (s *SyncGuildActBoss) OnChangeGuildActBoss(data *info.ActBoss2Client) {
	s.ActBoss2Client = *data
}
