package main

import (
	"os"
	"os/signal"
	"syscall"

	"sync"

	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/ban_account"
	"vcs.taiyouxi.net/platform/x/youmi_imcc/base"
)

var (
	WaitGroup sync.WaitGroup
	mir       base.MuteInfoRecorder
)

func startService() {
	go func() {
		mir.Start()
		WaitGroup.Add(1)
		defer WaitGroup.Done()
		logs.Debug("service start")
		for {
			select {
			case msg, ok := <-base.MsgChan:
				if ok {
					logs.Debug("receive msg %v", msg)
					if muteInfo := handleMsg(&msg); muteInfo != nil {
						optMuteInfo(muteInfo)
					}
				} else {
					logs.Debug("stop receive msg")
					return
				}
			case <-base.TimerChan:
				if muteInfos := handlerTimerMsg(base.MsgInfos); len(muteInfos) > 0 {
					for _, muteInfo := range muteInfos {
						optMuteInfo(muteInfo)
					}
				}
				base.ResetTimer()
			}
		}
	}()
	// it will block this go routine
	listenInterrupt()
}

func listenInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	stopService()
	os.Exit(1)
}

func stopService() {
	close(base.MsgChan)
	WaitGroup.Wait()
	mir.Stop()
	logs.Debug("service stop")
	logs.Close()
}

// gag rules
func handleMsg(netMsg *base.PlayerMsgInfoFromNet) *base.MuteInfo {
	msg := base.GenMsgInfo(netMsg)
	base.MsgInfos.AddMsg(msg)
	for i := 0; i < len(base.MsgHandler); i++ {
		if muteInfo := base.MsgHandler[i](msg); muteInfo != nil {
			return muteInfo
		}
	}
	return nil
}

func handlerTimerMsg(infos *base.MsgInfo) (ret []*base.MuteInfo) {
	ret = make([]*base.MuteInfo, 0)
	for i := 0; i < len(base.MsgTimerHandler); i++ {
		if muteInfo := base.MsgTimerHandler[i](infos); len(muteInfo) > 0 {
			ret = append(ret, muteInfo[:]...)
		}
	}
	return ret
}

func optMuteInfo(muteInfo *base.MuteInfo) {
	if muteInfo.IsMute {
		err := ban_account.MuteHMTSec(muteInfo.AcID, muteInfo.MuteTime)
		if err != nil {
			logs.Error("mute info:%v by err %v", *muteInfo, err)
		}
	}
	mir.Record(muteInfo, time.Now())
	base.MsgInfos.Remove(muteInfo.AcID)
}
