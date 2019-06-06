package logiclog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/timesking/seelog"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/eslogger"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func Start(c *cli.Context) {
	logTr := LogTransfer{
		mode:    c.String("mode"),
		logsC:   make(chan *string, 512),
		outPutC: make(chan string, 512),
		stopC:   make(chan os.Signal, 1),
	}
	switch logTr.mode {
	case "hero":
		logTr.outPutMode = &heroLog{}
	default:
		fmt.Fprintln(os.Stderr, "mode param err:", logTr.outPutMode)
		return
	}

	var wg util.WaitGroupWrapper
	wg.Wrap(func() { logTr.input(os.Stdin) })
	wg.Wrap(func() { logTr.handleLog() })
	wg.Wrap(func() { logTr.signalKillHandle() })
	wg.Wait()
}

type LogTransfer struct {
	mode       string
	logsC      chan *string
	outPutC    chan string
	outPutMode seelog.CustomReceiver
	stopC      chan os.Signal
}

func (logTr *LogTransfer) input(input io.Reader) {
	reader := bufio.NewReader(input)
	line := []byte{}
	for {
		b, isPrefix, err := reader.ReadLine()

		if len(b) <= 0 {
			break
		}
		if err != nil {
			logs.Error("reading standard input:%v", err)
			break
		}
		line = append(line, b...)

		if isPrefix {
			continue
		} else {
			var log eslogger.ESLoggerInfo
			err := json.Unmarshal(line, &log)
			if err != nil {
				logs.Error("reading standard input json.Unmarshal : %v line %s", err, string(b))
				continue
			}

			// 放入log
			logStr := string(line)
			logTr.logsC <- &logStr

			line = []byte{}
		}
	}
}

func (logTr *LogTransfer) handleLog() {
	if err := logTr.outPutMode.AfterParse(seelog.CustomReceiverInitArgs{}); err != nil {
		logs.Error("handleLog  outPutMode.AfterParse err: %v", err)
		return
	}
	for {
		select {
		case pLogStr := <-logTr.logsC:
			logTr.outPutMode.ReceiveMessage(*pLogStr, seelog.ErrorLvl, nil)
		case <-logTr.stopC:
			logTr.outPutMode.Close()
			return
		}
	}
}

func (logTr *LogTransfer) signalKillHandle() {
	signal.Notify(logTr.stopC, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	v := <-logTr.stopC
	logs.Info("Got %v by signal %v", v)
}
