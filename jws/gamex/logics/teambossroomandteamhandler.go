package logics

import (
	"encoding/json"

	"strings"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	helper2 "vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice/teamboss"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GetTBTeamList : 获取组队BOSS的队伍列表
// 获取组队BOSS的队伍信息
func (p *Account) GetTBTeamListHandler(req *reqMsgGetTBTeamList, resp *rspMsgGetTBTeamList) uint32 {
	diff := uint32(req.DifficultyId)
	//条件检查
	if diff < account.TEAM_BOSS_DIFF_1 || diff > account.TEAM_BOSS_DIFF_5 {
		logs.Info("<TBoss get>teamboss room list diff err")
	}

	//跨服房间列表结构体
	team := &helper.RoomListInfo{
		RoomLevel: diff,
	}

	teamList, code, err := teamboss.GetRoomList(p.AccountID.ShardId, p.AccountID.String(), team)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> get team list crossservice err %v", err)
		return errCode.CommonInner
	}
	if teamList.Code != 0 {
		return CrossError2ClientError(teamList.Code)
	}
	resp.TeamList = make([][]byte, 0)
	count := 0
	for count < len(teamList.List) {
		tempAva := make([]int64, 0)
		tempAvaStar := make([]int64, 0)
		for _, ava := range teamList.List[count].BattleAvatar {
			tempAva = append(tempAva, int64(ava.Avatar))
			tempAvaStar = append(tempAvaStar, int64(ava.StarLv))
		}
		//如果人满了 状态为full
		var state int64
		state = int64(teamList.List[count].RoomStatus)
		if teamList.List[count].PlayerCount >= helper.RoomPlayerMaxCount {
			state = int64(account.TEAM_SETTING_FULL)
		}
		myTbTeamSimp := TBTeamSimple{
			TeamId:           teamList.List[count].RoomID,
			TeamMemberCount:  int64(teamList.List[count].PlayerCount),
			LeaderPlayerName: teamList.List[count].LeadName,
			FightAvatarIds:   tempAva,
			AvatarStarLevel:  tempAvaStar,
			TeamState:        state,
			LeaderPlayerSid:  int64(teamList.List[count].ServerID),
		}
		resp.TeamList = append(resp.TeamList, encode(myTbTeamSimp))
		count++
	}
	nowDay := int64(time.Now().Weekday())
	//星期日程序中为0，表中为7
	if nowDay == 0 {
		resp.TBTeamDayOfWeek = nowDay + 7 - 1
	} else {
		resp.TBTeamDayOfWeek = nowDay - 1
	}
	return 0
}

// TBTeamReady : 在队中准备
// 在队中准备
func (p *Account) TBTeamReadyHandler(req *reqMsgTBTeamReady, resp *rspMsgTBTeamReady) uint32 {
	teamId := req.ReadyTeamId
	nowStatus := int(req.ReadyNowStatus)
	diff := uint32(req.ReadyTeamDiff)

	if nowStatus == account.TEAM_UNREADY {
		//判断玩家金币是否足够
		if !p.Profile.GetSC().HasSC(gamedata.SC_Money, int64(gamedata.GetTBossDiffMap()[diff].GetGoldCost())) {
			logs.Info("<TBoss> players gold is now enough,gold: %v", p.Profile.GetSC())
			return errCode.CommonLessMoney
		}
		nowStatus = account.TEAM_READY
	} else if nowStatus == account.TEAM_READY {
		nowStatus = account.TEAM_UNREADY
	}
	readyInfo := &helper.ReadyFightInfo{
		RoomID: teamId,
		AcID:   p.AccountID.String(),
		Status: nowStatus,
	}
	ret, code, err := teamboss.ReadyFight(p.AccountID.ShardId, p.AccountID.String(), readyInfo)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> ready team crossservice err %v", err)
		return errCode.CommonInner
	}
	if ret.Code != 0 {
		if nowStatus == account.TEAM_UNREADY {
			return 0
		} else {
			return CrossError2ClientError(ret.Code)
		}
	}
	return 0
}

