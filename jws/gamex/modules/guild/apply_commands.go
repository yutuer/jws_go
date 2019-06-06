package guild

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	warnCode "vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (aw *ApplyWorker) newGuild(c *applyCommand) {
	// 名字检查
	if len(c.BaseInfo.Name) > 64 {
		logs.Error("Guild name size to long %d", len(c.BaseInfo.Name))
		c.resChan <- genErrRes(Err_CODE_ERR_Name_Len)
		return
	}

	// 检查敏感词
	if gamedata.CheckSymbol(c.BaseInfo.Name) || gamedata.CheckSensitive(c.BaseInfo.Name) {
		c.resChan <- genWarnRes(warnCode.GuildWordIllegal)
		return
	}

	// 检查名字重复
	err, isHas := checkGuildName(c.BaseInfo.Name, c.Applicant.AccountID)
	if err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	if isHas {
		c.resChan <- genWarnRes(warnCode.GuildNameRepeat)
		return
	}

	// 检查玩家是否已经在公会中
	if errRes := checkPlayerInGuild(c.Applicant.AccountID); errRes.ret.HasError() {
		c.resChan <- errRes
		return
	}

	// new guild
	c.Applicant.GuildPosition = gamedata.Guild_Pos_Chief
	_g := newGuildInfo(&c.BaseInfo, &c.Applicant)
	if _g == nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	// 清此玩家所有申请
	playerApply := aw.playerApply[c.Applicant.AccountID]
	_, chgGa, _, _, _ := aw._delPlayerAllApply(playerApply)
	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := playerApply.DBDel(cb); err != nil {
			return err
		}
		for _, ga := range chgGa {
			if err := ga.DBSave(cb); err != nil {
				return err
			}
		}
		return nil
	})
	if errCode != 0 {
		logs.Error("newGuild but del apply failed, acid=%s", c.Applicant.AccountID)
	}
	assignID, assignTime := getAssignInfo(c.AssignID, c.AssignTimes, c.LastLeaveTime)
	_g.Inventory.SetAssignTimesByAcID(c.Applicant.AccountID, assignID, assignTime)

	// log
	logicLog(c.Applicant.AccountID, c.Channel, logiclog.LogicTag_GuildCreate, _g, c.Applicant.AccountID, "", "")

	res := guildCommandRes{}
	res.guildInfo = *_g
	c.resChan <- res
}

func (aw *ApplyWorker) getPlayerApplyList(c *applyCommand) {
	res := guildCommandRes{}
	pa := aw.playerApply[c.Applicant.AccountID]
	now_time := game.GetNowTimeByOpenServer(c.shard)
	pa.updateApply(now_time)
	res.playerApply = make([]PlayerApplyInfo2Client, 0, pa.ApplyNum)
	for i := pa.ApplyNum - 1; i >= 0; i-- {
		aply := pa.ApplyList[i]
		res.playerApply = append(res.playerApply, PlayerApplyInfo2Client{
			GuildUuid: aply.GuildUuid,
			ApplyTime: aply.ApplyTime,
		})
	}
	for i := 0; i < len(res.playerApply); i++ {
		aply := &res.playerApply[i]
		if ga, gOk := aw.guildApply[aply.GuildUuid]; gOk {
			aply.GuildName = ga.Guild.Name
			aply.GuildLvl = ga.Guild.Level
			aply.GuildNotice = ga.Guild.Notice
		}
	}
	// save db
	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := pa.DBSave(cb); err != nil {
			return err
		}
		return nil
	})
	if errCode != 0 {
		logs.Error("guild ApplyWorker getPlayerApplyList savedb errcode %d", errCode)
	}
	c.resChan <- res
	return
}

