package account

import (
	"fmt"
	"testing"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

func TestPackage(t *testing.T) {
	account := Debuger.GetNewAccount()

	giveData := &gamedata.CostData{}
	giveData.IAPChannel = "debug"
	giveData.IAPGoodIndex = 125
	giveData.IAPPkgInfo.PkgId = 1
	giveData.IAPPkgInfo.SubPkgId = 3
	giveData.AddItem("VI_EN", 50)
	giveData.AddItem("VI_GT", 50)
	reason := "test"

	if !GiveBySync(account, giveData, nil, reason) {
		fmt.Println("Give By Sync Fail")
	}
}
