package csrob

import (
	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

//Player 玩家状态句柄
type Player struct {
	groupID uint32
	res     *resources

	acid string
	info *PlayerInfo
}

func (p *Player) initPlayer(param *PlayerParam) *Player {
	p.acid = param.Acid

	logs.Trace("[CSRob] initPlayer acid [%s]", p.acid)

	info, err := p.res.PlayerDB.getInfo(p.acid)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return nil
	}

	if nil == info {
		p.info = genPlayerInfo(param)
		p.res.CommandMod.notifyRefreshPlayerCacheBySelf(param.Acid, param.GuildID, param.Name, param.GuildPosition)
	} else {
		p.info = info
	}
	logs.Debug("[CSRob] initPlayer info [%v]", p.info)
	p.refreshInfo(false, param)
	err = p.saveInfo()
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return nil
	}

	return p
}

func (p *Player) loadPlayer(acid string) *Player {
	logs.Trace("[CSRob] loadPlayer acid [%s]", acid)
	info, err := p.res.PlayerDB.getInfo(acid)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return nil
	}

	if nil == info {
		return nil
	}

	p.acid = acid
	p.info = info
	logs.Debug("[CSRob] initPlayer info [%v]", p.info)
	p.refreshInfo(true, nil)

	return p
}

func (p *Player) refreshInfo(justLoad bool, param *PlayerParam) {
	now := time.Now()
	logs.Debug("[CSRob] refreshInfo justLoad [%v] now [%d]", justLoad, now.Unix())

	cache := p.res.poolName.GetPlayerCache(p.info.Acid)
	if cache.GuildID != p.info.GuildID {
		p.info.GuildID = p.res.poolName.GetPlayerGuildID(p.info.Acid)
		if false == justLoad {
			if nil != param.FormationTeamFunc {
				p.SetFormation(param.FormationNew, param.FormationTeamFunc(param.FormationNew))
			}
			//通知刷新自己的阵容
			// player_msg.Send(p.info.Acid, player_msg.PlayerMsgCSRobSetFormation,
			// 	player_msg.PlayerCSRobSetFormation{})
		}
	}
	if false == gamedata.CSRobCheckSameDay(p.info.UpdateTime, now.Unix()) {
		logs.Debug("[CSRob] refreshInfo, is not same day")
		p.info.Count = PlayerCount{}
		p.info.CurrFormation = []int{}

		if false == justLoad {
			if nil != param.FormationTeamFunc {
				p.SetFormation(param.FormationNew, param.FormationTeamFunc(param.FormationNew))
			}
			//通知刷新自己的阵容
			// player_msg.Send(p.info.Acid, player_msg.PlayerMsgCSRobSetFormation,
			// 	player_msg.PlayerCSRobSetFormation{})

			//清理db里面的昨日数据
			carList := []uint32{}
			for _, car := range p.info.CarList {
				carList = append(carList, car.CarID)
			}
			if err := p.res.PlayerDB.clearMyCars(p.info.Acid, carList); nil != err {
				logs.Error(fmt.Sprintf("%v", err))
			}
			p.info.CarList = []PlayerCarListElem{}
			if err := p.res.PlayerDB.clearMyAppealBefore(p.info.Acid, gamedata.CSRobTodayStartTime()); nil != err {
				logs.Error(fmt.Sprintf("%v", err))
			}
			if err := p.res.PlayerDB.trimMyRecords(p.info.Acid, gamedata.CSRobRecordTrim()*scaleSaveRecordsNum); nil != err {
				logs.Error(fmt.Sprintf("%v", err))
			}
		}
	}

	//每周清除仇敌
	if false == gamedata.CSRobCheckSameWeek(p.info.UpdateTime, now.Unix()) {
		logs.Debug("[CSRob] refreshInfo, is not same week")
		if false == justLoad {
			if err := p.res.PlayerDB.clearMyEnemies(p.info.Acid); nil != err {
				logs.Error(fmt.Sprintf("[CSRob] %v", err))
			}
		}
	}

	p.info.UpdateTime = now.Unix()

	status, err := p.getPlayerStatus(p.info.Acid, now)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
	} else {
		p.info.Count.Help = status.AcceptAppealCount
	}
	if nil != param {
		if err := p.res.PlayerDB.setPlayerStatusVIP(p.info.Acid, param.Vip); nil != err {
			logs.Warn("[CSRob] Player refreshInfo failed, %v", err)
		}
	}
}

