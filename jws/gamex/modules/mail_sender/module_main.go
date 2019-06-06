package mail_sender

import (
	//"fmt"
	"sync"

	"github.com/gin-gonic/gin"

	"time"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
)

func genMailSenderModule(sid uint) *MailSenderModule {
	return &MailSenderModule{
		shardId: sid,
	}
}

type MailToUser struct {
	Mail timail.MailReward
	Typ  int64
	Uid  string
}

type MailSenderModule struct {
	shardId            uint
	waitter            sync.WaitGroup
	command_chan       chan MailToUser
	command_batch_chan chan mailForBatch
	quit_chan          chan struct{}
	tWheel             *timingwheel.TimingWheel
}

func (r *MailSenderModule) AfterStart(g *gin.Engine) {
}

func (r *MailSenderModule) BeforeStop() {
}

func (r *MailSenderModule) Start() {
	r.waitter.Add(1)
	defer r.waitter.Done()

	err := initMail()
	if err != nil {
		logs.Error("MailSenderModule DB Open Err by %s", err.Error())
		return
	}

	r.tWheel = timingwheel.NewTimingWheel(time.Second, 30)

	r.command_chan = make(chan MailToUser, timail.Mail_Id_Gen_Base)
	r.quit_chan = make(chan struct{}, 1)

	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		var addon timail.MailAddonCounter

		for command := range r.command_chan {
			func(cmd MailToUser) {
				defer logs.PanicCatcherWithInfo("MailSenderModule sendMailImp Panic")
				logs.Trace("MailSenderModule command %v", command)
				sendMailImp(game.GetNowTimeByOpenServer(r.shardId), addon.Get(), &cmd)

			}(command)
		}
		logs.Info("MailSenderModule command_chan close")
	}()

	r.command_batch_chan = make(chan mailForBatch, 10240)

	// 此goroutine关服不用关
	go func() {
		for {
			select {
			case mail, ok := <-r.command_batch_chan:
				if !ok {
					logs.Warn("MailSenderModule command_batch_chan close")
					return
				}
				func(m mailForBatch) {
					defer logs.PanicCatcherWithInfo("MailSenderModule mail2Cache Panic")
					// update id if none
					if m.Mail.Idx == 0 {
						m.Mail.Idx = timail.MkMailIdAuto(m.Typ, time.Now().Unix())
					}
					if err := mail2Cache(r.shardId, m, m.IsActivity); err != nil {
						logs.Error("MailSenderModule batch mail2Cache err: %s", err.Error())
					}
				}(mail)
			}
		}
	}()

	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		timerChan := r.tWheel.After(time.Second)
		for {
			select {
			case <-r.quit_chan:
				logs.Warn("MailSenderModule quit_chan close")
				return
			case <-timerChan:
				func() {
					defer logs.PanicCatcherWithInfo("MailSenderModule cache2DB Panic")
					if err := cache2DB(r.shardId); err != nil {
						logs.Error("MailSenderModule batch cache2DB err: %s", err.Error())
					}
				}()
				timerChan = r.tWheel.After(time.Second)
			}
		}
	}()
}

func (r *MailSenderModule) Stop() {
	r.tWheel.Stop()
	close(r.quit_chan)
	close(r.command_chan)
	r.waitter.Wait()
}

func (r *MailSenderModule) sendMailCmd(m *MailToUser) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	select {
	case r.command_chan <- *m:
	case <-ctx.Done():
		logs.Error("MailSenderModule sendMailCmd timeout %v", *m)
	}
}
