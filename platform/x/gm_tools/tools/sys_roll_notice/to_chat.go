package sys_roll_notice

import (
	"encoding/json"
)

type NoticeParam struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type SysRollNotice struct {
	MsgId  int32         `json:"msgId"`
	KV     []NoticeParam `json:"info"`
	LangKV []NoticeParam `json:"lang"`
}

type SysRollNoticeToChat struct {
	Typ string `json:"type"`
	Shd string `json:"shd"`
	Msg string `json:"msg"`
	msg SysRollNotice
}

func NewSysRollNoticeToChat(title, server string, msg []string) *SysRollNoticeToChat {
	re := &SysRollNoticeToChat{}
	re.Typ = "SysRollNotice"
	re.Shd = server
	re.msg.MsgId = 0
	re.msg.KV = []NoticeParam{{"0", msg[0]}}
	if len(msg) > 3 && msg[2] == "1" {
		re.msg.LangKV = []NoticeParam{{"0", msg[3]}}
	}
	return re
}

func (s *SysRollNoticeToChat) ToData() ([]byte, error) {
	j1, err := json.Marshal(s.msg)
	if err != nil {
		return nil, err
	}

	s.Msg = string(j1)

	j2, err := json.Marshal(*s)
	if err != nil {
		return nil, err
	}

	return j2, nil
}
