package ccumetrics

import (
	"fmt"
	"runtime"
	"time"

	"github.com/astaxie/beego/httplib"

	//"vcs.taiyouxi.net/platform/planx/servers/gate"

	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type loginConnector struct {
	ticker *time.Ticker
	cfg    Config
	//gatei  GateInfo

	loginStatusChan chan LoginStatus

	stopSign chan struct{}
}

func newLoginConnector(c Config) *loginConnector {
	logs.Trace("LoginConnectorTickTime %v", c)
	tick_time := time.Duration(c.LoginConnectorTickTime)

	l := &loginConnector{
		ticker:          time.NewTicker(tick_time * time.Second),
		cfg:             c,
		loginStatusChan: make(chan LoginStatus, 1024),
		stopSign:        make(chan struct{}),
	}
	return l
}

func httppost(url, gameip, rpcip string, ccu int64, sid uint) (string, error) {
	req := httplib.Post(url)
	req.Param("gameipaddrport", gameip)
	req.Param("rpcipaddrport", rpcip)
	req.Param("ccu", fmt.Sprintf("%d", ccu))
	req.Param("shardid", fmt.Sprintf("%d", sid))

	var str string
	err := req.ToJSON(&str)
	if err != nil {
		return "", fmt.Errorf("login server update error %s", err.Error())
	}

	rsp, _ := req.Response()
	defer rsp.Body.Close()

	if str != "ok" {
		return "", fmt.Errorf("login server update error %s", str)
	}
	//logs.Info("login server update success")
	return str, nil
}

func postLoginStatus(url, loginToken, accountId, RPCAddrPort string) {
	logs.Debug("postLoginStatus: %s", url)
	req := httplib.Post(url)
	req.Param("logintoken", loginToken)
	req.Param("accountid", accountId)
	req.Param("rpcaddrport", RPCAddrPort)

	logs.Debug("<Gate> postLoginStatus %s %s %s %s",
		url, loginToken, accountId, RPCAddrPort)

	var str string
	err := req.ToJSON(&str)
	if err != nil {
		logs.Error("<Gate> postLoginStatus failed err with %v", err)
		return
	}
	if str != "ok" {
		logs.Error("<Gate> postLoginStatus failed with %v", str)
		return
	}

}

func (l *loginConnector) notifyLogInOff() {
	c := l.loginStatusChan
	rpc := RpcIP
	var counter uint64
	for i := range c {
		url := l.cfg.LoginUrl
		var posturl string
		if i.LogInOff {
			addHandShakeCCU()
			posturl = l.cfg.URINotifyLogin
		} else {
			reduceHandShakeCCU()
			posturl = l.cfg.URINotifyLogout
		}

		if posturl == "" {
			runtime.Gosched()
			continue
		}

		url = fmt.Sprintf("%s/%s", url, posturl)

		go postLoginStatus(url,
			i.LoginToken,
			i.AccountID,
			rpc,
		)
		counter++
		if counter%10 == 0 {
			runtime.Gosched()
		}
	}
}

func (l *loginConnector) start() {
	gid := -1
	// find my gid from etcd
	g, err := etcd.FindServerGidBySid(l.cfg.EtcdRoot+"/", ShardID[0])
	if err != nil || g < 0 {
		logs.Error("etcd GetServerGidSid err %s or not found", err.Error())
	} else {
		gid = g
		logs.Debug("etcd get gid %d", gid)
	}

	url := fmt.Sprintf("%s/%s", l.cfg.LoginUrl, l.cfg.URICcu)
	logs.Info("Gate.LoginConnector start with %s", url)

	go l.notifyLogInOff()

	publicip := PublicIP
	internalip := RpcIP
	ccu := _ccu.Count()

	if isLocal {
		for _, sid := range ShardID {
			//logs.Trace("Gate.LoginConnector upload ccu %d, to shard %d", ccu, sid)
			_, err := httppost(url, publicip, internalip, ccu, sid)
			if err != nil {
				logs.Warn("Gate can't update its information to auth server, %s", err.Error())
			}
		}
	}

Loop:
	for {
		select {
		case <-l.ticker.C:
			publicip = PublicIP
			internalip = RpcIP
			ccu = _ccu.Count()

			for _, sid := range ShardID {
				if gid >= 0 {
					// to etcd
					//logs.Trace("Gate.LoginConnector ccu %d, upload shard %d to etcd", ccu, sid)
					keyIp := fmt.Sprintf("%s/%d/%d/%s", l.cfg.EtcdRoot, gid, sid, etcd.KeyIp)
					if err := etcd.Set(keyIp, publicip, time.Duration(l.cfg.LoginConnectorTickTime+1)*time.Second); err != nil {
						logs.Warn("register ip to etcd failed")
					} /*else {
						logs.Debug("send to etcd %s ip %s", keyIp, publicip)
					}*/
					keyInternalIp := fmt.Sprintf("%s/%d/%d/%s", l.cfg.EtcdRoot, gid, sid, etcd.KeyInternalIp)
					if err := etcd.Set(keyInternalIp, internalip, time.Duration(l.cfg.LoginConnectorTickTime+1)*time.Second); err != nil {
						logs.Warn("register internalip to etcd failed")
					} /*else {
						logs.Debug("send to etcd %s ip %s", keyInternalIp, internalip)
					}*/
				} else {
					logs.Error("not found gid from etcd, register failed")
				}

				if isLocal {
					// to auth 早期用法，直接给auth发gamex的ip
					//logs.Trace("Gate.LoginConnector upload ccu %d, to shard %d", ccu, sid)
					if str, err := httppost(url, publicip, internalip, ccu, sid); err != nil || str != "ok" {
						logs.Warn("register to auth failed, %s", err.Error())
					}
				}
			}
		case <-l.stopSign:
			break Loop
		}
	}
	l.ticker.Stop()

	logs.Info("Gate.LoginConnector quit")
}

func (l *loginConnector) stop() {
	close(l.stopSign)
	close(l.loginStatusChan)
}

//func externalIP() (string, error) {
//ifaces, err := net.Interfaces()
//if err != nil {
//return "", err
//}
//for _, iface := range ifaces {
//if iface.Flags&net.FlagUp == 0 {
//continue // interface down
//}
//if iface.Flags&net.FlagLoopback != 0 {
//continue // loopback interface
//}
//addrs, err := iface.Addrs()
//if err != nil {
//return "", err
//}
//for _, addr := range addrs {
//var ip net.IP
//switch v := addr.(type) {
//case *net.IPNet:
//ip = v.IP
//case *net.IPAddr:
//ip = v.IP
//}
//if ip == nil || ip.IsLoopback() {
//continue
//}
//ip = ip.To4()
//if ip == nil {
//continue // not an ipv4 address
//}
//return ip.String(), nil
//}
//}
//return "", errors.New("are you connected to the network?")
//}
