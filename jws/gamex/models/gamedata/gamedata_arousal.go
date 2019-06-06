package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type AvatarArousalLevelData struct {
	AvatarLvNeedByThisLevel []uint32   // 觉醒到idx等级的角色等级要求
	CostToThisLevel         []CostData // 觉醒到idx等级要消耗的物品
	Attack                  []float32  // 攻击力加成
	Defense                 []float32  // 防御力加成
	Hp                      []float32  // 武器生命值加成
	CritRate                []float32  // 暴击率加成
	CritValue               []float32  // 暴击伤害加成
}

func (a *AvatarArousalLevelData) AddCost(lv int, typ string, num uint32) {
	if a.CostToThisLevel == nil {
		a.CostToThisLevel = make([]CostData, 0, 32)
	}
	for len(a.CostToThisLevel) <= lv {
		a.CostToThisLevel = append(a.CostToThisLevel, CostData{})
		a.AvatarLvNeedByThisLevel = append(a.AvatarLvNeedByThisLevel, 0)
		a.Attack = append(a.Attack, 0)
		a.Defense = append(a.Defense, 0)
		a.Hp = append(a.Hp, 0)
		a.CritRate = append(a.CritRate, 0)
		a.CritValue = append(a.CritValue, 0)
	}
	a.CostToThisLevel[lv].AddItem(typ, num)
}

var (
	gdAvatarArousalData [AVATAR_NUM_MAX]AvatarArousalLevelData
)

func GetAvatarArousalData(avatar_id int) *AvatarArousalLevelData {
	if avatar_id < 0 || avatar_id >= AVATAR_NUM_MAX {
		return nil
	} else {
		return &gdAvatarArousalData[avatar_id]
	}
}

func loadAvatarArousalConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.AROUSAL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	for _, a := range dataList.GetItems() {
		avatar_id := int(a.GetRole())
		level := int(a.GetArousalLevel())

		coin_t := a.GetArousalCoin()
		coin_v := a.GetArousalCost()

		if avatar_id < 0 || avatar_id > AVATAR_NUM_CURR || avatar_id > AVATAR_NUM_MAX {
			logs.Error("avatar_id %d in loadAvatarArousalConfig Err", avatar_id)
			continue
		}

		gdAvatarArousalData[avatar_id].AddCost(level, coin_t, coin_v)
		gdAvatarArousalData[avatar_id].AvatarLvNeedByThisLevel[level] = a.GetLevelLimit()

		gdAvatarArousalData[avatar_id].Attack[level] = a.GetAttack()
		gdAvatarArousalData[avatar_id].Defense[level] = a.GetDefense()
		gdAvatarArousalData[avatar_id].Hp[level] = a.GetHp()
		gdAvatarArousalData[avatar_id].CritRate[level] = a.GetCritRate()
		gdAvatarArousalData[avatar_id].CritValue[level] = a.GetCritValue()

		for _, m := range a.GetMatierial_Table() {
			gdAvatarArousalData[avatar_id].AddCost(
				level,
				m.GetMatierialID(),
				m.GetMatierialCount())
		}
	}
	//logs.Trace("loadAvatarArousalConfig %v", gdAvatarArousalData)

}
