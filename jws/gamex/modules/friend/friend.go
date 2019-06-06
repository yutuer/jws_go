package friend

import (
	"time"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type friendModule struct {
	sid         uint
	friendCache friendCache
	GiftInfo    GiftInfo
	waitter     util.WaitGroupWrapper
	GiftChan    chan GiftCmd
	timeChan    <-chan time.Time
}

func (m *friendModule) Start() {
	m.Init()
	m.InitGiftInfo()
	m.waitter.Wrap(func() {
		for {
			select {
			case cmd, ok := <-m.GiftChan:
				if !ok {
					logs.Info("gift channel close")
					return
				}
				func() {
					defer logs.PanicCatcherWithInfo("friend command fatal error")
					m.handleGiftCmd(&cmd)
				}()
			case <-m.timeChan:
				m.Save()
				m.initTimer()
			}
		}
	})
}

func (m *friendModule) AfterStart(g *gin.Engine) {
}

func (m *friendModule) BeforeStop() {
}

func (m *friendModule) Stop() {
	close(m.GiftChan)
	m.waitter.Wait()
	m.Save()
}

func (m *friendModule) Save() {
	err := m.GiftInfo.saveGiftInfo(m.sid)
	if err != nil {
		logs.Error("friend module save db err by %v", err)
	}
	logs.Info("save friend info success")
}
