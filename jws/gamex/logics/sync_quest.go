package logics

import (
	"vcs.taiyouxi.net/jws/gamex/logics/notify"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (s *SyncResp) mkQuestInfo(p *Account) {
	profile := &p.Profile
	acid := p.AccountID.String()
	vip_lv := profile.GetVipLevel()
	nowT := profile.GetProfileNowTime()

	if p.questCurTimeRedPoint() {
		s.OnChangeRedPoint(notify.RedPointTyp_Quest)
	}

	playerQuest := p.Profile.GetQuest()

	if s.quest_need_all || playerQuest.IsNeedSync() {
		playerQuest.SetHadSync()
		playerQuest.DailyTaskReset(p.Account)
		playerQuest.UpdateCanReceiveList(p.Account)
		playerQuest.UpdateAccount7DayQuest(p.Profile.CreateTime,
			p.Profile.GetProfileNowTime())

		can_receive := playerQuest.GetCanReceivedQuest()
		s.SyncQuestCanReceiveAll = make([][]byte, 0, len(can_receive))
		for i := 0; i < len(can_receive); i++ {
			s.SyncQuestCanReceiveAll = append(
				s.SyncQuestCanReceiveAll,
				encode(questCanReceive2Client{
					can_receive[i].Id,
					i,
				}))
		}
		s.SyncQuestCaneceiveNeed = true

		received := playerQuest.GetReceivedQuest()
		s.SyncQuestReceivedAll = make([][]byte, 0, len(received))
		s.SyncQuestPoint = playerQuest.GetQuestPoint(nowT)
		s.SyncAccount7DayQuestPoint = playerQuest.Account7DayQuestPoint

		for i := 0; i < len(received); i++ {
			if !received[i].IsVailed() {
				gives_data := gamedata.GetQuestGiveData(received[i].Id)
				if gives_data == nil {
					logs.Error("Quest Err By No GiveData %d", received[i].Id)
					continue
				}

				items := gives_data.Ids
				counts := gives_data.Counts

				//logs.Trace("reward %d -->  %v --> %v %v", i, gives_data, items, counts)
				// 为了VIP每日任务单独做的逻辑
				if items != nil && len(items) > 0 && items[0] == gamedata.VI_HcByVIP {

					cfg := profile.GetMyVipCfg()
					if cfg == nil {
						logs.SentryLogicCritical(acid,
							"SyncQuest GetMyVipCfg Err by %d", vip_lv)
						items = []string{}
						counts = []uint32{}
					} else {
						items = cfg.VIPDailyGift.Ids
						counts = cfg.VIPDailyGift.Counts
					}
				}

				progress, all := received[i].GetProgress(p.Account)
				s.SyncQuestReceivedAll = append(
					s.SyncQuestReceivedAll,
					encode(questReceived2Client{
						received[i].Id,
						i,
						progress, all,
						items, counts,
					}))
			}
		}

		dailyTaskClosed := playerQuest.GetDailyTaskClosed()
		account7TaskClosed := playerQuest.GetAccount7Closed()
		for i := 0; i < len(dailyTaskClosed); i++ {
			gives := gamedata.GetQuestGiveData(dailyTaskClosed[i])
			if gives == nil {
				continue
			}
			s.SyncQuestDailyClosed = append(
				s.SyncQuestDailyClosed,
				encode(questBossClosed2Client{
					dailyTaskClosed[i],
					gives.Ids,
					gives.Counts,
				}))
		}

		if nowT < gamedata.GetAccount7DayOverTime(p.Profile.CreateTime) {
			for i := 0; i < len(account7TaskClosed); i++ {
				gives := gamedata.GetQuestGiveData(account7TaskClosed[i])
				if gives == nil {
					continue
				}
				s.SyncQuestDailyClosed = append(
					s.SyncQuestDailyClosed,
					encode(questBossClosed2Client{
						account7TaskClosed[i],
						gives.Ids,
						gives.Counts,
					}))
			}
		}
	}
}
