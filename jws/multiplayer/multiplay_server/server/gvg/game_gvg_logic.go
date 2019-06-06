package gvg

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/notify"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/gvg_proto"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	GVGGameCommandNull = iota
	GVGGameCommandPlayerEnter
	GVGGameCommandPlayerLeave
)

const (
	GVGGameCommandMsgTypNull = iota
	GVGGameCommandMsgTypEnterGame
)

const (
	GVGGameManagerCommandMsgTypNull = iota
	GVGGameManagerCommandMsgCreateGame
	GVGGameManagerCommandMsgGetGame
	GVGGameManagerCommandMsgGameOver
)

type GVGGameCommandResMsg struct {
	Code  int
	Idx   int
	Datas gvg_proto.GVGGameDatas
	Stat  gvg_proto.GVGGameState
}

type LossPlayerReq struct {
	AcID string
}

type GVGGameCommandMsg struct {
	req_enterMG       *multiplayMsg.EnterMultiplayGameReq
	req_getStat       *multiplayMsg.GetGameStateReq
	req_getData       *multiplayMsg.GetGameDatasReq
	req_leave         *multiplayMsg.LeaveMultiplayGameNotify
	req_ready         *multiplayMsg.ReadyMultiplayGameNotify
	req_hp            *multiplayMsg.HPNotify
	req_getReward     *multiplayMsg.GetGameRewardsReq
	req_enemyHP       *multiplayMsg.EnemyHP
	req_ping          *multiplayMsg.Ping
	req_change_avatar *multiplayMsg.ChangeAvatar
	req_lossPlayer    *LossPlayerReq
	number            int64
	data              interface{}
	ResChann          chan GVGGameCommandResMsg
	msg               []byte
}

func (r *GVGGame) loadGameDatas(data *helper.GVGStartFightData) error {
	// 加载GameData

	r.Stat.GameScene = ""
	//cfg := gamedata.GetTBBossData(data.BossID)
	//levelCfg := gamedata.GetTBEnemyLevelData(data.Level)
	//if cfg == nil || levelCfg == nil {
	//	return fmt.Errorf("load gamedata err")
	//}
	//enemyHPCfg := gamedata.GetTBEnemyData(data.SceneID)
	//enemyHP := make([][]uint32, 0, len(enemyHPCfg.GetTroop_Table()))
	//for _, item := range enemyHPCfg.GetTroop_Table() {
	//	eCfg := gamedata.GetTBBossData(item.GetWBossID())
	//
	//	if eCfg == nil {
	//		return fmt.Errorf("load gamedata err")
	//	}
	//	sub := make([]uint32, item.GetEnemyNumber())
	//	for i := range sub {
	//		sub[i] = uint32(eCfg.GetHitPointCoefficient() * float32(levelCfg.GetHitPoint()))
	//	}
	//	enemyHP = append(enemyHP, sub)
	//}
	//bossHP := uint32(cfg.GetHitPointCoefficient() * float32(levelCfg.GetHitPoint()))
	//ATK := float32(0.0)
	//for _, item := range data.Info {
	//	if item.Attr.ATK > ATK {
	//		ATK = item.Attr.ATK
	//	}
	//}
	//bossArmor := (ATK - (cfg.GetPhysicalResistCoefficient() * float32(levelCfg.GetPhysicalResist()))) * float32(cfg.GetThresholdRatio())
	//if bossArmor > float32(cfg.GetThresholdMax()) {
	//	bossArmor = float32(cfg.GetThresholdMax())
	//} else if bossArmor < float32(cfg.GetThresholdMin()) {
	//	bossArmor = float32(cfg.GetThresholdMin())
	//}
	//logs.Debug("bossHP: %v, bossArmor: %v, enemyHP: %v", bossHP, bossArmor, enemyHP)
	//if bossArmor < 0 {
	//	bossArmor = 0
	//}

	r.Stat.InitBossHp(0, int(0))
	r.Stat.Boss[0].Hp = int(0)
	r.Stat.Boss[0].Armor = int64(0)
	r.Stat.EnemyWaveHP = nil
	return nil
}

func (r *GVGGame) loadAccountDatas(data *helper.GVGStartFightData) {
	r.Datas.AddPlayer(data)
	logs.Trace("GameStart %v", data)
}

