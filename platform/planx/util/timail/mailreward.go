package timail

import (
	"encoding/json"
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type MailInJson struct {
	IDS    uint32   `json:"ids"`
	Param  []string `json:"param"`
	ItemId []string `json:"item"`
	Count  []uint32 `json:"count"`
	Reason string   `json:"reason"`
	Tag    string   `json:"tag"`
}

const (
	MailStateIdle       = iota
	MailStateSyncUpdate //玩家在线新生成的邮件==SyncToDynamoDB
	MailStateDel        //需要删除DelOnDynamoDB
)

type MailKey struct {
	Uid string `bson:"uid"`
	Idx int64  `bson:"t"`
}

//MailReward 数据结构，内存中的邮件数据结构
type MailReward struct {
	IdsID uint32
	Param []string

	ItemId []string
	Count  []uint32

	TimeBegin int64
	TimeEnd   int64

	Reason string
	Tag    string

	IsGetted bool
	IsRead   bool

	Idx   int64
	state int

	Account2Send []string
	CreateBefore int64
	CreateEnd    int64
}

func (m *MailReward) IsAvailable() bool {
	return (m.state != MailStateDel) && !m.IsGetted
}

func (m *MailReward) GetState() int {
	return m.state
}

func (m *MailReward) SetState(n int) {
	m.state = n
}

func (m *MailReward) IsNeedNoDelBeforeGetted() bool {
	return GetMailSendTyp(m.Idx) == Mail_Send_By_Sys ||
		GetMailSendTyp(m.Idx) == Mail_Send_By_AndroidIAP ||
		GetMailSendTyp(m.Idx) == Mail_Send_By_IOSIAP
}

func (m *MailReward) AddReward(item_id string, count uint32) {
	m.ItemId = append(m.ItemId, item_id)
	m.Count = append(m.Count, count)
}

func (m *MailReward) FormDB(mailDB *MailRes) {
	m_json := mailDB.Mail
	mail_info := MailInJson{}
	err := json.Unmarshal([]byte(m_json), &mail_info)
	if err != nil {
		logs.Error("LoadAllMail Err by MainInfo %s in %s", err.Error(), m_json)
		return
	}

	if mailDB.Acc2send != "" {
		m.Account2Send = strings.Split(mailDB.Acc2send, `;`)
	}

	m.CreateBefore = mailDB.Ctb2send
	m.CreateEnd = mailDB.Cte2send
	m.IdsID = mail_info.IDS
	m.Param = mail_info.Param
	m.Idx = mailDB.Id
	m.TimeBegin = mailDB.Begin
	m.TimeEnd = mailDB.End
	m.IsGetted = mailDB.IsGet
	m.IsRead = mailDB.IsRead
	m.Reason = mail_info.Reason
	m.Tag = mail_info.Tag

	for i, id := range mail_info.ItemId {
		if i >= len(mail_info.Count) {
			logs.Error("LoadAllMail Err by mail_info %s in %s", id, m_json)
			continue
		}
		m.AddReward(id, mail_info.Count[i])
	}
}

func (m *MailReward) ToDB(mailDB *MailRes) error {
	b, err := json.Marshal(MailInJson{
		IDS:    m.IdsID,
		Param:  m.Param,
		ItemId: m.ItemId,
		Count:  m.Count,
		Reason: m.Reason,
		Tag:    m.Tag,
	})
	if err != nil {
		return err
	}

	if m.Account2Send != nil && len(m.Account2Send) > 0 {
		mailDB.Acc2send = strings.Join(m.Account2Send[:], ";")
	}

	mailDB.Id = m.Idx
	mailDB.Begin = m.TimeBegin
	mailDB.End = m.TimeEnd
	mailDB.Mail = string(b)
	mailDB.IsGet = m.IsGetted
	mailDB.IsRead = m.IsRead
	mailDB.Ctb2send = m.CreateBefore
	mailDB.Cte2send = m.CreateEnd
	return nil
}
