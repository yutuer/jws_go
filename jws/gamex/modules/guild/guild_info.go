package guild

import (
	"fmt"

	"time"

	"reflect"

	"sort"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	. "vcs.taiyouxi.net/jws/gamex/models/account/warm"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/base"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/common/guild_player_rank"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	oldVer        = 0
	newVer        = 1
	newTimeStramp = 1489006800
)

type GuildInfo struct {
	GuildInfoBase
	posNum          [gamedata.Guild_Pos_Count]int
	memSyncReceiver []helper.IGuildMemSyncReceiver
	dirtiesCheck    map[string]interface{}
	Ver             int64 `redis:"version"`

	needNotifySomething bool
	needNotifyAct       [base.GuildActCount]bool
	lastNotifyTime      int64
	needStore2DB        bool
	shardId             uint
	saveReqCount        int // 存库请求次数， 累计一定次数后存一次库
	lastSaveTime        int64
}

func (g *GuildInfo) Tick(nowT int64) {
	g.TryRefresh(g.shardId)
	if g.needNotifySomething {
		// cheat改时间后不能同步，加个时间revert
		if nowT-g.lastNotifyTime > 3 || g.lastNotifyTime > nowT {
			g.lastNotifyTime = nowT
			g.notifyAllExp(g.needNotifyAct, time.Now().Unix()) // 这里使用服务器时间
			g.needNotifySomething = false
			g.needNotifyAct = [base.GuildActCount]bool{}
		}
	}
}

// TODO by ljz tmp value
func (g *GuildInfo) SetGuildTmpVer() {
	logs.Debug("shardId: %d, serverstarttime: %d", g.shardId, game.ServerStartTime(g.shardId))
	if game.ServerStartTime(g.shardId) > newTimeStramp {
		g.Base.GuildTmpVer = newVer
	} else {
		g.Base.GuildTmpVer = oldVer
	}
	logs.Debug("GuildTmpVer is %d", g.Base.GuildTmpVer)
}

func (g *GuildInfo) SetGuildMaxMemNum(lv uint32) {
	if g.Base.GuildTmpVer == oldVer {
		g.Base.MaxMemNum = int(gamedata.GetGuildLevelMemLimit(lv))
	} else {
		g.Base.MaxMemNum = int(gamedata.GetGuildLevelNewMemLimit(lv))
	}
}

func (g *GuildInfo) SetNeedSave2DB() {
	g.needStore2DB = true
}

func (g *GuildInfo) IsNeedSave2DB() bool {
	return g.needStore2DB
}

func (g *GuildInfo) SetNoNeedSave2DB() {
	g.needStore2DB = false
}

func (g *GuildInfo) GetGuildLv() uint32 {
	return g.Base.Level
}

func (g *GuildInfo) UpdateGs(gsAdd int64) {
	old := g.Base.GuildGSSum
	g.Base.GuildGSSum += gsAdd
	sid, _ := guild_info.GetShardIdByGuild(g.GuildInfoBase.Base.GuildUUID)
	rank.GetModule(sid).RankGuildGs.Add(&g.GuildInfoBase.Base, g.Base.GuildGSSum, old, false)
}

func (g *GuildInfo) AddXP(c int64) {
	logs.Trace("GuildAddXP %s %d", g.Base.GuildUUID, c)
	g.Base.XpCurr += c
	data := gamedata.GetGuildXpNeedNext(g.Base.Level)
	data1 := gamedata.GetGuildXpNeedNext(g.Base.Level + 1)
	ol := g.Base.Level
	for data > 0 && data1 > 0 && g.Base.XpCurr >= data {
		g.Base.XpCurr -= data
		g.Base.Level += 1
		data = gamedata.GetGuildXpNeedNext(g.Base.Level)
	}
	if g.Base.Level > ol {
		if g.Base.Level == 2 {
			g.Base.Guild2LvlTimes = g.GetDebugNowTime(g.shardId)
		}
		g.GuildLog.AddLog(IDS_GUILD_LOG_9, []string{fmt.Sprintf("%d", g.Base.Level)})
		sid, _ := guild_info.GetShardIdByGuild(g.Base.GuildUUID)
		GetModule(sid).updateGuildInfo2AW(g.Base)
		logs.Debug("Guild LevelUp %d->%d", ol, g.Base.Level)
		logiclog.LogGuildLvUp(g.Base.GuildUUID, ol, g.Base.Level, "")
	}
}