func (aw *ApplyWorker) getGuildApplyList(c *applyCommand) {
	res := guildCommandRes{}
	ga, gOk := aw.guildApply[c.BaseInfo.GuildUUID]
	if !gOk || ga.Guild.GuildID <= 0 {
		c.resChan <- genWarnRes(warnCode.GuildNotFound)
	} else {
		now_time := game.GetNowTimeByOpenServer(c.shard)
		ga.updateApply(now_time)
		res.guildApply = make([]GuildApplyInfo, 0, ga.ApplyNum)
		for i := ga.ApplyNum - 1; i >= 0; i-- {
			aply := ga.ApplyList[i]
			res.guildApply = append(res.guildApply, GuildApplyInfo{
				ApplyTime:  aply.ApplyTime,
				PlayerInfo: aply.PlayerInfo,
			})
		}
		// save db
		errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
			if err := ga.DBSave(cb); err != nil {
				return err
			}
			return nil
		})
		if errCode != 0 {
			logs.Error("guild ApplyWorker getGuildApplyList savedb errcode %d", errCode)
		}
	}
	c.resChan <- res
	return
}

func (aw *ApplyWorker) applyGuild(c *applyCommand) {
	res := guildCommandRes{}
	guildApply, gOk := aw.guildApply[c.BaseInfo.GuildUUID]
	if !gOk || guildApply.Guild.GuildID <= 0 {
		c.resChan <- genWarnRes(warnCode.GuildNotFound)
	} else {
		playerApply := aw.playerApply[c.Applicant.AccountID]
		now_time := game.GetNowTimeByOpenServer(c.shard)
		_updateApply(playerApply, guildApply, now_time)
		// 检查玩家自己的申请
		if errCode := playerApply.checkPlayerGuildApply(c.BaseInfo.GuildUUID,
			now_time); errCode != 0 {
			logs.Warn("applyGuild err %d", errCode)
			c.resChan <- genWarnRes(warnCode.ClickTooQuickly)
			return
		}
		// 检查玩家是否已经在公会中
		if errRes := checkPlayerInGuild(c.Applicant.AccountID); errRes.ret.HasError() {
			c.resChan <- errRes
			return
		}
		// 公会满
		if guildApply.Guild.MemNum >= guildApply.Guild.MaxMemNum {
			c.resChan <- genWarnRes(warnCode.GuildFull)
			return
		}
		// 公会申请列表检查
		if guildApply.ApplyNum >= MaxGuildApply {
			c.resChan <- genWarnRes(warnCode.GuildApplyFull)
			return
		}
		// 申请gs是否满足条件
		if guildApply.Guild.ApplyGsLimit > 0 &&
			guildApply.Guild.ApplyGsLimit > c.Applicant.CurrCorpGs {
			c.resChan <- genWarnRes(warnCode.GuildApplyGsNotEnough)
			return
		}

		if guildApply.Guild.ApplyAuto {
			guildInfo, errRet := aw.m.GetGuildInfo(c.BaseInfo.GuildUUID)
			if errRet.HasError() {
				c.resChan <- genWarnRes(warnCode.GuildNotFound)
				return
			}
			// 加入
			pres, _ := aw._delApply(c, playerApply, guildApply)
			if pres != nil && pres.ret.HasError() {
				c.resChan <- *pres
				return
			}
			c.Applicant.SetOnline(true)
			pres = aw._addMem(c, &c.Applicant, guildApply, guildInfo, "")
			if pres != nil && pres.ret.HasError() {
				c.resChan <- *pres
				return
			}
			res.isApplySyncGuild = true
		} else {
			// 存缓存
			_addApply(playerApply, guildApply, now_time, c.BaseInfo.GuildUUID, c.Applicant, c.AssignID,
				c.AssignTimes, c.LastLeaveTime)
			// save db
			errCode := aw._savedb(playerApply, guildApply)
			if errCode != 0 {
				c.resChan <- genErrRes(errCode)
				return
			}

			// 红点通知
			aw.m.noticeHasApply(c.BaseInfo.GuildUUID, true)
		}
		c.resChan <- res

		return
	}
}