//压缩属性和战力
func (p *Account) compressAttrAndGs(attr *gamedata.AvatarAttr, avatarGs, avatarId int, diff uint32, a gs.DataToCalculateGS) (helper2.AvatarAttr_, int) {
	acfg := gamedata.GetHeroData(avatarId)
	star := p.Profile.GetHero().GetStar(avatarId)
	scfg := acfg.LvData[int(star)]
	ADHgs := uint32(attr.GS_OnlyATKDEF(scfg.Cfg) + 0.5)
	compressAttr := attr.GetCompressAttr(diff, ADHgs)
	logs.Debug("<TBoss> compress attr is %v", compressAttr)

	compressGs := int(compressAttr.GS_2(scfg.Cfg) + 0.5)
	talentGS := gamedata.GetHeroCommonConfig().GetHSPointGS()
	cost_talent := a.GetHeroTalentPointCost(avatarId)
	compressGs += int(cost_talent * talentGS)
	logs.Debug("<TBoss> compress GS is %v", compressGs)

	return compressAttr.AvatarAttr_, compressGs
}

// CreatTBTeam : 创建组队BOSS队伍
// 创建组队BOSS队伍
func (p *Account) CreatTBTeamHandler(req *reqMsgCreatTBTeam, resp *rspMsgCreatTBTeam) uint32 {
	diff := uint32(req.CreateDifficultyId)
	//条件检查
	if diff < account.TEAM_BOSS_DIFF_1 || diff > account.TEAM_BOSS_DIFF_5 {
		logs.Info("<TBoss> create tb team diff error")
	}

	//分别选出四种种类最强的武将
	bestHeroInfo := p.GetBestHeroByType()
	sid := p.AccountID.ServerString()
	newSid := strings.Split(sid, ":")
	playerDetailInfo := WSPVPRankInfo{
		Sid:               newSid[1],
		Name:              p.Profile.Name,
		VipLevel:          int64(p.Profile.Vip.V),
		CorpLevel:         int64(p.Profile.GetCorp().Level),
		AllGs:             int64(p.Profile.GetData().CorpCurrGS),
		GuildName:         p.Account.GuildProfile.GuildName,
		BestHeroIdx:       make([]int64, 4),
		BestHeroLevel:     make([]int64, 4),
		BestHeroStarLevel: make([]int64, 4),
		BestHeroBaseGs:    make([]int64, 4),
		BestHeroExtraGs:   make([]int64, 4),
		EquipAttr:         make([]int64, 3),
		DestinyAttr:       make([]int64, 3),
		JadeAttr:          make([]int64, 3),
	}
	for i, best := range bestHeroInfo {
		playerDetailInfo.BestHeroIdx[i] = int64(best.HeroId)
		playerDetailInfo.BestHeroLevel[i] = int64(best.HeroLevel)
		playerDetailInfo.BestHeroStarLevel[i] = int64(best.HeroStarLevel)
		playerDetailInfo.BestHeroBaseGs[i] = int64(best.HeroBaseGs)
		playerDetailInfo.BestHeroExtraGs[i] = int64(best.HeroGS)
	}

	_, _, _, _, _, extraAttrs := gs.GetCurrAttrForWspvp(account.NewAccountGsCalculateAdapter(p.Account))

	for i := 0; i < 3; i++ {
		playerDetailInfo.EquipAttr[i] = int64(extraAttrs[0][i])
		playerDetailInfo.DestinyAttr[i] = int64(extraAttrs[1][i])
		playerDetailInfo.JadeAttr[i] = int64(extraAttrs[2][i])
	}

	playerDetailEncodeInfo := encode(playerDetailInfo)

	//房主信息
	joinInfo := helper.PlayerJoinInfo{
		PlayerDetailInfo: playerDetailEncodeInfo,
		AcID:             p.AccountID.String(),
		GS:               p.Profile.GetData().CorpCurrGS,
		Avatar:           p.Profile.GetCurrAvatar(),
		Name:             p.Profile.Name,
		Level:            p.Profile.Level,
		VIP:              int(p.Profile.Vip.V),
		Sid:              p.AccountID.ShardId,
	}

	teaminfo := &helper.CreateRoomInfo{
		RoomLevel: diff,
		JoinInfo:  joinInfo,
	}

	//判断如果玩家等级不够不能创建队伍
	playerLv := p.Profile.GetCorp().GetLvlInfo()
	needLv := gamedata.GetTBossDiffMap()[diff-1].GetPlayerNeedLevel()
	if playerLv < needLv {
		return errCode.TBossNeedLvIsNotEnough
	}
	teamInfo, code, err := teamboss.CreateRoom(p.AccountID.ShardId, p.AccountID.String(), teaminfo)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> create team crossservice err %v", err)
		return errCode.TBossRoomIsNotExist
	}
	if teamInfo.Code != 0 {
		return CrossError2ClientError(teamInfo.Code)
	}

	p.Profile.GetTeamBossTeamInfo().NowTeamID = teamInfo.Info.RoomID

	members := make([][]byte, 0)

	leader := TeamMember{
		PosId:         0,
		HeroId:        -1,
		TeamMemAcid:   p.AccountID.String(),
		TBMemLevel:    int64(p.Profile.Level),
		TeamMemGs:     int64(p.Profile.GetData().CorpCurrGS),
		TeamMemName:   p.Profile.Name,
		TeamMemVIPLvl: int64(p.Profile.Vip.V),
		TeamMemIcon:   int64(p.Profile.GetCurrAvatar()),
		TeamMemSid:    int64(p.AccountID.ShardId),
	}

	member := TeamMember{
		HeroId: -1,
	}

	members = append(members, encode(leader), encode(member))

	tbDetail := TBTeamDetail{
		MyTeamId:         teamInfo.Info.RoomID,
		MyTeamDifficulty: int64(teamInfo.Info.Level),
		MyTeamSetting:    account.TEAM_SETTING_OPEN,
		TBTeamTypeId:     int64(teamInfo.Info.TeamTypID),
		LevelInfoId:      teamInfo.Info.SceneID,
		TBWbossId:        teamInfo.Info.BossID,
		RedBoxTickState:  account.TEAM_REDBOX_UNTICK,
		MyTeamMember:     members,
	}
	resp.MyTeamInfo = encode(tbDetail)
	return 0
}

