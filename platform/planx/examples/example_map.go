package main

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers"
	// "time"
)

// map是引用类型的，发在结构体中，相互赋值会发生什么?
func main() {
	fmt.Println("...")
	a := servers.Request{
		Code:      "ABC",
		SessionID: "aaaaaa",
		Data: map[string]interface{}{
			"a": 1,
		},
	}
	b := a
	b.Code = "CBA"
	a.Data["a"] = 100
	b.Data["b"] = "22"
	fmt.Println(a, b)
}
