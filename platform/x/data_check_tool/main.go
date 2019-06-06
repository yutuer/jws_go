package main

import (
	"os"

	"fmt"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func main() {
	if len(os.Args) < 2 {
		logs.Error("os.Args Err %v, Use By ./tool {path_to_data}", os.Args)
		return
	}

	rootAddress := os.Args[1]
	if len(os.Args) == 3 {
		v := os.Args[2]
		if v != "v" {
			logs.Close()
		}
	} else {
		logs.Close()
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Errorf("LoadGameData Panic For %v", err)
			os.Exit(1)
		}
	}()
	gamedata.LoadGameData(rootAddress)
	logs.Close()
	os.Exit(0)
}
