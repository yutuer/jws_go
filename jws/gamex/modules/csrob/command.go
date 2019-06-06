package csrob

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//type of command for mux
const (
	commandNone = iota
	commandReward
	commandPushGuildList
	commandRefreshGuildRecommend

	commandRefreshPlayerCache
	commandRefreshGuildCache
	commandDismissGuild
	commandPlayerLeaveGuild

	commandRewardGuildWeek
)

type command struct {
	Type int

	commandParam
}

type commandParam struct {
	Acid      string
	CarID     uint32
	CacheInit bool

	bGuildID bool
	GuildID  string

	bName bool
	Name  string

	bGuildPos bool
	GuildPos  int

	Time int64
}

type commandMod struct {
	waiter util.WaitGroupWrapper

	commandChan chan *command

	res *resources
}

func newCommandMod(res *resources) *commandMod {
	return &commandMod{
		commandChan: make(chan *command, 1024),
		res:         res,
	}
}

func (cm *commandMod) start() {
	// external reg
	logs.Debug("[CSRob] commandMod Start")

	// go routine for command queue
	cm.waiter.Wrap(func() {
		for cc := range cm.commandChan {
			logs.Debug("[CSRob] commandMod command [%v]", cc)

			func() {
				defer logs.PanicCatcherWithInfo("[CSRob] commandMod command process panic")
				cm.dispatch(cc)
			}()
		}
		logs.Warn("[CSRob] commandMod commandChan close")
	})
}

func (cm *commandMod) stop() {
	logs.Debug("[CSRob] commandMod Stop")

	close(cm.commandChan)
	cm.waiter.Wait()
}

func (cm *commandMod) commandExecAsync(cmd *command) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case cm.commandChan <- cmd:
	case <-ctx.Done():
		logs.Error(fmt.Sprintf("[CSRob] commandExecAsync channel full, cmd put timeout[%s]", util.ASyncCmdTimeOut.String()))
	}
}

func (cm *commandMod) dispatch(cmd *command) {
	switch cmd.Type {
	case commandReward:
		cm.cmdReward(cmd)
	case commandPushGuildList:
		cm.cmdPushGuildList(cmd)
	case commandRefreshGuildRecommend:
		cm.cmdGuildRecommendRefresh()
	case commandRefreshPlayerCache:
		cm.cmdRefreshPlayerCache(cmd)
	case commandRefreshGuildCache:
		cm.cmdRefreshGuildCache(cmd)
	case commandDismissGuild:
		cm.cmdDismissGuild(cmd)
	case commandPlayerLeaveGuild:
		cm.cmdPlayerLeaveGuild(cmd)
	case commandRewardGuildWeek:
		cm.cmdRewardGuildWeek(cmd)
	default:
		logs.Warn("[CSRob] Unknown command type [%d]", cmd.Type)
	}
}

func makeCommand(cmdType int, param commandParam) *command {
	return &command{
		Type:         cmdType,
		commandParam: param,
	}
}

func (cm *commandMod) notifyReward(acid string, car uint32) {
	logs.Debug("[CSRob] notifyReward")
	cmd := makeCommand(commandReward, commandParam{Acid: acid, CarID: car})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) cmdReward(cmd *command) {
	player := cm.res.PlayerMod.Player(cmd.Acid)
	if nil == player {
		logs.Error(fmt.Sprintf("[CSRob] cmdReward send reward to [%s], but he isn't exist", cmd.Acid))
		return
	}

	ret, data, goods, err := player.DoneDrivingCar(cmd.CarID)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return
	}

	if RetInvalid == ret {
		logs.Error(fmt.Sprintf("[CSRob] send reward to [%s] for car [%d] failed, ret [%v]", cmd.Acid, cmd.CarID, ret))
		return
	} else if RetOK != ret {
		logs.Warn("[CSRob] send reward to [%s] for car [%d] failed, ret [%v]", cmd.Acid, cmd.CarID, ret)
		return
	}

	logs.Info("[CSRob] send reward to [%s] for car [%d], reward: {%v}", cmd.Acid, cmd.CarID, goods)

	if nil != data.Helper && 0 == len(data.Robbers) {
		helperPlayer := cm.res.PlayerMod.Player(data.Helper.Acid)
		if nil == player {
			logs.Error(fmt.Sprintf("[CSRob] cmdReward send reward to Helper [%s], but he isn't exist", data.Helper.Acid))
			return
		}
		helperPlayer.DoneHelpCar(cmd.Acid, data)
		logs.Info("[CSRob] send reward to Helper [%s] for car [%d], reward: {%v}", cmd.Acid, cmd.CarID, goods)
	}
}

