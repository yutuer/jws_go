package logics

import (
	//"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	//"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/stage_star"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) IsStageCanPlay(stage_id string, avatars []int, is_sweep bool) (uint32, int) {
	const (
		_             = iota
		Err_Energy    // 体力不够
		Err_Lv        // 战队等级不够
		Err_AvatarLv  // 角色等级不够
		Err_AvatarSep // 没有专属角色
		Err_Sweep     // 无法扫荡
		Err_Count     // 每日次数不够
		Err_Pre       // 前置关卡
		Err_SweepTicket
		Err
		Err_HighEnergy // 包子不够
		Err_MiniLvl_Already_Pass
	)

	stage_data := gamedata.GetStageData(stage_id)
	if stage_data == nil {
		logs.SentryLogicCritical(p.AccountID.String(), "stagedata not find %s", stage_id)
		return Err, 0
	}

	player_stage := p.Profile.GetStage()
	player_stage_info := player_stage.GetStageInfo(
		gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()),
		stage_id,
		p.GetRand())
	switch stage_data.Type {
	case gamedata.LEVEL_TYPE_HELL, gamedata.LEVEL_TYPE_ELITE:
		if !p.Profile.GetSC().HasSC(gamedata.SC_BaoZi, int64(stage_data.HighEnergy)) {
			return Err_HighEnergy, 0
		}
	default:
		if !p.Profile.GetEnergy().Has(int64(stage_data.Energy)) {
			return Err_Energy, 0
		}
	}

	// 检查次数
	if stage_data.MaxDailyAccess != 0 &&
		player_stage_info.T_count >= stage_data.MaxDailyAccess {
		return Err_Count, 0
	}

	// 检查战队等级
	corp_lv, _ := p.Profile.GetCorp().GetXpInfo()
	if stage_data.CorpLvRequirement != 0 &&
		corp_lv < uint32(stage_data.CorpLvRequirement) {
		return Err_Lv, 0
	}

	// 检查角色等级
	player_avatar := p.Profile.GetAvatarExp()

	for i := 0; i < len(avatars); i++ {
		lv, _ := player_avatar.Get(avatars[i])
		if lv < uint32(stage_data.LevelRequirement) {
			return Err_AvatarLv, 0
		}
	}
	// 检查关卡类型
	if stage_data.Type == gamedata.LEVEL_TYPE_MINILEVEL {
		if p.Profile.GetStage().IsStagePass(stage_id) {
			return Err_MiniLvl_Already_Pass, 0
		}
	}

	if is_sweep {
		if stage_star.GetStarCount(player_stage_info.MaxStar) < MAX_STAR_Count {
			return Err_Sweep, 0
		}

		/*
			暂时屏蔽扫荡券
			cost_ok := p.Profile.GetSC().HasSC(
				helper.SC_SweepTicket, 1)

			if !cost_ok {
				return Err_SweepTicket
			}

		*/
	}

	// 检查专属角色
	if stage_data.RoleOnly >= 0 {
		if len(avatars) != 1 || avatars[0] != int(stage_data.RoleOnly) {
			return Err_AvatarSep, 0
		}
	}

	// 检查前置关卡
	if !player_stage.IsAllPreStagePass(stage_data.PreLevelID) {
		return Err_Pre, 0
	}

	// TBD 前置检查各种限制 以后还会有新的

	const (
		CODE_MIN_GAMEMODE = 20
	)
	// 活动相关检查
	if stage_data.GameModeId > 0 {
		if ok, errcode, warncode := p.Profile.GameMode.IsCanEnterGameMode(p.Account, stage_data.GameModeId); !ok {
			return CODE_MIN_GAMEMODE + errcode, warncode
		}
	}
	return 0, 0
}

