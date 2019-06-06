package global_info

import "github.com/gin-gonic/gin"

func genGlobalInfoModule(sid uint) *GlobalInfoModule {
	return &GlobalInfoModule{
		sid:          sid,
		gLevelFinish: &GlobalLevelFinishInfo{},
	}
}

type GlobalInfoModule struct {
	sid          uint
	gLevelFinish *GlobalLevelFinishInfo
}

func (r *GlobalInfoModule) AfterStart(g *gin.Engine) {
}

func (r *GlobalInfoModule) BeforeStop() {
}

func (r *GlobalInfoModule) Start() {
	r.gLevelFinish.start(r.sid)
}

func (r *GlobalInfoModule) Stop() {
	r.gLevelFinish.stop()
}

func OnLevelFinish(shardId uint, levelId, acid, name string) {
	GetModule(shardId).gLevelFinish.levelFinishReq(typ_Level, levelId, acid, name)
}

func OnTrialFinish(shardId uint, levelId, acid, name string) {
	GetModule(shardId).gLevelFinish.levelFinishReq(typ_Trial, levelId, acid, name)
}

func OnBossFinish(shardId uint, levelId, acid, name string) {
	GetModule(shardId).gLevelFinish.levelFinishReq(typ_Boss, levelId, acid, name)
}
