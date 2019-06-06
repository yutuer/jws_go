package account

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"math"

	gm "github.com/rcrowley/go-metrics"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	. "vcs.taiyouxi.net/jws/gamex/models/account/warm"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	gveHelper "vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/security"

	"encoding/json"

	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/account/simple_info"
	"vcs.taiyouxi.net/jws/gamex/models/world_boss"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice/teamboss"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice/worldboss"
	"vcs.taiyouxi.net/jws/gamex/modules/friend"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
	"vcs.taiyouxi.net/jws/gamex/modules/hour_log"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/modules/room"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	helper2 "vcs.taiyouxi.net/jws/helper"
)

var (
	monoNewAccountCounter gm.Counter
	monoLoginCounter      gm.Counter
)

func init() {
	monoNewAccountCounter = metrics.NewCounter("gamex.login.newAccount")
	monoLoginCounter = metrics.NewCounter("gamex.login.newLogin")
}

type Account struct {
	//当前玩家的Account
	AccountID db.Account

	Profile           Profile
	Tmp               TmpProfile
	BagProfile        PlayerBag
	StoreProfile      PlayerStores
	GeneralProfile    PlayerGenerals
	GuildProfile      PlayerGuild
	AntiCheat         PlayerAntiCheat
	SimpleInfoProfile simple_info.AccountSimpleInfoProfile
	Friend            Friend

	ip     string
	rander *rand.Rand // 随机数生成器
	//dbHandler        driver.DBHandle
	dbRequestCounter uint32
	dbError          chan error           //dbroutine退出和出错的信号
	msgChannel       chan servers.Request // msg channel
	pushCh           chan game.INotifySyncMsg
	quit             chan struct{}
	requestSeq       uint64 // 当前登录后产生的请求序列号
	//lastLevelEnemyLoot map[string]string

	handle events.Handlers //一些事件通知
}

func IsHasReg(uid string) (bool, error) {
	str := fmt.Sprintf("{\"uid\":\"%s\"}", uid)
	res, err := util.HttpPost(game.Cfg.IsRegUrl, util.JsonPostTyp, []byte(str))
	logs.Trace("IsHasReg Res %s", string(res))
	return string(res) == "yes", err
}

//NewAccount 返回Account 和 这个账号是否是初始化的
func NewAccount(dbaccount db.Account, ip string) (*Account, bool) {
	accountID := dbaccount
	p := &Account{
		AccountID:         accountID,
		Profile:           NewProfile(accountID),
		Tmp:               NewTmpProfile(accountID),
		BagProfile:        NewPlayerBag(accountID),
		StoreProfile:      NewPlayerStores(accountID),
		GeneralProfile:    NewPlayerGenerals(accountID),
		GuildProfile:      NewPlayerGuild(accountID),
		AntiCheat:         NewPlayerAntiCheat(accountID),
		SimpleInfoProfile: simple_info.NewSimpleInfoProfile(accountID),
		Friend:            NewFriend(accountID),

		ip: ip,
		//dbHandler:        driver.GetDBService().CreateDBHandle(),

		dbError:    make(chan error, 1),
		msgChannel: make(chan servers.Request, 32),
		pushCh:     make(chan game.INotifySyncMsg, 32),
		quit:       make(chan struct{}),
	}

	is_need_init := false
	shardID := dbaccount.ShardId
	acID := dbaccount.String()
	//isHasReg, err := IsHasReg(acID)
	//if err != nil {
	//	panic(err)
	//}
	isHasReg := false
	logs.Trace("IsHasReg %s %v", acID, isHasReg)

	is_need_init = PanicIfErr(TryWarmData(true, shardID, &p.Profile, isHasReg))
	PanicIfErr(TryWarmData(true, shardID, &p.BagProfile, isHasReg))
	PanicIfErr(TryWarmData(true, shardID, &p.StoreProfile, isHasReg))
	PanicIfErr(TryWarmData(true, shardID, &p.GeneralProfile, isHasReg))
	PanicIfErr(TryWarmData(true, shardID, &p.GuildProfile, isHasReg))
	PanicIfErr(TryWarmData(true, shardID, &p.AntiCheat, isHasReg))
	PanicIfErr(TryWarmData(true, shardID, &p.Tmp, isHasReg))
	PanicIfErr(TryWarmData(true, shardID, &p.SimpleInfoProfile, isHasReg))
	PanicIfErr(TryWarmData(true, shardID, &p.Friend, isHasReg))

	p.rander = rand.New(&p.Profile.Rng)

	// 初始化各种Handle, 这个需要在所有的数据初始化之前初始化
	p.initHandle()
	if is_need_init {
		logs.Trace("[%s]Account Init", p.AccountID)
		p.OnAccountInit()
	}

	p.Profile.LoginTimes += 1
	monoLoginCounter.Inc(1)
	return p, is_need_init
}

