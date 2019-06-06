package teamboss

import (
	"time"

	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/teamboss_proto"
)

// 6. [Push]当前战斗状态
func (r *TBGame) PushGameState() {
	// select lead
	if r.lead != "" {
		leadPlayer := r.Stat.Player[r.Stat.GetPlayerIdx(r.lead)]
		if (time.Now().Unix()-leadPlayer.LastHPDeltaTime > 2 && leadPlayer.State == teamboss_proto.PlayerStateFighting) || leadPlayer.IsExit() {
			r.SetLeadExcept(r.lead)
		}
	}

	data := teamboss_proto.GenStatePush(&r.Stat, r.lead)
	r.broadcastMsg(data)
}

//SetNeedPushGameState 设置什么时机应该给客户端强推游戏更新
// 最宽泛的限时抵达
// 客户端状态切换的时机 ready -> fight（Game状态）
// 每8s, 进行一次同步,主要是boss血量和仇恨
// 有玩家进入的时候,需要通知所有人
func (r *TBGame) SetNeedPushGameState() {
	r.isNeedPushGameState = true
}

func (r *TBGame) CheckPushGameState() {
	if r.isNeedPushGameState {
		r.PushGameState()
		r.isNeedPushGameState = false
	}
}

// 6. [Push]当前战斗状态
func (r *TBGame) PushEnemyHP() {

}