// TBTeamJoinSetting : 队中邀请加入设定 0开放 1仅限邀请
// 队中邀请加入设定
func (p *Account) TBTeamJoinSettingHandler(req *reqMsgTBTeamJoinSetting, resp *rspMsgTBTeamJoinSetting) uint32 {
	newSetting := int(req.TeamJoinSetting)
	teamId := req.TeamSettingTeamId
	if newSetting != account.TEAM_SETTING_OPEN && newSetting != account.TEAM_SETTING_INV_ONLY {
		logs.Info("<TBoss> join setting error")
		return 0
	}

	status := &helper.ChangeRoomStatusInfo{
		RoomID:     teamId,
		RoomStatus: newSetting,
		AcID:       p.AccountID.String(),
	}

	ret, code, err := teamboss.ChangeRoomStatus(p.AccountID.ShardId, p.AccountID.String(), status)
	if code != crossservice.ErrOK {
		return errCode.CommonInner
		logs.Error("<TBoss> change join setting crossservice err %", err)
	}
	if ret.Code != 0 {
		return CrossError2ClientError(ret.Code)
	}

	resp.TeamJoinSetting = int64(ret.RoomStatus)
	return 0
}

// JoinTBTeam : 加入组队BOSS的队伍
// 加入组队BOSS的队伍
func (p *Account) JoinTBTeamHandler(req *reqMsgJoinTBTeam, resp *rspMsgJoinTBTeam) uint32 {
	teamId := req.JoinTeamId
	nowTime := p.Profile.GetProfileNowTime()
	leaderID := req.JoinLeaderId
	//条件判断
	if teamId == "" {
		resp.JoinTBTeamResult = 1
		return 0
	}
	var isInvited bool
	//是邀请的
	if leaderID == "" {
		isInvited = false
	} else {
		isInvited = true
	}
	if isInvited {
		diff, _, _ := helper.ParseRoomID(teamId)
		playerDiff := gamedata.GetHighestLevelByCorpLv(p.GetCorpLv())
		if playerDiff == 0 || diff < playerDiff {
			resp.JoinTBTeamResult = 4
			return 0
		}
	}
	if teamId != p.Profile.GetTeamBossTeamInfo().NowTeamID && p.Profile.GetTeamBossTeamInfo().NowTeamID != "" {
		leaveInfo := &helper.LeaveRoomInfo{
			OptAcID: p.AccountID.String(),
			TgtAcID: p.AccountID.String(),
			RoomID:  p.Profile.GetTeamBossTeamInfo().NowTeamID,
		}
		ret, code, err := teamboss.LeaveRoom(p.AccountID.ShardId, p.AccountID.String(), leaveInfo)
		if code != crossservice.ErrOK {
			logs.Error("<TBoss> on exit leave team crossservice err %v", err)
			resp.JoinTBTeamResult = 4
			return 0
		}
		if ret.Code != 0 {
			logs.Info("<TBoss> on exit leave team code err %v", ret.Code)
		}
		if ret.Param.IsRefresh {
			p.Profile.GetTeamBossTeamInfo().TeamBossLeaveInfo.SetTeamBossLeaveInfo(0, p.Profile.GetTeamBossTeamInfo().NowTeamID, nowTime)
			p.Profile.GetTeamBossTeamInfo().NowTeamID = ""
		}
	}

	if !p.Profile.GetTeamBossTeamInfo().TeamBossLeaveInfo.IsCanJoinTBossTeamNow(teamId, p.Profile.GetTeamBossTeamInfo().TeamBossLeaveInfo.LeaveType, nowTime) {
		resp.JoinTBTeamResult = 2
		return 0
	}

	bestHeroInfo := p.GetBestHeroByType()
	sid := p.AccountID.ServerString()
	newSid := strings.Split(sid, ":")
	playerDetailInfo := WSPVPRankInfo{
		Sid:               newSid[1],
		Name:              p.Profile.Name,
		VipLevel:          int64(p.Profile.Vip.V),
		CorpLevel:         int64(p.Profile.GetCorp().Level),
		AllGs:             int64(p.Profile.GetData().CorpCurrGS),
		GuildName:         p.Account.GuildProfile.GuildName,
		BestHeroIdx:       make([]int64, 4),
		BestHeroLevel:     make([]int64, 4),
		BestHeroStarLevel: make([]int64, 4),
		BestHeroBaseGs:    make([]int64, 4),
		BestHeroExtraGs:   make([]int64, 4),
		EquipAttr:         make([]int64, 3),
		DestinyAttr:       make([]int64, 3),
		JadeAttr:          make([]int64, 3),
	}
	for i, best := range bestHeroInfo {
		playerDetailInfo.BestHeroIdx[i] = int64(best.HeroId)
		playerDetailInfo.BestHeroLevel[i] = int64(best.HeroLevel)
		playerDetailInfo.BestHeroStarLevel[i] = int64(best.HeroStarLevel)
		playerDetailInfo.BestHeroBaseGs[i] = int64(best.HeroBaseGs)
		playerDetailInfo.BestHeroExtraGs[i] = int64(best.HeroGS)
	}

	_, _, _, _, _, extraAttrs := gs.GetCurrAttrForWspvp(account.NewAccountGsCalculateAdapter(p.Account))

	for i := 0; i < 3; i++ {
		playerDetailInfo.EquipAttr[i] = int64(extraAttrs[0][i])
		playerDetailInfo.DestinyAttr[i] = int64(extraAttrs[1][i])
		playerDetailInfo.JadeAttr[i] = int64(extraAttrs[2][i])
	}

	playerDetailEncodeInfo := encode(playerDetailInfo)

	playerInfo := helper.PlayerJoinInfo{
		PlayerDetailInfo: playerDetailEncodeInfo,
		AcID:             p.AccountID.String(),
		GS:               p.Profile.GetData().CorpCurrGS,
		Avatar:           p.Profile.GetCurrAvatar(),
		Name:             p.Profile.Name,
		VIP:              int(p.Profile.Vip.V),
		Sid:              p.AccountID.ShardId,
		IsInvited:        isInvited,
	}

	joinInfo := &helper.JoinRoomInfo{
		RoomID:   teamId,
		JoinInfo: playerInfo,
	}

	teamInfo, code, err := teamboss.JoinRoom(p.AccountID.ShardId, p.AccountID.String(), joinInfo)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> join team crossservice err %v", err)
		resp.JoinTBTeamResult = 3
		return 0
	}
	if teamInfo.Code != 0 {
		if teamInfo.Code == helper.RetCodeRoomCantEntry {
			resp.JoinTBTeamResult = 5
			return 0
		} else if teamInfo.Code == helper.RetCodeRoomNotExist {
			resp.JoinTBTeamResult = 1
			return 0
		} else {
			resp.JoinTBTeamResult = 3
			return 0
		}
	}
	mems := make([][]byte, 0)

	var leaderIndex int
	var memIndex int

	//决定谁是队长
	for i, player := range teamInfo.Info.SimpleInfo {
		if player.AcID == teamInfo.Info.LeadID {
			leaderIndex = i
		}
	}
	if leaderIndex == 1 {
		memIndex = 0
	} else if leaderIndex == 0 {
		memIndex = 1
	}

	simpleInfo := teamInfo.Info.SimpleInfo

	member1 := TeamMember{
		PosId:             int64(GetPositionByAcid(simpleInfo[leaderIndex].AcID, teamInfo.Info.PositionAcID)),
		HeroId:            int64(simpleInfo[leaderIndex].BattleAvatar),
		HeroMagicPet:      int64(simpleInfo[leaderIndex].MagicPet),
		HeroSwing:         int64(simpleInfo[leaderIndex].Wing),
		HeroFashion:       simpleInfo[leaderIndex].Fashion,
		HeroGlory:         simpleInfo[leaderIndex].ExclusiveWeapon,
		TeamMemAcid:       simpleInfo[leaderIndex].AcID,
		TBMemLevel:        int64(simpleInfo[leaderIndex].Level),
		TBMemHeroStar:     int64(simpleInfo[leaderIndex].StarLevel),
		TeamMemGs:         int64(simpleInfo[leaderIndex].GS),
		TeamMemName:       simpleInfo[leaderIndex].Name,
		TeamMemReadyState: int64(simpleInfo[leaderIndex].Status),
		TeamMemCompressGs: int64(simpleInfo[leaderIndex].CompressGS),
		TeamMemVIPLvl:     int64(simpleInfo[leaderIndex].VIP),
		TeamMemIcon:       int64(simpleInfo[leaderIndex].Avatar),
		TeamMemSid:        int64(simpleInfo[leaderIndex].Sid),
	}

	member2 := TeamMember{
		PosId:             int64(GetPositionByAcid(simpleInfo[leaderIndex].AcID, teamInfo.Info.PositionAcID)),
		HeroId:            int64(simpleInfo[memIndex].BattleAvatar),
		HeroMagicPet:      int64(simpleInfo[memIndex].MagicPet),
		HeroSwing:         int64(simpleInfo[memIndex].Wing),
		HeroFashion:       simpleInfo[memIndex].Fashion,
		HeroGlory:         simpleInfo[memIndex].ExclusiveWeapon,
		TeamMemAcid:       simpleInfo[memIndex].AcID,
		TBMemLevel:        int64(simpleInfo[memIndex].Level),
		TBMemHeroStar:     int64(simpleInfo[memIndex].StarLevel),
		TeamMemGs:         int64(simpleInfo[memIndex].GS),
		TeamMemName:       simpleInfo[memIndex].Name,
		TeamMemReadyState: int64(simpleInfo[memIndex].Status),
		TeamMemCompressGs: int64(simpleInfo[memIndex].CompressGS),
		TeamMemVIPLvl:     int64(simpleInfo[memIndex].VIP),
		TeamMemIcon:       int64(simpleInfo[memIndex].Avatar),
		TeamMemSid:        int64(simpleInfo[memIndex].Sid),
	}

	mems = append(mems, encode(member1), encode(member2))

	tbDetail := TBTeamDetail{
		MyTeamId:         teamInfo.Info.RoomID,
		MyTeamSetting:    int64(teamInfo.Info.RoomStatus),
		TBTeamTypeId:     int64(teamInfo.Info.TeamTypID),
		LevelInfoId:      teamInfo.Info.SceneID,
		TBWbossId:        teamInfo.Info.BossID,
		RedBoxTickState:  int64(teamInfo.Info.BoxStatus),
		MyTeamMember:     mems,
		MyTeamDifficulty: int64(teamInfo.Info.Level),
	}

	resp.JoinTBTeamResult = 0
	resp.MyTeamInfo = encode(tbDetail)
	p.Profile.GetTeamBossTeamInfo().NowTeamID = teamId
	return 0
}

