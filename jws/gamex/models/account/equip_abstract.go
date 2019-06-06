package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/mn_selector"
)

type abstractInfo struct {
	AimEquipId    uint32
	AimEquipTrick string
	AimEquipIdx   int
}

type PlayerAbstractCancelInfo struct {
	info          map[uint32]abstractInfo
	AbstractCount int                        `json:"abstractCount"` // 洗练**累计**次数
	MN            mnSelector.MNSelectorState `json:"mn"`
}

func (p *PlayerAbstractCancelInfo) Add(equip_id uint32, trick string, idx int) {
	if p.info == nil {
		p.info = make(map[uint32]abstractInfo, 16)
	}
	p.info[equip_id] = abstractInfo{
		AimEquipId:    equip_id,
		AimEquipTrick: trick,
		AimEquipIdx:   idx,
	}
}

func (p *PlayerAbstractCancelInfo) Get(equip_id uint32) *abstractInfo {
	if p.info == nil {
		return nil
	}

	res, ok := p.info[equip_id]
	if !ok {
		return nil
	} else {
		return &res
	}
}

func (p *PlayerAbstractCancelInfo) Clean(equip_id uint32) {
	if p.info == nil {
		return
	}
	delete(p.info, equip_id)
}

func (p *PlayerAbstractCancelInfo) AddCount() {
	p.AbstractCount++
}

func (p *PlayerAbstractCancelInfo) GetNumAndSpace() (int64, int64) {
	tData := gamedata.GetEquipTrickSelectData(p.AbstractCount)
	return tData.Num, tData.Space
}
