package core

import (
	"sync"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
)

const (
	HeroFlag      = "hero"
	AlibabaUCFlag = "alibaba_uc"
	HMT           = "hmt_gift"
)

var (
	mailDB  map[string]timail.Timail
	Waitter sync.WaitGroup
)

func GetTiMail(gidStr string) timail.Timail {
	return mailDB[gidStr]
}

func InitDynamoDB(mailDBName map[string]config.MailDBName) {
	mailDB = make(map[string]timail.Timail, 0)
	for k, v := range mailDBName {
		if db, err := mailhelper.NewMailDriver(mailhelper.MailConfig{
			AWSRegion:    config.CommonConfig.AWS_Region,
			DBName:       v.Name,
			AWSAccessKey: config.CommonConfig.AWS_AccessKey,
			AWSSecretKey: config.CommonConfig.AWS_SecretKey,
			DBDriver:     "DynamoDB",
		}); err != nil {
			logs.Error("init dynamodb err by %v for config: %v, %v", err, k, v)
		} else {
			mailDB[k] = db
		}
	}
}

type Handle interface {
	GenError()
	Reg(e *gin.Engine, cfg config.Config) error
}

type Handler struct {
	ErrorInfo map[int]string
}

func (h *Handler) GetMsg(code int) string {
	return h.ErrorInfo[code]
}