func (p *Player) saveInfo() error {
	logs.Trace("[CSRob] saveInfo acid [%s]", p.info.Acid)
	p.info.UpdateTime = time.Now().Unix()
	err := p.res.PlayerDB.setInfo(p.info)
	if nil != err {
		return err
	}

	return nil
}

func (p *Player) checkGuild() bool {
	logs.Trace("[CSRob] checkGuild acid [%s]", p.info.Acid)
	return "" != p.info.GuildID
}

//GetPlayerInfo 取玩家状态信息
func (p *Player) GetPlayerInfo() *PlayerInfo {
	logs.Trace("[CSRob] GetPlayerInfo acid [%s]", p.info.Acid)
	return p.info
}

//GetCurrCar 取玩家当前车子
func (p *Player) GetCurrCar() *PlayerRob {
	logs.Trace("[CSRob] GetCurrCar acid [%s]", p.info.Acid)
	if 0 == len(p.info.CarList) {
		return nil
	}

	now := time.Now().Unix()
	var ret *PlayerRob
	for _, car := range p.info.CarList {
		if car.EndStamp < now || car.StartStamp > now {
			continue
		}

		rob, err := p.res.PlayerDB.getRob(p.acid, car.CarID)
		if nil != err {
			logs.Error(fmt.Sprintf("%v", err))
			continue
		}

		if nil != rob.Helper {
			rob.Helper.Name = p.res.poolName.GetPlayerCSName(rob.Helper.Acid)
		}
		rob.AlreadyAppeal = car.AlreadySendHelp

		ret = rob
		break
	}

	return ret
}

//CheckCurrCar 检查有没有当前车子
func (p *Player) CheckCurrCar() bool {
	logs.Trace("[CSRob] GetCurrCar acid [%s]", p.info.Acid)
	if 0 == len(p.info.CarList) {
		return false
	}
	now := time.Now().Unix()
	for _, car := range p.info.CarList {
		if car.EndStamp >= now && car.StartStamp < now {
			return true
		}
	}
	return false
}

//GetRecords 取玩家日志
func (p *Player) GetRecords() []PlayerRecord {
	logs.Trace("[CSRob] GetRecords acid [%s]", p.info.Acid)
	list, err := p.res.PlayerDB.getRecords(p.acid, gamedata.CSRobRecordTrim())
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return []PlayerRecord{}
	}
	for i, record := range list {
		if "" != record.DriverID {
			list[i].DriverName = p.res.poolName.GetPlayerCSName(list[i].DriverID)
		}
		if "" != record.RobberID {
			list[i].RobberName = p.res.poolName.GetPlayerCSName(list[i].RobberID)
		}
		if "" != record.HelperID {
			list[i].HelperName = p.res.poolName.GetPlayerCSName(list[i].HelperID)
		}
	}
	logs.Debug("[CSRob] GetRecords getRecords list {%v}", list)
	return list
}

//GetAppeals 取玩家得到的求援信
func (p *Player) GetAppeals() []PlayerAppeal {
	logs.Trace("[CSRob] GetAppeals acid [%s]", p.info.Acid)
	list, err := p.res.PlayerDB.getAppeals(p.acid, maxLoadAppealList)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return []PlayerAppeal{}
	}

	now := time.Now().Unix()
	retList := []PlayerAppeal{}
	for _, appeal := range list {
		if appeal.EndStamp < now {
			logs.Debug("[CSRob] GetAppeals acid [%s], ignore appeal {%v}", p.info.Acid, appeal)
			continue
		}

		data, err := p.res.PlayerDB.getRob(appeal.Acid, appeal.CarID)
		if nil != err {
			logs.Debug("[CSRob] GetAppeals acid [%s], getRob {%v} failed, %v", p.info.Acid, appeal, err)
			continue
		}

		if nil != data.Helper {
			appeal.HasHelper = true
			appeal.HelperIsMe = (data.Helper.Acid == p.info.Acid)
		}
		appeal.Robbers = data.Robbers
		appeal.Name = p.res.poolName.GetPlayerName(appeal.Acid)
		retList = append(retList, appeal)
	}
	return retList
}

