package logics

import (
	"fmt"
	"math"
	"math/rand"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ActivateExclusiveWeapon : 激活专属兵器
// 激活专属兵器
func (p *Account) ActivateExclusiveWeaponHandler(req *reqMsgActivateExclusiveWeapon, resp *rspMsgActivateExclusiveWeapon) uint32 {
	avatarId := req.AvatarId
	weapon := &p.Profile.GetHero().HeroExclusiveWeapon[avatarId]
	// 检查配置该武将是否有神兵
	if !gamedata.ContainsGloryWeapon(int(avatarId)) {
		logs.Warn("no glory weapon, avatarId=%d", avatarId)
		return errCode.ClickTooQuickly
	}
	// 条件判断
	if errCode := canActivateWeapon(p, int(avatarId), weapon); errCode != 0 {
		logs.Warn("can not activate weapon, avatarId=%d, errCode=%d", avatarId, errCode)
		return errCode
	}

	// 扣神兵碎片
	if errCode := costByActivateWeapon(p, int(avatarId), resp); errCode != 0 {
		logs.Warn("cost fail when activate weapon, avatarId=%d", avatarId)
		return errCode
	}
	// 激活
	weapon.OnActivate()
	vip, _ := p.Profile.GetVip().GetVIP()
	// 更新神兵等级排行榜
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByExclusiveWeapon.Add(&info)
	logiclog.LogCommonInfo(p.getBIBaseInfo(), logiclog.ActiveExclusiveWeapon{
		VIP:    vip,
		Avatar: int(avatarId),
		GS:     p.Profile.GetData().CorpCurrGS,
	}, logiclog.LogicTag_ActivateGloryWeapon, "")
	resp.OnChangeUpdateHeroExclusive(int(avatarId))
	return 0
}

func canActivateWeapon(p *Account, avatarId int, weapon *account.HeroExclusiveWeapon) uint32 {
	level := p.Profile.GetHero().HeroLevel[avatarId]
	startLevel := p.Profile.GetHero().GetStar(int(avatarId))
	unLockLevel := gamedata.GetHeroCommonConfig().GetCompanionUnlockLv()
	unLockStarLevel := gamedata.GetHeroCommonConfig().GetCompanionUnlockStar()
	if level < unLockLevel || startLevel < unLockStarLevel {
		return errCode.ClickTooQuickly
	}
	if weapon.IsActive {
		return errCode.ClickTooQuickly
	}
	return 0
}

func costByActivateWeapon(p *Account, avatarId int, resp *rspMsgActivateExclusiveWeapon) uint32 {
	costData := &gamedata.CostData{}
	costItemId, costItemCount := gamedata.GetActivateGloryWeaponCost(int(avatarId))
	if costItemId == "" {
		return errCode.ClickTooQuickly
	}
	costData.AddItem(costItemId, uint32(costItemCount))
	reason := fmt.Sprintf("activate exclusive weapon %d", avatarId)
	if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	return 0
}

// EvolveExclusiveWeapon : 专属兵器升品
// 专属兵器升品
func (p *Account) EvolveExclusiveWeaponHandler(req *reqMsgEvolveExclusiveWeapon, resp *rspMsgEvolveExclusiveWeapon) uint32 {
	avatarId := req.AvatarId
	weapon := &p.Profile.GetHero().HeroExclusiveWeapon[avatarId]
	if !weapon.IsActive {
		return errCode.ClickTooQuickly
	}
	if weapon.Quality >= int(gamedata.GloryWeaponCfg.MaxWeaponQuality) {
		return errCode.ClickTooQuickly
	}
	// 消耗材料
	costData := &gamedata.CostData{}
	costId, costCount := gamedata.GetEvolveGloryWeaponCost(int(avatarId), weapon.Quality+1)
	for i := range costId {
		costData.AddItem(costId[i], costCount[i])
	}
	reason := fmt.Sprintf("evolve exclusive weapon, avatarId=%d new quality=%d", avatarId, weapon.Quality+1)
	if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	// 升级
	weapon.Quality++

	// 自动装备 用==只加一次， 防止多加
	if weapon.Quality == int(gamedata.GetHeroCommonConfig().GetGWAppearanceUnlockQuality()) {
		errCode := autoEquipExclusiveWeapon(p, int(avatarId), resp)
		if errCode != 0 {
			return errCode
		}
	}

	vip, _ := p.Profile.GetVip().GetVIP()
	logiclog.LogCommonInfo(p.getBIBaseInfo(), logiclog.EvolveExclusiveWeapon{
		VIP:     vip,
		Avatar:  int(avatarId),
		GS:      p.Profile.GetData().CorpCurrGS,
		BeforeQ: weapon.Quality - 1,
		AfterQ:  weapon.Quality,
	}, logiclog.LogicTag_EnvolveGloryWeapon, "")
	p.Profile.GetData().SetNeedCheckMaxGS()
	// 更新神兵等级排行榜
	info := p.GetSimpleInfo()
	rank.GetModule(p.AccountID.ShardId).RankByExclusiveWeapon.Add(&info)
	resp.OnChangeUpdateHeroExclusive(int(avatarId))
	return 0
}

func autoEquipExclusiveWeapon(p *Account, avatarId int, resp *rspMsgEvolveExclusiveWeapon) uint32 {
	// 生成装备
	cfg := gamedata.GetGloryWeaponListCfg(avatarId)
	if cfg == nil {
		return errCode.ClickTooQuickly
	}
	account.AvatarGiveAndThenEquip(p.Account, avatarId, cfg.GetGloryWeapon(), gamedata.FashionPart_Weapon)
	resp.OnChangeFashionBag()
	resp.OnChangeAvatarEquip()
	return 0
}

// PromoteExclusiveWeapon : 培养兵器升品
// 培养兵器升品
const (
	promoteType_promote = iota
	promoteType_save
	promoteType_cancel
)

func (p *Account) PromoteExclusiveWeaponHandler(req *reqMsgPromoteExclusiveWeapon, resp *rspMsgPromoteExclusiveWeapon) uint32 {
	avatarId := req.AvatarId
	weapon := &p.Profile.GetHero().HeroExclusiveWeapon[avatarId]
	if !weapon.IsActive {
		return errCode.ClickTooQuickly
	}

	weapon.Clear()

	errCode := promoteWeapon(p, int(avatarId), weapon, req.PromoteByTen, resp)

	if errCode != 0 {
		return errCode
	}

	if weapon.CanSave() {
		weapon.Save()
	}
	p.Profile.GetData().SetNeedCheckMaxGS()
	resp.OnChangeUpdateHeroExclusive(int(avatarId))
	return errCode
}

func promoteWeapon(p *Account, avatarId int, weapon *account.HeroExclusiveWeapon, isTen bool, resp *rspMsgPromoteExclusiveWeapon) uint32 {
	// 消耗材料
	costData := &gamedata.CostData{}
	heroCfg := gamedata.GetHeroCommonConfig()
	realCount := getRealPromoteCount(p, isTen)
	if realCount <= 0 {
		return errCode.ClickTooQuickly
	}
	costData.AddItem(heroCfg.GetGloryWeaponDevelopCoin1(), heroCfg.GetGWDevelopCoinUnit1()*realCount)
	costData.AddItem(heroCfg.GetGloryWeaponDevelopCoin2(), heroCfg.GetGWDevelopCoinUnit2()*realCount)
	reason := fmt.Sprintf("evolve exclusive weapon, avatarId=%d", avatarId)
	if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}
	promoteCfg := gamedata.GetEvolveGloryWeaponCfg(weapon.Quality)
	weapon.ExtraAttr, weapon.ExtraHasAttr = randomAttrs(promoteCfg, realCount)
	weapon.PromoteCount += int(realCount)
	modifyAttrs(weapon, promoteCfg)
	resp.RealPromoteTime = int64(realCount)
	resp.PromoteByTen = isTen
	return 0
}

