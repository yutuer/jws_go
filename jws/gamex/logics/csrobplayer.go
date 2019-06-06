package logics

import (
	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/simple_info"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/distinct"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func csrobCheckTimeIn() bool {
	now := time.Now().Unix()
	nowDaily := now - util.DailyBeginUnix(now)
	if nowDaily < gamedata.CSRobDailyStartTime() || nowDaily > gamedata.CSRobDailyEndTime() {
		logs.Debug("[CSRob] csrobCheckTimeIn failed, now %d, start %d, end %d", nowDaily, gamedata.CSRobDailyStartTime(), gamedata.CSRobDailyEndTime())
		return false
	}
	return true
}

func makeTodayFormation(p *Account) []int {
	data := p.Profile.GetData()
	logs.Debug("[CSRob] refreshInfo data.HeroGs, {%v}", data.HeroGs)

	now := time.Now().Unix()
	_, nat := gamedata.CSRobBattleIDAndHeroID(now)
	heroList := map[int]int{}
	for idx, gs := range data.HeroGs {
		hero := gamedata.GetHeroData(idx)
		if nil == hero || gs < 1 || hero.Nationality != nat {
			continue
		}

		heroList[idx] = gs
	}

	list := util.SortIntMapKeyByValue(heroList)
	formation := list[:]
	if 3 < len(list) {
		formation = list[:3]
	}
	return formation
}

func refreshFormation(p *Account, player *csrob.Player) {
	formation := makeTodayFormation(p)
	//写入新的阵容
	team := p.buildHeroList(formation)
	player.SetFormation(formation, team)
}

//协议------押运粮草:取自己的数据
type reqMsgCSRobPlayerInfo struct {
	Req
}

type rspMsgCSRobPlayerInfo struct {
	Resp
	Info []byte `codec:"info"` //我的活动数据 CSRobPlayerInfo
}

//CSRobPlayerInfo ..
func (p *Account) CSRobPlayerInfo(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobPlayerInfo)
	rsp := new(rspMsgCSRobPlayerInfo)

	initReqRsp(
		"Attr/CSRobPlayerInfoRsp",
		r.RawBytes,
		req, rsp, p)

	//活动时间检查
	//if false == csrobCheckTimeIn() {
	//	return rpcWarn(rsp, errCode.CommonNotInTime)
	//}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	simpleInfo := p.Account.GetSimpleInfo()
	param := &csrob.PlayerParam{
		Acid:              p.AccountID.String(),
		GuildID:           p.GuildProfile.GuildUUID,
		Name:              simpleInfo.Name,
		GuildPosition:     simpleInfo.GuildPosition,
		Vip:               simpleInfo.Vip,
		FormationNew:      makeTodayFormation(p),
		FormationTeamFunc: p.buildHeroList,
	}
	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.PlayerWithNew(param)
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	info := player.GetPlayerInfo()

	ret := buildCSRobPlayerInfo(info)
	ret.FormationGS = 0
	data := p.Profile.GetData()
	for _, id := range info.CurrFormation {
		ret.FormationGS += int64(data.HeroGs[id])
	}

	rob := player.GetCurrCar()
	if nil != rob {
		ret.HasCurrCar = true
		netRob := buildCSRobCarInfo(rob)
		netRob.AppealLeast = gamedata.CSRobAppealLimit() - netRob.AppealNum
		ret.CurrCar = encode(netRob)
	} else {
		ret.HasCurrCar = false
	}
	rsp.Info = encode(ret)

	return rpcSuccess(rsp)
}

//协议------押运粮草:取日志/仇敌/求援
type reqMsgCSRobGetRecords struct {
	Req
	Type string `codec:"type"` // 查看的类型: log|appeal|enemy|all
}

type rspMsgCSRobGetRecords struct {
	Resp
	Logs    [][]byte `codec:"logs"`    //日志
	Appeals [][]byte `codec:"appeals"` //求援
	Enemies [][]byte `codec:"enemies"` //仇敌

	AutoAcceptBottom []int64 `codec:"auto_accept_bottom"`
}

func (p *Account) CSRobGetRecords(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobGetRecords)
	rsp := new(rspMsgCSRobGetRecords)

	initReqRsp(
		"Attr/CSRobGetRecordsRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Logs = [][]byte{}
	rsp.Appeals = [][]byte{}
	rsp.Enemies = [][]byte{}

	//活动时间检查
	//if false == csrobCheckTimeIn() {
	//	return rpcWarn(rsp, errCode.CommonNotInTime)
	//}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	if "log" == req.Type || "all" == req.Type {
		list := player.GetRecords()
		rsp.Logs = make([][]byte, 0, len(list))
		for _, record := range list {
			rsp.Logs = append(rsp.Logs, encode(buildCSRobRecord(&record)))
		}
	}

	if "appeal" == req.Type || "all" == req.Type {
		list := player.GetAppeals()
		rsp.Appeals = make([][]byte, 0, len(list))
		for _, appeal := range list {
			rsp.Appeals = append(rsp.Appeals, encode(buildCSRobAppeal(&appeal)))
		}
	}

	if "enemy" == req.Type || "all" == req.Type {
		list := player.GetEnemies()
		rsp.Enemies = make([][]byte, 0, len(list))
		for _, enemy := range list {
			rsp.Enemies = append(rsp.Enemies, encode(buildCSRobEnemy(&enemy)))
		}
	}

	aab, err := player.GetAutoAccept()
	if nil != err {
		logs.Warn("[CSRob] CSRobGetRecords GetAutoAccept, failed %v", err)
	}
	rsp.AutoAcceptBottom = []int64{}
	for _, g := range aab {
		rsp.AutoAcceptBottom = append(rsp.AutoAcceptBottom, int64(g))
	}

	return rpcSuccess(rsp)
}

