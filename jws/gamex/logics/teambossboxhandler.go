package logics

import (
	"fmt"

	"math/rand"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice/teamboss"
	helper2 "vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// TBBattleStart : 组队BOSS战开始
// 组队BOSS战开始
func (p *Account) TBBattleStartHandler(req *reqMsgTBBattleStart, resp *rspMsgTBBattleStart) uint32 {
	p.Profile.GetTeamBossStorageInfo().TryDailyReset(p.Profile.GetProfileNowTime())
	teamId := req.TBBattleSTeamId
	startInfo := &helper2.StartFightInfo{
		RoomID: teamId,
		AcID:   p.AccountID.String(),
	}
	url, code, err := teamboss.StartFight(p.AccountID.ShardId, p.AccountID.String(), startInfo)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> start battle crossservice err %v", err)
		return errCode.CommonInner
	}
	if url.Code != 0 {
		return CrossError2ClientError(url.Code)
	}
	p.Profile.GetTeamBossTeamInfo().GlobalRoomId = url.GlobalRoomID
	resp.TBBattleServUrl = url.ServerUrl
	resp.TBBattleSGlobalTeamId = url.GlobalRoomID
	return 0
}

// TBBattleEnd : 组队BOSS战结束
// 组队BOSS战结束
func (p *Account) TBBattleEndHandler(req *reqMsgTBBattleEnd, resp *rspMsgTBBattleEnd) uint32 {
	p.Profile.GetTeamBossStorageInfo().TryDailyReset(p.Profile.GetProfileNowTime())
	teamId := req.TBBattleETeamId
	battelId := req.TBBattleEGlobalTeamId
	var rewardBoxId string
	endInfo := &helper2.EndFightInfo{
		RoomID:       teamId,
		GlobalRoomID: battelId,
		AcID:         p.AccountID.String(),
	}
	endFight, code, err := teamboss.EndFight(p.AccountID.ShardId, p.AccountID.String(), endInfo)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> end battle crossservice err %v", err)
		return errCode.CommonInner
	}
	if endFight.Code != 0 {
		return CrossError2ClientError(endFight.Code)
	}
	if endFight.HasReward {
		costGoldData := &gamedata.CostData{}
		costHCData := &gamedata.CostData{}
		mainData := gamedata.GetTBossMainDataByDiff(endFight.Level)
		if mainData == nil {
			return errCode.CommonInner
		}
		giveData := &gamedata.CostData{}
		for _, itemR := range mainData.GetReward_Table() {
			if itemR.GetItemID() != "" {
				giveData.AddItem(itemR.GetItemID(), itemR.GetItemNum())
			}
		}
		reason := fmt.Sprintf("tbEndBattle")
		if ok := account.GiveBySync(p.Account, giveData, resp, reason); !ok {
			return errCode.CommonConditionFalse
		}

		bossInfo := p.Profile.GetTeamBossStorageInfo()
		costGold := mainData.GetGoldCost()
		costGoldData.AddItem(gamedata.VI_Sc0, costGold)
		reason = fmt.Sprintf("tbCostGold")
		if ok := account.CostBySync(p.Account, costGoldData, resp, reason); !ok {
			return errCode.ClickTooQuickly
		}
		if endFight.HasRedBox {
			if endFight.IsCost {
				costHc := mainData.GetRedBoxCost()
				costHCData.AddItem(gamedata.VI_Hc, costHc)
				reason = fmt.Sprintf("tbRedboxHC")
				if ok := account.CostBySync(p.Account, costHCData, resp, reason); !ok {
					return errCode.ClickTooQuickly
				}
			}
			rewardBoxId = gamedata.GetRedBoxId(endFight.Level)
			//判断如果没拿到红宝箱id
			if rewardBoxId == "" {
				return errCode.CommonInner
			}
			bossInfo.AddNewBox(rewardBoxId)
			logs.Debug("<TBoss> end fight red box is: %v", rewardBoxId)
		} else {
			boxCtrlTimes := bossInfo.BoxCtrlTimes
			boxCtrlCfg := gamedata.GetTBossVipCtrl(p.Profile.Vip.V)

			if boxCtrlTimes >= boxCtrlCfg.GetGoodBoxControl() {
				bossInfo.ResetControlTimes()
				logs.Debug("box sepcial drop group")
				rewardBoxId = gamedata.RandomTBBox(mainData.GetSepcialDropGroup())
			} else {
				logs.Debug("box normal drop group")
				rewardBoxId = gamedata.RandomTBBox(mainData.GetBoxDropGroup())
				if gamedata.IsRedOrGoldenBox(rewardBoxId) {
					bossInfo.ResetControlTimes()
				} else {
					bossInfo.IncreaseControlTimes()
				}
			}
			bossInfo.AddNewBox(rewardBoxId)

		}

		//如果宝箱仓库的三个空没满
		if p.Profile.GetTeamBossStorageInfo().GetTBossBoxCount() < account.BOX_POS_TYPE_COUNT {
			resp.TBBoxIsFull = false
		} else {
			resp.TBBoxIsFull = true
		}
		resp.RewardID = append(resp.RewardID, rewardBoxId)
		resp.RewardCount = append(resp.RewardCount, 1)
		resp.RewardData = append(resp.RewardData, "")
	}
	p.Profile.GetTeamBossTeamInfo().GlobalRoomId = ""
	return 0
}

