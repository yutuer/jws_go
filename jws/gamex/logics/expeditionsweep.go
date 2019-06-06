package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// ExpeditionSweep : 扫荡远征
// 扫荡远征

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgExpeditionSweep 扫荡远征请求消息定义
type reqMsgExpeditionSweep struct {
	Req
	FightHeroId []int64 `codec:"fight_hero_id"` // 上阵的武将ID
	EnemyId     int64   `codec:"enemy_id"`      // 挑战的关卡
}

// rspMsgExpeditionSweep 扫荡远征回复消息定义
type rspMsgExpeditionSweep struct {
	SyncRespWithRewards
	FightResult [][]byte `codec:"fight_ret"` // 每场战斗的细节
	IsWin       int64    `codec:"is_win"`    //是否胜利 0 是失败，1是胜利
}

// ExpeditionSweep 扫荡远征: 扫荡远征
func (p *Account) ExpeditionSweep(r servers.Request) *servers.Response {
	req := new(reqMsgExpeditionSweep)
	rsp := new(rspMsgExpeditionSweep)

	initReqRsp(
		"Attr/ExpeditionSweepRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ExpeditionSweepHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end
	rsp.OnChangerExpeditionInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// ExpeditionFightDetail 扫荡远征
type ExpeditionFightDetail struct {
	AvatarId      int64 `codec:"avatar"`   // 武将的ID
	BeforeHp      int64 `codec:"be_hp"`    // 战斗前血量
	AfterHp       int64 `codec:"af_hp"`    // 战斗后血量
	EnemyId       int64 `codec:"enemy"`    // 敌方武将的ID
	EnemyBeforeHp int64 `codec:"en_be_hp"` // 敌方战斗前血量
	EnemyAfterHp  int64 `codec:"en_af_hp"` // 敌方战斗后血量
}
