package main

import (
	"fmt"
	"time"
)

// GODEBUG="gctrace=1" go run example_gc_chan.go |grep gc
// chan 的释放是GC托管
// 参考 http://stackoverflow.com/questions/8593645/is-it-ok-to-leave-a-channel-open

func main() {
	fmt.Println("...")
	i := 0
	for {
		select {
		case <-time.After(time.Second * 1):
			fmt.Println(i)
			i += 1
			c := make(chan bool, 1024)
			c <- true
		}
	}
}
