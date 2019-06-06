package base

import (
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type ActBase struct {
	LastRefershTime int64 `json:"lrt"`
	guildHandler    GuildHandler
}

func (a *ActBase) IsNeedRefersh(
	nowT int64,
	dailyRefershTime util.TimeToBalance) bool {
	//logs.Trace("IsNeedRefersh %v %v %v",
	//	a.LastRefershTime, nowT, dailyRefershTime)
	return !util.IsSameUnixByStartTime(
		nowT,
		a.LastRefershTime,
		dailyRefershTime)
}

func (a *ActBase) SetHasRefersh(nowT int64) {
	a.LastRefershTime = nowT
	logs.Trace("SetHasRefersh %v", a.LastRefershTime)
}

func (a *ActBase) SetGuildHandler(g GuildHandler) {
	a.guildHandler = g
}

func (a *ActBase) GetGuildHandler() GuildHandler {
	return a.guildHandler
}

type GuildHandler interface {
	AddGuildInventory(loots []string, cs []uint32, reason string)
	NotifyAll(typ int)
	SetNeedSave2DB()
	GetGuildLv() uint32
	// 返回受科技加成后的军魂数量
	OnGuildBossDied(bossName string, itemIds []string, count []uint32) uint32
}
