package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) updateCondition(ctyp, p1, p2 int, p3, p4 string, resp helper.ISyncRsp) {
	logs.Trace("[%s]updateCondition %d : %d %d %s %s", p.AccountID, ctyp, p1, p2, p3, p4)

	isUpdated := p.Profile.GetQuest().DailyTaskReset(p.Account)
	isUpdated = isUpdated || p.Profile.GetQuest().UpdateCanReceiveList(p.Account)

	player_cond := p.Profile.GetQuest().GetReceivedQuest()
	// TBD 按照ctyp分类更新，减少更新时遍历的数量
	for i := 0; i < len(player_cond); i++ {
		c := &player_cond[i].Condition
		if c.Ctyp == uint32(ctyp) {
			account.UpdateCondition(c, int64(p1), int64(p2), p3, p4)
			isUpdated = true
		}
	}

	if isUpdated {
		// TBD 差量同步
		p.Profile.GetQuest().SetNeedSync()
	}
}

func (p *Account) questCurTimeRedPoint() bool {
	received := p.Profile.GetQuest().GetReceivedQuest()
	for _, quest := range received {
		if !quest.IsVailed() && quest.Condition.Ctyp == account.COND_TYP_Curr_Time {
			progress, all := quest.GetProgress(p.Account)
			if progress >= all {
				return true
			}
		}
	}
	return false
}
