package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// ExpeditionHeroChoose : 远征英雄选择界面
// 用来传递远征玩法可以商场的英雄

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgExpeditionHeroChoose 远征英雄选择界面请求消息定义
type reqMsgExpeditionHeroChoose struct {
	Req
}

// rspMsgExpeditionHeroChoose 远征英雄选择界面回复消息定义
type rspMsgExpeditionHeroChoose struct {
	SyncRespWithRewards
	HeroId      []int64   `codec:"heroid"`  // 可选择的英雄ID
	HeroState   []int64   `codec:"hestate"` // 主将状态
	HeroBlood   []float32 `codec:"heblood"` // 主将血量
	HeroWuSkill []float32 `codec:"hewu"`    // 主将无双技能槽
	HeroExSkill []float32 `codec:"heex"`    // 主将普通技能槽
}

// ExpeditionHeroChoose 远征英雄选择界面: 用来传递远征玩法可以上场的英雄
func (p *Account) ExpeditionHeroChoose(r servers.Request) *servers.Response {
	req := new(reqMsgExpeditionHeroChoose)
	rsp := new(rspMsgExpeditionHeroChoose)

	initReqRsp(
		"Attr/ExpeditionHeroChooseRsp",
		r.RawBytes,
		req, rsp, p)
	pg := p.Profile.GetExpeditionInfo()
	for i := 0; i < helper.AVATAR_NUM_CURR; i++ {
		if p.Profile.GetCorp().IsAvatarHasUnlock(int(i)) {
			rsp.HeroId = append(rsp.HeroId, int64(i))
			rsp.HeroState = append(rsp.HeroState, int64(pg.ExpeditionMyHero[i].HeroIslive))
			rsp.HeroBlood = append(rsp.HeroBlood, pg.ExpeditionMyHero[i].HeroHp)
			rsp.HeroWuSkill = append(rsp.HeroWuSkill, pg.ExpeditionMyHero[i].HeroWuSkill)
			rsp.HeroExSkill = append(rsp.HeroExSkill, pg.ExpeditionMyHero[i].HeroExSkill)
		}
	}
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