func (aw *ApplyWorker) cancelApplyGuild(c *applyCommand) {
	res := guildCommandRes{}
	guildApply, gOk := aw.guildApply[c.BaseInfo.GuildUUID]

	// 检查玩家是否已经在公会中
	if errRes := checkPlayerInGuild(c.Applicant.AccountID); errRes.ret.HasError() {
		c.resChan <- errRes
		return
	}

	if !gOk || guildApply.Guild.GuildID <= 0 {
		c.resChan <- genWarnRes(warnCode.GuildNotFound)
	} else {
		playerApply := aw.playerApply[c.Applicant.AccountID]
		// 刷新申请列表
		now_time := game.GetNowTimeByOpenServer(c.shard)
		_updateApply(playerApply, guildApply, now_time)
		// 检查玩家自己的申请, 存在则继续
		if playerApply.hasPlayerGuildApply(c.BaseInfo.GuildUUID, now_time) {
			// 删除申请
			_delApply(playerApply, guildApply, c.BaseInfo.GuildUUID, c.Applicant.AccountID)
			// 存db

			errCode := aw._savedb(playerApply, guildApply)
			if errCode != 0 {
				c.resChan <- genErrRes(errCode)
				return
			}
			// 红点通知
			aw.m.noticeHasApply(c.BaseInfo.GuildUUID, guildApply.ApplyNum > 0)
		}
		c.resChan <- res
		return
	}
}

func (aw *ApplyWorker) delApply(c *applyCommand) {
	res := guildCommandRes{}
	guildApply, gOk := aw.guildApply[c.BaseInfo.GuildUUID]
	// 检查玩家是否已经在公会中
	if errRes := checkPlayerInGuild(c.Applicant.AccountID); errRes.ret.HasError() {
		c.resChan <- errRes
		return
	}

	if !gOk || guildApply.Guild.GuildID <= 0 {
		c.resChan <- genWarnRes(warnCode.GuildNotFound)
	} else {

		playerApply := aw.playerApply[c.Applicant.AccountID]
		now_time := game.GetNowTimeByOpenServer(c.shard)
		_updateApply(playerApply, guildApply, now_time)
		// 检查玩家自己的申请, 存在则继续
		if playerApply.hasPlayerGuildApply(c.BaseInfo.GuildUUID, now_time) {
			// 检查审批者
			guildInfo, errRet := aw.m.GetGuildInfo(c.BaseInfo.GuildUUID)
			if errRet.HasError() {
				c.resChan <- genWarnRes(warnCode.GuildNotFound)
				return
			}
			appoveMem := guildInfo.GetGuildMemInfo(c.Approver.AccountID)
			if appoveMem == nil || !gamedata.CheckApprovePosition(appoveMem.GuildPosition) {
				c.resChan <- genWarnRes(warnCode.GuildPositionErr)
				return
			}
			// 删除申请
			_delApply(playerApply, guildApply, c.BaseInfo.GuildUUID, c.Applicant.AccountID)
			// 存db
			errCode := aw._savedb(playerApply, guildApply)
			if errCode != 0 {
				c.resChan <- genErrRes(errCode)
				return
			}
			// 红点通知
			aw.m.noticeHasApply(c.BaseInfo.GuildUUID, guildApply.ApplyNum > 0)
			// 邮件通知
			mail_sender.SendGangApplyRefuse(c.Applicant.AccountID, guildApply.Guild.Name)
		}
		c.resChan <- res
		return
	}
}

