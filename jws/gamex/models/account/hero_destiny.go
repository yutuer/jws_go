package account

type HeroDestiny struct {
	ActivateDestiny    []int         `json:"activate_destiny"`     // 340前的版本
	NewActivateDestiny []DestinyInfo `json:"new_activate_destiny"` // 340以后的版本
}

type DestinyInfo struct {
	Id    int `json:"id"`
	Level int `json:"sub_id"`
}

func (h *HeroDestiny) GetHeroDestinyAllLvl() int {
	var count int
	for _, value := range h.NewActivateDestiny {
		count += value.Level
	}
	return count
}

func (h *HeroDestiny) GetActivateDestiny() []DestinyInfo {
	if h.NewActivateDestiny == nil {
		h.NewActivateDestiny = make([]DestinyInfo, 0)
	}
	return h.NewActivateDestiny
}

func (h *HeroDestiny) GetHeroDestinyById(id int) *DestinyInfo {
	for i, des := range h.GetActivateDestiny() {
		if des.Id == id {
			return &h.NewActivateDestiny[i]
		}
	}
	return nil
}

func (h *HeroDestiny) AddOrUpdate(id int, level int) {
	destinyList := h.GetActivateDestiny()
	for i, des := range destinyList {
		if des.Id == id {
			destinyList[i].Level = level
			return
		}
	}
	h.NewActivateDestiny = append(h.NewActivateDestiny, DestinyInfo{
		Id:    id,
		Level: level,
	})
}
