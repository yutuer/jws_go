package gve

import (
	"math/rand"

	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/gve_notify/post_data"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/logiclog"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/notify"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/gve_proto"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	GVEGameCommandNull = iota
	GVEGameCommandPlayerEnter
	GVEGameCommandPlayerLeave
)

const (
	GVEGameCommandMsgTypNull = iota
	GVEGameCommandMsgTypEnterGame
)

const (
	GVEGameManagerCommandMsgTypNull = iota
	GVEGameManagerCommandMsgCreateGame
	GVEGameManagerCommandMsgGetGame
	GVEGameManagerCommandMsgGameOver
)

type GVEGameCommandResMsg struct {
	Code  int
	Idx   int
	Datas gve_proto.GVEGameDatas
	Stat  gve_proto.GVEGameState
}

type LossPlayerReq struct {
	AcID string
}

type GVEGameCommandMsg struct {
	req_enterMG    *multiplayMsg.EnterMultiplayGameReq
	req_getStat    *multiplayMsg.GetGameStateReq
	req_getData    *multiplayMsg.GetGameDatasReq
	req_leave      *multiplayMsg.LeaveMultiplayGameNotify
	req_ready      *multiplayMsg.ReadyMultiplayGameNotify
	req_hp         *multiplayMsg.HPNotify
	req_getReward  *multiplayMsg.GetGameRewardsReq
	req_lossPlayer *LossPlayerReq
	number         int64
	ResChann       chan GVEGameCommandResMsg
}

func (r *GVEGame) loadGameDatas() {
	// 加载GameData
	cLv := r.getGameAttrCorpLv()
	datas := gamedata.GetGVEGameData(cLv, r.Stat.IsHard, rand.New(&r.Stat.Rng))
	r.Datas.AppendBoss(datas.Boss, datas.BossModel)
	r.Stat.InitBossHp(0, int(float32(datas.BossModel.GetHitPoint())*datas.Boss.GetHitPointCoefficient()))
	r.Stat.GameScene = datas.LevelID
}

func (r *GVEGame) loadAccountDatas() error {
	// 加载Account信息
	logs.Trace("loadAccountDatas %v", r.AcIDs)
	acCount := len(r.AcIDs)
	// TODO by ljz not good
	// 只有自己一个人,或许队友是机器人
	if len(r.AcIDs) == 1 {
		acID := r.AcIDs[0]
		data, err := notify.GetNotify().GameStart(acID, r.GameID, acCount, postService.GetGamexIDByAcID(acID))
		if err != nil {
			return err
		}
		playerData := r.Datas.AppendAccount(acID, data)
		logs.Trace("GameStart %v", data)
		r.Stat.InitPlayerHp(acID, int(playerData.HP))
		if data.RobotData != nil {
			for _, robot := range data.RobotData {
				if robot == nil {
					continue
				}
				robot_data := &post_data.StartGVEPostResData{
					Data:     *robot,
					Reward:   []string{},
					Count:    []uint32{},
					IsUseHc:  false,
					IsDouble: false,
				}
				robotData := r.Datas.AppendAccount(robot_data.Data.AcID, robot_data)
				logs.Trace("GameStart %v", data)
				r.Stat.InitPlayerHp(robot_data.Data.AcID, int(robotData.HP))
			}
		}
	} else {
		for _, acID := range r.AcIDs {

			data, err := notify.GetNotify().GameStart(acID, r.GameID, acCount, postService.GetGamexIDByAcID(acID))
			if err != nil {
				return err
			}
			playerData := r.Datas.AppendAccount(acID, data)
			logs.Trace("GameStart %v", data)
			r.Stat.InitPlayerHp(acID, int(playerData.HP))
		}
	}
	return nil
}

// 根据玩家等级算出计算Boss属性用的"平均"等级
func (r *GVEGame) getGameAttrCorpLv() uint32 {
	var lvSum uint32
	for i := 0; i < len(r.Datas.PlayerDatas); i++ {
		lvSum += r.Datas.PlayerDatas[i].Data.CorpLv
	}
	return lvSum / uint32(len(r.Datas.PlayerDatas))
}

func (r *GVEGame) loop() {
	// 检查是不是已经结束了
	nowT := time.Now().Unix()
	if r.Stat.State == gve_proto.GameStateOver {
		if !r.isHasOnOver {
			r.OnGameOver()
			r.isHasOnOver = true
		}
		return
	} else if nowT > r.Stat.EndTime { // 战斗截至事件到达
		if !r.isHasOnOver {
			r.Stat.IsSuccess = false
			r.Stat.State = gve_proto.GameStateOver
			r.SetNeedPushGameState()
		}
	}

	if nowT-r.lastPushTime > 8 { // 每 8 seconds push一下状态
		r.SetNeedPushGameState()
		r.lastPushTime = nowT
	}

	// 检查状态更新
	isStateChange := r.Stat.UpdateGameState()
	if isStateChange { //游戏的大状态改变
		r.SetNeedPushGameState()
	}

	// 最后检查是不是要push
	r.CheckPushGameState()
}

