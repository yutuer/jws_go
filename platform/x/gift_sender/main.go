package main

import (
	"os"

	"github.com/gin-gonic/gin"

	fmt "fmt"

	"sync"

	"github.com/fvbock/endless"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gift_sender/config"
	"vcs.taiyouxi.net/platform/x/gift_sender/core"
	"vcs.taiyouxi.net/platform/x/gift_sender/sdk/alibaba_uc"
	"vcs.taiyouxi.net/platform/x/gift_sender/sdk/hero"
	"vcs.taiyouxi.net/platform/x/gift_sender/sdk/hmt_gift"
)

func main() {
	defer logs.Close()

	config.LoadConfig("conf/config.toml")

	logs.Debug("Config: %v", config.CommonConfig)

	config.LoadGameData("")

	err := etcd.InitClient(config.CommonConfig.EtcdEndPoint)
	if err != nil {
		logs.Error("etcd InitClient err %s", err.Error())
		logs.Critical("\n")
		logs.Close()
		os.Exit(1)
	}
	logs.Info("etcd Init client done.")

	core.InitDynamoDB(config.CommonConfig.MailDB)

	Run()
}

func Run() {
	core.Waitter = sync.WaitGroup{}
	for k, v := range config.CommonConfig.Host {
		var handler core.Handle
		switch k {
		case core.HeroFlag:
			handler = &hero.Handler
		case core.AlibabaUCFlag:
			handler = &alibaba_uc.Handler
		case core.HMT:
			handler = &hmt_gift.Handler
		}
		handler.GenError()
		if handler != nil {
			engine := gin.Default()
			err := handler.Reg(engine, config.CommonConfig)
			if err != nil {
				panic(fmt.Sprintf("reg handle err by %v", err))
			}
			core.Waitter.Add(1)
			go func(host string) {
				defer core.Waitter.Done()
				endless.ListenAndServe(host, engine)
			}(v.Host)
		}
	}
	core.Waitter.Wait()
}
