package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type condition_initer func(c *gamedata.Condition)
type condition_update func(c *gamedata.Condition, p1, p2 int64, p3, p4 string)
type condition_get_progress func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int)

// 通过这两个map区分不同类型条件行为的不同
var (
	condition_initer_map       map[uint32]condition_initer
	condition_updater_map      map[uint32]condition_update
	condition_get_progress_map map[uint32]condition_get_progress
)

func NewCondition(ctyp uint32, p1, p2 int64, p3, p4 string) *gamedata.Condition {
	n := &gamedata.Condition{
		Ctyp:   ctyp,
		Param1: p1,
		Param2: p2,
		Param3: p3,
		Param4: p4,
	}
	InitCondition(n)
	return n
}

func DelCondition(c *gamedata.Condition) {
	//p.Conds[c.cidx] = nil
	//p.Cond_next = append(p.Cond_next, c.cidx)
}

func InitCondition(c *gamedata.Condition) {
	f, ok := condition_initer_map[c.Ctyp]
	if !ok {
		logs.Error("condition typ %d unknown initer", c.Ctyp)
		return
	}
	f(c)
	return
}

// 条件状态更新逻辑
func UpdateCondition(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
	f, ok := condition_updater_map[c.Ctyp]
	if !ok {
		logs.Error("condition typ %d unknown updater", c.Ctyp)
		return
	}
	f(c, p1, p2, p3, p4)
	return
}

// 获取条件进度
func GetConditionProgress(
	c *gamedata.Condition,
	p *Account,
	ctyp uint32,
	p1, p2 int64,
	p3, p4 string) (progress, all int) {

	f, ok := condition_get_progress_map[c.Ctyp]
	progress = 0
	all = 100

	if !ok {
		logs.Error("condition typ %d unknown get_progress", c.Ctyp)
		return
	}

	progress, all = f(c, p, p1, p2, p3, p4)
	if progress > all {
		progress = all
	}
	return
}

// 检测进度
func CheckCondition(
	p *Account,
	ctyp uint32,
	p1, p2 int64,
	p3, p4 string) bool {

	f, ok := condition_get_progress_map[ctyp]

	if !ok {
		logs.Error("condition typ %d unknown get_progress", ctyp)
		return false
	}
	progress, all := f(nil, p, p1, p2, p3, p4)
	return progress >= all
}

func regConditionImpFunc(ctyp uint32,
	init condition_initer,
	up condition_update,
	gp condition_get_progress) {
	condition_initer_map[ctyp] = init
	condition_updater_map[ctyp] = up
	condition_get_progress_map[ctyp] = gp
}