//GetEnemies 取玩家的仇敌
func (p *Player) GetEnemies() []PlayerEnemy {
	logs.Trace("[CSRob] GetEnemies acid [%s]", p.info.Acid)
	list, err := p.res.PlayerDB.getEnemies(p.acid)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return []PlayerEnemy{}
	}

	now := time.Now().Unix()
	limit := int(gamedata.CSRobShowEnemiesLimit())
	retList := []PlayerEnemy{}
	for _, enemy := range list {
		obj := enemy
		pc := p.res.poolName.GetPlayerCache(enemy.Acid)
		if pc.GuildID == p.info.GuildID {
			continue
		}

		obj.Name = p.res.poolName.GetPlayerCSName(enemy.Acid)

		//取他的当前车状态
		info, err := p.res.PlayerDB.getInfo(enemy.Acid)
		if nil != err {
			logs.Error(fmt.Sprintf("%v", err))
			continue
		}
		if 0 != len(info.CarList) {
			for _, car := range info.CarList {
				if now < car.StartStamp || now > car.EndStamp {
					continue
				}

				rob, err := p.res.PlayerDB.getRob(enemy.Acid, car.CarID)
				if nil != err {
					logs.Error(fmt.Sprintf("%v", err))
					continue
				}
				obj.CurrCar = rob
				obj.CurrCar.Name = obj.Name
				obj.CurrCar.GuildID = pc.GuildID
				obj.CurrCar.GuildPos = pc.GuildPos
				obj.CurrCar.GuildName = p.res.poolName.GetGuildCSName(pc.GuildID)
				break
			}
		}

		retList = append(retList, obj)
		if len(retList) >= limit {
			break
		}
	}

	return retList
}

//GetFormation 取玩家当前阵容
func (p *Player) GetFormation() []int {
	logs.Trace("[CSRob] GetFormation acid [%s]", p.info.Acid)
	return p.info.CurrFormation
}

//SetFormation 设置玩家阵容
func (p *Player) SetFormation(formation []int, team []HeroInfo) bool {
	logs.Trace("[CSRob] SetFormation acid [%s]", p.info.Acid)
	p.info.CurrFormation = formation[:]
	err := p.res.PlayerDB.setInfo(p.info)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return false
	}

	guildTeam := &GuildTeam{
		Acid: p.acid,
		Hero: team,
	}
	_, nat := gamedata.CSRobBattleIDAndHeroID(time.Now().Unix())
	err = p.res.GuildDB.pushTeam(p.info.GuildID, guildTeam, nat)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
	}

	return true
}

//SetGradeRefresh 设置玩家粮车刷新状态
func (p *Player) SetGradeRefresh(gr PlayerGradeRefresh) bool {
	logs.Trace("[CSRob] RandNewGrade acid [%s]", p.info.Acid)
	p.info.GradeRefresh = gr

	err := p.res.PlayerDB.setInfo(p.info)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return false
	}

	return true
}

//GetGradeRefresh 取玩家粮车刷新状态
func (p *Player) GetGradeRefresh() PlayerGradeRefresh {
	logs.Trace("[CSRob] GetCurrGrade acid [%s]", p.info.Acid)
	return p.info.GradeRefresh
}

//BuildCarSkip 一键发车
func (p *Player) BuildCarSkip(team []HeroInfo, num uint32, keepTime int64) ([]PlayerRob, error) {
	logs.Trace("[CSRob] BuildCarSkip acid [%s]", p.info.Acid)

	grade := gamedata.CSRobBestGrade()
	now := time.Now().Unix()
	startTime := now

	list := make([]PlayerRob, 0, num)
	for i := uint32(0); i < num; i++ {
		robInfo := PlayerRobInfo{
			CarID:      p.info.NextCarID,
			Grade:      grade,
			Team:       team[:],
			StartStamp: startTime,
			EndStamp:   startTime + keepTime,
		}
		err := p.res.PlayerDB.buildCar(p.acid, robInfo)
		if nil != err {
			logs.Error(fmt.Sprintf("%v", err))
			continue
		}

		p.info.Count.Build++
		p.info.NextCarID++
		elem := PlayerCarListElem{
			PlayerRobInfo:   robInfo,
			AlreadySendHelp: []string{},
		}
		p.info.CarList = append(p.info.CarList, elem)

		startTime = robInfo.EndStamp + 1
		p.res.ticker.regReward(p.acid, robInfo.CarID, robInfo.EndStamp)

		rob := PlayerRob{
			CarID:   robInfo.CarID,
			Info:    robInfo,
			Robbing: false,
			Robbers: []string{},

			Acid: p.acid,

			Helper: nil,
		}
		list = append(list, rob)

		if true == p.checkGuild() {
			elem := GuildRobElem{
				Acid:       p.acid,
				CarID:      robInfo.CarID,
				StartStamp: robInfo.StartStamp,
				EndStamp:   robInfo.EndStamp,
			}
			_, nat := gamedata.CSRobBattleIDAndHeroID(now)
			err := p.res.GuildDB.pushCar(p.info.GuildID, nat, elem)
			if nil != err {
				logs.Error(fmt.Sprintf("%v", err))
				continue
			}
		}
	}

	p.res.CommandMod.notifyPushGuildList(p.info.GuildID)

	if 0 != len(list) {
		p.info.GradeRefresh.reset()
		p.info.GradeRefresh.LastBuildTime = now
		if err := p.saveInfo(); nil != err {
			logs.Error(fmt.Sprintf("%v", err))
		}
	}

	return list, nil
}

