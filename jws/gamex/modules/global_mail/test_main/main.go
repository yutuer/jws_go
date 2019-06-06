package main

import (
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/global_mail"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/signalhandler"
)

func main() {
	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)
	//For Mails
	global_mail.SetConfig(mailhelper.MailConfig{
		AWSRegion:    "cn-north-1",
		DBName:       "Mail_prod",
		AWSAccessKey: "AKIAO4YSP5CZU5CDQS4A",
		AWSSecretKey: "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",
	})

	modules.StartModule([]uint{1000})

	var waitGroup util.WaitGroupWrapper
	//handle kill signal
	waitGroup.Wrap(func() { signalhandler.SignalKillHandle() })
	waitGroup.Wait()

	// modules需要Stop做一些收尾工作
	modules.StopModule()
}
