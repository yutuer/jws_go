package gates_enemy

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/gates_enemy/cmd"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func genGatesEnemyModule(sid uint) *GatesEnemyModule {
	return &GatesEnemyModule{
		sid: sid,
	}
}

type GatesEnemyModule struct {
	sid          uint
	waitter      sync.WaitGroup
	acts         map[string]*GatesEnemyActivity
	actsOverTime map[string]int64 // <guildUUID, endTime>
	cmdChannel   chan GatesEnemyCommandMsg
	quitChan     chan bool
}

func (r *GatesEnemyModule) AfterStart(g *gin.Engine) {
}

func (r *GatesEnemyModule) BeforeStop() {
}

func (r *GatesEnemyModule) Start() {
	r.waitter.Add(1)
	defer r.waitter.Done()
	r.acts = make(map[string]*GatesEnemyActivity, 1024)
	r.actsOverTime = make(map[string]int64, 1024)
	r.cmdChannel = make(chan GatesEnemyCommandMsg, 64)
	r.quitChan = make(chan bool, 1)

	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		timerChan := uutil.TimerMS.After(time.Second)
		for {
			select {
			case command, _ := <-r.cmdChannel:
				logs.Trace("command %v", command)
				r.waitter.Add(1)
				r.processCmd(&command)
			case <-timerChan:
				timerChan = uutil.TimerMS.After(time.Second)
				r.waitter.Add(1)
				func() {
					defer r.waitter.Done()
					r.checkGatesEnemyActStop()
				}()
			case <-r.quitChan:
				return
			}
		}
	}()
}

func (r *GatesEnemyModule) Stop() {
	r.quitChan <- true
	r.waitter.Wait()
}

func (r *GatesEnemyModule) readyGatesEnemyAct(guildID string, endTime int64,
	members []helper.AccountSimpleInfo) {

	logs.Debug("GatesEnemyModule readyGatesEnemyAct %v", guildID)

	a, ok := r.acts[guildID]
	if !ok {
		newAct := r._startGatesEnemy(guildID, endTime, members)
		a = newAct
	}
	a.needPush = true
	return
}

func (r *GatesEnemyModule) startGatesEnemyAct(guildID string, endTime int64,
	members []helper.AccountSimpleInfo,
	resChan chan<- chan<- GatesEnemyCommandMsg) {

	act := r.startGatesEnemyActASync(guildID, endTime, members)
	if act == nil {
		resChan <- nil
		return
	}
	resChan <- act.GetCommandChan()
	return
}

func (r *GatesEnemyModule) startGatesEnemyActASync(guildID string, endTime int64,
	members []helper.AccountSimpleInfo) *GatesEnemyActivity {

	a, ok := r.acts[guildID]
	if ok && a.state == GatesEnemyActivityStateStarted {
		logs.Warn("StartGatesEnemyAct Err By Has Act %s Now!", guildID)
		return nil
	}

	logs.Debug("GatesEnemyModule startGatesEnemyAct %v", guildID)

	newAct := a
	if newAct == nil {
		newAct = r._startGatesEnemy(guildID, endTime, members)
	}
	newAct.state = GatesEnemyActivityStateStarted
	newAct.needPush = true
	return newAct
}

func (r *GatesEnemyModule) _startGatesEnemy(guildID string, endTime int64,
	members []helper.AccountSimpleInfo) *GatesEnemyActivity {

	newAct := &GatesEnemyActivity{}
	newAct.cmdChannel = make(chan GatesEnemyCommandMsg, 64)
	newAct.members = append([]helper.AccountSimpleInfo{}, members...)
	newAct.buffCurLv = 0
	newAct.buffMemAcid = [helper.GateEnemyBuffCount]string{}
	newAct.buffMemName = [helper.GateEnemyBuffCount]string{}
	newAct.init(endTime)
	newAct.Start(guildID)
	r.acts[guildID] = newAct
	r.actsOverTime[guildID] = endTime

	return newAct
}

func (r *GatesEnemyModule) stopGatesEnemyAct(guildID string, isGuildDismiss bool) *guild_info.GuildSimpleInfo {
	var resSimpleInfo *guild_info.GuildSimpleInfo
	act, ok := r.acts[guildID]
	if ok {
		// Notify Guild
		logs.Trace("stopGatesEnemyAct %s", guildID)
		act.Stop()

		// 如果是公会解散的话不要有任何结束操作
		// 如果不是, 则要向公会发送最后的结果
		// 并且更新排行榜
		if !isGuildDismiss {
			sid, _ := guild_info.GetShardIdByGuild(guildID)
			res, info := guild.GetModule(sid).OnGateEnemyStop(guildID,
				act.GetMemSyncReceiverID(),
				act.mkDataToPlayers())
			if res.HasError() {
				logs.Error("stopGatesEnemyAct %s %v err by %v",
					guildID, act.mkDataToPlayers(), res)
			}

			if info != nil && info.GatesEnemyData.Point > 0 {
				//rank.GetModule(sid).RankGuildGateEnemy.Add(&info.Base)
				resSimpleInfo = &info.Base
			}
		}

		delete(r.acts, guildID)
		delete(r.actsOverTime, guildID)
	}
	return resSimpleInfo
}

func (r *GatesEnemyModule) onPlayerIntoAct(accountID string,
	guildID string,
	info helper.AccountSimpleInfo,
	resChann chan<- chan<- GatesEnemyCommandMsg) {
	act, ok := r.acts[guildID]
	if ok {
		act.OnPlayerIntoAct(accountID, info)
		if resChann != nil {
			resChann <- act.GetCommandChan()
		}
	} else {
		if resChann != nil {
			resChann <- nil
		}
	}
}

