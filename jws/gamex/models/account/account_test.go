package account

import (
	"os"
	"testing"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

// TestMain 用于初始化account目录下所有unittest
func TestMain(m *testing.M) {
	InitDebuger()
	gamedata.DebugLoadLocalGamedata()

	retCode := m.Run()

	os.Exit(retCode)
}
