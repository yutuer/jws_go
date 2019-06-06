package util

import (
	"fmt"
	"os"
	"time"

	"sync"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/leaktest"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	LeakTestPorts_start = 7051
	LeakTestPorts_end   = 7080
)

var (
	lsnapshot_orig map[string]bool

	mutx sync.RWMutex
)

func LeakTestStart() {
	leakTestSnapshot()

	g := gin.Default()
	g.GET("/leaktest/snapshot", func(c *gin.Context) {
		leakTestSnapshot()
		c.String(200, "ok")
	})
	g.GET("/leaktest/check", func(c *gin.Context) {
		mutx.RLock()

		ts := time.Now().Format("2006-1-2 15:04:05")
		fmt.Fprintf(os.Stderr, "[leak-test] %s leakTestCheck start \r\n", ts)

		lt := &leakTestReporter{
			msg: make([]string, 0, 128),
		}
		leaktest.Check2(lsnapshot_orig, lt)
		if lt.failed {
			fmt.Fprintf(os.Stderr, "[leak-test] %s found leak goroutine %d \r\n",
				ts, len(lt.msg))
			fmt.Fprintln(os.Stderr, lt.msg)
		}
		fmt.Fprintf(os.Stderr, "[leak-test] %s leakTestCheck end \r\n",
			time.Now().Format("2006-1-2 15:04:05"))

		mutx.RUnlock()

		c.String(200, "ok")
	})
	go func() {
		for port := LeakTestPorts_start; port <= LeakTestPorts_end; port++ {
			url := fmt.Sprintf(":%d", port)
			logs.Info("leaktest try listen port %d ", port)
			mode := os.FileMode(0644)
			f, err := os.OpenFile("leaktest", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				logs.Error("leaktest start failed for can't open new file: %s", err)
				return
			}
			f.Write([]byte(url))
			f.Close()

			if err := g.Run(url); err != nil {
				logs.Warn("leaktest server try %s err %v", url, err)
			} else {
				break
			}
		}
	}()
}

func leakTestSnapshot() {
	ts := time.Now().Format("2006-1-2 15:04:05")

	mutx.Lock()
	defer mutx.Unlock()

	fmt.Fprintf(os.Stderr, "[leak-test] %s leakTestSnapshot \r\n", ts)

	lsnapshot_orig = leaktest.Snapshot()
}

type leakTestReporter struct {
	failed bool
	msg    []string
}

func (tr *leakTestReporter) Errorf(format string, args ...interface{}) {
	tr.failed = true
	tr.msg = append(tr.msg, fmt.Sprintf(format+"\r\n", args))
}