func getRealPromoteCount(p *Account, isTen bool) uint32 {
	if !isTen {
		return 1
	} else {
		coin1 := gamedata.GetHeroCommonConfig().GetGloryWeaponDevelopCoin1()
		coin2 := gamedata.GetHeroCommonConfig().GetGloryWeaponDevelopCoin2()
		ownCount1 := p.Profile.GetSC().GetSC(helper.SCId(coin1))
		ownCount2 := p.Profile.GetSC().GetSC(helper.SCId(coin2))
		realCount1 := math.Floor(float64(ownCount1 / int64(gamedata.GetHeroCommonConfig().GetGWDevelopCoinUnit1())))
		realCount2 := math.Floor(float64(ownCount2 / int64(gamedata.GetHeroCommonConfig().GetGWDevelopCoinUnit2())))
		realCount := math.Min(realCount1, realCount2)
		return uint32(math.Min(realCount, 10))
	}
}

func randomAttrs(promoteCfg *ProtobufGen.GLORYWEAPON, realCount uint32) ([account.ExclusiveWeaponMaxAttr]float32,
	[account.ExclusiveWeaponMaxAttr]bool) {
	var result [account.ExclusiveWeaponMaxAttr]float32 // attrId = index + 1
	var hasExtra [account.ExclusiveWeaponMaxAttr]bool
	var resultTemp [account.ExclusiveWeaponMaxAttr]float32 // attrId = index + 1
	var hasExtraTemp [account.ExclusiveWeaponMaxAttr]bool
	isAllNegative := true
	for i := 0; i < int(realCount); i++ {
		addParma, valueRet, hasRet := randomAttrByOnce(promoteCfg)
		if addParma == 1 {
			for j := 0; j < account.ExclusiveWeaponMaxAttr; j++ {
				result[j] += valueRet[j]
				if hasRet[j] {
					hasExtra[j] = true
				}
			}
			isAllNegative = false
		} else {
			for j := 0; j < account.ExclusiveWeaponMaxAttr; j++ {
				resultTemp[j] += valueRet[j]
				if hasRet[j] {
					hasExtraTemp[j] = true
				}
			}
		}
	}

	if isAllNegative {
		return resultTemp, hasExtraTemp
	} else {
		return result, hasExtra
	}
}

