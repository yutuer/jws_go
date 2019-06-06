package main

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func main() {

	fmt.Println("Hello World!")
	logs.Trace("trace")
	logs.Info("info")
	logs.Warn("warning")
	logs.Debug("debug")
	logs.Critical("critical, %d", 123)
	util.Exit(0)
}
