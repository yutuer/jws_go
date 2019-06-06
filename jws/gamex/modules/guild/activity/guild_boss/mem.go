package guild_boss

import "vcs.taiyouxi.net/jws/gamex/modules/guild/activity/base"

func (a *ActivityState) OnMemKick(acid string) {
	a.BigBoss.MVPRank.OnPlayerDel(acid)
	a.BigBoss.OnMemberKick(acid)
	for i := 0; i < len(a.Bosses); i++ {
		a.Bosses[i].MVPRank.OnPlayerDel(acid)
		a.Bosses[i].OnMemberKick(acid)
	}
	a.TodayDamages.OnPlayerDel(acid)
	a.OnMemKickedByPlayerActStat(acid)
	a.GetGuildHandler().NotifyAll(base.GuildActBoss)
	a.GetGuildHandler().SetNeedSave2DB()
}
