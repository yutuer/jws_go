package market_activity

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 公会红包  固定iap
func (ma *PlayerMarketActivitys) OnRedPacketByPay(acid string, iapId int,
	now_t int64, shardId uint, guildUuid, name string) bool {
	ma.UpdateMarketActivity(acid, now_t)

	activities, retCode := ma.getAllActivityByType(acid, now_t, gamedata.ActRedPacket)
	if retCode > 0 {
		logs.Debug("GuildRedPakcet:not find any red packet, %d", retCode)
		return false
	}
	result := false
	for _, pa := range activities {
		if ok := pa.isActAvailable(acid, now_t); !ok {
			logs.Debug("GuildRedPacket:act is not available, now time = %d", now_t)
			continue
		}
		ma.sendRedPacket2Guild(shardId, guildUuid, name)
		result = true // 实际只有一个活动同时存在
	}
	if result {
		ma.SyncObj.SetNeedSync()
	}
	return result
}

// 向公会发红包
func (ma *PlayerMarketActivitys) sendRedPacket2Guild(shardId uint, guildUuid, name string) {
	if guildUuid == "" {
		logs.Debug("GuildRedPacket: can not send red packet, no guild: %s", guildUuid)
		return
	}
	guild.GetModule(shardId).SendRedPacket(guildUuid, name)
}

func (ma *PlayerMarketActivitys) HasRedPacketActivity(acid string, now_t int64) (bool, uint32) {
	ma.UpdateMarketActivity(acid, now_t)

	activities, retCode := ma.getAllActivityByType(acid, now_t, gamedata.ActRedPacket)
	if retCode > 0 || len(activities) <= 0 {
		logs.Debug("not find any red packet, %d", retCode)
		return false, 0
	}

	for _, pa := range activities {
		if ok := pa.isActAvailable(acid, now_t); ok {
			ma.SyncObj.SetNeedSync()
			return true, pa.ActivityId
		}
	}

	return false, 0
}
