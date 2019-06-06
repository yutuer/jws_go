package main

import (
	"flag"
	"net"
	// "os"
	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	// "strconv"
	"time"
)

const (
	RECV_BUF_LEN = 1024
	MSG_SEND_GAP = 1
)

func main() {
	flag.Parse()

	logs.Info("Starting the client")

	conn, err := net.Dial("tcp", "127.0.0.1:6666")
	defer conn.Close()

	if err != nil {
		logs.Error("error Dial:", err.Error())
	}

	go func() {
		i := time.Duration(1)
		for {
			select {
			case <-time.After(i * time.Second):
				// i += 1 //to test SetReadDeadline as heartbeat
				logs.Info("current interval %d", i)
				client.SendBytes(conn, []byte("PING"), client.PacketIDPingPong)
				client.SendBytes(conn, []byte(flag.Arg(0)), client.PacketIDContent)
			}
		}
	}()

	go func() {
		for {
			buf, err := client.ReadPacket(conn)
			if err != nil {
				logs.Error("Read Error %v", err.Error())
				return
			}
			if buf.IsContent() {
				logs.Info("REV %v", buf)
			}
		}
	}()
	<-time.After(300 * time.Second)
}
