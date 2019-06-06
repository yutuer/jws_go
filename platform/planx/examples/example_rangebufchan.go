package main

import (
	"runtime"

	"fmt"
)

// range buffered chan is multi goroutine safe
/*
	// The close built-in function closes a channel, which must be either
	// bidirectional or send-only. It should be executed only by the sender,
	// never the receiver, and has the effect of shutting down the channel after
	// the last sent value is received. After the last value has been received
	// from a closed channel c, any receive from c will succeed without
	// blocking, returning the zero value for the channel element. The form
	//	x, ok := <-c
	// will also set ok to false for a closed channel.
	func close(c chan<- Type)
*/

func main() {
	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu)
	data := make(chan int, 100)
	exit := make(chan bool)

	for i := 0; i < 20; i++ {
		data <- i
	}
	// close(data) //only take effect after last value received
	for i := 0; i < 5; i++ {
		i := i
		datab := data //chan is a reference structure, it could be copied.
		go func() {
			for d := range datab {
				fmt.Println(i, ": ", d)
			}
			exit <- true
		}()
	}

	close(data)
	// close(data) //panic
	<-exit
	<-exit
	<-exit
	<-exit
	<-exit
}