func (g *GuildInfo) AddSP(acid string, c int64) {
	for i := 0; i < len(g.Members) && i < g.Base.MemNum; i++ {
		if g.Members[i].AccountID == acid {
			g.Members[i].GuildSp += c
		}
	}
}

func newGuildInfo(baseInfo *GuildSimpleInfo, player *helper.AccountSimpleInfo) *GuildInfo {
	g := &GuildInfo{
		GuildInfoBase: GuildInfoBase{
			Base: *baseInfo,
		},
		Ver: helper.CurrDBVersion,
	}
	a, err := db.ParseAccount(player.AccountID)
	if err != nil {
		return nil
	}
	err, id := genGuildId(a.ShardId)
	if err != nil {
		return nil
	}
	g.Base.GuildID = id
	g.Base.GuildUUID = GenGuildUUidByPlayer(a.GameId, a.ShardId)
	g.Base.CreateTS = time.Now().Unix()
	g.Base.Level = 1
	g.shardId = a.ShardId
	g.SetGuildTmpVer()
	g.SetGuildMaxMemNum(0)
	player.GuildPosition = gamedata.Guild_Pos_Chief
	g.addMem(player)
	g.GuildLog.AddLog(IDS_GUILD_LOG_1, []string{player.Name})
	// save db
	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := saveGuild(g.Base.Name, g.Base.GuildUUID, g.Base.GuildID, cb); err != nil {
			return err
		}
		if err := addGuildMem(player.AccountID, g.Base.GuildUUID, cb); err != nil {
			return err
		}
		if err := g.DBSave(cb); err != nil {
			return err
		}
		return nil
	})
	if errCode != 0 {
		return nil
	}
	g.memSyncReceiver = make([]helper.IGuildMemSyncReceiver, 0, 8)
	g.ActBoss.SetGuildHandler(g)
	g.GatesEnemyData.PlayerRank = *(guild_player_rank.NewPlayerSimpleInfoRankByCap(MaxGuildMember + 2))
	return g
}

func (p *GuildInfo) initGuildInfo(shardID uint, guildUUID string) error {
	p.Base.GuildUUID = guildUUID

	err := TryWarmData(false, shardID, p, true)
	if err != nil {
		if err != driver.RESTORE_ERR_Profile_No_Data {
			logs.Error("Load Guild %s Err By %s", guildUUID, err.Error())
		}
		return err
	}
	if p.Base.MemNum <= 0 {
		return fmt.Errorf("GuildInfo initGuildInfo MemNum <= 0 %s", p.Base.GuildUUID)
	}
	p.memSyncReceiver = make([]helper.IGuildMemSyncReceiver, 0, 8)
	p.shardId = shardID
	p.ActBoss.SetGuildHandler(p)
	p.ActBoss.Init()
	p.GatesEnemyData.PlayerRank.Init()
	p.CheckGuildChief()

	//ADD by qiaozhu
	logs.Debug("---goin guild.VerUpdate")
	err = VerUpdate(p)
	if nil != err {
		logs.Error("Update Guild %s Err By %s", guildUUID, err.Error())
		return err
	}
	//Add by qiaozhu End
	return nil
}

func (guildInfo *GuildInfo) GetGuildMemInfo(acid string) *helper.AccountSimpleInfo {
	for i := 0; i < int(guildInfo.Base.MemNum); i++ {
		mem := &guildInfo.Members[i]
		if mem.AccountID == acid {
			return mem
		}
	}
	return nil
}

func (g *GuildInfo) addMem(playerInfo *helper.AccountSimpleInfo) {
	g.Members[g.Base.MemNum] = *playerInfo
	g.Base.MemNum++
	for i := 0; i < len(g.memSyncReceiver); i++ {
		g.memSyncReceiver[i].OnGuildChange(g.Members[:g.Base.MemNum])
	}
	syncGuild2Player(playerInfo.AccountID,
		g.Base.GuildUUID,
		g.Base.Name,
		playerInfo.GuildPosition,
		int(g.Base.Level))
	syncGuildScience2Player(playerInfo.AccountID)
	syncGuildRedPacket2Players(playerInfo.AccountID)
	g.UpdateGs(int64(playerInfo.CurrCorpGs))
}

