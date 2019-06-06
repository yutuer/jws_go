package worldboss

import (
	"fmt"

	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//..
const (
	MarqueeTypeChampion = "champion"
)

//ParamMarquee ..
type ParamMarquee struct {
	Sid          uint32
	Batch        string
	MsgType      string
	ChampionName string
	ChampionSid  uint32
}

//CallbackMarquee ..
type CallbackMarquee struct {
	module.BaseMethod
}

func newCallbackMarquee(m module.Module) *CallbackMarquee {
	return &CallbackMarquee{
		module.BaseMethod{Method: CallbackMarqueeID, Module: m},
	}
}

//NewParam ..
func (m *CallbackMarquee) NewParam() module.Param {
	return &ParamMarquee{}
}

func (c *callbackHolder) Marquee() error {
	list := c.res.RankDamageMod.getRange(0, 1)
	if 0 == len(list) {
		logs.Warn("[WorldBoss] callbackHolder Marquee getRange empty")
		return nil
	}
	champion := list[0]
	playerInfo := c.res.PlayerMod.getPlayerInfo(champion.Acid)
	if nil == playerInfo {
		logs.Warn("[WorldBoss] callbackHolder Marquee getPlayerInfo empty")
		return nil
	}
	shardList := gamedata.GetWBGSids(c.res.group)
	for _, sid := range shardList {
		param := &ParamMarquee{
			Sid:          sid,
			Batch:        c.res.ticker.roundStatus.BatchTag,
			MsgType:      MarqueeTypeChampion,
			ChampionName: playerInfo.Name,
			ChampionSid:  playerInfo.Sid,
		}
		if err := c.res.module.Push(sid, ModuleID, CallbackMarqueeID, param); nil != err {
			logs.Error(fmt.Sprintf("[WorldBoss] callbackHolder Marquee, Push to Shard [%d] failed, %v ...Param %+v", sid, err, param))
		}
	}
	return nil
}
