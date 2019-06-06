package gve

import (
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/gve_proto"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 6. [Push]当前战斗状态
func (r *GVEGame) PushGameState() {
	data := gve_proto.GenStatePush(&r.Stat)
	logs.Trace("MkStatePush %d %v", len(data), data)
	r.broadcastMsg(data)
}

//SetNeedPushGameState 设置什么时机应该给客户端强推游戏更新
// 最宽泛的限时抵达
// 客户端状态切换的时机 ready -> fight（Game状态）
// 每8s, 进行一次同步,主要是boss血量和仇恨
// 有玩家进入的时候,需要通知所有人
func (r *GVEGame) SetNeedPushGameState() {
	r.isNeedPushGameState = true
}

func (r *GVEGame) CheckPushGameState() {
	if r.isNeedPushGameState {
		r.PushGameState()
		r.isNeedPushGameState = false
	}
}
