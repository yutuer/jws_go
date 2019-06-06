package main

import (
	"flag"
	_ "fmt"
	"net"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 接收链接到达max后，再进入链接会被accept然后Close，kick out。有效链接数在max
// 当然可以不Close，把超出max限定的连接数做成一个排队pool，这些Connection可以进行通信，告知排队状况
// 如果这样做需要额外加入一个参数限定排队人数
// PS: 受到这个启发，Edge服务器相对容易进行扩展进行服务，因此，如果能够使得排队人数变得任意多。
// 推荐使用这个模式

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
	logs.Info("%d", *maxconn)
	// The booleans representing the free active connection spaces.
	spaces := make(chan bool, *maxconn)
	// Initialize the spaces
	for i := uint(0); i < *maxconn; i++ {
		spaces <- true
	}

	listener, err := net.Listen("tcp", *listento)
	if err != nil {
		logs.Error(err.Error())
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logs.Error(err.Error())
		}

		go func(cp net.Conn) {
			for {
				select {
				case <-spaces:
					go handleConnection(cp, spaces)
					return
				case <-time.After(time.Millisecond * 1000):
					logs.Info("Out of service %s", cp.RemoteAddr())
					cp.Close()
					return
				}
			}
		}(conn)

	}
}
