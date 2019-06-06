package gvg

import (
	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/errorcode"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GVGCmd struct {
	Typ          int
	AcID         string
	GuID         string
	Name         string
	GuName       string
	Title        string
	Players      [GVG_AVATAR_COUNT]helper.AvatarState
	DestinySkill [helper.DestinyGeneralSkillMax]int
	CityID       int
	IsWin        bool
	retChan      chan GVGRet
	GuMember     int
	MutiplayRet  *battleRes
	RoomID       string

	DetailInfo [GVG_AVATAR_COUNT]*helper.Avatar2Client
}

type GVGRet struct {
	ErrCode        errorcode.ErrorCode
	SortItem       GVGSortArray
	Score          int
	WinStreakCount int // 连胜次数
	MatchCount     int
	Rank           int
	NowMatch       int
	StrVal         string
}

func (m *gvgModule) prepareFight(cmd *GVGCmd) {
	logs.Debug("Receive prepare fight cmd")
	w := m.world
	city, ok := w.cities[cmd.CityID]
	if !ok {
		cmd.retChan <- GVGRet{ErrCode: err_city}
		return
	}
	logs.Debug("City player count: %d", len(city.fightPlayers))
	player := city.getPlayerByID(cmd.AcID)
	if player == nil {
		logs.Warn("The Player(%s) has already leave city(%d)", cmd.AcID, cmd.CityID)
		cmd.retChan <- GVGRet{}
		return
	}
	if player.state != player_state_idle {
		logs.Warn("GVG player state error, now state is %d", player.state)
	}
	player.setState(player_state_prepare)
	cmd.retChan <- GVGRet{}
}

func (m *gvgModule) cancelMatch(cmd *GVGCmd) {
	logs.Debug("Receive cancel match cmd")
	w := m.world
	city, ok := w.cities[cmd.CityID]
	if !ok {
		cmd.retChan <- GVGRet{ErrCode: err_city}
		return
	}
	logs.Debug("City player count: %d", len(city.fightPlayers))
	player := city.getPlayerByID(cmd.AcID)
	logs.Debug("prepareFight player: %v", player)
	if player == nil {
		logs.Warn("The Player(%s) has already leave city(%d)", cmd.AcID, cmd.CityID)
		cmd.retChan <- GVGRet{}
		return
	}
	if player.state != player_state_prepare {
		logs.Warn("GVG player state error, now state is %d", player.state)
	}
	player.setState(player_state_idle)
	cmd.retChan <- GVGRet{}
}

func (m *gvgModule) enterCity(cmd *GVGCmd) {
	logs.Debug("Receive enter city cmd")
	w := m.world
	w.playerData[cmd.AcID] = cmd.DetailInfo
	city, ok := w.cities[cmd.CityID]
	if !ok {
		cmd.retChan <- GVGRet{ErrCode: err_city}
		return
	}
	logs.Debug("City player count: %d", len(city.fightPlayers))
	player := city.getPlayerByID(cmd.AcID)
	if player != nil {
		logs.Warn("GVG player state error, now state is %d", player.state)
		player.setState(player_state_idle)
		cmd.retChan <- GVGRet{}
		return
	}
	city.addPlayer(cmd.AcID, cmd.Players, cmd.DestinySkill, cmd.GuID, cmd.Title, cmd.Name, cmd.GuName)
	cmd.retChan <- GVGRet{}
}

func (m *gvgModule) leaveCity(cmd *GVGCmd) {
	logs.Debug("Receive leave city cmd")
	w := m.world
	city, ok := w.cities[cmd.CityID]
	if !ok {
		cmd.retChan <- GVGRet{ErrCode: err_city}
		return
	}
	logs.Debug("City Player: %d", len(city.fightPlayers))
	player := city.getPlayerByID(cmd.AcID)
	if player == nil {
		logs.Warn("The Player(%s) has already leave city(%d)", cmd.AcID, cmd.CityID)
		cmd.retChan <- GVGRet{}
		return
	}
	city.removePlayer(cmd.AcID)
	cmd.retChan <- GVGRet{}
}

func (m *gvgModule) mutiplayEnd(cmd *GVGCmd) {
	logs.Debug("Receive mutiplayEnd city cmd")
	m.battleRes[cmd.MutiplayRet.RoomID] = cmd.MutiplayRet
	cmd.retChan <- GVGRet{}
}