//BuildCar 普通发车
func (p *Player) BuildCar(team []HeroInfo, keepTime int64) (*PlayerRob, error) {
	logs.Trace("[CSRob] BuildCar acid [%s]", p.info.Acid)

	now := time.Now().Unix()
	robInfo := PlayerRobInfo{
		CarID:      p.info.NextCarID,
		Grade:      p.info.GradeRefresh.CurrGrade,
		Team:       team[:],
		StartStamp: now,
		EndStamp:   now + keepTime,
	}

	if err := p.res.PlayerDB.buildCar(p.acid, robInfo); nil != err {
		return nil, err
	}

	p.info.Count.Build++
	p.info.NextCarID++
	elem := PlayerCarListElem{
		PlayerRobInfo:   robInfo,
		AlreadySendHelp: []string{},
	}
	p.info.CarList = append(p.info.CarList, elem)
	p.info.GradeRefresh.reset()
	p.info.GradeRefresh.LastBuildTime = now
	if err := p.saveInfo(); nil != err {
		return nil, err
	}

	p.res.ticker.regReward(p.acid, robInfo.CarID, robInfo.EndStamp)

	if true == p.checkGuild() {
		elem := GuildRobElem{
			Acid:       p.acid,
			CarID:      robInfo.CarID,
			StartStamp: robInfo.StartStamp,
			EndStamp:   robInfo.EndStamp,
		}
		_, nat := gamedata.CSRobBattleIDAndHeroID(time.Now().Unix())
		err := p.res.GuildDB.pushCar(p.info.GuildID, nat, elem)
		if nil != err {
			return nil, err
		}
	}

	p.res.CommandMod.notifyPushGuildList(p.info.GuildID)

	rob := &PlayerRob{
		CarID:   robInfo.CarID,
		Info:    robInfo,
		Robbing: false,
		Robbers: []string{},

		Acid: p.acid,

		Helper: nil,
	}

	return rob, nil
}

//SetAutoAccept ..
func (p *Player) SetAutoAccept(bottom []uint32) error {
	return p.res.PlayerDB.setPlayerStatusAutoAcceptBottom(p.info.Acid, bottom)
}

//GetAutoAccept ..
func (p *Player) GetAutoAccept() ([]uint32, error) {
	auto, err := p.res.PlayerDB.getPlayerStatusAutoAcceptBottom(p.info.Acid)
	if nil != err {
		return []uint32{}, err
	}
	return auto, nil
}

//SendHelp 发送求援信息
func (p *Player) SendHelp(helper string, car uint32) (int, *PlayerCarListElem, error) {
	logs.Trace("[CSRob] SendHelp acid [%s]", p.info.Acid)

	var carInfo *PlayerCarListElem
	carIndex := -1
	for index, c := range p.info.CarList {
		if car == c.CarID {
			carInfo = &(p.info.CarList[index])
			carIndex = index
			break
		}
	}

	if nil == carInfo {
		return RetInvalid, nil, makeError("SendHelp for a no exist car, acid [%v], car [%v]", p.info.Acid, car)
	}

	if time.Now().Unix() > carInfo.EndStamp {
		return RetTimeout, nil, nil
	}

	if len(carInfo.AlreadySendHelp) >= int(gamedata.CSRobAppealLimit()) {
		return RetCountLimit, nil, nil
	}

	for _, as := range carInfo.AlreadySendHelp {
		if as == helper {
			return RetCannotAgain, nil, nil
		}
	}

	appeal := PlayerAppeal{
		Acid:       p.acid,
		CarID:      car,
		Grade:      carInfo.Grade,
		AppealTime: time.Now().Unix(),
		EndStamp:   carInfo.EndStamp,
	}
	err := p.res.PlayerDB.pushAppeal(helper, appeal)
	if nil != err {
		return RetInvalid, nil, err
	}

	if -1 != carIndex {
		p.info.CarList[carIndex].AlreadySendHelp = append(p.info.CarList[carIndex].AlreadySendHelp, helper)
		if err := p.saveInfo(); nil != err {
			return RetInvalid, nil, err
		}
	}

	return RetOK, carInfo, nil
}

