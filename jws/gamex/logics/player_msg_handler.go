package logics

import (
	"strconv"
	"vcs.taiyouxi.net/jws/gamex/logics/notify"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice/teamboss"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	"vcs.taiyouxi.net/jws/gamex/modules/gates_enemy"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) RegisterPlayerMsgHandler(r *servers.Mux) {
	r.HandleFunc("MSG/TEST", p.OnPlayerMsgTest)
	r.HandleFunc(player_msg.PlayerMsgGatesEnemyDataCode, p.OnPlayerMsgGatesEnemyData)
	r.HandleFunc(player_msg.PlayerMsgGuildInfoSyncCode, p.OnPlayerGuildUpdate)
	r.HandleFunc(player_msg.PlayerMsgGuildBossSyncCode, p.OnGuildBossUpdate)
	r.HandleFunc(player_msg.PlayerMsgGuildApplyInfoSyncCode, p.OnPlayerGuildApplyUpdate)
	r.HandleFunc(player_msg.PlayerMsgGuildScienceInfoSyncCode, p.OnGuildScienceUpdate)
	r.HandleFunc(player_msg.PlayerMsgGuildRedPacketSyncCode, p.OnGuildRedPacketUpdate)
	r.HandleFunc(player_msg.PlayerMsgRedPoint, p.OnPlayerRedPoint)
	r.HandleFunc(player_msg.PlayerMsgGank, p.OnPlayerGank)
	r.HandleFunc(player_msg.PlayerMsgGVEGameStartCode, p.OnPlayerMsgGVEStart)
	r.HandleFunc(player_msg.PlayerMsgGVEGameStopCode, p.OnPlayerMsgGVEStop)
	r.HandleFunc(player_msg.PlayerMsgTeamPvpRankChgCode, p.OnPlayerTeamPvpRankChg)
	r.HandleFunc(player_msg.PlayerMsgTitleCode, p.OnPlayerTitle)
	r.HandleFunc(player_msg.PlayerMsgRooms, p.OnPlayerRoom)
	r.HandleFunc(player_msg.PlayerMsgRoomEvent, p.OnPlayerRoomEvent)
	r.HandleFunc(player_msg.PlayerMsgGVGStartCode, p.OnPlayerMsgGVGStart)
	r.HandleFunc(player_msg.PlayerMsgOnLoginCode, p.OnPlayerLogin)
	r.HandleFunc(player_msg.PlayerMsgCSRobSetFormation, p.OnPlayerCSRobSetFormation)
	r.HandleFunc(player_msg.PlayerMsgCSRobRefreshSelf, p.OnPlayerCSRobRefreshPlayerCache)
	r.HandleFunc(player_msg.PlayerMsgCSRobAddPlayerRank, p.OnPlayerCSRobAddPlayerRank)
	r.HandleFunc(player_msg.PlayerMsgTeamBossKicked, p.OnPlayerTeamBossKicked)
	r.HandleFunc(player_msg.PlayerMsgRefreshRoom, p.OnPlayerRefreshRoomInfo)
	r.HandleFunc(player_msg.PlayerMsgTeamStartFight, p.OnPlayerStartFight)
	r.HandleFunc(player_msg.Playerh5MsgCostDiamond, p.OnH5CostDiamond)
}

func (p *Account) OnPlayerMsgTest(r servers.Request) *servers.Response {
	req := struct {
		AA string `codec:"aa"`
		BB string `codec:"bb"`
		CC string `codec:"cc"`
		DD string `codec:"dd"`
	}{}

	decode(r.RawBytes, &req)
	logs.Error("req OnPlayerMsg %v", req)

	return nil
}

// 此用来将登陆时，比较费的操作用msg改成异步执行，目的在于在特殊情况下也不阻止玩家登陆
func (p *Account) OnPlayerLogin(r servers.Request) *servers.Response {
	acid := p.AccountID.String()
	simpleInfo := p.GetSimpleInfo()
	profile := &p.Profile
	guild.GetModule(p.AccountID.ShardId).UpdateAccountInfo(simpleInfo)
	p.GuildProfile.OnAfterLogin(p.AccountID, p.Profile.GetClientTagInfo())
	guid := guild.GetPlayerGuild(acid)
	if guid != "" {
		gates_enemy.GetModule(p.AccountID.ShardId).OnPlayerIntoAct(acid,
			p.GuildProfile.GuildUUID, simpleInfo, nil)
	}
	profile.GetTeamPvp().SyncRank(p.Account)
	profile.GetGank().OnAfterLogin(p.AccountID.String())
	logs.Debug("%s Msg OnPlayerLogin", p.AccountID.String())
	return nil
}

