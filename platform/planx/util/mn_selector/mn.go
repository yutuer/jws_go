package mnSelector

import (
	"math/rand"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
)

/*
	随机算法参考TAOCP 3.4.2节中的算法S
	算法S是一次性生成结果，而我们需要分次随机，
	所以存储了算法S中的中间变量
*/

type MNSelectorState struct {
	Count          int64 `json:"c"`
	Space          int64 `json:"s"`
	TurnIdx        int64 `json:"ti"`
	isLastSelected bool
	DefaultNum     int64 `json:"dn"`
	DefaultSpace   int64 `json:"dm"`
}

//GoString GoStringer interface
func (m *MNSelectorState) GoString() string {
	str := ""

	str += fmt.Sprintf("{Count:%v,Space:%v,TurnIdx:%v,isLastSelected:%v,DefaultNum:%v,DefaultSpace:%v}",
		m.Count, m.Space, m.TurnIdx, m.isLastSelected, m.DefaultNum, m.DefaultSpace)

	return str
}

func (m *MNSelectorState) Init(num, space int64) {
	m.Reset(num, space)
	m.DefaultNum = num
	m.DefaultSpace = space
}

func (m *MNSelectorState) Reset(num, space int64) {
	m.Count = num
	m.Space = space
	m.TurnIdx = util.TiNowUnixNano()
}

func (m *MNSelectorState) IsNowNeedNewTurn() bool {
	return m.Space <= 0
}

func (m *MNSelectorState) LogicLog(acid, logTyp string, idx string) {
	// T2714 增加MN日志
	logiclog.Trace(acid, logTyp,
		struct {
			IsSpec bool   `json:"is_spec"`
			N_t    int64  `json:"N_t"`
			N_m    int64  `json:"n_M"`
			Time   int64  `json:"space_time"`
			Idx    string `json:"idx"`
		}{
			IsSpec: m.isLastSelected,
			N_t:    m.Count,
			N_m:    m.Space,
			Time:   m.TurnIdx,
			Idx:    idx,
		}, logTyp)
}

/*
	N/M M次必中N次，
	随机到true， M-- N--
	随机到false, M--
*/
func (m *MNSelectorState) Selector(rd *rand.Rand) (selected bool) {
	if m.Space <= 0 {
		m.Reset(m.DefaultNum, m.DefaultSpace)
		// 如果重置后仍然小于0，说明必不中
		if m.Space <= 0 {
			return false
		}
	}

	if m.Count <= 0 {
		selected = false
		m.Space -= 1
		return
	}

	selected = rd.Float64()*float64(m.Space) < float64(m.Count)

	if selected {
		m.Count -= 1
	}

	m.isLastSelected = selected
	m.Space -= 1
	return
}
