package account

const ExclusiveWeaponMaxAttr = 7

type HeroExclusiveWeapon struct {
	IsActive     bool                            `json:"is_active" codec:"is_active"`
	Quality      int                             `json:"quality" codec:"quality"`
	PromoteCount int                             `json:"promot_c" codec:"promot_c"`
	Attr         [ExclusiveWeaponMaxAttr]float32 `json:"attr" codec:"attr"` // 属性 @avatar_attr.go 赋值的时候下标实际-1
	ExtraAttr    [ExclusiveWeaponMaxAttr]float32 `json:"extra_attr" codec:"extra_attr"`
	ExtraHasAttr [ExclusiveWeaponMaxAttr]bool    `json:"extra_has_attr" codec:"extra_has_attr"`
}

func (h *HeroExclusiveWeapon) OnActivate() {
	h.IsActive = true
	h.Quality = 1
}

func (h *HeroExclusiveWeapon) Clear() {
	for i := range h.Attr {
		h.ExtraAttr[i] = 0
		h.ExtraHasAttr[i] = false
	}
}

func (h *HeroExclusiveWeapon) Save() {
	for i := range h.Attr {
		h.Attr[i] += h.ExtraAttr[i]
	}
}

func (h *HeroExclusiveWeapon) CanSave() bool {
	for i := range h.Attr {
		if h.ExtraHasAttr[i] && h.ExtraAttr[i] > 0 {
			return true
		}
	}
	return false
}
