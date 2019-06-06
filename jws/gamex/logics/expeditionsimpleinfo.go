package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
)

// Expeditionsimpleinfo : 查看基本主将信息
// 用来查看基本主将信息

// reqMsgExpeditionsimpleinfo 查看基本主将信息请求消息定义
type reqMsgExpeditionsimpleinfo struct {
	Req
	Enmy int64 `codec:"enmynum"` // 第几个敌人
}

// rspMsgExpeditionsimpleinfo 查看基本主将信息回复消息定义
type rspMsgExpeditionsimpleinfo struct {
	SyncResp
	EnmyId     string   `codec:"enmyid"`   // 敌人id
	EnmyName   string   `codec:"enmyname"` // 敌人姓名
	EnmyGs     int64    `codec:"enmygs"`   // 敌人战力
	EnmycLv    int64    `codec:"enmyclv"`
	EnmyHero   []int    `codec:"enmyhero"` // 敌人主将
	EnmyFAs    []uint32 `codec:"enmyfas"`  // 敌人主将等级
	EnmyStarLv []uint32 `codec:"enmystlv"` // 敌人主将星际

	EnmyHeroGs []int `codec:"enmyhgs"` // 敌人主将战力

	EnmyHp      []float32 `codec:"ehp"` //敌人的血量
	EnmyWuSkill []float32 `codec:"ews"` //敌人无双技能
	EnmyExSkill []float32 `codec:"exs"` //敌人普通技能
	EnmyState   int64     `codec:"est"` //敌人的状态 0没战斗过 1战斗过
}

// Expeditionsimpleinfo 查看基本主将信息: 用来查看基本主将信息
func (p *Account) Expeditionsimpleinfo(r servers.Request) *servers.Response {
	req := new(reqMsgExpeditionsimpleinfo)
	rsp := new(rspMsgExpeditionsimpleinfo)

	initReqRsp(
		"Attr/ExpeditionsimpleinfoRsp",
		r.RawBytes,
		req, rsp, p)
	rsp.EnmyId = p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[req.Enmy-1].Acid
	rsp.EnmyName = p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[req.Enmy-1].Name
	rsp.EnmyGs = int64(p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[req.Enmy-1].Gs)
	rsp.EnmycLv = int64(p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[req.Enmy-1].CorpLv)
	rsp.EnmyHero = p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[req.Enmy-1].HeroId[:]
	rsp.EnmyFAs = p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[req.Enmy-1].FAs[:]
	rsp.EnmyStarLv = p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[req.Enmy-1].FAStarLv[:]
	rsp.EnmyHeroGs = p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[req.Enmy-1].HeroGs[:]
	rsp.EnmyHp = p.Profile.GetExpeditionInfo().ExpeditionEnmySkillInfo[req.Enmy-1].Hp[:]
	rsp.EnmyWuSkill = p.Profile.GetExpeditionInfo().ExpeditionEnmySkillInfo[req.Enmy-1].WuSkill[:]
	rsp.EnmyExSkill = p.Profile.GetExpeditionInfo().ExpeditionEnmySkillInfo[req.Enmy-1].ExSkill[:]
	rsp.EnmyState = p.Profile.GetExpeditionInfo().ExpeditionEnmySkillInfo[req.Enmy-1].State

	rsp.OnChangerExpeditionInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

const DroidName = "IDS_EXPEDITION_BOT_NAME"

func (p *Account) SetExpeditionDroid(num int, info *account.ExpeditionEnmyInfo, dinfo *account.ExpeditionEnmyDetail) bool {
	//names := gamedata.RandRobotNames(1)
	levelUp := gamedata.GetExpeditionLvlCfgs()
	rCfg := gamedata.GetDroidForExpedition(p.Profile.GetCorp().Level - levelUp[num].GetLevelUp())
	TestHeroId := []int{0, 1, 2}
	HeroFirstStar := []uint32{2, 3, 3}
	TestHerolvl := []uint32{rCfg.CorpLv, rCfg.CorpLv, rCfg.CorpLv}
	HeroGs := []int{rCfg.HeroGs, rCfg.HeroGs, rCfg.HeroGs}
	p.Profile.GetExpeditionInfo().ExpeditionIds = append(p.Profile.GetExpeditionInfo().ExpeditionIds, string(num))
	info.Acid = rCfg.AcID
	info.CorpLv = rCfg.CorpLv
	info.Gs = rCfg.CorpGs
	info.Name = DroidName
	info.HeroId = TestHeroId[:]
	info.FAs = TestHerolvl[:]
	info.FAStarLv = HeroFirstStar
	info.HeroGs = HeroGs[:]
	for i := 0; i < 3; i++ {
		droid := gamedata.GetDroidForExpedition(p.Profile.GetCorp().Level - levelUp[num].GetLevelUp())
		a := helper.Avatar2Client{}
		err := account.FromAccountByDroid(&a, droid, i)
		if err != nil {
			return false
		}
		a.Name = info.Name
		dinfo.Enemies[i] = a
	}
	return true
}