/*
0.战队等级达到P1
1.P3关卡最高星级达到P1
2.完成P1任务
3.接取任务后以P2星级通关P3关卡P1次（接取任务后）
4.在P3关卡杀P4类型怪P1次
5.强化P1次（接取后）
6.精炼P1次（接取后）
7.刷某类关卡 TODO 现在关卡还没分类
12.精炼等级大于等于P1的部位数大于等于P2个--废弃，因为装备属于战队了
13.全身精炼(最小部位的精炼等级)等级大于等于P1的角色数量大于等于P2
14.觉醒等级达到P1的角色数量大于等于P2
15.VIP等级达到P1
16.在接取任务之后,在P2商店中购买道具的P1次
17.在接取任务之后,在P2类型Gacha中抽取道具的P1次(十连抽算十次)
18.在接取任务之后,参与神魔BOSS的击杀P1次（不要求杀死，但是需要结算）
19.在接取任务之后,通过任意关卡P1次(P2为关卡类型 0不限，1普通，2精英，3地狱)
20.在接取任务之后,完成副将委托P1次
21.当前时间在P3与P4之间
22.最高角色战力
23.计时分钟
24.购买体力P1次
25.购买金币P1次
26.洗炼P1次
27.金币关P1次
28.经验关P1次
29.参加1V1竞技场P1次
30.技能升级P1次
31.升星N次P1次
32.副将升级N次P1次
33.进攻公会bossP1次
34.同时镶嵌大于等于P1级的宝石部位数量超过P2个
35.任意升星等级大于等于P1超过P2个部位（必须同时存在）
36.指定ID为P2的神将升级到P1级
37.拥有P1个达到P2星的副将
38.精铁关P1次
39.通关过P1难度名将试炼
40.（接取后）参加N次及以上公会祈福
41.任务积分达到P1
42.完成任意N层爬塔（扫荡也算）
43.最高通关爬塔第N层
44.拥有N个品质大于等于X的副将
45.新玩家7天狂欢积分任务
46.钓鱼次数
47.每日登陆
48.每日购买钻石
49.每日消费钻石
50.穿戴N件N品质的装备
51.激活N个主将
52.将N件装备升阶N次
53.参加3V3次数（接受任务后）
54.参与组队BOSS次数（接受任务后）
55.接取副将任务P1次
56.激活神兽,P1神兽id
57.激活特定品质的主将,P1品质下限
58.任意主将达到某星级,P1星级
59.军团捐献科技点次数（接任务后）次数
60.攻打N次军团BOSS，无论成败(接取任务后)次数
61.在兵临城下中完成（取胜）N次战斗, 次数
62.参与N次吃包子小游戏 次数
*/

