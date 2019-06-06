package util

import (
	"time"

	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
)

var (
	tWheel_quick *timingwheel.TimingWheel
	tWheel_slow  *timingwheel.TimingWheel
)

func init() {
	tWheel_quick = timingwheel.NewTimingWheel(10*time.Millisecond, 100*30)
	tWheel_slow = timingwheel.NewTimingWheel(time.Second, 300)
}

func GetQuickTimeWheel() *timingwheel.TimingWheel {
	return tWheel_quick
}

func GetSlowTimeWheel() *timingwheel.TimingWheel {
	return tWheel_slow
}
