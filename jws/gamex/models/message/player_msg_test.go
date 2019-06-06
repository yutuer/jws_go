package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func initDriver() {
	//driver.SetupRedis("10.0.1.244:6379", 7, "", true)
	driver.SetupRedis(":6379", 7, "", true)
}

func TestSendMsg(t *testing.T) {
	initDriver()

	msg := PlayerMsg{}
	msg.Typ = 1
	msg.Params = []string{"1", "cc", "asdfadsf"}
	for i := 0; i < 100; i++ {
		msg.Params = append(msg.Params, "aaaaa")
		SendPlayerMsgs("1:10:test", "xxx", 50, msg)
	}

	msgs, _ := LoadPlayerMsgs("1:10:test", "xxx", 50)
	for i, m := range msgs {
		assert.True(t, i < 50)
		assert.NotEmpty(t, m)
		//logs.Info("msg %d --> %v", i, m)
	}

	logs.Close()
}