func randomAttrByOnce(promoteCfg *ProtobufGen.GLORYWEAPON) (int, [account.ExclusiveWeaponMaxAttr]float32, [account.ExclusiveWeaponMaxAttr]bool) {
	// 先随机1-3条属性
	randomCount := randomCount(promoteCfg)
	shuffleArray := util.ShuffleN(0, 5)
	randomResult := shuffleArray[:randomCount]
	var addParam int
	if rand.Int31n(100) < int32(promoteCfg.GetIsUpRate()*100) {
		addParam = 1
	} else {
		addParam = -1
	}

	var result [account.ExclusiveWeaponMaxAttr]float32 // attrId = index + 1
	var hasExtra [account.ExclusiveWeaponMaxAttr]bool

	// 再为每条属性随机一个值
	for _, randIndex := range randomResult {
		attrId := getAttrIdByIndex(randIndex)
		attrCfg := gamedata.GetWeaponAttrById(promoteCfg, attrId)
		attrValue := randomOneAttr(attrCfg.GetRandomInterval(), attrCfg.GetDecimal())
		result[attrId-1] = attrValue * float32(addParam)
		hasExtra[attrId-1] = true
	}
	return addParam, result, hasExtra
}

func randomCount(promoteCfg *ProtobufGen.GLORYWEAPON) int {
	index := util.RandomWeightInts([]int{int(promoteCfg.GetRandNumWight1()),
		int(promoteCfg.GetRandNumWight2()),
		int(promoteCfg.GetRandNumWight3())})
	return index + 1
}

