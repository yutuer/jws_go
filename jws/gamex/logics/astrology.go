package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//协议------星图系统:取自己的数据
type reqMsgAstrologyGetInfo struct {
	Req
}

type rspMsgAstrologyGetInfo struct {
	Resp
	Bag   []byte   `codec:"bag"`   //星魂背包 AstrologySoulBag
	Heros [][]byte `codec:"heros"` //武将镶嵌数据 []AstrologyHero
	Augur []byte   `codec:"augur"` //占星数据 AstrologyAugur
}

//AstrologyGetInfo ..
func (p *Account) AstrologyGetInfo(r servers.Request) *servers.Response {
	req := new(reqMsgAstrologyGetInfo)
	rsp := new(rspMsgAstrologyGetInfo)

	initReqRsp(
		"Attr/AstrologyGetInfoRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.Bag = []byte{}
	rsp.Heros = [][]byte{}
	rsp.Augur = []byte{}

	astrology := p.Profile.GetAstrology()

	rsp.Bag = encode(buildNetAstrologySoulBag(astrology.GetBag()))
	rsp.Augur = encode(buildNetAstrologyAugur(astrology.GetFactory()))
	heros := astrology.GetHeros()
	rsp.Heros = make([][]byte, 0, len(heros))
	for _, hero := range heros {
		rsp.Heros = append(rsp.Heros, encode(buildNetAstrologyHero(hero)))
	}

	return rpcSuccess(rsp)
}

//协议------星图系统:镶嵌
type reqMsgAstrologyInto struct {
	Req
	HeroID uint32 `codec:"hero"` //武将ID
	HoleID uint32 `codec:"hole"` //孔ID
	SoulID string `codec:"soul"` //星魂ID
}

type rspMsgAstrologyInto struct {
	SyncRespWithRewards
}

//AstrologyInto ..
func (p *Account) AstrologyInto(r servers.Request) *servers.Response {
	req := new(reqMsgAstrologyInto)
	rsp := new(rspMsgAstrologyInto)

	initReqRsp(
		"Attr/AstrologyIntoRsp",
		r.RawBytes,
		req, rsp, p)

	//检查参数
	if false == gamedata.CheckItemIsStarSoul(req.SoulID) {
		return rpcWarn(rsp, errCode.CommonInvalidParam)
	}

	//检查镶嵌条件
	soulCfg := gamedata.GetAstrologySoulCfg(req.SoulID)
	if soulCfg.GetEnableLevel() > int32(p.GetCorpLv()) {
		logs.Warn("[Astrology] AstrologyInto, Player Level(%d) < StarSoul(%s) EnableLevel(%d)",
			p.GetCorpLv(), req.SoulID, soulCfg.GetEnableLevel())
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}
	if soulCfg.GetPart() != gamedata.GetAstrologyHeroType(req.HeroID) {
		logs.Warn("[Astrology] AstrologyInto, Hero(%d) Star Soul Type(%s) is not StarSoul(%s) Type(%s) ",
			req.HeroID, gamedata.GetAstrologyHeroType(req.HeroID), req.SoulID, soulCfg.GetPart())
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}
	starLimit, levelLimit := gamedata.GetAstrologyIntoLimit(req.HoleID)
	if p.Profile.Hero.GetStar(int(req.HeroID)) < starLimit {
		logs.Warn("[Astrology] AstrologyInto, Hero(%d) Star(%d) < Hole(%d) Limit(%d)",
			req.HeroID, p.Profile.Hero.GetStar(int(req.HeroID)),
			req.HoleID, starLimit)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}
	if p.Profile.Hero.GetLevel(int(req.HeroID)) < levelLimit {
		logs.Warn("[Astrology] AstrologyInto, Hero(%d) Level(%d) < Hole(%d) Limit(%d)",
			req.HeroID, p.Profile.Hero.GetLevel(int(req.HeroID)),
			req.HoleID, levelLimit)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}

	//检查星魂数量
	astrology := p.Profile.GetAstrology()
	soul := astrology.GetBag().GetSoul(req.SoulID)
	if nil == soul || soul.Count < 1 {
		logs.Warn("[Astrology] AstrologyInto, Player Have not enough soul(%s)",
			req.SoulID)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}

	//扣除星魂
	if false == astrology.GetBag().SubSoul(req.SoulID, 1) {
		logs.Warn("[Astrology] AstrologyInto, SubSoul Failed",
			req.SoulID)
		return rpcWarn(rsp, errCode.CommonInner)
	}
	astrology.GetBag().UpdateSouls()

	//星魂装到武将身上
	astrologyHero := astrology.GetHero(req.HeroID)
	old := astrologyHero.IntoHole(req.HoleID, req.SoulID)

	//折算原星魂返还的费用
	if nil != old {
		materials := gamedata.AstrologyTranslateSoulToMaterial(old.HoleID, old.Rare, old.Upgrade)
		giveData := &gamedata.CostData{}
		for id, count := range materials {
			giveData.AddItem(id, count)
		}
		giveGroup := &account.GiveGroup{}
		giveGroup.AddCostData(giveData)

		if !giveGroup.GiveBySyncAuto(p.Account, rsp, fmt.Sprintf("AstrologyInto, replace StarSoul %d:%d:%d", old.HoleID, old.Rare, old.Upgrade)) {
			logs.SentryLogicCritical(p.AccountID.String(), "AstrologyInto Give Material Err by %d:%d:%d.",
				old.HoleID, old.Rare, old.Upgrade)
			logs.Error(fmt.Sprintf("[Astrology] AstrologyInto Give Material Err by %d:%d:%d.",
				old.HoleID, old.Rare, old.Upgrade))
		}
	}

	oldGs := p.Profile.GetData().CorpCurrGS

	p.Profile.GetData().SetNeedCheckMaxGS()
	rsp.OnChangeHeroStarMap()
	rsp.OnChangeSC()
	rsp.mkInfo(p)

	logiclog.LogAstrologyInto(
		p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		p.Profile.GetVip().V,
		req.HeroID,
		req.HoleID,
		uint32(soulCfg.GetRareLevel()),
		req.SoulID,
		oldGs,
		p.Profile.GetData().CorpCurrGS,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return rpcSuccess(rsp)
}

//协议------星图系统:分解:武将身上
type reqMsgAstrologyDestroyInHero struct {
	Req
	HeroID uint32 `codec:"hero"` //武将ID
	HoleID uint32 `codec:"hole"` //孔ID
}

type rspMsgAstrologyDestroyInHero struct {
	SyncRespWithRewards
}

//AstrologyDestroyInHero ..
func (p *Account) AstrologyDestroyInHero(r servers.Request) *servers.Response {
	req := new(reqMsgAstrologyDestroyInHero)
	rsp := new(rspMsgAstrologyDestroyInHero)

	initReqRsp(
		"Attr/AstrologyDestroyInHeroRsp",
		r.RawBytes,
		req, rsp, p)

	astrology := p.Profile.GetAstrology()
	astrologyHero := astrology.CheckHero(req.HeroID)
	if nil == astrologyHero {
		logs.Warn("[Astrology] AstrologyDestroyInHero, Hero(%d) has no StarSoul",
			req.HeroID)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}
	oldHole := astrologyHero.UnsetHole(req.HoleID)
	if nil == oldHole {
		logs.Warn("[Astrology] AstrologyDestroyInHero, Hero(%d) Hole(%d) has no StarSoul",
			req.HeroID, req.HoleID)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}

	//折算原星魂返还的费用
	materials := gamedata.AstrologyTranslateSoulToMaterial(oldHole.HoleID, oldHole.Rare, oldHole.Upgrade)
	giveData := &gamedata.CostData{}
	for id, count := range materials {
		giveData.AddItem(id, count)
	}
	giveGroup := &account.GiveGroup{}
	giveGroup.AddCostData(giveData)

	if !giveGroup.GiveBySyncAuto(p.Account, rsp, fmt.Sprintf("AstrologyDestroyInHero, replace StarSoul %d:%d:%d", oldHole.HoleID, oldHole.Rare, oldHole.Upgrade)) {
		logs.SentryLogicCritical(p.AccountID.String(), "AstrologyDestroyInHero Give Material Err by %d:%d:%d.",
			oldHole.HoleID, oldHole.Rare, oldHole.Upgrade)
		logs.Error(fmt.Sprintf("[Astrology] AstrologyDestroyInHero Give Material Err by %d:%d:%d.",
			oldHole.HoleID, oldHole.Rare, oldHole.Upgrade))
		return rpcWarn(rsp, errCode.CommonInner)
	}
	// 更新星图
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByAstrology.Add(&info)

	p.Profile.GetData().SetNeedCheckMaxGS()
	rsp.OnChangeHeroStarMap()
	rsp.OnChangeSC()
	rsp.mkInfo(p)

	soulCfg := gamedata.GetAstrologySoulIDByParam(req.HeroID, req.HoleID, oldHole.Rare)
	logiclog.LogAstrologyDestroy(
		p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		p.Profile.GetVip().V,
		req.HeroID,
		req.HoleID,
		oldHole.Rare,
		map[string]int64{soulCfg.GetID(): 1},
		oldHole.Upgrade,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return rpcSuccess(rsp)
}

//协议------星图系统:分解:背包单个
type reqMsgAstrologyDestroyInBag struct {
	Req
	SoulID string `codec:"soul"` //分解的星魂ID
}

type rspMsgAstrologyDestroyInBag struct {
	SyncRespWithRewards
}

//AstrologyDestroyInBag ..
func (p *Account) AstrologyDestroyInBag(r servers.Request) *servers.Response {
	req := new(reqMsgAstrologyDestroyInBag)
	rsp := new(rspMsgAstrologyDestroyInBag)

	initReqRsp(
		"Attr/AstrologyDestroyInBagRsp",
		r.RawBytes,
		req, rsp, p)

	astrology := p.Profile.GetAstrology()
	soul := astrology.GetBag().GetSoul(req.SoulID)
	if nil == soul || 0 >= soul.Count {
		logs.Warn("[Astrology] AstrologyDestroyInBag, Player have no StarSoul(%d)", req.SoulID)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}

	//从背包里面扣除
	if false == astrology.GetBag().SubSoul(req.SoulID, 1) {
		logs.Warn("[Astrology] AstrologyDestroyInBag, SubSoul Failed",
			req.SoulID)
		return rpcWarn(rsp, errCode.CommonInner)
	}
	astrology.GetBag().UpdateSouls()

	//折算原星魂返还的费用
	soulCfg := gamedata.GetAstrologySoulCfg(req.SoulID)
	materials := gamedata.AstrologyTranslateSoulToMaterial(soulCfg.GetStarHole(), uint32(soulCfg.GetRareLevel()), 0)
	giveData := &gamedata.CostData{}
	for id, count := range materials {
		giveData.AddItem(id, count)
	}
	giveGroup := &account.GiveGroup{}
	giveGroup.AddCostData(giveData)

	if !giveGroup.GiveBySyncAuto(p.Account, rsp, fmt.Sprintf("AstrologyDestroyInBag, StarSoul %d:%d:%d", soulCfg.GetStarHole(), uint32(soulCfg.GetRareLevel()), 0)) {
		logs.SentryLogicCritical(p.AccountID.String(), "AstrologyDestroyInBag Give Material Err by %d:%d:%d.",
			soulCfg.GetStarHole(), uint32(soulCfg.GetRareLevel()), 0)
		logs.Error(fmt.Sprintf("[Astrology] AstrologyDestroyInBag Give Material Err by %d:%d:%d.",
			soulCfg.GetStarHole(), uint32(soulCfg.GetRareLevel()), 0))
		return rpcWarn(rsp, errCode.CommonInner)
	}
	// 更新星图
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByAstrology.Add(&info)
	rsp.OnChangeSC()
	rsp.mkInfo(p)

	logiclog.LogAstrologyDestroy(
		p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		p.Profile.GetVip().V,
		0,
		soulCfg.GetStarHole(),
		uint32(soulCfg.GetRareLevel()),
		map[string]int64{req.SoulID: 1},
		0,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return rpcSuccess(rsp)
}

//协议------星图系统:分解:一键分解
type reqMsgAstrologyDestroySkip struct {
	Req
	Rares  []uint32 `codec:"rare"` //分解的品质数组
	HoleID uint32   `codec:"hole"` //孔位ID, 若为0, 则分解背包里对应所有孔位的相应品质星魂
	HeroID uint32   `codec:"hero"` //武将ID, 若为0, 则分解背包里对应所有武将类型的相应品质星魂
}

type rspMsgAstrologyDestroySkip struct {
	SyncRespWithRewards
}

//AstrologyDestroySkip ..
func (p *Account) AstrologyDestroySkip(r servers.Request) *servers.Response {
	req := new(reqMsgAstrologyDestroySkip)
	rsp := new(rspMsgAstrologyDestroySkip)

	initReqRsp(
		"Attr/AstrologyDestroySkipRsp",
		r.RawBytes,
		req, rsp, p)

	doRare := map[uint32]bool{}
	for _, r := range req.Rares {
		if r > gamedata.GetAstrologyResolveLimit() {
			return rpcWarn(rsp, errCode.CommonInvalidParam)
		}
		doRare[r] = true
	}

	astrology := p.Profile.GetAstrology()
	astrologyBag := astrology.GetBag()
	souls := astrologyBag.GetSouls()

	giveGroup := &account.GiveGroup{}
	didSouls := map[string]int64{}
	for _, soul := range souls {
		soulCfg := gamedata.GetAstrologySoulCfg(soul.SoulID)
		if false == doRare[uint32(soulCfg.GetRareLevel())] {
			continue
		}
		if 0 != req.HoleID {
			if req.HoleID != soulCfg.GetStarHole() {
				continue
			}
			if soulCfg.GetPart() != gamedata.GetAstrologyHeroType(req.HeroID) {
				continue
			}
		}

		//折算原星魂返还的费用
		materials := gamedata.AstrologyTranslateSoulToMaterial(soulCfg.GetStarHole(), uint32(soulCfg.GetRareLevel()), 0)
		giveData := &gamedata.CostData{}
		for id, count := range materials {
			giveData.AddItem(id, count*soul.Count)
		}
		giveGroup.AddCostData(giveData)

		//从背包里面扣除
		didSouls[soul.SoulID] = didSouls[soul.SoulID] + int64(soul.Count)
		astrologyBag.SubSoul(soul.SoulID, soul.Count)
	}
	astrologyBag.UpdateSouls()

	if !giveGroup.GiveBySyncAuto(p.Account, rsp, fmt.Sprintf("AstrologyDestroySkip, Rares %v", req.Rares)) {
		logs.SentryLogicCritical(p.AccountID.String(), "AstrologyDestroySkip Give Material Err by %v.", req.Rares)
		logs.Error(fmt.Sprintf("[Astrology] AstrologyDestroySkip Give Material Err by %v.", giveGroup))
		return rpcWarn(rsp, errCode.CommonInner)
	}

	rsp.OnChangeSC()
	rsp.mkInfo(p)

	logiclog.LogAstrologyDestroy(
		p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		p.Profile.GetVip().V,
		0,
		0,
		0,
		didSouls,
		0,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return rpcSuccess(rsp)
}

//协议------星图系统:星魂升级
type reqMsgAstrologySoulUpgrade struct {
	Req
	HeroID uint32 `codec:"hero"` //武将ID
	HoleID uint32 `codec:"hole"` //孔ID
}

type rspMsgAstrologySoulUpgrade struct {
	SyncResp
}

//AstrologySoulUpgrade ..
func (p *Account) AstrologySoulUpgrade(r servers.Request) *servers.Response {
	req := new(reqMsgAstrologySoulUpgrade)
	rsp := new(rspMsgAstrologySoulUpgrade)

	initReqRsp(
		"Attr/AstrologySoulUpgradeRsp",
		r.RawBytes,
		req, rsp, p)

	astrology := p.Profile.GetAstrology()
	astrologyHero := astrology.CheckHero(req.HeroID)
	if nil == astrologyHero {
		logs.Warn("[Astrology] AstrologySoulUpgrade, Hero(%d) has no StarSoul",
			req.HeroID)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}
	hole := astrologyHero.GetHole(req.HoleID)
	if nil == hole {
		logs.Warn("[Astrology] AstrologySoulUpgrade, Hero(%d) Hole(%d) has no StarSoul",
			req.HeroID, req.HoleID)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}

	newUpgrade := hole.Upgrade + 1
	if false == gamedata.CheckAstrologyUpgrade(hole.HoleID, hole.Rare, newUpgrade) {
		logs.Warn("[Astrology] AstrologySoulUpgrade, Hero(%d) Hole(%d) has no Upgrade (%d)",
			req.HeroID, req.HoleID, newUpgrade)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}

	//核算费用
	materials := gamedata.GetAstrologyUpgradeMaterial(hole.HoleID, hole.Rare, newUpgrade)
	costData := gamedata.CostData{}
	for t, c := range materials {
		costData.AddItem(t, c)
	}
	costGroup := &account.CostGroup{}
	if !costGroup.AddCostData(p.Account, &costData) {
		logs.SentryLogicCritical(p.AccountID.String(), "[Astrology] AstrologySoulUpgrade Cost Add Err, Need %+v.", materials)
		logs.Warn("[Astrology] AstrologySoulUpgrade, Cost Add Err, Need %+v.", materials)
		return rpcWarn(rsp, errCode.CommonConditionFalse)
	}
	//扣除费用
	if !costGroup.CostBySync(p.Account, rsp, fmt.Sprintf("AstrologySoulUpgrade")) {
		logs.SentryLogicCritical(p.AccountID.String(), "AstrologySoulUpgrade CostBySync.")
		logs.Error(fmt.Sprintf("AstrologySoulUpgrade CostBySync."))
		return rpcError(rsp, errCode.CommonInner)
	}

	//升级
	oldUpgrade := hole.Upgrade
	hole.Upgrade = newUpgrade

	oldGs := p.Profile.GetData().CorpCurrGS
	p.Profile.GetData().SetNeedCheckMaxGS()
	// 更新星图
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByAstrology.Add(&info)
	rsp.OnChangeHeroStarMap()
	rsp.OnChangeSC()
	rsp.mkInfo(p)

	soulCfg := gamedata.GetAstrologySoulIDByParam(req.HeroID, req.HoleID, hole.Rare)
	if nil != soulCfg {
		logiclog.LogAstrologyUpgrade(
			p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
			p.Profile.GetVip().V,
			req.HeroID,
			req.HoleID,
			hole.Rare,
			soulCfg.GetID(),
			oldUpgrade,
			newUpgrade,
			oldGs,
			p.Profile.GetData().CorpCurrGS,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			"")
	}

	return rpcSuccess(rsp)
}

//协议------星图系统:生产(占星)
type reqMsgAstrologyAugur struct {
	Req
	Skip bool `codec:"skip"` //是否为一键占星
}

type rspMsgAstrologyAugur struct {
	SyncResp
	Souls   []string `codec:"ss"`      //产生的星魂
	Augur   []byte   `codec:"augur"`   //占星数据 AstrologyAugur
	ACCost  uint32   `codec:"accost"`  //占星消耗的香火
	Process []uint32 `codec:"process"` //一键占星的过程
}

//AstrologyAugur ..
func (p *Account) AstrologyAugur(r servers.Request) *servers.Response {
	req := new(reqMsgAstrologyAugur)
	rsp := new(rspMsgAstrologyAugur)

	initReqRsp(
		"Attr/AstrologyAugurRsp",
		r.RawBytes,
		req, rsp, p)

	rsp.ACCost = 0
	rsp.Process = []uint32{}

	numLimit := uint32(1)
	if req.Skip {
		numLimit = gamedata.GetAstrologySkipNum()
	}

	//占星步骤
	astrology := p.Profile.GetAstrology()
	factory := astrology.GetFactory()
	giveDatas := gamedata.NewPriceDatas(int(numLimit))
	costGroup := &account.CostGroup{}
	count := uint32(0)
	oldAugurLevel := factory.CurrLevel
	for i := uint32(0); i < numLimit; i++ {
		elem := factory.GetCurrFactoryElem(p.GetRand())
		logs.Debug("[Astrology] AstrologyAugur GetCurrFactoryElem :%#v", elem)

		//占星费用
		augurCfg := gamedata.GetAstrologyAugurCfg(elem.AugurLevel)

		costData := gamedata.CostData{}
		costData.AddItem(augurCfg.GetAugurCoin(), augurCfg.GetAugurCoinCount())
		if !costGroup.AddCostData(p.Account, &costData) {
			break
		}
		rsp.ACCost += augurCfg.GetAugurCoinCount()

		//占星产出
		loot := elem.GetLoot(p.GetRand())
		giveData, err := gamedata.LootTemplateRand(p.GetRand(), loot)
		if nil != err {
			logs.Error(fmt.Sprintf("[Astrology] AstrologyAugur LootTemplateRand failed, %v", err))
			continue
		}
		giveDatas.AddOther(&giveData)

		count++

		//占星升级
		factory.TryUp(p.GetRand())
		rsp.Process = append(rsp.Process, factory.CurrLevel)
	}

	//扣除费用
	if !costGroup.CostBySync(p.Account, rsp, fmt.Sprintf("AstrologyAugur, num:%d", count)) {
		logs.SentryLogicCritical(p.AccountID.String(), "AstrologyAugur, num:%d.", count)
		logs.Error(fmt.Sprintf("AstrologyAugur, num:%d.", count))
		factory.CurrLevel = oldAugurLevel
		return rpcError(rsp, errCode.CommonInner)
	}

	//产出的星魂加入背包
	souls := map[string]int64{}
	astrologyBag := astrology.GetBag()
	for i := 0; i < giveDatas.Len(); i++ {
		has, itemID, num, _, _ := giveDatas.GetItem(i)
		if true == has {
			astrologyBag.AddSoul(itemID, num)
			for n := uint32(0); n < num; n++ {
				rsp.Souls = append(rsp.Souls, itemID)
			}
			souls[itemID] = souls[itemID] + int64(num)

			soulCfg := gamedata.GetAstrologySoulCfg(itemID)
			if uint32(soulCfg.GetRareLevel()) >= gamedata.GetAstrologyMarqueeMin() {
				for i := uint32(0); i < num; i++ {
					sysnotice.NewSysRollNotice(fmt.Sprintf("%d:%d", p.AccountID.GameId, p.AccountID.ShardId), gamedata.IDS_Astrology_Augur).
						AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
						AddParam(sysnotice.ParamType_ItemId, itemID).
						Send()
				}
			}
		}
	}

	//当前的占星状态
	rsp.Augur = encode(buildNetAstrologyAugur(factory))

	rsp.OnChangeSC()
	rsp.mkInfo(p)

	logiclog.LogAstrologyAugur(
		p.AccountID.String(), p.Profile.GetCurrAvatar(), p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		p.Profile.GetVip().V,
		oldAugurLevel,
		factory.CurrLevel,
		souls,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return rpcSuccess(rsp)
}
