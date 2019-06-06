package account

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

type GetLastSetCurLogType func(string) string

func (p *Account) GetGiveCurrencyLog(accountId string, avatar int, corpLvl uint32, channel string,
	reason string, typ string, oldV int64, chgV int64,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {
	// logiclog
	item, _ := gamedata.GetProtoItem(typ)
	itemType := item.GetType()
	logiclog.LogGiveCurrency(accountId, avatar, corpLvl,
		channel, reason, typ, oldV, chgV, "", p.GetIp(),
		p.Profile.GetVipLevel(), 0, itemType, p.Profile.Name,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

}
