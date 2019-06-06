package gvg

import (
	"fmt"
	"time"

	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/jws/multiplayer/util/post_service_on_etcd"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func (m *gvgModule) matchFight() {
	defer logs.PanicCatcherWithInfo("gvg match fight fatal error")

	w := m.world
	now_t := time.Now().Unix()
	maxMatchTime := int64(gamedata.GetGVGConfig().GetMatchLimitTime())
	for _, city := range w.cities {
		players := city.getPlayerByState(player_state_prepare)
		l := len(players)
		players = randFightPlayers(players)
		for l >= 2 {
			player1 := players[0]
			index := -1
			for i, p2 := range players {
				if p2.guID != player1.guID {
					index = i
					break
				}
			}
			if index == -1 {
				break
			}
			player2 := players[index]
			logs.Debug("Match Scuccess(true player), city: %v, player1: %v, player2: %v", city, *player1, *player2)
			heroA := make([]int, 0)
			for _, item := range player1.playerInfo {
				heroA = append(heroA, item.Avatar)
			}
			heroB := make([]int, 0)
			for _, item := range player2.playerInfo {
				heroB = append(heroB, item.Avatar)
			}
			url, roomID, err := notifyMultiplayStart(player1.acID, player2.acID, heroA, heroB, m.sid)
			if err != nil {
				logs.Error("notify multiplay err by %v", err)
				// TODO by ljz how to do
			}
			logs.Debug("Get multiplay info: %v, %v", url, roomID)
			m.notifyPlayerFight(player1, player2, url, roomID)
			player1.lastIsRobot = false
			player2.lastIsRobot = false
			player1.setState(player_state_fight)
			player2.setState(player_state_fight)
			players = removeFightPlayers(0, index, players)
			l -= 2
		}
		for _, player := range players {
			if now_t-player.lastMatchTime > maxMatchTime {
				m.notifyPlayerRobotFight(player)
				player.lastIsRobot = true
				player.setState(player_state_fight)
				logs.Debug("Match Scuccess(robot player), city: %v, player: %v", city, *player)
			}
		}
	}
}

func (m *gvgModule) notifyPlayerFight(player1, player2 *FightPlayer, url string, room_id string) {
	enemyData1 := m.world.playerData[player1.acID]
	enemyData2 := m.world.playerData[player2.acID]
	go func() {
		player_msg.Send(player1.acID, player_msg.PlayerMsgGVGStartCode,
			player_msg.PlayerMsgGVGStart{
				PlayerInfo:      player1.playerInfo,
				DestinySkill:    player1.destinySkill,
				EnemyAcID:       player2.acID,
				EnemyPlayerInfo: player2.playerInfo,
				EDestinySkill:   player2.destinySkill,
				EnemyData:       enemyData2,
				URL:             url,
				RoomID:          room_id,
			})
	}()

	go func() {
		player_msg.Send(player2.acID, player_msg.PlayerMsgGVGStartCode,
			player_msg.PlayerMsgGVGStart{
				PlayerInfo:      player2.playerInfo,
				DestinySkill:    player2.destinySkill,
				EnemyAcID:       player1.acID,
				EnemyPlayerInfo: player1.playerInfo,
				EDestinySkill:   player1.destinySkill,
				EnemyData:       enemyData1,
				URL:             url,
				RoomID:          room_id,
			})
	}()
}

func (m *gvgModule) notifyPlayerRobotFight(player *FightPlayer) {
	go func() {
		info := [GVG_AVATAR_COUNT]helper.AvatarState{}
		for i := 0; i < GVG_AVATAR_COUNT; i++ {
			info[i].HP = 1
			info[i].MP = 0.5
			info[i].WS = 0
		}
		player_msg.Send(player.acID, player_msg.PlayerMsgGVGStartCode,
			player_msg.PlayerMsgGVGStart{
				PlayerInfo:      player.playerInfo,
				DestinySkill:    player.destinySkill,
				EnemyAcID:       "0:0:RobotID",
				EnemyPlayerInfo: info,
				EDestinySkill:   [helper.DestinyGeneralSkillMax]int{-1, -1, -1},
			})
	}()
}

