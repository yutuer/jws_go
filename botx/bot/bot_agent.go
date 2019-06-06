package bot

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/rpc/jsonrpc"
	"net/textproto"
	"time"

	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/servers/gate"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (bot *PlayerBot) IsOnline() bool {
	return bot.agent != nil
}

var ErrAlreadyOnline = errors.New("All Ready Online")
var ErrAlreadyOffline = errors.New("All Ready Offline")

func (bot *PlayerBot) TryGoOnline() error {
	bot.agentLock.Lock()
	defer bot.agentLock.Unlock()

	if bot.IsOnline() {
		return ErrAlreadyOnline
	}

	if err := bot.login(bot.rpcAddr); err != nil {
		return err
	}

	if err := bot.connect(bot.gateSer); err != nil {
		return fmt.Errorf("session connect error: %s", err.Error())
	}

	return nil
}

func (bot *PlayerBot) GoOffline() error {
	bot.agentLock.Lock()
	defer bot.agentLock.Unlock()

	if !bot.IsOnline() {
		return ErrAlreadyOffline
	}

	close(bot.agentQuit)

	//bot.agentLock.Lock()
	//defer bot.agentLock.Unlock()

	//只有确保了当前agent退出后，才能保证服务器不会出现两个session读取了同一个存档
	//尽管其中给一个session数据还没有落地，另外一个session的数据就加载到内存了。
	<-bot.agent.GetGoneChan()
	//time.Sleep(time.Second * 30)

	bot.agent = nil
	bot.agentQuit = nil
	return nil
}

func (bot *PlayerBot) login(gateRPCAddr string) error {
	if raw_conn, err := net.DialTimeout("tcp", gateRPCAddr, time.Second*3); err != nil {
		//logs.Error("login json rpc dial error %s", err.Error())
		return fmt.Errorf("loginrpc dial err %s, token is %s", err.Error(), bot.loginToken)
	} else {
		//defer raw_conn.Close()
		var reply bool
		conn := jsonrpc.NewClient(raw_conn)

		defer conn.Close()

		rpcDone := conn.Go("GateServer.RegisterLoginToken",
			&gate.LoginNotify{bot.account, bot.loginToken, 0}, &reply, nil)
		select {
		case d := <-rpcDone.Done:
			if d.Error != nil {
				return fmt.Errorf("loginrpc conn error %s, token:%s", err.Error(), bot.loginToken)
			} else {
				logs.Trace("loginrpc with reply: %v, %v", reply, d.Reply)
			}
		case <-time.After(time.Second * 3):
			return fmt.Errorf("loginrpc  call timeout! token is %s", bot.loginToken)
		}
	}
	return nil
}

func (bot *PlayerBot) connect(serverAddr string) error {
	conn, err := net.DialTimeout("tcp", serverAddr, time.Second*3)
	if err != nil {
		//logs.Error("connect dial err: %s", err.Error())
		return fmt.Errorf("connect dial err, %s", err.Error())
	}
	logs.Trace("account:%s, Bot connect address la:%s, ra:%s", bot.account, conn.LocalAddr().String(), conn.RemoteAddr().String())

	//handshake
	cnc := &util.ConnNopCloser{conn}

	//nw := bufio.NewWriter(cnc)
	//tw := textproto.NewWriter(nw)

	fmt.Fprintf(conn, "%s\r\n", bot.loginToken)
	logs.Trace("account:%s, handshake token send out. %s", bot.account, bot.loginToken)
	//tw.PrintfLine(bot.loginToken)

	nr := bufio.NewReader(cnc)
	tr := textproto.NewReader(nr)

	for i := 0; i < 4; i++ {
		l, err := tr.ReadLine()
		if err != nil {
			//logs.Error("connect handshake step:%d, err:%s", i, err.Error())
			return fmt.Errorf("account:%s, connect handshake err, step %d, string:%s, err:%s", bot.account.String(), i, l, err.Error())
		}
		logs.Trace("account:%s, handshake %d:%s", bot.account, i+1, l)
	}
	logs.Trace("account:%s, make new agent", bot.account)
	agent := client.NewPacketConnAgent(bot.account.String(), conn)
	agent.WaitSendAll = true
	agent.IdleTimeSpan = time.Second * 10

	aquit := make(chan struct{})
	go agent.Start(aquit)
	go func() {
		for range agent.GetReadingChan() {
			//logs.Trace("Response")
		}
	}()

	//bot.agentLock.Lock()
	//defer bot.agentLock.Unlock()

	bot.agent = agent
	bot.agentQuit = aquit
	return nil
}
