package logics

import (
	"fmt"
	"golang.org/x/net/context"
	"time"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/worship"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) worship(r servers.Request) *servers.Response {
	req := &struct {
		Req
		AccountID string `codec:"accid"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}
	initReqRsp(
		"Attr/WorshipRspMsg",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		ErrHasWorship
		ErrGiveErr
	)

	if !p.Profile.GetCounts().Use(counter.CounterTypeWorshipTimes, p) {
		resp.mkInfo(p)
		return rpcSuccess(resp)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second)
	defer cancel()
	worship.Get(p.AccountID.ShardId).Worship(ctx, req.AccountID)

	gives, err := p.GetGivesByTemplate(
		gamedata.GetCommonCfg().GetWorshipLoot())

	if err != nil {
		logs.Error("Worship getLootByTemplate Err")
		return rpcErrorWithMsg(
			resp,
			ErrGiveErr,
			fmt.Sprintf(
				"Worship getLootByTemplate Err by %v",
				gamedata.GetCommonCfg().GetWorshipLoot()))
	}

	ok := account.GiveBySync(p.Account, gives.Gives(), resp, "Worship")
	if !ok {
		logs.Error("Worship GiveBySync Err")
		return rpcErrorWithMsg(
			resp,
			ErrGiveErr,
			fmt.Sprintf(
				"Worship GiveBySync Err by %v",
				gives))
	}

	resp.OnChangeGameMode(gamedata.CounterTypeWorshipTimes)
	resp.mkInfo(p)
	return rpcSuccess(resp)
}
