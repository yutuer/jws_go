package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
)

// 单个武将全部羁绊信息
type HeroCompanionInfo struct {
	// 下面是240的老数据，250以后不用
	Companions   []Companion `json:"hcp_info"`
	CompanionNum int         `json:"hcp_num"` // 记录激活数量, 解决关平ID为0的情况

	// 250新的数据结构
	NewCompanions []NewCompanion `json:"hcp_new_info"`
	EvolveLevel   int            `json:"hcp_lv"` // 已经进阶的等级 初始等级0
}

type Companion struct {
	CompanionId int  `json:"hcp_hero_idx" codec:"hcp_hero_idx"` // heroIdx  !!!关平的武将ID是0
	Active      bool `json:"hcp_active" codec:"hcp_active"`     // 是否激活  该值在v250之后废弃掉
}

type NewCompanion struct {
	Id     int  `json:"hcp_id"`     // 聚义ID
	Active bool `json:"hcp_active"` // 该字段用于标志聚义在1级的时候是否是激活状态, 也就是只有1级的时候可能会为false, 其他等级都是true
	// 缓存数据
	companionId int // 聚义武将ID
	level       int // 等级
}

func (h *HeroCompanionInfo) GetCompanion(heroIdx, companionId int) *NewCompanion {
	h.TryInitFromConfig(heroIdx)
	for i, c := range h.NewCompanions {
		if c.GetCompanionId() == companionId {
			return &h.NewCompanions[i]
		}
	}
	return nil
}

func (h *HeroCompanionInfo) CanEvolve() int {
	// 3个条件  情缘武将 > 0 && 都激活 && 未达到等级上限
	if !h.hasCompanions() {
		return errCode.HeroCompanionNotOpen
	}
	for _, companion := range h.NewCompanions {
		if companion.GetLevel() < h.EvolveLevel {
			return errCode.HeroCompanionNotAllActive // 检查是否都激活
		}
	}
	if h.EvolveLevel >= int(gamedata.MaxCompanionLevel) {
		return errCode.HeroCompanionEvolveMaxLevel
	}
	return 0
}

func (h *HeroCompanionInfo) IncEvolveLevel() {
	h.EvolveLevel++
}

// 每个武将第一次开启这个功能的时候读取配置表初始化
func (h *HeroCompanionInfo) TryInitFromConfig(heroIdx int) {
	if h.NewCompanions == nil || len(h.NewCompanions) == 0 {
		dataArray := gamedata.GetAllActiveConfig(heroIdx, 1)
		h.NewCompanions = make([]NewCompanion, len(dataArray))
		for i, data := range dataArray {
			h.NewCompanions[i] = NewCompanion{
				Id:          int(data.Config.GetUniqueID()),
				Active:      false,
				companionId: data.CompanionIdx,
				level:       1,
			}
		}
	}
}

func (h *HeroCompanionInfo) GetAllCompanions(heroIdx int) []NewCompanion {
	h.TryInitFromConfig(heroIdx)
	return h.NewCompanions
}

func (h *HeroCompanionInfo) HasCompanions(heroIdx int) bool {
	h.TryInitFromConfig(heroIdx)
	return h.hasCompanions()
}

func (h *HeroCompanionInfo) hasCompanions() bool {
	return h.NewCompanions != nil && len(h.NewCompanions) > 0
}

func (nh *NewCompanion) GetLevel() int {
	if nh.level == 0 {
		nh.UpdateLevelAndCompanion()
	}
	if !nh.Active {
		return nh.level - 1
	} else {
		return nh.level
	}
}

func (nh *NewCompanion) GetCompanionId() int {
	if nh.level == 0 {
		nh.UpdateLevelAndCompanion()
	}
	return nh.companionId
}

func (nh *NewCompanion) UpdateLevelAndCompanion() {
	cfg := gamedata.GetCompanionActiveConfigById(nh.Id)
	if cfg != nil {
		nh.level = int(cfg.Config.GetRelationLevel())
		nh.companionId = cfg.CompanionIdx
	}
}
