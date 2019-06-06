package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
)

func condition_impl_59() {
	// 59.军团捐献科技点次数（接任务后）次数
	regConditionImpFunc(COND_TYP_Add_Guild_Science_Point,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	// 60.攻打N次军团BOSS，无论成败(接取任务后)次数
	regConditionImpFunc(COND_TYP_Guild_Boss,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	// 61.在兵临城下中完成（取胜）N次战斗, 次数
	regConditionImpFunc(COND_TYP_GateEnemy_Finish,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	// 62.参与N次吃包子小游戏 次数
	regConditionImpFunc(COND_TYP_EatBaozi,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)

	// 63.单次吃包子的数目
	regConditionImpFunc(COND_TYP_EatBaoziCount,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return int(p.Profile.GetEatBaozi().GetMaxEatBaoziCount()), int(p1)
		})

	// 64.完成了烽火辽源多少小关
	regConditionImpFunc(COND_TYP_FengHuoSubLevelCount,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	// 65.历史总计参与1v1的次数
	regConditionImpFunc(COND_TYP_1v1_Times,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return p.Profile.GetData().Times_1V1, int(p1)
		})

	// 66.历史总计参与3v3的次数
	regConditionImpFunc(COND_TYP_3v3_Times,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return p.Profile.GetData().Times_3V3, int(p1)
		})
	//67.每日参与远征的次数
	regConditionImpFunc(COND_TYP_Expedition,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)

	//68.M个主将的幻甲等级达到N
	regConditionImpFunc(COND_TYP_Swing_Lvl_Together,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			swing := p.Profile.Hero
			count := 0
			for i := 0; i < len(swing.HeroSwings); i++ {
				g := &(swing.HeroSwings[i])
				if g.Lv >= int(p2) {
					count += 1
				}
			}
			return count, int(p1)
		})
	//69.有M个主将的星级达到N
	regConditionImpFunc(COND_TYP_Hero_Star_Together,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			hero := p.Profile.Hero
			count := 0
			for i := 0; i < len(hero.HeroStarLevel); i++ {
				g := hero.HeroStarLevel[i]
				if int64(g) >= p2 {
					count += 1
				}
			}
			return count, int(p1)
		})
	// 70.有M个主将的等级达到N
	regConditionImpFunc(COND_TYP_Hero_Lvl_Together,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			hero := p.Profile.Hero
			count := 0
			for i := 0; i < len(hero.HeroLevel); i++ {
				g := hero.HeroLevel[i]
				if int64(g) >= p2 {
					count += 1
				}
			}
			return count, int(p1)
		})
	// 71.M个主将的幻甲星级达到N
	regConditionImpFunc(COND_TYP_Swing_Star_Together,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			swing := p.Profile.Hero
			count := 0
			for i := 0; i < len(swing.HeroSwings); i++ {
				g := &(swing.HeroSwings[i])
				if g.StarLv >= int(p2) {
					count += 1
				}
			}
			return count, int(p1)
		})
	//72.日常任务-参加1次我要名将
	regConditionImpFunc(COND_TYP_IWant_Hero,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)
	//73.日常任务参见试炼之地的次数
	regConditionImpFunc(COND_TYP_Try_Test,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)
	// 74.GVG军团战连斩次数
	regConditionImpFunc(COND_TYP_GVG_WINSTREAK,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return 1, 1
		})
	// 75.GVG军团战某个军团占领所有城池
	regConditionImpFunc(COND_TYP_GVG_ONEWORLD,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			isLeader := gvg.GetModule(p.AccountID.ShardId).IsWorldLeader(p.AccountID.String())
			if isLeader {
				return 1, 1
			}
			return 0, 1
		})
	// 76.激活N个主将
	regConditionImpFunc(COND_TYP_HERO_ACTIVE,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return p.Profile.GetCorp().HasAvatarHasUnlok(), int(p1)
		})

	// 77.军团募捐次数
	regConditionImpFunc(COND_TYP_GUILD_COLLECTION,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)

	// 78.击杀FestivalBoss的数目
	regConditionImpFunc(COND_TYP_GUILD_FESTIVALBOSS,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return int(p.Profile.GetFestivalBossInfo().GetBossKillTime()), int(p2)
		})
	// 80.参与出奇制胜的次数
	regConditionImpFunc(COND_TYP_HERODIFF_FINISH,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)

	// 83.N主将激活神兵，参数为主将个数
	regConditionImpFunc(COND_TYP_ACT_EXCLUSIVE_WEAPON,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			heros := p.Profile.Hero
			count := 0
			for _, weapon := range heros.HeroExclusiveWeapon {
				if weapon.IsActive {
					count++
				}
			}
			return count, int(p1)
		})
	// 84.N个主将的神兵达到X品质, 参数1：主将个数 参数2：神兵品质
	regConditionImpFunc(COND_TYP_EVOLVE_EXCLUSIVE_WEAPON,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			heros := p.Profile.Hero
			count := 0
			for _, weapon := range heros.HeroExclusiveWeapon {
				if weapon.IsActive && weapon.Quality >= int(p2) {
					count++
				}
			}
			return count, int(p1)
		})
	// 82.军团膜拜次数
	regConditionImpFunc(COND_TYP_GUILD_WORSHIP,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)

	// 85.FaceBook分享多少次
	regConditionImpFunc(COND_TYP_FB_Share,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)
	// 86.给好用送一次包子礼物
	regConditionImpFunc(COND_TYP_Give_FriendGift,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)

	// 87.当天打一次世界boss
	regConditionImpFunc(COND_TYP_WorldBoss,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)
	// 88.兑换商店获得某一道具数量,
	regConditionImpFunc(COND_TYP_ExchangeShop,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return 0, 1
		})

}
