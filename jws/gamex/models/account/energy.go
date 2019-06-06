package account

import (
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type PlayerEnergy struct {
	Value     int64 `json:"v"`
	Last_time int64 `json:"last"`

	ForBossFight int64 `json:"boss"`

	handler events.Handler

	// 计算最大值需要其他信息
	player *Profile
}

func (p *PlayerEnergy) Init(player *Profile) {
	p.player = player
	p.Value = p.getMax()
	p.Last_time = player.GetRegTimeUnix()
}

// 返回类型为typ的值
// 客户端需要根据这些信息模拟体力增长
// 考虑到网络通信是有延时的，所以一方面要把体力值和体力增长时间的计算剩余值给客户端
// 另一方面要把计算体力更新的时间发送给客户端
func (p *PlayerEnergy) Get() (value, refersh_time, last_time int64) {
	// 体力增长时间的计算剩余值
	last_time = p.updateValue()
	value = p.Value
	refersh_time = p.Last_time
	return
}

func (p *PlayerEnergy) GetMax() int64 {
	return p.getMax()
}

// 增加类型为typ的值add点，返回增加后的值
func (p *PlayerEnergy) Add(account, reason string, add int64) int64 {
	p.updateValue()

	// logiclog
	a := Account{}
	a.GetGiveCurrencyLog(account, p.player.GetCurrAvatar(), p.player.GetCorp().GetLvlInfo(),
		p.player.ChannelId, reason, helper.VI_EN, p.Value, add,
		func(last string) string { return p.player.GetLastSetCurLogicLog(last) }, "")

	p.Value += add
	if p.Value > p.getMax() {
		p.Value = p.getMax()
	}

	return p.Value
}

// 同Add，为强制增加 如果超出最大上限返回false
func (p *PlayerEnergy) AddForce(account, reason string, add int64) bool {
	p.updateValue()
	a := Account{}
	// logiclog
	a.GetGiveCurrencyLog(account, p.player.GetCurrAvatar(), p.player.GetCorp().GetLvlInfo(),
		p.player.ChannelId, reason, helper.VI_EN, p.Value, add,
		func(last string) string { return p.player.GetLastSetCurLogicLog(last) }, "")

	//if p.Value >= int64(gamedata.GetCommonCfg().GetMaxEnergy()) {
	//	return false
	//}

	p.Value += add

	return true

}

// 设置typ值为最大值，如果值大于最大值不处理，
func (p *PlayerEnergy) Add2Max(account, reason string) int64 {
	p.updateValue()

	oldV := p.Value
	p.Value = p.getMax()

	if oldV < p.Value {
		// logiclog
		a := Account{}
		a.GetGiveCurrencyLog(account, p.player.GetCurrAvatar(), p.player.GetCorp().GetLvlInfo(),
			p.player.ChannelId, reason, helper.VI_EN, oldV, p.Value-oldV,
			func(last string) string { return p.player.GetLastSetCurLogicLog(last) }, "")
	}
	return p.Value
}

func (e *PlayerEnergy) SetHandler(handler events.Handler) {
	e.handler = handler
}

// 使用typ值v点，返回是否成功 TBD 体力消耗途径日志
func (p *PlayerEnergy) Use(account, reason string, v int64) bool {
	if p.Has(v) {
		// Has函数中会更新，这里就不调用更新了

		// logiclog
		logiclog.LogCostCurrency(account, p.player.GetCurrAvatar(), p.player.GetCorp().GetLvlInfo(),
			p.player.ChannelId, reason, helper.VI_EN, p.Value, v, p.player.GetVipLevel(),
			func(last string) string { return p.player.GetLastSetCurLogicLog(last) }, "")

		p.Value -= v
		p.ForBossFight += v

		if p.handler != nil {
			p.handler.OnEnergyUsed(v)
		}

		return true
	} else {
		return false
	}
}

// 是否拥有typ值v点
func (p *PlayerEnergy) Has(v int64) bool {
	p.updateValue()
	return p.Value >= v
}

func (p *PlayerEnergy) getMax() int64 {
	// TODO 体力最大值计算方法待确认 T4204
	lv, _ := p.player.GetCorp().GetXpInfo()
	info := gamedata.GetCorpLvConfig(lv)
	if info == nil {
		logs.Error("getMax Energy Error %d", lv)
		return 20 // TODO暂时做个保护 T4204
	}

	return int64(info.MaxEnergy)
}

// 获取每次体力增长所需的时间，以秒为单位
func (p *PlayerEnergy) getTimePreAdd() int64 {
	min := gamedata.GetCommonCfg().GetEnergyRecover()
	return int64(min) * 60
}

// 更新体力值
func (p *PlayerEnergy) updateValue() int64 {
	//加个保护
	if p.Last_time <= 0 {
		p.Init(p.player)
	}

	max := p.getMax()

	now_unix_sec := p.player.GetRegTimeUnix()

	// 如果现在已经是最大值以上的话，不更新
	if p.Value >= max {
		p.Last_time = now_unix_sec
		return 0
	}

	one_need := p.getTimePreAdd()

	// 计算增量
	add_point, s := util.AccountTime2Point(
		now_unix_sec,
		p.Last_time,
		one_need)

	// 更新值
	p.Value += add_point
	// 时间增长不会超过最大值
	if p.Value > max {
		p.Value = max
	}

	// 依据文档中的算法上次更新时间应该置为this
	p.Last_time = now_unix_sec - s

	return s
}
