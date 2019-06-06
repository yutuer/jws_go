package config

import (
	"fmt"

	gm "github.com/rcrowley/go-metrics"
	"vcs.taiyouxi.net/platform/planx/metrics"
)

const (
	notice_prefix = "notice"
)

var (
	noticeCCU_C gm.Counter
)

func InitNoticeMetrics() {
	NewNoticeCounter := func(name string) gm.Counter {
		return metrics.NewCounter(fmt.Sprintf("%s.%s.%s", notice_prefix, metrics.GetIPToken(), name))
	}
	noticeCCU_C = NewNoticeCounter("get")
}

func AddNoticeCCUCount() {
	noticeCCU_C.Inc(1)
}
