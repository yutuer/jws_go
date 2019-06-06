package ccumetrics

import gm "github.com/rcrowley/go-metrics"

//尝试连接进行handshake的玩家
var _ccu gm.Counter

//合法handshake通过的玩家
var _handShakeCCU gm.Counter

func AddCCU() {
	if _ccu != nil {
		_ccu.Inc(1)
	}
}

func ReduceCCU() {
	if _ccu != nil {
		_ccu.Dec(1)
	}
}

func addHandShakeCCU() {
	if _handShakeCCU != nil {
		_handShakeCCU.Inc(1)
	}
}

func reduceHandShakeCCU() {
	if _handShakeCCU != nil {
		_handShakeCCU.Dec(1)
	}
}

//##############
//The number of Request单调增长
var _nRequests gm.Counter

//Request per second
var _rps gm.Meter

func CountAddNewRequest(n int64) {
	if _rps != nil {
		_rps.Mark(n)
	}
	if _nRequests != nil {
		_nRequests.Inc(n)
	}
}