//根据acid返回选将的位置
func GetPositionByAcid(acid string, posAcid [helper.RoomPlayerMaxCount]string) int {
	if acid != "" {
		if posAcid[account.TEAM_HERO_POS_LEFT] == acid {
			return account.TEAM_HERO_POS_LEFT
		} else if posAcid[account.TEAM_HERO_POS_RIGHT] == acid {
			return account.TEAM_HERO_POS_RIGHT
		}
	}
	return -1
}

// GetRedBoxCostHC : 房间中勾选花费钻石一定获得红宝箱
// 房间中勾选花费钻石一定获得红宝箱
func (p *Account) GetRedBoxCostHCHandler(req *reqMsgGetRedBoxCostHC, resp *rspMsgGetRedBoxCostHC) uint32 {
	boxStatus := int(req.IsTickRedBox) //0没勾选 1勾选
	teamId := req.TBCostTeamID
	if teamId == "" {
		return errCode.TBossRoomIsNotExist
	}
	advanceInfo := &helper.CostAdvanceInfo{
		RoomID:    teamId,
		BoxStatus: boxStatus,
		AcID:      p.AccountID.String(),
	}

	ret, code, err := teamboss.CostAdvance(p.AccountID.ShardId, p.AccountID.String(), advanceInfo)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> tick red box crossservice err %v", err)
		return errCode.TBossAreadyTickRedBox
	}
	if ret.Code != 0 {
		return CrossError2ClientError(ret.Code)
	}

	return 0
}

