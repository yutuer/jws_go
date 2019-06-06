package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/festivalboss"
	"vcs.taiyouxi.net/jws/gamex/modules/moneycat_marquee"
)

// 热更版本目前有3份， 分别是 客户端， 服务器， 玩家存档
func (s *SyncResp) mkHotInfo(p *Account) {
	acid := p.AccountID.String()
	now_t := p.GetProfileNowTime()

	// market activity
	marketActivity := p.Profile.GetMarketActivitys()
	hotBuild := gamedata.GetHotDataVerCfg().Build
	s.HotDataVer = hotBuild

	resUp := p.Profile.GetMarketActivitys().UpdateMarketActivity(acid, now_t)
	// 同步配置表 一般情况下这两种情况会同时发生
	if hotBuild != s.hotDataClientVersion || resUp {
		s.makeHotConfig(p.Profile.ChannelQuickId)
		// 限时商城
		s.mkLimitGoodConfig(p)
		// 黑盒宝箱
		s.makeBlackGachaSettings()
		s.makeBlackGachaShow()
		s.makeBlackGachaLowest()
		s.makeGachaBox()
		s.makeGachaRank()
		s.makeGachaOption()
		s.makeGuildPacketItemCondition(p.Profile.GetProfileNowTime(),p.Profile.ChannelQuickId)
		s.SyncHotConfigNeed = true
	}
	updatePlayerMarketInfo(p)
	p.CheckBlackGachaActivity()

	// 更新玩家个人活动数据
	if s.SyncMarketActivityNeed || hotBuild != s.hotDataClientVersion || resUp ||
		marketActivity.SyncObj.IsNeedSync() {
		// market values
		s.makeMarketActivities(p)
	}
}

func (s *SyncResp) makeHotConfig(channelId string) []gamedata.HotActivityInfo2Client {
	actInfos, subInfos, rewInfos := gamedata.GetHotDatas().Activity.GetAllActivityInfo2Client(channelId)
	s.SyncHotActivityInfo = make([][]byte, 0, len(actInfos))

	for _, v := range actInfos {
		s.SyncHotActivityInfo = append(s.SyncHotActivityInfo, encode(v))
	}

	s.SyncMarketSubActivity = make([][]byte, 0, len(subInfos))
	for _, v := range subInfos {
		s.SyncMarketSubActivity = append(s.SyncMarketSubActivity, encode(v))
	}
	s.SyncMarketSubActivityReward = make([][]byte, 0, len(rewInfos))
	for _, v := range rewInfos {
		s.SyncMarketSubActivityReward = append(s.SyncMarketSubActivityReward, encode(v))
	}
	return actInfos
}

func updatePlayerMarketInfo(p *Account) {
	acts := gamedata.GetHotDatas().Activity.GetAllActivityInfoValid(p.Profile.ChannelQuickId, p.GetProfileNowTime())
	has_hitegg_act := false
	has_moneycat_act := false
	has_FestivalBoss_act := false
	has_whiteGacha_act := false
	has_wheelGacha_act := false
	now_t := p.GetProfileNowTime()

	for _, v := range acts {
		if v.ActivityType == gamedata.ActHitEgg {
			if now_t >= v.StartTime && now_t < v.EndTime {
				p.Profile.GetHitEgg().UpdateHitEggActivityTime(p.Account, v.StartTime, v.EndTime, now_t)
				has_hitegg_act = true
			}
		}
		if v.ActivityType == gamedata.ActMoneyCat {
			if now_t >= v.StartTime && now_t < v.EndTime {
				if v.ActivityId == p.Profile.GetMoneyCatInfo().MoneyCatActId {
					has_moneycat_act = true
					if !has_moneycat_act {
					}
				}
			}
		}
		if v.ActivityType >= gamedata.ActFestivalBoss_Begin && v.ActivityType <= gamedata.ActFestivalBoss_End {
			if now_t >= v.StartTime && now_t < v.EndTime {
				if v.ActivityId == p.Profile.FestivalBossInfo.FestivalBossActId {
					has_FestivalBoss_act = true
				}
			}
		}
		if v.ActivityType == gamedata.ActWhiteGacha {
			if now_t >= v.StartTime && now_t < v.EndTime {
				if v.ActivityId == p.Profile.GetWhiteGachaInfo().WhiteGachaActId {
					has_whiteGacha_act = true
				}
			}
		}
		if v.ActivityType == gamedata.ActLuckyWheel{
			if now_t >= v.StartTime && now_t < v.EndTime {
				if v.ActivityId == p.Profile.GetWhiteGachaInfo().WhiteGachaActId {
					has_wheelGacha_act = true
				}
			}
		}
	}
	if !has_hitegg_act {
		p.Profile.GetHitEgg().EndHitEggActivity(p.Account, now_t)
	}

	if !has_moneycat_act {
		p.Profile.GetMoneyCatInfo().SetMoneyCat2Zero()
		moneycat_marquee.GetModule(p.AccountID.ShardId).TrySetMoneyCatInfo2Zero(gamedata.ActMoneyCat,
			p.Profile.ChannelQuickId, p.Profile.GetProfileNowTime())

	}

	if !has_FestivalBoss_act {
		p.Profile.GetFestivalBossInfo().SetFbKillTime2zero()
		p.Profile.GetFestivalBossInfo().SetFbShopRewardTime2zero()
		festivalboss.GetModule(p.AccountID.ShardId).TrySetFestivalBoss2Zero()
	}

	if !has_whiteGacha_act {
		p.Profile.GetWhiteGachaInfo().SetWhiteGacha2Zero()
	}

	if !has_wheelGacha_act{
		p.Profile.GetWheelGachaInfo().ClearInfo()
	}
}

func (s *SyncResp) makeMarketActivities(p *Account) {
	s.SyncMarketActivityNeed = true
	acid := p.AccountID.String()
	now_t := p.GetProfileNowTime()
	marketActivity := p.Profile.GetMarketActivitys()

	p.Profile.GetMarketActivitys().OnLogin(acid, now_t)
	p.Profile.GetMarketActivitys().UpdateGVGDailySignOnLogin(acid, now_t)
	r, rs := marketActivity.GetMarketActivityForClient(p.AccountID.String(), p.Profile.GetProfileNowTime())
	s.SyncMarketActivityInfo = make([][]byte, 0, len(r))
	for _, _r := range r {
		s.SyncMarketActivityInfo = append(s.SyncMarketActivityInfo, encode(_r))
	}
	s.SyncMarketActivityStates = rs
	marketActivity.SyncObj.SetHadSync()
}
