package guild_info

import (
	"math/rand"

	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

const (
	GuildInventoryMsgTableKey = "GuildInventory"
	GuildInventoryMsgCount    = 60

	GuildLostInventoryMsgTableKey = "GuildLostInventory"
)

type GuildInventory struct {
	NextRefTime   int64                `json:"nrt"`
	NextResetTime int64                `json:"nrestt"`
	PrepareLoots  []GuildInventoryLoot `json:"prels"`

	Loots         []GuildInventoryLootAndMem `json:"lms"`
	LostActive    bool                       `json:"lost_active"`
	DisappearTime int64                      `json:"disappear_time"`
}

type GuildInventoryLoot struct {
	LootId string `json:"lid" codec:"lid"`
	Count  uint32 `json:"c" codec:"c"`
}

type GuildInventoryLootAndMem struct {
	Loot       GuildInventoryLoot        `json:"loot"`
	ApplyAcids []GuildInventoryLootApply `json:"aply_acids"`

	AssignMems []GuildInventoryLootAssign `json:"ams"`
}

type GuildInventoryLootApply struct {
	Acid      string `json:"acid"`
	TimeStamp int64  `json:"ts"`
}

type GuildInventoryLootAssign struct {
	Acid        string `json:"acid" codec:"acid"`
	AssignTimes int64  `json:"ats" codec:"ats"`
}

func (gi *GuildInventory) GetAssignTimesByAcID(acID string) (lootID []string, times []int64) {
	lootID = make([]string, 0)
	times = make([]int64, 0)
	for _, item := range gi.Loots {
		for _, item2 := range item.AssignMems {
			if item2.Acid == acID {
				lootID = append(lootID, item.Loot.LootId)
				times = append(times, item2.AssignTimes)
				break
			}
		}
	}
	return
}

func (gi *GuildInventory) SetAssignTimesByAcID(acID string, lootID []string, times []int64) {
	logs.Debug("Set assign info %v, %v", lootID, times)
	for i, id := range lootID {
		for index, item := range gi.Loots {
			if item.Loot.LootId == id {
				gi.Loots[index].AssignMems = append(item.AssignMems, GuildInventoryLootAssign{Acid: acID, AssignTimes: times[i]})
				break
			}
		}
	}
}

// 加到军团仓库准备区中
func (gi *GuildInventory) AddGuildInventory2Prepare(shard uint, guildUuid string,
	loots []GuildInventoryLoot, reason string, now_t int64) (isFull bool) {

	gi.UpdateGuildInventory(now_t)

	// 整理参数
	logItem := make([]logiclog.LogicInfo_ItemC, len(loots))
	_loots := make(map[string]*GuildInventoryLoot, len(loots))
	for i, l := range loots {
		if o, ok := _loots[l.LootId]; ok {
			_loots[l.LootId] = &GuildInventoryLoot{
				LootId: l.LootId,
				Count:  l.Count + o.Count,
			}
		} else {
			_loots[l.LootId] = &loots[i]
		}
		nl := _loots[l.LootId]
		if nl.Count > gamedata.GuildInventoryHeapUpLimit {
			nl.Count = gamedata.GuildInventoryHeapUpLimit
		}
		logItem[i] = logiclog.LogicInfo_ItemC{l.LootId, l.Count}
	}

	// log
	logiclog.LogAddGuildInventory(guildUuid, logItem, reason, "")

	// PrepareLoots map化
	preMap := make(map[string]*GuildInventoryLoot, len(gi.PrepareLoots))
	for i := 0; i < len(gi.PrepareLoots); i++ {
		l := &gi.PrepareLoots[i]
		preMap[l.LootId] = l
	}

	isFull = false
	countLimit := gamedata.GetCommonCfg().GetGuildBagRoom()
	lc := len(gi.Loots)
	for _, nl := range _loots {
		o, ok := preMap[nl.LootId]
		if ok {
			o.Count += nl.Count
			if o.Count > gamedata.GuildInventoryHeapUpLimit {
				o.Count = gamedata.GuildInventoryHeapUpLimit
				logs.Warn("AddGuildInventory2Prepare full, drop %s", o.LootId)
				isFull = true
			}
		} else {
			if uint32(lc+len(gi.PrepareLoots)) >= countLimit {
				logs.Warn("AddGuildInventory2Prepare full, drop %s", nl.LootId)
				isFull = true
				continue
			}
			gi.PrepareLoots = append(gi.PrepareLoots, *nl)
		}
	}
	return isFull
}

// 直接往军团仓库的可分配区里加
func (gi *GuildInventory) AddGuildInventory(shard uint, guildUuid string,
	loots []GuildInventoryLoot, reason string, now_t int64) (isFull bool) {

	gi.UpdateGuildInventory(now_t)

	if gi.Loots == nil {
		gi.Loots = []GuildInventoryLootAndMem{}
	}

	// 整理参数
	logItem := make([]logiclog.LogicInfo_ItemC, len(loots))
	_loots := make(map[string]*GuildInventoryLoot, len(loots))
	for i, l := range loots {
		if o, ok := _loots[l.LootId]; ok {
			_loots[l.LootId] = &GuildInventoryLoot{
				LootId: l.LootId,
				Count:  l.Count + o.Count,
			}
		} else {
			_loots[l.LootId] = &loots[i]
		}
		nl := _loots[l.LootId]
		if nl.Count > gamedata.GuildInventoryHeapUpLimit {
			nl.Count = gamedata.GuildInventoryHeapUpLimit
		}
		logItem[i] = logiclog.LogicInfo_ItemC{l.LootId, l.Count}
	}

	// log
	logiclog.LogAddGuildInventory(guildUuid, logItem, reason, "")

	// Loots map化
	lootsMap := make(map[string]*GuildInventoryLootAndMem, len(gi.Loots))
	for i := 0; i < len(gi.Loots); i++ {
		l := &gi.Loots[i]
		lootsMap[l.Loot.LootId] = l
	}

	isFull = false
	countLimit := gamedata.GetCommonCfg().GetGuildBagRoom()
	lc := len(gi.Loots)
	for _, nl := range _loots {
		o, ok := lootsMap[nl.LootId]
		if ok {
			o.Loot.Count += nl.Count
			if o.Loot.Count > gamedata.GuildInventoryHeapUpLimit {
				o.Loot.Count = gamedata.GuildInventoryHeapUpLimit
				logs.Warn("AddGuildInventory full, drop %s", o.Loot.LootId)
				isFull = true
			}
		} else {
			if uint32(lc) >= countLimit {
				logs.Warn("AddGuildInventory full, drop %s", nl.LootId)
				isFull = true
				continue
			}
			gi.Loots = append(gi.Loots, GuildInventoryLootAndMem{
				Loot:       *nl,
				AssignMems: make([]GuildInventoryLootAssign, 0, 4),
				ApplyAcids: make([]GuildInventoryLootApply, 0, 4),
			})
		}
	}
	return isFull
}

func (gi *GuildInventory) ApplyGuildInventory(loot string, now_t int64,
	mem *helper.AccountSimpleInfo, channelId string) int {

	gi.UpdateGuildInventory(now_t)

	// 找到要申请的物品
	lt, _ := gi.findGuildInventory(loot)
	if lt == nil {
		return Inventory_Loot_Not_Found
	}
	// 申请数上限
	if len(lt.ApplyAcids) >= int(gamedata.GetCommonCfg().GetBagApplyTime()) {
		return Inventory_Apply_Count_UpLimit
	}

	// 加记录
	lt.ApplyAcids = append(lt.ApplyAcids, GuildInventoryLootApply{
		Acid:      mem.AccountID,
		TimeStamp: time.Now().Unix(),
	})

	return Inventory_Success
}

func (gi *GuildInventory) findGuildInventory(loot string) (*GuildInventoryLootAndMem, int) {
	for i, v := range gi.Loots {
		if v.Loot.LootId == loot {
			return &gi.Loots[i], i
		}
	}
	return nil, -1
}

func (gi *GuildInventory) ExchangeGuildInventory(loot string, acid,
	guildUuid, guildName string, now_t int64, rand *rand.Rand,
	mem *helper.AccountSimpleInfo, channelId string) int {

	gi.UpdateGuildInventory(now_t)

	// 找到要申请的物品
	lt, ltIndex := gi.findGuildInventory(loot)
	if lt == nil || lt.Loot.Count <= 0 {
		return Inventory_Loot_Not_Found
	}

	// 减少数量
	lt.Loot.Count--
	// 发邮件
	_sendMailLoot(acid, guildUuid, guildName, loot, rand)
	gi.autoRefuseApplyInventory(lt, ltIndex)
	gi.checkApplyInventory(now_t)

	return Inventory_Success
}

func (gi *GuildInventory) ApproveGuildInventory(loot, acid,
	guildUuid, guildName string,
	timestamp int64, aggree bool, rand *rand.Rand, now_t int64) (
	int, []string, []int64) {

	gi.UpdateGuildInventory(now_t)

	// 找到loot
	lt, ltIndex := gi.findGuildInventory(loot)
	if lt == nil {
		resAcid, resST := gi._genApplyList(loot)
		return Inventory_Loot_Not_Found, resAcid, resST
	}
	// 找申请记录，并清除本条
	found_rec := false
	for i, ap := range lt.ApplyAcids {
		if ap.Acid == acid && ap.TimeStamp == timestamp {
			tmp := lt.ApplyAcids[:i]
			tmp1 := lt.ApplyAcids[i+1:]
			lt.ApplyAcids = make([]GuildInventoryLootApply, 0, len(tmp)+len(tmp1))
			lt.ApplyAcids = append(lt.ApplyAcids, tmp...)
			lt.ApplyAcids = append(lt.ApplyAcids, tmp1...)
			found_rec = true
			break
		}
	}
	// 新的申请列表
	resAcid, resST := gi._genApplyList(loot)
	if !found_rec {
		return Inventory_Apply_Record_Not_Found, resAcid, resST
	}
	// 分配
	if aggree { // 同意
		if lt.Loot.Count <= 0 {
			return Inventory_Loot_Not_Found, resAcid, resST
		}
		// 扣数量
		lt.Loot.Count--
		// 发邮件
		_sendMailLoot(acid, guildUuid, guildName, loot, rand)
		ok := gi.autoRefuseApplyInventory(lt, ltIndex)
		gi.checkApplyInventory(now_t)
		if ok {
			return Inventory_Success, []string{}, []int64{}
		}
	} else { // 拒绝
		cfg := gamedata.GetGuildLostInventoryCfg(lt.Loot.LootId)
		items := make(map[string]uint32, 1)
		items[cfg.GetItemPrice()] = cfg.GetPrice()
		mail_sender.BatchSendMail2Account(acid,
			timail.Mail_Send_By_Guild_Inventory,
			mail_sender.IDS_MAIL_GUILD_GVEBOSSBAG_REFUSE_TITLE,
			[]string{}, items, "GuildInventoryRefuse",false)
	}
	return Inventory_Success, resAcid, resST
}

func (gi *GuildInventory) GetGuildInventoryApplyList(loot string, now_t int64) (
	[]string, []int64) {

	gi.UpdateGuildInventory(now_t)

	resAcid, resST := gi._genApplyList(loot)
	return resAcid, resST
}

func (gi *GuildInventory) OnDelMem(acid string) {
	for i, l := range gi.Loots {
		_ApplyAcids := make([]GuildInventoryLootApply, 0, len(l.ApplyAcids))
		for _, aa := range l.ApplyAcids {
			if aa.Acid != acid {
				_ApplyAcids = append(_ApplyAcids, aa)
			} else {
				// 发送拒绝邮件
				cfg := gamedata.GetGuildLostInventoryCfg(l.Loot.LootId)
				items := make(map[string]uint32, 1)
				items[cfg.GetItemPrice()] = cfg.GetPrice()
				mail_sender.BatchSendMail2Account(acid,
					timail.Mail_Send_By_Guild_Inventory,
					mail_sender.IDS_MAIL_GUILD_GVEBOSSBAG_REFUSE_TITLE,
					[]string{}, items, "GuildInventoryRefuse" , false)
			}
		}
		gi.Loots[i].ApplyAcids = _ApplyAcids
	}
}

func (gi *GuildInventory) UpdateGuildInventory(now_t int64) bool {
	if now_t > gi.NextRefTime || now_t > gi.NextResetTime {
		gi._mergeItem(now_t)
		gi._clearApply(now_t)
	}

	if gi.LostActive && len(gi.Loots) == 0 && now_t >= gi.DisappearTime {
		gi.LostActive = false
	}
	return true
}

func (gi *GuildInventory) _mergeItem(now_t int64) {
	if now_t >= gi.NextRefTime {
		gi.NextRefTime = util.GetNextDailyTime(util.DailyBeginUnixByStartTime(now_t,
			gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypGVGGuildGiftGet)), now_t)
		if gi.PrepareLoots == nil {
			gi.PrepareLoots = []GuildInventoryLoot{}
		}
		if gi.Loots == nil {
			gi.Loots = []GuildInventoryLootAndMem{}
		}
		// prepare map 化
		preMap := make(map[string]*GuildInventoryLoot, len(gi.PrepareLoots))
		for i := 0; i < len(gi.PrepareLoots); i++ {
			loot := &gi.PrepareLoots[i]
			preMap[loot.LootId] = loot
		}
		// 合并物品并清次数
		_loots := make([]GuildInventoryLootAndMem, 0, 20)
		for i := 0; i < len(gi.Loots); i++ {
			lt := &gi.Loots[i]
			if preLoot, ok := preMap[lt.Loot.LootId]; ok {
				lt.Loot.Count += preLoot.Count
				if lt.Loot.Count > gamedata.GuildInventoryHeapUpLimit {
					lt.Loot.Count = gamedata.GuildInventoryHeapUpLimit
					logs.Warn("UpdateGuildInventory full, drop %s", lt.Loot.LootId)
				}
			}
			// 合并后,或之前没用完的,进入loots,并清次数
			if lt.Loot.Count > 0 {
				lt.AssignMems = make([]GuildInventoryLootAssign, 0, 4)
				_loots = append(_loots, *lt)
				delete(preMap, lt.Loot.LootId)
			}
		}
		for _, v := range preMap {
			_loots = append(_loots, GuildInventoryLootAndMem{
				Loot:       *v,
				AssignMems: make([]GuildInventoryLootAssign, 0, 4),
				ApplyAcids: make([]GuildInventoryLootApply, 0, 4),
			})
		}
		gi.Loots = _loots
		gi.PrepareLoots = make([]GuildInventoryLoot, 0, 20)
	}
}

