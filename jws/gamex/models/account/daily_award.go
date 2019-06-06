package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
)

const (
	AwardState_NotGet = iota
	AwardState_Got
	AwardState_TimeOut
	AwardState_Not_Yet
)

type DailyAward struct {
	AwardStates     []int `json:"s"`
	NextRefreshTime int64 `json:"nreft"`
}

type PlayerDailyAwards struct {
	DailyAwards []DailyAward `json:"das"`
}

func (das *PlayerDailyAwards) onAfterLogin(curr_time int64) {
	if das.DailyAwards == nil {
		das.DailyAwards = make([]DailyAward, gamedata.GetDailyAwardCount())
	}
	if len(das.DailyAwards) < gamedata.GetDailyAwardCount() {
		for i := len(das.DailyAwards); i < gamedata.GetDailyAwardCount(); i++ {
			das.DailyAwards = append(das.DailyAwards, DailyAward{})
		}
	}
	das.UpdateDailyAwards(curr_time)
}

func (das *PlayerDailyAwards) UpdateDailyAwards(curr_time int64) {
	for i, _ := range das.DailyAwards {
		a := &(das.DailyAwards[i])
		id := i + 1
		subCfgs := gamedata.GetDailyAwardbyId(uint32(id))
		if curr_time > a.NextRefreshTime {
			a.NextRefreshTime = util.GetNextDailyTime(gamedata.GetCommonDayBeginSec(curr_time), curr_time)
			a.AwardStates = make([]int, len(subCfgs))
		}
		curr_begin_time := a.NextRefreshTime - util.DaySec
		for subId, cfg := range subCfgs {
			if a.AwardStates[subId-1] != AwardState_Got {
				be := util.DailyTime2UnixTime(curr_begin_time, util.DailyTimeFromString(cfg.GetFCValueSP1()))
				te := util.DailyTime2UnixTime(curr_begin_time, util.DailyTimeFromString(cfg.GetFCValueSP2()))
				if curr_time < be {
					a.AwardStates[subId-1] = AwardState_Not_Yet
				} else if curr_time > te {
					a.AwardStates[subId-1] = AwardState_TimeOut
				} else {
					a.AwardStates[subId-1] = AwardState_NotGet
				}
			}
		}
	}
}

const (
	_ = iota
	DailyAward_Err_Cfg_Not_Found
	DailyAward_Err_Award_Already_Got
	DailyAward_Err_Award_Time_Not_Satisfy
	DailyAward_Err_Give_Reward
	DailyAward_Err_Hc_Not_Enough
)

func (das *PlayerDailyAwards) AwardDailyAward(a *Account, id, subId uint32,
	sync interfaces.ISyncRspWithRewards) (int, string, int) {
	subCfgs := gamedata.GetDailyAwardbyId(id)
	if subCfgs == nil {
		return DailyAward_Err_Cfg_Not_Found, "Err_Cfg_Not_Found", 0
	}
	subCfg, ok := subCfgs[subId]
	if !ok {
		return DailyAward_Err_Cfg_Not_Found, "Err_Cfg_Not_Found", 0
	}
	award := &(das.DailyAwards[id-1])
	state := award.AwardStates[subId-1]
	if state == AwardState_NotGet { // 免费领
		now_time := a.Profile.GetProfileNowTime()
		tb := util.DailyTime2UnixTime(now_time, util.DailyTimeFromString(subCfg.GetFCValueSP1()))
		te := util.DailyTime2UnixTime(now_time, util.DailyTimeFromString(subCfg.GetFCValueSP2()))
		if now_time < tb || now_time > te {
			return DailyAward_Err_Award_Time_Not_Satisfy, "Err_Award_Time_Not_Satisfy", errCode.ClickTooQuickly
		}
		return award._award(a, subId, subCfg, sync)
	} else if state == AwardState_TimeOut { // 花钻石领取
		hc := subCfg.GetCompensateHC()
		cost := &CostGroup{}
		if !cost.AddHc(a, int64(hc)) || !cost.CostBySync(a, sync, "DailyAward") {
			return DailyAward_Err_Hc_Not_Enough, "Err_Hc_Not_Enough", errCode.ClickTooQuickly
		}
		return award._award(a, subId, subCfg, sync)
	}
	//return DailyAward_Err_Award_Already_Got, "Err_Award_Already_Got"
	return 0, "", 0
}

func (award *DailyAward) _award(a *Account, subId uint32, subCfg *ProtobufGen.DAILYAWARD,
	sync interfaces.ISyncRspWithRewards) (int, string, int) {
	data := &gamedata.CostData{}
	data.AddItem(subCfg.GetItem(), subCfg.GetCount())
	give := &GiveGroup{}
	give.AddCostData(data)
	if !give.GiveBySyncAuto(a, sync, "DailyAward") {
		return DailyAward_Err_Give_Reward, "Err_Give_Reward", 0
	}
	award.AwardStates[subId-1] = AwardState_Got
	return 0, "", 0
}

type dailyAwardForClient struct {
	AwardsStates     []int
	AwardsStatesLen  []int
	AwardNextRefTime []int64
}

func (das *PlayerDailyAwards) GetDailyAwardsForClient() dailyAwardForClient {
	res := dailyAwardForClient{}
	awards := das.DailyAwards
	res.AwardsStatesLen = make([]int, 0, len(awards))
	res.AwardNextRefTime = make([]int64, 0, len(awards))
	res.AwardsStates = make([]int, 0, len(awards)*5)
	for _, award := range awards {
		res.AwardsStates = append(res.AwardsStates, award.AwardStates...)
		res.AwardsStatesLen = append(res.AwardsStatesLen, len(award.AwardStates))
		res.AwardNextRefTime = append(res.AwardNextRefTime,
			award.NextRefreshTime)
	}
	return res
}
