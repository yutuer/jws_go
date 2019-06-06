package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	default_recover_count = 10
)

type RecoverItem struct {
	RecoverId      uint32 `json:"rid" codec:"rid"`
	GiveScTyp      string `json:"sc_t" codec:"sc_t"`
	FreeGiveGotten bool   `json:"fggten" codec:"fggten"`
	FreeGiveCount  uint32 `json:"fgc" codec:"fgc"`
	HcGiveCount    uint32 `json:"hcgc" codec:"hcgc"` // 若FreeGiveCount领取时，要减去FreeGiveCount
	HcTyp          string `json:"hc_t" codec:"hc_t"`
	HcCost         uint32 `json:"hcc" codec:"hcc"`
}

type PlayerRecover struct {
	RecoverItems []RecoverItem `json:"recovers"` // index对应cfg的recoverid
}

func (r *PlayerRecover) GetRecoverForClient() []RecoverItem {
	return r.RecoverItems
}

const (
	_ = iota
	Recover_Err_Param
	Recover_Err_NoReward
	Recover_Err_Hc_Not_Enough
	Recover_Err_Free_Already_Gotten
)

func (r *PlayerRecover) Award(a *Account, rid uint32, ishc bool, give *GiveGroup,
	cost *CostGroup, sync interfaces.ISyncRspWithRewards) (resCode uint32, errMsg string) {

	if rid >= uint32(len(r.RecoverItems)) {
		return Recover_Err_Param, "Err_Param"
	}
	rec_info := &(r.RecoverItems[rid])
	if rec_info.GiveScTyp == "" {
		return Recover_Err_NoReward, "Err_NoReward"
	}
	if ishc { // hc 领奖
		costdata := &gamedata.CostData{}
		costdata.AddItem(rec_info.HcTyp, rec_info.HcCost)
		if !cost.AddCostData(a, costdata) {
			return Recover_Err_Hc_Not_Enough, "Err_Hc_Not_Enough"
		}
		_giveAward(a, give, rec_info.GiveScTyp, rec_info.HcGiveCount, sync)
		rec_info.clear()
	} else { // 免费领奖
		if rec_info.FreeGiveGotten {
			return Recover_Err_Free_Already_Gotten, "Err_Free_Already_Gotten"
		}
		_giveAward(a, give, rec_info.GiveScTyp, rec_info.FreeGiveCount, sync)
		rec_info.HcGiveCount -= rec_info.FreeGiveCount
		rec_info.FreeGiveGotten = true
	}
	return 0, ""
}

func (r *PlayerRecover) initRecover() {
	if r.RecoverItems == nil {
		r.RecoverItems = make([]RecoverItem, default_recover_count)
		for i, _ := range r.RecoverItems {
			r.RecoverItems[i].RecoverId = uint32(i)
		}
	}
}

func (r *PlayerRecover) onAfterLogin(a *Account, loginTime, lastLogoutTime int64) {
	r.initRecover()
	if lastLogoutTime <= 0 { // 首次登陆
		return
	}
	start := gamedata.GetCommonDayBeginSec(lastLogoutTime)
	end := gamedata.GetCommonDayBeginSec(loginTime)
	diffDay := (end - start) / util.DaySec
	if diffDay > 0 { // 累计离线时间超过一天，就要重新计算追回奖励
		configMaxDayCount := gamedata.GetRecoverSetting().GetRecoverDays()
		if diffDay > int64(configMaxDayCount) {
			diffDay = int64(configMaxDayCount)
		}
		for id, recoverCfg := range gamedata.GetAllRecoverCfgs() {
			rec_info := &(r.RecoverItems[id])
			rec_info.clear()
			for _, retailCfg := range recoverCfg.Retails {
				switch retailCfg.GetRecoverType() {
				case gamedata.RecoverTyp_GameMode:
					modCondId, res := gamedata.GetGameMode2CondModId(retailCfg.GetRecoverPara())
					if !res {
						continue
					}
					if !CondCheck(modCondId, a) {
						continue
					}
					rec_info.addReward(recoverCfg.Recover, retailCfg)
				case gamedata.RecoverTyp_Quest:
					if a.Profile.GetQuest().CheckQuestReceiveConds(a,
						gamedata.GetQuestNeedCheckById(retailCfg.GetRecoverPara())) {
						rec_info.addReward(recoverCfg.Recover, retailCfg)
					}
				}
			}
			rec_info.addRewardFinial(uint32(diffDay), recoverCfg.Recover)

		}
	}
}

func (r *RecoverItem) addReward(recoverCfg *ProtobufGen.RECOVER,
	retailCfg *ProtobufGen.RECOVERDETAIL) {

	r.GiveScTyp = recoverCfg.GetRecoverAward()
	freeGive := uint32(float32(retailCfg.GetRecoverNum()) * recoverCfg.GetRecoverRatio())
	r.FreeGiveCount += freeGive
	r.HcGiveCount += retailCfg.GetRecoverNum()
	r.HcCost += retailCfg.GetRecoverCost()
}

func (r *RecoverItem) addRewardFinial(dayCount uint32, recoverCfg *ProtobufGen.RECOVER) {
	if r.HcGiveCount > 0 {
		r.FreeGiveCount = r.FreeGiveCount * dayCount
		r.HcGiveCount = r.HcGiveCount * dayCount
		r.HcCost = r.HcCost * dayCount
		if r.HcCost > recoverCfg.GetCoinMax() {
			r.HcCost = recoverCfg.GetCoinMax()
		}
		r.HcTyp = recoverCfg.GetRecoverCoin()
		logs.Info("PlayerRecover id %d dayCount %d reward %v", r.RecoverId, dayCount, r)
	} else {
		r.clear()
	}
}

func (r *RecoverItem) clear() {
	r.GiveScTyp = ""
	r.FreeGiveCount = 0
	r.HcGiveCount = 0
	r.HcTyp = ""
	r.HcCost = 0
	r.FreeGiveGotten = false
}

func _giveAward(
	a *Account, give *GiveGroup,
	scTyp string, count uint32,
	sync interfaces.ISyncRspWithRewards) {
	costdata := &gamedata.CostData{}
	costdata.AddItem(scTyp, count)
	give.AddCostData(costdata)
}
