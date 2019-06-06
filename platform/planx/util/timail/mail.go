package timail

type Timail interface {
	LoadAllMail(user_id string) ([]MailReward, error)
	LoadAllMailByGM(user_id string) ([]MailReward, error)

	SendMail(user_id string, mr MailReward) error
	SyncMail(user_id string, mails_add []MailReward, mails_del []int64) error
	BatchWriteMails(user_id []string, mails_add []MailReward) (error, []MailKey) // 返回未发送成功的邮件
	MailExist(user_id string, idx int64) (bool, error)
	Close() error
}
