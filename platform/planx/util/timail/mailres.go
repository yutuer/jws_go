package timail

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type MailRes struct {
	Userid   string `bson:"uid"` // 主要为batchwrite用
	Id       int64  `bson:"t"`
	Mail     string `bson:"mail"`
	Acc2send string `bson:"acc2send,omitempty"` // 发送的账号,只有玩家id在此列表中才可以领取
	Ctb2send int64  `bson:"ctb2send"`           // 在此时间开始之后注册的账号可以领取 为0表示不限开始时间
	Cte2send int64  `bson:"cte2send"`           // 在此时间之前注册的账号可以领取 为0表示不限结束时间
	Begin    int64  `bson:"begint"`
	End      int64  `bson:"endt"`
	IsGet    bool   `bson:"isget,omitempty"`
	IsRead   bool   `bson:"isread,omitempty"`
	Reason   string `bson:"reason,omitempty"`
}

func (mr *MailRes) SetBeginEnd(b, e int64) {
	mr.Begin = b
	mr.End = e
}

func (mr *MailRes) SetID(id int64) {
	mr.Id = id
}

func (mr *MailRes) SetMail(m string) {
	mr.Mail = m
}

func (mr *MailRes) SetUserID(uid string) {
	mr.Userid = uid
}

func (mr *MailRes) SetIsGet(b bool) {
	mr.IsGet = b
}

func (mr *MailRes) SetIsRead(b bool) {
	mr.IsRead = b
}

func (mr *MailRes) SetAcc2send(acc string) {
	mr.Acc2send = acc
}

func (mr *MailRes) SetCtb2send(ctb int64) {
	mr.Ctb2send = ctb
}

func (mr *MailRes) SetCte2send(cte int64) {
	mr.Cte2send = cte
}

const GETTED_STR = "getted"
const READ_STR = "read"

func HelperBatchWriteMails(user_id []string, mails_add []MailReward) ([]MailRes, error) {
	if len(user_id) == 0 || len(mails_add) == 0 ||
		len(user_id) != len(mails_add) || len(user_id) > MAX_Mail_ONE_SEND {
		return nil, fmt.Errorf("BatchWriteMails param err")
	}

	sends := make([]MailRes, len(mails_add), len(mails_add))

	for i := 0; i < len(mails_add); i++ {
		mails_add[i].ToDB(&(sends[i]))
		sends[i].SetUserID(user_id[i])
	}
	return sends, nil
}

func HelperBatchDupDrop(mail_add []MailRes) map[MailKey]MailRes {
	mmail := make(map[MailKey]MailRes, len(mail_add))

	for _, mail := range mail_add {
		k := MailKey{mail.Userid, mail.Id}
		if _, has := mmail[k]; has {
			logs.Error("HelperBatchDupDrop mail duplicate %v", mail)
			continue
		}
		mmail[k] = mail
	}
	return mmail
}
