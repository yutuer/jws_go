package astrology

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//SoulBag 星魂背包
type SoulBag struct {
	Souls []*Soul `json:"souls,omitempty"`

	mapSouls map[string]*Soul
}

//GoString GoStringer interface
func (b *SoulBag) GoString() string {
	str := ""

	str += fmt.Sprintf("{Souls:{")
	for _, soul := range b.Souls {
		str += fmt.Sprintf("%#v,", soul)
	}
	str += "}}"

	return str
}

func newSoulBag() *SoulBag {
	return &SoulBag{
		Souls:    []*Soul{},
		mapSouls: map[string]*Soul{},
	}
}

//afterLogin ..
func (b *SoulBag) afterLogin() {
	b.mapSouls = map[string]*Soul{}
	for _, soul := range b.Souls {
		b.mapSouls[soul.SoulID] = soul
	}
}

//AddSoul ..
func (b *SoulBag) AddSoul(id string, num uint32) {
	logs.Debug("[Astrology] SoulBag AddSoul, %v(%v)", id, num)

	soul, exist := b.mapSouls[id]
	if false == exist {
		soul = newSoul(id)
		b.mapSouls[id] = soul
		b.Souls = append(b.Souls, soul)

		logs.Debug("[Astrology] SoulBag AddSoul, newSoul:%#v", soul)
	}

	soul.Count += num
}

//SubSoul ..
func (b *SoulBag) SubSoul(id string, num uint32) bool {
	logs.Debug("[Astrology] SoulBag SubSoul, %v(%v)", id, num)

	soul, exist := b.mapSouls[id]
	if false == exist {
		logs.Debug("[Astrology] SoulBag SubSoul, no exist")
		return false
	}
	if soul.Count < num {
		logs.Debug("[Astrology] SoulBag SubSoul, count less, %v < %v", soul.Count, num)
		return false
	}
	soul.Count -= num
	return true
}

//GetSoul ..
func (b *SoulBag) GetSoul(id string) *Soul {
	return b.mapSouls[id]
}

//GetSouls ..
func (b *SoulBag) GetSouls() []*Soul {
	return b.Souls
}

//UpdateSouls ..
func (b *SoulBag) UpdateSouls() {
	clearIDs := []string{}
	for _, soul := range b.Souls {
		if 0 == soul.Count {
			clearIDs = append(clearIDs, soul.SoulID)
		}
	}

	for _, soulID := range clearIDs {
		for i := 0; i < len(b.Souls); i++ {
			if soulID == b.Souls[i].SoulID {
				if i+1 < len(b.Souls) {
					b.Souls = append(b.Souls[:i], b.Souls[i+1:]...)
				} else {
					b.Souls = b.Souls[:i]
				}
				delete(b.mapSouls, soulID)
				break
			}
		}
	}
}
