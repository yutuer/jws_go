package main

import (
	"flag"
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	// "vcs.taiyouxi.net/platform/planx/util/logs"
)

func main() {
	fmt.Println("PlanX Game Auth Server, version:", game.GetVersion())
	listento := flag.String("listen", ":8444", "0.0.0.0:8444, :8444")
	flag.Parse()

	server := game.NewAuthServer(*listento)
	server.Start()
}
