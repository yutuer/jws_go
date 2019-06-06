package csrob

import (
	"time"

	"fmt"

	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Guild ..
type Guild struct {
	groupID uint32
	guildID string

	info *GuildInfo

	res *resources
}

func (g *Guild) initGuild(guildID string, name string) *Guild {
	g.guildID = guildID

	logs.Trace("[CSRob] initGuild guildID [%s]", guildID)

	info, err := g.res.GuildDB.getInfo(g.guildID)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return nil
	}

	if nil == info {
		g.info = genGuildInfo(g.guildID, name)

		g.res.CommandMod.notifyRefreshGuildCacheBySelf(guildID)
		g.res.CommandMod.notifyPushGuildList(g.guildID)
	} else {
		g.info = info
	}
	g.refreshInfo(false)
	err = g.saveInfo()
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return nil
	}

	return g
}

func (g *Guild) loadGuild(guildID string) *Guild {
	logs.Trace("[CSRob] loadGuild guildID [%s]", guildID)

	info, err := g.res.GuildDB.getInfo(guildID)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return nil
	}

	if nil == info {
		return nil
	}

	g.info = info
	g.guildID = guildID
	g.refreshInfo(true)

	return g
}

func (g *Guild) saveInfo() error {
	err := g.res.GuildDB.setInfo(g.info)
	if nil != err {
		return err
	}

	return nil
}

func (g *Guild) refreshInfo(justload bool) {
	now := time.Now()
	lastUpdate := g.info.UpdateTime
	if false == gamedata.CSRobCheckSameDay(lastUpdate, now.Unix()) {
		g.info.UpdateTime = now.Unix()
		_, nat := gamedata.CSRobBattleIDAndHeroID(now.Unix())
		if err := g.res.GuildDB.removeCar(g.info.GuildID, nat, util.DailyBeginUnix(now.Unix())); nil != err {
			logs.Warn("[CSRob] Guild refreshInfo, removeCar failed, %v", err)
		}
	}
}

//GetInfo ..
func (g *Guild) GetInfo() *GuildInfo {
	g.info.GuildName = g.res.poolName.GetGuildCSName(g.info.GuildID)
	return g.info
}

//GetCarList ..
func (g *Guild) GetCarList() ([]PlayerRob, error) {
	if true == g.res.poolName.GetGuildCache(g.guildID).Dismissed {
		return []PlayerRob{}, nil
	}

	now := time.Now().Unix()
	_, nat := gamedata.CSRobBattleIDAndHeroID(now)
	infoList, err := g.res.GuildDB.getCars(g.guildID, nat)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return nil, err
	}

	robList := make([]PlayerRob, 0, len(infoList))
	for _, car := range infoList {
		if car.EndStamp < now || car.StartStamp > now {
			continue
		}

		rob, err := g.res.PlayerDB.getRob(car.Acid, car.CarID)
		if nil != err {
			logs.Error(fmt.Sprintf("%v", err))
			continue
		}

		if nil == rob {
			continue
		}

		if int(getBeRobLimit()) <= len(rob.Robbers) {
			continue
		}

		rob.Acid = car.Acid
		pc := g.res.poolName.GetPlayerCache(car.Acid)
		rob.Name = g.res.poolName.GetPlayerName(car.Acid)
		rob.GuildID = g.guildID
		rob.GuildName = g.res.poolName.GetGuildCSName(pc.GuildID)
		rob.GuildPos = pc.GuildPos
		if nil != rob.Helper {
			rob.Helper.Name = g.res.poolName.GetPlayerName(rob.Helper.Acid)
		}

		robList = append(robList, *rob)
	}

	return robList[:], nil
}

func (g *Guild) GetEnemies() []GuildEnemy {
	if true == g.res.poolName.GetGuildCache(g.guildID).Dismissed {
		return []GuildEnemy{}
	}

	list, err := g.res.GuildDB.getEnemy(g.guildID)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return []GuildEnemy{}
	}

	sort.Sort(sortEnemyList(list))

	limit := int(gamedata.CSRobShowGuildEnemiesLimit())
	if limit < len(list) {
		list = list[:limit]
	}

	retList := []GuildEnemy{}
	for _, enemy := range list {
		if "" == enemy.GuildID {
			continue
		}
		//如果仇敌已经解散了,从仇敌名单中删除它
		if true == g.res.poolName.GetGuildCache(enemy.GuildID).Dismissed {
			if err := g.res.GuildDB.removeEnemy(g.guildID, enemy.GuildID); nil != err {
				logs.Error(fmt.Sprintf("%v", err))
			}
			continue
		}

		obj := enemy
		obj.GuildName = g.res.poolName.GetGuildCSName(obj.GuildID)
		obj.BestGrade = g.res.cache.guild.getGrade(obj.GuildID)

		retList = append(retList, obj)
		if len(retList) >= limit {
			break
		}
	}
	return retList
}