//协议------押运粮草:阵容配置
type reqMsgCSRobSetFormation struct {
	Req
	Formation []int `codec:"formation"` // 阵容设置
}

type rspMsgCSRobSetFormation struct {
	Resp
	Ret string `codec:"ret"` // 成功:ok 不匹配:unmatch 失败:fail
}

func (p *Account) CSRobSetFormation(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobSetFormation)
	rsp := new(rspMsgCSRobSetFormation)

	initReqRsp(
		"Attr/CSRobSetFormationRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Ret = "fail"

	//活动时间检查
	if false == csrobCheckTimeIn() {
		return rpcWarn(rsp, errCode.CommonNotInTime)
	}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	_, bh := gamedata.CSRobBattleIDAndHeroID(time.Now().Unix())
	for _, idx := range req.Formation {
		hero := gamedata.GetHeroData(idx)
		if nil == hero || bh != hero.Nationality {
			rsp.Ret = "unmatch"
			return rpcSuccess(rsp)
		}
	}

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	//写入新的阵容
	team := p.buildHeroList(req.Formation)
	ret := player.SetFormation(req.Formation, team)
	if true == ret {
		rsp.Ret = "ok"
	}

	return rpcSuccess(rsp)
}

//协议------押运粮草:随机粮车
type reqMsgCSRobRandCar struct {
	Req
}

type rspMsgCSRobRandCar struct {
	SyncResp
	Grade uint32 `codec:"grade"` // 随机出的品质
	Ret   string `codec:"ret"`   // 成功:ok 超过时间限制:timeout
}

func (p *Account) CSRobRandCar(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobRandCar)
	rsp := new(rspMsgCSRobRandCar)

	initReqRsp(
		"Attr/CSRobRandCarRsp",
		r.RawBytes,
		req, rsp, p)

	//活动时间检查
	if false == csrobCheckTimeIn() {
		return rpcWarn(rsp, errCode.CommonNotInTime)
	}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	//检查时间限制
	vipCfg := gamedata.GetVIPCfg(int(p.Profile.Vip.V))
	now := time.Now().Unix()
	if now > gamedata.CSRobTodayEndTime()-int64(vipCfg.CSRobCarKeep)*util.MinSec {
		rsp.Ret = "timeout"
		return rpcSuccess(rsp)
	}

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	//生成新的当前品质
	gradeRefresh := player.GetGradeRefresh()
	if gamedata.CSRobBestGrade() == gradeRefresh.CurrGrade {
		return rpcWarn(rsp, errCode.CommonMaxLimit)
	}

	costDes := uint32(1)
	if !util.IsSameDayUnix(gradeRefresh.LastBuildTime, now) {
		costDes = uint32(2)
	}

	//0是初始无效值,用来描述'没有选过车'实际计算提升概率的时候视为从1开始升
	if 0 == gradeRefresh.CurrGrade {
		gradeRefresh.CurrGrade = gamedata.CSRobGradeFirstRefresh(p.GetRand().Float32())
		gradeRefresh.CarSumGradeRefresh = 1
	} else {
		prob, cost := gamedata.CSRobGradeRefresh(gradeRefresh.CarSumGradeRefresh, gradeRefresh.CurrGrade)
		cost = cost / costDes
		gradeRefresh.CarSumGradeRefresh++
		gradeRefresh.CarSumGradeCost += cost
		if gradeRefresh.CarSumGradeRefresh >= gamedata.CSRobGradeTopEdge() {
			gradeRefresh.CurrGrade = gamedata.CSRobBestGrade()
		} else {
			if p.GetRand().Float32() < prob {
				gradeRefresh.CurrGrade++
			}
		}

		//核算花费
		costData := gamedata.CostData{}
		costGroup := &account.CostGroup{}
		costData.AddItem(gamedata.VI_Hc, cost)
		if !costGroup.AddCostData(p.Account, &costData) {
			logs.SentryLogicCritical(p.AccountID.String(), "CSRob AddCostData Err by RandCar - %s : %d.",
				gamedata.VI_Hc, cost)
			return rpcWarn(rsp, errCode.CommonLessMoney)
		}

		//先扣钱
		if !costGroup.CostBySync(p.Account, rsp, "CSRobRandCar") {
			logs.SentryLogicCritical(p.AccountID.String(), "CSRob CostBySync Err by RandCar.")
			return rpcWarn(rsp, errCode.CommonLessMoney)
		}
		rsp.OnChangeHC()
	}

	//写入新的品质
	ret := player.SetGradeRefresh(gradeRefresh)
	if false == ret {
		return rpcWarn(rsp, errCode.CommonInner)
	}

	rsp.Grade = gradeRefresh.CurrGrade
	rsp.Ret = "ok"

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

//协议------押运粮草:出车
type reqMsgCSRobBuildCar struct {
	Req
	Type string `codec:"type"` // 一般出车:normal  一键出车:skip
}