// TBOpenStorage : 打开组队BOSS仓库
// 打开组队BOSS仓库
func (p *Account) TBOpenStorageHandler(req *reqMsgTBOpenStorage, resp *rspMsgTBOpenStorage) uint32 {
	p.Profile.GetTeamBossStorageInfo().TryDailyReset(p.Profile.GetProfileNowTime())
	nowTime := p.Profile.GetProfileNowTime()
	storageInfo := p.Profile.GetTeamBossStorageInfo()
	for i := range storageInfo.BoxList[1:] {
		p.Profile.GetTeamBossStorageInfo().MoveStorageBox(i, nowTime)
	}

	resp.TBBoxHCOpenTimes = int64(storageInfo.OpenTimes)
	resp.TBBoxInfo = make([][]byte, 4)
	for i, box := range storageInfo.BoxList {
		if box.TBBoxId != "" {
			resp.TBBoxInfo[i] = encode(storage2Client(box, i))
		}
	}
	return 0
}

func storage2Client(info account.TeamBossBoxInfo, index int) TBBoxInfo {
	return TBBoxInfo{
		TBBoxId:      info.TBBoxId,
		TBBoxEndTime: info.TBBoxEndTime,
		TBBoxPos:     int64(index),
	}
}

// TBOpenBox : 打开组队BOSS宝箱
// 打开组队BOSS宝箱
func (p *Account) TBOpenBoxHandler(req *reqMsgTBOpenBox, resp *rspMsgTBOpenBox) uint32 {
	p.Profile.GetTeamBossStorageInfo().TryDailyReset(p.Profile.GetProfileNowTime())
	index := int(req.TBBoxOpIndex)
	if index < account.BOX_POS_TYPE_1 || index > account.BOX_POS_TYPE_3 {
		return errCode.CommonInvalidParam
	}
	boxInfo := p.Profile.GetTeamBossStorageInfo().BoxList[index]
	// 空箱子
	if boxInfo.TBBoxId == "" {
		return errCode.CommonInvalidParam
	}

	// 判断花钻石和时间是否一致
	if retCode := p.checkBoxTime(req.TBBoxOpType, boxInfo); retCode != 0 {
		return retCode
	}

	nowTime := p.Profile.GetProfileNowTime()

	// 扣钻石
	costData := &gamedata.CostData{}
	costHc := 0
	if req.TBBoxOpType == 1 {
		logs.Debug("<TBoss> tboss already open times: %v", p.Profile.GetTeamBossStorageInfo().OpenTimes)
		if p.Profile.GetTeamBossStorageInfo().OpenTimes >= gamedata.BoxCfg.HCOpenBoxNum {
			return errCode.CommonCountLimit
		}
		costHc = gamedata.CalTBOpenCost(boxInfo.TBBoxEndTime - nowTime)

	}
	costData.AddItem(gamedata.VI_Hc, uint32(costHc))
	reason := fmt.Sprintf("tbOpenboxHC")
	if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}

	// 开宝箱
	giveData := &gamedata.CostData{}
	lootTable := gamedata.GetTBBoxLootTableByBoxID(boxInfo.TBBoxId)
	for _, item := range lootTable {
		if rand.Float32() < item.GetLootChance() {
			if item.GetItemID() != "" {
				giveData.AddItem(item.GetItemID(), item.GetLootNumber())
			}
		}
	}

	if req.TBBoxOpType == 1 {
		p.Profile.GetTeamBossStorageInfo().OpenTimes++
	}

	// 重置宝箱
	p.Profile.GetTeamBossStorageInfo().ResetBox(index)

	// 花费钻石
	reason = fmt.Sprintf("tbOpenBoxHC")
	if ok := account.GiveBySync(p.Account, giveData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}

	// 如果预存仓库有宝箱，需要挪到倒计时位置里面
	p.Profile.GetTeamBossStorageInfo().MoveStorageBox(index, nowTime)

	// 返回客户端信息
	storageInfo := p.Profile.GetTeamBossStorageInfo()
	resp.TBBoxHCOpenTimes = int64(storageInfo.OpenTimes)
	resp.TBBoxInfo = make([][]byte, 0)
	for i, box := range storageInfo.BoxList {
		resp.TBBoxInfo = append(resp.TBBoxInfo, encode(storage2Client(box, i)))
	}
	return 0
}