// GetTBMemberInfo : 获取组队BOSS的队友信息
// 获取组队BOSS的队友信息
func (p *Account) GetTBMemberInfoHandler(req *reqMsgGetTBMemberInfo, resp *rspMsgGetTBMemberInfo) uint32 {
	teamId := req.TeamMemTeamId
	tgtAcid := req.TeamMemAcid
	if tgtAcid == "" || teamId == "" {
		logs.Info("<TBoss> get team member info params error")
		return errCode.CommonInvalidParam
	}
	memDetail := &helper.GetPlayerDetailInfo{
		RoomID: teamId,
		AcID:   p.AccountID.String(),
		TgtID:  tgtAcid,
	}
	teamMemInfo, code, err := teamboss.GetPlayerDetail(p.AccountID.ShardId, p.AccountID.String(), memDetail)
	if code != crossservice.ErrOK {
		return errCode.CommonInner
	}
	if teamMemInfo.Code != 0 {
		logs.Info("<TBoss> get member info crossservice err %v", err)
		return CrossError2ClientError(teamMemInfo.Code)
	}
	resp.TBTeammate = teamMemInfo.Detail
	return 0
}

//TBChooseHero : 房间中选将界面确认选将
//房间中选将界面确认选将
func (p *Account) TBChooseHeroHandler(req *reqMsgTBChooseHero, resp *rspMsgTBChooseHero) uint32 {
	avaId := int(req.TbChooseHeroId)
	teamId := req.TbChooseHeroTeamId
	pos := int(req.TbChooseHeroPos)
	diff := uint32(req.TbChooseHeroDiff)
	if teamId == "" {
		return errCode.TBossRoomIsNotExist
	}
	var compressedGs int
	avaInfo := &helper.ChangeAvatarInfo{}
	if avaId >= 0 && avaId < helper2.AVATAR_NUM_MAX {
		accData := &helper2.Avatar2ClientByJson{}
		account.FromAccount2Json(accData, p.Account, avaId)
		accData.Attr, compressedGs = p.compressAttrAndGs(&gamedata.AvatarAttr{AvatarAttr_: accData.Attr}, accData.Gs, avaId, diff, account.NewAccountGsCalculateAdapter(p.Account))
		accData.HP = accData.Attr.HP
		accData.Gs = compressedGs
		detail, err := json.Marshal(accData)
		if err != nil {
			logs.Info("json marshal error by %v", err)
		}

		avaInfo.RoomID = teamId
		avaInfo.AcID = p.AccountID.String()
		avaInfo.BattleAvatar = avaId
		avaInfo.Wing = p.Profile.GetHero().GetSwing(avaId).CurSwing
		avaInfo.Fashion = p.getEquipFashionTids(avaId)
		avaInfo.MagicPet = int(p.Profile.GetHero().GetMagicPetFigure(avaId))
		avaInfo.StarLevel = int(p.Profile.GetHero().GetStar(avaId))
		avaInfo.Position = pos
		avaInfo.BattleInfo = detail
		avaInfo.CompressGs = compressedGs

		resp.ChooseHeroMagicPet = int64(p.Profile.GetHero().GetMagicPetFigure(avaId))
		resp.ChooseHeroSwing = int64(p.Profile.GetHero().GetSwing(avaId).CurSwing)
		resp.ChooseHeroFashion = p.getEquipFashionTids(avaId)
		resp.ChooseHeroPost = int64(pos)
		resp.ChooseHeroCompressGs = int64(compressedGs)

	} else {
		avaInfo.RoomID = teamId
		avaInfo.AcID = p.AccountID.String()
		avaInfo.BattleAvatar = avaId
		avaInfo.Position = pos

	}

	ret, code, err := teamboss.ChangeAvatar(p.AccountID.ShardId, p.AccountID.String(), avaInfo)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> choose hero crossservice err %v", err)
		return errCode.CommonInner
	}
	if ret.Code != 0 {
		return CrossError2ClientError(ret.Code)
	}
	resp.ChooseHeroId = int64(avaId)

	return 0
}