func (gi *GuildInventory) _clearApply(now_t int64) {
	if now_t >= gi.NextResetTime {
		gi.NextResetTime = util.GetNextDailyTime(
			gamedata.GetGuildBagRefuseBeginSec(now_t), now_t)
		if gi.PrepareLoots == nil {
			gi.PrepareLoots = []GuildInventoryLoot{}
		}
		if gi.Loots == nil {
			gi.Loots = []GuildInventoryLootAndMem{}
		}

		xp_map := make(map[string]map[string]uint32, 10)

		for i := 0; i < len(gi.Loots); i++ {
			lt := &gi.Loots[i]
			// 发自动回退邮件
			cfg := gamedata.GetGuildLostInventoryCfg(lt.Loot.LootId)
			for _, ap := range lt.ApplyAcids {
				value, ok := xp_map[ap.Acid]
				if ok {
					own, ok := value[cfg.GetItemPrice()]
					if ok {
						value[cfg.GetItemPrice()] = own + cfg.GetPrice()
					} else {
						value[cfg.GetItemPrice()] = cfg.GetPrice()
					}
				} else {
					xp_map[ap.Acid] = make(map[string]uint32, 1)
					xp_map[ap.Acid][cfg.GetItemPrice()] = cfg.GetPrice()
				}
			}
			// 清理记录
			lt.ApplyAcids = make([]GuildInventoryLootApply, 0, 4)
		}
		for k, v := range xp_map {
			mail_sender.BatchSendMail2Account(k,
				timail.Mail_Send_By_Guild_Inventory,
				mail_sender.IDS_MAIL_GUILD_GVEBOSSBAG_AUTO_TITLE,
				[]string{}, v, "GuildInventoryClear",false )
		}
	}
}

