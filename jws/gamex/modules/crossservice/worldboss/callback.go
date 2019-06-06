package worldboss

import (
	"fmt"

	"vcs.taiyouxi.net/jws/crossservice/module"
	cs_worldboss "vcs.taiyouxi.net/jws/crossservice/module/worldboss"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func init() {
	crossservice.RegCallbackHandle(cs_worldboss.ModuleID, cs_worldboss.CallbackDamageRankID, CallbackDamageRankDoReward)
	crossservice.RegCallbackHandle(cs_worldboss.ModuleID, cs_worldboss.CallbackMarqueeID, CallbackMarquee)
}

//CallbackDamageRankDoReward ..
func CallbackDamageRankDoReward(p module.Param) {
	param := p.(*cs_worldboss.ParamDamageRank)
	logs.Warn("[WorldBoss] CallbackDamageRankDoReward Param %+v", param)
	// rank rewards
	for _, item := range param.Rank {
		rewards := gamedata.GetWBRankRewards(item.Pos)
		if len(rewards) > 0 {
			items := make(map[string]uint32, 0)
			for _, item := range rewards {
				items[item.GetRankLootID()] = item.GetRankLootNumber()
			}
			mail_sender.BatchSendMail2Account(item.Acid, timail.Mail_send_By_Common,
				mail_sender.IDS_MAIL_WB_RANKREWARD_TITLE,
				[]string{fmt.Sprintf("%d", item.Pos)}, items, "WorldBossRankRewards", false)
		}
	}

	// boss rewards
	rewards := gamedata.GetWBBossCfg(param.BossLevel).GetLoot_Table()
	if len(rewards) > 0 {
		items := make(map[string]uint32, 0)
		for _, item := range rewards {
			items[item.GetItemID()] = item.GetItemNum()
		}
		for _, item := range param.Rank {
			if item.Pos <= gamedata.GetKillBossRewardRankLimit() {
				mail_sender.BatchSendMail2Account(item.Acid, timail.Mail_send_By_Common,
					mail_sender.IDS_MAIL_WB_KILLBOSSREWARD_TITLE,
					[]string{fmt.Sprintf("%d", param.BossLevel)}, items, "KillWorldBossRewards", false)
			}

		}
	}

}

//CallbackMarquee ..
func CallbackMarquee(p module.Param) {
	param := p.(*cs_worldboss.ParamMarquee)
	logs.Warn("[WorldBoss] CallbackMarquee Param %+v", param)
	switch param.MsgType {
	case cs_worldboss.MarqueeTypeChampion:
		sysnotice.NewSysRollNotice(fmt.Sprintf("%d:%d", game.Cfg.Gid, param.Sid), int32(gamedata.IDS_WORLDBOSS_CHAMPINE)).
			AddParam(sysnotice.ParamType_RollName, param.ChampionName).
			Send()
	}
}
