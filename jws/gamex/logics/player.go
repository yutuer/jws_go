package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logics/notify"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/crossservice/teamboss"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func CreatePlayer(accountid, ip string) game.Player {
	pr, err := NewPlayerResource(accountid)
	if err != nil {
		panic(fmt.Errorf("Prepare player error in GetMux: %s", err.Error()))
	}

	mux := servers.NewMux()
	NewGameData(mux)

	accountId := pr.AccountID()

	account := NewAccount(mux, accountId, ip) //mux, accountid
	logs.Trace("Account has been set up! %v", account)
	pr.SetMux(mux)
	pr.SetErrorChan(account.GetDBErrChan())
	pr.SetMsgChan(account.GetMsgChan())
	pr.SetPushNotify(account.GetPushChan())
	pr.SetOnExitFunc(
		func() {
			account.OnExit()
		})
	pr.account = account
	return pr
}

type ProfileSaver helper.DBProfile

const DBSAVER_MAX = 10

type PlayerResource struct {
	accountID      db.Account
	mux            *servers.Mux
	onExitCallFunc func()
	errNotify      <-chan error
	msgChan        <-chan servers.Request
	pushNotify     <-chan game.INotifySyncMsg
	account        *Account
}

func NewPlayerResource(account string) (*PlayerResource, error) {
	accountID, err := db.ParseAccount(account)
	if err != nil {
		return nil, err
	}
	return &PlayerResource{
		accountID: accountID,
	}, nil
}

//interface for game.Player
func (p *PlayerResource) AccountID() db.Account {
	return p.accountID
}

//interface for game.Player
func (p *PlayerResource) GetMux() *servers.Mux {
	return p.mux
}

func (p *PlayerResource) SetMux(m *servers.Mux) {
	p.mux = m
}

func (p *PlayerResource) SetMsgChan(ec <-chan servers.Request) {
	p.msgChan = ec
}

func (p *PlayerResource) SetErrorChan(ec <-chan error) {
	p.errNotify = ec
}

func (p *PlayerResource) ErrorNotify() <-chan error {
	return p.errNotify
}

func (p *PlayerResource) MsgChannel() <-chan servers.Request {
	return p.msgChan
}

func (p *PlayerResource) SetPushNotify(pc <-chan game.INotifySyncMsg) {
	p.pushNotify = pc
}

func (p *PlayerResource) PushNotify() <-chan game.INotifySyncMsg {
	return p.pushNotify
}

func (p *PlayerResource) SetOnExitFunc(onExitCallFunc func()) {
	p.onExitCallFunc = onExitCallFunc
}

func (p *PlayerResource) MkSyncNotifyInfo(im game.INotifySyncMsg) *servers.Response {
	switch im.(type) {
	case *notify.NotifySyncMsg:
		m := im.(*notify.NotifySyncMsg)
		resp := &SyncRespNotify{}
		resp.Init("push", im.GetAddr(), gamedata.GetHotDataVerCfg().Build, p.account)
		resp.FromSyncMsg(*m)
		resp.MkNotifyInfo(p.account)
		return rpcSuccess(resp)
	case *teamboss.Msg:
		m := im.(*teamboss.Msg)
		resp := &servers.Response{
			Code:     m.GetAddr(),
			RawBytes: encode(m),
		}
		return resp
	default:
		return nil
	}
	return nil
}

func (p *PlayerResource) OnExit() {
	if nil != p.account {
		p.account.onExit()
	}
	if p.onExitCallFunc != nil {
		p.onExitCallFunc()
	}
}