func (m *gvgModule) transScoreToGuild(timeInfo gamedata.GVGInfo2Client) {
	type tmpStruct struct {
		score []int64
		acID  []string
	}
	guildMap := make(map[string]tmpStruct, 0)

	for k, v := range m.world.players {
		gv, ok := guildMap[v.guildID]
		if !ok {
			gv.score = make([]int64, 0)
			gv.acID = make([]string, 0)
		}
		guildMap[v.guildID] = tmpStruct{
			score: append(gv.score, convertTrueScore(v.score)),
			acID:  append(gv.acID, k),
		}
	}
	for k, v := range guildMap {
		guild.GetModule(m.sid).UpdateGVGScore(k, v.score, v.acID)
	}
}

func (m *gvgModule) balance() {
	w := m.world
	// 攻城礼包
	for cityID, city := range w.cities {
		// 攻城礼包 for top N
		for i, guildID := range city.topNLeader {
			if guildID != "" {
				cfg := gamedata.GetGVGActivityGiftCfg(uint32(cityID), i)
				if cfg == nil {
					logs.Error("gvgModule balance GetGVGActivityGiftCfg nil, city %d", cityID)
					continue
				}
				logs.Debug("GVG City Gift")
				for acid, player := range m.world.players {
					if player.guildID == guildID && player.cityScore[cityID] > 0 {
						items := make(map[string]uint32, 5)
						for _, loot := range cfg.GetLoot_Table() {
							items[loot.GetGuildItemID()] = loot.GetGuildItemNum()
						}
						logs.Debug("GVG City Gift for guild: %s", guildID)
						logs.Debug("GVGActivityBalance rewards: %v", items)
						if len(items) > 0 {
							mail_sender.BatchSendMail2Account(
								acid, timail.Mail_Send_By_GVG,
								mail_sender.IDS_MAIL_GVG_FIGHT_GIFT_TITLE,
								[]string{fmt.Sprintf("%d", cityID)}, items, "GVGActivityBalance", true)
						}
					}
				}
			}
		}

	}

	// 参与礼包
	for acid, player := range m.world.players {
		if player.score > 0 {
			logs.Debug("GVG City Gift For Player: %v", player)
			cfg := gamedata.GetGVGPointGiftCfg(uint32(convertTrueScore(player.score)))
			items := make(map[string]uint32, 5)
			for _, loot := range cfg.GetLoot_Table() {
				items[loot.GetItemID()] = loot.GetItemNum()
			}
			logs.Debug("GVGPointBalance rewards: %v", items)
			if len(items) > 0 {
				mail_sender.BatchSendMail2Account(
					acid, timail.Mail_Send_By_GVG,
					mail_sender.IDS_MAIL_GVG_POINT_GIFT_TITLE,
					[]string{fmt.Sprintf("%d", convertTrueScore(player.score))}, items, "GVGPointBalance", true)
			}
		}
	}

	// 占领长安城跑马灯
	city := w.cities[CHANGAN_ID]
	if city != nil && city.leadGuildID != "" {
		info, ret := guild.GetModule(m.sid).GetGuildInfo(city.leadGuildID)
		if info == nil {
			logs.Error("gvg_opt err by %v", ret.ErrMsg)
		} else {
			logs.Trace("GVG SysNotice")
			sysnotice.NewSysRollNotice(fmt.Sprintf("%d:%d", game.Cfg.Gid, m.sid), gamedata.IDS_GVG_MARQUEE_CHANGAN_OCCUPIED).
				AddParam(sysnotice.ParamType_Value, city.leadGuildName).
				AddParam(sysnotice.ParamType_RollName, info.Base.LeaderName).
				Send()
			// 向新长安城城主发送称号
			leaderID := info.Base.LeaderAcid
			if leaderID != "" {
				logs.Debug("Send Title to leader")
				player_msg.Send(leaderID, player_msg.PlayerMsgTitleCode,
					player_msg.DefaultMsg{})
				w.setLastChangAnLeader(leaderID)
				logs.Debug("LastChanganLeader is %s", w.getLastChangAnLeader())
			}
		}
	}

	// 更新本次一统天下军团，保留一周，如果在此期间玩家没上线，将不会获得该称号
	leaders := w.genWorldLeader()
	player_msg.SendToPlayers(leaders, player_msg.PlayerMsgTitleCode,
		player_msg.DefaultMsg{})
	w.setWorldLeader(leaders)
}

