package account

import (
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// 玩家上线事件，通知此玩家上线了，在afterAccountLogin之后调用的，已保证初始化完成
func (p *Account) OnAccountOnline() {
	player_msg.Send(p.AccountID.String(),
		player_msg.PlayerMsgOnLoginCode, servers.Request{})
}

// 玩家下线事件，通知此玩家离线了
func (p *Account) OnAccountOffline() {

}
