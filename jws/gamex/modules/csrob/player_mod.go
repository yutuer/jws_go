package csrob

import (
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//PlayerMod 玩家数据的统合
type PlayerMod struct {
	groupID uint32

	res *resources
}

func (p *PlayerMod) init(res *resources) {
	p.groupID = res.groupID
	p.res = res
}

//PlayerWithNew 去玩家句柄,没有则创建
func (p *PlayerMod) PlayerWithNew(param *PlayerParam) *Player {
	player := &Player{}
	player.groupID = p.groupID
	player.res = p.res
	return player.initPlayer(param)
}

//Player 去玩家句柄,不创建
func (p *PlayerMod) Player(acid string) *Player {
	player := &Player{}
	player.groupID = p.groupID
	player.res = p.res
	return player.loadPlayer(acid)
}

//RefreshPlayerCacheBySelf 刷新玩家完整的缓存
func (p *PlayerMod) RefreshPlayerCacheBySelf(acid, guildID, name string, pos int) {
	p.res.CommandMod.notifyRefreshPlayerCacheBySelf(acid, guildID, name, pos)
}

//Rename 玩家改名
func (p *PlayerMod) Rename(acid string, guid string, name string) {
	p.res.CommandMod.NotifyPlayerRename(acid, guid, name)
}

//ChangeGuildPos 玩家改公会职位
func (p *PlayerMod) ChangeGuildPos(acid string, guid string, pos int) {
	p.res.CommandMod.NotifyPlayerGuildPos(acid, guid, pos)
}

//LeaveGuild 玩家离开公会(被踢)
func (p *PlayerMod) LeaveGuild(acid string) {
	p.res.CommandMod.notifyPlayerLeaveGuild(acid)
}

//JoinGuild 玩家加入公会
func (p *PlayerMod) JoinGuild(acid string, guid string) {
	p.res.CommandMod.NotifyPlayerGuildJoin(acid, guid)
}

func (p *PlayerMod) testDBLink() {
	if err := p.res.PlayerDB.testLink(); nil != err {
		logs.Error("[CSRob] DBLink Test Failed")
	} else {
		logs.Info("[CSRob] DBLink Test OK")
	}
}