// 审批公会申请者
func (aw *ApplyWorker) approveApply(c *applyCommand) {
	res := guildCommandRes{}
	guildApply, gOk := aw.guildApply[c.BaseInfo.GuildUUID]

	// 检查玩家是否已经在其他公会中
	if errRes := checkPlayerInGuild(c.Applicant.AccountID); errRes.ret.HasError() {
		c.resChan <- errRes
		return
	}

	if !gOk || guildApply.Guild.GuildID <= 0 {
		c.resChan <- genWarnRes(warnCode.GuildNotFound)
	} else {
		playerApply := aw.playerApply[c.Applicant.AccountID]
		now_time := game.GetNowTimeByOpenServer(c.shard)
		_updateApply(playerApply, guildApply, now_time)
		// 检查玩家自己的申请, 存在则继续
		if playerApply.hasPlayerGuildApply(c.BaseInfo.GuildUUID, now_time) {
			// 检查审批者
			guildInfo, errRet := aw.m.GetGuildInfo(c.BaseInfo.GuildUUID)
			if errRet.HasError() {
				c.resChan <- genWarnRes(warnCode.GuildNotFound)
				return
			}
			approveMem := guildInfo.GetGuildMemInfo(c.Approver.AccountID)
			if approveMem == nil {
				c.resChan <- genWarnRes(warnCode.GuildPositionErr)
				return
			}
			// 权限
			if !gamedata.CheckApprovePosition(approveMem.GuildPosition) {
				c.resChan <- genWarnRes(warnCode.GuildPositionErr)
				return
			}
			// 加入
			pres, applicantInfo := aw._delApply(c, playerApply, guildApply)
			if pres != nil && pres.ret.HasError() {
				c.resChan <- *pres
				return
			}
			pres = aw._addMem(c, applicantInfo, guildApply, guildInfo, approveMem.Name)
			if pres != nil && pres.ret.HasError() {
				c.resChan <- *pres
				return
			}
			// 红点通知
			aw.m.noticeHasApply(c.BaseInfo.GuildUUID, guildApply.ApplyNum > 0)
		} else {
			c.resChan <- genWarnRes(warnCode.GuildApplicantNotFound)
			return
		}

		c.resChan <- res
		return
	}
}

func (aw *ApplyWorker) _delApply(c *applyCommand,
	playerApply *PlayerApply, guildApply *GuildApply) (
	*guildCommandRes, *helper.AccountSimpleInfo) {

	// 删除被审批者身上的申请
	applyInfo, chgGa, assignID, assignTimes, lastLeaveTime := aw._delPlayerAllApply(playerApply)
	// 删除申请
	_delApply(playerApply, guildApply, c.BaseInfo.GuildUUID, c.Applicant.AccountID)
	// save db
	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := playerApply.DBDel(cb); err != nil {
			return err
		}
		if err := guildApply.DBSave(cb); err != nil {
			return err
		}
		for _, ga := range chgGa {
			if err := ga.DBSave(cb); err != nil {
				return err
			}
		}
		return nil
	})
	if errCode != 0 {
		res := genErrRes(errCode)
		return &res, applyInfo
	}
	if len(c.AssignID) <= 0 {
		c.AssignID = assignID
	}
	if len(c.AssignTimes) <= 0 {
		c.AssignTimes = assignTimes
	}
	if c.LastLeaveTime == 0 {
		c.LastLeaveTime = lastLeaveTime
	}
	logs.Debug("c.AssignTimes: %v| assignTimes: %V", c.AssignTimes, assignTimes)
	return nil, applyInfo
}
func (aw *ApplyWorker) _addMem(c *applyCommand,
	applicantInfo *helper.AccountSimpleInfo,
	guildApply *GuildApply,
	guildInfo *GuildInfo, approveName string) *guildCommandRes {

	// 检查玩家是否已经在公会中
	if errRes := checkPlayerInGuild(c.Applicant.AccountID); errRes.ret.HasError() {
		return &errRes
	}
	// 公会满
	if guildInfo.Base.MemNum >= guildInfo.Base.MaxMemNum {
		res := genWarnRes(warnCode.GuildFull)
		return &res
	}
	// 加公会
	assignID, assignTime := getAssignInfo(c.AssignID, c.AssignTimes, c.LastLeaveTime)
	ret, gi := aw.m.AddMem(c.BaseInfo.GuildUUID, applicantInfo, approveName, assignID, assignTime)
	if ret.HasError() {
		res := genErrRes(ret.ErrCode)
		return &res
	}
	guildApply.updateGuildInfo(gi)
	// log
	logicLog(c.Approver.AccountID, c.Channel, logiclog.LogicTag_GuildAddMem, guildInfo, c.Applicant.AccountID, "", "")
	return nil
}

