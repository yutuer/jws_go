package guild

import (
	"encoding/json"
	"math/rand"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/message"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *GuildModule) AddGuildInventory(guildUUID string, ids []string, counts []uint32, reason string) GuildRet {
	if ids == nil || counts == nil || len(ids) <= 0 || len(ids) != len(counts) {
		return GuildRet{ErrCode: Err_Param, ErrMsg: "Err_Param"}
	}
	loots := make([]guild_info.GuildInventoryLoot, 0, len(ids))
	for i, id := range ids {
		loots = append(loots, guild_info.GuildInventoryLoot{id, counts[i]})
	}
	res := r.guildCommandExec(guildCommand{
		Type: Command_AddGuildInventory,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		inventoryLoot: loots,
		Reason:        reason,
	})

	return res.ret
}

func (g *GuildWorker) addGuildInventory(c *guildCommand) {
	res := guildCommandRes{}
	var isFull bool
	if c.Reason == "debug_lost" {
		isFull = g.guild.GuildInfoBase.LostInventory.AddGuildInventory(
			g.m.sid, g.guild.Base.GuildUUID, c.inventoryLoot,
			c.Reason, g.guild.GetDebugNowTime(g.m.sid))
		g.guild.GuildInfoBase.LostInventory.LostActive = true
		g.guild.GuildInfoBase.LostInventory.DisappearTime = 0
	} else {
		isFull = g.guild.GuildInfoBase.Inventory.AddGuildInventory(
			g.m.sid, g.guild.Base.GuildUUID, c.inventoryLoot,
			c.Reason, g.guild.GetDebugNowTime(g.m.sid))
	}

	info := g.guild
	// save guild
	errC := dbCmdBuffExec(func(cb redis.CmdBuffer) error {
		if err := info.DBSave(cb); err != nil {
			return err
		}
		return nil
	})
	if errC != 0 {
		c.resChan <- genErrRes(errC)
		return
	}
	if isFull {
		c.resChan <- genWarnRes(errCode.GuildInventoryFull)
		return
	}
	c.resChan <- res
	return
}

// ret, acids, names, sts, gss
func (r *GuildModule) GetApplyListGuildInventoryItem(guildUUID,
	acID, loot string) (
	GuildRet, []string, []string, []int64, []int64) {
	res := r.guildCommandExec(guildCommand{
		Type: Command_ApplyListGuildInventory,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acID},
		LootId:  loot,
	})
	return res.ret, res.ResStr, res.ResStr2, res.ResInt, res.ResInt2
}

func (g *GuildWorker) getApplyListGuildInventoryItem(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	// 是否在公会
	mem := info.GetGuildMemInfo(c.Player1.AccountID)
	if mem == nil {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Leave,
		}
		c.resChan <- res
		return
	}
	// 申请列表
	gi := g.guild.GuildInfoBase.LostInventory
	acids, sts := gi.GetGuildInventoryApplyList(c.LootId,
		g.guild.GetDebugNowTime(g.m.sid))
	_genApplyList(info, acids, sts, &res)
	c.resChan <- res
	return
}

// ret, acids, names, sts, gss
func (r *GuildModule) ApproveGuildInventoryItem(guildUUID,
	acID, acID2, loot string, timestamp int64, aggree bool, rand *rand.Rand) (
	GuildRet, []string, []string, []int64, []int64) {

	res := r.guildCommandExec(guildCommand{
		Type: Command_ApproveGuildInventory,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1:    helper.AccountSimpleInfo{AccountID: acID},
		Player2:    helper.AccountSimpleInfo{AccountID: acID2},
		LootId:     loot,
		ParamInts:  []int64{timestamp},
		ParamBools: []bool{aggree},
		Rand:       rand,
	})
	return res.ret, res.ResStr, res.ResStr2, res.ResInt, res.ResInt2
}

