package uutil

type JWSConfig struct {
	MatchUrl    string `toml:"match_url"`
	MatchToken  string `toml:"match_token"`
	GVEStartUrl string `toml:"gve_start_url"`
	GVEStopUrl  string `toml:"gve_stop_url"`
}

var JwsCfg JWSConfig

const (
	Hot_Value_Phone              = iota //绑定手机号
	Hot_Value_HitEeg                    //砸金蛋
	Hot_Value_Total_Query_day           //累计充值天数
	Hot_Value_Total_Enter_day           //累计登录天数
	Hot_Value_Total_Query               //累计充值金额
	Hot_Value_Total_Buy                 //累计消费金额 5
	Hot_Value_Total_Play                //累计参与玩法次数
	Hot_Value_Total_Buy_Resource        //累计购买资源次数
	Hot_Value_Day_Total_Query           //每日累计充值金额
	Hot_Value_Day_Total_Buy             //每日累计消费金额
	Hot_Value_Star_Hero                 //将星之路 10
	Hot_Value_Limit_Hero                //限时神将
	Hot_Value_Money_Cat                 //招财猫
	Hot_Value_GvG                       //攻城战
	Hot_Value_Limit_Store               //限时商店
	Hot_Value_HERO_FUND                 //英雄投资 15
	Hot_Value_Only_Pay                  //单笔充值
	Hot_Value_Red_Packet                //红包类型
	Hot_Value_FestivalBoss              //节日BOSS
	Hot_Value_WhiteGacha                //白盒宝箱
	Hot_Value_SevenRedPacket            //开服七日红包 20
	Hot_Value_SevenRank                 //开服七天排行榜
	Hot_Value_ActivityRank              //运营排行榜
	Hot_Value_FaceBookShare             //FaceBook分享
	Hot_Value_FaceBookInvite            //FaceBook邀请
	Hot_Value_FaceBookFocus             //FaceBook关注 25
	Hot_value_JumpStore                 //跳转商店评分
	Hot_value_ExchangeShop              //兑换商店  27
	Hot_value_WeaponsBlackGacha         //神兵黑盒宝箱
	Hot_value_HeroBlackGacha            //主将黑盒宝箱
	Hot_value_SevenRankTwo              //开服第二周排行榜
	Hot_value_EG                        //EG账号绑定 31
	Hot_Value_Num
)

const (
	Android_Platform = "android"
	IOS_Platform     = "ios"
)
