package account

type MoneyCatInfo struct {
	MoneyCatTime  int64  `json:"money_cat_time" codec:"money_cat_time"`
	MoneyCatActId uint32 `json:"money_cat_act_id"`
}

func (e *MoneyCatInfo) GetMoneyCatTime() int64 {
	return e.MoneyCatTime
}
func (e *MoneyCatInfo) UpdateMoneyCatTime() {
	e.MoneyCatTime = e.MoneyCatTime + 1
}

func (e *MoneyCatInfo) UpdateMoneyCatActId(actId uint32) {
	e.MoneyCatActId = actId

}

func (e *MoneyCatInfo) SetMoneyCat2Zero() {
	e.MoneyCatTime = 0
}