const (
	COND_TYP_Corp_lv                 = 0
	COND_TYP_Stage_Max_Star          = 1
	COND_TYP_Finish_Quest            = 2
	COND_TYP_Stage_Pass              = 3
	COND_TYP_Kill                    = 4
	COND_TYP_Upgrade                 = 5
	COND_TYP_Evolution               = 6
	COND_TYP_Stage_Pass_Class        = 7
	COND_TYP_Boss_Fight_Score        = 8
	COND_TYP_Boss_Fight_Kill         = 9
	COND_TYP_Boss_Share              = 10
	COND_TYP_General_Has             = 11
	COND_TYP_One_Evolution           = 12
	COND_TYP_All_Evolution           = 13
	COND_TYP_Arousals                = 14
	COND_TYP_VIP                     = 15
	COND_TYP_Buy_In_Store            = 16
	COND_TYP_GachaOne                = 17
	COND_TYP_Boss_Fight_Count        = 18
	COND_TYP_Any_Stage_Pass          = 19
	COND_TYP_General_Quest           = 20
	COND_TYP_Curr_Time               = 21
	COND_TYP_Max_Avatar_GS           = 22
	COND_TYP_Online_Time             = 23
	COND_TYP_BuyEnergy               = 24
	COND_TYP_BuyMoney                = 25
	COND_TYP_EquipAbstract           = 26
	COND_TYP_GoldLevel               = 27
	COND_TYP_ExpLevel                = 28
	COND_TYP_SimplePvp               = 29
	COND_TYP_AvatarSkill             = 30
	COND_TYP_EquipStarUp             = 31
	COND_TYP_GeneralLevelUp          = 32
	COND_TYP_GuildBoss               = 33
	COND_TYP_EquipedJade             = 34
	COND_TYP_EquipStarPartCount      = 35
	COND_TYP_DestinyGeneralLv        = 36
	COND_TYP_GeneralCount            = 37
	COND_TYP_FiLevel                 = 38
	COND_TYP_BossFightDegree         = 39
	COND_TYP_GuildSign               = 40
	COND_TYP_QuestPoint              = 41
	COND_TYP_Trial                   = 42
	COND_TYP_TrialMaxLv              = 43
	COND_TYP_GeneralCountWithRare    = 44
	COND_TYP_Quest7DayPoint          = 45
	COND_TYP_FishTimes               = 46
	COND_TYP_7Day_Login_Today        = 47
	COND_TYP_7Day_BuyHC_Today        = 48
	COND_TYP_7Day_CostHC_Today       = 49
	COND_TYP_Equip_Wear              = 50
	COND_TYP_General_Active          = 51
	COND_TYP_Equip_Mat_Enh           = 52
	COND_TYP_TeamPvp_Times           = 53
	COND_TYP_Gve_Times               = 54
	COND_TYP_General_Quest_Count     = 55
	COND_TYP_Activate_DestingGeneral = 56

	COND_TYP_Activate_Hero 			 = 57
	COND_TYP_HeroLvl      			 = 58

	//TODO: by YZH 所有条件和注释统一到一行,或者上下行.
	//TODO: by YZH 应该考虑不实用init()??

	//NOTE: YZH 请重新命名下面常量为合适的常量
	COND_TYP_Add_Guild_Science_Point = 59 // 59.军团捐献科技点次数（接任务后）次数
	COND_TYP_Guild_Boss              = 60 // 60.攻打N次军团BOSS，无论成败(接取任务后)次数
	COND_TYP_GateEnemy_Finish        = 61 // 61.在兵临城下中完成（取胜）N次战斗, 次数
	COND_TYP_EatBaozi                = 62 // 62.参与N次吃包子小游戏 次数
	COND_TYP_EatBaoziCount           = 63 // 63.单次吃包子的数目
	COND_TYP_FengHuoSubLevelCount    = 64 // 64.完成了烽火辽源多少小关
	COND_TYP_1v1_Times               = 65 // 65.历史总计参与1v1的次数
	COND_TYP_3v3_Times               = 66 // 66.历史总计参与3v3的次数
	COND_TYP_Expedition              = 67 // 67.每日参与远征的次数
	COND_TYP_Swing_Lvl_Together      = 68 // 68.M个主将的幻甲等级达到N
	COND_TYP_Hero_Star_Together      = 69 // 69.有M个主将的星级达到N
	COND_TYP_Hero_Lvl_Together       = 70 // 70.有M个主将的等级达到N
	COND_TYP_Swing_Star_Together     = 71 // 71.M个主将的幻甲星级达到N
	COND_TYP_IWant_Hero              = 72 // 72.参与我要名将N次
	COND_TYP_Try_Test                = 73 // 73.通关试炼之地N次
	COND_TYP_GVG_WINSTREAK           = 74 // 74.GVG军团战连斩次数
	COND_TYP_GVG_ONEWORLD            = 75 // 75.GVG军团战某个军团占领所有城池
	COND_TYP_HERO_ACTIVE             = 76 // 76.激活N个主将
	COND_TYP_GUILD_COLLECTION        = 77 // 77.团募捐n次
	COND_TYP_GUILD_FESTIVALBOSS      = 78 // 78.击杀了FestivalBoss次数
	COND_TYP_HERODIFF_FINISH         = 80 // 80.参与出奇制胜的次数

	COND_TYP_GUILD_WORSHIP           = 82 // 88.军团膜拜次数
	COND_TYP_ACT_EXCLUSIVE_WEAPON    = 83 // 83.N主将激活神兵，参数为主将个数
	COND_TYP_EVOLVE_EXCLUSIVE_WEAPON = 84 // 84.N个主将的神兵达到X品质
	COND_TYP_FB_Share                = 85 // 85.FaceBook分享多少次
	COND_TYP_Give_FriendGift         = 86 // 86.给好用送一次包子礼物
	COND_TYP_WorldBoss               = 87 // 87.当天打一次世界boss
	COND_TYP_ExchangeShop            = 88 // 88.兑换商店获得某一道具数量

	COND_TYP_Null 					 = 999
)

