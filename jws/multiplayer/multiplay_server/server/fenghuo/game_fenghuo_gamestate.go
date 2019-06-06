package fenghuo

import (
	"time"

	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/fenghuomsg"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/fenghuoproto"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *FenghuoGame) hpSyncable() bool {
	if r.GameState > FenghuoGameStatusWaitingInit && r.GameState < FenghuoGameStatusGameOver {
		return true
	}
	return false
}

func (r *FenghuoGame) HpSyncable() bool {
	r.channelMutex.RLock()
	defer r.channelMutex.RUnlock()

	return r.hpSyncable()
}

func (r *FenghuoGame) checkLoss() {
	for _, p := range r.acID2Players {
		if p.Player.GetSession().IsClosed() {
			r.PlayerOnline[p.IDX] = false
		}
	}
}

func (r *FenghuoGame) loop() {
	//// 检查状态更新
	r.UpdateGameState()
}

func (r *FenghuoGame) GetPlayerHPs() [fenghuoMaxPlayer]int {
	r.channelMutex.RLock()
	defer r.channelMutex.RUnlock()
	return r.PlayerHPs
}

func (r *FenghuoGame) SetPlayerHP(idx, HP int) {
	if idx < 0 || idx > fenghuoMaxPlayer-1 {
		return
	}
	r.channelMutex.Lock()
	defer r.channelMutex.Unlock()

	r.PlayerHPs[idx] = HP
}

func (r *FenghuoGame) GetGameStatus() int {
	r.channelMutex.RLock()
	defer r.channelMutex.RUnlock()
	return r.GameState
}

func (r *FenghuoGame) GetEnemyHPs() []int {
	r.channelMutex.RLock()
	defer r.channelMutex.RUnlock()
	ehps := make([]int, r.NumEnemies)
	for i := 0; i < r.NumEnemies; i++ {
		ehps[i] = r.EnemyHps[i]
	}
	return ehps
}

func (r *FenghuoGame) AllHPNotify2Client() {
	if r.GameState == FenghuoGameStatusSubLevelStart {
		builder := proto.GetNewBuilder()
		ehp := fenghuoproto.GenIntArray(builder, AllHPNotifyStartEnemiesHpVector, r.EnemyHps[:])
		hp := fenghuoproto.GenIntArray(builder, AllHPNotifyStartHpsVector, r.PlayerHPs[:])

		AllHPNotifyStart(builder)
		AllHPNotifyAddSublevel(builder, int32(r.SubLevel-1))
		AllHPNotifyAddHps(builder, hp)
		AllHPNotifyAddEnemiesHp(builder, ehp)
		sfn := AllHPNotifyEnd(builder)
		b := fenghuoproto.GenPacketRspBasic(builder,
			int32(msgprocessor.MsgTypNotify),
			0, 0,
			DatasAllHPNotify, sfn)

		r.Channel.Broadcast(b)
		logs.Debug("FenghuoGame OnHPNotify, Enemies:%v", r.EnemyHps)
		logs.Debug("FenghuoGame OnHPNotify, Players:%v", r.PlayerHPs)
	}
}

func (r *FenghuoGame) StartFightNotify2Client() {
	builder := proto.GetNewBuilder()
	ehp := fenghuoproto.GenIntArray(builder, StartFightNotifyStartEnemiesHpVector, r.EnemyHps[:])
	hp := fenghuoproto.GenIntArray(builder, StartFightNotifyStartHpsVector, r.PlayerHPs[:])

	StartFightNotifyStart(builder)
	StartFightNotifyAddSublevel(builder, int32(r.SubLevel-1))
	StartFightNotifyAddEnemiesHp(builder, ehp)
	StartFightNotifyAddHps(builder, hp)

	sfn := StartFightNotifyEnd(builder)
	b := fenghuoproto.GenPacketRspBasic(builder,
		int32(msgprocessor.MsgTypNotify),
		0, 0,
		DatasStartFightNotify, sfn)
	r.broadcastMsg(b)
}

func (r *FenghuoGame) UpdateGameState() bool {
	r.channelMutex.RLock()
	defer r.channelMutex.RUnlock()
	switch r.GameState {
	case FenghuoGameStatusWaitingInit:

		if r.SubLevel-1 >= 0 {
			allWanna := true
			for _, s := range r.SubLevelStatus[r.SubLevel-1] {
				allWanna = allWanna && (s == FenghuoGameStatusSubLevelStart)
			}
			if allWanna {
				//所有人都准备好了
				r.SubLevelDoneCount++
				r.GameState = FenghuoGameStatusSubLevelStart
				r.StartFightNotify2Client()
			}
		}

	case FenghuoGameStatusSubLevelStart:
		r.AllHPNotify2Client()
	case FenghuoGameStatusSubLevelDone:
	case FenghuoGameStatusGameOver:
	}
	return r.GameState != r.LastGameState
}

