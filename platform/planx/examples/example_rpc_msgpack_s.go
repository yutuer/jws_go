package main

import (
	"fmt"
	"github.com/ugorji/go/codec"
	"log"
	"net"
	"net/rpc"
)

// rpc.ServeCodec 类似rpc.ServeConn是开启一个tcp链接，进行有编码的通信
// json 编码也是这样处理的, 可以参考http://crosbymichael.com/golang-json-rpc.html
// using server.ServeCodec will close `conn`, `ServeRequest` will not

type Listener int

func (l *Listener) GetLine(line []byte, ack *bool) error {
	fmt.Println(string(line))
	return nil
}

func main() {
	listener := new(Listener)
	rpc.Register(listener)
	// rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	addy, err := net.ResolveTCPAddr("tcp", "0.0.0.0:42586")
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.ListenTCP("tcp", addy)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		var mh codec.MsgpackHandle
		rpcCodec := codec.MsgpackSpecRpc.ServerCodec(conn, &mh)
		go rpc.ServeCodec(rpcCodec)
	}

	// rpc.Accept(inbound)
}
