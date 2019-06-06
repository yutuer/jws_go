package guild

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *GuildModule) GetGuildScienceBonus(guildUUID string, acid string,
	scienceIdx gamedata.GST_Typ) []float32 {
	res := r.guildCommandExec(guildCommand{
		Type: Command_GetGSTBonus,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1:   helper.AccountSimpleInfo{AccountID: acid},
		ParamInts: []int64{int64(scienceIdx)},
	})
	if res.ret.HasError() {
		return []float32{0.0, 0.0}
	}
	return res.ResFloat
}

func (g *GuildWorker) getGuildScienceBonus(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	m := info.GetMember(c.Player1.AccountID)
	if m == nil {
		c.resChan <- genWarnRes(errCode.GuildPlayerNotFound)
		return
	}
	scienceIdx := gamedata.GST_Typ(c.ParamInts[0])
	sc := &info.Sciences[int(scienceIdx)]
	bonus := gamedata.GetGuildScienceBonus(scienceIdx, sc.Lvl)
	res.ResFloat = bonus

	logs.Debug("getGuildScienceBonus %d %d %f", scienceIdx, sc.Lvl, bonus)
	c.resChan <- res
}

func (r *GuildModule) AddGuildSciencePoint(guildUUID string, acid string,
	scienceIdx int64, point, aftPoint int64) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_GSTLevelUp,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1:   helper.AccountSimpleInfo{AccountID: acid},
		ParamInts: []int64{scienceIdx, point, aftPoint},
	})
	if len(res.ResInt) > 0 {
		return res.ret
	}
	return res.ret
}

func (g *GuildWorker) addGuildSciencePoint(c *guildCommand) {
	res := guildCommandRes{
		ResInt: make([]int64, 1),
	}
	info := g.guild

	m := info.GetMember(c.Player1.AccountID)
	if m == nil {
		c.resChan <- genWarnRes(errCode.GuildPlayerNotFound)
		return
	}

	scienceIdx := gamedata.GST_Typ(c.ParamInts[0])
	point := c.ParamInts[1]
	aftPoint := c.ParamInts[2]

	// 加经验
	sc := &info.Sciences[int(scienceIdx)]
	cfg := gamedata.GetGuildScienceConfig(scienceIdx, sc.Lvl+1)
	if cfg == nil {
		c.resChan <- genWarnRes(errCode.GuildScienceLevelFull)
		return
	}
	m.GuildSp = aftPoint
	sc.Exp += uint32(point)

	// 记录log
	info.UpdateGuildScience()
	now_t := info.GetDebugNowTime(g.m.sid)
	m.Other.GSTDay.AddGuildSPRecord(uint32(point), now_t)
	m.Other.GSTWeek.AddGuildSPRecord(uint32(point), now_t)

	if info.Base.Level < cfg.GetLvlNeedGLv(scienceIdx) {
		c.resChan <- genWarnRes(errCode.ClickTooQuickly)
		return
	}
	need_sp := cfg.GetLvlNeedSP(scienceIdx)

	// 判断升级
	oldLv := sc.Lvl
	for sc.Exp >= need_sp {
		sc.Lvl++
		sc.Exp -= need_sp
		cfg = gamedata.GetGuildScienceConfig(scienceIdx, sc.Lvl+1)
		if cfg == nil {
			sc.Exp = 0
			break
		}
		if info.Base.Level < cfg.GetLvlNeedGLv(scienceIdx) {
			break
		}
		need_sp = cfg.GetLvlNeedSP(scienceIdx)
	}
	if sc.Lvl > oldLv { // 同步push
		if scienceIdx == gamedata.GST_MemCap {
			info.SetGuildMaxMemNum(sc.Lvl)
			g.m.updateGuildInfo2AW(info.Base)
		}
		mems := make([]string, 0, info.Base.MemNum)
		for i := 0; i < len(info.Members) && i < info.Base.MemNum; i++ {
			mems = append(mems, info.Members[i].AccountID)
		}
		syncGuildScience2Players(mems)
	}

	errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := info.DBSave(cb); err != nil {
			return err
		}
		return nil
	})
	if errCode != 0 {
		c.resChan <- genErrRes(errCode)
		return
	}

	logs.Debug("addGuildSciencePoint %d %v", scienceIdx, sc)

	// logiclog statistic
	info.ActivePlayerStatistic.JoinStic(c.Player1.AccountID)
	c.resChan <- res
}

func (r *GuildModule) GetGuildScienceLog(guildUUID string, acid string,
	bGstWeek bool) (GuildRet, []string, []int64, []int64) {
	typ := Command_GetGSTDay
	if bGstWeek {
		typ = Command_GetGSTWeek
	}
	res := r.guildCommandExec(guildCommand{
		Type: typ,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acid},
	})
	return res.ret, res.ResStr, res.ResInt, res.ResInt2
}

func (g *GuildWorker) scienceDayLog(c *guildCommand, bGstWeek bool) {
	res := guildCommandRes{}
	info := g.guild

	m := info.GetMember(c.Player1.AccountID)
	if m == nil {
		c.resChan <- genWarnRes(errCode.GuildPlayerNotFound)
		return
	}

	info.UpdateGuildScience()

	res.ResStr = make([]string, 0, 20)
	res.ResInt = make([]int64, 0, 20)
	res.ResInt2 = make([]int64, 0, 20)
	for i := 0; i < len(info.Members) && i < info.Base.MemNum; i++ {
		mem := info.Members[i]
		dInfo := mem.Other.GSTDay
		if bGstWeek {
			dInfo = mem.Other.GSTWeek
		}
		if dInfo.GSP > 0 {
			res.ResStr = append(res.ResStr, mem.Name)
			res.ResInt = append(res.ResInt, int64(dInfo.GSP))
			res.ResInt2 = append(res.ResInt2, dInfo.LastTimeStamp)
		}
	}

	c.resChan <- res
}

func (r *GuildModule) DebugSetGuildScienceLevel(guildUUID string, acid string,
	scienceIdx gamedata.GST_Typ, lvl int64) {
	r.guildCommandExec(guildCommand{
		Type: Command_DebugSetGSTLevel,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1:   helper.AccountSimpleInfo{AccountID: acid},
		ParamInts: []int64{int64(scienceIdx), lvl},
	})
}

func (g *GuildWorker) debugSetGuildScienceLevel(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	m := info.GetMember(c.Player1.AccountID)
	if m == nil {
		c.resChan <- genWarnRes(errCode.GuildPlayerNotFound)
		return
	}
	scienceIdx := gamedata.GST_Typ(c.ParamInts[0])
	lvl := c.ParamInts[1]
	sc := &info.Sciences[int(scienceIdx)]
	cfg := gamedata.GetGuildScienceConfig(scienceIdx, uint32(lvl))
	if cfg != nil {
		sc.Lvl = uint32(lvl)
		sc.Exp = 0
		errCode := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
			if err := info.DBSave(cb); err != nil {
				return err
			}
			return nil
		})
		if errCode != 0 {
			c.resChan <- genErrRes(errCode)
			return
		}
	}

	c.resChan <- res
}
