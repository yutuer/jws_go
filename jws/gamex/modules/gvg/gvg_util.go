package gvg

import (
	"sort"
	"time"

	"math/rand"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func convertTrueScore(value int64) int64 {
	return int64(value / GVG_SCORE_BASE)
}

func convertStoreScore(value int64) int64 {
	return int64(value)*GVG_SCORE_BASE + time.Now().Unix()
}

func updateStoreScore(value1 int64, value2 int64) int64 {
	return value2 + value1/GVG_SCORE_BASE*GVG_SCORE_BASE
}

func gvgLogiclog(world *GVGWorld) {
	var palyers int = 0
	for guild_id, guild_data := range world.guilds {
		for cityID, city := range world.cities {
			if guild_id == city.leadGuildID {
				for _, data := range world.players {
					if data.guildID == guild_id && data.score > 0 {
						palyers += 1
					}
				}
				logiclog.LogGvgGuildInfo(guild_id, guild_data.name, guild_data.guildNum, palyers, cityID, convertTrueScore(guild_data.cityScore[cityID]), "", "")
				continue
			}
			logiclog.LogGvgGuildInfo(guild_id, guild_data.name, guild_data.guildNum, palyers, 0, convertTrueScore(guild_data.cityScore[cityID]), "", "")
		}

	}

}

func gvgLogiclog2(world *GVGWorld) {
	allInfo := &logiclog.LogicInfo_GvgGuildCityScoreInfo{}
	allInfo.Infos = make([]logiclog.LogicInfo_GvgGuildCityScoreItem, 0, len(world.cities))
	for cityID, _ := range world.cities {
		cityInfo := logiclog.LogicInfo_GvgGuildCityScoreItem{}
		cityInfo.CityID = cityID

		world.sortItem = world.sortItem[:0]
		for guildID, v := range world.guilds {
			cityScore, ok := v.cityScore[cityID]
			if !ok {
				continue
			}
			item := &GVGSortItem{
				StrKey: guildID,
				IntVal: cityScore,
			}
			world.sortItem = append(world.sortItem, item)
			logs.Debug("GetGuildRank, item: %v", *item)
		}
		sort.Sort(world.sortItem)
		cityInfo.GuildInfo = make([]logiclog.LogicInfo_GvgGuildScoreItem, 0, BILOG_CITY_GUILDINFO_COUNT)
		for i := 0; i < BILOG_CITY_GUILDINFO_COUNT && i < len(world.sortItem); i++ {
			detailInfo := world.guilds[world.sortItem[i].StrKey]
			guildInfo := logiclog.LogicInfo_GvgGuildScoreItem{
				GuildID:   world.sortItem[i].StrKey,
				GuildName: detailInfo.name,
				Score:     int(convertTrueScore(world.sortItem[i].IntVal)),
			}
			cityInfo.GuildInfo = append(cityInfo.GuildInfo, guildInfo)
		}
		if len(cityInfo.GuildInfo) > 0 {
			allInfo.Infos = append(allInfo.Infos, cityInfo)
		}
	}
	logiclog.LogGvgGuildScoreInfo(allInfo, "")
}

func mergeBalance(sid uint) {
	_db := driver.GetDBConn()
	defer _db.Close()
	data := &Gvg2DB{}
	err := driver.RestoreFromHashDB(_db.RawConn(),
		TableGVGMerge(sid), data, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		logs.Error("restore from hash db err by %v", err)
		return
	}
	if err == driver.RESTORE_ERR_Profile_No_Data {
		logs.Info("No MergeData for GVG")
		return
	}
	// delete
	_, err = _db.Do("DEL", TableGVGMerge(sid))
	if err != nil {
		logs.Error("del error by %v", err)
		return
	}
	nowT := time.Now().Unix()
	timeInfo, _ := gamedata.GetHotDatas().GvgConfig.GetGVGTime(sid, nowT)
	balanceDay := gamedata.GetCommonDayDiff(data.LastDayBalanceTime, timeInfo.GVGResetTime)
	guildBalanceDay := gamedata.GetGVGGuildDayDiff(data.LastGuildDayBalanceTime, timeInfo.GVGResetTime)
	logs.Debug("guildBalanceDay: %d", guildBalanceDay)
	logs.Debug("balanceDay: %d", balanceDay)
	logs.Debug("gvg merge data: %v", data)
	// 最多发七天
	for _, city := range data.CityInfo {
		for i := int64(0); i < balanceDay && i < 7; i++ {
			if city.LeaderGuildID != "" {
				guild.GetModule(sid).GVGBalanceForPlayer(city.LeaderGuildID, city.ID)
			}
		}

		for i := int64(0); i < guildBalanceDay && i < 7; i++ {
			for j, guildID := range city.TopNLeader {
				if guildID != "" {
					guild.GetModule(sid).GVGBalanceForInventory(guildID, city.ID, j)
				}
			}
		}
	}
}

func sortFightPlayers(players []*FightPlayer) []*FightPlayer {
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			if players[i].winRateCount < players[j].winRateCount {
				players[i], players[j] = players[j], players[i]
			}
		}
	}
	return players
}

func randFightPlayers(players []*FightPlayer) []*FightPlayer {
	playerlen := len(players)
	if playerlen == 1 {
		return players
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var pos int
	for i := 0; i < playerlen; i++ {
		pos = r.Intn(playerlen - 1)
		players[i], players[pos] = players[pos], players[i]
	}
	return players
}

func removeFightPlayers(index1, index2 int, players []*FightPlayer) []*FightPlayer {
	if index1 == index2 {
		return players
	}
	return append(append(players[:index1], players[index1+1:index2]...), players[index2+1:]...)
}

// TODO by ljz implementation
func notifyMultiplayStart(acidA string, acidB string, heroA []int, heroB []int, sid uint) (string, string, error) {
	info := helper.GVGStartFightData{
		Acid1:   acidA,
		Acid2:   acidB,
		Avatar1: heroA,
		Avatar2: heroB,
		Sid:     sid,
	}
	data, err := GetNotify(helper.GVGToken).GVGNotify(info)
	if err != nil {
		return "", "", err
	}
	return genGVGMultiplayInfo(data)
}

type battleRes helper.GVGStopInfo

func genGVGMultiplayInfo(data []byte) (string, string, error) {
	info := &helper.GVGStartFigntRetData{}
	err := json.Unmarshal(data, info)
	return info.WebsktUrl, info.RoomID, err

}
