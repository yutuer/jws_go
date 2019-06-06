package gate

import (
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/cache"

	"fmt"

	gm "github.com/rcrowley/go-metrics"

	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/safecache"
	"vcs.taiyouxi.net/platform/planx/util/secure"
)

const defaultTokenAliveDur = 36 * 3600 //36 HOUR
const cachedToken = "CachedToken"

type sessionInfo struct {
	id         int64
	kickFunc   func(string, int, int)
	notifyFunc func(InfoNotifyToClient)
}

type loginInfo struct {
	//保存login服务器玩家getGate后通知过来的login token
	//这里的token都代表，玩家Maybe可能会来服务器
	loginCache cache.Cache // <loginToken, LoginNotify>

	//玩家使用login token完成Handshake后，这里会有token
	mu         sync.RWMutex
	autoIncId  int64
	sessionMap map[string]*sessionInfo // <accountId, sessionInfo>

	// 存储活跃的sessionId
	activeSessionCache cache.Cache // <sessionId, struct>
}

type LoginNotify struct {
	db.Account
	LoginToken string
	NumberID   int
}

func newLoginInfo() *loginInfo {
	l := &loginInfo{
		sessionMap: make(map[string]*sessionInfo),
	}
	cc, err := safecache.NewSafeCache("gate_cache", `{"interval":60}`)
	if err != nil {
		logs.Critical("Gate Server Startup Failed, due to: %s", err.Error())
		panic("Gate Server Startup Failed, by loginCache init failed")
	}
	l.loginCache = cc
	actCache, err := safecache.NewSafeCache("gate_act_cache", `{"interval":60}`)
	if err != nil {
		logs.Critical("Gate Server Startup Failed 1, due to: %s", err.Error())
		panic("Gate Server Startup Failed, by loginCache init failed 1")
	}
	l.activeSessionCache = actCache
	return l
}

//queryUserLoginInfo handshake用来调用查询是否这个Logintoken已经推送到当前gate
func (l *loginInfo) queryUserLoginInfo(loginToken string, agent *client.PacketConnAgent) (*LoginNotify, int64) {
	realLoginToken := loginToken
	// 重登的话客户端发过来的是加过re前缀的
	if strings.HasPrefix(loginToken, "re:") {
		realLoginToken = parseReloginToken(loginToken)
		if realLoginToken == "" {
			logs.Warn("<Gate> parse relogin token result %s", realLoginToken)
			return nil, 0
		}
	}
	// step 1: 检测 login token
	infoNotify := l.getAcidByLoginToken(realLoginToken)
	if infoNotify == nil {
		logs.Warn("<Gate> not find acid")
		return nil, 0
	}
	logs.Debug("<Gate> get acid by loginToken %v", infoNotify)
	acid := infoNotify.Account.String()

	// step 2: 设置新的连接信息
	newSessionInfo := &sessionInfo{
		kickFunc: getKickFunc(agent, acid),
		notifyFunc: func(info InfoNotifyToClient) {
			sendInfoNotify(agent.SendPacket, info)
		},
	}
	l.mu.Lock()
	oldSessionInfo, ok := l.sessionMap[acid]
	l.autoIncId++
	newSessionInfo.id = l.autoIncId
	l.sessionMap[acid] = newSessionInfo
	agent.SessionId = newSessionInfo.id
	sessionSize := len(l.sessionMap)
	l.mu.Unlock()
	if ok {
		oldSessionInfo.kickFunc(ReLogin_Kick, 0, 0)
	}
	// metrics
	logs.Info("update gate count metrics, %d", sessionSize)
	updateGaugeByGate(infoNotify.GameId, infoNotify.ShardId, int64(sessionSize))

	// step 3: 等待上次连接结束
	counter := 30
	kickSuccess := false
	for {
		if oldSessionInfo == nil {
			kickSuccess = true
			break
		}
		exist := l.activeSessionCache.IsExist(fmt.Sprintf("%d", oldSessionInfo.id))

		counter = counter - 1
		if counter <= 0 {
			//logs.Error("queryUserLoginInfo, relogin failed 3, %s", loginToken)
			return nil, 0
		}
		if !exist {
			//time.Sleep(time.Second * 2)
			kickSuccess = true
			break
		} else {
			logs.Trace("<Gate> queryUserLoginInfo waiting....")
			time.Sleep(time.Second)
		}
	}

	// step 4: 设置为活跃的sessionId
	if kickSuccess {
		l.activeSessionCache.Put(fmt.Sprintf("%d", newSessionInfo.id), 0, defaultTokenAliveDur)
	} else {
		// 如果没有T掉上一个连接， 这里不让登录
		logs.Error("<Gate> fail to kick old login, acid %s %s %d", acid, loginToken, newSessionInfo.id)
	}

	return infoNotify, newSessionInfo.id
}