//ProcessAutoReceive ..
func (p *Player) ProcessAutoReceive(car *PlayerCarListElem, helper string, helpTeam []HeroInfo) (bool, error) {
	hasHelper, err := p.res.PlayerDB.getRobHelper(p.info.Acid, car.CarID)
	if nil != err {
		return false, fmt.Errorf("getRobHelper [%s]:[%d] failed, %v", p.info.Acid, car.CarID, err)
	}
	if nil != hasHelper {
		logs.Debug("[CSRob] Player ProcessAutoReceive, Refuse, HasHelper [%v]", hasHelper.Acid)
		return false, nil
	}

	now := time.Now()
	helperStatus, err := p.getPlayerStatus(helper, now)
	if nil != err {
	}
	if false == p.checkAutoAcceptIn(car.Grade, helperStatus.AutoAcceptBottom) {
		logs.Debug("[CSRob] Player ProcessAutoReceive, Refuse, AutoAcceptBottom [%v], car [%d]", helperStatus.AutoAcceptBottom, car.Grade)
		return false, nil
	}
	maxAccept := getPlayerHelpLimit(helperStatus.VIP)
	if helperStatus.AcceptAppealCount > maxAccept {
		logs.Debug("[CSRob] Player ProcessAutoReceive, Refuse, AcceptAppealCount [%d], maxAccept [%d]", helperStatus.AcceptAppealCount, maxAccept)
		return false, nil
	}

	newCount, err := p.res.PlayerDB.pushPlayerStatusAutoAppeal(helper, now, 1)
	if nil != err {
		return false, fmt.Errorf("pushPlayerStatusAutoAppeal [%s] failed, %v", helper, err)
	}
	if newCount > maxAccept {
		logs.Debug("[CSRob] Player ProcessAutoReceive, Refuse, newCount [%d], maxAccept [%d]", newCount, maxAccept)
		p.res.PlayerDB.pushPlayerStatusAutoAppeal(helper, now, -1)
		return false, nil
	}

	help := &PlayerRobHelper{
		Acid: helper,
		Team: helpTeam[:],
	}
	ok, err := p.res.PlayerDB.setRobHelper(p.info.Acid, car.CarID, help)
	if nil != err {
		return false, err
	}
	if false == ok {
		logs.Debug("[CSRob] Player ProcessAutoReceive, Refuse, too late when setRobHelper")
		p.res.PlayerDB.pushPlayerStatusAutoAppeal(helper, now, -1)
		return false, nil
	}

	return true, nil
}

func (p *Player) getPlayerStatus(acid string, now time.Time) (*PlayerStatus, error) {
	status, err := p.res.PlayerDB.getPlayerStatus(acid)
	if nil != err {
		return nil, err
	}
	auto, err := p.res.PlayerDB.getPlayerStatusAutoAcceptBottom(acid)
	if nil != err {
		return nil, err
	}
	status.AutoAcceptBottom = auto
	if gamedata.CSRobCheckSameDay(status.LastUpdate, now.Unix()) {
		return status, nil
	}

	newStatus, err := p.res.PlayerDB.resetPlayerStatus(acid, status, now)
	if nil != err {
		return status, err
	}
	newStatus.AutoAcceptBottom = auto
	return newStatus, nil
}

func (p *Player) checkAutoAcceptIn(check uint32, auto []uint32) bool {
	for _, g := range auto {
		if g == check {
			return true
		}
	}
	return false
}

func getPlayerHelpLimit(vip uint32) uint32 {
	vipCfg := gamedata.GetVIPCfg(int(vip))
	if nil != vipCfg {
		return vipCfg.CSRobHelpLimit
	}
	return 0
}

