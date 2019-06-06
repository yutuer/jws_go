package gates_enemy

import (
	"time"

	"sync"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/gates_enemy/cmd"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const PlayerWaitSeconds = 5

const (
	GatesEnemyActivityStateNull = iota
	GatesEnemyActivityStateNoBegin
	GatesEnemyActivityStateWaitStart
	GatesEnemyActivityStateStarted
	GatesEnemyActivityStateEnding
	GatesEnemyActivityStateStarting
)

const (
	GatesEnemyMemStateNull = iota
	GatesEnemyMemStateExit // 不在
	GatesEnemyMemStateFight
	GatesEnemyMemStateWait
)

type GatesEnemyActivity struct {
	cmdChannel chan GatesEnemyCommandMsg
	members    []helper.AccountSimpleInfo

	enemyInfo           [][]byte
	enemyLastUpdateTime []int64
	state               int
	stateOverTime       int64
	killPoint           int
	gePointAll          int
	bossMax             int

	memberData map[string]*MemberInGatesEnemyActivityData

	buffCurLv   uint32
	buffMemAcid [helper.GateEnemyBuffCount]string
	buffMemName [helper.GateEnemyBuffCount]string

	waitter      sync.WaitGroup
	quitChan     chan bool
	currSubState int // 中间状态标记
	lastPushTime int64
	needPush     bool
}

func (g *GatesEnemyActivity) init(endTime int64) {
	g.state = GatesEnemyActivityStateWaitStart
	g.stateOverTime = endTime
	g.killPoint = 0
	g.bossMax = 0
	g.initEnemys()
}

func (g *GatesEnemyActivity) Start(guildID string) {
	logs.Warn("GatesEnemyActivity start %s", guildID)

	g.waitter.Add(1)
	defer g.waitter.Done()
	g.quitChan = make(chan bool, 1)

	g.waitter.Add(1)

	sid, _ := guild_info.GetShardIdByGuild(guildID)
	guild.GetModule(sid).AddMemSyncReceiver(guildID, g)
	g.updateMemberInGatesEnemyActivityData()
	g.SendDataToPlayers()
	go func() {
		defer g.waitter.Done()
		timerChan := uutil.TimerMS.After(time.Second)
		for {
			nowT := time.Now().Unix()
			select {
			case command, _ := <-g.cmdChannel:
				g.waitter.Add(1)
				func() {
					defer g.waitter.Done()
					logs.Trace("command %v", command)
					g.processCmd(&command)
				}()
			case <-timerChan:
				timerChan = uutil.TimerMS.After(time.Second)
				g.waitter.Add(1)
				func() {
					defer g.waitter.Done()
					//g.checkGatesEnemyActStop(nowT)
					g.updateGatesEnemy(nowT)
					g.checkIsPush(nowT)
				}()
			case <-g.quitChan:
				g.state = GatesEnemyActivityStateNoBegin
				g.stateOverTime = nowT
				g.needPush = true
				g.checkIsPush(nowT + 32) // 强制push下去
				logs.Warn("GatesEnemyActivity over %s", guildID)
				return
			}
		}
	}()
}

func (g *GatesEnemyActivity) Stop() {
	g.quitChan <- true
	g.waitter.Wait()
}

// 处理逻辑
func (r *GatesEnemyActivity) onGuildChange(members []helper.AccountSimpleInfo) {
	logs.Trace("onGuildChange %v", members)
	r.members = append([]helper.AccountSimpleInfo{}, members...)
	r.updateMemberInGatesEnemyActivityData()
	r.needPush = true
}

func (r *GatesEnemyActivity) onPlayerIntoAct(accountID string,
	info helper.AccountSimpleInfo) {
	r.needPush = true
}

func (r *GatesEnemyActivity) onPlayerOutAct(accountID string) {

}

func (g *GatesEnemyActivity) processCmd(msg *GatesEnemyCommandMsg) {
	if msg == nil {
		logs.Error("processCmd msg nil")
		return
	}
	switch msg.Type {
	case GatesEnemyCommandMsgTypGuildChange:
		g.onGuildChange(msg.Members)
	case GatesEnemyCommandMsgTypPlayerLogin:
		g.onPlayerIntoAct(msg.AccountID, msg.Members[0])
	case GatesEnemyCommandMsgTypPlayerLogout:
		g.onPlayerOutAct(msg.AccountID)
	case GatesEnemyCommandMsgTypFightBegin:
		g.onFightBegin(msg.AccountID, msg.Members[0], msg.EnemyTyp, msg.EnemyIdx, msg.OkChann)
	case GatesEnemyCommandMsgTypFightEnd:
		g.onFightEnd(msg.AccountID, msg.Members[0], msg.EnemyTyp, msg.EnemyIdx, msg.IsSuccess, msg.OkChann)
	case GatesEnemyCommandMsgTypFightBossBegin:
		g.onFightBossBegin(msg.AccountID, msg.Members[0], msg.BossIdx, msg.OkChann)
	case GatesEnemyCommandMsgTypFightBossEnd:
		g.onFightBossEnd(msg.AccountID, msg.Members[0], msg.BossIdx, msg.IsSuccess, msg.OkChann)
	case GatesEnemyCommandMsgTypEnterAct:
		g.onEnterAct(msg.AccountID, msg.Members[0], msg.OkChann)
	case GatesEnemyCommandMsgTypLeaveAct:
		g.onLeaveAct(msg.AccountID, msg.OkChann)
	case GatesEnemyCommandMsgTypAddBuff:
		g.addBuff(msg.AccountID, msg.Members[0].Name, msg.OkChann)
	case GatesEnemyCommandMsgTypDebugOp:
		g.onDebugOp(msg.AccountID,
			msg.GuildID,
			msg.Members[0].InfoUpdateTime,
			msg.Members[1].InfoUpdateTime,
			msg.Members[2].InfoUpdateTime,
			msg.OkChann)
	default:
		logs.Error("GatesEnemyActivity processCmd Err By %v", msg)
	}
}

func (g *GatesEnemyActivity) checkIsPush(nowT int64) {
	if nowT-g.lastPushTime < 1 {
		return
	}
	if g.needPush || nowT-g.lastPushTime >= 10 { // 10s会定时给所有人发推送
		g.SendDataToPlayers()
		g.needPush = false
		g.lastPushTime = nowT
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *GatesEnemyActivity) sendCmd(msg *GatesEnemyCommandMsg, wait int) {
	if wait > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(wait)*time.Second)
		defer cancel()

		select {
		case r.cmdChannel <- *msg:
		case <-ctx.Done():
			logs.Warn("[GatesEnemyActivity] sendCmd chann full, cmd put timeout")
			return
		}
	} else {
		select {
		case r.cmdChannel <- *msg:
		default:
			logs.Warn("[GatesEnemyActivity] sendCmd chann full, cmd put timeout")
			return
		}
	}
	logs.Trace("[GatesEnemyActivity] sendCmd %v", *msg)
}