func getKickFunc(agent *client.PacketConnAgent, acid string) func(string, int, int) {
	return func(reason string, after, nologin int) {
		if after <= 0 {
			after = 1
		}
		sendKickNotify(agent.SendPacket, reason, after, nologin)
		go func() {
			<-time.After(time.Duration(after) * time.Second)
			agent.Close()
			logs.Debug("<Gate> handleConnection kick agent close %s", acid)
		}()
	}
}

func parseReloginToken(reloginToken string) string {
	relogin := strings.TrimPrefix(reloginToken, "re:")
	loginTokenb, err := secure.Decode64FromNet(relogin)
	if err != nil {
		logs.Error("<Gate> queryUserLoginInfo, relogin failed 1, %s", reloginToken)
		return ""
	}
	return string(loginTokenb)
}

func (l *loginInfo) getAcidByLoginToken(loginToken string) *LoginNotify {
	logs.Debug("<Gate> loginCache %v", l.loginCache)
	v := l.loginCache.Get(loginToken)

	if v == nil {
		return nil
	} else {
		loginNotify := v.(*LoginNotify)
		return loginNotify
	}
}

//RegisterLoginNotify 通过login server通知rpc方式注册进来
// 通常玩家**准备**登录，在login server成功获取getGate后，此接口会被调用
func (l *loginInfo) RegisterLoginNotify(loginToken string, lt *LoginNotify) bool {
	logs.Trace("<Gate> loginInfo.RegisterLoginInfo coming! token got:%s", loginToken)
	if err := l.loginCache.Put(loginToken, lt, defaultTokenAliveDur); err != nil {
		logs.Error("<Gate> loginInfo.RegisterLoginNotify get Error: %s", err.Error())
		return false
	}
	l.loginCache.Incr(cachedToken)
	return true
}

/////////////////////////

func (l *loginInfo) GetSession(acid string) (*sessionInfo, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	f, ok := l.sessionMap[acid]
	return f, ok
}

func (l *loginInfo) GetInfoNotifyHandler(accountId string) (func(InfoNotifyToClient), bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	f, ok := l.sessionMap[accountId]
	if ok {
		return f.notifyFunc, ok
	} else {
		return nil, ok
	}
}

///////////////////////////////////

func (g *GateServer) RegisterLoginToken(args *LoginNotify, reply *bool) error {
	logs.Trace("<Gate> Happy!! Gate RPC: user_id %d logged in with loginToken %s", args.UserId, args.LoginToken)
	*reply = g.info.RegisterLoginNotify(args.LoginToken, args)
	return nil
}

const (
	ReLogin_Kick = "IDS_KICKNOTIFY_NEWLOGIN"
	GM_Kick      = "IDS_GM_CLOSEDOWN"
)

type KickOfflineParam struct {
	LoginToken      string
	AccountID       string
	Reason          string
	AfterDuration   int
	NoLoginDuration int
}

// KickItOffline 返回值如果是1, 发现这个玩家，并要求其断线
// 用途1：login服务器收到同一个帐号的登录信息，会给已经登录的人发送Kick RPC
// 用途2：客服系统可能需要对某些玩家进行强制踢出游戏测操作，并禁止登录N分钟
// 用途3：系统运维前，需要提出所有在线玩家，强迫下线
func (g *GateServer) KickItOffline(args *KickOfflineParam, reply *int) error {
	logs.Info("<Gate> GateServer got kick info %v", args)
	*reply = 0

	if args == nil {
		return nil
	}
	session, ok := g.info.GetSession(args.AccountID)
	if ok {
		exist := g.info.activeSessionCache.IsExist(fmt.Sprintf("%d", session.id))
		if exist {
			session.kickFunc(args.Reason, args.AfterDuration, args.NoLoginDuration)
			*reply = 1
			logs.Info("<Gate> GateServer kick %s success, %v", args.AccountID, args)
		}
	}

	return nil
}

type NotifyInfoParam struct {
	LoginToken string
	AccountID  string
	Info       InfoNotifyToClient
}

// 向玩家Client推送信息
// 用途1：若玩家被禁言,则通过这个通知客户端
func (g *GateServer) NotifyInfo(args *NotifyInfoParam, reply *int) error {
	logs.Info("GateServer got kick info %v", args)
	*reply = 0

	if args == nil {
		return nil
	}

	h, ok := g.info.GetInfoNotifyHandler(args.AccountID)
	if ok {
		if h != nil {
			logs.Info("GateServer NotifyInfo %s success, %v", args.AccountID, args)
			h(args.Info)
		}

		*reply = 1
	}

	return nil
}

var gateMetricLock sync.RWMutex
var gateMetrics gm.Gauge

func updateGaugeByGate(gid, sid uint, value int64) {
	metricsKey := fmt.Sprintf("g%d.gamex%d.gatecount", gid, sid)
	logs.Debug("<gate count metics>%s", metricsKey)
	updateGauge(metricsKey, value)
}

func updateGauge(name string, value int64) {
	gateMetricLock.Lock()
	defer gateMetricLock.Unlock()
	if gateMetrics == nil {
		gateMetrics = metrics.NewCustomGauge(name)
	}
	gateMetrics.Update(value)
}
