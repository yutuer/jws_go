package cs_metrics

import (
	gm "github.com/rcrowley/go-metrics"
	"vcs.taiyouxi.net/platform/planx/metrics"
)

var (
	connMetric gm.Gauge
	optMetric  gm.Counter
)

func Reg() {
	connMetric = metrics.NewGauge("ConnNum")
	optMetric = metrics.NewCounter("OptNum")
}

func UpdateConn(value int64) {
	connMetric.Update(value)
}

func AddOpt(value int64) {
	optMetric.Inc(value)
}