// 加载账号所有信息 只是加载信息,如果数据库中没有的话就返回错误,不走任何初始化
func LoadFullAccount(dbaccount db.Account, needCheckReg bool) (*Account, error) {
	accountID := dbaccount
	p := &Account{
		AccountID:         accountID,
		Profile:           NewProfile(accountID),
		Tmp:               NewTmpProfile(accountID),
		BagProfile:        NewPlayerBag(accountID),
		StoreProfile:      NewPlayerStores(accountID),
		GeneralProfile:    NewPlayerGenerals(accountID),
		GuildProfile:      NewPlayerGuild(accountID),
		AntiCheat:         NewPlayerAntiCheat(accountID),
		SimpleInfoProfile: simple_info.NewSimpleInfoProfile(accountID),
		Friend:            NewFriend(accountID),
	}
	shardID := dbaccount.ShardId
	var isHasReg bool
	var err error
	//if needCheckReg {
	//	isHasReg, err = IsHasReg(dbaccount.String())
	//	if err != nil {
	//		return nil, err
	//	}
	//} else {
	//	isHasReg = true
	//}
	err = TryWarmData(true, shardID, &p.Profile, isHasReg)
	if err != nil {
		return nil, err
	}
	TryWarmData(true, shardID, &p.BagProfile, isHasReg)
	TryWarmData(true, shardID, &p.StoreProfile, isHasReg)
	TryWarmData(true, shardID, &p.GeneralProfile, isHasReg)
	TryWarmData(true, shardID, &p.GuildProfile, isHasReg)
	TryWarmData(true, shardID, &p.AntiCheat, isHasReg)
	TryWarmData(true, shardID, &p.Tmp, isHasReg)
	TryWarmData(true, shardID, &p.SimpleInfoProfile, isHasReg)
	TryWarmData(true, shardID, &p.Friend, isHasReg)

	return p, nil
}

// 加载Pvp账号信息 只是加载信息,如果数据库中没有的话就返回错误,不走任何初始化
func LoadPvPAccount(dbaccount db.Account) (*Account, error) {
	accountID := dbaccount
	p := &Account{
		AccountID:  accountID,
		Profile:    NewProfile(accountID),
		BagProfile: NewPlayerBag(accountID),
		//GeneralProfile: NewPlayerGenerals(accountID),
		GuildProfile: NewPlayerGuild(accountID),
	}
	shardID := dbaccount.ShardId
	//isHasReg, err := IsHasReg(dbaccount.String())
	//if err != nil {
	//	return nil, err
	//}
	isHasReg := false
	err := TryWarmData(true, shardID, &p.Profile, isHasReg)
	if err != nil {
		return nil, err
	}
	TryWarmData(true, shardID, &p.BagProfile, isHasReg)
	TryWarmData(true, shardID, &p.GuildProfile, isHasReg)
	//TryWarmData(true, shardID, &p.GeneralProfile, isHasReg)
	return p, nil
}

func (p *Account) GetRand() *rand.Rand {
	return p.rander
}

func (p *Account) GetDBErrChan() <-chan error {
	return p.dbError
}

func (p *Account) GetMsgChan() <-chan servers.Request {
	return p.msgChannel
}

func (p *Account) GetMsgNotifyChan() chan<- servers.Request {
	return p.msgChannel
}

func (p *Account) GetPushChan() <-chan game.INotifySyncMsg {
	return p.pushCh
}

func (p *Account) GetPushNotifyChan() chan game.INotifySyncMsg {
	return p.pushCh
}

func (p *Account) SendRespByPush(resp game.INotifySyncMsg) {
	select {
	case p.pushCh <- resp:
	default:
		logs.Warn("send push fail acid: %s push: %v", p.AccountID.String(), p.pushCh)
	}
}

func (p *Account) PreRequest(req servers.Request) {
	//senderr := func(err error) {
	//	select {
	//	case p.dbError <- err:
	//	default:
	//	}
	//}
	// 是否限制特定ip   by zhangzhen 限定都在auth上做吧
	//if security.LimitIpKick(p.ip) {
	//	logs.Warn("LimitIpKick ip %s", p.ip)
	//	senderr(fmt.Errorf("LimitIpKick"))
	//	return
	//}
	// ratelimit
	if err := security.Consume(security.GenSource(p.AccountID.String(), req.Code)); err != nil {
		logs.Warn("ratelimit illegal %s %s %s", p.AccountID.String(), req.Code, err.Error())
	}

	// 角色每次登陆在执行第一个协议之前，执行一次p.AfterAccountLogin()
	if p.requestSeq == 0 {
		p.afterAccountLogin()
		p.OnAccountOnline()
	}
	p.requestSeq++
}

func (p *Account) PostRequest(resp *servers.Response) {
	if !strings.HasPrefix(resp.Code, "Debug") {
		//请求处理完成了，需要增加计数器。此计数器是两次存盘之间的请求数量
		//防止因为机器人上限发送大量Debug消息，引起大量存盘
		atomic.AddUint32(&p.dbRequestCounter, 1)
	}

	// 每小时bilog记录活跃玩家数量，在PostRequest里因为需要玩家channel已经有了
	// 加时间判断，是为了一小时只写一次
	ts := hour_log.GetHourEndTS()
	if p.Profile.LastBIHourActiveTS < ts {
		hour_log.Get(p.AccountID.ShardId).OnActive(p.AccountID.String(), p.Profile.ChannelId)
		p.Profile.LastBIHourActiveTS = ts
	}

	//TODO dbRequestCounter is 20, 这个写操作计数可能需要配置
	if p.dbRequestCounter > 20 || resp.ForceDBChange {
		p.MakeDBSave(p.AccountID.ShardId, false)

		// 逻辑业务：检查是否需要更新好友模块信息
		if p.Profile.GetData().IsNeedUpdateFriend() {
			simpleInfo := p.GetSimpleInfo()
			logs.Debug("UpdateFriendInfo!!!")
			friend.GetModule(p.AccountID.ShardId).UpdateFriendInfo(&simpleInfo, 0)
			p.Profile.GetData().SetNeedUpdateFriend(false)
		}

	}
}