func _sendMailLoot(acid, guildUuid, guildName, loot string, rand *rand.Rand) {
	// 发邮件
	pds := gamedata.NewPriceDataSet(5)
	cfg := gamedata.GetGuildLostInventoryCfg(loot)
	for _, cfgLoot := range cfg.GetLoot_Table() {
		if cfgLoot.GetLootGroupID() == "" {
			continue
		}
		for i := 0; i < int(cfgLoot.GetLootTime()); i++ {
			gives, err := gamedata.LootTemplateRand(rand, cfgLoot.GetLootGroupID())
			if err == nil && gives.IsNotEmpty() {
				pds.AppendData(gives)
			} else {
				logs.Error("AssignGuildInventoryItem LootItemGroupRand err or empty %s %s %s",
					loot, cfgLoot.GetLootGroupID(), err.Error())
			}
		}
	}
	pd := pds.Mk2One()
	if pd.IsNotEmpty() {
		mail_sender.SendGuildInventory(acid, guildName, loot,
			pd.Item2Client, pd.Count2Client)
	} else {
		logs.Error("AssignGuildInventoryItem LootItemGroupRand empty %s", loot)
	}
	// log
	logiclog.LogAssignGuildInventory(guildUuid, loot, acid, pd.Item2Client, pd.Count2Client, "")
}

