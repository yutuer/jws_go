package gvg

import (
	"sync"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

type LastWorldInfo struct {
	LastBalanceTime         int64    `json:"lbt"`
	LastDayBalanceTime      int64    `json:"ldbt"`
	LastResetTime           int64    `json:"lrt"`
	LastGuildDayBalanceTime int64    `json:"lgdb"`
	LastChangAnLeader       string   `json:"lcal"`
	LastOneWorldLeader      []string `json:"lowl"`
}

type GVGWorld struct {
	LastWorldInfo
	cities  map[int]*GVGCity
	players map[string]*PlayerInfo
	guilds  map[string]*GuildInfo

	// 玩家数据缓存
	playerData map[string][GVG_AVATAR_COUNT]*helper.Avatar2Client

	sortItem    GVGSortArray
	removeGuild []guildandcity

	// debug
	DebugOffSetTime int64

	mutex sync.RWMutex
}

func (w *GVGWorld) SaveDataToDB(dbWorld *Gvg2DB) {
	dbCities := make([]GvgCity2DB, 0, len(w.cities))
	for k, v := range w.cities {
		city2db := GvgCity2DB{
			ID:              k,
			LeaderGuildID:   v.leadGuildID,
			LeaderGuildName: v.leadGuildName,
			TopNLeader:      v.topNLeader,
		}
		dbCities = append(dbCities, city2db)
	}
	players := make([]gvgPlayer2DB, 0, 1500)
	for acID, pl := range w.players {
		player := gvgPlayer2DB{}
		player.AcID = acID
		cityScore := make([]gvgPlayerScore2DB, 0, len(w.cities))
		for k, _ := range w.cities {
			s, ok := pl.cityScore[k]
			if ok {
				cityScore = append(cityScore, gvgPlayerScore2DB{
					CityID: k,
					Score:  s,
				})
			}
		}
		player.CityScore = cityScore
		player.Name = pl.name
		player.GuildID = pl.guildID
		players = append(players, player)
	}
	guilds := make([]gvgGuild2DB, 0, 20)
	for guildID, gu := range w.guilds {
		guild := gvgGuild2DB{}
		guild.GuildID = guildID
		cityScore := make([]gvgGuildScore2DB, 0, len(w.cities))
		for k, _ := range w.cities {
			s, ok := gu.cityScore[k]
			if ok {
				cityScore = append(cityScore, gvgGuildScore2DB{
					CityID: k,
					Score:  s,
				})
			}
		}
		guild.CityScore = cityScore
		guild.Name = gu.name
		guilds = append(guilds, guild)
	}
	dbWorld.LastWorldInfo = w.LastWorldInfo

	dbWorld.Players = players
	dbWorld.CityInfo = dbCities
	dbWorld.Guilds = guilds
	dbWorld.LastOneWorldLeader = w.genWorldLeader()
}

func (w *GVGWorld) LoadDataFromDB(db *Gvg2DB) {
	cities_data := gamedata.GetGVGCityID()
	w.players = make(map[string]*PlayerInfo, 1500)
	w.guilds = make(map[string]*GuildInfo, 20)
	w.cities = make(map[int]*GVGCity, len(cities_data))
	w.playerData = make(map[string][GVG_AVATAR_COUNT]*helper.Avatar2Client, 1000)

	w.sortItem = make([]*GVGSortItem, 0, 1000)
	for _, cityID := range cities_data {
		city := GVGCity{}
		city.fightPlayers = make([]*FightPlayer, 0, 30)
		w.cities[cityID] = &city
	}
	for _, item := range db.CityInfo {
		if _, ok := w.cities[item.ID]; ok {
			w.cities[item.ID].leadGuildID = item.LeaderGuildID
			w.cities[item.ID].leadGuildName = item.LeaderGuildName
			w.cities[item.ID].topNLeader = item.TopNLeader
		}
	}
	for _, item := range db.Players {
		city_score := &PlayerInfo{}
		city_score.cityScore = make(map[int]int64, len(cities_data))
		allScore := int64(0)
		for _, item := range item.CityScore {
			city_score.cityScore[item.CityID] = item.Score
			allScore += item.Score
		}
		city_score.score = allScore
		city_score.name = item.Name
		city_score.guildID = item.GuildID
		w.players[item.AcID] = city_score
	}
	for _, item := range db.Guilds {
		city_score := &GuildInfo{}
		city_score.cityScore = make(map[int]int64, len(cities_data))
		allScore := int64(0)
		for _, it := range item.CityScore {
			city_score.cityScore[it.CityID] = it.Score
			allScore += it.Score
		}
		city_score.score = allScore
		city_score.name = item.Name
		w.guilds[item.GuildID] = city_score
	}

	w.LastWorldInfo = db.LastWorldInfo
}

func (w *GVGWorld) cleanWorld() {
	w.playerData = make(map[string][GVG_AVATAR_COUNT]*helper.Avatar2Client, 1000)
	for _, city := range w.cities {
		city.fightPlayers = city.fightPlayers[:0]
	}
}

// 两个值公用一个锁

func (w *GVGWorld) getLastChangAnLeader() string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.LastChangAnLeader
}