func (g *GuildWorker) approveGuildInventoryItem(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	acids, sts := info.LostInventory.GetGuildInventoryApplyList(c.LootId,
		g.guild.GetDebugNowTime(g.m.sid))
	_genApplyList(info, acids, sts, &res)
	// 是否在公会
	mem := info.GetGuildMemInfo(c.Player1.AccountID)
	if mem == nil {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Leave,
		}
		c.resChan <- res
		return
	}
	pmem := info.GetGuildMemInfo(c.Player2.AccountID)
	if pmem == nil {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Leave,
		}
		c.resChan <- res
		return
	}
	// 权限
	cfg := gamedata.GetGuildPosData(mem.GuildPosition)
	if cfg == nil || cfg.GetAllotPower() <= 0 {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Position,
		}
		c.resChan <- res
		return
	}
	// approve
	aggree := c.ParamBools[0]
	timestamp := c.ParamInts[0]
	code, acids, sts := info.LostInventory.ApproveGuildInventory(c.LootId,
		c.Player2.AccountID, info.Base.GuildUUID,
		info.Base.Name, timestamp, aggree,
		c.Rand, g.guild.GetDebugNowTime(g.m.sid))
	res.ret = GuildRet{
		CodeLevel: Code_Inner_Msg,
		ErrCode:   code,
	}
	// 填充申请列表
	_genApplyList(info, acids, sts, &res)
	if code == guild_info.Inventory_Success {
		g._saveAndLog(info, c, mem.Name, pmem.Name, aggree)
	}
	c.resChan <- res
	return
}

func (r *GuildModule) ApplyGuildInventoryItem(guildUUID, acID, loot string) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_ApplyGuildInventory,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acID},
		LootId:  loot,
	})
	return res.ret
}

func (g *GuildWorker) applyGuildInventoryItem(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	// 是否在公会
	pasMem := info.GetGuildMemInfo(c.Player1.AccountID)
	if pasMem == nil {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Leave,
		}
		c.resChan <- res
		return
	}
	// 申请
	code := info.LostInventory.ApplyGuildInventory(c.LootId,
		g.guild.GetDebugNowTime(g.m.sid),
		pasMem,
		c.Channel)
	res.ret = GuildRet{
		CodeLevel: Code_Inner_Msg,
		ErrCode:   code,
	}
	if code == guild_info.Inventory_Success {
		// save guild
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
	return
}

func (r *GuildModule) ExchangeGuildInventoryItem(guildUUID,
	acID, loot string, rand *rand.Rand) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_ExchangeGuildInventory,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acID},
		LootId:  loot,
		Rand:    rand,
	})
	return res.ret
}

func (g *GuildWorker) exchangeGuildInventoryItem(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	// 是否在公会
	mem := info.GetGuildMemInfo(c.Player1.AccountID)
	if mem == nil {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Leave,
		}
		c.resChan <- res
		return
	}
	// 权限
	cfg := gamedata.GetGuildPosData(mem.GuildPosition)
	if cfg == nil || cfg.GetAllotPower() <= 0 {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Position,
		}
		c.resChan <- res
		return
	}

	code := info.LostInventory.ExchangeGuildInventory(c.LootId,
		c.Player1.AccountID, info.Base.GuildUUID, info.Base.Name,
		g.guild.GetDebugNowTime(g.m.sid), c.Rand,
		mem, c.Channel)
	res.ret = GuildRet{
		CodeLevel: Code_Inner_Msg,
		ErrCode:   code,
	}
	// 填充申请列表
	if code == guild_info.Inventory_Success {
		g._saveAndLog(info, c, mem.Name, mem.Name, true)
	}
	c.resChan <- res
	return
}

func (g *GuildWorker) _saveAndLog(info *GuildInfo,
	c *guildCommand, name1, name2 string, islog bool) {

	// save guild
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
	// 记log
	if islog {
		r := AssignInventroyRecord{
			Time:         time.Now().Unix(),
			AssignerName: name1,
			ItemId:       c.LootId,
			ReceiverName: name2,
		}
		bb, err := json.Marshal(r)
		if err != nil {
			logs.Error("assignGuildInventoryItem json.Marshal log err %s", err.Error())
		} else {
			message.SendPlayerMsgs(g.guild.Base.GuildUUID,
				guild_info.GuildLostInventoryMsgTableKey, guild_info.GuildInventoryMsgCount,
				message.PlayerMsg{
					Params: []string{string(bb)},
				})
		}
	}
}