type rspMsgCSRobBuildCar struct {
	SyncResp          //TODO 需要返回出车消耗, 希望后续改为增量发送, 不发送整个sync
	CarList  [][]byte `codec:"car_list"` //新增的粮车 []CSRobCarInfo
	Ret      string   `codec:"ret"`      // 成功:ok 超出时间限制timeout
}

func (p *Account) CSRobBuildCar(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobBuildCar)
	rsp := new(rspMsgCSRobBuildCar)

	initReqRsp(
		"Attr/CSRobBuildCarRsp",
		r.RawBytes,
		req, rsp, p)

	//活动时间检查
	if false == csrobCheckTimeIn() {
		return rpcWarn(rsp, errCode.CommonNotInTime)
	}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	// 检查是否已有车子在跑
	if true == player.CheckCurrCar() {
		return rpcWarn(rsp, errCode.ClickTooQuickly)
	}

	//校验次数限制
	vipCfg := gamedata.GetVIPCfg(int(p.Profile.Vip.V))
	info := player.GetPlayerInfo()
	if info.Count.Build >= vipCfg.CSRobBuildCarTimes {
		return rpcWarn(rsp, errCode.CommonCountLimit)
	}

	now := time.Now().Unix()
	timeAllowNum := (gamedata.CSRobTodayEndTime() - now) / int64(vipCfg.CSRobCarKeep*util.MinSec)
	if 0 > timeAllowNum {
		logs.Error("[CSRob] CSRobBuildCar timeAllowNum error, CSRobTodayEndTime:[%d], now:[%d]", gamedata.CSRobTodayEndTime(), now)
		return rpcWarn(rsp, errCode.CommonInner)
	}

	skipNum := vipCfg.CSRobBuildCarTimes - info.Count.Build
	if vipCfg.CSRobBuildCarTimes < info.Count.Build {
		skipNum = 0
	}
	if skipNum > uint32(timeAllowNum) {
		logs.Debug("[CSRob] CSRobBuildCar skipNum [%d] is larger than timeAllowNum [%d]", skipNum, timeAllowNum)
		skipNum = uint32(timeAllowNum)
	}
	logs.Debug("[CSRob] CSRobBuildCar skipNum is [%d]", skipNum)

	//检查时间限制
	if now > gamedata.CSRobTodayEndTime()-int64(vipCfg.CSRobCarKeep)*util.MinSec+gamedata.CSRobBuildCarOffsetTime() {
		rsp.Ret = "timeout"
		return rpcSuccess(rsp)
	}

	//检查阵容
	formation := player.GetFormation()
	if 0 == len(formation) {
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}

	//计算开销
	costData := gamedata.CostData{}
	cost := &account.CostGroup{}
	if "skip" == req.Type {
		sumCost := gamedata.CSRobSkipBuildCost() * skipNum
		gr := player.GetGradeRefresh()

		if !util.IsSameDayUnix(gr.LastBuildTime, now) {
			//从一键发车的费用中减去刷新车子花掉的费用, 如果刷新的花费超过一键发车单价, 只减去一个单价, 或者,如果已经刷到顶级, 减去一个单价
			if gr.CarSumGradeCost < (gamedata.CSRobSkipBuildCost()/2) && gamedata.CSRobBestGrade() != gr.CurrGrade {
				sumCost -= gr.CarSumGradeCost + (gamedata.CSRobSkipBuildCost() / 2)
			} else {
				sumCost -= gamedata.CSRobSkipBuildCost()
			}
		} else {
			//从一键发车的费用中减去刷新车子花掉的费用, 如果刷新的花费超过一键发车单价, 只减去一个单价, 或者,如果已经刷到顶级, 减去一个单价
			if gr.CarSumGradeCost < gamedata.CSRobSkipBuildCost() && gamedata.CSRobBestGrade() != gr.CurrGrade {
				sumCost -= gr.CarSumGradeCost
			} else {
				sumCost -= gamedata.CSRobSkipBuildCost()
			}
		}

		costData.AddItem(gamedata.VI_Hc, sumCost)
		if !cost.AddCostData(p.Account, &costData) {
			logs.SentryLogicCritical(p.AccountID.String(), "CSRob AddCostData Err by BuildCar Skip - %s : %d * %d - %d.",
				gamedata.VI_Hc, gamedata.CSRobSkipBuildCost(), skipNum, gr.CarSumGradeCost)
			return rpcWarn(rsp, errCode.CommonLessMoney)
		}

		//先扣钱
		if !cost.CostBySync(p.Account, rsp, "CSRobBuildCarSkip") {
			logs.SentryLogicCritical(p.AccountID.String(), "CSRob CostBySync Err by BuildCar.")
			return rpcWarn(rsp, errCode.CommonLessMoney)
		}
		rsp.OnChangeHC()
	} else {
		//单独发车并没有的消耗
	}

	//当前阵容筹备
	team := p.buildHeroList(formation)
	if nil == team || 0 == len(team) {
		return rpcWarn(rsp, errCode.CommonInner)
	}

	//出车
	list := []csrob.PlayerRob{}
	if "skip" == req.Type {
		cl, err := player.BuildCarSkip(team, skipNum, int64(vipCfg.CSRobCarKeep)*util.MinSec)
		if nil != err {
			return rpcWarn(rsp, errCode.CommonInner)
		}
		if nil != cl {
			list = cl[:]
		}
		logiclog.LogCSRobBuildCar(
			p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
			p.Profile.Vip.V, gamedata.CSRobBestGrade(), true, skipNum,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	} else {
		c, err := player.BuildCar(team, int64(vipCfg.CSRobCarKeep)*util.MinSec)
		if nil != err {
			return rpcWarn(rsp, errCode.CommonInner)
		}
		if nil != c {
			list = append(list, *c)
			logiclog.LogCSRobBuildCar(
				p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
				p.Profile.Vip.V, c.Info.Grade, false, 1,
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
		}
	}

	if nil != list {
		rsp.CarList = make([][]byte, 0, len(list))
		for _, c := range list {
			rsp.CarList = append(rsp.CarList, encode(buildCSRobCarInfo(&c)))
		}
	}

	rsp.Ret = "ok"

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

//协议------押运粮草:自动护送设置
type reqMsgCSRobAutoAcceptSet struct {
	Req
	GradeBottom []uint32 `codec:"grade_bottom"`
}

type rspMsgCSRobAutoAcceptSet struct {
	Resp
	Ret string `codec:"ret"` // 成功:ok 失败:fail
}

//CSRobAutoAcceptSet ..
func (p *Account) CSRobAutoAcceptSet(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobAutoAcceptSet)
	rsp := new(rspMsgCSRobAutoAcceptSet)

	initReqRsp(
		"Attr/CSRobAutoAcceptSetRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Ret = "fail"

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	err := player.SetAutoAccept(req.GradeBottom)
	if nil != err {
		logs.Error(fmt.Sprintf("[CSRob] CSRobAutoAcceptSet failed, %v", err))
		return rpcWarn(rsp, errCode.CommonInner)
	}
	logiclog.LogCSRobSetAutoAccept(
		p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		p.Profile.Vip.V, fmt.Sprintf("%v", req.GradeBottom),
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	rsp.Ret = "ok"

	return rpcSuccess(rsp)
}

//协议------押运粮草:求援
type reqMsgCSRobSendHelp struct {
	Req
	Acid  string `codec:"acid"`   // 向他请求救援
	CarID uint32 `codec:"car_id"` // 求援的车
}

type rspMsgCSRobSendHelp struct {
	Resp
	Ret string `codec:"ret"` // 成功:ok 超过车的求助次数了:limit 车子已经运完了:timeout 已经给他发过求援了:again 目标人物无效:target_invalid 失败:fail
}

func (p *Account) CSRobSendHelp(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobSendHelp)
	rsp := new(rspMsgCSRobSendHelp)

	initReqRsp(
		"Attr/CSRobSendHelpRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Ret = "fail"

	//活动时间检查
	if false == csrobCheckTimeIn() {
		return rpcWarn(rsp, errCode.CommonNotInTime)
	}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	//公会信息
	guildInfo, retErr := guild.GetModule(p.AccountID.ShardId).GetGuildInfo(p.GuildProfile.GuildUUID)
	if true == retErr.HasError() {
		logs.Error(fmt.Sprintf("[CSRob] CSRobSendHelp GetGuildInfo [%s] failed, %v", p.GuildProfile.GuildUUID, retErr.ErrMsg))
		return rpcWarn(rsp, errCode.CommonInner)
	}

	helperInfo := guildInfo.GetGuildMemInfo(req.Acid)
	if nil == helperInfo {
		rsp.Ret = "target_invalid"
		return rpcSuccess(rsp)
	}

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	ret, carInfo, err := player.SendHelp(req.Acid, req.CarID)
	if nil != err {
		logs.Error("%v", err)
		return rpcWarn(rsp, errCode.CommonInner)
	}

	switch ret {
	case csrob.RetCountLimit:
		rsp.Ret = "limit"
	case csrob.RetTimeout:
		rsp.Ret = "timeout"
	case csrob.RetCannotAgain:
		rsp.Ret = "again"
	case csrob.RetOK:
		rsp.Ret = "ok"
		if nil != carInfo {
			logiclog.LogCSRobSendHelp(
				p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
				p.Profile.Vip.V, carInfo.Grade, req.Acid, helperInfo.Vip,
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

			helperTeam, err := csrob.GetModule(p.AccountID.ShardId).GuildMod.GetPlayerTeam(p.GuildProfile.GuildUUID, req.Acid)
			if nil != err {
				logs.Warn("[CSRob] CSRobSendHelp, ProcessAutoReceive, GetPlayerTeam failed, %v", err)
			} else {
				if auto, err := player.ProcessAutoReceive(carInfo, req.Acid, helperTeam.Hero); nil != err {
					logs.Warn("[CSRob] CSRobSendHelp, ProcessAutoReceive failed, %v", err)
				} else {
					logs.Debug("[CSRob] CSRobSendHelp, ProcessAutoReceive Result [%v]", auto)
				}
			}
		}
	default:
		logs.Error("[CSRob] CSRobReceiveHelp Unexpect Ret, %d", ret)
		return rpcWarn(rsp, errCode.CommonInner)
	}

	return rpcSuccess(rsp)
}

//协议------押运粮草:同意求援
type reqMsgCSRobReceiveHelp struct {
	Req
	Acid  string `codec:"acid"`   // 同意他的求援
	CarID uint32 `codec:"car_id"` // 守护这辆车
}

type rspMsgCSRobReceiveHelp struct {
	Resp
	Ret string `codec:"ret"` // 成功:ok 正在被打:lock 已经被抢够了:limit 已经被别人帮了:hashelper 粮车已结束:timeout 失败:fail
}

func (p *Account) CSRobReceiveHelp(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobReceiveHelp)
	rsp := new(rspMsgCSRobReceiveHelp)

	initReqRsp(
		"Attr/CSRobReceiveHelpRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Ret = "fail"

	//活动时间检查
	if false == csrobCheckTimeIn() {
		return rpcWarn(rsp, errCode.CommonNotInTime)
	}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	//校验次数限制
	info := player.GetPlayerInfo()
	vipCfg := gamedata.GetVIPCfg(int(p.Profile.Vip.V))
	if info.Count.Help >= vipCfg.CSRobHelpLimit {
		return rpcWarn(rsp, errCode.CommonCountLimit)
	}

	//当前阵容筹备
	formation := player.GetFormation()
	team := p.buildHeroList(formation)
	if nil == team || 0 == len(team) {
		return rpcWarn(rsp, errCode.CommonInner)
	}

	ret, data, err := player.ReceiveHelp(req.Acid, req.CarID, team)
	if nil != err {
		logs.Error("[CSRob] CSRobReceiveHelp ReceiveHelp failed, %v", err)
		return rpcWarn(rsp, errCode.CommonInner)
	}

	switch ret {
	case csrob.RetLocked:
		rsp.Ret = "lock"
	case csrob.RetCountLimit:
		rsp.Ret = "limit"
	case csrob.RetTimeout:
		rsp.Ret = "timeout"
	case csrob.RetHasHelper:
		rsp.Ret = "hashelper"
	case csrob.RetOK:
		rsp.Ret = "ok"

		sender, err := db.ParseAccount(req.Acid)
		if nil != err {
			logs.Warn("[CSRob] ParseAccount Failed for Sender When LogCSRobSendHelp, %v", err)
		} else {
			senderInfo, err := simple_info.LoadAccountSimpleInfoProfile(sender)
			if nil != err {
				logs.Warn("[CSRob] LoadAccountSimpleInfoProfile Failed for Sender When LogCSRobSendHelp, %v", err)
			} else {
				logiclog.LogCSRobReceiveHelp(
					p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
					p.Profile.Vip.V, data.Info.Grade, req.Acid, senderInfo.Vip,
					func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			}
		}
	default:
		logs.Error("[CSRob] CSRobReceiveHelp Unexpect Ret, %d", ret)
		return rpcWarn(rsp, errCode.CommonInner)
	}

	return rpcSuccess(rsp)
}

//协议------押运粮草:抢他!!
type reqMsgCSRobRobIt struct {
	Req
	Acid  string `codec:"acid"`   // 抢他的车
	CarID uint32 `codec:"car_id"` // 抢这辆车
}

type rspMsgCSRobRobIt struct {
	Resp
	Ret     string `codec:"ret"`      // 成功:ok 被他人锁定:lock 攻击人数超出:limit 粮车已过期:timeout 工会伙伴:guildmate 失败:fail
	CarInfo []byte `codec:"car_info"` //车子的数据 CSRobCarInfo
}

func (p *Account) CSRobRobIt(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobRobIt)
	rsp := new(rspMsgCSRobRobIt)

	initReqRsp(
		"Attr/CSRobRobItRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Ret = "fail"
	rsp.CarInfo = []byte{}

	//活动时间检查
	if false == csrobCheckTimeIn() {
		return rpcWarn(rsp, errCode.CommonNotInTime)
	}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	//检查是不是本公会成员
	dAcid, err := db.ParseAccount(req.Acid)
	if nil != err {
		logs.Warn("[CSRob] CSRobRobIt ParseAccount target account failed, %v", err)
		return rpcWarn(rsp, errCode.CommonInvalidParam)
	}
	if dAcid.ShardId == p.AccountID.ShardId {
		sim, err := simple_info.LoadAccountSimpleInfoProfile(dAcid)
		if nil != err {
			logs.Warn("[CSRob] CSRobRobIt LoadAccountSimpleInfoProfile failed, %v", err)
			return rpcWarn(rsp, errCode.CommonInitFailed)
		}
		if sim.GuildName == p.GuildProfile.GuildName {
			rsp.Ret = "guildmate"
			return rpcSuccess(rsp)
		}
	}

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	//校验次数限制
	vipCfg := gamedata.GetVIPCfg(int(p.Profile.Vip.V))
	info := player.GetPlayerInfo()
	if info.Count.Rob >= vipCfg.CSRobRobTimes {
		return rpcWarn(rsp, errCode.CommonCountLimit)
	}

	ret, carData, err := player.RobCar(req.Acid, req.CarID)
	if nil != err {
		logs.Error("[CSRob] CSRobRobIt RobCar failed, %v", err)
		return rpcWarn(rsp, errCode.CommonInner)
	}

	switch ret {
	case csrob.RetLocked:
		rsp.Ret = "lock"
	case csrob.RetCountLimit:
		rsp.Ret = "limit"
	case csrob.RetTimeout:
		rsp.Ret = "timeout"
	case csrob.RetOK:
		rsp.Ret = "ok"
		rsp.CarInfo = encode(buildCSRobCarInfo(carData))
	default:
		logs.Error("[CSRob] CSRobRobIt Unexpect Ret, %d", ret)
		return rpcWarn(rsp, errCode.CommonInner)
	}

	return rpcSuccess(rsp)
}

//协议------押运粮草:挑战结果
type reqMsgCSRobRobResult struct {
	ReqWithAnticheat
	Acid    string `codec:"acid"`    // 抢他的车
	CarID   uint32 `codec:"car_id"`  // 抢这辆车
	Success bool   `codec:"success"` // 抢成功了没
}

type rspMsgCSRobRobResult struct {
	RespWithAnticheat
	Ret     string   `codec:"ret"`      // 成功:ok 被他人锁定:lock 攻击人数超出:limit 粮车已过期:timeout 失败:fail
	GoodID  []string `codec:"good_id"`  // 奖励物品ID
	GoodNum []int    `codec:"good_num"` // 奖励物品数量
}

func (p *Account) CSRobRobResult(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobRobResult)
	rsp := new(rspMsgCSRobRobResult)

	initReqRsp(
		"Attr/CSRobRobResultRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Ret = "fail"
	rsp.GoodID = []string{}
	rsp.GoodNum = []int{}

	//活动时间检查
	if false == csrobCheckTimeIn() {
		return rpcWarn(rsp, errCode.CommonNotInTime)
	}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	//战斗校验
	if cheatCode := p.AntiCheatCheckWithCode(&rsp.RespWithAnticheat, &req.ReqWithAnticheat, 0, account.Anticheat_Typ_CSRob); cheatCode != 0 {
		return rpcWarn(rsp, cheatCode)
	}

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	//校验次数限制
	vipCfg := gamedata.GetVIPCfg(int(p.Profile.Vip.V))
	info := player.GetPlayerInfo()
	if info.Count.Rob >= vipCfg.CSRobRobTimes {
		return rpcWarn(rsp, errCode.CommonCountLimit)
	}

	if false == req.Success {
		err := player.CancelRobCar(req.Acid, req.CarID)
		if nil != err {
			logs.Error("[CSRob] CSRobRobResult CancelRobCar failed, %v", err)
			return rpcWarn(rsp, errCode.CommonInner)
		}
		rsp.Ret = "ok"

		target, err := db.ParseAccount(req.Acid)
		if nil != err {
			logs.Warn("[CSRob] ParseAccount Failed for RobResult When LogCSRobSendHelp, %v", err)
		} else {
			targetInfo, err := simple_info.LoadAccountSimpleInfoProfile(target)
			if nil != err {
				// logs.Warn("[CSRob] LoadAccountSimpleInfoProfile Failed for RobResult When LogCSRobSendHelp, %v", err)
				logiclog.LogCSRobRobResult(
					p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
					p.Profile.Vip.V, p.Profile.GetData().CorpCurrGS, req.Acid, 0, 0, 0, true,
					func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			} else {
				logiclog.LogCSRobRobResult(
					p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
					p.Profile.Vip.V, p.Profile.GetData().CorpCurrGS, req.Acid, targetInfo.Vip, targetInfo.CurrCorpGs, 0, false,
					func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			}
		}
		return rpcSuccess(rsp)
	}

	ret, data, goods, err := player.DoneRobCar(req.Acid, req.CarID)
	if nil != err {
		logs.Error("[CSRob] CSRobRobIt RobCar failed, %v", err)
		return rpcWarn(rsp, errCode.CommonInner)
	}

	switch ret {
	case csrob.RetLocked:
		rsp.Ret = "lock"
	case csrob.RetCountLimit:
		rsp.Ret = "limit"
	case csrob.RetTimeout:
		rsp.Ret = "timeout"
	case csrob.RetOK:
		rsp.Ret = "ok"
		for id, num := range goods {
			rsp.GoodID = append(rsp.GoodID, id)
			rsp.GoodNum = append(rsp.GoodNum, int(num))
		}

		target, err := db.ParseAccount(req.Acid)
		if nil != err {
			logs.Warn("[CSRob] ParseAccount Failed for RobResult When LogCSRobSendHelp, %v", err)
		} else {
			grade := uint32(0)
			if nil != data {
				grade = data.Info.Grade
			}
			targetInfo, err := simple_info.LoadAccountSimpleInfoProfile(target)
			if nil != err {
				// logs.Warn("[CSRob] LoadAccountSimpleInfoProfile Failed for RobResult When LogCSRobSendHelp, %v", err)
				logiclog.LogCSRobRobResult(
					p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
					p.Profile.Vip.V, p.Profile.GetData().CorpCurrGS, req.Acid, 0, 0, grade, true,
					func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			} else {
				logiclog.LogCSRobRobResult(
					p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
					p.Profile.Vip.V, p.Profile.GetData().CorpCurrGS, req.Acid, targetInfo.Vip, targetInfo.CurrCorpGs, grade, true,
					func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			}
		}
	default:
		logs.Error("[CSRob] CSRobRobIt Unexpect Ret, %d", ret)
		return rpcWarn(rsp, errCode.CommonInner)
	}

	return rpcSuccess(rsp)
}

//协议------押运粮草:跨服跑马灯炫耀我把谁谁谁打劫了
type reqMsgCSRobSendMarquee struct {
	Req
	Acid  string `codec:"acid"`   // 抢他的车
	CarID uint32 `codec:"car_id"` // 抢这辆车
}

type rspMsgCSRobSendMarquee struct {
	SyncResp
	Ret string `codec:"ret"` // 成功:ok 失败:fail
}

func (p *Account) CSRobSendMarquee(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobSendMarquee)
	rsp := new(rspMsgCSRobSendMarquee)

	initReqRsp(
		"Attr/CSRobSendMarqueeRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Ret = "fail"

	//活动时间检查
	if false == csrobCheckTimeIn() {
		return rpcWarn(rsp, errCode.CommonNotInTime)
	}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(p.AccountID.String())
	if nil == player {
		return rpcWarn(rsp, errCode.CommonInitFailed)
	}

	//取缓存的跑马灯数据
	cache := player.GetMarqueeCache()

	if req.Acid != cache.Driver || req.CarID != cache.CarID {
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}

	if true == cache.Sent {
		return rpcWarn(rsp, errCode.CommonCountLimit)
	}

	//核算花费
	costData := gamedata.CostData{}
	costGroup := &account.CostGroup{}
	costData.AddItem(gamedata.VI_Hc, gamedata.CSRobMarqueeCost())
	if !costGroup.AddCostData(p.Account, &costData) {
		logs.SentryLogicCritical(p.AccountID.String(), "CSRob AddCostData Err by CSRobSendMarquee - %s : %d.",
			gamedata.VI_Hc, gamedata.CSRobMarqueeCost())
		return rpcWarn(rsp, errCode.CommonLessMoney)
	}

	//先扣钱
	if !costGroup.CostBySync(p.Account, rsp, "CSRobSendMarquee") {
		logs.SentryLogicCritical(p.AccountID.String(), "CSRob CostBySync Err by CSRobSendMarquee.")
		return rpcWarn(rsp, errCode.CommonLessMoney)
	}
	rsp.OnChangeHC()

	//发跑马灯
	gid := p.AccountID.GameId
	groupId := gamedata.GetCSRobGroupId(uint32(p.AccountID.ShardId))
	sids_ := gamedata.GetCSRobSids(groupId)
	gsids := make([]string, 0, len(sids_))
	for _, sid := range sids_ {
		gsString := sysnotice.GetRealSid(fmt.Sprintf("%d:%v", gid, sid))
		if gsString != "" {
			gsids = append(gsids, gsString)
		}
	}
	distinct_gsids, err := distinct.ValuesAndDisinct(gsids)
	if err != nil {
		logs.Error("CSRob distinct_sids error:%v")
		return rpcWarn(rsp, errCode.CommonInner)
	}
	if true == cache.HasHelper {
		for _, toGSid := range distinct_gsids {
			sysnotice.NewSysRollNotice(fmt.Sprintf("%v", toGSid), gamedata.IDS_CSRob_Rob_With_Helper).
				AddParam(sysnotice.ParamType_RollName, cache.HelperName).
				AddParam(sysnotice.ParamType_RollName, cache.DriverName).
				AddParam(sysnotice.ParamType_RollName, cache.RobberName).
				Send()
		}
	} else {
		for _, toGSid := range distinct_gsids {
			sysnotice.NewSysRollNotice(fmt.Sprintf("%v", toGSid), gamedata.IDS_CSRob_Rob_Without_Helper).
				AddParam(sysnotice.ParamType_RollName, cache.DriverName).
				AddParam(sysnotice.ParamType_RollName, cache.RobberName).
				Send()
		}
	}

	rsp.Ret = "ok"
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

//协议------押运粮草:四国战力排行榜
type reqMsgCSRobNationalityRank struct {
	Req
	Nat uint32 `codec:"nat"` //国家
}

type rspMsgCSRobNationalityRank struct {
	Resp
	MyPos  uint32   `codec:"mp"`   //我的排名
	MyTeam []byte   `codec:"mt"`   //我的榜上阵容
	TopN   [][]byte `codec:"list"` //榜上人
}

//CSRobNationalityRank ..
func (p *Account) CSRobNationalityRank(r servers.Request) *servers.Response {
	req := new(reqMsgCSRobNationalityRank)
	rsp := new(rspMsgCSRobNationalityRank)

	initReqRsp(
		"Attr/CSRobNationalityRankRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.MyPos = 0
	rsp.MyTeam = []byte{}
	rsp.TopN = [][]byte{}

	//参数检查
	if false == gamedata.CSRobNatCheck(req.Nat) {
		return rpcWarn(rsp, errCode.CommonInvalidParam)
	}

	// 自己是否不在公会中
	if !p.GuildProfile.InGuild() {
		return rpcWarn(rsp, uint32(errCode.GuildPlayerNotIn))
	}

	top := csrob.GetModule(p.AccountID.ShardId).PlayerRanker.GetRank(req.Nat)
	team, pos := csrob.GetModule(p.AccountID.ShardId).PlayerRanker.GetPos(req.Nat, p.AccountID.String())

	rsp.MyPos = pos
	if nil != team {
		rsp.MyTeam = encode(buildCSRobNationalityRankElem(team, pos))
	}

	rsp.TopN = make([][]byte, 0, len(top))
	for index, topTeam := range top {
		rsp.TopN = append(rsp.TopN, encode(buildCSRobNationalityRankElem(topTeam, uint32(index+1))))
	}

	return rpcSuccess(rsp)
}

func (p *Account) debugTest() {
	go func() {
		<-time.After(3 * time.Second)
		acid := p.AccountID.String()

		now := time.Now().Unix()
		logs.Debug("---[CSRob] now [%d], StartTime [%d] EndTime [%d] DailyBeginUnix [%d]", now, gamedata.CSRobTodayStartTime(), gamedata.CSRobTodayEndTime(), util.DailyBeginUnix(now))

		simpleInfo := p.Account.GetSimpleInfo()
		param := &csrob.PlayerParam{
			Acid:              p.AccountID.String(),
			GuildID:           p.GuildProfile.GuildUUID,
			Name:              simpleInfo.Name,
			GuildPosition:     simpleInfo.GuildPosition,
			Vip:               simpleInfo.Vip,
			FormationNew:      makeTodayFormation(p),
			FormationTeamFunc: p.buildHeroList,
		}
		player := csrob.GetModule(p.AccountID.ShardId).PlayerMod.PlayerWithNew(param)
		logs.Debug("---[CSRob] CSRobPlayerInfo get player {%v}", player)

		team := p.buildHeroList([]int{1, 2, 3})

		ret := player.SetFormation([]int{1, 2, 3}, team)
		logs.Debug("---[CSRob] SetFormation {%v}", ret)
		ret = player.SetGradeRefresh(csrob.PlayerGradeRefresh{})
		logs.Debug("---[CSRob] SetCurrGrade {%v}", ret)
		cl, err := player.BuildCar(team, 600)
		if nil != err {
			logs.Debug("---[CSRob] BuildCar Failed {%v}", err)
		} else {
			logs.Debug("---[CSRob] BuildCar {%v}", cl)
		}
		car := player.GetCurrCar()
		logs.Debug("---[CSRob] CSRobPlayerInfo GetCurrCar {%v}", car)
		robret, _, err := player.SendHelp(acid, car.CarID)
		logs.Debug("---[CSRob] SendHelp %d {%v}", robret, err)
		robret, _, err = player.ReceiveHelp(acid, car.CarID, team)
		logs.Debug("---[CSRob] ReceiveHelp %d {%v}", robret, err)

		player = csrob.GetModule(p.AccountID.ShardId).PlayerMod.Player(acid)
		if nil == player {
			logs.Error("---[CSRob] PlayerMod.Player Failed")
			return
		}
		robret, data, err := player.RobCar(acid, car.CarID)
		if nil != err {
			logs.Debug("---[CSRob] RobCar Failed {%v}", err)
		} else {
			logs.Debug("---[CSRob] RobCar Ret {%d} Data {%v}", robret, data)
		}
		robret, data, err = player.RobCar(acid, car.CarID)
		if nil != err {
			logs.Debug("---[CSRob] RobCar Failed {%v}", err)
		} else {
			logs.Debug("---[CSRob] RobCar Ret {%d} Data {%v}", robret, data)
		}
		robret, data, goods, err := player.DoneRobCar(acid, car.CarID)
		if nil != err {
			logs.Debug("---[CSRob] DoneRobCar Failed {%v}", err)
		} else {
			logs.Debug("---[CSRob] DoneRobCar Ret {%d} Data {%v}, Goods {%v}", robret, data, goods)
		}
		robret, data, err = player.RobCar(acid, car.CarID)
		if nil != err {
			logs.Debug("---[CSRob] RobCar Failed {%v}", err)
		} else {
			logs.Debug("---[CSRob] RobCar Ret {%d} Data {%v}", robret, data)
		}
		logs.Debug("---[CSRob] Wait to check lock")
		//<-time.After(15 * time.Second)
		err = player.CancelRobCar(acid, car.CarID)
		if nil != err {
			logs.Debug("---[CSRob] CancelRobCar Failed {%v}", err)
		} else {
			logs.Debug("---[CSRob] CancelRobCar OK")
		}

		logs.Debug("---[CSRob] GetRecords {%v}", player.GetRecords())
		logs.Debug("---[CSRob] GetEnemies {%v}", player.GetEnemies())
		logs.Debug("---[CSRob] GetAppeals {%v}", player.GetAppeals())

		guild := csrob.GetModule(p.AccountID.ShardId).GuildMod.GuildWithNew(p.Account.GuildProfile.GuildUUID, p.Account.GuildProfile.GuildName)
		if nil == guild {
			logs.Error("---[CSRob] GuildMod.Guild Failed")
			return
		}
		logs.Debug("---[CSRob] guild.GetInfo {%v} ", guild.GetInfo())
		carlist, err := guild.GetCarList()
		if nil != err {
			logs.Debug("---[CSRob] GetCarList Failed {%v}", err)
		} else {
			logs.Debug("---[CSRob] GetCarList {%v}", carlist)
		}

		logs.Debug("---[CSRob] guild.GetEnemies {%v}", guild.GetEnemies())
		logs.Debug("---[CSRob] guild.GetList {%v}", guild.GetList())
		logs.Debug("---[CSRob] guild.GetTeams {%v}", guild.GetTeams())
		logs.Debug("---[CSRob] guild.GetRankList {%v}", guild.GetRankList())
		logs.Debug("---[CSRob] guild.GetMyRank {%v}", guild.GetMyRank())

		logs.Debug("---[CSRob] csrob.PlayerRanker.GetRank {%v}", csrob.GetModule(p.AccountID.ShardId).PlayerRanker.GetRank(1))
		myTeam, myPos := csrob.GetModule(p.AccountID.ShardId).PlayerRanker.GetPos(1, p.AccountID.String())
		logs.Debug("---[CSRob] csrob.PlayerRanker.GetPos {%v} {%d}", myTeam, myPos)
	}()
}
