package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logics/notify"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
)

func (s *SyncResp) mkGiftInfo(p *Account) {
	if s.SyncGiftMonthlyNeed {
		now_time := p.Profile.GetProfileNowTime()
		dayInMonth := gamedata.GetNowMonthlyGiftDayth(now_time)
		gift_monthly := p.Profile.GetGiftMonthly()
		gift_monthly.Update(now_time)
		s.SyncGiftMonthlyId = gift_monthly.Curr_activity_id
		s.SyncGiftMonthlyCurrIdx = gift_monthly.Curr_gift_idx
		s.SyncGiftMonthlyCurrStat = int(gift_monthly.Curr_gift_stat)
		s.SyncGiftLeftReSignCount = dayInMonth - s.SyncGiftMonthlyCurrIdx - 1
	}

	if s.SyncGiftsNeed {
		now_time := p.Profile.GetProfileNowTime()

		gifts := p.Profile.GetGifts()
		gifts.Update(now_time)
		s.SyncGifts = make([][]byte, 0, len(gifts.Gifts))
		leftTime2Update := util.GetNextDailyTime(
			gamedata.GetCommonDayBeginSec(now_time), now_time)

		for i := 0; i < len(gifts.Gifts); i++ {
			flag, count := gifts.Gifts[i].IsHasReward()
			if flag {
				// 这个判断是为了最后一次领奖后，立刻返回空，即功能关闭
				if gifts.Gifts[i].CurrGiftIdx == count-1 && gifts.Gifts[i].CurrGiftStat > 0 {
					continue
				}
				leftTime := leftTime2Update
				if gifts.Gifts[i].StartActivityTime <= 0 {
					leftTime = 0
				}

				allRewards := gifts.Gifts[i].GetAllGiftData(p.Profile.GetVipLevel())

				giftInfo := activityGiftInfo2Client{
					GiftId:              gifts.Gifts[i].CurrActivityId,
					GiftActivityTime:    gifts.Gifts[i].StartActivityTime,
					GiftCurrIdx:         gifts.Gifts[i].CurrGiftIdx,
					GiftCurrStat:        int(gifts.Gifts[i].CurrGiftStat),
					GiftNextUpdateSec:   leftTime,
					GiftAllRewardStat:   make([]int, 0, len(gifts.Gifts[i].AllGiftStats)),
					GiftAllRewardCount:  make([]int, 0, len(allRewards)),
					GiftAllRewardAID:    make([]string, 0, len(allRewards)*3),
					GiftAllRewardACount: make([]uint32, 0, len(allRewards)*3),
					GiftAllRewardAData:  make([]string, 0, len(allRewards)*3),
				}

				for _, s := range gifts.Gifts[i].AllGiftStats {
					giftInfo.GiftAllRewardStat = append(giftInfo.GiftAllRewardStat, int(s))
				}
				for _, reward := range allRewards {
					giftInfo.GiftAllRewardCount = append(giftInfo.GiftAllRewardCount, len(reward.Item2Client))
					giftInfo.GiftAllRewardAID = append(giftInfo.GiftAllRewardAID, reward.Item2Client...)
					giftInfo.GiftAllRewardACount = append(giftInfo.GiftAllRewardACount, reward.Count2Client...)
					giftInfo.GiftAllRewardAData = append(giftInfo.GiftAllRewardAData, reward.Data2Client...)
				}

				// TODO delete begin
				_, giftReward := gifts.Gifts[i].GetGiftToGet(
					p.Profile.GetVipLevel())
				if giftReward != nil {
					giftInfo.GiftRewardID = giftReward.Item2Client
					giftInfo.GiftRewardCount = giftReward.Count2Client
					giftInfo.GiftRewardData = giftReward.Data2Client
				}
				// TODO delete end
				s.SyncGifts = append(s.SyncGifts, encode(giftInfo))
			}
		}
	}

	if s.act_gift_by_cond_need_sync {
		s.SyncActGiftByCond = p.Account.Profile.GetActGiftByCond().GetAllInfo(p.Account)
		s.SyncActGiftByCondCount = len(s.SyncActGiftByCond)
	}

	if s.act_gift_by_time_need_sync {
		s.SyncActGiftByTime = p.Account.Profile.GetActGiftByTime().GetAllInfo(p.Account)
		s.SyncActGiftByTimeCount = 1
	}

	if p.Account.Profile.GetActGiftByCond().RefreshRedPoint(p.Account) ||
		p.Account.Profile.GetActGiftByTime().RefreshRedPoint(p.Account) {
		s.OnChangeRedPoint(notify.RedPointTyp_CondActGift)
	}
}
