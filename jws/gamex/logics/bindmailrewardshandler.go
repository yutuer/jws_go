package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// BindMailRewards : 客户端绑定邮箱可以发奖
//
func (p *Account) BindMailRewardsHandler(req *reqMsgBindMailRewards, resp *rspMsgBindMailRewards) uint32 {
	info := p.Profile.GetBindMailRewardInfo()
	switch req.ActivityRewardID {
	case gamedata.ActSpecRewardMailReward:
		if info.BindMailRewardGet {
			return errCode.CommonConditionFalse
		}
	case gamedata.ActSpecRewardEGReward:
		if info.BindEGRewardGet {
			return errCode.CommonConditionFalse
		}
	default:
		logs.Error("no activity id: %d", req.ActivityRewardID)
		return 0
	}
	reward := gamedata.GetActivitySpecRewards(int(req.ActivityRewardID))
	if !account.GiveBySync(p.Account, &reward.Cost, resp, "bind eg mail reward") {
		return errCode.RewardFail
	}
	switch req.ActivityRewardID {
	case gamedata.ActSpecRewardMailReward:
		info.BindMailRewardGet = true
	case gamedata.ActSpecRewardEGReward:
		info.BindEGRewardGet = true
	}
	resp.OnChangeBindMailReward()
	return 0
}
