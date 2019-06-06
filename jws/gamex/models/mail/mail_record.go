package mail

import (
	"encoding/json"
	"sort"

	"vcs.taiyouxi.net/platform/planx/util/timail"
)

type mailGettedList []int64

func (p mailGettedList) Len() int           { return len(p) }
func (p mailGettedList) Less(i, j int) bool { return p[i] < p[j] }
func (p mailGettedList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// mailRecord
// 邮件领取， 玩家已经领取或者邮件已经过期
type mailRecord struct {
	MailGetted mailGettedList `json:"get"` // 已领取的mail列表

	// mail接收截止时间,在这个时间之前的mail都算过期的,也就是已领取的
	MailGettedTimeout int64 `json:"t"`
}

func (m *mailRecord) isMailGot(id int64) bool {
	t := timail.GetTimeFromMailID(id)
	if t < m.MailGettedTimeout {
		return true
	}

	res := sort.Search(len(m.MailGetted),
		func(i int) bool { return m.MailGetted[i] >= id })
	return res < len(m.MailGetted) && m.MailGetted[res] == id
}

// MarkMailAsGot 标记邮件为玩家已经领取的状态
func (m *mailRecord) markMailAsGot(id int64) {
	m.MailGetted = append(m.MailGetted, id)
	sort.Sort(m.MailGetted)
	return
}

func (m *mailRecord) Load(json_str []byte) error {
	return json.Unmarshal(json_str, m)
}

func (m *mailRecord) Save() ([]byte, error) {
	return json.Marshal(*m)
}
