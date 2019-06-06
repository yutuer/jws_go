package account

import (
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/gates_enemy"
	"vcs.taiyouxi.net/jws/gamex/modules/gates_enemy/cmd"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type playerGatesEnemyData struct {
	pushData     player_msg.PlayerMsgGatesEnemyData
	CurrBoss     []byte `json:"cb"`
	CurrActTime  int64  `json:"cst"`
	ActBeginTime int64  `json:"act_begin_ts"`
	ActEndTime   int64  `json:"act_end_ts"`

	isHasGetChannel bool
	cmdChannel      chan<- gates_enemy_cmd.GatesEnemyCommandMsg
}

func (p *playerGatesEnemyData) CleanDatas() {
	p.CurrBoss = make([]byte, 32, 32)
	logs.Warn("p.CurrBoss %v", p.CurrBoss)
	for i := 0; i < len(p.CurrBoss); i++ {
		p.CurrBoss[i] = 0
	}
	p.CurrActTime = 0
	p.isHasGetChannel = false
	p.pushData = player_msg.PlayerMsgGatesEnemyData{}
	p.pushData.State = gates_enemy.GatesEnemyActivityStateNoBegin
}

func (p *playerGatesEnemyData) OnPushData(shardID uint, acID, guildID string,
	info helper.AccountSimpleInfo,
	pushData *player_msg.PlayerMsgGatesEnemyData) {
	p.pushData = *pushData
	logs.Debug("Test Gates Enemy: state %d", pushData.State)
	// 关闭时清空数据
	if pushData.State == gates_enemy.GatesEnemyActivityStateNoBegin {
		p.isHasGetChannel = false
		p.CurrActTime = 0
	} else {
		logs.Debug("Test Gates Enemy: OnPushData")
		p.InitGetChannel(shardID, acID, guildID, info)
	}

	if pushData.StateOverTime != p.CurrActTime {
		p.CurrBoss = make([]byte, 32, 32)
		p.CurrActTime = pushData.StateOverTime
		p.isHasGetChannel = false
	}
}

func (p *playerGatesEnemyData) InitGetChannel(shardID uint, acID, guildID string,
	info helper.AccountSimpleInfo) {
	logs.Debug("Test Gates Enemy: InitGetChannle")
	if !p.isHasGetChannel {
		resChan := make(chan chan<- gates_enemy_cmd.GatesEnemyCommandMsg, 1)
		gates_enemy.GetModule(shardID).OnPlayerIntoAct(acID, guildID, info, resChan)
		channel := <-resChan
		if channel != nil {
			p.cmdChannel = channel
			p.isHasGetChannel = true
		}
	}
}

func (p *playerGatesEnemyData) GetPushData() *player_msg.PlayerMsgGatesEnemyData {
	return &p.pushData
}

func (p *playerGatesEnemyData) SetPushData(data player_msg.PlayerMsgGatesEnemyData) {
	p.pushData = data
}

func (r *playerGatesEnemyData) OnFightBegin(accountID string,
	info helper.AccountSimpleInfo,
	enemyTyp, enemyIDx int) uint32 {
	msg := &gates_enemy_cmd.GatesEnemyCommandMsg{
		Type:      gates_enemy_cmd.GatesEnemyCommandMsgTypFightBegin,
		AccountID: accountID,
		EnemyIdx:  enemyIDx,
		EnemyTyp:  enemyTyp,
	}
	msg.Members = []helper.AccountSimpleInfo{info}
	return r.waitResCode(r.sendCmd(msg))
}

func (r *playerGatesEnemyData) OnFightEnd(accountID string,
	info helper.AccountSimpleInfo,
	enemyTyp, enemyIDx int, isSuccess bool) uint32 {
	msg := &gates_enemy_cmd.GatesEnemyCommandMsg{
		Type:      gates_enemy_cmd.GatesEnemyCommandMsgTypFightEnd,
		AccountID: accountID,
		EnemyIdx:  enemyIDx,
		EnemyTyp:  enemyTyp,
		IsSuccess: isSuccess,
	}
	msg.Members = []helper.AccountSimpleInfo{info}
	return r.waitResCode(r.sendCmd(msg))
}

func (r *playerGatesEnemyData) OnEnterAct(accountID string,
	info helper.AccountSimpleInfo) uint32 {
	msg := &gates_enemy_cmd.GatesEnemyCommandMsg{
		Type:      gates_enemy_cmd.GatesEnemyCommandMsgTypEnterAct,
		AccountID: accountID,
	}
	msg.Members = []helper.AccountSimpleInfo{info}
	return r.waitResCode(r.sendCmd(msg))
}

func (r *playerGatesEnemyData) OnLeaveAct(accountID string) uint32 {
	msg := &gates_enemy_cmd.GatesEnemyCommandMsg{
		Type:      gates_enemy_cmd.GatesEnemyCommandMsgTypLeaveAct,
		AccountID: accountID,
	}
	return r.waitResCode(r.sendCmd(msg))
}

func (r *playerGatesEnemyData) OnAddBuff(accountID, name string) gates_enemy_cmd.GatesEnemyRet {
	msg := &gates_enemy_cmd.GatesEnemyCommandMsg{
		Type:      gates_enemy_cmd.GatesEnemyCommandMsgTypAddBuff,
		AccountID: accountID,
	}
	msg.Members = make([]helper.AccountSimpleInfo, 1)
	msg.Members[0].Name = name
	return r.waitRet(r.sendCmd(msg))
}

func (r *playerGatesEnemyData) OnFightBossBegin(accountID string,
	info helper.AccountSimpleInfo,
	bossIDx int) uint32 {
	if bossIDx >= len(r.CurrBoss) || r.CurrBoss[bossIDx] > 0 {
		return gates_enemy_cmd.RESErrStateErr
	}

	msg := &gates_enemy_cmd.GatesEnemyCommandMsg{
		Type:      gates_enemy_cmd.GatesEnemyCommandMsgTypFightBossBegin,
		AccountID: accountID,
		BossIdx:   bossIDx,
	}
	msg.Members = []helper.AccountSimpleInfo{info}
	return r.waitResCode(r.sendCmd(msg))
}

func (r *playerGatesEnemyData) OnFightBossEnd(accountID string,
	info helper.AccountSimpleInfo,
	bossIDx int, isSuccess bool) uint32 {
	msg := &gates_enemy_cmd.GatesEnemyCommandMsg{
		Type:      gates_enemy_cmd.GatesEnemyCommandMsgTypFightBossEnd,
		AccountID: accountID,
		BossIdx:   bossIDx,
		IsSuccess: isSuccess,
	}
	msg.Members = []helper.AccountSimpleInfo{info}
	if isSuccess {
		for bossIDx >= len(r.CurrBoss) {
			r.CurrBoss = append(r.CurrBoss, 0)
		}
		r.CurrBoss[bossIDx] = 1
	}
	return r.waitResCode(r.sendCmd(msg))
}

func (p *playerGatesEnemyData) GetActTime(now_t int64) (int64, int64) {
	//if now_t < p.ActEndTime {
	//	return p.ActBeginTime, p.ActEndTime
	//}
	p.ActBeginTime, p.ActEndTime = gamedata.GetGETime(now_t)
	return p.ActBeginTime, p.ActEndTime
}

func (p *playerGatesEnemyData) DebugGetActTime() {
	now_t := time.Now().Unix()
	p.ActBeginTime, p.ActEndTime = gamedata.GetGETime(now_t)
}

func (r *playerGatesEnemyData) OnDebugOp(accountID, guildID string,
	p1, p2, p3 int64) uint32 {
	msg := &gates_enemy_cmd.GatesEnemyCommandMsg{
		Type:      gates_enemy_cmd.GatesEnemyCommandMsgTypDebugOp,
		AccountID: accountID,
		GuildID:   guildID,
	}
	msg.Members = make([]helper.AccountSimpleInfo, 3)
	msg.Members[0].InfoUpdateTime = p1
	msg.Members[1].InfoUpdateTime = p2
	msg.Members[2].InfoUpdateTime = p3

	return r.waitResCode(r.sendCmd(msg))
}

func (r *playerGatesEnemyData) sendCmd(msg *gates_enemy_cmd.GatesEnemyCommandMsg) chan gates_enemy_cmd.GatesEnemyRet {
	resChan := make(chan gates_enemy_cmd.GatesEnemyRet, 1)
	msg.OkChann = resChan
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	logs.Debug("send cmd, type %d, %v", msg.Type, r.cmdChannel)
	defer cancel()

	select {
	case r.cmdChannel <- *msg:
	case <-ctx.Done():
		logs.Warn("[playerGatesEnemyData] sendCmd chann full, cmd put timeout")
		return nil
	}
	logs.Trace("[playerGatesEnemyData] sendCmd %v", *msg)
	return resChan
}

func (r *playerGatesEnemyData) waitRet(resChan chan gates_enemy_cmd.GatesEnemyRet) gates_enemy_cmd.GatesEnemyRet {
	if resChan == nil {
		return gates_enemy_cmd.GenGatesEnemyRet(gates_enemy_cmd.RESTimeOut)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var res gates_enemy_cmd.GatesEnemyRet

	select {
	case res = <-resChan:
	case <-ctx.Done():
		logs.Warn("[playerGatesEnemyData] sendCmd chann full, cmd put timeout")
		return gates_enemy_cmd.GenGatesEnemyRet(gates_enemy_cmd.RESTimeOut)
	}
	return res
}

func (r *playerGatesEnemyData) waitResCode(resChan chan gates_enemy_cmd.GatesEnemyRet) uint32 {
	ret := r.waitRet(resChan)
	return ret.Code
}
