package account

type playerRedeemCodeTypHasToken struct {
	BatchIDs []int64 `json:"bid"`
}

func (p *playerRedeemCodeTypHasToken) IsHasToken(bID int64) bool {
	for i := 0; i < len(p.BatchIDs); i++ {
		if p.BatchIDs[i] == bID {
			return true
		}
	}
	return false
}

func (p *playerRedeemCodeTypHasToken) SetHasToken(bID int64) {
	if p.BatchIDs == nil {
		p.BatchIDs = make([]int64, 0, 16)
	}
	p.BatchIDs = append(p.BatchIDs, bID)
}