func (r *GatesEnemyModule) checkGatesEnemyActStop() {
	guilds := make([]*guild_info.GuildSimpleInfo, 0, len(r.actsOverTime))
	timeNow := time.Now().Unix()
	for actID, stopTime := range r.actsOverTime {
		if stopTime < timeNow {
			g := r.stopGatesEnemyAct(actID, false)
			if g != nil {
				guilds = append(guilds, g)
			}
		}
	}
	if len(guilds) > 0 {
		logs.Debug("checkGatesEnemyActStop %v", guilds)
		rank.GetModule(r.sid).RankGuildGateEnemy.Adds(guilds)
	}
}

func (r *GatesEnemyModule) getActChan(guildID string,
	resChann chan<- chan<- GatesEnemyCommandMsg) {
	logs.Trace("getActChan %s", guildID)
	act, ok := r.acts[guildID]
	if ok && act.state == GatesEnemyActivityStateStarted {
		logs.Trace("getActChan chan")
		resChann <- act.GetCommandChan()
	} else {
		logs.Trace("getActChan res nil")
		resChann <- nil
	}
}

func (r *GatesEnemyModule) processCmd(msg *GatesEnemyCommandMsg) {
	defer r.waitter.Done()
	if msg == nil {
		return
	}
	switch msg.Type {
	case GatesEnemyCommandMsgTypStartAct:
		r.startGatesEnemyAct(msg.GuildID, msg.EndTime, msg.Members, msg.ResChann)
	case GatesEnemyCommandMsgTypStartActASync:
		r.startGatesEnemyActASync(msg.GuildID, msg.EndTime, msg.Members)
	case GatesEnemyCommandMsgTypReadyAct:
		r.readyGatesEnemyAct(msg.GuildID, msg.EndTime, msg.Members)
	case GatesEnemyCommandMsgTypStopAct:
		r.stopGatesEnemyAct(msg.GuildID, true)
	case GatesEnemyCommandMsgTypGetAct:
		r.onPlayerIntoAct(
			msg.AccountID,
			msg.GuildID,
			msg.Members[0],
			msg.ResChann)
	case GatesEnemyCommandMsgTypGetActChan:
		r.getActChan(msg.GuildID, msg.ResChann)
	case GatesEnemyCommandMsgTypPlayerLogin:
		r.onPlayerIntoAct(
			msg.AccountID,
			msg.GuildID,
			msg.Members[0],
			msg.ResChann)
	default:
		logs.Error("GatesEnemyModule processCmd Err By %v", msg)
	}
}

func (r *GatesEnemyModule) StartGatesEnemyAct(guildID string, endTime int64,
	members []helper.AccountSimpleInfo) bool {
	resChan := make(chan chan<- GatesEnemyCommandMsg, 1)
	msg := &GatesEnemyCommandMsg{
		Type:     GatesEnemyCommandMsgTypStartAct,
		GuildID:  guildID,
		EndTime:  endTime,
		Members:  members,
		ResChann: resChan,
	}
	r.sendCmd(msg, PlayerWaitSeconds)
	res := <-resChan
	return res != nil
}

func (r *GatesEnemyModule) StartGatesEnemyActASync(guildID string, endTime int64,
	members []helper.AccountSimpleInfo) {

	msg := &GatesEnemyCommandMsg{
		Type:    GatesEnemyCommandMsgTypStartActASync,
		GuildID: guildID,
		EndTime: endTime,
		Members: members,
	}
	r.sendCmd(msg, PlayerWaitSeconds)
}

func (r *GatesEnemyModule) ReadyGatesEnemyActASync(guildID string,
	endTime int64,
	members []helper.AccountSimpleInfo) {

	msg := &GatesEnemyCommandMsg{
		Type:    GatesEnemyCommandMsgTypReadyAct,
		GuildID: guildID,
		EndTime: endTime,
		Members: members,
	}
	r.sendCmd(msg, PlayerWaitSeconds)
}

// 由Guild方面调用, 结束一个GatesEnemy活动, 一般是因为公会解散, 因为公会解散情况并不多, 所以每次公会解散都会发包
func (r *GatesEnemyModule) StopGatesEnemyAct(guildID string) {
	logs.Trace("StopGatesEnemyAct %s", guildID)
	msg := &GatesEnemyCommandMsg{
		Type:    GatesEnemyCommandMsgTypStopAct,
		GuildID: guildID,
	}
	r.sendCmd(msg, 0)
}

func (r *GatesEnemyModule) OnPlayerIntoAct(accountID string,
	guildID string,
	info helper.AccountSimpleInfo,
	resChann chan<- chan<- GatesEnemyCommandMsg) {
	msg := &GatesEnemyCommandMsg{
		Type:      GatesEnemyCommandMsgTypPlayerLogin,
		AccountID: accountID,
		GuildID:   guildID,
		ResChann:  resChann,
	}
	msg.Members = []helper.AccountSimpleInfo{info}
	r.sendCmd(msg, PlayerWaitSeconds)
}

func (r *GatesEnemyModule) IsActHasStart(guildID string) bool {
	resChan := make(chan chan<- GatesEnemyCommandMsg, 1)
	msg := &GatesEnemyCommandMsg{
		Type:     GatesEnemyCommandMsgTypGetActChan,
		GuildID:  guildID,
		ResChann: resChan,
	}
	r.sendCmd(msg, PlayerWaitSeconds)
	res := <-resChan
	return res != nil
}

func (r *GatesEnemyModule) sendCmd(msg *GatesEnemyCommandMsg, wait int) {
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
	logs.Trace("[GatesEnemyActivity] sendCmd %v", msg.Type)
}
