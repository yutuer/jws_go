package market_activity

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Command_Null = iota
	Command_HotDataUpdate
	Command_MakeSnapShoot
	Command_SendReward
	Command_RankParentID
)

type marketCommand struct {
	Type    int
	resChan chan marketCommandRes

	marketCommandParam
}

type marketCommandParam struct {
	HotDataVer *gamedata.DataVerConf

	ActivityType uint32
	ActivityID   uint32
	Acid         string
}

type marketCommandRes struct {
}

func makeMarketCommand(cmdtype int, param marketCommandParam) *marketCommand {
	return &marketCommand{
		Type:               cmdtype,
		resChan:            make(chan marketCommandRes, 1),
		marketCommandParam: param,
	}
}

func (ma *MarketActivityModule) dispatch(cmd *marketCommand) {
	switch cmd.Type {
	case Command_HotDataUpdate:
		ma.cmdHotDataUpdate(cmd)
	case Command_MakeSnapShoot:
		ma.cmdMakeSnapShoot(cmd)
	case Command_SendReward:
		ma.cmdSendReward(cmd)
	case Command_RankParentID:
		ma.cmdRankParentID(cmd)
	default:
		logs.Warn("[MarketActivityModule] Unkown command type [%d]", cmd.Type)
	}
}