func (r *GVGGame) loop() {
	// 检查是不是已经结束了
	nowT := time.Now().Unix()
	if r.Stat.State == gvg_proto.GameStateOver {
		if !r.isHasOnOver {
			r.OnGameOver()
			r.isHasOnOver = true
		}
		return
	} else if nowT > r.Stat.EndTime && r.Stat.EndTime != 0 { // 战斗截至事件到达
		if !r.isHasOnOver {
			r.Stat.JudgeWinner()
			logs.Info("stat over for over time")
			r.Stat.State = gvg_proto.GameStateOver
			r.SetNeedPushGameState()
		}
	}
	for i, p := range r.Stat.Player {
		if nowT-p.LastPingTime > int64(gamedata.BoxCfg.BackstageTime) && p.LastPingTime != 0 {
			r.Stat.SetPlayerLost(p.AcID)
			r.Stat.Player[i].IsBackLeave = true
		}
	}
	if nowT-r.lastPushTime > 10 { // 每 10 seconds push一下状态
		r.SetNeedPushGameState()
		r.lastPushTime = nowT
	}

	// 检查状态更新
	isStateChange := r.Stat.UpdateGameState()
	if r.Stat.State == gvg_proto.GameStateOver {
		// 通知玩家结束
		acids := make([]string, 0)
		status := helper.Invalid
		for _, p := range r.Stat.Player {
			acids = append(acids, p.AcID)
			if p.State != gvg_proto.PlayerStateKilled && p.State != gvg_proto.PlayerStateHasLeave && p.State != gvg_proto.PlayerStateLost {
				status = helper.Normal
			}
		}

		// TODO by ljz retry?
		err := notify.GetNotify().GVGGameStop(helper.GVGStopInfo{
			RoomID: r.GameID,
			Winner: r.Stat.Winner,
			AcIDs:  acids,
			Status: status,
		}, postService.GetGamexIDBySID(r.SID)) // 只有击杀时在线的玩家才能领奖
		if err != nil {
			logs.Error("Team boss fight stop err by %v from cross service", err)
		}
		//大数据埋点
		//avaId := make([]int, 0)
		//compressGS := make([]int, 0)
		//vip := make([]uint32, 0)
		//battleTime := make([]int64, 0)
		//playerHP := make([]int, 0)
		//playerState := make([]int, 0)
		//backLeave := make([]bool, 0)
		//for _, info := range r.Datas.PlayerDatas {
		//	avaId = append(avaId, info.AvatarId)
		//	compressGS = append(compressGS, info.Gs)
		//	vip = append(vip, info.VipLv)
		//}
		//for _, player := range r.Stat.Player {
		//	playerHP = append(playerHP, player.Hp)
		//	battleTime = append(battleTime, player.LastPingTime-r.Stat.StartTime)
		//	playerState = append(playerState, player.State)
		//	backLeave = append(backLeave, player.IsBackLeave)
		//}
		//ret := logiclog2.LogicInfo_TBossBattleEnd{
		//	BattleID:     r.GameID,
		//	AccounID:     r.AcIDs,
		//	AvatarID:     avaId,
		//	CompressGS:   compressGS,
		//	BossID:       r.BossID,
		//	VIP:          vip,
		//	BattleTime:   battleTime,
		//	PlayerLeftHP: playerHP,
		//	BossLeftHP:   r.Stat.Boss[0].Hp,
		//	IsWin:        r.Stat.IsSuccess,
		//	PlayerState:  playerState,
		//	IsBackLeave:  backLeave,
		//	Difficulty:   r.Level,
		//}
		//TypeInfo := logiclog2.LogicTag_TBossBattleEnd
		//format := logiclog2.BITag
		//logiclog.MultiInfo(TypeInfo, ret, format)
	}
	if isStateChange { //游戏的大状态改变
		r.SetNeedPushGameState()
	}

	// 最后检查是不是要push
	r.CheckPushGameState()
}

// 逻辑上游戏结束
func (r *GVGGame) OnGameOver() {
	logs.Info("OnGameOver %s %v", r.GameID, r.Stat)
	// log
	// 通知清除Game
	GVGGamesMgr.GVGGameOver(r.GameID)
	r.Stop()
}

func (r *GVGGame) processMsg(msg *GVGGameCommandMsg) {
	rsp := GVGGameCommandResMsg{}

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
	} else if msg.req_enemyHP != nil {
		r.onEnemyHPDeta(msg.req_enemyHP)
	} else if msg.req_ping != nil {
		r.onPing(msg.req_ping)
	} else if msg.msg != nil {
		r.broadcastMsg(msg.msg)
	} else if msg.req_change_avatar != nil {
		r.onChangeAvatar(msg.req_change_avatar)
	} else {
		logs.Error("GVGGameCommandMsg Null msg send!")
	}

	if msg.ResChann != nil {
		msg.ResChann <- rsp
	}
}

// GVGGame Grountinue 结束
func (r *GVGGame) onExitGame() {
	close(r.cmdChannel)
	r.cmdChannel = nil
}

//checkLoss 检查玩家是否掉线
func (game *GVGGame) checkLoss() {
	for _, p := range game.acID2Players {
		if p.GetSession().IsClosed() {
			game.PlayerLoss(p.GetAcID())
		}
	}
}