var (
	null_updater = func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
		return
	}

	null_initer = func(c *gamedata.Condition) {
		return
	}

	p1_count_initer = func(c *gamedata.Condition) {
		c.Param1 = 0
		return
	}

	p1_count_add_updater = func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
		//logs.Trace("p1_count_add_updater up %v %d %d %s %s",
		//	*c, p1, p2, p3, p4)
		if c.Param2 == p2 &&
			c.Param3 == p3 &&
			c.Param4 == p4 {
			c.Param1 += p1
		}
		//logs.Trace("p1_count_add_updater before %v %s %s %s %s",
		//	*c, c.Param2 == p2, c.Param3 == p3, c.Param4 == p4, p4)
	}

	p1_count_get_progress = func(c *gamedata.Condition,
		p *Account,
		p1, p2 int64,
		p3, p4 string) (int, int) {
		if c == nil {
			return 0, int(p1)
		}
		//logs.Trace("p1_count_get_progress %v %s",
		//	*c, p1)
		return int(c.Param1), int(p1)
	}
)

func init() {
	condition_updater_map = make(map[uint32]condition_update)
	condition_get_progress_map = make(map[uint32]condition_get_progress)
	condition_initer_map = make(map[uint32]condition_initer)
	// null update, 不需要更新逻辑的用这个

	// 0.战队等级达到P1
	regConditionImpFunc(COND_TYP_Corp_lv,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			lv, _ := p.Profile.GetCorp().GetXpInfo()
			return int(lv), int(p1)
		})

	// 1.P3关卡最高星级达到P1
	regConditionImpFunc(COND_TYP_Stage_Max_Star,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			star := p.Profile.GetStage().GetStarCount(p3)
			return int(star), int(p1)
		})

	// 2.完成P1任务
	regConditionImpFunc(COND_TYP_Finish_Quest,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			is_has_finish := p.Profile.GetQuest().IsHasClosed(uint32(p1))
			if is_has_finish {
				return 1, 1
			} else {
				return 0, 1
			}
		})

	// 3.接取任务后以P2星级通关P3关卡P1次 （接取任务后）
	regConditionImpFunc(COND_TYP_Stage_Pass,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			logs.Trace("COND_TYP_Stage_Pass up %v %v %v %s",
				*c, p1, p2, p3)
			if c.Param2 <= p2 &&
				c.Param3 == p3 {
				c.Param1 += p1
			}
		},
		p1_count_get_progress)

	// 4.在P4关卡杀P3类型怪P1次 P4为空表示全部关卡
	regConditionImpFunc(COND_TYP_Kill,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			logs.Trace("COND_TYP_Kill up %v - %v %v %s %s",
				*c, p1, p2, p3, p4)
			if c.Param3 == p3 {
				logs.Trace("COND_TYP_Kill add")
				if c.Param4 == "" || c.Param4 == p4 {
					logs.Trace("COND_TYP_Kill add %d", p1)
					c.Param1 += p1
				}
			}
		},
		p1_count_get_progress)

	// 5.强化P1次（接取后）
	regConditionImpFunc(COND_TYP_Upgrade,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	// 6.精炼P1次（接取后）
	regConditionImpFunc(COND_TYP_Evolution,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	// 7.刷某类关卡 TODO 现在关卡还没分类
	regConditionImpFunc(COND_TYP_Stage_Pass_Class,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	// 8.PveBoss累积功勋大于等于P1
	regConditionImpFunc(COND_TYP_Boss_Fight_Score,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			//scoreAll := p.Profile.GetBoss().ScoreAll
			return 0, 1
		})

	// 9.击杀PveBoss次数大于等于P1
	regConditionImpFunc(COND_TYP_Boss_Fight_Kill,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	// 10.PveBoss分享 TODO 实现 imp
	regConditionImpFunc(COND_TYP_Boss_Share,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			logs.Trace("COND_TYP_Boss_Share up %v - %v %v %s %s",
				*c, p1, p2, p3, p4)
			c.Param1 += p1
		},
		p1_count_get_progress)

	// 11.拥有P3与P4副将,P3或P4为空时不起作用
	regConditionImpFunc(COND_TYP_General_Has,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			progress, all := 0, 0
			if p3 != "" {
				all += 1
				if p.GeneralProfile.IsExistGeneral(p3) {
					progress += 1
				}
			}

			if p4 != "" {
				all += 1
				if p.GeneralProfile.IsExistGeneral(p4) {
					progress += 1
				}
			}

			return progress, all
		})
	// 12.精炼等级大于等于P1的部位数大于等于P2个
	regConditionImpFunc(COND_TYP_One_Evolution,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			equip := p.Profile.GetEquips()

			var part_lager_then_p1 int64
			for i := 0; i < gamedata.GetEquipSlotNum(); i++ {
				if int64(equip.GetEvolution(i)) >= p1 {
					part_lager_then_p1 += 1
				}
			}
			if part_lager_then_p1 >= p2 {
				return 1, 1
			}
			return 0, 1
		})
	// 13.全身精炼(最小部位的精炼等级)等级大于等于P1的角色数量大于等于P2--废弃，因为装备属于战队了
	regConditionImpFunc(COND_TYP_All_Evolution,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			equip := p.Profile.GetEquips()

			avatar_count := 0

			for avatar := 0; avatar < AVATAR_NUM_CURR; avatar++ {
				var part_min int64 = 65535 // 特大的值
				for i := 0; i < gamedata.GetEquipSlotNum(); i++ {
					e := int64(equip.GetEvolution(i))
					if e < part_min {
						part_min = e
					}
				}
				logs.Trace("avatar id min %d %d", avatar, part_min)
				if part_min >= p1 {
					avatar_count += 1
				}
			}

			return avatar_count, int(p2)
		})

	// 14.觉醒等级达到P1的角色数量大于等于P2
	regConditionImpFunc(COND_TYP_Arousals,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			exp := p.Profile.GetAvatarExp()
			avatar_count := 0
			for avatar := 0; avatar < AVATAR_NUM_CURR; avatar++ {
				if int64(exp.GetArousalLv(avatar)) >= p1 {
					avatar_count += 1
				}
			}
			return avatar_count, int(p2)
		})

	// 15.VIP等级达到P1
	regConditionImpFunc(COND_TYP_VIP,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return int(p.Profile.GetVipLevel()), int(p1)
		})

	// 16.在接取任务之后,在P2商店中购买道具的P1次
	regConditionImpFunc(COND_TYP_Buy_In_Store,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	// 17.在接取任务之后,在P2类型Gacha中抽取道具的P1次(十连抽算十次)
	regConditionImpFunc(COND_TYP_GachaOne,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			// 对于钻石Gacha 来说 单抽的id为1 十抽的id为4 这两个都被算入单抽的条件
			if c.Param2 == p2 &&
				c.Param3 == p3 &&
				c.Param4 == p4 {
				c.Param1 += p1
				return
			}
			// 对于钻石Gacha 来说 单抽的id为1 十抽的id为4 这两个都被算入单抽的条件
			if p2 == GachaHCTenID && c.Param2 == GachaHCOneID {
				c.Param1 += p1
				return
			}
		},
		p1_count_get_progress)
	// 18.在接取任务之后,参与神魔BOSS的击杀P1次（不要求杀死，但是需要结算）
	regConditionImpFunc(COND_TYP_Boss_Fight_Count,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	// 19.在接取任务之后,通过任意关卡P1次(P2为关卡类型 0不限，1普通，2精英) 不限扫荡
	regConditionImpFunc(COND_TYP_Any_Stage_Pass,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	// 20.在接取任务之后,完成副将委托P1次
	regConditionImpFunc(COND_TYP_General_Quest,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	// 21.当前时间在P3与P4之间
	regConditionImpFunc(COND_TYP_Curr_Time,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			curr_time := p.Profile.GetProfileNowTime()
			tb := util.DailyTime2UnixTime(curr_time, util.DailyTimeFromString(p3))
			te := util.DailyTime2UnixTime(curr_time, util.DailyTimeFromString(p4))
			logs.Trace("COND_TYP_Curr_Time %s-- %d %d %d -- %s", p3, tb, curr_time, te, p4)
			if curr_time >= tb && curr_time <= te {
				return 1, 1
			} else {
				return 0, 1
			}
		})

	//22.当前单个角色最大战力达到P1
	regConditionImpFunc(COND_TYP_Max_Avatar_GS,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			currGS := p.Profile.GetData().CorpCurrGS
			return currGS, int(p1)
		})
	//23.在线时间
	regConditionImpFunc(COND_TYP_Online_Time,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			logs.Trace("COND_TYP_Online_Time %v %v", c.Param1, int(p.GetOnlineTimeCurrDay()))
			return int(p.GetOnlineTimeCurrDay()), int(p1 * 60)
		})

	//24.购买体力P1次
	regConditionImpFunc(COND_TYP_BuyEnergy,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	//25.购买金币P1次
	regConditionImpFunc(COND_TYP_BuyMoney,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	//26.洗炼P1次
	regConditionImpFunc(COND_TYP_EquipAbstract,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	//27.金币关P1次
	regConditionImpFunc(COND_TYP_GoldLevel,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	//28.经验关P1次
	regConditionImpFunc(COND_TYP_ExpLevel,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	//29.参加1V1竞技场P1次
	regConditionImpFunc(COND_TYP_SimplePvp,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	//30.技能升级P1次
	regConditionImpFunc(COND_TYP_AvatarSkill,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	//31.升星N次P1次
	regConditionImpFunc(COND_TYP_EquipStarUp,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	//32.副将升级N次P1次
	regConditionImpFunc(COND_TYP_GeneralLevelUp,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)
	//33.进攻公会bossP1次
	regConditionImpFunc(COND_TYP_GuildBoss,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	//34.同时镶嵌大于等于P1级的宝石部位数量大于等于P2个
	//   包括所有生效的宝石 包括神兽和角色 需要支持红点
	regConditionImpFunc(COND_TYP_EquipedJade,
		p1_count_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			count := 0
			avatarJades := p.Profile.GetEquipJades()
			destJades := p.Profile.GetDestGeneralJades()
			JadeBag := p.Profile.GetJadeBag()
			for i := 0; i < len(avatarJades.Jades); i++ {
				jadeID := avatarJades.Jades[i]
				if jadeID != 0 {
					data, ok := JadeBag.GetJadeData(jadeID)
					if ok && (int64(data.GetJadeLevel()) >= p1) {
						count += 1
					}
				}
			}

			for i := 0; i < len(destJades.DestinyGeneralJade); i++ {
				jadeID := destJades.DestinyGeneralJade[i]
				if jadeID != 0 {
					data, ok := JadeBag.GetJadeData(jadeID)
					if ok && (int64(data.GetJadeLevel()) >= p1) {
						count += 1
					}
				}
			}

			// 没达到条件时正常更新 如果已经完成状态 就不再回到未完成状态
			if c.Param1 < p2 {
				c.Param1 = int64(count)
			}

			return int(c.Param1), int(p2)
		})

	//35.任意升星等级大于等于P1大于等于P2个部位（必须同时存在） 需要支持红点
	regConditionImpFunc(COND_TYP_EquipStarPartCount,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			count := 0
			_, _, _, avatarStars, _, _, _, _ := p.Profile.GetEquips().Curr()
			for i := 0; i < len(avatarStars); i++ {
				if int64(avatarStars[i]) >= p1 {
					count += 1
				}
			}
			return count, int(p2)
		})

	//36.指定ID为P2的神兽升级大于等于P1级 --> 大等级 从 0 到 50 而不是 从 0 到 201 需要支持红点
	regConditionImpFunc(COND_TYP_DestinyGeneralLv,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			dg := p.Profile.GetDestinyGeneral()
			d := dg.GetGeneral(int(p2))
			if d != nil {
				if d.LevelIndex == 0 {
					return 0, int(p1)
				}
				return d.LevelIndex, int(p1)
			}
			return 0, int(p1)
		})

	//37.拥有P1个达到P2星的副将 需要支持红点
	regConditionImpFunc(COND_TYP_GeneralCount,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			general := p.GeneralProfile
			count := 0
			for i := 0; i < len(general.GeneralAr); i++ {
				g := &(general.GeneralAr[i])
				if g.IsHas() && int(g.StarLv) >= int(p2) {
					count += 1
				}
			}
			return count, int(p1)
		})

	//38.精铁关P1次
	regConditionImpFunc(COND_TYP_FiLevel,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	//39.通关过P1难度名将试炼
	regConditionImpFunc(COND_TYP_BossFightDegree,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			if p.Profile.GetBoss().MaxDegree >= int(p1) {
				return 1, 1
			} else {
				return 0, 1
			}
		})

	//40.（接取后）参加N次及以上公会祈福
	regConditionImpFunc(COND_TYP_GuildSign,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	//41.任务积分大于等于P1
	regConditionImpFunc(COND_TYP_QuestPoint,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			nowT := p.Profile.GetProfileNowTime()
			curr := p.Profile.GetQuest().GetQuestPoint(nowT)
			return curr, int(p1)
		})

	//42.完成任意N层爬塔（扫荡也算）
	regConditionImpFunc(COND_TYP_Trial,
		p1_count_initer,
		p1_count_add_updater,
		p1_count_get_progress)

	//43.最高通关爬塔第N层
	regConditionImpFunc(COND_TYP_TrialMaxLv,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return int(p.Profile.GetPlayerTrial().MostLevelId), int(p1)
		})

	//44.拥有N个品质大于等于X的副将
	regConditionImpFunc(COND_TYP_GeneralCountWithRare,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			general := p.GeneralProfile
			count := 0
			for i := 0; i < len(general.GeneralAr); i++ {
				g := &(general.GeneralAr[i])
				data := gamedata.GetGeneralInfo(g.Id)
				if data != nil {
					if g.IsHas() && int(data.GetRareLevel()) >= int(p2) {
						count += 1
					}
				}
			}
			return count, int(p1)
		})

	// 45.7天活动任务积分，大于等于P1
	regConditionImpFunc(COND_TYP_Quest7DayPoint,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			curr := p.Profile.GetQuest().GetAccount7DayQuestPoint()
			return curr, int(p1)
		})

	// 46.钓鱼次数，大于等于P1
	regConditionImpFunc(COND_TYP_FishTimes,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			cfg := gamedata.GetGameModeControlData(counter.CounterTypeFish)
			has, _ := p.Profile.GetCounts().Get(counter.CounterTypeFish, p)
			lc := cfg.GetCount - has

			cfg = gamedata.GetGameModeControlData(counter.CounterTypeFishHC)
			has, _ = p.Profile.GetCounts().Get(counter.CounterTypeFishHC, p)
			lc += cfg.GetCount - has
			return lc, int(p1)
		})

	// 47.每日登陆, P1次数, P2所在天数(0为每天)
	regConditionImpFunc(COND_TYP_7Day_Login_Today,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			dayNum := gamedata.GetCommonDayDiff(p.Profile.CreateTime,
				p.Profile.GetProfileNowTime()) + 1
			if p2 > 0 && dayNum == p2 {
				return int(dayNum), int(p1)
			}
			if p2 == 0 {
				return int(dayNum), int(p1)
			}
			return 0, int(p1)
		})

	// 48.每日购买钻石, P1次数, P2所在天数(0为所有历史钻石，1为指定天钻石)
	regConditionImpFunc(COND_TYP_7Day_BuyHC_Today,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			dayNum := gamedata.GetCommonDayDiff(p.Profile.CreateTime,
				p.Profile.GetProfileNowTime()) + 1
			if p2 == 0 { // 所有的钻石
				hc := p.Profile.GetHC().BuyFromHc
				return int(hc), int(p1)
			}
			if dayNum == p2 { // 指定天的钻石
				n := p.Profile.GetHC().GetBuyFromHcToday(p.Profile.GetProfileNowTime())
				return int(n), int(p1)
			}

			return 0, int(p1)
		})

	// 49.每日消费钻石, P1次数, P2所在天数
	regConditionImpFunc(COND_TYP_7Day_CostHC_Today,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			dayNum := gamedata.GetCommonDayDiff(p.Profile.CreateTime,
				p.Profile.GetProfileNowTime()) + 1
			if dayNum == p2 {
				n := p.Profile.GetHC().GetCostHcToday(p.Profile.GetProfileNowTime())
				return int(n), int(p1)
			}
			return 0, int(p1)
		})

	// 50.穿戴N件N品质的装备, P1装备数, P2品质
	regConditionImpFunc(COND_TYP_Equip_Wear,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			n := 0
			eqs, _, _, _, _, _, _, _ := p.Profile.GetEquips().Curr()
			for i := 0; i < len(eqs) && i < gamedata.GetEquipSlotNum(); i++ {
				eq := eqs[i]
				cfg, ok := p.BagProfile.GetItemData(eq)
				if ok {
					if cfg.GetRareLevel() == int32(p2) {
						n++
					}
				}
			}
			return n, int(p1)
		})

	// 51.激活N个主将, P1主将数
	regConditionImpFunc(COND_TYP_General_Active,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			return p.GeneralProfile.GetAllExistGeneralCount(), int(p1)
		})

	// 52.将N件装备升阶N次, P1装备数, P2升阶等级
	regConditionImpFunc(COND_TYP_Equip_Mat_Enh,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			_, _, _, _, _, lv_eq_mat_enh, _, _ := p.Profile.GetEquips().Curr()
			n := 0
			for _, l := range lv_eq_mat_enh {
				if l >= uint32(p2) {
					n++
				}
			}
			return n, int(p1)
		})

	// 53.参加3V3次数（接受任务后）
	regConditionImpFunc(COND_TYP_TeamPvp_Times,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)

	// 54.参与组队BOSS次数（接受任务后）
	regConditionImpFunc(COND_TYP_Gve_Times,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)

	// 55.接取副将任务P1次
	regConditionImpFunc(COND_TYP_General_Quest_Count,
		p1_count_initer,
		func(c *gamedata.Condition, p1, p2 int64, p3, p4 string) {
			c.Param1 += p1
		},
		p1_count_get_progress)

	// 56.激活神兽,P1神兽id
	regConditionImpFunc(COND_TYP_Activate_DestingGeneral,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			if p.Profile.GetDestinyGeneral().GetGeneral(int(p1)) != nil {
				return 1, int(p2)
			}
			return 0, int(p2)
		})

	// 57.激活特定品质的主将,P1品质下限
	regConditionImpFunc(COND_TYP_Activate_Hero,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			for idx, lvl := range p.Profile.Hero.HeroStarLevel {
				if lvl > 0 {
					info := gamedata.GetHeroData(idx)
					if info.RareLv >= uint32(p1) {
						return 1, int(p2)
					}
				}
			}
			return 0, int(p2)
		})

	// 58.任意主将达到某星级,P1星级
	regConditionImpFunc(COND_TYP_HeroLvl,
		null_initer,
		null_updater,
		func(c *gamedata.Condition, p *Account, p1, p2 int64, p3, p4 string) (int, int) {
			for _, lvl := range p.Profile.Hero.HeroStarLevel {
				if lvl >= uint32(p1) {
					return 1, int(p2)
				}
			}
			return 0, int(p2)
		})

	condition_impl_59()

}
