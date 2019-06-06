package dynamodb

import (

	//"github.com/cenkalti/backoff"
	//"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	. "vcs.taiyouxi.net/platform/planx/util/timail"
)

type MailDynamoDB struct {
	db        *DynamoDB
	region    string
	db_name   string
	accessKey string
	secretKey string

	is_inited bool
}

func NewMailDynamoDB(region, db_name, accessKey, secretKey string) MailDynamoDB {
	db := &DynamoDB{}

	return MailDynamoDB{
		db:        db,
		region:    region,
		db_name:   db_name,
		accessKey: accessKey,
		secretKey: secretKey,
	}
}

func (s *MailDynamoDB) Clone() MailDynamoDB {
	db := &DynamoDB{}
	return MailDynamoDB{
		db:        db,
		region:    s.region,
		db_name:   s.db_name,
		accessKey: s.accessKey,
		secretKey: s.secretKey,
	}
}

func (s *MailDynamoDB) Open() error {
	db, err := DynamoConnectInitFromPool(s.db,
		s.region,
		[]string{s.db_name},
		s.accessKey,
		s.secretKey,
		"")

	if err != nil {
		return err
	}
	s.db = db
	s.is_inited = true

	return nil
}

func (s *MailDynamoDB) Close() error {
	return nil
}

func (s *MailDynamoDB) IsHasInited() bool {
	return s.is_inited
}

func (s *MailDynamoDB) LoadAllMail(user_id string) ([]MailReward, error) {
	res, err := s.db.QueryMail(s.db_name, user_id, 0, MkMailId(0, 0), MAX_Mail_ONE_GET)
	re := make([]MailReward, 0, len(res)+1)
	if err != nil {
		return re[:], err
	}
	for _, mail_from_db := range res {
		new_mail := MailReward{}
		new_mail.FormDB(&mail_from_db)
		re = append(re, new_mail)
	}

	return re, nil
}

func (s *MailDynamoDB) SendMail(user_id string, mr MailReward) error {
	var m MailRes
	if err := mr.ToDB(&m); err != nil {
		return err
	}
	return s.db.SendMail(s.db_name, user_id, m)
}

func (s *MailDynamoDB) SyncMail(user_id string, mails_add []MailReward, mails_del []int64) error {
	if len(mails_add) == 0 && len(mails_del) == 0 {
		return nil
	}
	if (len(mails_add) + len(mails_del)) > MAX_Mail_ONE_SEND {
		if len(mails_add) > 0 {
			next_idx := 0
			for {
				if next_idx >= len(mails_add) {
					break
				}
				sends := make([]MailRes, 0, MAX_Mail_ONE_SEND)
				for i := 0; i < MAX_Mail_ONE_SEND && next_idx < len(mails_add); i++ {
					m := &MailRes{}
					mails_add[next_idx].ToDB(m)
					sends = append(sends, *m)
					next_idx++
				}
				if err := s.db.SyncMail(s.db_name,
					user_id, sends, []int64{}); err != nil {
					logs.Error("SyncMail len(mails_add) %d err %s", len(mails_add), err.Error())
					return err
				}
			}
		}
		if len(mails_del) > 0 {
			next_idx := 0
			for {
				if next_idx >= len(mails_del) {
					break
				}
				sends := make([]int64, 0, MAX_Mail_ONE_SEND)
				for i := 0; i < MAX_Mail_ONE_SEND && next_idx < len(mails_del); i++ {
					sends = append(sends, mails_del[next_idx])
					next_idx++
				}
				if err := s.db.SyncMail(s.db_name,
					user_id, []MailRes{}, sends); err != nil {
					logs.Error("SyncMail len(mails_del) %d err %s", len(mails_del), err.Error())
					return err
				}
			}
		}
		return nil
	} else {
		// 生成mail id
		sends := make([]MailRes, len(mails_add), len(mails_add))

		for i := 0; i < len(mails_add) && i < MAX_Mail_ONE_SEND; i++ {
			mails_add[i].ToDB(&(sends[i]))
		}

		return s.db.SyncMail(s.db_name,
			user_id,
			sends,
			mails_del)
	}
}

func (s *MailDynamoDB) BatchWriteMails(user_id []string, mails_add []MailReward) (error, []MailKey) {
	sends, err := HelperBatchWriteMails(user_id, mails_add)
	if err != nil {
		return err, nil
	}

	return s.db.BatchSendMail(s.db_name, sends)
}

func (s *MailDynamoDB) MailExist(user_id string, idx int64) (bool, error) {
	res, err := s.db.Get(s.db_name, user_id, idx)
	if err != nil {
		return false, err
	}

	if res == nil || len(res) <= 0 {
		return false, nil
	}
	return true, nil
}
