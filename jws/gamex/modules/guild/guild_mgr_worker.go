package guild

import (
	"time"

	"sync"

	"fmt"

	"math/rand"

	"golang.org/x/net/context"
	warnCode "vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
	管理所有公会的worker
	本身只处理查询公会，随机公会，公会解散回调，这几个操作，其他都会下方到各个公会worker处理
*/

type GuildMgrWorker struct {
	sid          uint
	m            *GuildModule
	guildWorks   map[string]*GuildWorker // 这里的工会是活跃工会
	waitter      util.WaitGroupWrapper
	command_chan chan guildCommand

	// for random guild
	guildSlice    []string
	guildUuid2Idx map[string]int
	guildSetMutex sync.RWMutex
}

func (gmw *GuildMgrWorker) Start(sid uint) {
	gmw.guildWorks = make(map[string]*GuildWorker, 1024)
	gmw.command_chan = make(chan guildCommand, 2048)
	gmw.sid = sid
	err := gmw.initGuildMgrWorker()
	if err != nil {
		panic(fmt.Errorf("initGuildMgrWorker err %s", err.Error()))
	}
	gmw.waitter.Wrap(func() {
		for cc := range gmw.command_chan {
			func(c guildCommand) {
				//by YZH 这个让parent never dead, 应该如此吗？
				defer logs.PanicCatcherWithInfo("GuildMgrWorker Panic")
				guid := cc.BaseInfo.GuildUUID
				switch c.Type {
				case GuildMgr_Cmd_GetRandomGuild:
					gmw.getRandomGuild(&c)
				case GuildMgr_Cmd_FindGuild:
					gmw.findGuild(&c)
				case GuildMgr_Cmd_Dismiss_CallBack:
					if gw, ok := gmw.guildWorks[guid]; ok {
						delete(gmw.guildWorks, guid)
						go gw.Stop()
					}
				default:
					if guid == "" {
						logs.Error("GuildMgrWorker cmd guid is empty  %v", c)
						c.resChan <- genWarnRes(warnCode.GuildNotFound)
						return
					}
					if err := gmw.loadGuild(guid); err != nil {
						c.resChan <- genWarnRes(warnCode.GuildNotFound)
						return
					}
					gw := gmw.guildWorks[guid]

					// send to guild chan
					ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
					defer cancel()
					select {
					case gw.command_chan <- c:
					case <-ctx.Done():
						logs.Error("[GuildMgrWorker] Start put cmd timeout, guid %s len %d",
							guid, len(gw.command_chan))
					}
				}
			}(cc)
		}
		logs.Warn("GuildWorker command_chan close!")
	})
}

func (gmw *GuildMgrWorker) Stop() {
	close(gmw.command_chan)
	gmw.waitter.Wait()
	for _, gw := range gmw.guildWorks {
		gw.Stop()
	}
}

func (gmw *GuildMgrWorker) loadGuild(guid string) error {
	if _, ok := gmw.guildWorks[guid]; !ok {
		gw := &GuildWorker{
			m: gmw.m,
		}
		g := &GuildInfo{}
		//DBLoad
		err := g.initGuildInfo(gmw.sid, guid)
		if err != nil {
			return err
		} else {
			gw.guild = g
		}
		gw.Start()
		gmw.guildWorks[guid] = gw
		gmw.addGuildUuid(guid)
	}
	// 老档base里没有leader信息，在这里补上
	gw := gmw.guildWorks[guid]
	if gw.guild.Base.LeaderAcid == "" {
		chief := gw.guild.GetGuildChief()
		if chief != nil {
			gw.guild.Base.LeaderAcid = chief.AccountID
			gw.guild.Base.LeaderName = chief.Name
		} else {
			logs.Warn("GuildMgrWorker guild.GetGuildChief nil %s", gw.guild.Base.GuildUUID)
		}
	}
	return nil
}

func (gmw *GuildMgrWorker) loadGuildSimple(guid string) *guild_info.GuildSimpleInfo {
	if g, ok := gmw.guildWorks[guid]; ok {
		return &g.guild.Base
	} else {
		err, res := loadGuildSimple(guid)
		if err != nil {
			logs.Error("GuildMgrWorker loadGuildSimple guid %s err %s",
				guid, err.Error())
			return nil
		}
		return res
	}
}

// 随机获取几个公会
func (gmw *GuildMgrWorker) getRandomGuild(c *guildCommand) {
	res := guildCommandRes{}

	uuids := gmw.randGuildUuid(MaxRandGuilds)
	res.guilds = make([]guild_info.GuildSimpleInfo, 0, len(uuids))
	for _, uuid := range uuids {
		info := gmw.loadGuildSimple(uuid)
		if info != nil {
			res.guilds = append(res.guilds, *info)
		}
	}

	c.resChan <- res
}

// 用公会id查找公会，缓存没有会从db里load
func (gmw *GuildMgrWorker) findGuild(c *guildCommand) {
	res := guildCommandRes{}

	err, uuid := findGuildUuid(c.Player1.AccountID, c.BaseInfo.GuildID)
	if err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}
	if uuid == "" {
		c.resChan <- genWarnRes(warnCode.GuildNotFound)
		return
	}
	if err := gmw.loadGuild(uuid); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	if info, ok := gmw.guildWorks[uuid]; ok {
		res.guildInfo = *info.guild
	}

	c.resChan <- res
}

func (gmw *GuildMgrWorker) initGuildMgrWorker() error {
	err, guildIds := loadAllGuildUuid(gmw.sid)
	if err != nil {
		return err
	}
	gmw.guildSlice = guildIds
	gmw.guildUuid2Idx = make(map[string]int, len(guildIds))
	for i, id := range guildIds {
		gmw.guildUuid2Idx[id] = i
	}
	return nil
}

func (gmw *GuildMgrWorker) addGuildUuid(uuid string) {
	gmw.guildSetMutex.Lock()
	defer gmw.guildSetMutex.Unlock()
	_, ok := gmw.guildUuid2Idx[uuid]
	if !ok {
		gmw.guildSlice = append(gmw.guildSlice, uuid)
		gmw.guildUuid2Idx[uuid] = len(gmw.guildSlice) - 1
	}
}

func (gmw *GuildMgrWorker) delGuildUuid(uuid string) {
	gmw.guildSetMutex.Lock()
	defer gmw.guildSetMutex.Unlock()
	idx, ok := gmw.guildUuid2Idx[uuid]
	if ok {
		tmp := gmw.guildSlice[:idx]
		tmp1 := gmw.guildSlice[idx+1:]
		gmw.guildSlice = append(tmp, tmp1...)
		delete(gmw.guildUuid2Idx, uuid)
		// 更新idx后面的uuid为新的idx
		for i := idx; i < len(gmw.guildSlice); i++ {
			gmw.guildUuid2Idx[gmw.guildSlice[i]] = i
		}
	}
}

func (gmw *GuildMgrWorker) randGuildUuid(c int) []string {
	gmw.guildSetMutex.RLock()
	defer gmw.guildSetMutex.RUnlock()

	l := len(gmw.guildSlice)
	if c >= l {
		return gmw.guildSlice[:]
	}
	r := rand.Int31n((int32(l - c + 1)))
	return gmw.guildSlice[r : int(r)+c]
}
