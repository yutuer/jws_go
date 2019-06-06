package account

import (
	"math"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

type GameModeInfo struct {
	IsDoing              bool  // 是否当前正在活动关卡中
	tmpLastSlotBeginUnit int64 // 临时进入时所在活动时间段开始unix时间
	LastSlotBeginUnit    int64 `json:"lsbu"` // 进入时所在活动时间段开始unix时间
	LastSecTimeAfC       int64 `json:"lst"`  // 进入时刻时间，以创建角色时间为基准
}

type GameModesInfo struct {
	Info []GameModeInfo `json:"if"` // 某活动上次进入时的信息
}

type gameModeSyncInfo struct {
	GameModeId     uint32 `codec:"id"`
	CostTimes      uint32 `codec:"ct"`
	LastEnterTime  int64  `codec:"let"`
	NextUpdateTime int64  `codec:"nxtrf"` // 下次刷新时间，只是上次刷新时间加24小时，gamemode刷新规则比较复杂，这个时间可能不会真的刷新，但最小刷新间隔是24小时
	RefershTime    int64  `codec:"refersht"`
	LastTime       int64  `codec:"lastt"`
}

const (
	_ = iota
	CFG_ERR
	GAMEMODE_NOT_START
	COND_NOT_SATISFY
	TIMES_NOT_ENOUGH
	NEED_ENTER_FIRST
	IN_CD
	NO_CD_CFG
	NOT_IN_CD
	HC_NOT_ENOUGH
	CFG_Mode2Cond_ERR
	GAMEMODE_LEVEL_NOT_SWEEP
	VIP_ERR
)

func (gm *GameModesInfo) IsCanEnterGameMode(a *Account, gameModeId uint32) (bool, uint32, int) {
	gm.initGameMode()
	gmInfo := &gm.Info[gameModeId]
	// 时间检查
	bValid, sbu, err := gamedata.GetNowValidGameModeIndex(gameModeId, a.Profile.GetProfileNowTime())
	if err != nil {
		return false, CFG_ERR, 0
	}
	if !bValid {
		return false, GAMEMODE_NOT_START, 0
	}
	// 开启条件检查
	modCondId, res := gamedata.GetGameMode2CondModId(gameModeId)
	if !res {
		return false, CFG_Mode2Cond_ERR, 0
	}
	if !CondCheck(modCondId, a) {
		return false, 0, errCode.ActivityNotValid
	}
	// 次数检查
	if !a.Profile.GetCounts().Has(int(gameModeId), a) {
		return false, 0, errCode.ClickTooQuickly
	}
	// CD检查
	cdSec, _, _ := gamedata.GetGameModeCDSec(gameModeId)
	if gmInfo.LastSecTimeAfC > 0 && cdSec > 0 && a.Profile.GetRegTimeUnix()-gmInfo.LastSecTimeAfC < cdSec {
		return false, 0, errCode.ClickTooQuickly
	}

	gmInfo.tmpLastSlotBeginUnit = sbu

	return true, 0, 0
}

func (gm *GameModesInfo) EnterGameMode(gameModeId uint32) {
	gmInfo := &gm.Info[gameModeId]

	gmInfo.LastSlotBeginUnit = gmInfo.tmpLastSlotBeginUnit
	gmInfo.IsDoing = true
}

func (gm *GameModesInfo) CostGameMode(a *Account, gameModeId uint32) (bool, uint32, uint32) {
	gmInfo := &gm.Info[gameModeId]
	if !gmInfo.IsDoing {
		return false, NEED_ENTER_FIRST, errCode.ClickTooQuickly
	}
	// 时间检查
	bValid, sbu, err := gamedata.GetNowValidGameModeIndex(gameModeId, a.Profile.GetProfileNowTime())
	if err != nil {
		return false, CFG_ERR, 0
	}
	// 是否夸阶段了
	if !bValid || sbu != gmInfo.LastSlotBeginUnit {
		return true, 0, 0
	}
	// 次数检查
	if !a.Profile.GetCounts().Use(int(gameModeId), a) {
		return false, TIMES_NOT_ENOUGH, errCode.ClickTooQuickly
	}

	gmInfo.IsDoing = false
	gmInfo.LastSecTimeAfC = a.Profile.GetRegTimeUnix()
	return true, 0, 0
}

func (gm *GameModesInfo) ResetCD(a *Account, gameModeId uint32) (bool, uint32) {
	gmInfo := &gm.Info[gameModeId]
	// 时间检查
	bValid, _, err := gamedata.GetNowValidGameModeIndex(gameModeId, a.Profile.GetProfileNowTime())
	if err != nil {
		return false, CFG_ERR
	}
	if !bValid {
		return false, GAMEMODE_NOT_START
	}
	// 是否配置了cd
	cdSec, costHc, _ := gamedata.GetGameModeCDSec(gameModeId)
	if cdSec <= 0 {
		// return false, NO_CD_CFG
		return true, 0
	}
	// 次数检查
	if !a.Profile.GetCounts().Has(int(gameModeId), a) {
		return false, TIMES_NOT_ENOUGH
	}
	// CD检查
	if gmInfo.LastSecTimeAfC > 0 &&
		cdSec > 0 &&
		a.Profile.GetRegTimeUnix()-gmInfo.LastSecTimeAfC >= cdSec {
		return false, NOT_IN_CD
	}
	// 消耗hc
	if !a.Profile.GetHC().UseHcGiveFirst(
		a.AccountID.String(),
		int64(costHc),
		a.Profile.GetProfileNowTime(),
		"ResetGameModeCD") {
		return false, HC_NOT_ENOUGH
	}
	gmInfo.LastSecTimeAfC = 0
	return true, 0
}
func (gm *GameModesInfo) GameModeLevelSweep(a *Account, gameModeId uint32, sync helper.ISyncRsp) (
	ok bool, errCode uint32, warnCode int, leftCount int) {
	return gm.GameModeCheckAndCostWithChange(a, gameModeId, -1, sync)
}

func (gm *GameModesInfo) GameModeCheckAndCost(
	a *Account,
	gameModeId uint32,
	costTimes int,
	sync helper.ISyncRsp) (
	ok bool, errcode, warnCode uint32, leftCount int) {
	gm.initGameMode()
	gmInfo := &gm.Info[gameModeId]

	// 开启条件检查
	modCondId, res := gamedata.GetGameMode2CondModId(gameModeId)
	if !res {
		return false, CFG_Mode2Cond_ERR, 0, 0
	}
	if !CondCheck(modCondId, a) {
		return false, COND_NOT_SATISFY, errCode.ClickTooQuickly, 0
	}
	// 时间检查
	bValid, _, err := gamedata.GetNowValidGameModeIndex(gameModeId, a.Profile.GetProfileNowTime())
	if err != nil {
		return false, CFG_ERR, 0, 0
	}
	// 是否结束了
	if !bValid {
		return false, GAMEMODE_NOT_START, 0, 0
	}
	// CD检查
	cdSec, _, _ := gamedata.GetGameModeCDSec(gameModeId)
	if gmInfo.LastSecTimeAfC > 0 &&
		cdSec > 0 &&
		a.Profile.GetRegTimeUnix()-gmInfo.LastSecTimeAfC >= cdSec {
		return false, NOT_IN_CD, errCode.ClickTooQuickly, 0
	}
	// 消耗
	leftCount, _ = a.Profile.GetCounts().Get(int(gameModeId), a)
	if leftCount <= 0 {
		return false, TIMES_NOT_ENOUGH, errCode.ClickTooQuickly, 0
	}
	if costTimes < 0 {
		costTimes = leftCount
	} else if leftCount < costTimes {
		return false, TIMES_NOT_ENOUGH, errCode.ClickTooQuickly, 0
	}
	cfg := gamedata.GetGameModeCfg(gameModeId)
	hc := cfg.GetCostOperation()
	if hc > 0 {
		cost := &CostGroup{}
		if !cost.AddHc(a, int64(hc)) || !cost.CostBySync(a, sync, "GameModeLevelSweep") {
			return false, HC_NOT_ENOUGH, errCode.ClickTooQuickly, 0
		}
	}
	// 次数减少
	for i := 0; i < costTimes; i++ {
		a.Profile.GetCounts().Use(int(gameModeId), a)
	}

	gmInfo.IsDoing = false
	gmInfo.LastSecTimeAfC = a.Profile.GetRegTimeUnix()
	return true, 0, 0, costTimes
}

func (gm *GameModesInfo) GameModeCheckAndNoCost(
	a *Account,
	gameModeId uint32,
	costTimes int,
	sync helper.ISyncRsp) (
	ok bool, errcode, warnCode uint32, leftCount int) {
	gm.initGameMode()
	gmInfo := &gm.Info[gameModeId]

	// 开启条件检查
	modCondId, res := gamedata.GetGameMode2CondModId(gameModeId)
	if !res {
		return false, CFG_Mode2Cond_ERR, 0, 0
	}
	if !CondCheck(modCondId, a) {
		return false, COND_NOT_SATISFY, errCode.ClickTooQuickly, 0
	}
	// 时间检查
	bValid, _, err := gamedata.GetNowValidGameModeIndex(gameModeId, a.Profile.GetProfileNowTime())
	if err != nil {
		return false, CFG_ERR, 0, 0
	}
	// 是否结束了
	if !bValid {
		return false, GAMEMODE_NOT_START, 0, 0
	}
	// CD检查
	cdSec, _, _ := gamedata.GetGameModeCDSec(gameModeId)
	if gmInfo.LastSecTimeAfC > 0 &&
		cdSec > 0 &&
		a.Profile.GetRegTimeUnix()-gmInfo.LastSecTimeAfC < cdSec {
		return false, IN_CD, errCode.ClickTooQuickly, 0
	}
	// 消耗
	leftCount, _ = a.Profile.GetCounts().Get(int(gameModeId), a)
	if leftCount <= 0 {
		return false, TIMES_NOT_ENOUGH, errCode.ClickTooQuickly, 0
	}
	if costTimes < 0 {
		costTimes = leftCount
	} else if leftCount < costTimes {
		return false, TIMES_NOT_ENOUGH, errCode.ClickTooQuickly, 0
	}
	// 次数减少
	for i := 0; i < costTimes; i++ {
		a.Profile.GetCounts().Use(int(gameModeId), a)
	}

	gmInfo.IsDoing = false
	gmInfo.LastSecTimeAfC = a.Profile.GetRegTimeUnix()
	return true, 0, 0, costTimes
}

func (gm *GameModesInfo) GameModeCheckAndCostWithChange(a *Account, gameModeId uint32, costTimes int, sync helper.ISyncRsp) (
	ok bool, errcode uint32, warnCode int, leftCount int) {
	gm.initGameMode()
	gmInfo := &gm.Info[gameModeId]

	// 开启条件检查
	modCondId, res := gamedata.GetGameMode2CondModId(gameModeId)
	if !res {
		return false, CFG_Mode2Cond_ERR, 0, 0
	}
	if !CondCheck(modCondId, a) {
		return false, COND_NOT_SATISFY, errCode.ClickTooQuickly, 0
	}
	// 时间检查
	bValid, _, err := gamedata.GetNowValidGameModeIndex(
		gameModeId,
		a.Profile.GetProfileNowTime())
	if err != nil {
		return false, CFG_ERR, 0, 0
	}
	// 是否结束了
	if !bValid {
		return false, GAMEMODE_NOT_START, 0, 0
	}
	// vip检查
	if !gamedata.SweepVipVaild(gameModeId, a.Profile.GetVipLevel()) {
		return false, VIP_ERR, 0, 0
	}

	// 消耗
	leftCount, _ = a.Profile.GetCounts().Get(int(gameModeId), a)
	max := a.Profile.GetCounts().GetDailyMax(int(gameModeId))
	if leftCount <= 0 {
		return false, 0, errCode.ClickTooQuickly, 0
	}
	if costTimes < 0 {
		costTimes = leftCount
	} else if leftCount < costTimes {
		return false, 0, errCode.ClickTooQuickly, 0
	}
	cfg := gamedata.GetGameModeCfg(gameModeId)
	hc := cfg.GetCostOperation()
	if hc > 0 {
		hcNeed := float64(hc) * float64(leftCount) / float64(max)
		hcNeed = math.Ceil(hcNeed)

		cost := &CostGroup{}
		if !cost.AddHc(a, int64(hcNeed)) ||
			!cost.CostBySync(a, sync, "GameModeLevelSweep") {
			return false, HC_NOT_ENOUGH, 0, 0
		}
	}
	// 次数减少
	for i := 0; i < costTimes; i++ {
		a.Profile.GetCounts().Use(int(gameModeId), a)
	}

	gmInfo.IsDoing = false
	gmInfo.LastSecTimeAfC = a.Profile.GetRegTimeUnix()
	return true, 0, 0, costTimes
}

func (gm *GameModesInfo) GetSyncInfo(a *Account, gameModeId uint32) gameModeSyncInfo {
	gm.initGameMode()
	gmInfo := &gm.Info[gameModeId]
	costTimes, nextRefTime, refTime, lastTime :=
		a.Profile.GetCounts().GetWithTimeValue(int(gameModeId), a)
	return gameModeSyncInfo{
		GameModeId:     gameModeId,
		CostTimes:      costTimes,
		LastEnterTime:  getLastEnterTimeForClient(a, gmInfo.LastSecTimeAfC),
		NextUpdateTime: nextRefTime,
		// for inc by time
		RefershTime: refTime,
		LastTime:    lastTime,
	}
}

func (gm *GameModesInfo) GetAllSyncInfo(a *Account) []gameModeSyncInfo {
	gm.initGameMode()
	res := []gameModeSyncInfo{}
	for _, gameModeId := range gamedata.GetAllGameModeId() {
		if int(gameModeId) >= len(gm.Info) {
			continue
		}
		gmInfo := &gm.Info[gameModeId]
		costTimes, nextRefTime, refTime, lastTime :=
			a.Profile.GetCounts().GetWithTimeValue(int(gameModeId), a)
		info := gameModeSyncInfo{
			GameModeId:     gameModeId,
			CostTimes:      costTimes,
			LastEnterTime:  getLastEnterTimeForClient(a, gmInfo.LastSecTimeAfC),
			NextUpdateTime: nextRefTime,
			RefershTime:    refTime,
			LastTime:       lastTime,
		}
		res = append(res, info)
	}
	return res
}

func getLastEnterTimeForClient(a *Account, lastSecTimeAfC int64) int64 {
	if lastSecTimeAfC <= 0 {
		return 0
	}
	return lastSecTimeAfC + a.Profile.CreateTime
}

func (gm *GameModesInfo) initGameMode() {
	if len(gm.Info) <= 0 {
		gm.Info = make([]GameModeInfo, gamedata.CounterTypeCountMax)
	}
	if len(gm.Info) < gamedata.CounterTypeCountMax {
		new := make([]GameModeInfo, gamedata.CounterTypeCountMax)
		copy(new, gm.Info)
		gm.Info = new
	}
}
