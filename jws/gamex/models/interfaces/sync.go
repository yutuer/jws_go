package interfaces

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

type ISyncRspWithRewards interface {
	helper.ISyncRsp
	AddResReward(g *gamedata.CostData2Client)
	MergeReward()
}
