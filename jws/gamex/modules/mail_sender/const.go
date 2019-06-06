package mail_sender

import "vcs.taiyouxi.net/jws/gamex/models/helper"

// 一个要消耗/赠与东西的列表
// 这个主要是配合逻辑中得CostGroup和GiveGroup使用
// 虽然名字是CostData 但也可以根据这个给玩家赠送东西
//

const (
	VI_Sc0                = helper.VI_Sc0
	VI_Sc1                = helper.VI_Sc1
	VI_Hc_Buy             = helper.VI_Hc_Buy
	VI_Hc_Give            = helper.VI_Hc_Give
	VI_Hc_Compensate      = helper.VI_Hc_Compensate
	VI_Hc                 = helper.VI_Hc
	VI_XP                 = helper.VI_XP
	VI_CorpXP             = helper.VI_CorpXP
	VI_EN                 = helper.VI_EN
	VI_GoldLevelPoint     = helper.VI_GoldLevelPoint
	VI_ExpLevelPoint      = helper.VI_ExpLevelPoint
	VI_BossFightPoint     = helper.VI_BossFightPoint
	VI_BossFightRankPoint = helper.VI_BossFightRankPoint
	VI_HcByVIP            = "VI_HcByVIP"
	VI_BossCoin           = "VI_BC"
	VI_PvpCoin            = "VI_PVPC"
	VI_StarBlessCoin      = helper.VI_StarBlessCoin
	VI_BaoZi              = helper.VI_BaoZi
)