func (r *GuildModule) AssignGuildInventoryItem(guildUUID, acID, pasAcID, lootId string, rand *rand.Rand) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_AssignGuildInventory,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		Player1: helper.AccountSimpleInfo{AccountID: acID},
		Player2: helper.AccountSimpleInfo{AccountID: pasAcID},
		LootId:  lootId,
		Rand:    rand,
	})
	return res.ret
}

// 此操作返回值标志是成功,只是带有不同错误码,客户端根据不同错误码,决定怎么刷新UI, 而不是弹出错误提示框
func (g *GuildWorker) assignGuildInventoryItem(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	// 职位检查
	mem := info.GetGuildMemInfo(c.Player1.AccountID)
	if mem == nil {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Act_Leave,
		}
		c.resChan <- res
		return
	}
	cfg := gamedata.GetGuildPosData(mem.GuildPosition)
	if cfg == nil || cfg.GetAllotPower() <= 0 {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Position,
		}
		c.resChan <- res
		return
	}
	// 是否在公会
	pasMem := info.GetGuildMemInfo(c.Player2.AccountID)
	if pasMem == nil {
		res.ret = GuildRet{
			CodeLevel: Code_Inner_Msg,
			ErrCode:   guild_info.Inventory_Leave,
		}
		c.resChan <- res
		return
	}
	code := info.Inventory.AssignGuildInventoryItem(g.m.sid, info.Base.GuildUUID, c.LootId,
		c.Player2.AccountID, info.Base.Name, c.Rand, g.guild.GetDebugNowTime(g.m.sid))
	res.ret = GuildRet{
		CodeLevel: Code_Inner_Msg,
		ErrCode:   code,
	}
	if code == guild_info.Inventory_Success {
		// save guild
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
		// 记log
		r := AssignInventroyRecord{
			Time:         time.Now().Unix(),
			AssignerName: mem.Name,
			ItemId:       c.LootId,
			ReceiverName: pasMem.Name,
		}
		bb, err := json.Marshal(r)
		if err != nil {
			logs.Error("assignGuildInventoryItem json.Marshal log err %s", err.Error())
		} else {
			message.SendPlayerMsgs(g.guild.Base.GuildUUID,
				guild_info.GuildInventoryMsgTableKey, guild_info.GuildInventoryMsgCount,
				message.PlayerMsg{
					Params: []string{string(bb)},
				})
		}
	}
	c.resChan <- res
	return
}

type AssignInventroyRecord struct {
	Time         int64  `json:"t"`
	AssignerName string `json:"as_nm"`
	ItemId       string `json:"iid"`
	ReceiverName string `json:"re_nm"`
}

func (r *GuildModule) DebugSetGuildInventoryTime(guildUUID string, t int64) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_DebugSetGuildInventory,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		DebugTime: t,
	})
	return res.ret
}

func (g *GuildWorker) debugSetGuildInventoryTime(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	info.SetDebugTime(c.DebugTime, g.m.sid)
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
	c.resChan <- res
	return
}

func (r *GuildModule) DebugResetGuildInventoryTime(guildUUID string) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_DebugResetGuildInventory,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	})
	return res.ret
}

func (g *GuildWorker) debugResetGuildInventoryTime(c *guildCommand) {
	res := guildCommandRes{}
	info := g.guild

	info.Inventory.NextRefTime = 0
	info.Inventory.NextResetTime = 0
	info.DebugTimeAbsolute = 0
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
	c.resChan <- res
	return
}

func _genApplyList(info *GuildInfo, acids []string,
	sts []int64, res *guildCommandRes) {
	mems := make(map[string]*helper.AccountSimpleInfo, info.Base.MemNum)
	for i := 0; i < info.Base.MemNum; i++ {
		m := &info.Members[i]
		mems[m.AccountID] = m
	}
	names := make([]string, len(acids))
	gss := make([]int64, len(acids))
	for i, acid := range acids {
		mem := mems[acid]
		if mem == nil {
			logs.Error("GuildInventory _genApplyList mem not found %s", acid)
			continue
		}
		names[i] = mem.Name
		gss[i] = int64(mem.CurrCorpGs)
	}
	res.ResStr = acids
	res.ResStr2 = names
	res.ResInt = sts
	res.ResInt2 = gss
}
