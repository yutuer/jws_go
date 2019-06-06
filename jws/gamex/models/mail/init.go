package mail

import (
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

var (
	db timail.Timail
)

func InitMail(mc mailhelper.MailConfig) error {
	var err error
	db, err = mailhelper.NewMailDriver(mc)
	return err
}
