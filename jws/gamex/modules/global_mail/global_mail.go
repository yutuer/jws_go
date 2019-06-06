package global_mail

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func genGlobalMailModule(sid uint) *GlobalMailModule {
	return &GlobalMailModule{
		sid: sid,
	}
}

type globalMailCommand struct {
	res_chan   chan<- []timail.MailReward
	createTime int64
	accountID  string
}

type globalMailReward struct {
	m        timail.MailReward
	accounts map[string]bool
}

type GlobalMailModule struct {
	sid                 uint
	mails               []globalMailReward
	mails_last_get_time int64

	db timail.Timail

	waitter sync.WaitGroup

	command_chan chan globalMailCommand
}

func (r *GlobalMailModule) AfterStart(g *gin.Engine) {
}

func (r *GlobalMailModule) BeforeStop() {
}

func (r *GlobalMailModule) copyMails(acID string, createTime int64) []timail.MailReward {
	res := make([]timail.MailReward, 0, len(r.mails))
	for _, m := range r.mails {
		if m.accounts != nil && len(m.accounts) > 0 {
			_, ok := m.accounts[acID]
			if !ok {
				continue
			}
		}
		if m.m.CreateBefore != 0 && createTime < m.m.CreateBefore {
			continue
		}

		if m.m.CreateEnd != 0 && m.m.CreateEnd < createTime {
			continue
		}

		res = append(res, m.m)
	}
	logs.Trace("copy %v", res)
	return res[:]
}

func (r *GlobalMailModule) reloadMails() {
	res, err := r.db.LoadAllMail(TableGlobalMailName(r.sid))
	if err != nil {
		logs.Error("LoadServerAllMail Err by %s", err.Error())
	}

	logs.Trace("reloadMails %v, sid:%d", res, r.sid)

	if res != nil {
		r.mails = make([]globalMailReward, 0, len(res))
		for i := 0; i < len(res); i++ {
			ng := globalMailReward{}
			ng.m = res[i]
			if ng.m.Account2Send != nil && len(ng.m.Account2Send) > 0 {
				ng.accounts = make(map[string]bool, len(ng.m.Account2Send))
				for _, acc := range ng.m.Account2Send {
					ng.accounts[acc] = true
				}
			}
			r.mails = append(r.mails, ng)
		}
	}
}

func (r *GlobalMailModule) Start() {
	r.waitter.Add(1)
	defer r.waitter.Done()
	var err error
	r.db, err = mailhelper.NewMailDriver(cfg)
	if err != nil {
		logs.Error("GlobalMailModule DB Open Err by %s", err.Error())
		return
	}

	r.mails = make([]globalMailReward, 0, 64)
	r.command_chan = make(chan globalMailCommand, 64)

	r.reloadMails()

	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		timerChan30 := uutil.TimerMS.After(30 * time.Second)
		for {
			select {
			case command, ok := <-r.command_chan:
				logs.Trace("command %v", command)

				if !ok {
					logs.Warn("command_chan close")
					return
				}
				if command.res_chan == nil {
					logs.Warn("command_chan command res_chann error")
				} else {
					command.res_chan <- r.copyMails(command.accountID, command.createTime)
				}
			case <-timerChan30:
				timerChan30 = uutil.TimerMS.After(30 * time.Second)
				r.waitter.Add(1)
				func() {
					defer r.waitter.Done()
					//logs.Info("reload All Mails")
					r.reloadMails()
				}()
			}
		}
	}()

}

func (r *GlobalMailModule) Stop() {
	close(r.command_chan)
	r.waitter.Wait()
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func GetGlobalMail(accID string, createTime int64) []timail.MailReward {
	a, _ := db.ParseAccount(accID)
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	errRet := []timail.MailReward{}
	res_chan := make(chan []timail.MailReward, 1) // 只会发一个信息给这个chann

	select {
	case GetModule(a.ShardId).command_chan <- globalMailCommand{
		res_chan:   res_chan,
		accountID:  accID,
		createTime: createTime,
	}:
	case <-ctx.Done():
		logs.Error("GetGlobalMail cmd put timeout, acid %s", accID)
		return errRet
	}

	select {
	case res := <-res_chan:
		return res
	case <-ctx.Done():
		logs.Error("GetGlobalMail  <-res_chan timeout")
		return errRet
	}
}
