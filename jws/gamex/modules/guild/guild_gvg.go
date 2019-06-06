package guild

import (
	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func (r *GuildModule) UpdateGVGScore(guildUuId string, score []int64, acId []string) {
	r.guildCommandExecAsyn(guildCommand{
		Type: Command_UpdateGVGScore,
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUuId,
		},
		ParamInts: score,
		ParamStrs: acId,
	})
	return
}

func (r *GuildModule) GVGBalanceForPlayer(guildUUID string, id int) {

	c := guildCommand{
		Type:      Command_GVGBalanceForPlayer,
		ParamInts: []int64{int64(id)},
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	}
	r.guildCommandExecAsyn(c)
}

func (r *GuildModule) GVGBalanceForInventory(guildUUID string, id int, rank int) {

	c := guildCommand{
		Type:      Command_GVGBalanceForInventory,
		ParamInts: []int64{int64(id), int64(rank)},
		BaseInfo: guild_info.GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
	}
	r.guildCommandExecAsyn(c)
}

func (g *GuildWorker) gvgBalanceForPlayer(c *guildCommand) {
	guildInfo := g.guild
	logs.Warn("guild gvg balance for player %v", c)
	cityID := c.ParamInts[0]
	// 向每一个成员发奖
	items := make(map[string]uint32, 5)
	memberDayGift := gamedata.GetGVGActivityDailyGift(uint32(cityID))
	if memberDayGift != nil {
		for _, loot := range memberDayGift.GetLoot_Table() {
			items[loot.GetDailyItemID()] = loot.GetDailyItemNum()
			logs.Debug("GVG Member gift: %v", *loot)
		}
		if len(items) > 0 {
			for i := 0; i < guildInfo.Base.MemNum; i++ {
				mail_sender.BatchSendMail2Account(
					guildInfo.Members[i].AccountID, timail.Mail_Send_By_GVG,
					mail_sender.IDS_MAIL_GVG_DAILY_GIFT_TITLE,
					[]string{fmt.Sprintf("%d", cityID)}, items,
					"GVGActivityDayBalance", true)
			}
		}
	}

	if err := g.saveGuild(guildInfo); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	c.resChan <- guildCommandRes{
		guildInfo: *guildInfo,
	}
	return
}

func (g *GuildWorker) gvgBalanceForInventory(c *guildCommand) {
	guildInfo := g.guild
	logs.Warn("guild gvg balance for inventory %v", c)
	cityID := c.ParamInts[0]
	rank := c.ParamInts[1]

	// 向工会发奖,放入仓库
	guildGift := gamedata.GetGVGActivityGuildGift(uint32(cityID), rank)
	if guildGift.Ids != nil && guildGift.Counts != nil {
		guildItems := make([]guild_info.GuildInventoryLoot, 0, 5)
		if len(guildGift.Counts) == len(guildGift.Ids) {
			for i := 0; i < len(guildGift.Counts); i++ {
				guildItems = append(guildItems, guild_info.GuildInventoryLoot{
					LootId: guildGift.Ids[i],
					Count:  guildGift.Counts[i],
				})
			}
			logs.Debug("GVG Inventory gift: %v", guildItems)
			guildInfo.Inventory.AddGuildInventory(g.m.sid, c.BaseInfo.GuildUUID, guildItems, "GVG Guild Gift", time.Now().Unix())
		}
	}
	if err := g.saveGuild(guildInfo); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}

	c.resChan <- guildCommandRes{
		guildInfo: *guildInfo,
	}
	return
}

func (g *GuildWorker) updateGVGScore(c *guildCommand) {
	guildInfo := g.guild
	logs.Trace("update gvg scorefor player %v", c)
	for i := 0; i < guildInfo.Base.MemNum; i++ {
		guildInfo.Members[i].GVGScore = 0
	}
	for i := 0; i < guildInfo.Base.MemNum; i++ {
		for j := 0; j < len(c.ParamStrs); j++ {
			if guildInfo.Members[i].AccountID == c.ParamStrs[j] {
				guildInfo.Members[i].GVGScore = int(c.ParamInts[j])
				break
			}
		}
	}
	if err := g.saveGuild(guildInfo); err != nil {
		c.resChan <- genErrRes(Err_DB)
		return
	}
	c.resChan <- guildCommandRes{
		guildInfo: *guildInfo,
	}
}