func modifyAttrs(weapon *account.HeroExclusiveWeapon, promoteCfg *ProtobufGen.GLORYWEAPON) {
	for i, value := range weapon.ExtraAttr {
		if value != 0 {
			if weapon.Attr[i]+value < 0 {
				weapon.ExtraAttr[i] = -weapon.Attr[i]
				continue
			}
			attrCfg := gamedata.GetWeaponAttrById(promoteCfg, i+1)
			if weapon.Attr[i]+value > attrCfg.GetValue() {
				weapon.ExtraAttr[i] = attrCfg.GetValue() - weapon.Attr[i]
			}
		}
	}
}

func getAttrIdByIndex(n int) int {
	switch n {
	case 0:
		return gamedata.ATK
	case 1:
		return gamedata.DEF
	case 2:
		return gamedata.HP
	case 3:
		return gamedata.DE_CRI_RATE
	case 4:
		return gamedata.DE_CRI_DAMAGE
	}
	return gamedata.ATK
}

// 将随机区间根据随机位数转化成整数 比如 -0.0005 ~ 0.0005, 按保留4位随机，实际上等同于是在-5~5随机一个数
func randomOneAttr(policyId float32, decimal uint32) float32 {
	tempPow := math.Pow10(int(decimal)) // 扩大倍数
	policyCfg := gamedata.RandPolicyConfig(policyId)

	// 随机一个范围
	tempIntervalMin := int32(policyCfg.GetMinValue() * float32(tempPow)) // 实际随机范围
	tempIntervalMax := int32(policyCfg.GetMaxValue() * float32(tempPow))
	randNum := util.RandomInt(tempIntervalMin, tempIntervalMax)
	return float32(randNum) / float32(tempPow)
}

// ResetExclusiveWeapon : 重置神兵
// 重置神兵
func (p *Account) ResetExclusiveWeaponHandler(req *reqMsgResetExclusiveWeapon, resp *rspMsgResetExclusiveWeapon) uint32 {
	avatarId := req.AvatarId
	weapon := &p.Profile.GetHero().HeroExclusiveWeapon[avatarId]
	if !weapon.IsActive {
		return errCode.ClickTooQuickly
	}

	// 钻石消耗
	costHc := gamedata.GetHeroCommonConfig().GetGWRebornCost() * uint32(weapon.PromoteCount)
	costData := &gamedata.CostData{}
	costData.AddItem(gamedata.VI_Hc, costHc)
	reason := fmt.Sprintf("reset exclusive weapon, avatarId=%d", avatarId)
	if ok := account.CostBySync(p.Account, costData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}

	backRate := gamedata.GetHeroCommonConfig().GetGWRebornRate()
	if backRate > 1 {
		logs.Error("error config, reset exclusive weapon rate %f > 1", backRate)
		return errCode.ClickTooQuickly
	}

	// 重置所有培养属性
	count := weapon.PromoteCount
	weapon.ExtraAttr = [account.ExclusiveWeaponMaxAttr]float32{}
	weapon.Attr = [account.ExclusiveWeaponMaxAttr]float32{}
	weapon.ExtraHasAttr = [account.ExclusiveWeaponMaxAttr]bool{}
	weapon.PromoteCount = 0

	// 返还材料
	giveData := &gamedata.CostData{}
	coin1 := gamedata.GetHeroCommonConfig().GetGloryWeaponDevelopCoin1()
	coin2 := gamedata.GetHeroCommonConfig().GetGloryWeaponDevelopCoin2()
	unit1 := gamedata.GetHeroCommonConfig().GetGWDevelopCoinUnit1()
	unit2 := gamedata.GetHeroCommonConfig().GetGWDevelopCoinUnit2()
	giveData.AddItem(coin1, uint32(float32(unit1)*float32(count)*backRate))
	giveData.AddItem(coin2, uint32(float32(unit2)*float32(count)*backRate))
	reason = fmt.Sprintf("evolve exclusive weapon, avatarId=%d", avatarId)
	if ok := account.GiveBySync(p.Account, giveData, resp, reason); !ok {
		return errCode.ClickTooQuickly
	}

	p.Profile.GetData().SetNeedCheckMaxGS()
	resp.OnChangeUpdateHeroExclusive(int(avatarId))
	return 0
}
