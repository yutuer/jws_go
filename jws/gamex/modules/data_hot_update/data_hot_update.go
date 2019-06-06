package data_hot_update

import (
	"sync"

	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	instance = &DataHotUpdateModule{
		startOnce:      &sync.Once{},
		afterStartOnce: &sync.Once{},
		stopOnce:       &sync.Once{},
		cmd_chan:       make(chan struct{}, 128),
	}
)

func genDataHotUpdateModule(sid uint) *DataHotUpdateModule {
	return instance
}

type DataHotUpdateModule struct {
	startOnce      *sync.Once
	afterStartOnce *sync.Once
	stopOnce       *sync.Once
	cmd_chan       chan struct{}
	waitter        util.WaitGroupWrapper
}

func (m *DataHotUpdateModule) Start() {
	m.startOnce.Do(func() {
		//
		m.waitter.Wrap(func() {
			for _ = range m.cmd_chan {
				func() {
					//by YZH 这个让parent never dead, 应该如此吗？
					defer logs.PanicCatcherWithInfo("DataHotUpdate Worker Panic")

					if err := uutil.LoadHotData2LocalFromS3(
						game.Cfg.HotDataS3Bucket(),
						game.Cfg.GetHotDataVerC(),
						gamedata.GetHotDataPath()); err != nil {
						logs.Error("DataHotUpdateModule LoadHotData2LocalFromS3 err %s", err.Error())
						return
					}

					// 数据热更
					logs.Info("DataHotUpdateModule data update build %s", game.Cfg.GetHotDataVerC())
					if err := gamedata.LoadHotGameDataFromUpdate(
						gamedata.GetHotDataPath(),
						gamedata.GetHotDataRelPath()); err != nil {
						logs.Error("DataHotUpdateModule LoadHotGameDataFromUpdate err %s", err.Error())
						return
					}
				}()
			}
		})
	})
}

func (m *DataHotUpdateModule) AfterStart(g *gin.Engine) {
	m.afterStartOnce.Do(func() {
		g.GET(game.Cfg.HotDataUrl, func(c *gin.Context) {
			logs.Info("DataHotUpdate rec signal ...")

			if err := gamedata.LoadHotDataVerFromEtcd(); err != nil {
				c.String(http.StatusNotFound, "ok")
				return
			}
			if game.Cfg.IsHotDataValid() {
				if !m.CommandExec() {
					c.String(http.StatusRequestTimeout, "timeout")
					return
				}
			}
			c.String(http.StatusOK, "ok")
		})
	})
}

func (m *DataHotUpdateModule) BeforeStop() {
}

func (m *DataHotUpdateModule) Stop() {
	m.stopOnce.Do(func() {
		close(m.cmd_chan)
		m.waitter.Wait()
	})
}

func (m *DataHotUpdateModule) CommandExec() bool {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	chann := m.cmd_chan
	select {
	case chann <- struct{}{}:
	case <-ctx.Done():
		logs.Error("DataHotUpdateModule CommandExec chann full, cmd put timeout")
		return false
	}
	return true
}