func (cm *commandMod) notifyPushGuildList(guildID string) {
	if "" == guildID {
		return
	}
	logs.Debug("[CSRob] notifyPushGuildList")
	cmd := makeCommand(commandPushGuildList, commandParam{GuildID: guildID})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) cmdPushGuildList(cmd *command) {
	info, ret := guild.GetModule(cm.res.sid).GetGuildInfo(cmd.GuildID)
	if true == ret.HasError() {
		logs.Error(fmt.Sprintf("[CSRob] cmdPushGuildList GetGuildInfo failed, %v", ret.ErrMsg))
		return
	}

	err := cm.res.GuildDB.pushGuildToList(cmd.GuildID, info.Base.GuildGSSum)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return
	}
}

func (cm *commandMod) notifyGuildRecommendRefresh() {
	logs.Debug("[CSRob] notifyGuildRecommendRefresh")
	cmd := makeCommand(commandRefreshGuildRecommend, commandParam{})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) cmdGuildRecommendRefresh() {
	nextTime := util.DailyBeginUnix(time.Now().Unix()) + gamedata.CSRobGuildListRefreshOffset() + util.DaySec
	cm.res.ticker.regCommand(
		func() { cm.notifyGuildRecommendRefresh() },
		nextTime,
	)

	status, err := cm.res.GuildDB.getCommonStatus()
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return
	}
	cm.res.status = status

	now := time.Now().Unix()
	if gamedata.CSRobCheckSameDay(status.RecommendRefreshTime, now) {
		logs.Warn("[CSRob] cmdGuildRecommendRefresh again in the same day")
		return
	}
	logs.Debug("[CSRob] cmdGuildRecommendRefresh do...")

	list, err := cm.res.GuildDB.loadAllGuildIDs()
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return
	}

	for _, guildID := range list {
		sid, err := guild_info.GetShardIdByGuild(guildID)
		if nil != err {
			logs.Error(fmt.Sprintf("[CSRob] cmdGuildRecommendRefresh GetShardIdByGuild failed %v", err))
			continue
		}
		if sid != cm.res.sid {
			continue
		}

		info, ret := guild.GetModule(cm.res.sid).GetGuildInfo(guildID)
		if true == ret.HasError() {
			logs.Error(fmt.Sprintf("[CSRob] cmdGuildRecommendRefresh GetGuildInfo [%s] failed, %v", guildID, ret.ErrMsg))
			continue
		}

		err = cm.res.GuildDB.pushGuildToList(guildID, info.Base.GuildGSSum)
		if nil != err {
			logs.Error(fmt.Sprintf("%v", err))
			return
		}
	}

	status.RecommendRefreshTime = now
	if err := cm.res.GuildDB.setCommonStatus(status); nil != err {
		logs.Warn("[CSRob] cmdGuildRecommendRefresh setCommonStatus failed")
	}
	cm.res.status = status
}

func (cm *commandMod) NotifyPlayerRename(acid, guildID, name string) {
	logs.Debug("[CSRob] NotifyPlayerRename")
	cmd := makeCommand(
		commandRefreshPlayerCache,
		commandParam{
			Acid:      acid,
			GuildID:   guildID,
			bGuildID:  true,
			Name:      name,
			bName:     true,
			CacheInit: false,
		})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) NotifyPlayerGuildPos(acid, guildID string, pos int) {
	logs.Debug("[CSRob] NotifyPlayerGuildPos")
	cmd := makeCommand(
		commandRefreshPlayerCache,
		commandParam{
			Acid:      acid,
			GuildID:   guildID,
			bGuildID:  true,
			GuildPos:  pos,
			bGuildPos: true,
			CacheInit: false,
		})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) NotifyPlayerGuildJoin(acid, guildID string) {
	logs.Debug("[CSRob] NotifyPlayerGuildJoin")
	cmd := makeCommand(
		commandRefreshPlayerCache,
		commandParam{
			Acid:      acid,
			GuildID:   guildID,
			bGuildID:  true,
			CacheInit: false,
		})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) NotifyRefreshPlayerCache(acid, guildID string) {
	logs.Debug("[CSRob] NotifyRefreshPlayerCache")
	cmd := makeCommand(
		commandRefreshPlayerCache,
		commandParam{
			Acid:      acid,
			GuildID:   guildID,
			bGuildID:  true,
			CacheInit: false,
		})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) notifyRefreshPlayerCacheBySelf(acid, guildID, name string, pos int) {
	logs.Debug("[CSRob] notifyRefreshPlayerCacheBySelf")
	cmd := makeCommand(
		commandRefreshPlayerCache,
		commandParam{
			Acid:      acid,
			GuildID:   guildID,
			bGuildID:  true,
			Name:      name,
			bName:     true,
			GuildPos:  pos,
			bGuildPos: true,
			CacheInit: true,
		})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) cmdRefreshPlayerCache(cmd *command) {
	logs.Debug("[CSRob] cmdRefreshPlayerCache, cmd [%v]", cmd)
	cache := cm.res.poolName.GetPlayerCache(cmd.Acid)
	if "" == cache.Acid && false == cmd.CacheInit {
		return
	}
	logs.Debug("[CSRob] cmdRefreshPlayerCache, GetPlayerCache cache [%v]", cache)

	cache.Acid = cmd.Acid
	cache.Sid = cm.res.sid

	if true == cmd.bName {
		cache.Name = cmd.Name
	}
	if true == cmd.bGuildPos {
		cache.GuildPos = cmd.GuildPos
	}
	if true == cmd.bGuildID {
		cache.GuildID = cmd.GuildID
	}

	logs.Debug("[CSRob] cmdRefreshPlayerCache SetPlayerCache, cache [%v]", cache)
	cm.res.poolName.SetPlayerCache(cmd.Acid, cache)
}