func (p *GuildInfo) delMember(acID, kicker string) guildCommandRes {
	if acID == "" {
		return guildCommandRes{}
	}

	for i := 0; i < len(p.Members) && i < p.Base.MemNum; i++ {
		if p.Members[i].AccountID == acID {
			name := p.Members[i].Name
			gs := p.Members[i].CurrCorpGs
			if i != p.Base.MemNum-1 {
				p.Members[i] = p.Members[p.Base.MemNum-1]
			}

			p.Members[p.Base.MemNum-1] = helper.AccountSimpleInfo{}
			p.Base.MemNum -= 1
			for i := 0; i < len(p.memSyncReceiver); i++ {
				p.memSyncReceiver[i].OnGuildChange(p.Members[:p.Base.MemNum])
			}
			if kicker != "" {
				p.GuildLog.AddLog(IDS_GUILD_LOG_4, []string{kicker, name})
			} else {
				p.GuildLog.AddLog(IDS_GUILD_LOG_3, []string{name})
			}
			assignID, assignTimes := p.Inventory.GetAssignTimesByAcID(acID)
			p.Inventory.OnDelMem(acID)
			p.LostInventory.OnDelMem(acID)
			p.GatesEnemyData.PlayerRank.OnPlayerDel(acID)
			p.UpdateGs(-int64(gs))
			nowDebugTime := p.GetDebugNowTime(p.shardId)
			nextEnterGuildTime := util.GetNextDailyTime(gamedata.GetCommonDayBeginSec(nowDebugTime), nowDebugTime)
			syncLeaveGuild2Player(acID, nowDebugTime, nextEnterGuildTime, assignID, assignTimes)
			syncGuildRedPacket2Players(acID)
			return guildCommandRes{}
		}
	}
	return genWarnRes(errCode.GuildPlayerNotFound)
}

func (p *GuildInfo) AddMemSyncReceiver(r helper.IGuildMemSyncReceiver) {
	p.memSyncReceiver = append(p.memSyncReceiver, r)
}

func (p *GuildInfo) DelMemSyncReceiver(id int) {
	n := make([]helper.IGuildMemSyncReceiver, 0, len(p.memSyncReceiver))
	for i := 0; i < len(p.memSyncReceiver); i++ {
		if p.memSyncReceiver[i].GetMemSyncReceiverID() != id {
			n = append(n, p.memSyncReceiver[i])
		}
	}
	p.memSyncReceiver = n
}

func (p *GuildInfo) UseGateEnemyCount() bool {
	c := p.Base.GetGateEnemyCount()
	if c > 0 {
		p.Base.GatesEnemyCount -= 1
		return true
	}
	return false
}

func (p *GuildInfo) SyncApply2AllMem(hasApply bool) {
	mems := make([]string, 0, len(p.Members))
	for i := 0; i < len(p.Members) && i < p.Base.MemNum; i++ {
		mem := p.Members[i]
		if gamedata.CheckApprovePosition(mem.GuildPosition) {
			mems = append(mems, mem.AccountID)
		}
	}

	syncGuildApply2Players(mems, hasApply)
}

func (p *GuildInfo) DBName() string {
	return fmt.Sprintf("%s:%s", Table_Guild, p.Base.GuildUUID)
}

func (p *GuildInfo) DBSave(cb redis.CmdBuffer) error {
	p.saveReqCount++
	now := time.Now().Unix()
	if p.saveReqCount < 30 && now-p.lastSaveTime < 30 {
		return nil
	}
	p.lastSaveTime = now
	p.saveReqCount = 0
	return p.dbSave(cb)
}

func (p *GuildInfo) ForceDBSave(cb redis.CmdBuffer) error {
	p.saveReqCount = 0
	p.lastSaveTime = time.Now().Unix()
	return p.dbSave(cb)
}

func (p *GuildInfo) dbSave(cb redis.CmdBuffer) error {
	key := p.DBName()
	err, newDirtyCheck, chged := driver.DumpToHashDBCmcBufferCheckDirty(
		cb, key, p, p.dirtiesCheck)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	if !game.Cfg.IsRunModeProd() {
		if !reflect.DeepEqual(p.dirtiesCheck, newDirtyCheck) {
			logs.Trace("Save Guild Data: %s %v", p.Base.GuildUUID, chged)
		} else {
			logs.Trace("Save Guild Data Clean: %s %v", p.Base.GuildUUID, chged)
		}
	}

	p.dirtiesCheck = newDirtyCheck
	return nil
}

