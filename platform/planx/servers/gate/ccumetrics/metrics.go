package ccumetrics

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	ShardID  []uint
	PublicIP string
	RpcIP    string
	_cfg     Config
	_login   *loginConnector
	isLocal  bool
)

func Start(cfg Config, shardId []uint, publicip, rpcip string, islocal bool) {
	logs.Info("gate ccumetric started.")
	_cfg = cfg
	ShardID = shardId
	PublicIP = publicip
	RpcIP = rpcip
	isLocal = islocal

	dataprefix := fmt.Sprintf("gatex.%s", metrics.GetIPToken())
	_ccu = metrics.NewCounter(fmt.Sprintf("%s.ccu", dataprefix))
	_handShakeCCU = metrics.NewCounter(fmt.Sprintf("%s.hs_ccu", dataprefix))

	_rps = metrics.NewMeter(fmt.Sprintf("%s.requests", dataprefix))
	_nRequests = metrics.NewCounter(fmt.Sprintf("%s.request_total", dataprefix))

	_login = newLoginConnector(cfg)
	_login.start()
}

func Stop() {
	_login.stop()
}

func NotifyLogInOff(s LoginStatus) {
	_login.loginStatusChan <- s
}