//LeaveTBTeam : 离开组队BOSS的队伍
//离开组队BOSS的队伍
func (p *Account) LeaveTBTeamHandler(req *reqMsgLeaveTBTeam, resp *rspMsgLeaveTBTeam) uint32 {
	teamId := req.LeaveTeamId
	if teamId == "" {
		logs.Debug("<TBoss>leave room id is nil")
	}
	//如果acid和目标acid一样是离开
	leaveInfo := &helper.LeaveRoomInfo{
		OptAcID: p.AccountID.String(),
		TgtAcID: p.AccountID.String(),
		RoomID:  teamId,
	}
	ret, code, err := teamboss.LeaveRoom(p.AccountID.ShardId, p.AccountID.String(), leaveInfo)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> leave team crossservice err %v", err)
	}
	if ret.Code != 0 {
		logs.Debug("cross server is error, code is %v", ret.Code)
	}

	if ret.Param.IsRefresh {
		//将离开信息存到profile
		p.Profile.GetTeamBossTeamInfo().TeamBossLeaveInfo.SetTeamBossLeaveInfo(account.TEAM_LEAVE_BY_SELF, ret.Param.RoomID, ret.Param.LeaveTime)
		p.Profile.GetTeamBossTeamInfo().NowTeamID = ""
	}
	return 0
}

// TBTeamKick : 踢出组队boss某人
// 踢出组队boss某人
func (p *Account) TBTeamKickHandler(req *reqMsgTBTeamKick, resp *rspMsgTBTeamKick) uint32 {
	kickAcid := req.BeKickedAcid
	teamid := req.BeKickedTeamId
	if kickAcid == "" {
		return errCode.TBossKickOneIsNotExist
	}
	kickInfo := &helper.LeaveRoomInfo{
		OptAcID: p.AccountID.String(),
		TgtAcID: kickAcid,
		RoomID:  teamid,
	}
	ret, code, err := teamboss.LeaveRoom(p.AccountID.ShardId, p.AccountID.String(), kickInfo)
	if code != crossservice.ErrOK {
		logs.Error("<TBoss> kick someone crossservice err %v", err)
		return errCode.CommonInner
	}
	if ret.Code != 0 {
		return CrossError2ClientError(ret.Code)
	}
	return 0
}
