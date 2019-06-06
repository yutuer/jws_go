package main

import (
	"fmt"
	"time"
)

// 每个time.After都会生成Timer实例，如果在高负载核心大量使用会对GC带来过多无效Timer
// GC的时间也会变长

func main() {
	fmt.Println("...")
	i := 1
	timer := time.NewTimer(time.Second * 1)
	for {
		time.Sleep(time.Second * 5)

		fmt.Println("=")
		v := timer.Reset(time.Second * time.Duration(i))
		fmt.Println("=", v, len(timer.C))
		// for len(timer.C) > 0 {
		// 	<-timer.C
		// }
		for ii := range timer.C {
			fmt.Println("...", ii)
		}
		select {
		case <-timer.C:
			fmt.Println(i, len(timer.C))
			i += 2
		}
	}
}
