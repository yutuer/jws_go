package util

import (
	"net/http"
	_ "net/http/pprof"

	"fmt"
	"os"

	"runtime"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	PProfPorts_start = 7010
	PProfPorts_end   = 7050
)

func PProfStart() {
	go func() {
		for port := PProfPorts_start; port <= PProfPorts_end; port++ {
			url := fmt.Sprintf(":%d", port)
			logs.Info("pprof try listen port %d ", port)
			mode := os.FileMode(0644)
			f, err := os.OpenFile("pprof", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				logs.Error("pprof start failed for can't open new file: %s", err)
				return
			}
			f.Write([]byte(url))
			f.Close()

			if err := http.ListenAndServe(url, nil); err != nil {
				logs.Warn("pprof server try %s err %v", url, err)
			} else {
				break
			}
		}
	}()
	runtime.SetBlockProfileRate(2)
}
