package modules

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"errors"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type ServerModule interface {
	Start()
	AfterStart(g *gin.Engine)
	BeforeStop()
	Stop()
}

type GenServerModule func(sid uint) ServerModule

var (
	server_gen_modules = make(map[string]GenServerModule, 16)
	server_modules     = make(map[uint]map[string]ServerModule, 6)

	gins []*gin.Engine
)

func RegModule(name string, module GenServerModule) {
	_, ok := server_gen_modules[name]
	if ok {
		logs.Warn("module has been Reg by %s", name)
	}
	server_gen_modules[name] = module
}

func GetModule(shard uint, name string) ServerModule {
	modules_mng, ok := server_modules[shard]
	if ok {
		return modules_mng[name]
	}
	return nil
}

func StartModule(shards []uint) {
	defer func() {
		if err := recover(); err != nil {
			logs.Critical("StartModule Panic, Err %v", err)
			panic(fmt.Errorf("StartModule Panic, Err %v", err))
		}
	}()

	gins = make([]*gin.Engine, 0, 32)

	for _, shard := range shards {
		ginEngine := gin.New()
		n := 0

		modules_mng := server_modules[shard]
		if modules_mng == nil {
			modules_mng = make(map[string]ServerModule, 16)
			server_modules[shard] = modules_mng
		}
		for name, m := range server_gen_modules {
			module := m(shard)
			modules_mng[name] = module
			logs.Warn("Module %s New", name)
		}
		for _, name := range server_modules_seq {
			logs.Warn("Module %s Start", name)
			module := modules_mng[name]
			module.Start()
			n++
		}
		for _, name := range server_modules_seq {
			logs.Warn("Module %s AfterStart", name)
			m := modules_mng[name]
			m.AfterStart(ginEngine)
		}

		hasRun := false
		for i, sid := range game.Cfg.ShardId {
			if shard == uint(sid) {
				go func() {
					err := ginEngine.Run(game.Cfg.ListenPostAddr[i])
					if err != nil {
						logs.Critical(
							"ginEngine for modules run err By %v",
							err.Error())
						panic(err)
					}
				}()
				hasRun = true
				break
			}
		}

		gins = append(gins, ginEngine)

		if !hasRun {
			panic(errors.New("ginEngine for modules no post url cfg"))
		}
		logs.Debug("shard %d start module count %d", shard, n)
	}
}

//必须放在最后一个调用，至少所有请求都结束。
//moudle中的接口可能被Request直接调用
func StopModule() {
	for s, modules_mng := range server_modules {
		n := 0

		for i := len(server_modules_seq) - 1; i >= 0; i-- {
			name := server_modules_seq[i]
			m := modules_mng[name]
			logs.Warn("Module %s BeforeStop", name)
			m.BeforeStop()
		}

		for i := len(server_modules_seq) - 1; i >= 0; i-- {
			name := server_modules_seq[i]
			m := modules_mng[name]
			logs.Warn("Module %s Stop", name)
			m.Stop()
			n++
		}

		logs.Debug("shard %d Stop Module count %d", s, n)
	}
}
