package mbot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"

	"vcs.taiyouxi.net/botx/bot"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

var _melody *melody.Melody
var (
	mu       sync.RWMutex
	sessions map[*melody.Session]*bot.PlayerBot
)

func init() {
	sessions = make(map[*melody.Session]*bot.PlayerBot)
	m := melody.New()
	size := 65536
	m.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  size,
		WriteBufferSize: size,
	}
	m.Config.MaxMessageSize = int64(size)
	m.Config.MessageBufferSize = 2048
	_melody = m

	m.HandleConnect(locustAddBot)
	m.HandleMessage(locustMessage)
	m.HandleDisconnect(locustStopBot)
}

func websocket_handler(r *gin.Engine) {
	r.GET("/ws", func(c *gin.Context) {
		_melody.HandleRequest(c.Writer, c.Request)
	})
}

type Session struct {
	*melody.Session
}

func (s *Session) Fire(t, e string) {
	raw, _ := json.Marshal(
		struct {
			Type string
			Name string
		}{
			Type: t,
			Name: e,
		})
	go func(data []byte) {
		defer func() {
			//因为无法准确断定Session的结束与否
			//，所以这里做一个简单的处理
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		s.Session.Write(data)
	}(raw)
}

func locustAddBot(s *melody.Session) {
	logsfiles := _bf.GetLogs()
	nlogfiles := len(logsfiles)
	idx := rand.Intn(nlogfiles)
	id := logsfiles[idx]
	bmi := BotMakeInfo{
		Identify: id,
		Server:   _bf.server,
		Rpc:      _bf.rpc,
		Speed:    _bf.speedParam,
	}

	bmi.Account = _bf.nextAccount
	_bf.nextAccount.UserId = db.NewUserID()

	mker, err := _bf.getMaker(id)
	if err != nil {
		logs.Error("bot %s start error, %s", bmi.Account.String(), err.Error())
		s.Close()
		panic("getmaker failed")
	}
	pbot := mker.RunABot(bmi, &Session{s})

	mu.Lock()
	defer mu.Unlock()
	sessions[s] = pbot
}

func locustStopBot(s *melody.Session) {
	mu.Lock()
	defer mu.Unlock()
	//logs.Info("locustStopBot")
	if pbot, ok := sessions[s]; ok {
		//logs.Info("locustStopBot2")
		pbot.Stop()
		delete(sessions, s)
	}

}

func locustMessage(s *melody.Session, msg []byte) {
}