func (p *Account) OnPlayerMsgGatesEnemyData(r servers.Request) *servers.Response {
	req := player_msg.PlayerMsgGatesEnemyData{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerMsgGatesEnemyData %v", req)

	p.Account.Profile.GetGatesEnemy().OnPushData(
		p.AccountID.ShardId,
		p.AccountID.String(),
		p.GuildProfile.GetCurrGuildUUID(),
		p.Account.GetSimpleInfo(),
		&req,
	)

	respPush := notify.NotifySyncMsg{}
	respPush.OnChangeGatesEnemyPushData()
	sendPush("NOTIFY/GatesEnemyNotify", respPush, p)
	return nil
}

func (p *Account) OnPlayerGuildUpdate(r servers.Request) *servers.Response {
	req := player_msg.PlayerGuildInfoUpdate{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerGuildUpdate %v", req)

	if req.GuildUUID == "" {
		p.Account.Profile.GetGatesEnemy().CleanDatas()
	}

	p.GuildProfile.SyncGuildInfo(
		p.AccountID.String(),
		req.GuildUUID,
		req.GuildName,
		req.GuildPosition,
		req.LeaveTime,
		req.NextJoinTime,
		p.Profile.GetClientTagInfo(),
		req.AssignID,
		req.AssignTimes)

	respPush := notify.NotifySyncMsg{}
	respPush.OnChangePlayerGuild()
	respPush.OnChangeGuildInfo()
	respPush.OnChangeGuildMemsInfo()
	sendPush("NOTIFY/GuildInfoNotify", respPush, p)
	return nil
}

func (p *Account) OnGuildBossUpdate(r servers.Request) *servers.Response {
	req := player_msg.PlayerGuildBossUpdate{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnGuildBossUpdate %v", req)

	p.Profile.GetGuildBossInfo().Info = &req.Info
	respPush := notify.NotifySyncMsg{}
	respPush.SetNeedSyncGuildActBoss()
	sendPush("NOTIFY/GuildInfoNotify", respPush, p)
	return nil
}

func (p *Account) OnPlayerGuildApplyUpdate(r servers.Request) *servers.Response {
	req := player_msg.PlayerGuildApplyUpdate{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerGuildApplyUpdate %v", req)

	p.GuildProfile.HasApplyCanApprove = req.HasApplyCanApprove
	respPush := notify.NotifySyncMsg{}
	sendPush("NOTIFY/GuildApplyNotify", respPush, p)
	return nil
}

func (p *Account) OnGuildScienceUpdate(r servers.Request) *servers.Response {
	respPush := notify.NotifySyncMsg{}
	respPush.OnChangeGuildScience()
	sendPush("NOTIFY/GuildScienceNotify", respPush, p)
	return nil
}

func (p *Account) OnGuildRedPacketUpdate(r servers.Request) *servers.Response {
	respPush := notify.NotifySyncMsg{}
	respPush.OnChangeGuildRedPacket()
	sendPush("NOTIFY/GuildRedPacketNotify", respPush, p)
	return nil
}

func (p *Account) OnPlayerRedPoint(r servers.Request) *servers.Response {
	req := player_msg.PlayerRedPoint{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerRedPoint %v", req)

	respPush := notify.NotifySyncMsg{}
	respPush.OnChangeRedPoint(req.RedPointId)
	sendPush("NOTIFY/RedPointNotify", respPush, p)
	return nil
}

func (p *Account) OnPlayerGank(r servers.Request) *servers.Response {
	req := player_msg.PlayerGank{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerGank %v", req)

	p.Profile.GetGank().SetNewestLogTS(req.LogTS)
	respPush := notify.NotifySyncMsg{}
	respPush.OnChangeRedPoint(notify.RedPointTyp_Gank)
	sendPush("NOTIFY/RedPointNotify", respPush, p)
	return nil
}

func (p *Account) OnPlayerTeamPvpRankChg(r servers.Request) *servers.Response {
	req := player_msg.PlayerMsgTeamPvpRankChg{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerTeamPvpRankChg %v", req)

	p.Profile.GetTeamPvp().SetRank(req.Rank, true)
	p.Profile.GetFirstPassRank().OnRank(
		gamedata.FirstPassRankTypTeamPvp,
		req.Rank)

	respPush := notify.NotifySyncMsg{}
	respPush.OnChangeTeamPvp()
	sendPush("NOTIFY/TeamPvpNotify", respPush, p)
	return nil
}

func (p *Account) OnPlayerTitle(r servers.Request) *servers.Response {
	respPush := notify.NotifySyncMsg{}
	respPush.OnChangeTitle()
	sendPush("NOTIFY/TitleNotify", respPush, p)
	return nil
}

func (p *Account) OnPlayerRoom(r servers.Request) *servers.Response {
	if r.RawBytes != nil {
		respPush := notify.NotifySyncMsg{}
		respPush.SyncRoom = r.RawBytes
		sendPush("NOTIFY/Room", respPush, p)
	}
	return nil
}

func (p *Account) OnPlayerRoomEvent(r servers.Request) *servers.Response {
	if r.RawBytes != nil {
		respPush := notify.NotifySyncMsg{}
		respPush.SyncRoomEvent = r.RawBytes
		sendPush("NOTIFY/RoomEvent", respPush, p)
	}
	return nil
}

func (p *Account) OnPlayerCSRobSetFormation(r servers.Request) *servers.Response {
	req := player_msg.PlayerCSRobSetFormation{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerCSRobSetFormation %v", req)

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return nil
	}

	refreshFormation(p, player)
	logs.Trace("[CSRob] OnPlayerCSRobSetFormation {%v}", player.GetFormation())
	return nil
}

//OnPlayerCSRobRefreshPlayerCache ..
func (p *Account) OnPlayerCSRobRefreshPlayerCache(r servers.Request) *servers.Response {
	csrob.GetModule(p.AccountID.ShardId).PlayerMod.RefreshPlayerCacheBySelf(
		p.AccountID.String(),
		p.GuildProfile.GuildUUID,
		p.Profile.Name,
		p.GuildProfile.GetCurrPosition(),
	)
	return nil
}

//OnPlayerCSRobAddPlayerRank ..
func (p *Account) OnPlayerCSRobAddPlayerRank(r servers.Request) *servers.Response {
	p.updateCSRobPlayerRank(true)
	return nil
}

func (p *Account) OnPlayerTeamBossKicked(r servers.Request) *servers.Response {
	req := teamboss.Msg{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerTeamBossKicked %v", req)
	req.SetAddr(teamboss.Prefix)
	p.Account.SendRespByPush(&req)

	data := helper.LeaveRoomParam{}
	decode(req.PushMsg, &data)
	logs.Debug("OnPlayerTeamBossKicked: %v", data)
	p.Profile.GetTeamBossTeamInfo().TeamBossLeaveInfo.SetTeamBossLeaveInfo(1, data.RoomID, data.LeaveTime)
	return nil
}

func (p *Account) OnPlayerRefreshRoomInfo(r servers.Request) *servers.Response {
	req := teamboss.Msg{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerRefreshRoomInfo %v", req)
	req.SetAddr(teamboss.Prefix)
	p.Account.SendRespByPush(&req)
	return nil
}

func (p *Account) OnPlayerStartFight(r servers.Request) *servers.Response {
	req := teamboss.Msg{}
	decode(r.RawBytes, &req)
	logs.Trace("req OnPlayerStartFight %v", req)
	req.SetAddr(teamboss.Prefix)
	p.Account.SendRespByPush(&req)
	data := teamboss.ParamPlayerStart{}
	decode(req.PushMsg, &data)
	logs.Debug("OnPlayerStartFight: %v", data)
	p.Profile.GetTeamBossTeamInfo().GlobalRoomId = data.GlobalRoomID
	return nil
}

// 返回resp给playerMsg的发起者
func (p *Account) sendPlayerMsgResp(resp *servers.Response) {
	player_msg.SendResponseBySync(p.AccountID.String(), resp)
}
func (p *Account) OnH5CostDiamond(r servers.Request) *servers.Response {
	req := player_msg.PlayerCostDiamondInfo{}
	decode(r.RawBytes, &req)
	logs.Trace("req PlayerCostDiamondInfo %v", req)
	num, err := strconv.Atoi(req.DiamNum)
	logs.Debug("I will reomve %d diamond", num)
	if err != nil {
		logs.Error("Convert number from string to int err: %v", err)
	}
	ok := p.Profile.HC.UseHcGiveFirst(req.Roleid, int64(num),
		p.GetProfileNowTime(), "h5shop")
	logs.Debug("will try to Return")

	//rspBytes := encode()
	toRes := "0"
	if ok {
		logs.Debug("Cost Diamond Success")
		respPush := notify.NotifySyncMsg{}
		respPush.OnChangePlayerHc()
		sendPush("NOTIFY/PlayerHcNotify", respPush, p)
		toRes = "1"
	} else {
		logs.Debug("Cost Diamond Fail")
	}
	nowHc := p.Profile.HC.GetHC()
	HcS := strconv.Itoa(int(nowHc))
	p.sendPlayerMsgResp(&servers.Response{
		toRes,
		[]byte(HcS),
		false,
	})
	logs.Debug("OnH5CostDiamond End")
	return nil
}