func (m *gvgModule) reset() {
	logs.Debug("Reset GVG")
	m.battleRes = make(map[string]*battleRes, 0)
	m.world.LoadDataFromDB(&Gvg2DB{
		LastWorldInfo: m.world.LastWorldInfo,
	})
	m.saveToDB()
}

func (m *gvgModule) dayBalance() {
	for id, city := range m.world.cities {
		if city.leadGuildID != "" {
			guild.GetModule(m.sid).GVGBalanceForPlayer(city.leadGuildID, id)
		}
	}
}

func (m *gvgModule) dayGuildBalance() {
	for cityID, _ := range m.world.cities {
		m.dayBalanceFor1Guild(cityID)
	}
}

func (m *gvgModule) dayBalanceFor1Guild(cityID int) {
	city, ok := m.world.cities[cityID]
	if !ok {
		logs.Error("No city data for id: %d", cityID)
		return
	}
	for i, guildID := range city.topNLeader {
		if guildID != "" {
			guild.GetModule(m.sid).GVGBalanceForInventory(guildID, cityID, i)
		}
	}
}

type guildandcity struct {
	Cityid  int
	Guildid string
}

type cityprioinfo struct {
	Cityid   int
	Citytype int
}

/*
先获取城池id和城池类型进行排序
*/
func (m *gvgModule) removeGuildFirstSecond() {
	w := m.world
	holdMax := gamedata.GetGVGConfig().GetCityHoldMax()
	logs.Debug("begin to remove the %d rank Guild.  Hold Max:%d", holdMax)
	allCitys := make([]cityprioinfo, 0)
	tPrio := gamedata.GetGVGCityPrio()
	for i, value := range gamedata.GetGVGCityID() {
		allCitys = append(allCitys, cityprioinfo{value, tPrio[i]})
		logs.Debug("allCitys  append city:%d  cityType:%d", value, tPrio[i])
	}

	for i := 0; i < len(allCitys); i++ {
		for j := i + 1; j < len(allCitys); j++ {
			if allCitys[i].Citytype > allCitys[j].Citytype {
				logs.Debug("CityA %d:%d swap CityB %d:%d", allCitys[i].Cityid, allCitys[i].Citytype, allCitys[j].Cityid, allCitys[j].Citytype)
				allCitys[i], allCitys[j] = allCitys[j], allCitys[i]
			}
		}
	}
	logs.Debug("Swap end")
	logs.Debug("Exchange Citis by cityeType finished")
	m.world.removeGuild = make([]guildandcity, 0)
	logs.Debug("Continue to remove citis")
	cityofGuild := make(map[string]int, GVG_CITY_COUNT)

	for i, value := range allCitys {
		m.world.sortItem = m.world.sortItem[:0]
		for k, v := range w.guilds {
			cityScore, ok := v.cityScore[value.Cityid]
			if !ok {
				continue
			}
			item := &GVGSortItem{
				StrKey: k,
				IntVal: cityScore,
			}
			m.world.sortItem = append(m.world.sortItem, item)
		}
		sort.Sort(m.world.sortItem)

		if i < int(holdMax) {
			tlen := len(m.world.sortItem)
			if tlen > 0 {
				cityofGuild[m.world.sortItem[0].StrKey]++
				logs.Debug("The guild:%s the first of city:%d", m.world.sortItem[0].StrKey, value.Cityid)
			}
			if tlen > 1 {
				cityofGuild[m.world.sortItem[1].StrKey]++
				logs.Debug("The guild:%s the second of city:%d", m.world.sortItem[1].StrKey, value.Cityid)
			}
			continue
		} else {
			flag := 0
			for _, item := range m.world.sortItem {
				if flag < 2 {
					if cityofGuild[item.StrKey] >= int(holdMax) {
						m.world.removeGuild = append(m.world.removeGuild, guildandcity{value.Cityid, item.StrKey})
						logs.Debug("The city:%d guild:%s was been remove from %d", value.Cityid, item.StrKey, flag+1)
						delete(m.world.guilds[item.StrKey].cityScore, value.Cityid)
						continue
					} else {
						cityofGuild[item.StrKey]++
						flag++
						logs.Debug("The guild:%s the %d of city:%d %d", item.StrKey, flag+1, value.Cityid, cityofGuild[item.StrKey])
					}
				}
			}
		}
	}

	for _, value := range m.world.removeGuild {
		cityID := value.Cityid
		guildID := value.Guildid
		if cityID != 0 {
			logs.Debug("Give guild %d the remove reward of city %d", guildID, cityID)
			cfg := gamedata.GetGVGActivityGiftCfg(uint32(cityID), 1)
			if cfg == nil {
				logs.Error("gvgModule balance GetGVGActivityGiftCfg nil, city %d", guildID)
				continue
			}
			logs.Debug("GVG City Gift")
			for acid, player := range m.world.players {
				if player.guildID == guildID && player.cityScore[cityID] > 0 {
					items := make(map[string]uint32, 5)
					for _, loot := range cfg.GetLoot_Table() {
						items[loot.GetGuildItemID()] = loot.GetGuildItemNum()
					}
					logs.Debug("GVG City Gift for guild: %s", guildID)
					logs.Debug("GVGRMOVE rewards: %v", items)
					if len(items) > 0 {
						mail_sender.BatchSendMail2Account(
							acid, timail.Mail_Send_By_GVG,
							mail_sender.IDS_MAIL_GVG_FIGHT_GIFT_TITLE,
							[]string{fmt.Sprintf("%d", cityID)}, items, "GVGControlMaxhold", true)
					}
				}
			}
		}
	}
}

