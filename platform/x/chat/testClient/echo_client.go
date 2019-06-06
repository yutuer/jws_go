package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"vcs.taiyouxi.net/platform/planx/funny/link"
)

func main() {
	var wg sync.WaitGroup

	runtime.GOMAXPROCS(32)

	wg.Add(1)
	for i := 0; i < 10; i++ {
		CreateAnAgent(fmt.Sprintf("Agent%d", i))
	}

	wg.Wait()
	println("bye")
}

func CreateAnAgent(name string) {
	addr := "127.0.0.1:10010"

	session, err := link.Connect("tcp", addr, link.String())
	if err != nil {
		panic(err)
	}

	go func(n string) {
		var msg string
		for {
			if err := session.Receive(&msg); err != nil {
				println("byebye " + n)
				return
			}
			fmt.Printf("%s : %s\n", n, msg)
		}
	}(name)

	go func(n string) {
		c := 1
		Reg(session, n, "1")
		for range time.Tick(time.Second * 5) {
			now := time.Now().Format("2006-01-02 15:04:05")
			err := Say(session, fmt.Sprintf("Say %d %s", c, now), "1")
			if err != nil {
				println("bye " + n)
				return
			}
			c++
		}
	}(name)
}

func SendMsg(session *link.Session, format string, params ...interface{}) error {
	msg := fmt.Sprintf(format, params...)
	if err := session.Send(msg); err != nil {
		return err
	}
	return nil
}

func Reg(session *link.Session, name, roomkey string) error {
	return SendMsg(session, "Reg:%s:%s\n", roomkey, name)
}

func UnReg(session *link.Session, info, roomkey string) error {
	return SendMsg(session, "UnReg:%s:%s\n", roomkey, info)
}

func Say(session *link.Session, info, roomkey string) error {
	return SendMsg(session, "Say:%s:%s\n", roomkey, info)
}
