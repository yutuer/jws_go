package main

import (
	"crypto/md5"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codegangsta/cli"

	"vcs.taiyouxi.net/jws/crossservice/metrics"
	"vcs.taiyouxi.net/jws/crossservice/util/connect"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "server",
			Usage:  "server mode",
			Action: serverAction,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "p",
					Usage: "port of server listening",
					Value: 32333,
				},
				cli.BoolFlag{
					Name:  "v",
					Usage: "show verbose",
				},
				cli.StringSliceFlag{
					Name:  "filter",
					Usage: "filter array like \"192.168.2.0/24\" ",
				},
			},
		},
		cli.Command{
			Name:   "client",
			Usage:  "client mode",
			Action: clientAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "ip of server",
					Value: "127.0.0.1",
				},
				cli.IntFlag{
					Name:  "p",
					Usage: "port of server",
					Value: 32333,
				},
				cli.BoolFlag{
					Name:  "v",
					Usage: "show verbose",
				},
				cli.IntFlag{
					Name:  "n",
					Usage: "num of connection pool capcity",
					Value: 3,
				},
				cli.IntFlag{
					Name:  "t",
					Usage: "interval of send per goroutine",
					Value: 0,
				},
			},
		},
	}

	app.Run(os.Args)
}

type message struct {
	checksum [md5.Size]byte
	data     []byte
}

func makeMessageRand(rander *rand.Rand) *message {
	m := &message{}
	// l := rander.Uint32() % 0xFFFFF
	l := 0xFFF0*3 - md5.Size
	buf := make([]byte, l)
	_, err := rander.Read(buf)
	if nil != err {
		log.Printf("makeMessageRand Failed, rand read failed, %v", err)
		return nil
	}
	m.data = buf
	m.checksum = md5.Sum(m.data)
	return m
}

func (m *message) toMsg() *connect.Message {
	msg := &connect.Message{}
	msg.Length = len(m.data) + md5.Size
	msg.Payload = make([]byte, msg.Length)
	copy(msg.Payload[:md5.Size], m.checksum[:md5.Size])
	copy(msg.Payload[md5.Size:], m.data)
	return msg
}

func checkMessage(msg *connect.Message) bool {
	recvChecksum := msg.Payload[:md5.Size]
	dataChecksum := md5.Sum(msg.Payload[md5.Size:])
	return string(recvChecksum) == string(dataChecksum[:])
}

func serverAction(c *cli.Context) {
	port := c.Int("p")
	verbose := c.Bool("v")
	ipArr := c.StringSlice("filter")

	cs_metrics.Reg()

	server, err := connect.NewServer("tcp4", "", port)
	if nil != err {
		log.Printf("NewServer failed, %v", err)
		return
	}
	log.Printf("Listen at %v", server.LocalAddr().String())

	if 0 != len(ipArr) {
		filter := connect.NewIPFilter()
		for _, str := range ipArr {
			filter.Add(str)
		}
		server.SetIPFilter(filter)
	}

	go func() {
		err := server.Run()
		if nil != err {
			log.Printf("server end with error: %v", err)
			if connect.ErrClosed != err {
				server.Close()
			}
		}
	}()

	go func() {
		for {
			msg, conn, err := server.Recv()
			if nil != err {
				log.Printf("server receive error: %v", err)
				if connect.ErrClosed == err {
					break
				}
				continue
			}
			if verbose {
				log.Printf("Recv Msg Checksum [%x], length [%d]", msg.Payload[:md5.Size], msg.Length)
				log.Printf("Recv Msg From [%s], length [%d]", conn.RemoteAddr().String(), msg.Length)
			}
			if !checkMessage(msg) {
				log.Printf("!!! Recv Msg From [%s], checksum fail", conn.RemoteAddr().String())
			}
		}
	}()

	blockWithSignal()
	server.Close()
}

func clientAction(c *cli.Context) {
	ip := c.String("i")
	port := c.Int("p")
	verbose := c.Bool("v")
	max := c.Int("n")
	interval := c.Int("t")

	client, err := connect.NewClient("tcp4", ip, port, uint32(max))
	if nil != err {
		log.Printf("NewClient failed, %v", err)
		return
	}

	for i := 0; i < max; i++ {
		go func() {
			var rander = rand.New(rand.NewSource(time.Now().UnixNano()))
			for {
				if 0 != interval {
					<-time.After(time.Millisecond * time.Duration(interval))
				}
				m := makeMessageRand(rander)
				msg := m.toMsg()
				conn, err := client.GetConn()
				if nil != err {
					log.Printf("client GetConn failed, %v", err)
					break
				}

				if verbose {
					log.Printf("Send Msg Checksum [%x], length [%d]", m.checksum, msg.Length)
					log.Printf("Send Msg From [%s], %v", conn.LocalAddr().String(), msg.Payload)
				}

				err = conn.Send(msg)
				if nil != err {
					log.Printf("client send error: %v", err)
				}
				conn.Release()
			}
		}()
	}

	blockWithSignal()
	client.Close()
}

func blockWithSignal() {
	nc := make(chan os.Signal, 10)
	signal.Notify(nc, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	s := <-nc
	log.Printf("")
	log.Printf("got signal %v", s)
}