/*
	参照mail.xlsx
	http://wiki.taiyouxi.net/w/三国-设计文档区/系统设计/邮件/
*/
const (
	IDS_MAIL_SYS_TITLE                            = iota // 服务器自定义邮件
	IDS_MAIL_PVEBOSS_RANKREWARD_ACCU_TITLE               // 1.名将乱入累计军功排行榜奖励
	IDS_MAIL_PVEBOSS_RANKREWARD_SINGLE_TITLE             // 2.名将乱入单场军功排行榜奖励
	IDS_MAIL_SIMPLEPPVP_RANKREWARD_TITLE                 // 3.单人竞技场排行奖励
	IDS_MAIL_GENERAL_QUESTREWARD_TITLE                   // 4.派兵遣将
	IDS_MAIL_GUILD_KICKEDOUT_TITLE                       // 5.你被踢出了公会
	IDS_MAIL_GUILD_DISBANDED_TITLE                       // 6.公会解散
	IDS_MAIL_FASHION_TIMEOUT_TITLE                       // 7.时装过期
	IDS_MAIL_GUILD_DECLINE_TITLE                         // 8.公会申请被拒
	IDS_MAIL_ACTIVITY_7DAYRANK_TITLE                     // 9.7日个人战力排行榜
	IDS_MAIL_ACTIVITY_7DAYAWARD_TITLE                    // 10.7日个人战力奖励
	IDS_MAIL_ACTIVITY_7DAYGUILD_TITLE                    // 11.7日公会战力排行榜（会员）
	IDS_MAIL_ACTIVITY_7DAYGUILDLEAD_TITLE                // 12.7日公会战力排行榜（会长）
	IDS_MAIL_GUILD_CHANGELEVEL_TITLE                     // 13.公会职位变更
	IDS_MAIL_TEAMPVP_REWARD                              // 14.3v3竞技场奖励
	IDS_MAIL_GUILD_GVEBOSSBAG_TITLE                      // 15.公会仓库
	IDS_MAIL_GUILD_FIRSTBACKHC_TITLE                     // 16.首次充值反钻
	IDS_MAIL_GUILD_SCENDBACKHC_TITLE                     // 17.二次充值反钻
	IDS_MAIL_ACTIVITY_ACCLOGIN_TITLE                     // 18.累计登录
	IDS_MAIL_ACTIVITY_ACCPAYDAY_TITLE                    // 19.累计充值天数
	IDS_MAIL_ACTIVITY_ACCPAYSUM_TITLE                    // 20.累计充值金额
	IDS_MAIL_ACTIVITY_ACCCONSUME_TITLE                   // 21.累计消费金额
	IDS_MAIL_ACTIVITY_ACCLEVEL_TITLE                     // 22.累计完成活动次数
	IDS_MAIL_ACTIVITY_ACCRESOURCE_TITLE                  // 23.累计购买资源次数
	IDS_MAIL_ACTIVITY_HGRBOX_TITLE                       // 24.限时神将结束，宝箱没领的奖励发放。
	IDS_MAIL_SIMPLEPPVP_RANKREWARDWEEK_TITLE             // 25.单人竞技场周排行奖励
	IDS_MAIL_ACTIVITY_HGRRANK_TITLE                      // 26.限时神将发放排名奖励
	IDS_MAIL_GUILD_GVEBOSSBAG_REFUSE_TITLE               // 27.仓库拒绝（手动）
	IDS_MAIL_GUILD_GVEBOSSBAG_AUTO_TITLE                 // 28.仓库拒绝（自动）
	IDS_MAIL_ACTIVITY_ACCPAYDAYSUM_TITLE                 // 29.当日累计充值金额
	IDS_MAIL_ACTIVITY_HERO_STAR_TITLE                    // 30.将星
	IDS_MAIL_ACTIVITY_ACCDAYCONSUME_TITLE                // 31.当日累计消费金额
	IDS_MAIL_ACTIVITY_YYB_GIFT_TITLE                     // 32.应用宝礼包邮件
	IDS_MAIL_GVG_FIGHT_GIFT_TITLE                        // 33.军团战攻城礼包奖励
	IDS_MAIL_GVG_POINT_GIFT_TITLE                        // 34.军团战积分礼包奖励
	IDS_MAIL_GVG_DAILY_GIFT_TITLE                        // 35.军团战每日礼包奖励
	IDS_MAIL_GUILD_AUTO_CHANGE_CHIEF                     // 36.工会自动更换军团长
	IDS_MAIL_HERO_FUND_ON_BANLANCE                       // 37.名将投资结束时发送未领取的奖励
	IDS_MAIL_SINGLEPAY_CONTENT                           // 38.单笔充值活动奖励
	IDS_MAIL_YYBPAY_TITLE                                // 39.应用宝充值奖励
	IDS_MAIL_REGUILDNAME_TITLE                           // 40.军团更名通知
	IDS_MAIL_WORSHIP_TITLE                               // 41.军团膜拜
	IDS_MAIL_REPACKT_TITLE                               // 42.开服七日红包
	IDS_MAIL_ON_GUILD_BOSS_DIED_TITLE                    // 43.军团BOSS死亡邮件
	IDS_MAIL_HMT_PAY_FEED_BACK_TITLE                     // 44.HMT封測儲值鑽石返還
	IDS_MAIL_VI_GB_DEL_TITLE                             // 45.军魂清空补偿邮件
	IDS_MAIL_VI_GB_INVENTORY_TITLE                       // 46.未申请批准的军魂补偿邮件
	IDS_MAIL_CAPACITY_TITLE                              // 47.开服战力排行榜奖励
	IDS_MAIL_GROUPLV_TITLE                               // 48.开服战队等级排行榜奖励
	IDS_MAIL_DG_TITLE                                    // 49.开服神兽等级排行榜奖励
	IDS_MAIL_JADE_TITLE                                  // 50.开服宝石等级排行榜奖励
	IDS_MAIL_EQUIPSTAR_TITLE                             // 51.开服装备星级排行榜奖励
	IDS_MAIL_HEROSTAR_TITLE                              // 52.开服主将星级排行榜奖励
	IDS_MAIL_AUTO_EXCHANGE_SHOP_PROP_TITLE               // 53.兑换商店奖励
	IDS_MAIL_CROPS_TITLE                                 // 54.劫营夺粮运粮奖励
	IDS_MAIL_ROB_TITLE                                   // 55.劫营夺粮夺粮奖励
	IDS_MAIL_ESCORT_TITLE                                // 56.劫营夺粮护送奖励
	IDS_MAIL_BLACK_GACHA_HERO_TITLE                      // 57.黑盒宝箱武将结算奖励邮件
	IDS_MAIL_BLACK_GACHA_GWC_TITLE                       // 58.黑盒宝箱神兵结算奖励邮件
	IDS_MAIL_ASTROLOGY_TITLE                             // 59.开服星图排行榜奖励
	IDS_MAIL_HERODESTINY_TITLE                           // 60.开服羁绊排行榜奖励
	IDS_MAIL_HWSTAR_TITLE                                // 61.开服幻甲星级排行榜奖励
	IDS_MAIL_GWC_TITLE                                   // 62.开服神兵积分排行榜奖励
	IDS_MAIL_WS_CAPACITY_TITLE                           // 63.开服无双战力排行榜奖励
	IDS_MAIL_WB_RANKREWARD_TITLE                         // 64.世界boss排行榜奖励
	IDS_MAIL_WB_KILLBOSSREWARD_TITLE                     // 65.世界boss杀boss奖励
	IDS_MAIL_WB_DPSREWARD_TITLE                          // 66.世界boss伤害量忘领后发给玩家
	IDS_MAIL_DAILY_UCGIFT_TITLE                          // 67.九游-每日礼包
	IDS_MAIL_WEEKLY_UCGIFT_TITLE                         // 68.九游-每周礼包
	IDS_MAIL_LV30_UCGIFT_TITLE                           // 69.九游-等级礼包=30级
	IDS_MAIL_LV50_UCGIFT_TITLE                           // 70.九游-等级礼包=50级
	IDS_MAIL_LV60_UCGIFT_TITLE                           // 71.九游-等级礼包=60级
	IDS_MAIL_PAY300_UCGIFT_TITLE                         // 72.九游-充值礼包=300钻
	IDS_MAIL_VIPDAILY_UCGIFT_TITLE                       // 73.九游-会员每日礼包
	IDS_MAIL_VIPONCE_UCGIFT_TITLE                        // 74.九游-会员一次性礼包
	IDS_MAIL_ONCE_UCGIFT_TITLE                           // 75.九游-全体礼包
	IDS_MAIL_AUTO_EXCHANGE_SHOP_PROP_CUSTOM_TITLE        // 76.兑换商店自定义奖励邮件
)

const (
	Mail_DB_Counter_Name = "Mail_DB"
	mail_batch_table     = "mails"
)