//ReceiveHelp 接受求援
func (p *Player) ReceiveHelp(who string, car uint32, team []HeroInfo) (int, *PlayerRob, error) {
	logs.Trace("[CSRob] ReceiveHelp acid [%s]", p.info.Acid)

	data, err := p.res.PlayerDB.getRob(who, car)
	if nil != err {
		return RetInvalid, nil, err
	}

	if nil != data.Helper {
		return RetHasHelper, nil, nil
	}

	if true == data.Robbing {
		return RetLocked, nil, nil
	}

	if int(getBeRobLimit()) <= len(data.Robbers) {
		return RetCountLimit, nil, nil
	}

	now := time.Now()
	if now.Unix() > data.Info.EndStamp {
		return RetTimeout, nil, nil
	}

	status, err := p.getPlayerStatus(p.info.Acid, now)
	if nil != err {
		return RetInvalid, nil, fmt.Errorf("getPlayerStatus [%s] failed, %v", p.info.Acid, err)
	}
	maxAccept := getPlayerHelpLimit(status.VIP)
	if status.AcceptAppealCount > maxAccept {
		return RetCountLimit, nil, nil
	}

	newCount, err := p.res.PlayerDB.pushPlayerStatusAutoAppeal(p.info.Acid, now, 1)
	if nil != err {
		return RetInvalid, nil, fmt.Errorf("pushPlayerStatusAutoAppeal [%s] failed, %v", p.info.Acid, err)
	}
	if newCount > maxAccept {
		p.res.PlayerDB.pushPlayerStatusAutoAppeal(p.info.Acid, now, -1)
		return RetCountLimit, nil, nil
	}
	p.info.Count.Help = status.AcceptAppealCount

	help := &PlayerRobHelper{
		Acid: p.acid,
		Team: team[:],
	}
	ok, err := p.res.PlayerDB.setRobHelper(who, car, help)
	if nil != err {
		return RetInvalid, nil, err
	}

	if false == ok {
		p.res.PlayerDB.pushPlayerStatusAutoAppeal(p.info.Acid, now, -1)
		return RetHasHelper, nil, nil
	}

	return RetOK, data, nil
}

//RobCar 开始抢夺一辆车
func (p *Player) RobCar(acid string, car uint32) (int, *PlayerRob, error) {
	logs.Trace("[CSRob] RobCar acid [%s]", p.info.Acid)

	data, err := p.res.PlayerDB.getRob(acid, car)
	if nil != err {
		return RetInvalid, nil, err
	}
	if nil == data {
		return RetInvalid, nil, makeError("car doesn't exist, acid [%s], car [%d]", acid, car)
	}

	if int(getBeRobLimit()) <= len(data.Robbers) {
		return RetCountLimit, nil, nil
	}

	if true == data.Robbing {
		return RetLocked, nil, nil
	}

	now := time.Now().Unix()
	if now >= data.Info.EndStamp {
		return RetTimeout, nil, nil
	}

	ok, err := p.res.PlayerDB.touchRob(acid, car, p.acid, now+gamedata.CSRobRobTimeout())
	if nil != err {
		return RetInvalid, nil, err
	}

	if false == ok {
		return RetLocked, nil, nil
	}

	if nil != data.Helper {
		data.Helper.Name = p.res.poolName.GetPlayerCSName(data.Helper.Acid)
	}

	p.res.CommandMod.notifyPushGuildList(p.info.GuildID)

	return RetOK, data, nil
}

//CancelRobCar 取消抢夺一辆车
func (p *Player) CancelRobCar(acid string, car uint32) error {
	logs.Trace("[CSRob] CancelRobCar acid [%s]", p.info.Acid)
	return p.res.PlayerDB.unTouchRob(acid, car, p.acid)
}

