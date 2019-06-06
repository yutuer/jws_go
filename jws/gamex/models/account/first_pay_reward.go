package account

type PlayerFirstPayReward struct {
	FirstPayReward []uint32 `json:"fp"`
}

func (fp *PlayerFirstPayReward) onAfterLogin() {
	if fp.FirstPayReward == nil {
		fp.FirstPayReward = make([]uint32, 0, 5)
	}
}

func (fp *PlayerFirstPayReward) HadGot(id uint32) bool {
	for _, idx := range fp.FirstPayReward {
		if id == idx {
			return true
		}
	}
	return false
}

func (fp *PlayerFirstPayReward) GotReward(id uint32) {
	fp.FirstPayReward = append(fp.FirstPayReward, id)
}
