package tmp

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func PayFeedBack(a *account.Account, f1, f2, money float64) {
	logs.Debug("[AuthNotifyNewRole] try payfeedback %s %s", a.AccountID.String(), a.Profile.AccountName)
	req := httplib.Post(game.Cfg.AuthNotifyNewRoleUrl).SetTimeout(8*time.Second, 8*time.Second)
	req.Param("gid", fmt.Sprintf("%d", a.AccountID.GameId))
	req.Param("shardid", fmt.Sprintf("%d", a.AccountID.ShardId))
	req.Param("uid", a.AccountID.UserId.String())
	req.Param("newrole", fmt.Sprintf("%d", a.Profile.GetCorp().GetLvlInfo()))
	req.Param("PayFeedBack", "wantGet")

	ret := struct {
		Res            string
		HadGotFeedBack bool
	}{}
	err := req.ToJSON(&ret)
	if err != nil {
		logs.Error("[AuthNotifyNewRole try payfeedback] req.ToJson failed err with %v", err)
		return
	}

	rsp, _ := req.Response()
	defer rsp.Body.Close()

	if ret.Res != "ok" {
		logs.Error("[AuthNotifyNewRole try payfeedback] res failed with %v", ret)
		return
	}
	logs.Trace("[AuthNotifyNewRole try payfeedback] success %s", a.AccountID.String())

	if ret.HadGotFeedBack && (f1 > 0 || f2 > 0) { // 反钻邮件
		_payFeedBack(a, f1, f2, money)
		a.Profile.GotPayFeedBack = true
	}
}

func _payFeedBack(a *account.Account, f, s, money float64) {
	acid := a.AccountID.String()

	logs.Debug("PayFeedBack %s %v %v", acid, f, s)
	if f <= 0 && s <= 0 {
		return
	}
	// 发邮件
	var firstMoney, firstBack, secondMoney, secondBack int
	// 越南 同SDK帐号，只反1服，1倍充值2倍非充值
	if f > 0 {
		m := int(money)
		firstMoney = int(f)
		mail_sender.SendPayFeedBack_VN(acid, m, uint32(firstMoney), uint32(firstMoney*2))
	}
	// 港澳台
	//if f > 0 {
	//	firstMoney = int(f)
	//	a.Profile.Vip.AddRmbPoint(a, uint32(f), "payfeedback")
	//	mail_sender.SendPayFeedBack_HMT(acid, firstMoney*2)
	//}
	// 国服
	//cfg := gamedata.GetIAPConfig()
	//if f > 0 {
	//	firstMoney = int(f)
	//	b1 := int(f) * int(cfg.GetCTDrate())
	//	b2 := int(f * float64(cfg.GetCTDbonus()) * float64(cfg.GetCTDrate()))
	//	firstBack = b1 + b2
	//	mail_sender.SendPayFeedBack(acid, int(f), b1, b2, true)
	//}
	//if s > 0 {
	//	secondMoney = int(s)
	//	b1 := int(s) * int(cfg.GetCTDrate())
	//	b2 := int(s * float64(cfg.GetCTDbonus()) * float64(cfg.GetCTDrate()))
	//	secondBack = b1 + b2
	//	mail_sender.SendPayFeedBack(acid, int(s), b1, b2, false)
	//}
	logiclog.LogPayFeedBack(acid, a.Profile.AccountName, a.Profile.GetCurrAvatar(),
		a.Profile.GetCorp().GetLvlInfo(), a.Profile.ChannelId,
		firstMoney, firstBack, secondMoney, secondBack,
		func(last string) string { return a.Profile.GetLastSetCurLogicLog(last) }, "")
}