func (p *Account) MakeDBSave(shardId uint, forceDirty bool) error {
	// 玩家自身信息存储 用于其他模块读取
	simpleInfo := p.GetSimpleInfo()
	p.SimpleInfoProfile.SetFromOther(&simpleInfo)

	dbs := driver.GetDBService(shardId)
	dbs.MetricsCountDBSaves()
	senderr := func(err error) {
		select {
		case p.dbError <- err:
			//logs.Error("dbError Chan len %d", len(p.dbError))
		default:
			//logs.Trace("Too many errors in dbroutine")
			//太多的错误是不必要的，可以忽略掉
		}
	}
	if raw, err := p.redisSave(forceDirty); err != nil {
		senderr(err)
		dbs.LogDBError(p.AccountID.String(), err, raw)
		return err
	}
	return nil
}

func (p *Account) redisSave(forceDirty bool) ([]byte, error) {
	cb := redis.NewCmdBuffer()
	p.Profile.DBSave(cb, forceDirty)
	p.Tmp.DBSave(cb, forceDirty)
	p.BagProfile.DBSave(cb, forceDirty)
	p.StoreProfile.DBSave(cb, forceDirty)
	p.GeneralProfile.DBSave(cb, forceDirty)
	p.GuildProfile.DBSave(cb, forceDirty)
	p.AntiCheat.DBSave(cb, forceDirty)
	p.SimpleInfoProfile.DBSave(cb, forceDirty)
	p.Friend.DBSave(cb, forceDirty)

	raw := cb.Bytes()

	rc := driver.GetDBConn()
	defer rc.Close()
	if rc.IsNil() {
		logs.Error("Save Error:Account DB Save, cant get redis conn")
		return raw, fmt.Errorf("Account DB Save, cant get redis conn")
	}
	//logs.Trace("[%s]RedisSave %s", p.AccountID.String(), cb.String())

	_, err := rc.DoCmdBuffer(cb, true)
	if err != nil {
		logs.Error("Save Error:DoCmdBuffer, %s", err)
		return raw, err
	}

	atomic.StoreUint32(&p.dbRequestCounter, 0)

	return nil, nil
}

func (p *Account) SendMatchReq(isHard bool, isCancel bool) error {
	data := gveHelper.MatchValue{}
	data.AccountID = p.AccountID.String()
	data.IsHard = isHard
	data.CorpLv = p.Profile.GetCorp().GetLvlInfo()
	d, _ := json.Marshal(data)

	url := fmt.Sprintf("%s%s/%s", uutil.JwsCfg.MatchUrl, gveHelper.MatchPostUrlAddressV2,
		"oneboss", //boss
	)
	token := uutil.JwsCfg.MatchToken
	if token == "" {
		token = gveHelper.MatchDefaultToken
	}
	url += fmt.Sprintf("?token=%s", token)

	if isCancel {
		url += "&cancel=1"
	}

	_, err := util.HttpPost(url, util.JsonPostTyp, d)
	return err
}

func (p *Account) cancelGVEMatchOnLogout() {
	if p.Tmp.IsCurrWaittingGVE() {
		p.SendMatchReq(p.Tmp.GameIsHard, true)
		p.Tmp.CleanGVEData()
	}
}

func (p *Account) LeaveGVG() {
	if p.Tmp.GVGCity != 0 {
		gvg.GetModule(p.AccountID.ShardId).CommandExec(gvg.GVGCmd{
			Typ:    gvg.Cmd_Typ_LeaveCity,
			AcID:   p.AccountID.String(),
			CityID: p.Tmp.GVGCity,
		})
	}
}