func (m *gvgModule) endFight(cmd *GVGCmd) {
	logs.Debug("Receive end fight cmd")
	// 只有分数更新需要做时间限制,保证分数准确,超时不更新分数
	now_t := m.GetNowTime()
	timeInfo, _ := gamedata.GetHotDatas().GvgConfig.GetGVGTime(m.sid, now_t)
	if !m.isActive() || now_t > timeInfo.GVGEndTime || now_t < timeInfo.GVGOpeningTime {
		logs.Warn("The Activity has already over")
		cmd.retChan <- GVGRet{}
		return
	}
	cmd.IsWin = false
	// TODO by ljz more strict check
	if res, ok := m.battleRes[cmd.RoomID]; ok {
		if ok {
			cmd.IsWin = res.Winner == cmd.AcID
		} else {
			logs.Warn("No battle info for player: %v, room: %v", cmd.AcID, cmd.RoomID)
		}
	}

	w := m.world
	city, ok := w.cities[cmd.CityID]
	if !ok {
		cmd.retChan <- GVGRet{ErrCode: err_city}
		return
	}
	logs.Debug("City Player: %d", len(city.fightPlayers))
	player := city.getPlayerByID(cmd.AcID)
	if player == nil {
		logs.Warn("The Player(%s) has already leave city(%d)", cmd.AcID, cmd.CityID)
		city.addPlayer(cmd.AcID, cmd.Players, cmd.DestinySkill, cmd.GuID, cmd.Title, cmd.Name, cmd.GuName)
		player = city.getPlayerByID(cmd.AcID)
		player.state = player_state_fight
	}
	if player.state != player_state_fight {
		logs.Warn("GVG player state error, now state is %d", player.state)
	}
	// update hp mp ws destinyskill

	player.playerInfo = cmd.Players
	// 保证数据一致id 和 detail对应
	detail, ok := w.playerData[cmd.AcID]
	if ok {
		for _, state := range cmd.Players {
			had := false
			for _, info := range detail {
				if info.AvatarId == state.Avatar {
					had = true
					break
				}
			}
			if had == false {
				// 数据不一致，理论上不可能发生
				logs.Error("fatal error gvg, ID and detail doesn't match")
				delete(w.playerData, cmd.AcID)
				break
			}
		}
	}

	player.destinySkill = cmd.DestinySkill
	player.setState(player_state_idle)

	score := int64(0)
	config := gamedata.GetGVGConfig()
	if cmd.IsWin {
		if player.lastIsRobot {
			score = int64(config.GetRobotGVGPoint())
		} else {
			player.winStreakCount++
			score = int64(config.GetBasicGVGPoint()) + int64(gamedata.GetGVGWinScore(player.winStreakCount))
		}
		player.winRateCount++
	} else {
		score = int64(config.GetFailGVGPoint())
		player.winStreakCount = 0
		player.winRateCount--
	}
	// 计算时间戳与分数

	score = convertStoreScore(score)
	logs.Debug("gvg get score: %d", score)
	// 更新个人分数
	player_info, ok := w.players[cmd.AcID]
	if ok {
		cityScore, ok := player_info.cityScore[cmd.CityID]
		if !ok {
			player_info.cityScore[cmd.CityID] = score
		} else {
			player_info.cityScore[cmd.CityID] = updateStoreScore(cityScore, score)
		}
		player_info.score = updateStoreScore(player_info.score, score)
		w.players[cmd.AcID] = player_info
	} else {
		playerInfo := &PlayerInfo{}
		playerInfo.score = score
		playerInfo.name = player.name
		playerInfo.cityScore = make(map[int]int64, GVG_CITY_COUNT)
		playerInfo.cityScore[cmd.CityID] = score
		playerInfo.guildID = player.guID
		w.players[cmd.AcID] = playerInfo
	}

	// 更新所在工会分数
	guild_info, ok := w.guilds[cmd.GuID]
	if ok {
		cityScore, ok := guild_info.cityScore[cmd.CityID]
		if !ok {
			guild_info.cityScore[cmd.CityID] = score
		} else {
			guild_info.cityScore[cmd.CityID] = updateStoreScore(cityScore, score)
		}
		guild_info.score = updateStoreScore(guild_info.score, score)
		w.guilds[cmd.GuID] = guild_info
		w.guilds[cmd.GuID].guildNum = cmd.GuMember
	} else {
		guildInfo := &GuildInfo{}
		guildInfo.score = score
		guildInfo.name = player.guName
		guildInfo.cityScore = make(map[int]int64, GVG_CITY_COUNT)
		guildInfo.cityScore[cmd.CityID] = score
		w.guilds[cmd.GuID] = guildInfo
		w.guilds[cmd.GuID].guildNum = cmd.GuMember
	}
	// 将玩家score更新到公会信息中

	cmd.retChan <- GVGRet{
		Score:          int(convertTrueScore(score)),
		WinStreakCount: player.winStreakCount,
	}

}

