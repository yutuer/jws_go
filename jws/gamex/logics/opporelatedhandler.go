package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
)

// OppoSign : oppo签到
//
func (p *Account) OppoSignHandler(req *reqMsgOppoSign, resp *rspMsgOppoSign) uint32 {
	nowT := p.Profile.GetProfileNowTime()
	oppo := p.Profile.GetOppoRelated()
	oppo.UpdateTime(nowT, true)
	if gamedata.IsSameDayCommon(nowT, oppo.LastSignTime) {
		return errCode.OppoSignError
	}
	oppo.LastSignTime = nowT
	if rewardData := gamedata.GetOPPOSignData(oppo.SignDays); rewardData != nil {
		giveData := gamedata.CostData{}
		for _, item := range rewardData {
			giveData.AddItem(item.GetGuildItemID(), item.GetGuildItemNum())
		}
		if !account.GiveBySync(p.Account, &giveData, resp, "OPPOSign") {
			return errCode.RewardFail
		}
	} else {
		return errCode.OppoSignError
	}
	resp.OnChangeOppoRelated()
	return 0
}

// OppoDailyQuest : oppo每日任务
//
func (p *Account) OppoDailyQuestHandler(req *reqMsgOppoDailyQuest, resp *rspMsgOppoDailyQuest) uint32 {
	oppo := p.Profile.GetOppoRelated()
	nowT := p.Profile.GetProfileNowTime()
	if gamedata.IsSameDayCommon(nowT, oppo.LastDailyQuestTime) {
		return errCode.OppoDailyQuestError
	}
	oppo.LastDailyQuestTime = nowT
	if rewardData := gamedata.GetOPPODailyQuestData(); rewardData != nil {
		giveData := gamedata.CostData{}
		for _, item := range rewardData {
			giveData.AddItem(item.GetGuildItemID(), item.GetGuildItemNum())
		}
		if !account.GiveBySync(p.Account, &giveData, resp, "OPPODailyQuest") {
			return errCode.RewardFail
		}
	} else {
		return errCode.OppoDailyQuestError
	}
	resp.OnChangeOppoRelated()
	return 0
}

// OppoDailyQuest : oppo每日任务
//
func (p *Account) OppoLoginHandler(req *reqMsgOppoLogin, resp *rspMsgOppoLogin) uint32 {
	oppo := p.Profile.GetOppoRelated()
	nowT := p.Profile.GetProfileNowTime()
	oppo.UpdateTime(nowT, req.IsOppo)
	resp.OnChangeOppoRelated()
	return 0
}
