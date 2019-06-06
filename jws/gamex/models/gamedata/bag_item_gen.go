package gamedata

import (
	"encoding/json"
	"math/rand"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type BagItemData struct {
	Id              uint32
	TrickGroup      []string `json:"tricks"`
	TCJ             int32    `json:"TCJ"`     // 天地人
	JadeExp         uint32   `json:"JadeExp"` // 龙玉经验 TODO 测试删档时，要删掉
	WholeCharPieces uint32   `json:"wholeCharPiece"`
	IsWholeChar     uint32   `json:"isWholeChar"`
	binited         bool
}

func (b *BagItemData) IsNil() bool {
	return b.Id == 0 &&
		(b.TrickGroup == nil || len(b.TrickGroup) == 0) &&
		b.TCJ == 0 &&
		b.JadeExp == 0 &&
		b.WholeCharPieces == 0 && b.IsWholeChar == 0
}

func (b *BagItemData) IsInited() bool {
	return b.binited
}

func (b *BagItemData) ToData() (string, error) {
	d, err := json.Marshal(b)
	if err != nil {
		return "", err
	} else {
		return string(d), nil
	}
}

func (b *BagItemData) ToDataStr() string {
	if b.IsNil() {
		return ""
	}
	d, err := json.Marshal(b)
	if err != nil {
		return ""
	} else {
		return string(d)
	}
}

func (b *BagItemData) FromData(id uint32, data string) error {
	err := json.Unmarshal([]byte(data), b)
	b.Id = id
	if err == nil {
		b.binited = true
	}
	return err
}

func NewBagItemData() *BagItemData {
	return &BagItemData{
		TrickGroup: make([]string, 0, 8),
	}
}

// 生成装备随机属性
func MakeItemData(acid string, rd *rand.Rand, ItemID string) (re *BagItemData) {
	// 所有有Trick的道具都会有tricks项 形如{"tricks":[]}
	bd := NewBagItemData()
	re = bd
	// re, _ = bd.ToData() //目前就是 {"tricks":[]}]

	if isTCJ := IsTCJ(ItemID); isTCJ {
		if rd != nil {
			bd.TCJ = rd.Int31n(TCJ_COUNT) + 1
		} else {
			bd.TCJ = rand.Int31n(TCJ_COUNT) + 1
		}
	}

	// Trick Data
	id, is_need := GetEquipTrickRuleID(ItemID)
	//logs.Trace("GetEquipTrickRuleID %s %v", id, is_need)

	if is_need && id != "" {
		//logs.Trace("TrickRuleRand %s", id)

		// 先从1属性组中随处一个高阶属性, 再从0属性组中随机出两个低阶属性
		// 高阶属性要排在低阶之前, 属性可以重复
		randsets := GetTrickRandPool(id)
		if randsets == nil {
			logs.SentryLogicCritical(acid,
				"mkItemData TrickRuleRand Err : %s",
				id)
			return
		}

		if len(randsets) < MAX_Trick_RandSet_Typ {
			logs.SentryLogicCritical(acid,
				"mkItemData TrickRuleRand Err len err : %s",
				id)
			return
		}

		// 先从1属性组中随处一个高阶属性, 再从0属性组中随机出两个低阶属性
		// 高阶属性要排在低阶之前, 属性可以重复
		t0 := randsets[1].Rand(rd)
		t1 := randsets[0].Rand(rd)
		t2 := randsets[0].Rand(rd)

		// 这个要增加的话需要手工加。。。
		bd.TrickGroup = append(bd.TrickGroup, t0)
		bd.TrickGroup = append(bd.TrickGroup, t1)
		bd.TrickGroup = append(bd.TrickGroup, t2)
	}

	return
}