func (m *gvgModule) getGuildRank(cmd *GVGCmd) {
	logs.Debug("Receive get guild rank cmd")
	w := m.world
	ret := GVGRet{}
	ret.SortItem = make([]*GVGSortItem, 0, 20)
	for _, v := range w.guilds {
		cityScore, ok := v.cityScore[cmd.CityID]
		if !ok {
			continue
		}
		item := &GVGSortItem{
			StrKey: v.name,
			IntVal: cityScore,
		}
		ret.SortItem = append(ret.SortItem, item)
		logs.Debug("GetGuildRank, item: %v", *item)
	}
	sort.Sort(ret.SortItem)
	city, ok := w.cities[cmd.CityID]
	if !ok {
		cmd.retChan <- GVGRet{ErrCode: err_city}
		return
	}
	rank := 0
	score := 0
	for i := 0; i < len(ret.SortItem); i++ {
		ret.SortItem[i].IntVal = convertTrueScore(ret.SortItem[i].IntVal)
		if ret.SortItem[i].StrKey == cmd.GuName {
			rank = i + 1
			score = int(ret.SortItem[i].IntVal)
		}
	}
	ret.Rank = rank
	ret.Score = score
	ret.NowMatch = city.getNowMatchPlayer(cmd.GuID)
	cmd.retChan <- ret
}

func (m *gvgModule) getSelfGuildInfo(cmd *GVGCmd) {
	logs.Debug("Receive self guild info cmd")
	w := m.world

	ret := GVGRet{}
	ret.SortItem = make([]*GVGSortItem, 0, 30)
	// TODO by ljz not good
	for _, v := range w.players {
		if v.guildID != cmd.GuID {
			continue
		}

		score, ok := v.cityScore[cmd.CityID]
		if !ok {
			continue
		}
		item := &GVGSortItem{
			StrKey: v.name,
			IntVal: score,
		}
		ret.SortItem = append(ret.SortItem, item)
		logs.Debug("GetSelfGuildInfo, item: %v", *item)
	}
	sort.Sort(ret.SortItem)
	for i := 0; i < len(ret.SortItem); i++ {
		ret.SortItem[i].IntVal = convertTrueScore(ret.SortItem[i].IntVal)
	}
	city, ok := w.cities[cmd.CityID]
	if !ok {
		cmd.retChan <- GVGRet{ErrCode: err_city}
		return
	}

	ret.NowMatch = city.getNowMatchPlayer(cmd.GuID)
	cmd.retChan <- ret
}

func (m *gvgModule) getSelfGuildAllInfo(cmd *GVGCmd) {
	logs.Debug("Receive self guild all info cmd")
	w := m.world
	ret := GVGRet{}
	ret.SortItem = make([]*GVGSortItem, 0, 30)
	for _, v := range w.players {
		if v.guildID != cmd.GuID {
			continue
		}

		item := &GVGSortItem{
			StrKey: v.name,
			IntVal: v.score,
		}
		ret.SortItem = append(ret.SortItem, item)
		logs.Debug("GetSelfGuildAllInfo, item: %v", *item)
	}
	sort.Sort(ret.SortItem)
	for i := 0; i < len(ret.SortItem); i++ {
		ret.SortItem[i].IntVal = convertTrueScore(ret.SortItem[i].IntVal)
	}
	cmd.retChan <- ret
}

func (m *gvgModule) getPlayerInfo(cmd *GVGCmd) {
	logs.Debug("Receive player info cmd")
	w := m.world
	city, ok := w.cities[cmd.CityID]
	if !ok {
		cmd.retChan <- GVGRet{ErrCode: err_city}
		return
	}
	ret := GVGRet{}
	ret.SortItem = make([]*GVGSortItem, 0, 20)
	for _, item := range city.fightPlayers {
		if item.guID == cmd.GuID {
			continue
		}
		item := &GVGSortItem{
			StrKey: item.name,
			StrVal: item.guName,
		}
		ret.SortItem = append(ret.SortItem, item)
		logs.Debug("GetPlayerInfo, item: %v", *item)
	}
	ret.NowMatch = city.getNowMatchPlayer(cmd.GuID)
	cmd.retChan <- ret
}