func (gi *GuildInventory) _genApplyList(loot string) (
	resAcid []string, resST []int64) {

	resAcid = make([]string, 0, 10)
	resST = make([]int64, 0, 10)
	for _, v := range gi.Loots {
		if v.Loot.LootId == loot {
			for _, a := range v.ApplyAcids {
				resAcid = append(resAcid, a.Acid)
				resST = append(resST, a.TimeStamp)
			}
			break
		}
	}
	return
}

const (
	Inventory_Success = iota
	Inventory_ItemCount
	Inventory_Times
	Inventory_Position
	Inventory_Leave
	Inventory_Act_Leave
	Inventory_GB_Not_Enough
	Inventory_Loot_Not_Found
	Inventory_Apply_Record_Not_Found
	Inventory_Apply_Count_UpLimit
)

func (gi *GuildInventory) AssignGuildInventoryItem(shard uint,
	guildUuid, lootId, acid, guildName string,
	rand *rand.Rand, now_t int64) int {

	gi.UpdateGuildInventory(now_t)

	cfg := gamedata.GetGuildInventoryCfg(lootId)
	for i := 0; i < len(gi.Loots); i++ {
		loot := &gi.Loots[i]
		if loot.Loot.LootId == lootId {
			if loot.Loot.Count <= 0 {
				return Inventory_ItemCount
			}
			// 扣次数
			bCosted := false
			for j := 0; j < len(loot.AssignMems); j++ {
				mem := &loot.AssignMems[j]
				if mem.Acid == acid {
					if uint32(mem.AssignTimes) >= cfg.GetLimitTime() {
						return Inventory_Times
					}
					mem.AssignTimes++
					bCosted = true
					break
				}
			}
			if !bCosted {
				loot.AssignMems = append(loot.AssignMems, GuildInventoryLootAssign{
					Acid:        acid,
					AssignTimes: 1,
				})
			}
			// 扣数量
			loot.Loot.Count--
			// 发邮件
			pds := gamedata.NewPriceDataSet(5)
			cfg := gamedata.GetGuildInventoryCfg(loot.Loot.LootId)
			for _, cfgLoot := range cfg.GetLoot_Table() {
				if cfgLoot.GetLootGroupID() == "" {
					continue
				}
				for i := 0; i < int(cfgLoot.GetLootTime()); i++ {
					gives, err := gamedata.LootTemplateRand(rand, cfgLoot.GetLootGroupID())
					if err == nil && gives.IsNotEmpty() {
						pds.AppendData(gives)
					} else {
						logs.Error("AssignGuildInventoryItem LootItemGroupRand err or empty %s %s %s",
							loot.Loot.LootId, cfgLoot.GetLootGroupID(), err.Error())
					}
				}
			}
			pd := pds.Mk2One()
			if pd.IsNotEmpty() {
				mail_sender.SendGuildInventory(acid, guildName, loot.Loot.LootId,
					pd.Item2Client, pd.Count2Client)
			} else {
				logs.Error("AssignGuildInventoryItem LootItemGroupRand empty %s", loot.Loot.LootId)
			}
			// log
			logiclog.LogAssignGuildInventory(guildUuid, lootId, acid, pd.Item2Client, pd.Count2Client, "")
			return Inventory_Success
		}
	}
	return Inventory_ItemCount
}