func (p *GuildInfo) DBLoad(logInfo bool) error {
	key := p.DBName()

	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, p, false, logInfo)
	if err != nil {
		return err
	}

	for i := 0; i < len(p.Members); i++ {
		pos := p.Members[i].GuildPosition
		if pos < 0 || pos >= gamedata.Guild_Pos_Count {
			continue
		}
		p.posNum[pos] += 1
	}

	p.dirtiesCheck = driver.GenDirtyHash(p)

	logs.Trace("GuildInfo posNum %v", p.posNum)

	return err
}

func syncLeaveGuild2Player(acid string, leaveTime int64, nextJoinTime int64, lootID []string, times []int64) {
	player_msg.Send(acid, player_msg.PlayerMsgGuildInfoSyncCode,
		player_msg.PlayerGuildInfoUpdate{
			GuildUUID:     "",
			GuildName:     "",
			GuildPosition: 0,
			GuildLv:       0,
			LeaveTime:     leaveTime,
			NextJoinTime:  nextJoinTime,
			AssignID:      lootID,
			AssignTimes:   times,
		})
}

func syncGuild2Player(acid, guildUUID, name string, guildPosition, lv int) {
	player_msg.Send(acid, player_msg.PlayerMsgGuildInfoSyncCode,
		player_msg.PlayerGuildInfoUpdate{
			GuildUUID:     guildUUID,
			GuildName:     name,
			GuildPosition: guildPosition,
			GuildLv:       lv,
		})
}

func syncGuild2Players(acids []string, guildUUID, name string, guildPosition, lv int) {
	player_msg.SendToPlayers(acids, player_msg.PlayerMsgGuildInfoSyncCode,
		player_msg.PlayerGuildInfoUpdate{
			GuildUUID:     guildUUID,
			GuildName:     name,
			GuildPosition: guildPosition,
			GuildLv:       lv,
		})
}

func syncGuildAndApply2Player(acid string, guildUUID, name string, guildPosition int, hasApply bool) {
	player_msg.Send(acid, player_msg.PlayerMsgGuildInfoSyncCode,
		player_msg.PlayerGuildInfoUpdate{
			GuildUUID:     guildUUID,
			GuildName:     name,
			GuildPosition: guildPosition,
		})
	if gamedata.CheckApprovePosition(guildPosition) && hasApply {
		player_msg.Send(acid, player_msg.PlayerMsgGuildApplyInfoSyncCode,
			player_msg.PlayerGuildApplyUpdate{
				HasApplyCanApprove: hasApply,
			})
	}
}

func syncGuildApply2Players(acids []string, hasApply bool) {
	player_msg.SendToPlayers(acids, player_msg.PlayerMsgGuildApplyInfoSyncCode,
		player_msg.PlayerGuildApplyUpdate{
			HasApplyCanApprove: hasApply,
		})
}

func syncGuildScience2Player(acid string) {
	player_msg.Send(acid, player_msg.PlayerMsgGuildScienceInfoSyncCode,
		player_msg.DefaultMsg{})
}

func syncGuildScience2Players(acids []string) {
	player_msg.SendToPlayers(acids, player_msg.PlayerMsgGuildScienceInfoSyncCode,
		player_msg.DefaultMsg{})
}

func syncGuildRedPacket2Players(acid string) {
	player_msg.Send(acid, player_msg.PlayerMsgGuildRedPacketSyncCode,
		player_msg.DefaultMsg{})
}

func (g *GuildInfo) AddGuildInventory(ids []string, counts []uint32, reason string) {
	loots := make([]guild_info.GuildInventoryLoot, 0, len(ids))
	for i, id := range ids {
		loots = append(loots, guild_info.GuildInventoryLoot{id, counts[i]})
	}
	g.Inventory.AddGuildInventory2Prepare(g.shardId, g.Base.GuildUUID,
		loots, reason, g.GetDebugNowTime(g.shardId))
}

func (g *GuildInfo) UpdateGuildScience() {
	now_t := g.GetDebugNowTime(g.shardId)
	if now_t >= g.GSTTodayResetTime { // 清空本日所有记录
		for i := 0; i < g.Base.MemNum && i < len(g.Members); i++ {
			mem := &g.Members[i]
			mem.Other.GSTDay = helper.GuildSPRecord{}
		}
		g.GSTTodayResetTime = getGSTDayResetTime(now_t)
	}
	if now_t >= g.GSTWeekResetTime { // 清空本周所有记录
		for i := 0; i < g.Base.MemNum && i < len(g.Members); i++ {
			mem := &g.Members[i]
			mem.Other.GSTWeek = helper.GuildSPRecord{}
		}
		g.GSTWeekResetTime = getGSTWeekResetTime(now_t)
		logs.Debug("UpdateGuildScience GSTWeekResetTime %d", g.GSTWeekResetTime)
	}
}

