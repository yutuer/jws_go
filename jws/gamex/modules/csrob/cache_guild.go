package csrob

import (
	"fmt"
	"sync"

	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type cacheGuild struct {
	root      *utilCache
	cache     *cacheGuildCache
	cacheLock sync.RWMutex

	holdIDs map[string]bool
}

func newCacheGuild(r *utilCache) *cacheGuild {
	return &cacheGuild{
		root: r,
		cache: &cacheGuildCache{
			m: map[string]*cacheGuildElem{},
		},
		holdIDs: map[string]bool{},
	}
}

func (c *cacheGuild) refresh(now time.Time) {
	ids, err := c.root.res.GuildDB.getGuildFromList(guildRecommendNum)
	if nil != err {
		logs.Error(fmt.Sprintf("%v", err))
		return
	}

	for _, id := range ids {
		c.holdIDs[id] = true
	}

	nowstamp := now.Unix()
	_, nat := gamedata.CSRobBattleIDAndHeroID(nowstamp)

	cg := &cacheGuildCache{
		m: map[string]*cacheGuildElem{},
	}
	for id := range c.holdIDs {
		cl, err := c.root.res.GuildDB.getCars(id, nat)
		if nil != err {
			logs.Warn(fmt.Sprintf("[CSRob] cacheGuild guild [%s] getCars%v", id, err))
			continue
		}
		bg := uint32(0)
		for _, car := range cl {
			if car.EndStamp < nowstamp || car.StartStamp > nowstamp {
				continue
			}

			rob, err := c.root.res.PlayerDB.getRob(car.Acid, car.CarID)
			if nil != err {
				logs.Warn(fmt.Sprintf("[CSRob] cacheGuild guild [%s] getRob [%s:%d]%v", id, car.Acid, car.CarID, err))
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
		cg.m[id] = &cacheGuildElem{
			GuildID:   id,
			BestGrade: bg,
		}
	}

	// logs.Debug("[CSRob] cacheGuild refresh")

	c.cacheLock.Lock()
	c.cache = cg
	c.cacheLock.Unlock()
}

func (c *cacheGuild) getGrade(guildID string) uint32 {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()
	g, exist := c.cache.m[guildID]
	if false == exist || nil == g {
		c.holdIDs[guildID] = true
		return 0
	}
	return g.BestGrade
}

func (c *cacheGuild) getMultiGrade(guildIDs []string) map[string]uint32 {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()

	gs := map[string]uint32{}
	for _, id := range guildIDs {
		g, exist := c.cache.m[id]
		if false == exist || nil == g {
			c.holdIDs[id] = true
			gs[id] = 0
		} else {
			gs[id] = g.BestGrade
		}
	}
	return gs
}

type cacheGuildCache struct {
	m map[string]*cacheGuildElem
}

type cacheGuildElem struct {
	GuildID   string
	BestGrade uint32
}
