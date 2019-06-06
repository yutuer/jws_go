package player_state

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util"
)

type PlayerActivityState struct {
	AcId        string
	LastActTime int64
	Params      []int64
}

type PlayerActivityStates struct {
	Players [helper.MaxGuildMember]PlayerActivityState `json:"players"`
	CurrNum int                                        `json:"currNum"`
}

func (p *PlayerActivityStates) Clean() {
	p.Players = [helper.MaxGuildMember]PlayerActivityState{}
	p.CurrNum = 0
}

func (p *PlayerActivityStates) GetActNum(nowT int64, dailyBegin util.TimeToBalance) int {
	res := 0
	for i := 0; i < len(p.Players) && i < p.CurrNum; i++ {
		if util.IsSameUnixByStartTime(
			p.Players[i].LastActTime,
			nowT,
			dailyBegin) {
			res++
		}
	}
	return res
}

func (p *PlayerActivityStates) SetAct(acid string, nowT int64) {
	for i := 0; i < len(p.Players); i++ {
		if p.Players[i].AcId == acid {
			p.Players[i].LastActTime = nowT
			return
		}
	}

	p.Players[p.CurrNum].AcId = acid
	p.Players[p.CurrNum].LastActTime = nowT
	p.CurrNum++
}

func (p *PlayerActivityStates) OnMemKickedByPlayerActStat(acid string) {
	for i := 0; i < len(p.Players); i++ {
		if p.Players[i].AcId == acid {
			if len(p.Players) <= 1 {
				// 不会出现公会里一个人都没有的情况
				return
			}
			if p.CurrNum-1 != i {
				p.Players[i] = p.Players[p.CurrNum-1]
			}
			p.Players[p.CurrNum-1] = PlayerActivityState{}
			p.CurrNum--
			return
		}
	}
}
