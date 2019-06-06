package metrics

import (
	//"log"
	"net"
	//"os"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	gm "github.com/rcrowley/go-metrics"
)

func startGolangStats(stop <-chan struct{}) {
	logs.Info("metric.golang stats started.")

	if !_cfg.GraphiteValid {
		logs.Info("metric.golang stats exit, GraphiteValid is false .")
		return
	}

	if _cfg.GraphiteHost == "" {
		logs.Info("metric.golang stats exit, GraphiteHost is null .")
		return
	}

	if _cfg.GraphitePrefix == "" {
		_cfg.GraphitePrefix = "YouForgotPrefix"
	}
	_gcStats := gm.NewRegistry()

	if _cfg.GraphiteGCStats || _cfg.GraphiteMemStats {
		if _cfg.GraphiteGCStats {
			gm.RegisterDebugGCStats(_gcStats)
		}
		if _cfg.GraphiteMemStats {
			gm.RegisterRuntimeMemStats(_gcStats)
		}

		debugTick := time.NewTicker(5e9) // 5e9 == 5 seconds
		go func() {
			logs.Trace("debugTick Start")
			for {
				select {
				case <-stop:
					debugTick.Stop()
					logs.Trace("debugTick stop")
					return
				case <-debugTick.C:
					//logs.Trace("debugTick comes")
					if _cfg.GraphiteGCStats {
						gm.CaptureDebugGCStatsOnce(_gcStats)
					}
					if _cfg.GraphiteMemStats {
						gm.CaptureRuntimeMemStatsOnce(_gcStats)
					}
				}
			}
		}()
	}

	//if _cfg.GraphiteOutPutStdErr {
	//gm.Log(_gcStats, 60e9, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	//}
	if _cfg.GraphiteHost != "" {
		addr, _ := net.ResolveTCPAddr("tcp", _cfg.GraphiteHost)
		graphiteDebugTick := time.NewTicker(time.Duration(_cfg.GraphiteFlushInterval))
		graphiteDebugConfig := gm.GraphiteConfig{
			Addr:          addr,
			Registry:      _gcStats,
			FlushInterval: time.Duration(_cfg.GraphiteFlushInterval),
			Prefix:        _cfg.GraphitePrefix,
			DurationUnit:  time.Millisecond,
			Percentiles:   []float64{0.5, 0.75, 0.99, 0.999},
		}
		go func() {
			logs.Trace("graphiteDebugTick Start")
			for {
				select {
				case <-stop:
					logs.Trace("graphiteDebugTick Stop")
					graphiteDebugTick.Stop()
					return
				case <-graphiteDebugTick.C:
					// logs.Trace("graphiteDebugTick comes")
					if err := gm.GraphiteOnce(graphiteDebugConfig); err != nil {
						logs.Error("GraphiteDebugTick error with %s", err.Error())
					}
				}
			}
		}()

		graphiteCCUTick := time.NewTicker(time.Second)
		graphiteCCUConfig := gm.GraphiteConfig{
			Addr:          addr,
			Registry:      _goStatsPerSeond,
			FlushInterval: time.Second,
			Prefix:        _cfg.GraphitePrefix,
			DurationUnit:  time.Millisecond,
			Percentiles:   []float64{0.5, 0.75, 0.99, 0.999},
		}
		go func() {
			logs.Trace("graphiteCCUTick Start")
			for {
				select {
				case <-stop:
					logs.Trace("graphiteCCUTick Stop")
					graphiteCCUTick.Stop()
					return
				case <-graphiteCCUTick.C:
					// logs.Trace("graphiteCCUTick comes")
					if err := gm.GraphiteOnce(graphiteCCUConfig); err != nil {
						logs.Error("graphiteCCUTick error with %s", err.Error())
					}
				}
			}
		}()
		graphiteCustomTick := time.NewTicker(time.Second)
		graphiteCustomConfig := gm.GraphiteConfig{
			Addr:          addr,
			Registry:      _goCurstomStatsPerSecond,
			FlushInterval: time.Second,
			Prefix:        "",
			DurationUnit:  time.Millisecond,
			Percentiles:   []float64{0.5, 0.75, 0.99, 0.999},
		}
		go func() {
			logs.Trace("graphiteCCUTick Start")
			for {
				select {
				case <-stop:
					logs.Trace("graphiteCCUTick Stop")
					graphiteCustomTick.Stop()
					return
				case <-graphiteCustomTick.C:
					// logs.Trace("graphiteCCUTick comes")
					if err := gm.GraphiteOnce(graphiteCustomConfig); err != nil {
						logs.Error("graphiteCustomTick error with %s", err.Error())
					}
				}
			}
		}()
	}
}
