package guild

import (
	"math/rand"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/modules/rank/award"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func guildRankAward() {
	go func() {
		var addon timail.MailAddonCounter
		rd := rand.New(rand.NewSource(time.Now().Unix()))
		for rankGuild := range award.GetGuildAwardChan() {
			logs.Trace("SeverOpenGuildRank award %v", rankGuild)
			rank := rankGuild.Pos
			uuid := rankGuild.Uuid
			rCfg, lCfg := gamedata.GetGuildRankAwardCfg(rank)
			if rCfg == nil || lCfg == nil {
				continue
			}
			sid, _ := guild_info.GetShardIdByGuild(uuid)
			info, code := GetModule(sid).GetGuildInfo(uuid)
			if code.HasError() {
				logs.Error("SeverOpenRank AwardGuild GetGuildInfo err")
			} else {
				var leader string
				time_now := game.GetNowTimeByOpenServer(sid)
				for i := 0; i < len(info.Members) && i < info.Base.MemNum; i++ {
					mem := info.Members[i]
					if mem.GuildPosition == gamedata.Guild_Pos_Chief {
						leader = mem.AccountID
					}

					ritems, rcounts := guildMemReward(rank, uuid, rCfg, rd)
					mail_sender.BatchSend7DayGuildReward(sid, mem.AccountID, time_now,
						rank, ritems, rcounts, addon.Get())
					logs.Trace("SeverOpenGuildRank award mail acid %s %d", mem.AccountID, rank)

				}
				// 会长奖
				litems, lcounts := guildLeaderReward(rank, uuid, lCfg, rd)
				mail_sender.BatchSend7DayGuildLeaderReward(sid, leader, time_now,
					rank, litems, lcounts, addon.Get())
				logs.Trace("SeverOpenGuildRank award mail leader %s %d", leader, rank)
			}
		}
	}()
}

func guildMemReward(rank int, uuid string, rCfg *ProtobufGen.GUILDRANK, rd *rand.Rand) ([]string, []uint32) {
	rs := rCfg.GetGuildAwardItems_Template()
	itemMap := make(map[string]uint32, len(rs))
	for _, a := range rs {
		_addItemMap(itemMap, a.GetItemID(), a.GetAmount())
	}
	if rCfg.GetLootData1ID() != "" {
		for i := 0; i < int(rCfg.GetLootData1Amount()); i++ {
			gives, err := gamedata.LootItemGroupRand(rd, rCfg.GetLootData1ID())
			if err != nil {
				logs.Error("guildRankAward LootItemGroupRand %d %s err %s",
					rank, uuid, err.Error())
			} else {
				for idx, itemID := range gives.Item2Client {
					_addItemMap(itemMap, itemID, gives.Count2Client[idx])
				}
			}
		}
	}
	if rCfg.GetLootData2ID() != "" {
		for i := 0; i < int(rCfg.GetLootData2Amount()); i++ {
			gives, err := gamedata.LootItemGroupRand(rd, rCfg.GetLootData2ID())
			if err != nil {
				logs.Error("guildRankAward LootItemGroupRand %d %s err %s",
					rank, uuid, err.Error())
			} else {
				for idx, itemID := range gives.Item2Client {
					_addItemMap(itemMap, itemID, gives.Count2Client[idx])
				}
			}
		}
	}
	ritems := make([]string, 0, len(itemMap))
	rcounts := make([]uint32, 0, len(itemMap))
	for k, v := range itemMap {
		ritems = append(ritems, k)
		rcounts = append(rcounts, v)
	}
	return ritems, rcounts
}

func guildLeaderReward(rank int, uuid string, lCfg *ProtobufGen.GUILDLEAD, rd *rand.Rand) ([]string, []uint32) {
	ls := lCfg.GetLeadAwardItems_Template()
	itemMap := make(map[string]uint32, len(ls))
	for _, a := range ls {
		_addItemMap(itemMap, a.GetItemID(), a.GetAmount())
	}
	if lCfg.GetLootData1ID() != "" {
		for i := 0; i < int(lCfg.GetLootData1Amount()); i++ {
			gives, err := gamedata.LootItemGroupRand(rd, lCfg.GetLootData1ID())
			if err != nil {
				logs.Error("guildRankAward leader LootItemGroupRand %d %s err %s", rank, uuid, err.Error())
			} else {
				for idx, itemID := range gives.Item2Client {
					_addItemMap(itemMap, itemID, gives.Count2Client[idx])
				}
			}
		}
	}
	if lCfg.GetLootData2ID() != "" {
		for i := 0; i < int(lCfg.GetLootData2Amount()); i++ {
			gives, err := gamedata.LootItemGroupRand(rd, lCfg.GetLootData2ID())
			if err != nil {
				logs.Error("guildRankAward leader LootItemGroupRand %d %s err %s", rank, uuid, err.Error())
			} else {
				for idx, itemID := range gives.Item2Client {
					_addItemMap(itemMap, itemID, gives.Count2Client[idx])
				}
			}
		}
	}
	litems := make([]string, 0, len(ls))
	lcounts := make([]uint32, 0, len(ls))
	for k, v := range itemMap {
		litems = append(litems, k)
		lcounts = append(lcounts, v)
	}
	return litems, lcounts
}

func _addItemMap(m map[string]uint32, item string, count uint32) {
	c, ok := m[item]
	if ok {
		m[item] = c + count
	} else {
		m[item] = count
	}
}
