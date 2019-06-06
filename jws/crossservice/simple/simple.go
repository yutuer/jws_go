package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fmt"

	"bytes"
	"encoding/gob"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/jws/crossservice/client"
	"vcs.taiyouxi.net/jws/crossservice/module/simple"
	"vcs.taiyouxi.net/jws/crossservice/module/simpledynamic"
	"vcs.taiyouxi.net/jws/crossservice/server"

	_ "vcs.taiyouxi.net/jws/crossservice/module/simple"        //simple module
	_ "vcs.taiyouxi.net/jws/crossservice/module/simpledynamic" //simpledynamic module
)

func main() {
	app := cli.NewApp()

	defaultGroupIDs := cli.IntSlice([]int{})
	defaultShardIDs := cli.IntSlice([]int{})
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "server",
			Usage:  "server mode of crossservice",
			Action: serverAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "ip of server to listen",
					Value: "127.0.0.1",
				},
				cli.IntFlag{
					Name:  "p",
					Usage: "port of server to listen",
					Value: 9527,
				},
				cli.BoolFlag{
					Name:  "v",
					Usage: "verbose",
				},
				cli.IntSliceFlag{
					Name:  "g",
					Usage: "group id list",
					Value: &defaultGroupIDs,
				},
			},
		},
		cli.Command{
			Name:   "client",
			Usage:  "client mode of crossservice",
			Action: clientAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "ip of server to access",
					Value: "127.0.0.1",
				},
				cli.IntFlag{
					Name:  "p",
					Usage: "port of server to access",
					Value: 9527,
				},
				cli.BoolFlag{
					Name:  "v",
					Usage: "verbose",
				},
				cli.IntSliceFlag{
					Name:  "g",
					Usage: "group id list",
					Value: &defaultGroupIDs,
				},
				cli.IntFlag{
					Name:  "t",
					Usage: "interval of send (ms)",
					Value: 1000,
				},
				cli.IntSliceFlag{
					Name:  "s",
					Usage: "shard id list",
					Value: &defaultShardIDs,
				},
				cli.StringFlag{
					Name:  "c",
					Usage: "call method",
					Value: "sync",
				},
			},
		},
	}

	app.Run(os.Args)
}

func clientAction(c *cli.Context) {
	shardList := []uint32{}
	if 0 == len(c.IntSlice("s")) {
		log.Printf("Need ShardID List")
		return
	}
	for _, s := range c.IntSlice("s") {
		shardList = append(shardList, uint32(s))
	}
	client := client.NewClient(0, shardList)
	groupList := []uint32{}
	if 0 == len(c.IntSlice("g")) {
		log.Printf("Need GroupID List")
		return
	}
	for _, g := range c.IntSlice("g") {
		groupList = append(groupList, uint32(g))
	}
	client.AddGroupIDs(groupList)
	err := client.Start()
	if nil != err {
		log.Printf("Client start failed, %v", err)
		return
	}

	tickerCall := time.NewTicker(time.Duration(c.Int("t")) * time.Millisecond)

	clm := c.String("c")

	type TransferParam struct {
		Hello string
		World string
	}

	rand.Seed(time.Now().Unix())
	go func() {
		for {
			<-tickerCall.C
			group := groupList[rand.Int()%len(groupList)]
			switch clm {
			case "sync":
				param := &simple.ParamSimpleSync{In: uint32(rand.Int() % 256)}
				source := fmt.Sprintf("%d", param.In)
				ret, _, err := client.CallSync(group, simple.ModuleID, simple.MethodSimpleSyncID, source, param)
				if nil != err {
					log.Printf("Error, CallSync, %v", err)
					continue
				}
				log.Printf("CallSync, param %+v, ret %+v", param, ret.(*simple.RetSimpleSync))
			case "async":
				rd := rand.Int() % 1024
				param := &simple.ParamSimpleAsync{Hello: fmt.Sprintf("hello:%d", rd)}
				source := fmt.Sprintf("%d", rd)
				_, err := client.CallAsync(group, simple.ModuleID, simple.MethodSimpleAsyncID, source, param)
				if nil != err {
					log.Printf("Error, CallAsync, %v", err)
					continue
				}
				log.Printf("CallAsync, param %+v", param)
			case "transfer":
				p := &TransferParam{
					Hello: fmt.Sprintf("%d", rand.Int()%4096),
					World: fmt.Sprintf("%d", rand.Int()%4096),
				}
				buf := new(bytes.Buffer)
				if err := gob.NewEncoder(buf).Encode(p); nil != err {
					log.Printf("Error, Pull Encode, %v", err)
					continue
				}
				param := &simple.ParamSimpleTransfer{ShardID: shardList[0], Payload: buf.Bytes()}
				source := "1"
				_, err := client.CallAsync(group, simple.ModuleID, simple.MethodSimpleTransferID, source, param)
				if nil != err {
					log.Printf("Error, Transfer CallAsync, %v", err)
					continue
				}
				log.Printf("Transfer CallAsync, param %+v", param)
			case "status":
				param := &simpledynamic.ParamGetStatus{}
				source := "status"
				ret, _, err := client.CallSync(group, simpledynamic.ModuleID, simpledynamic.MethodGetStatusID, source, param)
				if nil != err {
					log.Printf("Error, CallSync, %v", err)
					continue
				}
				log.Printf("Status CallSync, param %+v, ret %+v", param, ret.(*simpledynamic.RetGetStatus))
			default:
				log.Printf("NoCall [%s]", clm)
			}
		}
	}()
	go func() {
		for {
			data, _, err := client.Pull()
			if nil != err {
				log.Printf("Error, Pull, %v", err)
				break
			}
			p := &TransferParam{}
			if err := gob.NewDecoder(bytes.NewBuffer(data.Data)).Decode(p); nil != err {
				log.Printf("Error, Pull Decode, %v", err)
				continue
			}
			log.Printf("Pull, param %+v", p)
		}
	}()

	blockWithSignal()
	client.Stop()
	log.Printf("Client Exist")
}

func serverAction(c *cli.Context) {
	server := server.NewServer(0, c.String("i"), c.Int("p"))
	for _, g := range c.IntSlice("g") {
		server.AddGroupIDs([]uint32{uint32(g)})
	}
	err := server.Start()
	if nil != err {
		log.Printf("Server start failed, %v", err)
		return
	}
	log.Printf("Server start listen: %v", server.LocalAddr().String())

	blockWithSignal()
	server.Stop()
	log.Printf("Server Exist")
}

func blockWithSignal() {
	nc := make(chan os.Signal, 10)
	signal.Notify(nc, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	s := <-nc
	log.Printf("")
	log.Printf("got signal %v", s)
}
