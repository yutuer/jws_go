package guild

import (
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/base"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *GuildInfo) notifyAllExp(types [base.GuildActCount]bool, nowTime int64) {
	acids := p.GetAllMemberAcids()
	if types[base.GuildActBoss] {
		syncGuildBoss2Players(acids, p, nowTime)
	}
}

func (p *GuildInfo) NotifyAll(typ int) {
	logs.Trace("notify all %v", typ)
	p.needNotifySomething = true
	p.needNotifyAct[typ] = true
}
