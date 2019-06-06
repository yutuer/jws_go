package account

type EatBaoziInfo struct {
	MaxEatCount uint32 `json:"max_eat_count codec: "max_eat_count"`
}

func (e *EatBaoziInfo) GetMaxEatBaoziCount() uint32 {
	return e.MaxEatCount
}
func (e *EatBaoziInfo) UpdateMaxEatBaoziCount(newCount uint32) {
	if e.MaxEatCount < newCount {
		e.MaxEatCount = newCount
	}
}