//DoneRobCar 成功抢夺一辆车
func (p *Player) DoneRobCar(acid string, car uint32) (int, *PlayerRob, map[string]uint32, error) {
	logs.Trace("[CSRob] DoneRobCar acid [%s]", p.info.Acid)
	data, err := p.res.PlayerDB.getRob(acid, car)
	if nil != err {
		return RetInvalid, nil, nil, err
	}

	if nil == data {
		return RetInvalid, nil, nil, makeError("car doesn't exist, acid [%s], car [%d]", acid, car)
	}

	if int(getBeRobLimit()) <= len(data.Robbers) {
		return RetCountLimit, nil, nil, nil
	}

	now := time.Now().Unix()
	if now >= data.Info.EndStamp {
		return RetTimeout, nil, nil, nil
	}

	ok, err := p.res.PlayerDB.pushRob(acid, car, p.acid, getBeRobLimit())
	if nil != err {
		return RetInvalid, nil, nil, err
	}

	if false == ok {
		return RetLocked, nil, nil, nil
	}

	//取我的发奖记录
	rr := p.getRewardInfo()

	//生成奖励
	history, goods, _ := gamedata.CSRobRewardForRob(rr.DropHistory, data.Info.Grade)
	rr.DropHistory = history

	//记次数
	p.info.Count.Rob++

	//缓存预备发跑马灯的数据
	p.info.LastRob.setCacheForMarquee(p.info.Acid, data, goods)

	//写数据
	if err := p.saveInfo(); nil != err {
		return RetInvalid, nil, nil, err
	}
	//记录奖励历史
	p.setRewardInfo(rr)

	recordBeRob := PlayerRecord{Type: RecordBeRob, DriverID: acid, RobberID: p.acid, Grade: data.Info.Grade, Timestamp: now, Goods: goods}
	recordRob := PlayerRecord{Type: RecordRob, DriverID: acid, RobberID: p.acid, Grade: data.Info.Grade, Timestamp: now, Goods: goods}
	if nil != data.Helper {
		recordBeRob.HelperID = data.Helper.Acid
		recordRob.HelperID = data.Helper.Acid
	}
	p.res.PlayerDB.pushRecord(acid, recordBeRob)
	p.res.PlayerDB.pushRecord(p.acid, recordRob)
	p.res.PlayerDB.pushEnemy(acid, p.acid, 1)

	dGuildID := p.res.poolName.GetPlayerGuildID(acid)
	if "" != dGuildID && "" != p.info.GuildID {
		p.res.GuildDB.pushEnemy(dGuildID, p.info.GuildID, 1)
		p.res.GuildDB.incrRobTimes(p.info.GuildID, 1, now, p.res.ranker.batchStr)
		p.res.ranker.addTrig(p.info.GuildID)
	}

	//发奖励邮件
	mail_sender.BatchSendMail2Account(p.info.Acid,
		timail.Mail_send_By_CSRob,
		mail_sender.IDS_MAIL_ROB_TITLE,
		[]string{
			fmt.Sprintf("%s", p.res.poolName.GetPlayerCSName(acid)),
		},
		goods,
		"CSROB: send reward by DoneRobCar", false)

	data.Robbers = append(data.Robbers, p.acid)
	return RetOK, data, goods, nil
}

func (p *Player) getRewardInfo() *PlayerRewardInfo {
	logs.Trace("[CSRob] getRewardInfo acid [%s]", p.info.Acid)
	rewardInfo, err := p.res.PlayerDB.getRewardInfo(p.acid)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
	}

	now := time.Now().Unix()
	if nil == rewardInfo {
		rewardInfo = &PlayerRewardInfo{
			UpdateTime:  now,
			DropHistory: make(map[string]uint32),
		}
	} else if false == gamedata.CSRobCheckSameDay(rewardInfo.UpdateTime, now) {
		rewardInfo.UpdateTime = now
		rewardInfo.DropHistory = make(map[string]uint32)
	} else {
		rewardInfo.UpdateTime = now
	}

	return rewardInfo
}

func (p *Player) setRewardInfo(info *PlayerRewardInfo) {
	logs.Trace("[CSRob] setRewardInfo acid [%s]", p.info.Acid)
	err := p.res.PlayerDB.setRewardInfo(p.info.Acid, info)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
	}
}

//DoneDrivingCar 结束一辆车的运粮
func (p *Player) DoneDrivingCar(car uint32) (int, *PlayerRob, map[string]uint32, error) {
	logs.Trace("[CSRob] DoneDrivingCar acid [%s]", p.info.Acid)
	data, err := p.res.PlayerDB.getRob(p.info.Acid, car)
	if nil != err {
		return RetInvalid, nil, nil, err
	}

	if nil == data {
		return RetInvalid, nil, nil, makeError("car doesn't exist, acid [%s], car [%d]", p.info.Acid, car)
	}

	if nil != data.Reward && 0 != data.Reward.Time {
		logs.Warn("[CSRob] DoneDrivingCar try to send reward again, car {%v}", data)
		return RetCountLimit, nil, nil, nil
	}

	now := time.Now().Unix()
	if now < data.Info.EndStamp {
		return RetTimeout, nil, nil, nil
	}

	//取我的发奖记录
	rr := p.getRewardInfo()

	//生成奖励
	history, goods, dark := gamedata.CSRobRewardForDriver(rr.DropHistory, data.Info.Grade, uint32(len(data.Robbers)))
	rr.DropHistory = history

	//记录奖励已发
	reward := &PlayerRobReward{
		Time:   now,
		BeDark: dark,
		Goods:  goods,
	}
	err = p.res.PlayerDB.setRobReward(p.info.Acid, car, reward)
	if nil != err {
		return RetInvalid, nil, nil, err
	}

	recordDone := PlayerRecord{Type: RecordDoneDriving, DriverID: p.info.Acid, Grade: data.Info.Grade, Timestamp: now, Goods: goods}
	if nil != data.Helper {
		recordDone.HelperID = data.Helper.Acid
	}
	p.res.PlayerDB.pushRecord(p.info.Acid, recordDone)

	//记录奖励历史
	p.setRewardInfo(rr)

	//发奖励邮件
	mail_sender.BatchSendMail2Account(p.info.Acid,
		timail.Mail_send_By_CSRob,
		mail_sender.IDS_MAIL_CROPS_TITLE,
		[]string{
			fmt.Sprintf("%d", data.Info.Grade),
		},
		goods,
		"CSROB: send reward by DoneDrivingCar", false)

	return RetOK, data, goods, nil
}