// 在检测一遍可不可以打关卡，之后使用体力
func (p *Account) CostStagePay(id string, avatars []int, is_sweep bool, resp helper.ISyncRsp) (uint32, uint32) {
	const (
		_                      = iota
		Err_Energy             // 体力不够
		Err_Lv                 // 战队等级不够
		Err_AvatarLv           // 角色等级不够
		Err_AvatarSep          // 没有专属角色
		Err_Sweep              // 无法扫荡
		Err_Count              // 每日次数不够
		Err_Pre                // 前置关卡
		Err_SweepTicket        // 无法扫荡
		Err_VIP_CFG_NOT_FOUND  // 配置出错
		Err_GameMode_Not_Sweep // gamemode level 不能扫荡
		Err
		Err_HighEnergy // 包子不够
	)

	acid := p.AccountID.String()

	stage_data := gamedata.GetStageData(id)
	if stage_data == nil {
		logs.SentryLogicCritical(acid,
			"stagedata not find %s", id)
		return Err, 0
	}

	player_stage := p.Profile.GetStage()
	player_stage_info := player_stage.GetStageInfo(
		gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()),
		id,
		p.GetRand())

	if stage_data.MaxDailyAccess != 0 &&
		player_stage_info.T_count >= stage_data.MaxDailyAccess {
		logs.Warn("CostStagePay Err_Count")
		return 0, errCode.ClickTooQuickly
	}

	// 检查角色等级
	player_avatar := p.Profile.GetAvatarExp()

	for i := 0; i < len(avatars); i++ {
		lv, _ := player_avatar.Get(avatars[i])
		if lv < uint32(stage_data.LevelRequirement) {
			return Err_AvatarLv, 0
		}
	}

	// 检查专属角色
	if stage_data.RoleOnly >= 0 {
		if len(avatars) != 1 || avatars[0] != int(stage_data.RoleOnly) {
			return Err_AvatarSep, 0
		}
	}

	if is_sweep {
		curVip, _ := p.Profile.GetVip().GetVIP()
		vipInfo := gamedata.GetVIPCfg(int(curVip))
		if vipInfo == nil {
			return Err_VIP_CFG_NOT_FOUND, 0
		}
		if (stage_data.Type == gamedata.LEVEL_TYPE_HELL &&
			vipInfo.HellNotStarSweep) ||
			(stage_data.Type == gamedata.LEVEL_TYPE_ELITE &&
				vipInfo.EliteNotStarSweep) ||
			(stage_data.Type == gamedata.LEVEL_TYPE_MAIN &&
				vipInfo.OrdinaryNotStarSweep) {
			if player_stage_info.MaxStar <= 0 {
				return Err_Sweep, 0
			}
		} else if stage_star.GetStarCount(player_stage_info.MaxStar) < MAX_STAR_Count {
			return Err_Sweep, 0
		}

		/*
			暂时屏蔽扫荡券
			cost_ok := p.Profile.GetSC().UseSC(acid,
				helper.SC_SweepTicket,
				1, "StageSweep")

			if !cost_ok {
				return Err_SweepTicket
			}
		*/
	}

	switch stage_data.Type {
	case gamedata.LEVEL_TYPE_HELL, gamedata.LEVEL_TYPE_ELITE:
		data := &gamedata.CostData{}
		data.AddItem(gamedata.VI_BaoZi, uint32(stage_data.HighEnergy))
		if !account.CostBySync(p.Account, data, resp, "Stage") {
			return Err_HighEnergy, 0
		}
	default:
		if !p.Profile.GetEnergy().Use(acid, "Stage", int64(stage_data.Energy)) {
			return Err_Energy, 0
		}
	}

	// TBD 这里要返还体力么？
	if !player_stage.IsAllPreStagePass(stage_data.PreLevelID) {
		return Err_Pre, 0
	}

	const (
		CODE_MIN_GAMEMODE = 20
	)
	// 活动相关检查
	if stage_data.GameModeId > 0 {
		if is_sweep {
			return Err_GameMode_Not_Sweep, 0
		} else {
			logs.Warn("CostStagePay game mode ")
			if ok, errcode, warncode := p.Profile.GameMode.CostGameMode(p.Account, stage_data.GameModeId); !ok {
				if warncode > 0 {
					return 0, warncode
				}
				return CODE_MIN_GAMEMODE + errcode, 0
			}
		}

		resp.OnChangeGameMode(stage_data.GameModeId)
	}

	return 0, 0
}
