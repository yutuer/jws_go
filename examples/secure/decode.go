package main

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/util/secure"
)

func main() {
	e := "OnAJxjQTxzArt_QRvl5r2oO8vmprvl-8xDM82_ewxDM8sWiTgzA0sV9="
	er, _ := secure.DefaultEncode.Decode64FromNet(e)
	fmt.Println(string(er))
}