func (r *FenghuoGame) GenFenghuoAvatarsFlatbuffer(builder *flatbuffers.Builder) [fenghuoMaxPlayer]flatbuffers.UOffsetT {
	r.channelMutex.RLock()
	defer r.channelMutex.RUnlock()
	var AcVector [fenghuoMaxPlayer]flatbuffers.UOffsetT
	for i := range r.Avatars {
		AcVector[i] = fenghuoproto.GenAccountInfoData(builder, i, r.Avatars[i])
	}
	return AcVector
}

func (r *FenghuoGame) OnStartFightNotify(idx int, req *StartFightNotify) {
	if idx < 0 || idx > fenghuoMaxPlayer-1 {
		logs.Warn("FenghuoGame.OnStartFightNotify idx wrong %d", idx)
		return
	}
	if r.GameState >= FenghuoGameStatusGameOver {
		//8轮战斗已经结束 什么都不做
		logs.Debug("FenghuoGame.OnStartFightNotify Game Over.")
		return
	}
	logs.Trace("OnStartFightNotify ok")
	r.channelMutex.Lock()
	defer r.channelMutex.Unlock()

	sublvl := int(req.Sublevel() + 1)
	r.PlayerHPs[idx] = int(req.Hps(idx))

	ne := req.EnemiesHpLength()
	r.NumEnemies = ne
	r.EnemyHps = r.enemyHps[:ne]
	for i := 0; i < ne; i++ {
		r.EnemyHps[i] = int(req.EnemiesHp(i))
	}

	r.SubLevelStatus[r.SubLevel-1][idx] = FenghuoGameStatusSubLevelStart

	if r.SubLevel != sublvl {
		//第一次有人想玩新馆的时候,这里才会启动。后续发出相同关卡的情况不会触发这里
		//有任何玩家已经想开始r.SubLevel - 1关的游戏了
		go func(sublvl int) {
			<-time.After(5 * time.Second)

			r.channelMutex.Lock()
			defer r.channelMutex.Unlock()
			//强制伪装其他人允许进行游戏了,无论对方是因为掉线还是其他什么原因
			for i := range r.SubLevelStatus[sublvl] {
				r.SubLevelStatus[sublvl][i] = FenghuoGameStatusSubLevelStart
			}
			logs.Info("Fake StartFightNotify2Client %v", r.PlayerOnline)
		}(r.SubLevel - 1)
	}

	r.SubLevel = sublvl
}

func (r *FenghuoGame) isPlayerDead(idx int) bool {
	if idx < 0 || idx > fenghuoMaxPlayer-1 {
		return true
	}
	return r.PlayerHPs[idx] == 0
}

func (r *FenghuoGame) OnHPNotify(idx int, req *HPNotify) (dead bool) {
	if idx < 0 || idx > fenghuoMaxPlayer-1 {
		return
	}
	r.channelMutex.Lock()
	defer r.channelMutex.Unlock()

	if !r.hpSyncable() {
		logs.Debug("FenghuoGame.OnHPNotify hpSyncable false.")
		return r.isPlayerDead(idx)
	}

	r.PlayerHPs[idx] += int(req.MyHpD())
	if r.PlayerHPs[idx] < 0 {
		r.PlayerHPs[idx] = 0
	}

	dead = r.isPlayerDead(idx)

	num := req.EnemiesHpDLength()
	if num != r.NumEnemies || num != len(r.EnemyHps) {
		logs.Error("Client has diffent num of enmeies. client:%d, server:%d", num, r.NumEnemies)
		return
	}

	enemyclear := true
	for i := 0; i < r.NumEnemies; i++ {
		r.EnemyHps[i] += int(req.EnemiesHpD(i))
		if r.EnemyHps[i] < 0 {
			r.EnemyHps[i] = 0
		}
		enemyclear = enemyclear && (r.EnemyHps[i] == 0)
	}
	if enemyclear {
		if r.SubLevelDoneCount == fenghuoMaxSublevles {
			r.GameState = FenghuoGameStatusGameOver
		} else {
			r.GameState = FenghuoGameStatusSubLevelDone
		}
	}
	return
}
