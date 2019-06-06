package gacha

import (
	"vcs.taiyouxi.net/jws/gamex/logics"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/tools/dataChecker/utils"
)

var (
	repeatTimes = 100000
	hitRTimes   = 10000
	acc         *logics.Account
	reporter    *utils.Reporter
)

func init() {
	reporter = utils.NewReporter()

	// 生成用于测试Gacha的账号
	account.InitDebuger()
	account.Debuger.UnlockAllHero()
	a := account.Debuger.GetTestAccount()

	acc = new(logics.Account)
	acc.Account = &a
}

func RunAll() {
	GetAllNormalGachaLoot()
	GetGachaLootDetails(12)
}
