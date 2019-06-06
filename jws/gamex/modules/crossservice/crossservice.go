package crossservice

import (
	"github.com/gin-gonic/gin"

	"fmt"

	"time"

	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//CrossServiceModule ..
type CrossServiceModule struct {
	sid uint
	res *resources

	alive bool
}

type resources struct {
	sid    uint
	client *client

	shardList []uint32
	groupIDs  []uint32
}

func genCrossServiceModule(sid uint) *CrossServiceModule {
	cs := &CrossServiceModule{}

	mergeShardID := game.Cfg.GetShardIdByMerge(sid)
	cs.sid = mergeShardID
	cs.alive = false

	res := &resources{}
	res.sid = cs.sid
	cs.res = res

	res.shardList = []uint32{uint32(res.sid)}
	for _, ms := range game.Cfg.MergeRel {
		find := false
		for _, s := range res.shardList {
			if s == uint32(ms) {
				find = true
				break
			}
		}
		if false == find {
			res.shardList = append(res.shardList, uint32(ms))
		}
	}

	res.client = newClient(res)

	logs.Info("[CrossService] CrossServiceModule genCrossServiceModule")

	return cs
}

//Start ..
func (cs *CrossServiceModule) Start() {
	logs.Info("[CrossService] CrossServiceModule Start Begin ...")

	cs.res.groupIDs = getGroupIDs(cs.res.sid)

	if err := cs.res.client.start(); nil != err {
		logs.Error(fmt.Sprintf("[CrossService] CrossServiceModule Start failed, %v", err))
		return
	}
	cs.alive = true
	logs.Info("[CrossService] CrossServiceModule Start End ...")
}

//AfterStart ..
func (cs *CrossServiceModule) AfterStart(g *gin.Engine) {
}

//BeforeStop ..
func (cs *CrossServiceModule) BeforeStop() {

}

//Stop ..
func (cs *CrossServiceModule) Stop() {
	logs.Info("[CrossService] CrossServiceModule Stop Begin ...")
	cs.alive = false
	cs.res.client.stop()
	logs.Info("[CrossService] CrossServiceModule Stop End ...")
}

//CallSync ..
func (cs *CrossServiceModule) CallSync(groupID uint32, moName string, meName string, source string, param module.Param) (module.Ret, int, error) {
	if false == cs.alive {
		return nil, ErrNotAlive, fmt.Errorf("CallSync When CrossServiceModule is not Alive")
	}
	begin := time.Now()
	ret, errCode, err := cs.res.client.cs.CallSync(groupID, moName, meName, source, param)
	if message.ErrCodeOK == errCode {
		errCode = ErrOK
	}
	cost := time.Now().Sub(begin)
	logs.Info("[CrossService] CrossServiceModule CallSync cost %fs", cost.Seconds())
	return ret, errCode, err
}

//CallAsync ..
func (cs *CrossServiceModule) CallAsync(groupID uint32, moName string, meName string, source string, param module.Param) (int, error) {
	if false == cs.alive {
		return ErrNotAlive, fmt.Errorf("CallAsync When CrossServiceModule is not Alive")
	}
	begin := time.Now()
	errCode, err := cs.res.client.cs.CallAsync(groupID, moName, meName, source, param)
	if message.ErrCodeOK == errCode {
		errCode = ErrOK
	}
	cost := time.Now().Sub(begin)
	logs.Info("[CrossService] CrossServiceModule CallAsync cost %fs", cost.Seconds())
	return errCode, err
}