func (r *GatesEnemyActivity) OnGuildChange(members []helper.AccountSimpleInfo) {
	r.sendCmd(&GatesEnemyCommandMsg{
		Type:    GatesEnemyCommandMsgTypGuildChange,
		Members: members,
	}, 0)
}

func (r *GatesEnemyActivity) OnPlayerIntoAct(accountID string,
	info helper.AccountSimpleInfo) {
	msg := &GatesEnemyCommandMsg{
		Type:      GatesEnemyCommandMsgTypPlayerLogin,
		AccountID: accountID,
	}
	msg.Members = []helper.AccountSimpleInfo{info}
	r.sendCmd(msg, PlayerWaitSeconds)
}

func (r *GatesEnemyActivity) OnPlayerOutAct(accountID string) {
	r.sendCmd(&GatesEnemyCommandMsg{
		Type:      GatesEnemyCommandMsgTypPlayerLogout,
		AccountID: accountID,
	}, PlayerWaitSeconds)
}

func (r *GatesEnemyActivity) OnPlayerEnterAct(accountID string,
	info helper.AccountSimpleInfo) {
	msg := &GatesEnemyCommandMsg{
		Type:      GatesEnemyCommandMsgTypEnterAct,
		AccountID: accountID,
	}
	msg.Members[0] = info
	r.sendCmd(msg, PlayerWaitSeconds)
}

func (r *GatesEnemyActivity) OnPlayerLeaveAct(accountID string) {
	r.sendCmd(&GatesEnemyCommandMsg{
		Type:      GatesEnemyCommandMsgTypLeaveAct,
		AccountID: accountID,
	}, PlayerWaitSeconds)
}

func (g *GatesEnemyActivity) GetCommandChan() chan<- GatesEnemyCommandMsg {
	return g.cmdChannel
}

func (g *GatesEnemyActivity) GetMemSyncReceiverID() int {
	return helper.MemSyncReceiverIDGateEnemyAct
}