func (p *Account) checkBoxTime(openType int64, info account.TeamBossBoxInfo) uint32 {
	nowTime := p.GetProfileNowTime()
	if openType == 0 && nowTime < info.TBBoxEndTime {
		return errCode.CommonInvalidParam
	}
	if openType == 1 && nowTime > info.TBBoxEndTime {
		return errCode.CommonInvalidParam
	}
	return 0
}

// TBDelBox : 删除组队BOSS宝箱
// 删除组队BOSS宝箱
func (p *Account) TBDelBoxHandler(req *reqMsgTBDelBox, resp *rspMsgTBDelBox) uint32 {
	nowTime := p.Profile.GetProfileNowTime()
	p.Profile.GetTeamBossStorageInfo().TryDailyReset(nowTime)
	index := int(req.TBBoxPos)
	if index <= account.BOX_POS_TYPE_STORAGE || index >= account.BOX_POS_TYPE_COUNT {
		return errCode.CommonInvalidParam
	}
	p.Profile.GetTeamBossStorageInfo().ResetBox(index)
	// 如果预存仓库有宝箱，需要挪到倒计时位置里面
	p.Profile.GetTeamBossStorageInfo().MoveStorageBox(index, nowTime)
	resp.TBBoxInfo = make([][]byte, 0)
	for i, box := range p.Profile.GetTeamBossStorageInfo().BoxList {
		resp.TBBoxInfo = append(resp.TBBoxInfo, encode(storage2Client(box, i)))
	}
	resp.TBBoxHCOpenTimes = int64(p.Profile.GetTeamBossStorageInfo().OpenTimes)
	return 0
}

func CrossError2ClientError(code int) uint32 {
	switch code {
	case helper2.RetCodeSuccess:
		return 0
	case helper2.RetCodeFail:
		return errCode.CommonInner
	case helper2.RetCodeRoomNotExist:
		return errCode.TBossRoomIsNotExist
	case helper2.RetCodeOptInvalid:
		return errCode.CommonInner
	case helper2.RetCodeOptLimitPermission:
		return errCode.CommonCountLimit
	case helper2.RetCodeStartFightFailed:
		return errCode.TBossStartBattleFailed
	case helper2.RetCodeDataError:
		return errCode.CommonInitFailed
	case helper2.RetCodePositionOccupied:
		return errCode.TBossHeroChooseOccupied
	case helper2.RetCodeRoomPlayerFull:
		return errCode.TBossRoomIsFull
	case helper2.RetCodeRoomInBattle:
		return errCode.TBossRoomInBattle
	case helper2.RetCodeAlreadyTickRedBox:
		return errCode.TBossAreadyTickRedBox
	case helper2.RetCodeKickFightingRoom:
		return errCode.TBossRoomCantKickInBattle
	case helper2.RetCodeRoomCantEntry:
		return errCode.TBossRoomCantEnter
	case helper2.RetCodeReadyFailed:
		return errCode.TBossReadyFailed
	}
	return 0
}
