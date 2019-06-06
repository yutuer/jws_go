package astrology

import "fmt"

//Soul 星魂
type Soul struct {
	SoulID string `json:"id,omitempty"`
	Count  uint32 `json:"c,omitempty"`
}

//GoString GoStringer interface
func (s *Soul) GoString() string {
	str := ""

	str += fmt.Sprintf("{SoulID:%v,Count:%v}", s.SoulID, s.Count)

	return str
}

func newSoul(id string) *Soul {
	return &Soul{
		SoulID: id,
		Count:  0,
	}
}
