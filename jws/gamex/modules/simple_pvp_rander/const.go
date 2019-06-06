package sPvpRander

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
	VI_StarBlessCoin      = helper.VI_StarBlessCoin
	VI_BaoZi              = helper.VI_BaoZi

	AVATAR_NUM_CURR = helper.AVATAR_NUM_CURR
)
