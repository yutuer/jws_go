package account

import (
	"fmt"
	"time"

	"strings"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/counter"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	guild2 "vcs.taiyouxi.net/jws/gamex/models/guild"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/interfaces"
	"vcs.taiyouxi.net/jws/gamex/models/pay"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/hour_log"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 赠与接口
// 游戏中经常要赠与玩家一系列物品、软通等
// 这是一个统一的接口，可以生成一个列表，并一次性赠与玩家
// 由于类似与精炼返还一类的需求，赠与数据支持一些简单的运算
//
//
//
//

type GiveIAPData struct {
	IAPGoodIndex   uint32
	IAPGameOrderId string
	IAPOrder       string
	IAPPrice       uint32
	IAPPlatform    string
	IAPChannel     string
	IAPPayTime     string // 客户端带过来的支付时间
	IAPTrueAmount  uint32
	IAPPkgInfo     gamedata.PackageInfo
	IAPPayType     string
}

type GiveGroup struct {
	items      []string
	item_datas []gamedata.BagItemData
	count      []uint32

	general []string
	gcount  []uint32

	heroPiece               [AVATAR_NUM_MAX]uint32
	heroPieceWholeChar      [AVATAR_NUM_MAX]uint32
	heroPieceWholeCharCount [AVATAR_NUM_MAX]uint32

	sc                [SC_TYPE_COUNT]int64
	goldLevelPoint    uint32
	expLevelPoint     uint32
	dcLevelPoint      uint32
	corp_xp           uint32
	avatar_xp_all     uint32
	avatarIds         []int
	energy            uint32
	wheelcoin         uint32
	bossFightPoint    uint32
	EStageId          string
	EStageTimes       uint32
	HeroTalentPoint   uint32
	GameModeId        int
	GameModeTime      int
	GuildXp           uint32
	GuildScienceValid gamedata.GST_Typ
	GateEnemyBonus    float32
	// IAP
	IAPData *GiveIAPData

	avaters_xp map[int]uint32

	hc [HC_TYPE_COUNT]int64

	starsouls map[string]uint32

	giveHcFromVip int // 如果这个不为0 则按照vip等级中的Hc数量给hc, 之所以只加一个标签, 是因为给的数额是和VIP等级相关的

	price2Client gamedata.CostData2Client
}

func GiveBySync(
	p *Account,
	data *gamedata.CostData,
	sync interfaces.ISyncRspWithRewards,
	reason string) bool {

	g := GiveGroup{}
	g.AddCostData(data)
	return g.GiveBySyncAuto(p, sync, reason)
}

func GiveBySyncWithoutMerge(
	p *Account,
	data *gamedata.CostData,
	sync interfaces.ISyncRspWithRewards,
	reason string) bool {

	g := GiveGroup{}
	g.AddCostData(data)
	return g.GiveBySyncAutoWithoutMerge(p, sync, reason)
}

func GiveBySyncWith2Client(
	p *Account,
	data *gamedata.CostData,
	sync helper.ISyncRsp,
	reason string) (bool, *gamedata.CostData2Client) {

	g := GiveGroup{}
	g.AddCostData(data)
	return g.GiveBySyncWithRes(p, sync, reason)
}

func (g *GiveGroup) addSc(sc_t int, sc_v int64) {
	if sc_t >= SC_TYPE_COUNT || sc_t < 0 {
		logs.Error("GiveGroup Add Sc Err Typ %d", sc_t)
		return
	}

	g.sc[sc_t] += sc_v
}

func (g *GiveGroup) addHc(hc_t int, hc_v int64) {
	if hc_t >= HC_TYPE_COUNT || hc_t < 0 {
		logs.Error("GiveGroup Add Sc Err Typ %d", hc_t)
		return
	}

	g.hc[hc_t] += hc_v
}

func (g *GiveGroup) addGoldLevelPoint(glp uint32) {
	g.goldLevelPoint += glp
}

func (g *GiveGroup) addExpLevelPoint(elp uint32) {
	g.expLevelPoint += elp
}

func (g *GiveGroup) addDCLevelPoint(dclp uint32) {
	g.dcLevelPoint += dclp
}

func (g *GiveGroup) addCorpXp(xp uint32) {
	g.corp_xp += xp
}

func (g *GiveGroup) addAvatarAllXp(xp uint32) {
	g.avatar_xp_all += xp
}

func (g *GiveGroup) addAvatarXp(avatars []int, addxp uint32) {
	if g.avaters_xp == nil {
		g.avaters_xp = map[int]uint32{}
	}
	for _, avatar := range avatars {
		xp, ok := g.avaters_xp[avatar]
		if !ok {
			g.avaters_xp[avatar] = addxp
		} else {
			g.avaters_xp[avatar] = xp + addxp
		}
	}
}

func (g *GiveGroup) addGeneralGoodwill(gid string, gcount uint32) {
	for idx, id := range g.general {
		if id == gid {
			g.gcount[idx] += gcount
			return
		}
	}

	g.general = append(g.general, gid)
	g.gcount = append(g.gcount, gcount)
}

func (g *GiveGroup) addStarSoul(soulID string, num uint32) {
	if nil == g.starsouls {
		g.starsouls = map[string]uint32{}
	}
	g.starsouls[soulID] = g.starsouls[soulID] + num
}

func (g *GiveGroup) addItem(item_id string, count uint32) {
	for idx, id := range g.items {
		if id == item_id {
			g.count[idx] += count
			return
		}
	}

	g.items = append(g.items, item_id)
	g.count = append(g.count, count)
	g.item_datas = append(g.item_datas, gamedata.BagItemData{})
}

func (g *GiveGroup) addItemWithData(item_id string, data gamedata.BagItemData, count uint32) {
	// 当指定Data时两个id一样的装备也不一样
	if data.IsNil() {
		for idx, id := range g.items {
			if id == item_id {
				g.count[idx] += count
				return
			}
		}
	}

	g.items = append(g.items, item_id)
	g.count = append(g.count, count)
	g.item_datas = append(g.item_datas, data)
}

func (g *GiveGroup) AddItem(itemId string, count uint32) {
	cost := gamedata.CostData{}
	cost.AddItem(itemId, count)
	g.AddCostData(&cost)
}

func (g *GiveGroup) AddItemWithData(itemId string, data gamedata.BagItemData, count uint32) {
	cost := gamedata.CostData{}
	cost.AddItemWithData(itemId, data, count)
	g.AddCostData(&cost)
}

func (g *GiveGroup) AddOther(other *GiveGroup) {
	for sc_t, sc_v := range other.sc {
		g.addSc(sc_t, sc_v)
	}

	g.addGoldLevelPoint(other.goldLevelPoint)
	g.addExpLevelPoint(other.expLevelPoint)
	g.addDCLevelPoint(other.dcLevelPoint)

	for hc_t, hc_v := range other.hc {
		g.addHc(hc_t, hc_v)
	}

	for idx, item_id := range other.items {
		g.addItemWithData(
			item_id,
			other.item_datas[idx],
			other.count[idx])
	}

	for idx, item_id := range other.general {
		g.addGeneralGoodwill(item_id, other.gcount[idx])
	}

	for idx, v := range other.heroPiece {
		g.heroPiece[idx] += v
	}

	for idx, v := range other.heroPieceWholeChar {
		g.heroPieceWholeChar[idx] += v
	}

	for idx, v := range other.heroPieceWholeCharCount {
		g.heroPieceWholeCharCount[idx] = v
	}

	if nil != other.starsouls {
		for id, c := range other.starsouls {
			g.addStarSoul(id, c)
		}
	}

	g.addAvatarXp(other.avatarIds, other.avatar_xp_all)
	g.addCorpXp(other.corp_xp)
	g.energy += other.energy
	g.wheelcoin += other.wheelcoin
	g.bossFightPoint += other.bossFightPoint
	g.giveHcFromVip = other.giveHcFromVip
	g.GuildXp += other.GuildXp
	g.GuildScienceValid = other.GuildScienceValid
	g.GateEnemyBonus = other.GateEnemyBonus

	g.price2Client.AddOther2Client(&other.price2Client)
}

func (g *GiveGroup) AddCostData(data *gamedata.CostData) {
	for sc_t, sc_v := range data.Sc {
		g.addSc(sc_t, sc_v)
	}

	g.addGoldLevelPoint(data.GoldLevelPoint)
	g.addExpLevelPoint(data.ExpLevelPoint)
	g.addDCLevelPoint(data.DCLevelPoint)

	for hc_t, hc_v := range data.Hc {
		g.addHc(hc_t, hc_v)
	}

	for idx, item_id := range data.Items {
		g.addItemWithData(
			item_id,
			data.ItemData[idx],
			data.Count[idx])
	}

	for idx, item_id := range data.Generals {
		g.addGeneralGoodwill(item_id, data.GCount[idx])
	}

	for idx, v := range data.HeroPiece {
		g.heroPiece[idx] += v
	}

	g.addAvatarXp(data.AvatarIds, data.AvatarXp)
	g.addCorpXp(data.CorpXp)
	g.energy += data.Energy
	g.wheelcoin += data.WheelCoin
	g.bossFightPoint += data.BossFightPoint
	g.giveHcFromVip = data.GiveHcFromVip
	g.EStageId = data.EStageId
	g.EStageTimes = data.EStageTimes
	g.HeroTalentPoint = data.HeroTalentPoint
	g.GameModeId = data.GameModeId
	g.GameModeTime = data.GameModeTimes
	g.GuildXp = data.GuildXp
	g.GuildScienceValid = data.GuildScienceValid
	g.GateEnemyBonus = data.GateEnemyBonus

	g.IAPData = &GiveIAPData{}
	g.IAPData.IAPGoodIndex = data.IAPGoodIndex
	g.IAPData.IAPGameOrderId = data.IAPGameOrderId
	g.IAPData.IAPOrder = data.IAPGoodOrder
	g.IAPData.IAPPrice = data.IAPMoney
	g.IAPData.IAPPlatform = data.IAPPlatform
	g.IAPData.IAPChannel = data.IAPChannel
	g.IAPData.IAPPayTime = data.IAPPayTime
	g.IAPData.IAPTrueAmount = data.IAPTrueAmount
	g.IAPData.IAPPkgInfo = data.IAPPkgInfo
	g.IAPData.IAPPayType = data.IAPPayType
	for idx, v := range data.HeroPieceWholeChar {
		g.heroPieceWholeChar[idx] += v
	}

	for idx, v := range data.HeroPieceWholeCharCount {
		g.heroPieceWholeCharCount[idx] = v
	}

	if nil != data.StarSouls {
		for id, c := range data.StarSouls {
			g.addStarSoul(id, c)
		}
	}

	g.price2Client.AddOther2Client(data)
}

func (g *GiveGroup) GiveBySyncWithRes(p *Account, sync helper.ISyncRsp, reason string) (bool, *gamedata.CostData2Client) {
	logs.Trace("Give %v", *g)
	aid := p.AccountID.String()
	pt := p.Profile.GetTitle()
	typ := COND_TYP_ExchangeShop //需要获取的Title类型
	res := gamedata.CostData2Client{}
	res.Init2Client(32)
	for i := 0; i < g.price2Client.Len(); i++ {
		ok, it, c, _, ds := g.price2Client.GetItem(i)
		if logiclog.LogIsWingProp(it) {
			logiclog.LogHeroWingGetProp(p.AccountID.String(),
				p.Profile.GetCurrAvatar(),
				p.Profile.GetCorp().GetLvlInfo(),
				p.Profile.ChannelId, 0,
				func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
				[]string{it}, []uint32{c}, reason, "")
		}
		showToClient := false
		//遍历所有的COND_TYP_ExchangeShop称号，与判断条件3
		for _, cfg := range gamedata.GetTitleCond(typ) {
			if it != cfg.GetFCValueIP3() {
				continue
			}
			_, ok := pt.TitleCanActivate[cfg.GetTitleID()]
			_, ok2 := pt.TitleHadActivate[cfg.GetTitleID()]
			if !ok && !ok2 { //满足条件&&没有激活
				showToClient = true
				pt.TitleCanActivate[cfg.GetTitleID()] = struct{}{}
				pt.TitleForClient[cfg.GetTitleID()] = struct{}{}
				player_msg.Send(p.AccountID.String(), player_msg.PlayerMsgTitleCode,
					player_msg.DefaultMsg{})
				logs.Debug("PlayerTitle cond_type activate %s %d %s",
					p.AccountID.String(), typ, cfg.GetTitleID())
				break
			}
		}
		if it == "VI_VIPXP" {
			logs.Debug("%s add vimpoint %d", it, c)
			p.Profile.GetVip().AddRmbPoint(p, c, reason)
			sync.OnChangeVIP()
		}
		if ok && !gamedata.IsItemIdKnownBeforeGive(it) && !showToClient {
			res.AddItemWithData2Client(it, ds, c)
		}
	}

	if g.giveHcFromVip > 0 {
		cfg := p.Profile.GetMyVipCfg()
		if cfg == nil {
			logs.SentryLogicCritical(aid, "GiveBySync Get Vip Lv Err")
		} else {
			data := cfg.VIPDailyGift.GetData()
			// 注意要在最前面
			g.AddCostData(data)
			res.AddOther2Client(data)
		}
	}

	// 注意要在最前面
	iap_infos := make(map[string]GiveIAPData, 2)

	var Gpkgid int
	var Gsubpkgid int
	var Giap uint32
	if g.IAPData != nil {
		Gpkgid = g.IAPData.IAPPkgInfo.PkgId
		Giap = g.IAPData.IAPGoodIndex
		Gsubpkgid = g.IAPData.IAPPkgInfo.SubPkgId
		iap_infos[""] = GiveIAPData{
			IAPGoodIndex:   g.IAPData.IAPGoodIndex,
			IAPGameOrderId: g.IAPData.IAPGameOrderId,
			IAPOrder:       g.IAPData.IAPOrder,
			IAPPrice:       g.IAPData.IAPPrice,
			IAPPlatform:    g.IAPData.IAPPlatform,
			IAPChannel:     g.IAPData.IAPChannel,
			IAPPayTime:     g.IAPData.IAPPayTime,
			IAPTrueAmount:  g.IAPData.IAPTrueAmount,
			IAPPkgInfo:     g.IAPData.IAPPkgInfo,
			IAPPayType:     g.IAPData.IAPPayType,
		}
	}
	for _, item := range g.items {
		if ok, idx := gamedata.IsItemIAP(item); ok {
			gamedata.GetPlatformByIdx(idx)
			iap_infos[item] = GiveIAPData{
				IAPGoodIndex:   idx,
				IAPGameOrderId: "Debug",
				IAPPlatform:    gamedata.GetPlatformByIdx(idx),
				IAPPayTime:     fmt.Sprintf("%d", time.Now().Unix()),
				IAPOrder:       "Debug",
				IAPChannel:     "Debug",
				IAPPkgInfo:     gamedata.PackageInfo{0, 0},
			}
		}
	}

	for _, _iap := range iap_infos {
		hcBuy, hcGive, goodName := p.Profile.GetIAPGoodInfo().OnPayGoodSuccess(p.AccountID.ShardId,
			_iap.IAPGoodIndex, _iap.IAPPlatform, p.Profile.GetProfileNowTime(), _iap.IAPTrueAmount, _iap.IAPChannel)
		if Gpkgid < 0 || Gsubpkgid < 0 {
			Gpkgid = 0
			Gsubpkgid = 0
		}
		pkgValue := gamedata.GetHotDatas().HotKoreaPackge.GetHotPackage(int64(Gpkgid), int64(Gsubpkgid))
		if pkgValue == nil {
			logs.Error("can not find the package %d:%d", Gpkgid, Gsubpkgid)
			Gpkgid = 0
			Gsubpkgid = 0
		}
		if Gpkgid > 0 {
			isTrue := false
			IapId := pkgValue.GetIAPID()
			tc := strings.Split(IapId, ",")
			for i := 0; i < len(tc); i++ {
				if tc[i] == fmt.Sprintf("%d", _iap.IAPGoodIndex) {
					isTrue = true
					break

				}
			}
			if !isTrue {
				Gpkgid = 0
				Gsubpkgid = 0
			}
		}

		if Gpkgid != 0 && _iap.IAPChannel != "debug" {
			sync.OnChangeIAPGift()
			logs.Debug("korea packet iap step: find ipa, %d", uint32(_iap.IAPGoodIndex))
			if gamedata.GetHotDatas().HotKoreaPackge.GetHotPackageType(int64(Gpkgid)) == gamedata.ConditonPackage &&
				!p.Profile.Koreapackget.GetCondHaveBuy(int64(Gpkgid)) {
				/*
				 购买条件礼包的情况(不是领取礼包).物品通过res返回给客户端显示，p添加到账户
				*/
				p.Profile.Koreapackget.UpdateCondiPkgOne(int64(Gpkgid))
				logs.Debug("have got korea condition package %d:%d", Gpkgid, Gsubpkgid)
			} else {
				p.Profile.Koreapackget.UpdateLimitTimeOne(int64(Gpkgid), int64(Gsubpkgid), p.GetProfileNowTime())
				for _, value := range pkgValue.GetHotPackageGoods_Temp() {
					cost := gamedata.CostData{}
					cost.AddItemWithRes(value.GetGoodsID(), value.GetGoodsCount(), &res, p.GetRand())
					g.AddCostData(&cost)
					logs.Debug("korea package send item:%d count:%d", value.GetGoodsID(), value.GetGoodsCount())
				}
				hcValue := pkgValue.GetHCValue()
				if hcValue != 0 {
					logs.Debug("Package add hc_give %d", hcValue)
					g.addHc(gamedata.HC_From_Give, int64(hcValue))
					res.AddItem2Client(helper.VI_Hc, hcValue)
				}
			}
			hcGive = 0
			hcBuy = 0
		}

		if hcBuy > 0 {
			g.addHc(gamedata.HC_From_Buy, int64(hcBuy))
			g.addHc(gamedata.HC_From_Give, int64(hcGive))

			p.Profile.GetMarketActivitys().OnOnlyPay(p.AccountID.String(), _iap.IAPGoodIndex, p.Profile.GetProfileNowTime())

			p.Profile.GetMarketActivitys().OnHeroFundByPay(aid, int(_iap.IAPGoodIndex), p.GetSimpleInfo().CurrCorpGs, p.GetProfileNowTime())

			g.DoActiveRedPacket(p, int(_iap.IAPGoodIndex))

			sync.OnChangeIAPGoodInfo()
			if _iap.IAPGameOrderId != "" {
				if _iap.IAPChannel != "gm" {
					// hour log
					hour_log.Get(p.AccountID.ShardId).OnPay(p.AccountID.String(),
						p.Profile.ChannelId, int(_iap.IAPTrueAmount))
					// logiclog
					logiclog.LogIAP(aid, p.Profile.AccountName, p.Profile.Name,
						p.Profile.CurrAvatar, p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
						_iap.IAPGoodIndex, _iap.IAPGameOrderId, goodName,
						_iap.IAPOrder, _iap.IAPTrueAmount, _iap.IAPPlatform,
						_iap.IAPChannel, _iap.IAPPayTime, hcBuy, hcGive, p.GetIp(), p.Profile.GetVipLevel(),
						p.Profile.IAPGoodInfo.MoneySum, p.Profile.HC.GetHCFromBy(), p.Profile.HC.GetHCFromGive(),
						p.Profile.HC.GetHCFromCompensate(),
						func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

					if p.Profile.GetIAPGoodInfo().MoneySum <= 0 {
						logiclog.LogFirstPay(aid, p.Profile.CurrAvatar, p.Profile.GetCorp().GetLvlInfo(),
							p.Profile.ChannelId, _iap.IAPGoodIndex, p.Profile.GetData().FarthestStageIndex,
							p.Profile.GetData().FarthestEliteStageIndex, p.Profile.GetData().FarthestHellStageIndex,
							func(last string) string {
								return p.Profile.GetLastSetCurLogicLog(last)
							}, "")
					}

					// record iap for userinfo log
					p.Profile.GetIAPGoodInfo().AddRecOrder(_iap.IAPPayTime, _iap.IAPTrueAmount,
						_iap.IAPGoodIndex, _iap.IAPOrder, _iap.IAPPlatform, _iap.IAPPayType)
				}

				// write dynamodb
				if err := pay.LogIAP2DB(aid, p.Profile.AccountName, p.Profile.Name, p.Profile.CurrAvatar,
					_iap.IAPGoodIndex, goodName, _iap.IAPOrder, _iap.IAPTrueAmount, _iap.IAPPlatform,
					_iap.IAPChannel, _iap.IAPPayTime, hcBuy, hcGive); err != nil {
					logs.SentryLogicCritical(aid, "GiveBySync LogIAP2DB err: %v", err)
				}

			}
		} else {
			logs.Debug("IAPGoodIndex %d hcBuy <= 0", _iap.IAPGoodIndex)
		}
	}

	p.Tmp.GoldLevelPoint += g.goldLevelPoint
	p.Tmp.ExpLevelPoint += g.expLevelPoint
	p.Tmp.DCLevelPoint += g.dcLevelPoint
	flag := 0
	logs.Debug("before add vipPoint")
	for hc_t, hc_v := range g.hc {
		logs.Debug("package %d:%d addhc type:%d number:%d", Gpkgid, Gsubpkgid, hc_t, hc_v)
		if hc_v > 0 {
			if hc_t == HC_From_Buy {
				flag = 1
				logs.Debug("Package %d:%d have HcFromBuy", Gpkgid, Gsubpkgid)
			}
			p.Profile.GetHC().AddHC(aid, hc_t, hc_v, p.Profile.GetProfileNowTime(), reason)

			if sync != nil {
				sync.OnChangeHC()
			}
		}
		if hc_t == HC_From_Buy {
			logs.Debug("Hc from buy add hc VIPPoint")
			p.Profile.GetVip().AddRmbPoint(p, uint32(hc_v), reason)
			count := p.Profile.GetWheelGachaInfo().UpdataCoinByHc(hc_v)
			if count != 0 {
				res.AddItem2Client(helper.VI_WheelCoin, uint32(count))
			}
			if sync != nil {
				sync.OnChangeVIP()
				sync.OnChangeWheel()
			}
			// 充值获得砸蛋的锤子
			hmr_c := p.Profile.GetHitEgg().OnAddHcBug(p, hc_v, p.Profile.GetProfileNowTime())
			if hmr_c > 0 {
				g.sc[helper.SC_EggKey] = g.sc[helper.SC_EggKey] + hmr_c
				sync.OnChangeGameMode(counter.CounterTypeHitHammerDailyLimit)
				sync.OnChangeHitEgg()
			}
		}
	}
	/*
		如果没有钻石，走iap表里的流程
	*/
	if Gpkgid != 0 && flag == 0 {
		logs.Debug("Package %d:%d add VIPPoint %d", Gpkgid, Gsubpkgid,
			gamedata.GetIAPInfo(Giap).Info.GetVIPPoint())
		p.Profile.GetVip().AddRmbPoint(p, gamedata.GetIAPInfo(Giap).Info.GetVIPPoint(), reason)
		if sync != nil {
			sync.OnChangeVIP()
		}
	}

	var sc_bonus, gc_bonus, exp_bonus, gb_bonus float32
	if p.GuildProfile.InGuild() && g.GuildScienceValid != gamedata.GST_NULL {
		bonus := guild.GetModule(p.AccountID.ShardId).GetGuildScienceBonus(
			p.GuildProfile.GuildUUID, p.AccountID.String(), g.GuildScienceValid)
		switch g.GuildScienceValid {
		case gamedata.GST_BossFight:
			gb_bonus = bonus[0]
		case gamedata.GST_GoldBonus:
			sc_bonus = bonus[0]
		case gamedata.GST_GateEnemy:
			gc_bonus = bonus[0]
		case gamedata.GST_DailyTask:
			exp_bonus = bonus[0]
		}
		logs.Debug("give guildscience bonus %d sc %f gc %f exp %f",
			g.GuildScienceValid, sc_bonus, gc_bonus, exp_bonus)
	}

	// 加sc必须在加hc之后进行
	for sc_t, sc_v := range g.sc {
		if sc_v > 0 {
			sc_old := sc_v
			if sc_t == helper.SC_Money && sc_bonus > 0 {
				old := sc_v
				sc_v = int64(float32(sc_v) * (1 + sc_bonus))
				logs.Debug("give sc %d %f %d", old, sc_bonus, sc_v)
			}
			if sc_t == helper.SC_GuildCoin && gc_bonus > 0 {
				old := sc_v
				sc_v = int64(float32(sc_v) * (1 + gc_bonus))
				logs.Debug("give GuildCoin %d %f %d", old, gc_bonus, sc_v)
			}
			if sc_t == helper.SC_GB && gb_bonus > 0 {
				old := sc_v
				sc_v = int64(float32(sc_v) * (1 + gb_bonus))
				logs.Debug("give gb %d %f %d", old, gb_bonus, sc_v)
			}
			if sc_t == helper.SC_GuildSp {
				if p.GuildProfile.GuildUUID != "" {
					guild.GetModule(p.AccountID.ShardId).AddSp(p.GuildProfile.GuildUUID,
						aid, sc_v)
				} else {
					guild.GetModule(p.AccountID.ShardId).UpdateAccountInfo(p.GetSimpleInfo())
				}
			}

			if g.GateEnemyBonus > 0 {
				sc_v = int64(float32(sc_v)*g.GateEnemyBonus + 0.5)
				logs.Debug("give sc gateenemy bonus %f", g.GateEnemyBonus)
			}
			p.Profile.GetSC().AddSC(sc_t, sc_v, reason)
			if sync != nil {
				sync.OnChangeSC()
			}
			if sc_v-sc_old > 0 {
				res.AddItem2Client(helper.SCString(sc_t), uint32(sc_v-sc_old))
			}
		}
	}

	//idx_has_give := 0
	for i := 0; i < len(g.items); i++ {
		id := g.items[i]
		c := g.count[i]
		if _, ok := iap_infos[id]; ok { // iap物品不处理
			continue
		}
		if c > 0 {
			his := p.Profile.GetItemHistory()
			if id == helper.MatEevoUniversalItemID {
				//自适应掉落
				for j := 0; j < int(c); j++ {
					addItemID, idx := his.GetRandLootByHistory(p.GetRand())
					logs.Trace("MatEevoUniversalItemID,%s", addItemID)
					if addItemID == "" {
						logs.Error("MatEevoUniversalItemID Rand Err By Idx %d",
							idx)
						continue
					}
					his.Add(idx, 1)
					errCode, item_inner_type, idx2OldCount := p.BagProfile.AddToBag(p, g.item_datas[i], addItemID, 1)
					if errCode != helper.RES_AddToBag_Success {
						if errCode == helper.RES_AddToBag_MaxCount {
							logs.SentryLogicCritical(aid, "[ItemMaxCount]Give item AddToBag Error for MaxCount:%d,%s,%d,%s",
								errCode, addItemID, 1, reason)
						}
						logs.SentryLogicCritical(aid, "Give item AddToBag Error:%d,%s,%d,%s", errCode, addItemID, 1, reason)
						return false, nil
					}
					if idx2OldCount != nil && sync != nil {
						for bagId, oldCount := range idx2OldCount {
							sync.OnChangeUpdateItems(item_inner_type, bagId, oldCount, reason)
						}
					}
					res.AddItem2Client(addItemID, 1)
				}
			} else {
				oldc := c
				if g.GateEnemyBonus > 0 {
					b, _ := gamedata.IsJade(id)
					if b {
						c = uint32(float32(c)*g.GateEnemyBonus + 0.5)
						logs.Debug("give jade gateenemy bonus %f", g.GateEnemyBonus)
					}
				}
				errCode, item_inner_type, idx2OldCount := p.BagProfile.AddToBag(p, g.item_datas[i], id, c)
				if errCode != helper.RES_AddToBag_Success {
					if errCode == helper.RES_AddToBag_MaxCount {
						logs.SentryLogicCritical(aid, "[ItemMaxCount]Give item AddToBag Error for MaxCount:%d,%s,%d,%s", errCode, id, c, reason)
					}
					logs.SentryLogicCritical(aid, "Give item AddToBag Error:%d,%s,%d,%s", errCode, id, c, reason)
					return false, nil
				}
				und := gamedata.GetUniversalMaterialDataByItem(id)
				if und != nil {
					his.Add(und.ID, c)
				}
				if idx2OldCount != nil && sync != nil {
					for bagId, oldCount := range idx2OldCount {
						sync.OnChangeUpdateItems(item_inner_type, bagId, oldCount, reason)
					}
				}
				if c-oldc > 0 {
					res.AddItem2Client(id, c-oldc)
				}
				if item_inner_type == helper.Item_Inner_Type_Fashion {
					p.Profile.GetData().SetNeedCheckMaxGS()
				}
			}
		}
	}

	for i := 0; i < len(g.general); i++ {
		id := g.general[i]
		c := g.gcount[i]
		if c > 0 {
			p.GeneralProfile.AddGeneralNum(id, c, reason)
			if sync != nil {
				sync.OnChangeGeneralAllChange()
			}
			general := p.GeneralProfile.GetGeneral(id)
			if general != nil {
				logiclog.LogGeneralAddNum(aid, p.Profile.GetCurrAvatar(),
					p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, general.Id, c, reason,
					func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
			}
		}
	}

	if g.avatar_xp_all > 0 {
		p.Profile.GetAvatarExp().AddExp2All(p, g.avatar_xp_all, reason)
		if sync != nil {
			sync.OnChangeAvatarExp()
		}
	}

	if len(g.avaters_xp) > 0 {
		for avater, xp := range g.avaters_xp {
			p.Profile.GetHero().AddHeroExp(p, avater, xp)
		}
		if sync != nil {
			sync.OnChangeAvatarExp()
		}
	}

	if g.corp_xp > 0 {
		corp_xp_old := g.corp_xp
		corp_lvl_old := p.GetCorpLv()
		if exp_bonus > 0 {
			old := g.corp_xp
			g.corp_xp = uint32(float32(g.corp_xp) * (1 + exp_bonus))
			logs.Debug("give corp_xp %d %f %d", old, exp_bonus, g.corp_xp)
		}
		p.Profile.GetCorp().AddExp(aid, g.corp_xp, reason)
		if sync != nil {
			sync.OnChangeCorpExp()
		}
		if g.corp_xp-corp_xp_old > 0 {
			res.AddItem2Client(gamedata.VI_CorpXP, g.corp_xp-corp_xp_old)
		}
		if p.GetCorpLv() > corp_lvl_old {
			p.Profile.AutomationQuest(p)
		}
	}

	if g.EStageId != "" {
		if p.Profile.GetStage().AddEStageTimes(gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()),
			g.EStageId, g.EStageTimes) && sync != nil {
			sync.OnChangeStage(g.EStageId)
		}
	}

	if g.HeroTalentPoint > 0 {
		p.Profile.GetHeroTalent().BuyTalentPoint()
		sync.OnChangeHeroTalent()
	}

	if g.GameModeId > 0 {
		if p.Profile.GetCounts().Add(
			g.GameModeId,
			g.GameModeTime,
			p) {
			sync.OnChangeGameMode(uint32(g.GameModeId))
		} else {
			logs.Error("GiveBySync CounterTypeTeamPvp addtimes failed")
		}
	}

	hero := p.Profile.GetHero()
	for avatarID, v := range g.heroPiece {
		if v > 0 {
			old := v
			if g.GateEnemyBonus > 0 {
				v = uint32(float32(v)*g.GateEnemyBonus + 0.5)
				logs.Debug("give heroPiece gateenemy bonus %f", g.GateEnemyBonus)
			}
			hero.Add(p, avatarID, v, reason)
			if v-old > 0 {
				res.AddItem2Client(gamedata.GetHeroData(avatarID).Piece, v-old)
			}
		}
	}

	for avatarID, v := range g.heroPieceWholeChar {
		if v <= 0 {
			continue
		}

		for i := 0; i < int(v); i++ {
			if hero.GetStar(avatarID) > 0 || hero.IsWholeCharHasGot(avatarID) {
				hero.Add(p, avatarID, g.heroPieceWholeCharCount[avatarID], "WholeChar")
				res.AddItemWithData2Client(
					gamedata.GetWholeCharIdByAvatarId(avatarID), gamedata.BagItemData{
						WholeCharPieces: g.heroPieceWholeCharCount[avatarID],
						IsWholeChar:     0,
					}, 1)
			} else {
				hData := gamedata.GetHeroData(avatarID)
				hero.AddInit(p, avatarID,
					hData.UnlockPieceNeed,
					"WholeChar")
				hero.SetWholeCharHasGot(avatarID)
				res.AddItemWithData2Client(
					gamedata.GetWholeCharIdByAvatarId(avatarID), gamedata.BagItemData{
						WholeCharPieces: hData.UnlockPieceNeed,
						IsWholeChar:     1,
					}, 1)
			}
		}
	}

	if nil != g.starsouls {
		astrologyBag := p.Profile.GetAstrology().GetBag()
		for id, c := range g.starsouls {
			astrologyBag.AddSoul(id, c)
		}
	}

	if p.GuildProfile.InGuild() {
		if g.GuildXp > 0 {
			old := g.GuildXp
			if g.GateEnemyBonus > 0 {
				g.GuildXp = uint32(float32(g.GuildXp)*g.GateEnemyBonus + 0.5)
				logs.Debug("give GuildXp gateenemy bonus %f", g.GateEnemyBonus)
			}
			guild.GetModule(p.AccountID.ShardId).AddXp(p.GuildProfile.GuildUUID,
				aid, int64(g.GuildXp))
			if g.GuildXp-old > 0 {
				res.AddItem2Client(gamedata.VI_GuildXP, g.GuildXp-old)
			}
		}
	}

	ok := true
	// 这个在最后, 可能失败
	if g.energy > 0 {
		ok = p.Profile.GetEnergy().AddForce(aid, reason, int64(g.energy))
		if sync != nil {
			sync.OnChangeEnergy()
		}
	}

	//发放幸运转盘抽奖券
	if g.wheelcoin > 0 {
		ok = p.Profile.GetWheelGachaInfo().UpdataCoin(g.wheelcoin)
		if sync != nil {
			sync.OnChangeWheel()
		}
	}

	if g.bossFightPoint > 0 {
		ok = p.Profile.GetBossFightPoint().AddForce(aid, reason, int64(g.bossFightPoint))
		if sync != nil {
			sync.OnChangeBossFightPoint()
		}
	}
	return ok, &res
}

func (g *GiveGroup) GiveBySyncAuto(p *Account, sync interfaces.ISyncRspWithRewards, reason string) bool {
	ok, res := g.GiveBySyncWithRes(p, sync, reason)
	if sync != nil {
		sync.AddResReward(res)
		sync.MergeReward()
	}
	return ok
}

func (g *GiveGroup) GiveBySyncAutoWithoutMerge(p *Account, sync interfaces.ISyncRspWithRewards, reason string) bool {
	ok, res := g.GiveBySyncWithRes(p, sync, reason)
	if sync != nil {
		sync.AddResReward(res)
	}
	return ok
}

func (g *GiveGroup) IsCanAddItem(p *Account) bool {
	item2Count := make(map[string]uint32, len(g.items))
	for i := 0; i < len(g.items); i++ {
		id := g.items[i]
		c := g.count[i]
		item2Count[id] = item2Count[id] + c
	}
	for item, count := range item2Count {
		if !p.BagProfile.IsCanAddItem(item, count) {
			return false
		}
	}
	return true
}

func (g *GiveGroup) itemHasIap() {

}

func (g *GiveGroup) isPackageGroup(itemId string, packageGroup []string) bool {
	for _, id := range packageGroup {
		if id == itemId {
			return true
		}
	}
	return false
}
func (g *GiveGroup) DoActiveRedPacket(p *Account, iapGoodIndex int) {

	if _, ok := gamedata.GetHotDatas().RedPacketConfig.IapSet[uint32(iapGoodIndex)]; !ok {
		iapGoodIndex -= uutil.IAPID_ONESTORE_2_GOOGLE
		if _, ok := gamedata.GetHotDatas().RedPacketConfig.IapSet[uint32(iapGoodIndex)]; !ok {
			return
		}
	}
	g.doActiveRedPacket(p, iapGoodIndex)
}

func (g *GiveGroup) doActiveRedPacket(p *Account, iapGoodIndex int) {
	p.GuildProfile.RedPacketInfo.CheckDailyReset(p.GetProfileNowTime())
	if p.GuildProfile.RedPacketInfo.IpaStatus != guild2.RP_IPA_NONE {
		logs.Debug("can not active red packet, status = %d", p.GuildProfile.RedPacketInfo.IpaStatus)
		return
	}
	ok := p.Profile.GetMarketActivitys().OnRedPacketByPay(p.AccountID.String(), int(iapGoodIndex), p.GetProfileNowTime(),
		p.AccountID.ShardId, p.GuildProfile.GuildUUID, p.Profile.Name)
	if ok {
		p.GuildProfile.RedPacketInfo.IpaStatus = guild2.RP_IPA_HAS_PAY
		p.GuildProfile.RedPacketInfo.Sync.SetNeedSync()
	}
}
