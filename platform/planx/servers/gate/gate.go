package gate

import (
	"net"

	//"code.google.com/p/go-uuid/uuid"

	"reflect"
	//"strings"

	"github.com/BurntSushi/toml"
	"github.com/ugorji/go/codec"

	"strings"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/gate/ccumetrics"
	"vcs.taiyouxi.net/platform/planx/servers/gate/rpc"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var mh codec.MsgpackHandle

func init() {
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = true
	mh.WriteExt = true
}

type GateServer struct {
	server servers.ConServer

	quit     chan struct{}
	gameSrvs GameServerManager

	cfg  Config
	info *loginInfo

	waitGroup util.WaitGroupWrapper
}

func NewGateServer(cfgname string) *GateServer {
	var gate GateServer
	config.NewConfig(cfgname, true, func(lcfgname string, cmd config.LoadCmd) {
		switch cmd {
		case config.Load: //, config.Reload:
			var c Config
			if _, err := toml.DecodeFile(lcfgname, &c); err != nil {
				logs.Critical("App config load failed. %s, %s\n", lcfgname, err.Error())
			} else {
				logs.Info("Config loaded: %s\n", lcfgname)
				//if cmd == config.Reload {
				////XXX 不能更新所有配置，只更新自己需要的配置部分
				//gate.cfg.MetricsConfig = c.MetricsConfig
				//metrics.Reload(c.MetricsConfig)
				//} else {
				gate.cfg = c
				//}
			}

		}

	})

	cfg := gate.cfg.GateConfig

	logs.Debug("Gate cfg %v", cfg)

	scfg := servers.NewConnServerCfg{
		ListenTo:             cfg.Listen,
		NumberOfAcceptor:     cfg.NAcceptor,
		NumberOfWaitingQueue: cfg.NWaitingConn,
		SslCfg:               cfg.SslCfg,
		SslCaCfg:             cfg.SslCaCfg,
	}
	switch cfg.ConnServer {
	case "LimitConnServer":
		logs.Info("Start with LimitConnServer server")
		slcfg := servers.NewLimitConnServerCfg{
			NewConnServerCfg: scfg,
			MaxConn:          cfg.MaxConn,
		}
		gate.server = servers.NewLimitConnServer(slcfg)
	default:
		logs.Info("Start with default ConnServer server")
		gate.server = servers.NewConnServer(scfg)
	}

	gate.quit = make(chan struct{})
	gate.info = newLoginInfo()

	return &gate
}

func (g *GateServer) forwardToClients(quit chan struct{}, AccountID string, gs GameServer, agent *client.PacketConnAgent) {
	defer logs.PanicCatcherWithAccountID(AccountID)
	logs.Info("[GateServer] forwardToClients start, account id: %s", AccountID)
	pktChan := gs.GetReadingChan()
Loop:
	for {
		select {
		case <-quit:
			break Loop
		case <-g.quit:
			break Loop
		case <-gs.GetGoneChan():
			break Loop
		case pkt, ok := <-pktChan:
			if !ok || pkt == nil {
				break Loop
			}
			switch pkt.GetContentType() {
			case client.PacketIDPingPong:
				if string(pkt.GetBytes()) == "PING" {
					logs.Warn("[GateServer] forwardToClients get PING PONG!")
				}
			case client.PacketIDGatePkt:
				var (
					sessionpkt client.SessionPacket
				)
				dec := codec.NewDecoderBytes(pkt.GetBytes(), &mh)
				dec.Decode(&sessionpkt)

				//LOG MARKER
				client.Log("rsp", AccountID, sessionpkt.PacketData)

				if ok := agent.SendPacket(sessionpkt.PacketData); !ok {
					break Loop
				}
				//if err != nil {
				//if e, ok := err.(net.Error); ok && e.Timeout() {
				//logs.Warn("Gate.servive  reply to client failed1! %s", e.Error())
				//continue
				//}
				//logs.Warn("Gate.servive  reply to client failed2! %s", err.Error())
				////XXX 这里考虑回收GameServer然后等待客户端重连
				//break Loop
				//}
			}

		}
	}
	logs.Trace("[GateServer] forwardToClients exit, account id: %s", AccountID)

}

func (g *GateServer) forwardToGameServer(SessionID string, AccountID string, agent *client.PacketConnAgent, gs GameServer) {
	defer logs.PanicCatcherWithAccountID(AccountID)
	logs.Trace("[GateServer] forwardToGameServer start, for account %s", AccountID)
	//g.registerConn(SessionID, agent)
	agentReadingChan := agent.GetReadingChan()

Loop:
	for {
		select {
		case <-g.quit:
			break Loop
		case <-gs.GetGoneChan():
			break Loop
		case pkt, ok := <-agentReadingChan:
			if !ok {
				break Loop
			}

			//LOG MARKER
			client.Log("req", AccountID, pkt)

			switch pkt.GetContentType() {
			case client.PacketIDPingPong:
				//logs.Warn("[GateServer] forwardToServices should not get pingpong")
				//网关服务器会收到客户端的PING信息，目前不做任何处理
			default:
				//XXX: Will it be a performance factor? there is Lock inside it
				ccumetrics.CountAddNewRequest(1)

				var out []byte //XXX: If it comes from []byte buffer pool, it would be cool
				enc := codec.NewEncoderBytes(&out, &mh)
				enc.Encode(client.SessionPacket{SessionID, AccountID, pkt})
				SessPkt := client.NewPacket(out, client.PacketIDGatePkt)

				// XXX should not recreate always
				if ok := gs.SendPacket(SessPkt); !ok {
					break Loop
				}
				//logs.Trace("[GateServer] dispatch a message to game server done.")
			}
		}
	}
	//g.unRegisterConn(SessionID)
	logs.Trace("[GateServer] forwardToGameServer end, for account %s", AccountID)
}

func (g *GateServer) chooseGameServer(accountID string) string {
	// TODO 完善GameServer的选择，需要在实现多个Game服务同一个Gate的时候才需要考虑实现
	if len(g.cfg.GateConfig.GameServers) == 0 {
		return "allinone"
	}
	return g.cfg.GateConfig.GameServers[0]
}

//func (g *GateServer) GetLoginStatusChan() <-chan LoginStatus {
//return g.loginStatusChan
//}

//建立玩家和后端服务器的链接，后段服务器可以有多个
//玩家会被分配到固定的服务器上，所有后端服务器功能相同
//暂时不会按照功能拆分
//方便高效断线重连接模式的实现
// zz 不支持短时间内多次连接，目前会有问题；实现方案也是现实10s内不能重复登录
func (g *GateServer) handleConnection(con net.Conn) {
	var AccountID, IpAddr string
	//logs.Debug("Gate Server, handleConnection with %s", con.RemoteAddr().String())
	IpAddr, _, _ = net.SplitHostPort(con.RemoteAddr().String())
	//logs.Debug("Gate Server, handleConnection start run with ip %s", IpAddr)
	defer logs.PanicCatcherWithAll(AccountID, IpAddr)

	ccumetrics.AddCCU()
	defer func() {
		ccumetrics.ReduceCCU()
		g.server.ReleaseConnChan(con)
		con.Close()
	}()

	logs.Trace("<Gate> gate.handleConnection la:%s ra:%s", con.LocalAddr().String(), con.RemoteAddr().String())
	agent := client.NewPacketConnAgent(AccountID, con)
	defer agent.Close()
	if ok, ln, gzipLimit, sessionId := g.handShake(con, agent); ok {
		AccountID = ln.String() //makeAccountID(ln.GameID, ln.ShardID, ln.UserID)

		///////////////////////////////////////
		//通知login服务器我上线
		client.LogSession(AccountID, "online")
		ccumetrics.NotifyLogInOff(ccumetrics.LoginStatus{
			LoginToken: ln.LoginToken,
			AccountID:  AccountID,
			LogInOff:   true,
		})

		//清理掉线玩家的信息
		defer func() {
			g.info.activeSessionCache.Delete(fmt.Sprintf("%d", sessionId))
			g.info.mu.Lock()
			if info, ok := g.info.sessionMap[AccountID]; ok {
				if info.id == agent.SessionId {
					delete(g.info.sessionMap, AccountID)
				}
			}
			sessionSize := len(g.info.sessionMap)
			g.info.mu.Unlock()
			logs.Info("update gate count metrics, %d", sessionSize)
			modules.UpdateGaugeByGate(int64(sessionSize))
		}()

		//通知Login服务器我下线
		defer func() {
			client.LogSession(AccountID, "offline")
			ccumetrics.NotifyLogInOff(ccumetrics.LoginStatus{
				LoginToken: ln.LoginToken,
				AccountID:  AccountID,
				LogInOff:   false,
			})
		}()

		///////////////////////////////////////

		addr := g.chooseGameServer(AccountID)
		gs := g.gameSrvs.NewGameServer(AccountID, addr)
		defer g.gameSrvs.RecycleGameServer(gs)
		///////////////////////////////////////
		//go func() { //just for debug
		//<-time.After(5 * time.Second)
		//logs.Warn("Send out a kick")
		//sendKickNotify(agent)
		//}()

		///////////////////////////////////////
		//发送account id到后端建立服务关系
		var out []byte //XXX: If it comes from []byte buffer pool, it would be cool
		enc := codec.NewEncoderBytes(&out, &mh)
		enc.Encode(AccountID)
		enc.Encode(strings.Split(con.RemoteAddr().String(), ":")[0])
		enc.Encode(fmt.Sprintf("%d", gzipLimit))
		handShakePkt := client.NewPacket(out, client.PacketIDGateSession)
		gs.SendPacket(handShakePkt)

		var wg util.WaitGroupWrapper
		backendQuit := make(chan struct{})

		//agent, gs之间的退出是交叉退出，任何人出现问题，则大家都退出！
		wg.Wrap(func() {
			agent.Start(g.quit)
		})
		wg.Wrap(func() {
			g.forwardToGameServer(ln.LoginToken, AccountID, agent, gs)
			defer close(backendQuit) //TODO 实现高级断线重连就从这里切入，这里是临时实现
			agent.Stop()
		})
		wg.Wrap(func() {
			g.forwardToClients(backendQuit, AccountID, gs, agent)
			agent.Stop()
			// TODO gate game 分布式模式.如果gs比agent提前出现问题，这里就退出了，然后其他人都卡死.
			// gs重启，也不会有机会恢复这里
			// allinone 模式下，gs同样可能出问题，但是已经处理过相关退出了 201510.29
		})
		wg.Wait()
		logs.Debug("<Gate> handleConnection close %s", AccountID)
	}
	//结束链接
}

func (g *GateServer) loop() {
	cc := g.server.GetWaitingConnChan()
	for conn := range cc {
		myconn := conn
		g.waitGroup.Wrap(func() { g.handleConnection(myconn) })
	}
	//for {
	//select {
	//case <-g.quit:
	//logs.Info("Gate.loop, server quit")
	//return
	//case conn, ok := <-g.server.GetWaitingConnChan():
	//if ok {
	//g.waitGroup.Wrap(func() { g.handleConnection(conn) })
	//}
	//}
	//}
}

var _rpc *rpc.GateRPC

func (g *GateServer) Start(gsm GameServerManager) {

	cfg := g.cfg

	pip := cfg.GateConfig.PublicIP
	cfg.GateConfig.PublicIP = GetPublicIP(pip, cfg.GateConfig.Listen)

	logs.Info("[GateServer] with public ip %s binded.", cfg.GateConfig.PublicIP)
	//TODO: 这个action这样进行分类中转，显然不理想，需要处理。需要分离gate, game的时候再考虑
	// 临时提取到这个位置，需要考虑etcd进行动态的注册和调整？
	// 需要能够动态更新这些配置，或者通过etcd的配置进行更新，或者HUP信号的更新
	g.gameSrvs = gsm
	go g.server.Start()

	//rpc server of gate
	_rpc = rpc.NewGateRPCServer(cfg.RPCConfig)
	//signalhandler.SignalKillHandler(gatexRPC)
	var wg_service util.WaitGroupWrapper
	wg_service.Wrap(func() { _rpc.Start(g) })
	//report ccu to login server
	shards := make([]uint, 0, len(cfg.GateConfig.ShardID)+len(cfg.GateConfig.MergeRel))
	shards = append(shards, cfg.GateConfig.ShardID...)
	shards = append(shards, cfg.GateConfig.MergeRel...)

	rpcListen := cfg.RPCConfig.RPCListen
	if "" != cfg.RPCConfig.RPCListenDocker {
		rpcListen = cfg.RPCConfig.RPCListenDocker
	}

	wg_service.Wrap(func() {
		ccumetrics.Start(
			cfg.CCUMetricsConfig,
			shards,
			cfg.GateConfig.PublicIP,
			rpcListen,
			cfg.IsRunModeLocal(),
		)
	})
	go g.loop()
	//g.waitGroup.Wrap(func() { g.loop() })
	g.waitGroup.Wrap(func() { g.gameSrvs.WaitAllShutdown(g.quit) })
	g.waitGroup.Wait()
	//close(g.loginStatusChan)

	g.server.Stop()
	_rpc.Stop()
	ccumetrics.Stop()

	wg_service.Wait()
}

func (g *GateServer) Stop() {
	close(g.quit)
}

func (g *GateServer) SyncCfg2Etcd() bool {
	gid, err := etcd.FindServerGidBySid(g.cfg.CCUMetricsConfig.EtcdRoot+"/", g.cfg.GateConfig.ShardID[0])
	if err != nil || gid < 0 {
		logs.Error("etcd GetServerGidSid err %s or not found", err.Error())
		return false
	} else {
		_ss := make([]uint, 0, 4)
		_ss = append(_ss, g.cfg.GateConfig.ShardID...)
		_ss = append(_ss, g.cfg.GateConfig.MergeRel...)
		for _, sid := range _ss {
			key_parent := fmt.Sprintf("%s/%d/%d/gm/", g.cfg.CCUMetricsConfig.EtcdRoot, gid, sid)
			// private_ip
			ss := strings.Split(g.cfg.RPCConfig.RPCListen, ":")
			private_ip := ss[0]
			if err := etcd.Set(key_parent+"private_ip", private_ip, 0); err != nil {
				logs.Error("set etcd key %s err %s", key_parent+"private_ip", err)
				return false
			}
		}
	}
	return true
}

func (g *GateServer) IsReg2Etcd() bool {
	return g.cfg.GateConfig.RunMode == RunModeTest || g.cfg.GateConfig.RunMode == RunModeProd
}