func getGSTDayResetTime(now_t int64) int64 {
	cfg := gamedata.GetGSTConfig()
	t := time.Unix(now_t, 0)
	t = t.In(util.ServerTimeLocal)
	nt, err := time.ParseInLocation("2006-1-2 15:04",
		fmt.Sprintf("%d-%d-%d %s", t.Year(), t.Month(), t.Day(),
			cfg.GetLogDailyResetTime()), util.ServerTimeLocal)
	if err != nil {
		logs.Error("getGSTDayResetTime time.ParseInLocation err %v", err)
		return 0
	}
	if nt.Unix() <= now_t {
		return nt.Unix() + gamedata.Day2Second
	}
	return nt.Unix()
}

func getGSTWeekResetTime(now_t int64) int64 {
	cfg := gamedata.GetGSTConfig()
	return util.GetNextWeekTime(now_t,
		int(cfg.GetLogWeeklyResetDay()),
		cfg.GetLogDailyResetTime())
}

func (guildInfo *GuildInfo) CheckGuildChief() {
	// 统计所有会长头衔的成员
	chiefs := make([]*helper.AccountSimpleInfo, 0, 5)
	for i := 0; i < int(guildInfo.Base.MemNum); i++ {
		if guildInfo.Members[i].GuildPosition == gamedata.Guild_Pos_Chief {
			chiefs = append(chiefs, &guildInfo.Members[i])
		}
	}
	// 如果有多个会长，只保留一个
	if len(chiefs) > 1 {
		guildInfo.selectOneIfMulti(chiefs)
	} else if len(chiefs) == 0 {
		guildInfo.appointNewChiefIfNone()
	}
}

func (guildInfo *GuildInfo) appointNewChiefIfNone() {
	candidates := make(guild_info.GuildSortMember, 0, guildInfo.Base.MemNum) // 候选人
	for i := 0; i < guildInfo.Base.MemNum; i++ {
		candidates = append(candidates, &guildInfo.Members[i])
	}
	if len(candidates) < 1 {
		logs.Debug("check guild no any body %s", guildInfo.Base.GuildUUID)
		return
	}
	sort.Sort(candidates)
	newChief := candidates[0]
	oldPos := newChief.GuildPosition
	newChief.GuildPosition = gamedata.Guild_Pos_Chief
	guildInfo.posNum[oldPos]--
	guildInfo.posNum[gamedata.Guild_Pos_Chief] = 1
	syncGuild2Player(newChief.AccountID,
		guildInfo.Base.GuildUUID,
		guildInfo.Base.Name,
		gamedata.Guild_Pos_Chief,
		int(guildInfo.Base.Level))
	guildInfo.UpdateChiefInfo(newChief.AccountID, newChief.Name)
}

func (guildInfo *GuildInfo) selectOneIfMulti(chiefs []*helper.AccountSimpleInfo) {
	logs.Warn("find more than one chief, %d, %d", guildInfo.Base.GuildID, len(chiefs))
	rightChiefIndex := 0 // 默认选0位置
	for i, chief := range chiefs {
		if chief.AccountID == guildInfo.Base.LeaderAcid {
			rightChiefIndex = i
			break
		}
	}
	for i, chief := range chiefs {
		if i != rightChiefIndex {
			chief.GuildPosition = gamedata.Guild_Pos_Mem
			guildInfo.posNum[gamedata.Guild_Pos_Mem]++
			syncGuild2Player(chief.AccountID,
				guildInfo.Base.GuildUUID,
				guildInfo.Base.Name,
				gamedata.Guild_Pos_Mem,
				int(guildInfo.Base.Level))
		}
	}
	guildInfo.posNum[gamedata.Guild_Pos_Chief] = 1
	guildInfo.UpdateChiefInfo(chiefs[rightChiefIndex].AccountID, chiefs[rightChiefIndex].Name)
	logs.Warn("keep a chief %s, %s", chiefs[rightChiefIndex].AccountID, chiefs[rightChiefIndex].Name)
}

func (guildInfo *GuildInfo) GetAllMemberAcids() []string {
	acids := make([]string, guildInfo.Base.MemNum)
	for i := 0; i < guildInfo.Base.MemNum; i++ {
		acids[i] = guildInfo.Members[i].AccountID
	}
	return acids
}
