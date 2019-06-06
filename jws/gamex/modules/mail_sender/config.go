package mail_sender

import (
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
)

var cfg mailhelper.MailConfig

func SetConfig(mc mailhelper.MailConfig) {
	cfg = mc
}