// 物品数量为0, 自动处理剩余的申请为拒绝
func (gi *GuildInventory) autoRefuseApplyInventory(lt *GuildInventoryLootAndMem, ltIndex int) bool {
	if lt.Loot.Count <= 0 {
		lootCfg := gamedata.GetGuildLostInventoryCfg(lt.Loot.LootId)
		mergedApply := gi.mergeSameApply(lt)
		for acid, count := range mergedApply {
			gi.onRefuseApply(acid, lootCfg, count)
		}
		gi.deleteInventoryItem(ltIndex)
		return true
	}
	return false
}

func (gi *GuildInventory) checkApplyInventory(nowTime int64) {
	if gi.LostActive && len(gi.Loots) == 0 && gi.DisappearTime == 0 {
		gi.DisappearTime = util.GetNextDailyTime(gamedata.GetCommonDayBeginSec(nowTime), nowTime)
	}
}

// 合并每个人的申请
func (gi *GuildInventory) mergeSameApply(lt *GuildInventoryLootAndMem) map[string]int {
	retApply := make(map[string]int, len(lt.ApplyAcids)) // <acid, 申请数量count>
	for _, apply := range lt.ApplyAcids {
		if count, ok := retApply[apply.Acid]; ok {
			retApply[apply.Acid] = count + 1
		} else {
			retApply[apply.Acid] = 1
		}
	}
	return retApply
}

// 删除这个物品
func (gi *GuildInventory) deleteInventoryItem(ltIndex int) {
	preLoots := gi.Loots[:ltIndex]
	postLoots := gi.Loots[ltIndex+1:]
	gi.Loots = make([]GuildInventoryLootAndMem, 0, len(preLoots)+len(postLoots))
	gi.Loots = append(gi.Loots, preLoots...)
	gi.Loots = append(gi.Loots, postLoots...)
}

func (gi *GuildInventory) onRefuseApply(acid string, lootCfg *ProtobufGen.LOSTGOODSHOP, count int) {
	items := make(map[string]uint32, 1)
	items[lootCfg.GetItemPrice()] = lootCfg.GetPrice() * uint32(count)
	mail_sender.BatchSendMail2Account(acid,
		timail.Mail_Send_By_Guild_Inventory,
		mail_sender.IDS_MAIL_GUILD_GVEBOSSBAG_REFUSE_TITLE,
		[]string{}, items, "GuildInventoryRefuse" , false)
}
