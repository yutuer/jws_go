package update

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func init() {
	guild.VerAdd(6, V6ToV7)
}

// v240 -> v250
func V6ToV7(FromVersion int64, gi *guild.GuildInfo) error {
	if FromVersion != gi.Ver {
		return fmt.Errorf("v6Tov7 err FromVersion %d Profile.Ver %d",
			FromVersion, gi.Ver)
	}

	err := v7GuildInventoryChange(gi)
	if nil != err {
		return err
	}

	logs.Debug("v6Tov7 guild:[%s] ver[%d]", gi.Base.GuildUUID, gi.Ver)
	return nil
}

// v6tov7 Story:T20264 Task:T20504 清除军团仓库,取消军团成员的申请并折算军魂补偿
func v7GuildInventoryChange(gi *guild.GuildInfo) error {
	if uutil.IsHMTVer() {
		logs.Info("V6ToV7 v7GuildInventoryChange ignore, because HMT")
		return nil
	}

	if err := v7GuildInventoryApplyRet(gi); nil != err {
		return err
	}
	if err := v7ClearGuildInventory(gi); nil != err {
		return err
	}
	return nil
}

func v7GuildInventoryApplyRet(gi *guild.GuildInfo) error {
	logs.Debug("v7GuildInventoryApplyRet:%s Start", gi.Base.GuildUUID)

	uc := make(map[string]uint32, 0)
	for _, loot := range gi.Inventory.Loots {
		item := gamedata.GetGuildInventoryCfg(loot.Loot.LootId)
		if nil != item {
			cn := item.GetPrice()
			for _, apply := range loot.ApplyAcids {
				_, exist := uc[apply.Acid]
				if exist {
					uc[apply.Acid] += cn
				} else {
					uc[apply.Acid] = cn
				}
			}
		}
	}

	logs.Debug("v7GuildInventoryApplyRet:%s, collect apply infos:%v", gi.Base.GuildUUID, uc)

	for acid, bg_count := range uc {
		t := bg_count / 480
		if 0 != bg_count%480 {
			t += 1
		}
		logs.Debug("v7GuildInventoryApplyRet:%s, accout[%s] get %d pack", gi.Base.GuildUUID, acid, t)
		mail_sender.BatchSendMail2Account(acid,
			timail.Mail_send_By_Guild,
			mail_sender.IDS_MAIL_VI_GB_INVENTORY_TITLE,
			[]string{
				fmt.Sprintf("%d", bg_count),
				fmt.Sprintf("%d", t),
			},
			map[string]uint32{
				gamedata.VI_DC:    2000 * uint32(t),
				"VI_HERO_ZFB":     25 * uint32(t),
				"MAT_StarStone_5": 1 * uint32(t),
			},
			"V6toV7: recharge user's GuildBossCoin in GuildInventory to reward", false)
	}

	logs.Debug("v7GuildInventoryApplyRet:%s End", gi.Base.GuildUUID)
	return nil
}

func v7ClearGuildInventory(gi *guild.GuildInfo) error {
	logs.Debug("v7ClearGuildInventory:%s", gi.Base.GuildUUID)
	gi.Inventory.PrepareLoots = []guild_info.GuildInventoryLoot{}
	gi.Inventory.Loots = []guild_info.GuildInventoryLootAndMem{}
	return nil
}
