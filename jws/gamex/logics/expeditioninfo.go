package logics

import (
	"sort"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ExpeditionInfo : 远征界面协议
// 用来传输九个远征敌人的信息

const (
	_ = iota + 20
	Err_Not_Activate
)

// reqMsgExpeditionInfo 远征界面协议请求消息定义
type reqMsgExpeditionInfo struct {
	Req
	AccountID string `codec:"acid"` // 远征的玩家ID
}

// rspMsgExpeditionInfo 远征界面协议回复消息定义
type rspMsgExpeditionInfo struct {
	SyncResp
	ExpeditionIds     []string `codec:"expeditionids"` // 被远征的玩家ID
	ExpeditionNames   []string `codec:"ednames"`       // 被远征玩家姓名
	ExpeditionState   int64    `codec:"edstate"`       // 当前最远关卡
	ExpeditionAvard   int64    `codec:"edavard"`       // 当前最远宝箱
	ExpeditionNum     int64    `codec:"ednum"`         // 远征通关总计次数
	ExpeditionRestNum int64    `codec:"edrestnum"`     // 远征重置次数
	ExpeditionStep    bool     `codec:"edstep"`        //是否通过九关
}

// ExpeditionInfo 远征界面协议: 用来传输三个远征敌人的信息
func (p *Account) ExpeditionInfo(r servers.Request) *servers.Response {
	req := new(reqMsgExpeditionInfo)
	rsp := new(rspMsgExpeditionInfo)

	initReqRsp(
		"Attr/ExpeditionInfoRsp",
		r.RawBytes,
		req, rsp, p)

	pg := p.Profile.GetExpeditionInfo()
	now_t := p.Profile.GetProfileNowTime()
	if p.expeditionFirstActivate() {
		// 很大可能造成玩家多次点击
		p.Profile.GetExpeditionInfo().LoadEnemyToday(p.AccountID.String(),
			int64(p.Profile.GetData().CorpCurrGS_HistoryMax), now_t)
		if !p.setExpeditionEnmy() {
			logs.Warn("there is no Expedition enmy info")
			return rpcWarn(rsp, errCode.ClickTooQuickly)
		}
	}
	if p.Profile.GetExpeditionInfo().ExpeditionEnmyInfo[0].Acid == "" {
		if !p.setExpeditionEnmy() {
			logs.Error("there is no Expedition enmy info")
			return rpcError(rsp, 1)
		}
	}

	rsp.ExpeditionIds = pg.ExpeditionIds
	rsp.ExpeditionNames = p.Profile.GetExpeditionInfo().ExpeditionNames

	rsp.ExpeditionState = int64(pg.ExpeditionState)
	rsp.ExpeditionAvard = int64(pg.ExpeditionAward)
	rsp.ExpeditionNum = int64(pg.ExpeditionNum)
	rsp.ExpeditionRestNum = int64(pg.ExpeditionREstNum)
	rsp.ExpeditionStep = pg.ExpeditionStep
	rsp.OnChangerExpeditionInfo()
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) expeditionFirstActivate() bool {
	if !p.Profile.GetExpeditionInfo().IsActive {
		if account.CondCheck(gamedata.Mod_Expedition, p.Account) {
			p.Profile.GetExpeditionInfo().IsActive = true
			// 消耗一次重置次数
			if !p.Profile.GetCounts().Use(counter.CounterTypeExpedition, p.Account) {
				logs.Error("expeditionFirstActivate use counter failed")
			}
			p._initExpeditionInfo()
			logs.Debug("%s expedition unlock", p.AccountID.String())
			return true
		}
	}
	return false
}

func (p *Account) _initExpeditionInfo() {
	pg := p.Profile.GetExpeditionInfo()
	HpFirst := []float32{1, 1, 1}
	SkillFirst := []float32{0.5, 0.5, 0.5}
	WSkillFirst := []float32{0, 0, 0}
	for i := 0; i < 9; i++ {
		pg.ExpeditionEnmySkillInfo[i].State = 0
		pg.ExpeditionEnmySkillInfo[i].Hp = append([]float32{}, HpFirst[:]...)
		pg.ExpeditionEnmySkillInfo[i].ExSkill = append([]float32{}, SkillFirst[:]...)
		pg.ExpeditionEnmySkillInfo[i].WuSkill = append([]float32{}, WSkillFirst[:]...)
	}
	for i, _ := range p.Profile.GetExpeditionInfo().ExpeditionMyHero {
		n := &p.Profile.GetExpeditionInfo().ExpeditionMyHero[i]
		n.HeroExSkill = 0.5
		n.HeroHp = 1
		n.HeroIslive = 0
		n.HeroWuSkill = 0
	}
	pg.ExpeditionState = 1
	pg.ExpeditionAward = 1
	pg.ExpeditionStep = false

}

func (p *Account) setExpeditionEnmy() bool {
	ok, info := p.Profile.GetExpeditionInfo().GetEnemyToday(p.AccountID.String(), p.Account.GetProfileNowTime())
	if !ok {
		return false
	}
	logs.Debug("setExpeditionEnmy %s", p.AccountID.String())

	ep := p.Profile.GetExpeditionInfo()
	ep.ExpeditionEnmyInfo = info.ExpeditionEnmySimple
	ep.ExpeditionEnmyDetail = info.ExpeditionEnmyDetail
	enemies := make(enemys, len(ep.ExpeditionEnmyDetail))
	for i := 0; i < len(ep.ExpeditionEnmyInfo); i++ {
		sInfo := &ep.ExpeditionEnmyInfo[i]
		dInfo := &ep.ExpeditionEnmyDetail[i]
		if sInfo.Acid == "" || len(sInfo.HeroId) < 3 {
			if !p.SetExpeditionDroid(i, sInfo, dInfo) {
				return false
			}
		}
		enemies[i] = enemy{sInfo.Acid, sInfo.Gs, i}
	}
	sort.Sort(enemies)
	_ExpeditionEnmyInfo := [account.EXPEDITION_ENMY_NUM]account.ExpeditionEnmyInfo{}
	_ExpeditionEnmyDetail := [account.EXPEDITION_ENMY_NUM]account.ExpeditionEnmyDetail{}
	for i, e := range enemies {
		_ExpeditionEnmyInfo[i] = ep.ExpeditionEnmyInfo[e.Idx]
		_ExpeditionEnmyDetail[i] = ep.ExpeditionEnmyDetail[e.Idx]
	}
	ep.ExpeditionEnmyInfo = _ExpeditionEnmyInfo
	ep.ExpeditionEnmyDetail = _ExpeditionEnmyDetail

	for i, s := range ep.ExpeditionEnmyInfo {
		ds := ep.ExpeditionEnmyDetail[i]
		logs.Debug("GetEnemyToday simple %s %s %d", s.Acid, s.Name, s.HeroId)
		for _, d := range ds.Enemies {
			logs.Debug(""+
				" %s %s %d %s", d.Acid, d.Name, d.AvatarId, d.DestinyGeneralsID)
		}
	}
	return true
}

type enemy struct {
	Acid  string
	Score int
	Idx   int
}

type enemys []enemy

func (pq enemys) Len() int { return len(pq) }

func (pq enemys) Less(i, j int) bool {
	return pq[i].Score < pq[j].Score
}

func (pq enemys) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