//GetTeams ..
func (g *Guild) GetTeams() []GuildTeam {
	_, nat := gamedata.CSRobBattleIDAndHeroID(time.Now().Unix())
	list, err := g.res.GuildDB.getTeams(g.guildID, nat)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return []GuildTeam{}
	}
	now := time.Now()
	teamList := make([]GuildTeam, 0, len(list))
	for _, team := range list {
		obj := GuildTeam{
			Acid: team.Acid,
			Hero: team.Hero,
		}
		if autoAccept, err := g.res.PlayerDB.getPlayerStatusAutoAcceptBottom(team.Acid); nil != err {
			logs.Warn("[CSRob] Guild GetTeams [%s] getPlayerStatusAutoAcceptBottom failed, %v", team.Acid, err)
		} else {
			obj.AutoAccept = autoAccept
		}
		if vip, err := g.res.PlayerDB.getPlayerStatusVIP(team.Acid); nil != err {
			logs.Warn("[CSRob] Guild GetTeams [%s] getPlayerStatusVIP failed, %v", team.Acid, err)
			continue
		} else {
			maxAccept := getPlayerHelpLimit(vip)
			if status, err := g.res.PlayerDB.getPlayerStatus(team.Acid); nil != err {
				logs.Warn("[CSRob] Guild GetTeams [%s] getPlayerStatus failed, %v", team.Acid, err)
				continue
			} else {
				count := status.AcceptAppealCount
				if !gamedata.CSRobCheckSameDay(status.LastUpdate, now.Unix()) {
					count = 0
				}
				if count >= maxAccept {
					logs.Debug("[CSRob] Guild GetTeams [%s] filter because %d >= %d failed", count, maxAccept)
					continue
				}
			}
		}
		obj.Name = g.res.poolName.GetPlayerName(team.Acid)

		teamList = append(teamList, obj)
	}
	return teamList
}

//GetList ..
func (g *Guild) GetList() []GuildInfo {
	ids, err := g.res.GuildDB.getGuildFromList(guildRecommendNum)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return []GuildInfo{}
	}

	list := []GuildInfo{}
	for _, guildID := range ids {
		if "" == guildID {
			continue
		}
		if true == g.res.poolName.GetGuildCache(guildID).Dismissed {
			continue
		}

		list = append(list, GuildInfo{
			GuildID:   guildID,
			GuildName: g.res.poolName.GetGuildCSName(guildID),
			BestGrade: g.res.cache.guild.getGrade(guildID),
		})
	}

	return list
}

//GetRankList ..
func (g *Guild) GetRankList() []GuildRankElem {
	list, err := g.res.ranker.rangeFromRobRank(100)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return []GuildRankElem{}
	}

	for i, elem := range list {
		if "" == elem.GuildID {
			continue
		}
		gc := g.res.poolName.GetGuildCache(elem.GuildID)
		list[i].GuildName = g.res.poolName.GetGuildCSName(elem.GuildID)
		list[i].GuildMaster = gc.GuildMaster
	}

	return list
}

//GetMyRank ..
func (g *Guild) GetMyRank() GuildRankElem {
	rank, err := g.res.ranker.getRankFromRobRank(g.info.GuildID)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return GuildRankElem{}
	}

	count, robTime, err := g.res.GuildDB.getRobTimes(g.info.GuildID, g.res.ranker.batchStr)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return GuildRankElem{}
	}

	elem := GuildRankElem{
		GuildID:     g.info.GuildID,
		Rank:        rank,
		RobCount:    count,
		RobTime:     robTime,
		GuildName:   g.res.poolName.GetGuildCSName(g.info.GuildID),
		GuildMaster: g.res.poolName.GetGuildCache(g.info.GuildID).GuildMaster,
	}

	return elem
}
