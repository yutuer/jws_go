package csrob

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	guild_notify "vcs.taiyouxi.net/jws/gamex/modules/guild/notifycsrob"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type CSRobModule struct {
	sid     uint
	groupID uint32

	PlayerMod    *PlayerMod
	GuildMod     *GuildMod
	Ranker       *rankGuild
	PlayerRanker *rankPlayer
	res          *resources
}

type resources struct {
	sid     uint
	groupID uint32

	PlayerMod *PlayerMod
	GuildMod  *GuildMod

	PlayerDB   *PlayerDB
	GuildDB    *GuildDB
	RewardDB   *RewardDB
	NamePoolDB *NamePoolDB

	CommandMod *commandMod
	ticker     *ticker
	ranker     *rankGuild

	playerRanker *rankPlayer
	PlayerRankDB *PlayerRankDB

	poolName *poolName
	cache    *utilCache

	status *GuildCommonStatus
}

func genCSRobModule(sid uint) *CSRobModule {
	rm := &CSRobModule{}

	mergeShardID := game.Cfg.GetShardIdByMerge(sid)
	rm.sid = mergeShardID
	rm.groupID = gamedata.GetCSRobGroupId(uint32(mergeShardID))

	res := &resources{}
	res.status = &GuildCommonStatus{}
	res.sid = rm.sid
	res.groupID = rm.groupID

	res.ticker = newTicker()
	res.ticker.res = res

	res.PlayerDB = initPlayerDB(res)
	res.GuildDB = initGuildDB(res)
	res.RewardDB = initRewardDB(res)

	res.CommandMod = newCommandMod(res)

	res.PlayerMod = &PlayerMod{}
	res.PlayerMod.init(res)
	res.GuildMod = &GuildMod{}
	res.GuildMod.init(res)

	res.poolName = newPoolName(res)
	res.NamePoolDB = initNamePoolDB(res)
	res.ranker = newRankGuild(res)

	res.playerRanker = newRankPlayer(res)
	res.PlayerRankDB = initPlayerRankDB(res)

	res.cache = newUtilCache(res)

	rm.res = res
	rm.PlayerMod = res.PlayerMod
	rm.GuildMod = res.GuildMod
	rm.Ranker = res.ranker
	rm.PlayerRanker = res.playerRanker
	logs.Debug("[CSRob] CSRobModule genCSRobModule")

	return rm
}

func (rm *CSRobModule) Start() {
	rm.res.ticker.Start()
	rm.res.ranker.Start()
	rm.res.CommandMod.start()
	logs.Debug("[CSRob] CSRobModule Start")
}

func (rm *CSRobModule) AfterStart(g *gin.Engine) {
	rm.res.PlayerMod.testDBLink()
	rm.res.CommandMod.notifyGuildRecommendRefresh()
	rm.res.ticker.reloadRewardList()
	rm.res.ranker.loadWeekReward()

	guild_notify.RegRefreshCallback(rm.res.CommandMod.notifyRefreshGuildCache)
}

func (rm *CSRobModule) BeforeStop() {

}

func (rm *CSRobModule) Stop() {
	rm.res.CommandMod.stop()
	rm.res.ranker.Stop()
	rm.res.ticker.Stop()
	logs.Debug("[CSRob] CSRobModule Stop")
}
