package main

import (
	"bufio"
	"github.com/ugorji/go/codec"
	"log"
	"net"
	"net/rpc"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:42586")
	if err != nil {
		log.Fatal(err)
	}

	var mh codec.MsgpackHandle
	rpcCodec := codec.MsgpackSpecRpc.ClientCodec(conn, &mh)
	client := rpc.NewClientWithCodec(rpcCodec)

	in := bufio.NewReader(os.Stdin)
	for {
		line, _, err := in.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		var reply bool
		err = client.Call("Listener.GetLine", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	}
}