//DoneHelpCar 结束一辆车的护卫(援助)
func (p *Player) DoneHelpCar(driver string, data *PlayerRob) {
	logs.Trace("[CSRob] DoneHelpCar acid [%s]", p.info.Acid)

	//取我的发奖记录
	rr := p.getRewardInfo()

	//生成奖励
	history, goods, _ := gamedata.CSRobRewardForHelp(rr.DropHistory, data.Info.Grade)
	rr.DropHistory = history

	//记录奖励历史
	p.setRewardInfo(rr)

	now := time.Now().Unix()
	recordHelp := PlayerRecord{Type: RecordDoneHelp, DriverID: data.Acid, HelperID: p.info.Acid, Grade: data.Info.Grade, Timestamp: now, Goods: goods}
	p.res.PlayerDB.pushRecord(p.info.Acid, recordHelp)

	//发奖励邮件
	mail_sender.BatchSendMail2Account(p.info.Acid,
		timail.Mail_send_By_CSRob,
		mail_sender.IDS_MAIL_ESCORT_TITLE,
		[]string{
			fmt.Sprintf("%s", p.res.poolName.GetPlayerCSName(driver)),
		},
		goods,
		"CSROB: send reward by DoneHelpCar", false)
}

//GetMarqueeCache 取跑马灯缓存
func (p *Player) GetMarqueeCache() *CacheForMarquee {
	p.info.LastRob.DriverName = p.res.poolName.GetPlayerCSName(p.info.LastRob.Driver)
	p.info.LastRob.RobberName = p.res.poolName.GetPlayerCSName(p.info.LastRob.Robber)
	if true == p.info.LastRob.HasHelper {
		p.info.LastRob.HelperName = p.res.poolName.GetPlayerCSName(p.info.LastRob.Helper)
	}
	return &p.info.LastRob
}

func getBeRobLimit() uint32 {
	return gamedata.CSRobRobLimit()
}

//Debug for Cheat

//DebugClearCount cheat:清除计数
func (p *Player) DebugClearCount() bool {
	p.info.Count = PlayerCount{}
	err := p.res.PlayerDB.setInfo(p.info)
	if nil != err {
		logs.Error(fmt.Sprintf("[CSRob] DebugClearCount Failed, %v", err))
		return false
	}

	return true
}

//DebugSetMaxCount cheat:填满计数
func (p *Player) DebugSetMaxCount(count PlayerCount) bool {
	p.info.Count = count
	err := p.res.PlayerDB.setInfo(p.info)
	if nil != err {
		logs.Error(fmt.Sprintf("[CSRob] DebugSetMaxCount Failed, %v", err))
		return false
	}

	return true
}

//DebugAddEnemy cheat:增加自己为仇敌
func (p *Player) DebugAddEnemy() bool {
	record := PlayerRecord{Type: RecordRob, DriverID: p.info.Acid, RobberID: p.info.Acid, Grade: 1, Timestamp: 100}
	p.res.PlayerDB.pushRecord(p.info.Acid, record)
	p.res.PlayerDB.pushRecord(p.info.Acid, record)
	p.res.PlayerDB.pushEnemy(p.info.Acid, p.info.Acid, 1)
	p.res.GuildDB.pushEnemy(p.info.GuildID, p.info.GuildID, 1)
	return true
}