func (m *gvgModule) getCityLeader(cmd *GVGCmd) {
	logs.Debug("Receive get city leader cmd")
	cities := m.world.cities
	ret := GVGRet{}
	ret.SortItem = make([]*GVGSortItem, 0, len(cities))
	for k, city := range cities {
		item := &GVGSortItem{
			IntKey: k,
			StrVal: city.leadGuildName,
		}
		ret.SortItem = append(ret.SortItem, item)
		logs.Debug("GetCityLeader, item: %v", *item)
	}
	cmd.retChan <- ret
}

func (m *gvgModule) getGuildWorldRank(cmd *GVGCmd) {
	logs.Debug("Receive get guild world rank cmd")
	w := m.world
	ret := GVGRet{}
	ret.SortItem = make([]*GVGSortItem, 0, 20)

	for _, info := range w.guilds {
		item := &GVGSortItem{
			StrKey: info.name,
			IntVal: info.score,
		}
		ret.SortItem = append(ret.SortItem, item)
	}
	rank := 0
	score := 0
	sort.Sort(ret.SortItem)
	for i := 0; i < len(ret.SortItem); i++ {
		ret.SortItem[i].IntVal = convertTrueScore(ret.SortItem[i].IntVal)
		if cmd.GuName == ret.SortItem[i].StrKey {
			rank = rank + i + 1
			score = int(ret.SortItem[i].IntVal)
		}
	}
	ret.Rank = rank
	ret.Score = score
	cmd.retChan <- ret
}

func (m *gvgModule) getPlayerWorldInfo(cmd *GVGCmd) {
	logs.Debug("Receive get player world info cmd")
	w := m.world
	ret := GVGRet{}
	w.sortItem = w.sortItem[:0]
	ret.SortItem = make([]*GVGSortItem, 0, WORLD_PLAYER_INFO_COUNT)
	for _, info := range w.players {
		item := &GVGSortItem{
			StrKey: info.name,
			IntVal: info.score,
		}
		w.sortItem = append(w.sortItem, item)
	}
	sort.Sort(w.sortItem)
	for i := 0; i < len(w.sortItem) && i < WORLD_PLAYER_INFO_COUNT; i++ {
		ret.SortItem = append(ret.SortItem, &GVGSortItem{
			StrKey: w.sortItem[i].StrKey,
			IntVal: convertTrueScore(w.sortItem[i].IntVal),
		})
	}
	rank := 0
	score := 0
	for i, item := range w.sortItem {
		if cmd.Name == item.StrKey {
			rank = rank + i + 1
			score = int(convertTrueScore(item.IntVal))
			break
		}
	}
	ret.Rank = rank
	ret.Score = score
	cmd.retChan <- ret
}

func (m *gvgModule) removePlayer(cmd *GVGCmd) {
	logs.Debug("gvg cmd: remove player")
	ret := GVGRet{}
	delete(m.world.players, cmd.AcID)
	cmd.retChan <- ret
}

func (m *gvgModule) removeGuild(cmd *GVGCmd) {
	logs.Debug("gvg cmd: remove guild")
	ret := GVGRet{}
	delete(m.world.guilds, cmd.GuID)
	for _, city := range m.world.cities {
		if city.leadGuildID == cmd.GuID {
			city.leadGuildID = ""
			city.leadGuildName = ""
		}
	}
	cmd.retChan <- ret
}

func (m *gvgModule) renameGuild(cmd *GVGCmd) {
	logs.Debug("gvg cmd: rename guild")
	ret := GVGRet{}
	guild, ok := m.world.guilds[cmd.GuID]
	if ok {
		guild.name = cmd.GuName
	}
	for _, city := range m.world.cities {
		if city.leadGuildID == cmd.GuID {
			city.leadGuildName = cmd.GuName
		}
	}
	cmd.retChan <- ret
}

func (m *gvgModule) renamePlayer(cmd *GVGCmd) {
	logs.Debug("gvg cmd: rename player")
	ret := GVGRet{}
	player, ok := m.world.players[cmd.AcID]
	if ok {
		player.name = cmd.Name
	}
	cmd.retChan <- ret
}
