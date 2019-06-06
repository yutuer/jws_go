package update

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func V5toV6(FromVersion int64, justLoad bool, acc *account.Account) error {
	if FromVersion != acc.Profile.Ver {
		return fmt.Errorf("V5toV6 err FromVersion %d Profile.Ver %d",
			FromVersion, acc.Profile.Ver)
	}

	if acc.GuildProfile.GuildUUID != "" {
		res, code := guild.GetModule(acc.AccountID.ShardId).GetGuildInfo(
			acc.GuildProfile.GuildUUID)
		if !code.HasError() {
			for i := 0; i < res.GuildInfoBase.Base.MemNum; i++ {
				mem := res.GuildInfoBase.Members[i]
				if mem.AccountID == acc.AccountID.String() {
					acc.Profile.GetSC().Currency[gamedata.SC_GB] = mem.GuildBossCoin
					acc.Profile.GetSC().Currency[gamedata.SC_GuildSp] = mem.GuildSp
					break
				}
			}

		}
	}

	acc.Profile.Ver = FromVersion + 1
	logs.Debug("V5toV6 %s %d", acc.AccountID.String(), acc.Profile.Ver)
	return nil
}

// v240 -> v250
func V6toV7(FromVersion int64, justLoad bool, acc *account.Account) error {
	if FromVersion != acc.Profile.Ver {
		return fmt.Errorf("V6toV7 err FromVersion %d Profile.Ver %d",
			FromVersion, acc.Profile.Ver)
	}

	HeroCompanionV6Tov7(acc)

	// ADD By qiaozhu
	// 军魂处理
	clearGuildBossCoin(justLoad, acc)
	// ADD By qiaozhu End

	acc.Profile.Ver = FromVersion + 1
	logs.Debug("V6toV7 %s %d", acc.AccountID.String(), acc.Profile.Ver)
	return nil
}

// v240 -> 250
// 新增字段ID 用来标志聚义唯一ID 支持策划修改聚义关系
// 配置表新增字段OldCompanionId 支持老的存档结构转换到新的结构
// 新增字段LevelId 用来聚义升级后更改ID
func HeroCompanionV6Tov7(acc *account.Account) {
	companionList := acc.Profile.Hero.HeroCompanionInfos[:]
	// 每个武将
	for i := range companionList {
		companion := &companionList[i]
		companion.NewCompanions = make([]account.NewCompanion, companion.CompanionNum)
		// 每个情缘
		for j := 0; j < companion.CompanionNum; j++ {
			companion.NewCompanions[j] = newHeroCompanion(i, companion.Companions[j], companion.EvolveLevel)
		}
		companion.Companions = nil
	}
	acc.Profile.GetData().SetNeedCheckMaxGS()
}

func newHeroCompanion(heroId int, companion account.Companion, evolveLevel int) account.NewCompanion {
	level := evolveLevel
	if companion.Active {
		level++
	}
	defaultActive := true
	if level == 0 {
		level = 1
		defaultActive = false
	}
	cfg := gamedata.GetCompanionActiveConfig(heroId, companion.CompanionId, level)
	newCompanion := account.NewCompanion{
		Id:     int(cfg.Config.GetUniqueID()),
		Active: defaultActive,
	}
	newCompanion.UpdateLevelAndCompanion()
	return newCompanion
}

// v240 -> 250 军团仓库改版
func clearGuildBossCoin(justLoad bool, acc *account.Account) error {
	if uutil.IsHMTVer() {
		logs.Info("V6ToV7 clearGuildBossCoin ignore, because HMT")
		return nil
	}

	bg_count := acc.Profile.GetSC().Currency[gamedata.SC_GB]
	logs.Debug("V6toV7: Clean Guild Boss Coin before %d", acc.Profile.GetSC().Currency[gamedata.SC_GB])
	acc.Profile.GetSC().Currency[gamedata.SC_GB] = 0
	if bg_count > 0 {
		t := bg_count / 480
		if 0 != bg_count%480 {
			t += 1
		}
		if false == justLoad {
			logs.Debug("V6toV7: Clean Guild Boss Coin accout[%s] get %d pack", acc.AccountID.String(), t)
			mail_sender.BatchSendMail2Account(acc.AccountID.String(),
				timail.Mail_send_By_Guild,
				mail_sender.IDS_MAIL_VI_GB_DEL_TITLE,
				[]string{
					fmt.Sprintf("%d", bg_count),
					fmt.Sprintf("%d", t),
				},
				map[string]uint32{
					gamedata.VI_DC:    2000 * uint32(t),
					"VI_HERO_ZFB":     25 * uint32(t),
					"MAT_StarStone_5": 1 * uint32(t),
				},
				"V6toV7: recharge user's GuildBossCoin to reward", false)
		}
	}
	logs.Debug("V6toV7: Clean Guild Boss Coin after %d", acc.Profile.GetSC().Currency[gamedata.SC_GB])
	return nil
}