func (m *gvgModule) rankWinner() {
	w := m.world
	// cal top N
	m.removeGuildFirstSecond()
	for id, city := range w.cities {
		m.world.sortItem = m.world.sortItem[:0]
		for k, v := range w.guilds {
			cityScore, ok := v.cityScore[id]
			if !ok {
				continue
			}
			item := &GVGSortItem{
				StrKey: k,
				IntVal: cityScore,
			}
			m.world.sortItem = append(m.world.sortItem, item)
		}
		sort.Sort(m.world.sortItem)
		for i := 0; i < len(m.world.sortItem) && i < len(city.topNLeader); i++ {
			city.topNLeader[i] = m.world.sortItem[i].StrKey
		}
		city.leadGuildID = city.topNLeader[0]
		if v, ok := w.guilds[city.leadGuildID]; ok {
			city.leadGuildName = v.name
		} else {
			city.leadGuildName = ""
		}
		logs.Debug("Win Guild is %s, city: %d", city.leadGuildID, id)
	}
}

func (m *gvgModule) GetNowTime() int64 {
	return time.Now().Unix() + m.world.DebugOffSetTime
}

func (m *gvgModule) isActive() bool {
	return game.Cfg.GetHotActValidData(m.sid, uutil.Hot_Value_GvG)
}

func (m *gvgModule) IsWorldLeader(acID string) bool {
	return m.world.isWorldLeader(acID)
}
func (m *gvgModule) GetLastChangAnLeader() string {
	return m.world.getLastChangAnLeader()
}

func (m *gvgModule) RenameGuild(guildID string, newName string) bool {
	m.CommandExecAsync(GVGCmd{
		Typ:    Cmd_Typ_Rename_Guild,
		GuID:   guildID,
		GuName: newName,
	})
	return true
}

func (m *gvgModule) RenamePlayer(acid string, newName string) {
	m.CommandExecAsync(GVGCmd{
		Typ:  Cmd_Type_Rename_Player,
		AcID: acid,
		Name: newName,
	})
}

func (m *gvgModule) regETCD() {
	stopUrl := ""
	for i, sid := range game.Cfg.ShardId {
		if m.sid == uint(sid) {
			stopUrl = fmt.Sprintf("http://%s/%s", game.Cfg.ListenPostAddr[i], gvg_stop_url)
			logs.Info("Get GVG Stop Url: %v", stopUrl)
			break
		}
	}
	if stopUrl == "" {
		panic(fmt.Sprintf("no shard: %v  conf for gvg module", m.sid))
	}
	postService.RegGVGServices(game.Cfg.EtcdRoot, stopUrl, m.sid)
}
