package main

import (
	"flag"
	_ "fmt"
	"net"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//限定Connection数量会比设定数量大一倍，极限情况下，
//当前服务链接数是: max
//当前Waiting队列中滞留链接数: max, 这部分用户可以进行tcp通信
//阻塞中链接: 1
//所以实际链接数是2*max + 1
//这个模式对于**长链接服务**不是很理想，但是对于短链接服务确还不错，能够起到一定的缓冲效果

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
	// The channel of connections which are waiting to be processed.
	waiting := make(chan net.Conn, *maxconn)
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

	// Start the connection matcher.
	go matchConnections(waiting, spaces)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logs.Error(err.Error())
		}
		logs.Info("Received connection from %s", conn.RemoteAddr())
		waiting <- conn
	}
}

func matchConnections(waiting chan net.Conn, spaces chan bool) {
	// Iterate over each connection in the waiting channel
	for cp := range waiting {
		logs.Info("match connetion %s", cp.RemoteAddr())
		<-spaces
		go handleConnection(cp, spaces)
	}
}
