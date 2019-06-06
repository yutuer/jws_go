package main

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/tiprotogen/command"
)

func main() {
	defer logs.Close()
	d, err := command.LoadFromFile("./test.json")
	if err != nil {
		logs.Error("err %s", err.Error())
		return
	}
	fmt.Println(d)
	command.GenMultipleCSharpCode("./", d)

	command.GenGolangCodes("./", "msg", d)
}