func (cm *commandMod) notifyRefreshGuildCacheBySelf(guildID string) {
	logs.Debug("[CSRob] NotifyRefreshGuildCache")
	cmd := makeCommand(commandRefreshGuildCache, commandParam{GuildID: guildID, CacheInit: true})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) notifyRefreshGuildCache(guildID string) {
	logs.Debug("[CSRob] NotifyRefreshGuildCache")
	cmd := makeCommand(commandRefreshGuildCache, commandParam{GuildID: guildID, CacheInit: false})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) notifyRefreshGuildMasterName(guildID string, name string) {
	logs.Debug("[CSRob] NotifyRefreshGuildCache")
	cmd := makeCommand(commandRefreshGuildCache, commandParam{GuildID: guildID, bName: true, Name: name, CacheInit: false})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) cmdRefreshGuildCache(cmd *command) {
	cache := cm.res.poolName.GetGuildCache(cmd.GuildID)
	if "" == cache.GuildID && false == cmd.CacheInit {
		return
	}

	cache.GuildID = cmd.GuildID
	cache.Sid = cm.res.sid

	info, ret := guild.GetModule(cm.res.sid).GetGuildInfo(cmd.GuildID)
	if true == ret.HasError() {
		logs.Error(fmt.Sprintf("[CSRob] cmdRefreshGuildCache GetGuildInfo failed, %v", ret.ErrMsg))
		return
	}
	cache.GuildName = info.Base.Name
	cache.GuildMaster = info.Base.LeaderName

	if true == cmd.bName {
		cache.GuildMaster = cmd.Name
	}

	cm.res.poolName.SetGuildCache(cmd.GuildID, cache)
}

func (cm *commandMod) notifyDismissGuild(guildID string) {
	logs.Debug("[CSRob] notifyDismissGuild")
	cmd := makeCommand(commandDismissGuild, commandParam{GuildID: guildID})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) cmdDismissGuild(cmd *command) {
	cache := cm.res.poolName.GetGuildCache(cmd.GuildID)
	if "" == cache.GuildID {
		logs.Warn("[CSRob] cmdDismissGuild, want dismiss guild [%s], but guild is not exist", cmd.GuildID)
		return
	}

	//在缓存数据中设置公会已解散
	logs.Debug("[CSRob] cmdDismissGuild, set dismiss tag to guild cache")
	cache.Dismissed = true
	cm.res.poolName.SetGuildCache(cmd.GuildID, cache)

	//从军团推荐列表中删除
	logs.Debug("[CSRob] cmdDismissGuild, remove from guild recommend list")
	if err := cm.res.GuildDB.removeGuildFromList(cmd.GuildID); nil != err {
		logs.Error(fmt.Sprintf("%v", err))
	}

	//从军团榜单里面删除
	logs.Debug("[CSRob] cmdDismissGuild, remove from guild rank list")
	if err := cm.res.ranker.removeFromRobRank(cmd.GuildID); nil != err {
		logs.Error(fmt.Sprintf("%v", err))
	}
}

func (cm *commandMod) notifyPlayerLeaveGuild(acid string) {
	logs.Debug("[CSRob] notifyPlayerLeaveGuild")
	cmd := makeCommand(commandPlayerLeaveGuild, commandParam{Acid: acid})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) cmdPlayerLeaveGuild(cmd *command) {
	cache := cm.res.poolName.GetPlayerCache(cmd.Acid)
	if "" == cache.Acid && false == cmd.CacheInit {
		return
	}

	logs.Debug("[CSRob] cmdPlayerLeaveGuild, player [%s] leave guild [%s]", cmd.Acid, cache.GuildID)

	logs.Debug("[CSRob] cmdPlayerLeaveGuild, remove player [%s] from guild [%s] teamlist", cmd.Acid, cache.GuildID)
	natlist := gamedata.CSRobBattleIDList()
	for _, nat := range natlist {
		if err := cm.res.GuildDB.removeTeam(cache.GuildID, cmd.Acid, nat); nil != err {
			logs.Error(fmt.Sprintf("%v", err))
		}
	}

	cache.GuildID = ""
	logs.Debug("[CSRob] cmdPlayerLeaveGuild, player [%s] set cache [%v]", cmd.Acid, cache)
	cm.res.poolName.SetPlayerCache(cmd.Acid, cache)
}