func (w *GVGWorld) setLastChangAnLeader(acID string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.LastChangAnLeader = acID
}

func (w *GVGWorld) isWorldLeader(acID string) bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	for _, item := range w.LastOneWorldLeader {
		if item == acID {
			return true
		}
	}
	return false
}

func (w *GVGWorld) setWorldLeader(acIDs []string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.LastOneWorldLeader = acIDs
}

func (w *GVGWorld) genWorldLeader() []string {
	oneGuildID := ""
	if leadID := w.cities[CHANGAN_ID].leadGuildID; leadID != "" {
		leadNum := 1
		for _, v := range w.cities {
			if v.leadGuildID == leadID {
				leadNum++
			}
		}
		// TODO by ljz  tmp value
		if leadNum > 8 {
			oneGuildID = leadID
		}
	}

	leaders := make([]string, 0, 20)
	if oneGuildID != "" {
		for acid, player := range w.players {
			if player.guildID == oneGuildID {
				leaders = append(leaders, acid)
			}
		}
	}
	return leaders
}

type GVGCity struct {
	leadGuildID   string
	leadGuildName string
	/*-------------------------*/
	fightPlayers []*FightPlayer
	topNLeader   [gamedata.GVGTopN]string
}

func (c *GVGCity) getPlayerByID(acID string) *FightPlayer {
	for _, player := range c.fightPlayers {
		if player.acID == acID {
			return player
		}
	}
	return nil
}

func (c *GVGCity) getPlayerByGuild(guID string) []*FightPlayer {
	players := make([]*FightPlayer, 0, 10)
	for _, player := range c.fightPlayers {
		if player.guID == guID {
			players = append(players, player)
		}
	}
	return players
}

func (c *GVGCity) getPlayerByState(state int) []*FightPlayer {
	players := make([]*FightPlayer, 0, 10)
	for _, player := range c.fightPlayers {
		if player.state == state {
			players = append(players, player)
		}
	}
	return players
}

func (c *GVGCity) getNowMatchPlayer(guildID string) int {
	players := c.getPlayerByState(player_state_prepare)
	matchEnemy := 0
	for _, player := range players {
		if player.guID != guildID {
			matchEnemy += 1
		}
	}
	return matchEnemy
}

func (c *GVGCity) addPlayer(acID string, players [GVG_AVATAR_COUNT]helper.AvatarState,
	destinySkill [helper.DestinyGeneralSkillMax]int, guID string, title string,
	name string, guName string) {
	player := &FightPlayer{}
	player.acID = acID
	player.setState(player_state_idle)
	player.destinySkill = destinySkill
	player.playerInfo = players
	player.guID = guID
	player.title = title
	player.name = name
	player.guName = guName

	index := 0
	for i, fightPlayer := range c.fightPlayers {
		if fightPlayer.guID == guID {
			index = i
			break
		}
	}
	_rear := c.fightPlayers[index:]
	rear := make([]*FightPlayer, 0, len(_rear))
	rear = append(rear, _rear...)
	c.fightPlayers = append(c.fightPlayers[:index], player)
	c.fightPlayers = append(c.fightPlayers, rear[:]...)
}

func (c *GVGCity) removePlayer(acID string) {
	index := -1
	for i, player := range c.fightPlayers {
		if player.acID == acID {
			index = i
			break
		}
	}
	if index != -1 {
		c.fightPlayers = append(c.fightPlayers[:index], c.fightPlayers[index+1:]...)
	}
}

type FightPlayer struct {
	acID           string
	guID           string
	name           string
	guName         string
	title          string
	lastMatchTime  int64
	state          int
	winStreakCount int
	lastIsRobot    bool
	winRateCount   int
	/*-------------------------*/
	playerInfo   [GVG_AVATAR_COUNT]helper.AvatarState
	destinySkill [helper.DestinyGeneralSkillMax]int
}

func (fp *FightPlayer) setState(state int) {
	now_t := time.Now().Unix()
	switch state {
	case player_state_idle:
		fp.state = player_state_idle
		fp.lastMatchTime = 0
	case player_state_prepare:
		fp.state = player_state_prepare
		fp.lastMatchTime = now_t
	case player_state_fight:
		fp.state = player_state_fight
		fp.lastMatchTime = 0
	}
}

type GuildInfo struct {
	cityScore map[int]int64
	name      string
	score     int64
	guildNum  int
}

type PlayerInfo struct {
	cityScore map[int]int64
	name      string
	guildID   string
	score     int64
}