func (p *Account) OnExit() {
	//p.Save()
	//p.dbHandler.Close()
	p.pushCh = nil
	p.OnAccountOffline()
	p.Profile.LogoutTime = p.Profile.GetProfileNowTime()
	if player_msg.GetModule(p.AccountID.ShardId).OnPlayerLogout(p.AccountID.String()) {
		hour_log.DelCCU(p.Profile.ChannelId)
	}
	// 结算本次登陆的在线时长
	nowT := p.Profile.GetProfileNowTime()
	LastLoginTime := p.Profile.LoginTime
	if !gamedata.IsSameDayCommon(p.Profile.LoginTime, nowT) {
		LastLoginTime = gamedata.GetCommonDayBeginSec(nowT)
	}
	if !gamedata.IsSameDayCommon(p.Profile.OnlineTimeCurrDayLastSetTime, nowT) {
		p.Profile.OnlineTimeCurrDay = 0
		p.Profile.OnlineTimeCurrDayLastSetTime = nowT
	}
	p.Profile.OnlineTimeCurrDay += (nowT - LastLoginTime)

	// 通知公会，下线了
	if p.GuildProfile.GuildUUID != "" {
		guild.GetModule(p.AccountID.ShardId).NoticeGuildWhenOffline(
			p.GuildProfile.GuildUUID, p.AccountID.String())
	}

	p.cancelGVEMatchOnLogout()

	nowTime := p.Profile.GetProfileNowTime()

	//组队boss 下线发离开房间协议
	logs.Debug("<TBoss> on exit leave team nowteamID : %v", p.Profile.GetTeamBossTeamInfo().NowTeamID)
	if p.Profile.GetTeamBossTeamInfo().NowTeamID != "" {
		leaveInfo := &helper2.LeaveRoomInfo{
			OptAcID: p.AccountID.String(),
			TgtAcID: p.AccountID.String(),
			RoomID:  p.Profile.GetTeamBossTeamInfo().NowTeamID,
		}
		ret, code, err := teamboss.LeaveRoom(p.AccountID.ShardId, p.AccountID.String(), leaveInfo)
		if code != crossservice.ErrOK {
			if ret.Code != 0 {
				logs.Error("<TBoss> on exit leave team code err %v", ret.Code)
			}
			logs.Error("<TBoss> on exit leave team crossservice err %s", err.Error())
		}
		p.Profile.GetTeamBossTeamInfo().TeamBossLeaveInfo.SetTeamBossLeaveInfo(0, p.Profile.GetTeamBossTeamInfo().NowTeamID, nowTime)
		p.Profile.GetTeamBossTeamInfo().NowTeamID = ""
	}

	// Room离开
	if p.Tmp.CurrRoomNum > 0 {
		go func(currRoomNum int) {
			ctx, cancel := context.WithTimeout(
				context.Background(),
				1*time.Microsecond)
			defer cancel()
			room.Get(p.AccountID.ShardId).LeaveRoom(ctx,
				p.AccountID.String(),
				"",
				currRoomNum)
		}(p.Tmp.CurrRoomNum)
		p.Tmp.CurrRoomNum = 0
	}
	simpleInfo := p.GetSimpleInfo()

	friend.GetModule(p.AccountID.ShardId).UpdateFriendInfo(&simpleInfo, p.Profile.LogoutTime)
	logs.Debug("log out name: %s, time: %d", p.Profile.Name, p.Profile.LogoutTime)
	// bi
	logiclog.LogLogout(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.LogoutTime-p.Profile.LoginTime, p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	// 世界boss
	wbd := p.Profile.GetWorldBossData()
	team := p.Profile.GetHeroTeams().GetHeroTeam(gamedata.LEVEL_TYPE_WORLD_BOSS)
	if len(team) >= 1 {
		if p.Profile.GetWorldBossData().State == world_boss.Battle {
			oldBuffLevel := wbd.BuffLevel
			if wbd.IsHadCostTimes() {
				wbd.BuffLevel = 0
			}
			isCheat := false
			maxDamage := world_boss.AntiCheatAllDamage(wbd.CurDamage, wbd.MaxHeroATK, oldBuffLevel,
				p.Profile.GetProfileNowTime()-wbd.StartBattleTime)
			if wbd.CurDamage > maxDamage {
				logs.Warn("player cheat, round damage: %v, maDamage: %v", wbd.CurDamage, maxDamage)
			}
			_, code, err := worldboss.Leave(p.AccountID.ShardId, p.AccountID.String(), p.GenTeamInfoDetail(team, oldBuffLevel), isCheat)
			if code != crossservice.ErrOK {
				logs.Error("end worldboss battle err, code: %d, err: %v", code, err)
			}
			wbd.State = world_boss.Idle
		}
	}

	p.MakeDBSave(p.AccountID.ShardId, true)
	close(p.quit)
}

func (p *Account) GenTeamInfoDetail(heros []int, buffLv int) *worldboss.TeamInfoDetail {
	extraAttrs := gs.GetCurrAttrForWB(NewAccountGsCalculateAdapter(p))
	baseGS := p.Profile.GetData().HeroBaseGs
	GS := p.Profile.GetData().HeroGs
	team := make([]worldboss.HeroInfoDetail, 0)
	for _, idx := range heros {
		team = append(team, worldboss.HeroInfoDetail{
			Idx:       int(idx),
			StarLevel: int(p.Profile.GetHero().GetStar(idx)),
			Level:     int(p.Profile.GetHero().HeroLevel[idx]),
			BaseGs:    int64(baseGS[idx]),
			ExtraGs:   int64(GS[idx] - baseGS[idx]),
		})
	}
	equipAttr := make([]int64, 0)
	destinyAttr := make([]int64, 0)
	jadeAttr := make([]int64, 0)
	for i := 0; i < len(extraAttrs[0]); i++ {
		equipAttr = append(equipAttr, int64(extraAttrs[0][i]))
		destinyAttr = append(destinyAttr, int64(extraAttrs[1][i]))
		jadeAttr = append(jadeAttr, int64(extraAttrs[2][i]))
	}
	logs.Debug("gen info detail, equipAttr: %v, destinyAttr: %v, jadeAttr: %v", equipAttr, destinyAttr, jadeAttr)
	return &worldboss.TeamInfoDetail{
		EquipAttr:   equipAttr,
		DestinyAttr: destinyAttr,
		JadeAttr:    jadeAttr,
		Team:        team,
		BuffLevel:   uint32(buffLv),
	}

}

// 获取当天在线时间
func (p *Account) GetOnlineTimeCurrDay() int64 {
	nowT := p.Profile.GetProfileNowTime()
	LastLoginTime := p.Profile.LoginTime
	if !gamedata.IsSameDayCommon(p.Profile.LoginTime, nowT) {
		LastLoginTime = gamedata.GetCommonDayBeginSec(nowT)
	}
	t := nowT - LastLoginTime
	if !gamedata.IsSameDayCommon(p.Profile.OnlineTimeCurrDayLastSetTime, nowT) {
		p.Profile.OnlineTimeCurrDayLastSetTime = nowT
		p.Profile.OnlineTimeCurrDay = 0
	}
	return t + p.Profile.OnlineTimeCurrDay
}

// 创建账号后一些初始化工作
func (p *Account) OnAccountInit() {
	var seed int64 = 5927
	const seedMask = 10000
	nowTime := time.Now().Unix()

	//根据注册时间对每个玩家设置不同随机数Seed, in prod
	if game.Cfg.IsRunModeProd() {
		seed = nowTime%(util.DaySec)*seedMask + rand.Int63n(seedMask)
	}

	//LOG MARKER
	client.LogInitData(p.AccountID.String(), seed)

	p.Profile.Rng.Seed(seed)
	p.Profile.InitQuest()
	p.Profile.GetGifts().Init()

	p.Profile.CreateTime = time.Now().Unix()
	//设置7日红包
	p.Profile.GetRedPacket7day().SetCreatTime(p.Profile.CreateTime)
	p.initAvatars()
	monoNewAccountCounter.Inc(1)

	p.Profile.GetCorp().OnAccountInit()
	p.Profile.GetPrivilegeBuy().InitPrivilegeInfo()
	p.Profile.GetClientTagInfo().Init()
	p.Profile.GetPlayerTrial().Init()

	// bi
	logiclog.LogCreateProfile(p.AccountID.String(), 0, "", 1, p.Profile.ChannelId,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	nowT := p.Profile.GetProfileNowTime()
	if !gamedata.IsSameDayCommon(p.Profile.OnlineTimeCurrDayLastSetTime, nowT) {
		p.Profile.OnlineTimeCurrDay = 0
		p.Profile.OnlineTimeCurrDayLastSetTime = nowT
	}

	p.Profile.GetEquips().OnAccountInit()

}

func (p *Account) afterAccountLogin() {
	profile := &p.Profile
	if profile.LoginTime <= 0 {
		profile.LoginTodayNum++
		profile.Logindaynum++
	} else {
		if !util.IsSameDayUnix(profile.LoginTime, profile.GetProfileNowTime()) {
			profile.Logindaynum++
		}
		if !gamedata.IsSameDayCommon(profile.LoginTime, profile.GetProfileNowTime()) {
			profile.LoginTodayNum = 1
		} else {
			profile.LoginTodayNum++
		}
	}

	now_t := profile.GetProfileNowTime()
	profile.LoginTime = now_t

	profile.GetAvatarSkill().OnAfterLogin()
	profile.GetQuest().RegCondition(&profile.conditons)
	profile.Story.OnAfterLogin(&profile.conditons)
	profile.GetQuest().DailyTaskReset(p)
	profile.GetQuest().UpdateCanReceiveList(p)
	profile.AutomationQuest(p)
	profile.GetBuy().OnAfterLogin()
	profile.GetIAPGoodInfo().InitGoodInfo()
	profile.GetClientTagInfo().Init()
	profile.GetDailyAwards().onAfterLogin(now_t)
	p.Profile.GetTeamPvp().OnAfterLogin()
	profile.GetCorp().OnAfterLogin()

	acID := p.AccountID.String()

	p.Profile.GetBoss().OnAfterLogin(p.rander)

	player_msg.GetModule(p.AccountID.ShardId).OnPlayerLogin(acID, p.msgChannel)

	profile.GetRecover().onAfterLogin(p, profile.LoginTime, profile.LogoutTime)

	profile.GetAccount7Day().onAfterLogin()
	profile.GetItemHistory().Init()

	profile.GetHero().onAfterLogin(p)

	profile.FirstPayReward.onAfterLogin()

	profile.GetHeroTalent().OnAfterLogin(now_t)
	//将星被动技能
	p.Profile.GetHero().CheckHeroSkill()

	//将星之路运营活动
	profile.GetHero().HeroStarActivity(p)

	profile.GetEquips().CheckEquipInBag(acID, p.Profile, p.BagProfile)
	//远征主将存档升级
	profile.GetExpeditionInfo().UpExpedition()

	profile.GetData().OnAfterLogin()
	simpleInfo := p.GetSimpleInfo()
	friend.GetModule(p.AccountID.ShardId).UpdateFriendInfo(&simpleInfo, 0)
	p.Profile.GetPrivilegeBuy().InitPrivilegeInfo()
	p.Profile.Vip.Update(p)
	//7日红包
	p.Profile.GetRedPacket7day().SetCreatTime(p.GetProfileNowTime())
	p.Profile.GetRedPacket7day().SendRedPacket7daysMail(p.AccountID.String(), p.GetProfileNowTime())

	p.Profile.GetWSPVPInfo().OnAfterLogin(p.AccountID.ShardId, p.AccountID.String())
	p.Profile.OfflineRecoverInfo.OnAfterLogin(p.Profile.LogoutTime, p.Profile.GetProfileNowTime())


	//各个排行榜更新，异步
	rank.GetModule(p.AccountID.ShardId).RankByEquipStarLv.Add(&simpleInfo)
	rank.GetModule(p.AccountID.ShardId).RankByJade.Add(&simpleInfo)
	rank.GetModule(p.AccountID.ShardId).RankByDestiny.Add(&simpleInfo)
	rank.GetModule(p.AccountID.ShardId).RankByCorpLv.Add(&simpleInfo)
	// 更新第二周宝石等级排行榜
	rank.GetModule(p.AccountID.ShardId).RankByHeroJadeTwo.Add(&simpleInfo)
	rank.GetModule(p.AccountID.ShardId).RankByWingStar.Add(&simpleInfo)
	rank.GetModule(p.AccountID.ShardId).RankByHeroDestiny.Add(&simpleInfo)
	rank.GetModule(p.AccountID.ShardId).RankByHeroWuShuangGs.Add(&simpleInfo)
	rank.GetModule(p.AccountID.ShardId).RankByAstrology.Add(&simpleInfo)
	rank.GetModule(p.AccountID.ShardId).RankByExclusiveWeapon.Add(&simpleInfo)
	p.UpdateCountryGs(&simpleInfo)

	//组队boss战 如果战斗后掉线，上线给他结束战斗发奖励
	globalRoomID := p.Profile.GetTeamBossTeamInfo().GlobalRoomId
	if globalRoomID != "" {
		EndfightInfo := &helper2.EndFightInfo{
			GlobalRoomID: globalRoomID,
			AcID:         p.AccountID.String(),
		}
		endFight, code, err := teamboss.EndFight(p.AccountID.ShardId, p.AccountID.String(), EndfightInfo)
		if code != crossservice.ErrOK {
			if endFight.Code != 0 {
				logs.Error("<TBoss> end battle code err %v", endFight.Code)
				logs.Error("<TBoss> end battle crossservice err %s", err.Error())
			}
		}
		boxInfo := p.Profile.GetTeamBossStorageInfo()
		if endFight.HasReward {
			mainData := gamedata.GetTBossMainDataByDiff(endFight.Level)
			if endFight.HasRedBox {
				redBoxId := gamedata.GetRedBoxId(endFight.Level)
				boxInfo.AddNewBox(redBoxId)
			} else {
				var rewardBoxId string
				boxCtrlTimes := boxInfo.BoxCtrlTimes
				boxCtrlCfg := gamedata.GetTBossVipCtrl(p.Profile.Vip.V)
				if boxCtrlTimes >= boxCtrlCfg.GetGoodBoxControl() {
					boxInfo.ResetControlTimes()
					rewardBoxId = gamedata.RandomTBBox(mainData.GetSepcialDropGroup())
				} else {
					boxInfo.IncreaseControlTimes()
					rewardBoxId = gamedata.RandomTBBox(mainData.GetBoxDropGroup())
				}
				boxInfo.AddNewBox(rewardBoxId)
			}

		}
		p.Profile.GetTeamBossTeamInfo().GlobalRoomId = ""
	}
}

func (p *Account) initHandle() {
	profile := &p.Profile
	handle := &p.handle
	profile.GetAvatarExp().SetHandler(handle)
	profile.GetCorp().SetHandler(handle)
	profile.GetHero().SetHandler(handle)
	profile.GetVip().SetHandler(handle)
	profile.GetEnergy().SetHandler(handle)
	profile.GetSC().SetHandler(handle)
	profile.GetHC().SetHandler(handle)
}

func (p *Account) GetRequestSeq() uint64 {
	return p.requestSeq
}

func (p *Account) GetIp() string {
	return p.ip
}

func (p *Account) AddHandle(h events.Handler) {
	p.handle.Add(h)
}

func (p *Account) GetHandle() events.Handler {
	return &p.handle
}

func (p *Account) GetSimpleInfo() helper.AccountSimpleInfo {
	profile := p.Profile
	lv, _ := profile.GetCorp().GetXpInfo()
	currAvatar := profile.GetCurrAvatar()
	nowTime := time.Now().Unix()
	var eqStartLvl uint32
	eqStartLvl = math.MaxUint32
	for slot := gamedata.PartID_Chest; slot < gamedata.PartEquipCount; slot++ {
		star := profile.GetEquips().GetStarLv(slot)
		if star < eqStartLvl {
			eqStartLvl = star
		}
	}

	var equipStarLv uint32
	for slot := gamedata.PartID_Weapon; slot < gamedata.PartEquipCount; slot++ {
		star := profile.GetEquips().GetStarLv(slot)
		equipStarLv += star
	}

	fashions := [helper.FashionPart_Count]string{}
	avatarFashions := profile.GetAvatarEquips().CurrByAvatar(currAvatar)
	for idx, id := range avatarFashions {
		if id > 0 {
			ok, item := p.Profile.GetFashionBag().GetFashionInfo(id)
			if ok && idx >= 0 && idx < len(fashions) {
				fashions[idx] = item.TableID
			}
		}
	}
	teamPvpAvatars := [helper.TeamPvpAvatarsCount]int{}
	teamPvpAvatarLvs := [helper.TeamPvpAvatarsCount]int{}
	hero := profile.GetHero()
	var tpsgs int
	for i, a := range profile.GetTeamPvp().FightAvatars {
		teamPvpAvatars[i] = a
		teamPvpAvatarLvs[i] = int(hero.GetStar(a))
		tpsgs += p.Profile.GetData().HeroGs[a]
	}
	bestHero, gsHeroGs, gsHeroBaseGs := profile.GetData().GetHeroGsInfo()
	jadeBag := p.Profile.GetJadeBag()
	jadeScore := 0
	for _, item := range profile.GetEquipJades().Jades {
		jadeItem := jadeBag.GetJade(item)
		if jadeItem != nil {
			lv := gamedata.GetJadeLvlByExp(jadeItem.JadeExp)
			if lv > 0 {
				jadeScore += int(math.Pow(float64(3), float64(lv)-1))
			}
		}
	}
	for _, item := range profile.GetDestGeneralJades().DestinyGeneralJade {
		jadeItem := jadeBag.GetJade(item)
		if jadeItem != nil {
			lv := gamedata.GetJadeLvlByExp(jadeItem.JadeExp)
			if lv > 0 {
				jadeScore += int(math.Pow(float64(3), float64(lv)-1))
			}
		}
	}
	var swingStarLv int
	for _, item := range profile.GetHero().HeroSwings {
		if item.StarLv > 0 {
			swingStarLv += item.StarLv
		}
	}
	var suShuangGs int
	sortWuGs := profile.GetData().GetSortHeroGsInfo()
	if len(sortWuGs) > 9 {
		for j := 0; j < 9; j++ {
			suShuangGs += sortWuGs[j]
		}
	} else {
		for _, x := range sortWuGs {
			suShuangGs += x
		}
	}
	var exclusiveWeapon int
	for i, item := range profile.GetHero().HeroExclusiveWeapon {
		if item.Quality != 0 {
			exclusiveWeapon += int(gamedata.GetAllNeedGWChips(item.Quality) * gamedata.GetRankForGwcParam(uint32(i)))
		}
	}

	return helper.AccountSimpleInfo{
		Name:            profile.Name,
		AccountID:       p.AccountID.String(),
		CorpLv:          lv,
		Vip:             p.Profile.GetVipLevel(),
		GuildSp:         p.Profile.GetSC().GetSC(gamedata.SC_GuildSp),
		GuildPosition:   p.GuildProfile.GetCurrPosition(),
		LastLoginTime:   profile.LoginTime,
		CurrAvatar:      currAvatar,
		CurrCorpGs:      profile.GetData().CorpCurrGS,
		FashionEquips:   fashions,
		WeaponStartLvl:  profile.GetEquips().GetStarLv(gamedata.PartID_Weapon),
		EqStartLvl:      eqStartLvl,
		InfoUpdateTime:  nowTime,
		MaxTrialLv:      int64(profile.GetPlayerTrial().MostLevelId),
		AvatarStarLvl:   profile.GetHero().HeroStarLevel,
		TeamPvpAvatar:   teamPvpAvatars,
		TeamPvpAvatarLv: teamPvpAvatarLvs,
		TeamPvpGs:       tpsgs,
		GuildName:       p.GuildProfile.GuildName,
		TitleOn:         profile.GetTitle().TitleTakeOn,
		TitleTimeOut:    profile.GetTitle().GetNextRefTime(),
		GsHeroIds:       bestHero,
		GsHeroGs:        gsHeroGs,
		GsHeroBaseGs:    gsHeroBaseGs,
		Swing:           profile.GetHero().GetSwing(currAvatar).CurSwing,
		MagicPetfigure:  profile.GetHero().GetMagicPetFigure(currAvatar),
		HeroDiffScore:   profile.GetHeroDiff().GetHeroDiffScore(),
		DestinyLv:       int64(profile.GetDestinyGeneral().GetAllDestinyLv()),
		JadeLv:          int64(jadeScore),
		EquipStarLv:     int64(equipStarLv),
		SwingStarLv:     int64(swingStarLv),
		HeroDestinyLv:   int64(profile.HeroDestiny.GetHeroDestinyAllLvl()),
		WuShuangGs:      int64(suShuangGs),
		Astrology:       int64(profile.GetAstrology().CalculateHerosSoulExp()),
		ExclusivWeapon:  int64(exclusiveWeapon),
		TopGsByCountry:  p.GetTopGsByCountry(),
	}
}

func (p *Account) GetTopGsByCountry() [helper.Country_Count]int64 {
	var ret [helper.Country_Count]int64
	bestInfo := p.GetBestHeroByCountry()
	for i, infoList := range bestInfo {
		for _, info := range infoList {
			ret[i] += int64(info.HeroGs)
		}
	}
	return ret
}

type BestHeroByCountry struct {
	HeroId     int
	HeroGs     int
	HeroBaseGs int
}

type BestHeroByType struct {
	HeroId        int
	HeroLevel     int
	HeroStarLevel int
	HeroBaseGs    int
	HeroGS        int
}

type BestHeroSliceForSort []BestHeroByCountry

type BestHeroSliceForType []BestHeroByType

func (b BestHeroSliceForSort) Len() int {
	return len(b)
}

func (b BestHeroSliceForSort) Less(i, j int) bool {
	return b[i].HeroGs > b[j].HeroGs
}

func (b BestHeroSliceForSort) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b BestHeroSliceForType) Len() int {
	return len(b)
}

func (b BestHeroSliceForType) Less(i, j int) bool {
	return b[i].HeroGS > b[j].HeroGS
}

func (b BestHeroSliceForType) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (p *Account) GetBestHeroByCountry() [helper.Country_Count][]BestHeroByCountry {
	allGs := p.Profile.GetData().HeroGs
	baseGs := p.Profile.GetData().HeroBaseGs
	var ret [helper.Country_Count][]BestHeroByCountry
	var tempGroup [helper.Country_Count][]BestHeroByCountry
	for i := 0; i < helper.Country_Count; i++ {
		tempGroup[i] = make([]BestHeroByCountry, 0)
	}
	for i, gs := range allGs {
		if gs > 0 {
			countryId := gamedata.GetHeroCountry(int(i))
			tempGroup[countryId] = append(tempGroup[countryId], BestHeroByCountry{
				HeroId:     i,
				HeroGs:     gs,
				HeroBaseGs: baseGs[i],
			})
		}
	}
	for i := 0; i < helper.Country_Count; i++ {
		sort.Sort(BestHeroSliceForSort(tempGroup[i]))
		logs.Info("gs by country info: %v", tempGroup[i])
		if len(tempGroup[i]) > 3 {
			ret[i] = tempGroup[i][0:3]
		} else {
			ret[i] = tempGroup[i]
		}
	}
	return ret
}

func (p *Account) GetBestHeroByType() [helper.HeroDiff_Count]BestHeroByType {
	allGs := p.Profile.GetData().HeroGs
	baseGs := p.Profile.GetData().HeroBaseGs
	heroLvl := p.Profile.GetHero().HeroLevel
	heroStar := p.Profile.GetHero().HeroStarLevel
	var ret [helper.HeroDiff_Count]BestHeroByType
	var tempGroup [helper.HeroDiff_Count][]BestHeroByType
	for i := 0; i < helper.HeroDiff_Count; i++ {
		tempGroup[i] = make([]BestHeroByType, 0)
	}
	for i, gs := range allGs {
		if gs > 0 {
			typeID := gamedata.GetHeroType(int(i))
			logs.Debug("gs by type id: %v", typeID)
			tempGroup[typeID-1] = append(tempGroup[typeID-1], BestHeroByType{
				HeroId:        i,
				HeroLevel:     int(heroLvl[i]),
				HeroStarLevel: int(heroStar[i]),
				HeroBaseGs:    baseGs[i],
				HeroGS:        gs,
			})
		}
	}
	for i := 0; i < helper.HeroDiff_Count; i++ {
		sort.Sort(BestHeroSliceForType(tempGroup[i]))
		logs.Info("gs by type info: %v", tempGroup[i])
		if len(tempGroup[i]) >= 1 {
			ret[i] = tempGroup[i][0]
		} else {
			ret[i] = BestHeroByType{}
		}
	}
	return ret
}

type ClientInfo struct {
	MarketVer     string `json:"MarketVer"`
	Build         string `json:"Build"`
	BuildHash     string `json:"BuildHash"`
	BuildBranch   string `json:"BuildBranch"`
	Data          string `json:"Data"`
	DataHash      string `json:"DataHash"`
	DataBranch    string `json:"DataBranch"`
	DeviceInfo    string `json:"DeviceInfo"`
	DeviceId      string `json:"DeviceId"`
	NetInfo       string `json:"NetInfo"`
	AtlasId       string `json:"AtlasId"`
	AccountName   string `json:"AccountName"`
	IsRegister    string `json:"IsRegister"`
	PhoneNum      string `json:"PhoneNum"`
	ChannelId     string `json:"ChannelId,omitempty"`
	ChannelUid    string `json:"ChannelUid,omitempty"`
	PlatformId    string `json:"PlatformId,omitempty"`
	DeviceToken   string `json:"DeviceToken,omitempty"`
	MemSize       string `json:"MemSize,omitempty"`
	ConnectReason string `json:"ConnectReason"`
	IDFA          string `json:"IDFA"`
	BundleUpdate  string `json:"BundleUpdate"`
	DataUpdate    string `json:"DataUpdate"`
}

func (p *Account) UpdateCountryGs(info *helper.AccountSimpleInfo) {
	logs.Info("<country gs> %v", info.TopGsByCountry)
	rankModule := rank.GetModule(p.AccountID.ShardId)
	if info.TopGsByCountry[helper.Country_Wei] > 0 {
		score := info.TopGsByCountry[helper.Country_Wei]
		rankModule.RankByCorpGsOfWei.Add(info, score, score) // TODO
	}
	if info.TopGsByCountry[helper.Country_Shu] > 0 {
		score := info.TopGsByCountry[helper.Country_Shu]
		rankModule.RankByCorpGsOfShu.Add(info, score, score) // TODO
	}
	if info.TopGsByCountry[helper.Country_Wu] > 0 {
		score := info.TopGsByCountry[helper.Country_Wu]
		rankModule.RankByCorpGsOfWu.Add(info, score, score) // TODO
	}
	if info.TopGsByCountry[helper.Country_Qun] > 0 {
		score := info.TopGsByCountry[helper.Country_Qun]
		rankModule.RankByCorpGsOfQunXiong.Add(info, score, score) // TODO
	}
}