func (cm *commandMod) notifyRewardGuildWeek(t int64) {
	logs.Debug("[CSRob] notifyRewardGuildWeek")
	cmd := makeCommand(commandRewardGuildWeek, commandParam{Time: t})
	cm.commandExecAsync(cmd)
}

func (cm *commandMod) cmdRewardGuildWeek(cmd *command) {
	logs.Debug("[CSRob] cmdRewardGuildWeek")
	logs.Debug("[CSRob] cmdRewardGuildWeek, reward to week [%d]", cmd.Time)

	//先判断本周是不是发过一次了
	status, err := cm.res.GuildDB.getCommonStatus()
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return
	}
	cm.res.status = status
	if game.Cfg.IsRunModeProd() {
		if true == gamedata.CSRobCheckSameWeek(cmd.Time, status.WeekRewardTime) {
			logs.Warn("[CSRob] cmdRewardGuildWeek, already send reward this week at [%d]", status.WeekRewardTime)
			return
		}
	} else {
		if true == gamedata.CSRobCheckSameDayAndHour(cmd.Time, status.WeekRewardTime) {
			logs.Warn("[CSRob] cmdRewardGuildWeek, already send reward this day and hour at [%d]", status.WeekRewardTime)
			return
		}
	}
	//记下已发奖记录先
	status.WeekRewardTime = cmd.Time
	if err := cm.res.GuildDB.setCommonStatus(status); nil != err {
		logs.Error(fmt.Sprintf("[CSRob] cmdRewardGuildWeek setCommonStatus failed, %v", err))
		return
	}
	cm.res.status = status

	//清除奖励名单先
	oldSid := cm.res.ranker.mWeekRewardShardID
	oldMembers := cm.res.ranker.getWeekRewardList()
	cm.res.ranker.clearWeekReward()

	//删掉老名单的玩家称号
	if oldSid == cm.res.sid {
		for _, member := range oldMembers {
			player_msg.Send(member, player_msg.PlayerMsgTitleCode,
				player_msg.DefaultMsg{})
		}
	}

	//取冠军
	list, err := cm.res.ranker.rangeFromRobRank(1)
	if nil != err {
		logs.Error(fmt.Sprintf("[CSRob] cmdRewardGuildWeek rangeFromRobRank failed, %v", err))
		return
	}
	if 0 == len(list) {
		logs.Warn("[CSRob] cmdRewardGuildWeek, range list is empty")
		return
	}

	//冠军应该只有一个
	champion := list[0]

	//解析冠军军团所属的服
	sid, err := guild_info.GetShardIdByGuild(champion.GuildID)
	if nil != err {
		logs.Error(fmt.Sprintf("[CSRob] cmdRewardGuildWeek GetShardIdByGuild failed, %v", err))
		return
	}
	sid = game.Cfg.GetShardIdByMerge(sid)

	now := time.Now().Unix()

	//记下冠军所属的服
	cm.res.ranker.mWeekRewardShardID = sid
	cm.res.ranker.mWeekRewardTime = now

	//检查冠军是不是本服的, 不是的话就没别的要处理了
	if sid != cm.res.sid {
		logs.Warn("[CSRob] cmdRewardGuildWeek, the champion [%s] is not in this shard", champion.GuildID)
		return
	}
	logs.Warn("[CSRob] cmdRewardGuildWeek, the champion is [%s]", champion.GuildID)

	//取公会名单
	guildInfo, ret := guild.GetModule(sid).GetGuildInfo(champion.GuildID)
	if true == ret.HasError() {
		logs.Error(fmt.Sprintf("[CSRob] cmdRewardGuildWeek GetGuildInfo [%s] failed, %v", champion.GuildID, ret.ErrMsg))
		return
	}
	members := guildInfo.GetAllMemberAcids()
	logs.Warn("[CSRob] cmdRewardGuildWeek, send reward to acid [%s]", members)

	//记录奖励名单, 并写入DB
	cm.res.ranker.addWeekReward(members)
	reward := &RewardWeek{
		Time:    now,
		Sid:     sid,
		Count:   champion.RobCount,
		GuildID: champion.GuildID,
		Members: members,
	}
	if err := cm.res.RewardDB.setRewardWeek(reward); nil != err {
		logs.Error(fmt.Sprintf("[CSRob] cmdRewardGuildWeek setRewardWeek failed, %v", err))
	}

	//向玩家弹称号通知
	for _, member := range members {
		player_msg.Send(member, player_msg.PlayerMsgTitleCode,
			player_msg.DefaultMsg{})
	}
}
