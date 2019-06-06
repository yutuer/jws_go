package csrob

import (
	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GuildMod struct {
	groupID uint32
	res     *resources
}

func (g *GuildMod) init(res *resources) {
	g.groupID = res.groupID
	g.res = res
}

func (g *GuildMod) GuildWithNew(guid string, name string) *Guild {
	guild := &Guild{}
	guild.groupID = g.groupID
	guild.res = g.res
	return guild.initGuild(guid, name)
}

//Guild
func (g *GuildMod) Guild(guid string) *Guild {
	guild := &Guild{}
	guild.groupID = g.groupID
	guild.res = g.res
	return guild.loadGuild(guid)
}

//Rename 公会改名响应
func (g *GuildMod) Rename(guid string) {
	g.res.CommandMod.notifyRefreshGuildCache(guid)
}

//MasterRename 公会军团长改名 (军团长改自己的名字并不会同步修改军团数据中的军团长名字,所以不能简单触发军团刷新)
func (g *GuildMod) MasterRename(guid string, name string) {
	g.res.CommandMod.notifyRefreshGuildMasterName(guid, name)
}

//MasterChange 公会军团长变化响应
func (g *GuildMod) MasterChange(guid string) {
	g.res.CommandMod.notifyRefreshGuildCache(guid)
}

//Dismiss 公会解散响应
func (g *GuildMod) Dismiss(guid string) {
	//大营不清除,今天的车子还需要被人打
	g.res.CommandMod.notifyDismissGuild(guid)
}

//GetPlayerTeam ..
func (g *GuildMod) GetPlayerTeam(guid, acid string) (*GuildTeam, error) {
	_, nat := gamedata.CSRobBattleIDAndHeroID(time.Now().Unix())
	return g.res.GuildDB.getTeam(guid, acid, nat)
}

//GetGuildBestGrade ..
func (g *GuildMod) GetGuildBestGrade(guid string) uint32 {
	nowstamp := time.Now().Unix()
	_, nat := gamedata.CSRobBattleIDAndHeroID(time.Now().Unix())
	infoList, err := g.res.GuildDB.getCars(guid, nat)
	if nil != err {
		logs.Warn(fmt.Sprintf("%v", err))
		return 0
	}
	bg := uint32(0)
	for _, car := range infoList {
		if car.EndStamp < nowstamp || car.StartStamp > nowstamp {
			continue
		}

		rob, err := g.res.PlayerDB.getRob(car.Acid, car.CarID)
		if nil != err {
			logs.Warn(fmt.Sprintf("[CSRob] cacheGuild guild [%s] getRob [%s:%d]%v", guid, car.Acid, car.CarID, err))
			continue
		}
		if nil == rob {
			continue
		}
		if int(getBeRobLimit()) <= len(rob.Robbers) {
			continue
		}
		if bg < rob.Info.Grade {
			bg = rob.Info.Grade
		}
	}
	return bg
}
