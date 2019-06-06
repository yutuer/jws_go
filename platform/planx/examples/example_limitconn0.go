package main

import (
	"flag"
	"net"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 只Listen客户端尝试连接时，服务器是不会进行ESTABLISHED的。换句话说就是不消耗系统资源

func main() {
	listento := flag.String("listen", ":6666", "0.0.0.0:6666, :6666")
	flag.Parse()
	exitChan := make(chan struct{})

	listener, err := net.Listen("tcp", *listento)
	if err != nil {
		logs.Error(err.Error())
	}
	defer listener.Close()

	<-exitChan
}