// 逻辑上游戏结束
func (r *GVEGame) OnGameOver() {
	logs.Info("OnGameOver %s %v", r.GameID, r.Stat)
	// log
	r.log()
	// 通知玩家结束
	for idx, acID := range r.AcIDs {
		if r.Stat.Player[idx].State == gve_proto.PlayerStateOver ||
			r.Stat.Player[idx].State == gve_proto.PlayerStateKilled {
			postService.GetGamexIDByAcID(acID)
			notify.GetNotify().GameStop(helper.GameStopInfo{
				AcIDs:  acID,
				GameID: r.GameID,
				IsHasReward: r.Stat.Player[idx].State == gve_proto.PlayerStateOver ||
					r.Stat.Player[idx].State == gve_proto.PlayerStateKilled,
				IsSuccess: r.Stat.IsSuccess,
			}, postService.GetGamexIDByAcID(acID)) // 只有击杀时在线的玩家才能领奖
		}
	}
	// 通知清除Game
	GVEGamesMgr.GVEGameOver(r.GameID)
	r.Stop()
}

func (r *GVEGame) processMsg(msg *GVEGameCommandMsg) {
	rsp := GVEGameCommandResMsg{}

	if msg.req_enterMG != nil {
		rsp.Code = r.enterRoom(msg.number, msg.req_enterMG)
		rsp.Datas = r.Datas
		rsp.Stat = r.Stat
	} else if msg.req_getData != nil {
		//rsp.Code = r.getGameDatas(msg.number, msg.req_getData)
		rsp.Datas = r.Datas
	} else if msg.req_getStat != nil {
		//rsp.Code = r.getGameState(msg.number, msg.req_getStat)
		rsp.Stat = r.Stat
	} else if msg.req_hp != nil {
		r.onHpDeta(msg.req_hp)
	} else if msg.req_leave != nil {
		r.onLeaveRoom(msg.req_leave)
	} else if msg.req_ready != nil {
		r.onReadyToGame(msg.req_ready)
	} else if msg.req_getReward != nil {
		rsp.Datas = r.Datas
		rsp.Idx = r.Stat.GetPlayerIdx(string(msg.req_getReward.AccountId()))
	} else if msg.req_lossPlayer != nil {
		r.Stat.SetPlayerLost(msg.req_lossPlayer.AcID)
	} else {
		logs.Error("GVEGameCommandMsg Null msg send!")
	}

	if msg.ResChann != nil {
		msg.ResChann <- rsp
	}
}

// GVEGame Grountinue 结束
func (r *GVEGame) onExitGame() {
	close(r.cmdChannel)
	r.cmdChannel = nil
}

func (r *GVEGame) log() {
	duration := time.Now().Unix() - r.Stat.StartTime
	bossIds := make([]string, 0, len(r.Datas.BossAcDatas))
	for _, boss := range r.Datas.BossAcDatas {
		bossIds = append(bossIds, boss.GetBossID())
	}
	var lvlsum uint32
	for _, p := range r.Datas.PlayerDatas {
		lvlsum += p.Data.CorpLv
	}
	lvlavg := lvlsum / uint32(len(r.Datas.PlayerDatas))
	logiclog.GameOver(r.GameID, duration, r.Stat.IsSuccess, r.Stat.IsHard, bossIds, lvlavg)

	var player1, player2, player3 logiclog.LogicInfo_Player
	if len(r.Datas.PlayerDatas) > 0 && len(r.Stat.Player) > 0 {
		fillLogicInfoPlayer(&player1, &r.Stat.Player[0], &r.Datas.PlayerDatas[0])
	}
	if len(r.Datas.PlayerDatas) > 1 && len(r.Stat.Player) > 1 {
		fillLogicInfoPlayer(&player2, &r.Stat.Player[1], &r.Datas.PlayerDatas[1])
	}
	if len(r.Datas.PlayerDatas) > 2 && len(r.Stat.Player) > 2 {
		fillLogicInfoPlayer(&player3, &r.Stat.Player[2], &r.Datas.PlayerDatas[2])
	}
	bossHp := make([]int, 0, len(r.Stat.Boss))
	bossOrgHp := make([]int, 0, len(r.Stat.Boss))
	bossHpRate := make([]float32, 0, len(r.Stat.Boss))
	for i, boss := range r.Stat.Boss {
		bossHp = append(bossHp, boss.Hp)
		f_bossOrgHp := float32(r.Datas.BossModel[i].GetHitPoint()) * r.Datas.BossAcDatas[i].GetHitPointCoefficient()
		_bossOrgHp := int(f_bossOrgHp)
		bossOrgHp = append(bossOrgHp, _bossOrgHp)
		bossHpRate = append(bossHpRate, float32(boss.Hp)/f_bossOrgHp)
	}
	logiclog.GameOverWithPlayer(r.GameID, duration, r.Stat.IsSuccess, r.Stat.IsHard,
		bossIds, bossHp, bossOrgHp, bossHpRate, lvlavg,
		player1, player2, player3)
}

func fillLogicInfoPlayer(lp *logiclog.LogicInfo_Player, ps *gve_proto.GVEPlayerState, p *post_data.StartGVEPostResData) {
	lp.PlayerId = p.Data.AcID
	lp.Avatar = p.Data.AvatarId
	lp.Gs = p.Data.Gs
	lp.Hp = ps.Hp
	lp.HpOrg = int(p.Data.HP)
	lp.HpRate = float32(ps.Hp) / float32(p.Data.HP)
	lp.Lvl = p.Data.CorpLv
	lp.IsUseHc = p.IsUseHc
	lp.IsDouble = p.IsDouble
}

//checkLoss 检查玩家是否掉线
func (game *GVEGame) checkLoss() {
	for _, p := range game.acID2Players {
		if p.GetSession().IsClosed() {
			game.PlayerLoss(p.GetAcID())
		}
	}
}
