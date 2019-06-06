package guild

import (
	"time"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/base"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/info"
)

type guildCommandResWithActDatas struct {
	ActBoss *info.ActBoss2Client
}

func (g *guildCommandResWithActDatas) OnActRes(gi *GuildInfo, res *base.ActCommand) {
	nowT := time.Now().Unix()
	if res.NeedPushs[base.GuildActBoss] {
		g.ActBoss = gi.ActBoss.ToClient(nowT)
	}
}
