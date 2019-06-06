package bot

import (

	//"sync"

	"os"

	//"github.com/astaxie/beego/httplib"

	"github.com/ugorji/go/codec"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func NewPlayerBot(account, gateserver, rpcAddr string, speed float64) *PlayerBot {
	bot := &PlayerBot{
		logEntry:    make(chan LogEntry),
		stopChan:    make(chan struct{}),
		speedFactor: speed,
		rpcAddr:     rpcAddr,
		gateSer:     gateserver,
	}

	//87d5f092-7bb7-44a5-9869-b42fd9bf5858 是随机的
	act, err := db.ParseAccount(account)
	if err != nil {
		logs.Error("NewPlayerBot with illegal format of accounts")
		return nil
	}
	bot.account = act
	bot.loginToken = uuid.NewV4().String()

	return bot
}

func encode(value interface{}) []byte {
	var out []byte
	enc := codec.NewEncoderBytes(&out, &mh)
	enc.Encode(value)
	return out
}

// 过滤哪些logEntry是后面流程感兴趣的
// 因为在Bot.Run中仍然需要针对不同事件做不同的判断，只能说减少了不必要的diff time和sleep
func checkLogEntry(le LogEntry) bool {
	switch le.LogType {
	case "record":
		if le.Prefix == "req" {
			return true
		}
	case "init", "session":
		return true
	}
	return false
}

func OSExit(code int) {
	logs.Close()
	os.Exit(code)
}
