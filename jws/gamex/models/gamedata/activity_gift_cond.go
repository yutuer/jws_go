package gamedata

import ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"

const (
	ActTypInitLen  = 8
	ActGiftInitLen = 64
)

type ActivityGiftByCondition struct {
	ID     uint32
	Cond   Condition
	Reward givesData
	Desc   string
}

func (a *ActivityGiftByCondition) loadFromData(data *ProtobufGen.CDTGIFTVALUE) {
	a.ID = data.GetActivityID()
	for _, i := range data.GetGoalAward_Template() {
		a.Reward.AddItem(i.GetReward(), i.GetCount())
	}
	a.Desc = data.GetDesc()
	a.Cond = Condition{
		Ctyp:   data.GetFCType(),
		Param1: int64(data.GetFCValue()),
	}
}

func GetActivityGiftByTime() []ActivityGiftByCondition {
	return gdActivityGiftByTime[:]
}

func GetActivityGiftByCond() []ActivityGiftByConds {
	return gdActivityGiftByConds[:]
}