func getAssignInfo(assignID []string, assignTime []int64, lastLeaveTime int64) ([]string, []int64) {
	var retID []string
	var retTime []int64
	logs.Debug("time now: %v, last next refresh time: %v", time.Now().Unix(),
		util.GetNextDailyTime(gamedata.GetGVGGuildGiftBeginSec(lastLeaveTime), lastLeaveTime))
	if time.Now().Unix() > util.GetNextDailyTime(gamedata.GetGVGGuildGiftBeginSec(lastLeaveTime), lastLeaveTime) {
		retID = make([]string, 0)
		retTime = make([]int64, 0)
	} else {
		retID = assignID
		retTime = assignTime
	}
	return retID, retTime
}

func (aw *ApplyWorker) updateAccountInfoApply(c *applyCommand) {
	playerApply := aw.playerApply[c.Applicant.AccountID]
	now_time := game.GetNowTimeByOpenServer(c.shard)
	chgGuild := aw._updatePlayerInfo(playerApply, &c.Applicant, now_time)

	// save db
	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		for _, g := range chgGuild {
			if err := g.DBSave(cb); err != nil {
				return err
			}
		}
		return nil
	})
	if errCode != 0 {
		logs.Error("guild ApplyWorker updateAccountInfoApply savedb errcode %d", errCode)
	}
}

func (aw *ApplyWorker) updateGuildInfoCmd(c *applyCommand) {
	res := guildCommandRes{}
	guildApply, gOk := aw.guildApply[c.BaseInfo.GuildUUID]
	if !gOk {
		c.resChan <- genWarnRes(warnCode.GuildNotFound)
	} else {
		guildApply.updateGuildInfo(c.BaseInfo)
		if guildApply.ApplyNum > 0 {
			// 红点通知
			aw.m.noticeHasApply(c.BaseInfo.GuildUUID, true)
		}
		// save db
		errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
			if err := guildApply.DBSave(cb); err != nil {
				return err
			}
			return nil
		})
		if errCode != 0 {
			logs.Error("guild ApplyWorker updateAccountInfoApply savedb errcode %d", errCode)
		}
		c.resChan <- res
		return
	}
}

func (aw *ApplyWorker) dismissGuildCallBack(c *applyCommand) {
	guid := c.BaseInfo.GuildUUID
	ga, gOk := aw.guildApply[guid]
	if !gOk {
		return
	}
	pas := make([]*PlayerApply, 0, ga.ApplyNum)
	for i := ga.ApplyNum - 1; i >= 0; i-- {
		aply := ga.ApplyList[i]
		pa := aw.playerApply[aply.PlayerInfo.AccountID]
		for j := pa.ApplyNum - 1; j >= 0; j-- {
			if pa.ApplyList[j].GuildUuid == guid {
				pa._delApply(j)
				pas = append(pas, pa)
				break
			}
		}
	}
	delete(aw.guildApply, guid)
	dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := ga.DBDel(cb); err != nil {
			return err
		}
		for _, pa := range pas {
			if err := pa.DBSave(cb); err != nil {
				return err
			}
		}
		return nil
	})
}

func (aw *ApplyWorker) _savedb(playerApply *PlayerApply, guildApply *GuildApply) int {
	return dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := playerApply.DBSave(cb); err != nil {
			return err
		}
		if err := guildApply.DBSave(cb); err != nil {
			return err
		}
		return nil
	})
}
