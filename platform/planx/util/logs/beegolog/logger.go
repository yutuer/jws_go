package beegolog

/*
In beego project main.go
import _ "vcs.taiyouxi.net/platform/planx/util/logs/beegolog"
beego.SetLogger("seelog", "")
*/

import (
	"github.com/astaxie/beego/logs"

	xlog "vcs.taiyouxi.net/platform/planx/util/logs"
)

type BeegoLogAdapter struct{}

func NewBeegoLogAdapter() logs.LoggerInterface {
	return &BeegoLogAdapter{}
}

func (l *BeegoLogAdapter) Init(config string) error {
	xlog.SetStackTraceDepth(2)
	return nil
}

func (l *BeegoLogAdapter) WriteMsg(msg string, level int) error {
	switch level {
	case logs.LevelEmergency:
		fallthrough
	case logs.LevelAlert:
		fallthrough
	case logs.LevelCritical:
		xlog.Critical(msg)
	case logs.LevelError:
		xlog.Error(msg)
	case logs.LevelWarning:
		fallthrough
	case logs.LevelNotice:
		xlog.Warn(msg)
	case logs.LevelInformational:
		xlog.Info(msg)
	case logs.LevelDebug:
		xlog.Debug(msg)
	}
	return nil
}

func (l *BeegoLogAdapter) Destroy() {
	xlog.Close()
}

func (l *BeegoLogAdapter) Flush() {
	xlog.Flush()
}

func init() {
	logs.Register("seelog", NewBeegoLogAdapter)
}
