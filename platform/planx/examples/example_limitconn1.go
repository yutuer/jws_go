package main

import (
	"flag"
	_ "fmt"
	"net"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 接收链接到达max后，不会再有新链接建立
// Accept是并发的，所以响应速度可能还不错
// 可以扩展成limitconn2的模式，加一个waiting用户
// 简单粗暴

func handleConnection(connection net.Conn, exit chan bool) {
	logs.Info("handleConnection %s", connection.RemoteAddr())
	defer connection.Close()

	var bytes []byte = make([]byte, 256*1024)
	for {
		connection.SetReadDeadline(time.Now().Add(time.Second * 10))
		n, err := connection.Read(bytes)

		if err != nil {
			e, ok := err.(net.Error)
			if ok && (e.Temporary() || e.Timeout()) {
				continue
			} else {
				logs.Error("Read TCP Error, %s, %v", connection.RemoteAddr(), err)
				break
			}
		}
		logs.Info("%s", string(bytes[:n]))
	}
	logs.Info("handleConnection %s exit", connection.RemoteAddr())
	exit <- true
}

func main() {
	listento := flag.String("listen", ":6666", "0.0.0.0:6666, :6666")
	maxconn := flag.Uint("maxc", 2, "max connections to handle")
	flag.Parse()
	exitChan := make(chan struct{})

	listener, err := net.Listen("tcp", *listento)
	if err != nil {
		logs.Error(err.Error())
	}
	defer listener.Close()

	for i := uint(0); i < *maxconn; i++ {
		i := i
		go func() {
			again := make(chan bool)
			defer close(again)
			for {
				logs.Info("Acceptor %d ready!", i)
				conn, err := listener.Accept()
				if err != nil {
					logs.Error("Error accept: %s", err.Error())
					return
				}
				logs.Info("Got %s", conn.RemoteAddr())
				go handleConnection(conn, again)
				<-again
			}
		}()
	}
	<-exitChan
}
