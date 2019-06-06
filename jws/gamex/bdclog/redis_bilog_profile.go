package bdclog

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"sort"

	"strings"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/currency"
	"vcs.taiyouxi.net/jws/gamex/models/fashion"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/pay"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

func (cs *BiScanRedis) bi_profile(key string, skey []string, val []byte, rh storehelper.ReadHandler) error {
	t_now := time.Now().In(cs.loc)
	t_now = t_now.Add(-1 * time.Hour) // 如果用当前时间跑过去一天的数据，则当前时间减去1小时
	if cs.timestamp > 0 {
		t_now = time.Unix(cs.timestamp, 0).In(cs.loc)
	}

	acid := skey[1]
	a, err := db.ParseAccount(acid)
	if err != nil {
		return err
	}
	has := false
	for _, sid := range cs.shardId {
		if fmt.Sprintf("%d", a.ShardId) == sid {
			has = true
			break
		}
	}
	if !has {
		return nil
	}

	accountNameId := fmt.Sprintf("%d:%s", a.GameId, a.UserId)
	strChan := cs.gidInfo[int(cs.gid)]
	sid := fmt.Sprintf("%s%04s%06s", strChan, gameId, cs.shardId[0])

	var dat map[string]string
	if err := json.Unmarshal(val, &dat); err != nil {
		return fmt.Errorf("BiScanRedis unmarshal val %v %s", err, key)
	}
	deviceId := dat["device_id"]
	accountName := dat["account_name"]
	// iap
	payInfo := &pay.PayGoodInfos{}
	if _, ok := dat["iap_good_info"]; ok {
		if err := json.Unmarshal([]byte(dat["iap_good_info"]), payInfo); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal PayGoodInfos %v %s %s", err, key, dat["iap_good_info"])
		}
	}
	// hc
	hc := &currency.HardCurrency{}
	if _, ok := dat["hc"]; ok {
		if err := json.Unmarshal([]byte(dat["hc"]), hc); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal HardCurrency %v %s %s", err, key, dat["hc"])
		}
	}
	// level
	corp := &account.Corp{}
	if _, ok := dat["corp"]; ok {
		if err := json.Unmarshal([]byte(dat["corp"]), corp); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal Corp %v %s %s", err, key, dat["corp"])
		}
	}
	lvl := corp.Level
	// name
	name := dat["name"]
	// pay log
	for _, ord := range payInfo.RecentOrders {
		if strings.ToLower(ord.Order) == "debug" {
			continue
		}
		t_o := time.Unix(ord.TimeStamp, 0).In(cs.loc)
		if t_now.Day() == t_o.Day() && t_now.Month() == t_o.Month() &&
			t_now.Year() == t_o.Year() {
			payStr := fmt.Sprintf("%s$$%s$$%s$$%s$$%s$$%s$$%d$$%d$$%d$$%d$$%s$$%s$$%s$$%s",
				deviceId, accountNameId, accountName, dat["channel"], acid, name,
				lvl, hc.GetHC(), ord.Money, ord.Idx, ord.Order,
				time.Unix(ord.TimeStamp, 0).In(cs.loc).Format(timeLayout), sid, ord.PayType)
			if _, err := cs.paywriter.WriteString(payStr + "\n"); err != nil {
				return err
			}
		}
	}

	// sc
	sc := &currency.SoftCurrency{}
	if _, ok := dat["sc"]; ok {
		if err := json.Unmarshal([]byte(dat["sc"]), sc); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal SoftCurrency %v %s %s", err, key, dat["sc"])
		}
	}
	// exp
	exp := &account.AvatarExp{}
	if _, ok := dat["avatarExp"]; ok {
		if err := json.Unmarshal([]byte(dat["avatarExp"]), exp); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal avatarExp %v %s %s", err, key, dat["avatarExp"])
		}
	}
	arousalLvs := make([]uint32, 0, account.AVATAR_NUM_CURR)
	for i := 0; i < account.AVATAR_NUM_CURR; i++ {
		e := exp.Avatars[i]
		arousalLvs = append(arousalLvs, e.ArousalLv)
	}

	// vip
	v := &account.VIP{}
	if _, ok := dat["v"]; ok {
		if err := json.Unmarshal([]byte(dat["v"]), v); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal VIP %v %s %s", err, key, dat["v"])
		}
	}
	vip, _ := v.GetVIP()

	// energy
	eg := &account.PlayerEnergy{}
	if _, ok := dat["energy"]; ok {
		if err := json.Unmarshal([]byte(dat["energy"]), eg); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal PlayerEnergy %v %s %s", err, key, dat["energy"])
		}
	}
	energy := eg.Value

	// create time
	createTimeO := dat["createtime"]
	icreatetime, _ := strconv.Atoi(createTimeO)
	createTime := time.Unix(int64(icreatetime), 0).In(cs.loc).Format(timeLayout)
	// data
	data := &account.ProfileData{}
	if _, ok := dat["data"]; ok {
		if err := json.Unmarshal([]byte(dat["data"]), data); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal ProfileData %v %s %s", err, key, dat["data"])
		}
	}
	lastLoginTimeO, _ := strconv.ParseInt(dat["logintime"], 10, 64)
	lastLoginTime := time.Unix(lastLoginTimeO, 0).In(cs.loc).Format(timeLayout)
	loginDayNum, _ := strconv.Atoi(dat["logindaynum"])
	// phone
	var str_phone string
	ph := &account.PlayerPhoneData{}
	if _, ok := dat["phones"]; ok {
		if err := json.Unmarshal([]byte(dat["phones"]), ph); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal phone %v %s %s", err, key, dat["phones"])
		}
		if ph.IsHasBindPhone() {
			str_phone = ph.Phone
		}
	}
	// trial
	trial := &account.PlayerTrial{}
	if _, ok := dat["trial"]; ok {
		if err := json.Unmarshal([]byte(dat["trial"]), trial); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal trial %v %s %s", err, key, dat["trial"])
		}
	}
	// equip
	equips := &account.Equips{}
	if _, ok := dat["equips"]; ok {
		if err := json.Unmarshal([]byte(dat["equips"]), equips); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal Equip %v %s %s", err, key, dat["equips"])
		}
	}
	avatar_equips := &account.AvatarEquips{}
	if _, ok := dat["avatar_equips"]; ok {
		if err := json.Unmarshal([]byte(dat["avatar_equips"]), avatar_equips); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal avatarEquip %v %s %s", err, key, dat["avatar_equips"])
		}
	}
	bagKey := "bag:" + acid
	bagDatb, b := rh(bagKey)
	if !b {
		return fmt.Errorf("BiScanRedis get bag data fail %s", bagKey)
	}
	var bagMap map[string]string
	if err := json.Unmarshal(bagDatb, &bagMap); err != nil {
		return fmt.Errorf("BiScanRedis unmarshal bagdata %v %s", err, bagKey)
	}
	// jade bag
	jade_bag := &account.PlayerJadeBagDB{}
	jade_bag_map := make(map[uint32]account.JadeItem, 5)
	if _, ok := dat["jade_bag"]; ok {
		if err := json.Unmarshal([]byte(dat["jade_bag"]), jade_bag); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal PlayerJadeBagDB %v %s %s", err, key, dat["jade_bag"])
		}
		for _, j := range jade_bag.Jades {
			jade_bag_map[j.ID] = j
		}
	}
	// jade
	eq_jades := &account.EquipmentJades{}
	if _, ok := dat["av_jades"]; ok {
		if err := json.Unmarshal([]byte(dat["av_jades"]), eq_jades); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal AvatarJades %v %s %s", err, key, dat["av_jades"])
		}
	}

	eqRes := make([]Equip, 0, helper.EQUIP_SLOT_MAX)
	for slot := 0; slot < helper.EQUIP_SLOT_MAX; slot++ {
		equip_id := equips.GetEquip(slot)
		if equip_id <= 0 {
			continue
		}
		item, ok := bagMap[fmt.Sprintf("%d", equip_id)]
		if !ok {
			logs.Error("equip but not found in bag %v %v %v", bagKey, fmt.Sprintf("%d", equip_id), bagMap)
			continue
		}
		var bi bag.BagItem
		err := json.Unmarshal([]byte(item), &bi)
		if err != nil && bi.ID == equip_id {
			logs.Error("equip item Unmarshal err %v %v", bagKey, equip_id)
			continue
		}
		ejs := eq_jades.GetSlotJadeForLog(slot)
		jades_info := make([]string, 0, 16)
		for _, av_jade_id := range ejs {
			if av_jade_id <= 0 {
				continue
			}
			item := jade_bag_map[av_jade_id]
			jades_info = append(jades_info, item.TableID)
		}
		eqRes = append(eqRes, Equip{
			Slot:       gamedata.Slot2String(slot),
			Item:       bi.TableID,
			Lv_upgrade: equips.GetEvolution(slot),
			Lv_star:    equips.GetStarLv(slot),
			Lv_matenh:  equips.GetMatEnhLv(slot),
			TrickGroup: bi.ItemData.TrickGroup,
			Jades:      jades_info,
		})
	}
	// fashion bag
	fashionBag := &fashion.PlayerFashionBagDB{
		Items: []helper.FashionItem{},
	}
	if _, ok := dat["fashion_bag"]; ok {
		if err := json.Unmarshal([]byte(dat["fashion_bag"]), fashionBag); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal fashion_bag %v %s %s", err, key, dat["fashion_bag"])
		}
	}

	fashionMap := make(map[uint32]helper.FashionItem, len(fashionBag.Items))
	for _, f := range fashionBag.Items {
		fashionMap[f.ID] = f
	}
	aequipRes := make(map[int]AvatarEquips, helper.AVATAR_NUM_CURR)
	for i := 0; i < helper.AVATAR_NUM_CURR; i++ {
		aequipRes[i] = make(AvatarEquips, 0, helper.AVATAR_EQUIP_SLOT_MAX)
		for slot := 0; slot < helper.AVATAR_EQUIP_SLOT_MAX; slot++ {
			equip_id := avatar_equips.GetEquip(i, slot)
			if equip_id <= 0 {
				continue
			}
			var t string
			item, ok := fashionMap[equip_id]
			if !ok {
				item, ok := bagMap[fmt.Sprintf("%d", equip_id)]
				if !ok {
					logs.Error("AvatarEquips but not found in bag %v %v %v", bagKey, fmt.Sprintf("%d", equip_id), bagMap)
					continue
				}
				var bi bag.BagItem
				err := json.Unmarshal([]byte(item), &bi)
				if err != nil && bi.ID == equip_id {
					logs.Error("AvatarEquips item Unmarshal err %v %v", bagKey, equip_id)
					continue
				}
				t = bi.TableID
			} else {
				t = item.TableID
			}
			aequipRes[i] = append(aequipRes[i], AvatarEquip{
				Slot: gamedata.Slot2String(slot),
				Item: t,
			})
		}
	}

	dg_jades_info := make([]string, 0, 16)
	dg_jades := &account.DestGeneralJades{}
	if _, ok := dat["dg_jades"]; ok {
		if err := json.Unmarshal([]byte(dat["dg_jades"]), dg_jades); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal DestGeneralJades %v %s %s", err, key, dat["av_jades"])
		}
	}
	for _, dg_jade_id := range dg_jades.DestinyGeneralJade {
		if dg_jade_id <= 0 {
			continue
		}
		item := jade_bag_map[dg_jade_id]
		dg_jades_info = append(dg_jades_info, item.TableID)
	}
	// destinys
	dgs := &account.PlayerDestinyGeneral{}
	if _, ok := dat["destinys"]; ok {
		if err := json.Unmarshal([]byte(dat["destinys"]), dgs); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal PlayerDestinyGeneral %v %s %s", err, key, dat["destinys"])
		}
	}
	dg_info := make(DestinyGenerals, 0, len(dgs.Generals))
	for _, dg := range dgs.Generals {
		dg_info = append(dg_info, DestinyGeneral{
			Id:    dg.Id,
			Level: dg.LevelIndex,
		})
	}

	// title
	title := &account.PlayerTitleInDB{}
	if _, ok := dat["title"]; ok {
		if err := json.Unmarshal([]byte(dat["title"]), title); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal PlayerTitleInDB %v %s %s", err, key, dat["title"])
		}
	}
	title_info := make([]string, 0, len(title.TitleHadActivate)+2)
	title_info = append(title_info, title.TitleHadActivate...)
	if title.TitleSimplePvp != "" {
		title_info = append(title_info, title.TitleSimplePvp)
	}
	if title.TitleTeamPvp != "" {
		title_info = append(title_info, title.TitleTeamPvp)
	}
	// PlayerHeroSoul
	hsoul := &account.PlayerHeroSoul{}
	if v, ok := dat["herosoul"]; ok && v != "" {
		if err := json.Unmarshal([]byte(dat["herosoul"]), hsoul); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal PlayerHeroSoul %v %s %s", err, key, v)
		}
	}

	// hero
	heros := &account.PlayerHero{}
	if dat["hero"] != "" {
		if err := json.Unmarshal([]byte(dat["hero"]), heros); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal hero %v %s %s", err, key, dat["hero"])
		}
	}
	heroTalent := &account.PlayerHeroTalent{}
	if _, ok := dat["herotalent"]; ok {
		if err := json.Unmarshal([]byte(dat["herotalent"]), heros); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal PlayerHeroTalent %v %s %s", err, key, dat["herotalent"])
		}
	}
	// heroDestiny
	heroDestiny := &account.HeroDestiny{}
	if dat["hero_destiny"] != "" {
		if err := json.Unmarshal([]byte(dat["hero_destiny"]), heroDestiny); err != nil {
			return fmt.Errorf("BiScanRedis unmarshal hero_destiny %v %s %s", err, key, dat["hero_destiny"])
		}
	}
	logs.Debug("hero_Destiny %v", dat["hero_destiny"])
	heroDestinyId := make([]int, 0)
	for _, x := range heroDestiny.ActivateDestiny {
		heroDestinyId = append(heroDestinyId, x)
	}

	//主将时装
	heroFashion := &fashion.PlayerFashionBag{
		Items: fashionMap,
	}

	hasHeros := make(HeroInfos, 0, len(heros.HeroStarLevel))
	for i, h := range heros.HeroStarLevel {
		if h > 0 {
			hasHeros = append(hasHeros, HeroInfo{
				Id:           i,
				Lvl:          heros.HeroLevel[i],
				StarLvl:      h,
				Talent:       heroTalent.HeroTalentLevel[i][:helper.CurTalentCount],
				Pskill:       heros.HeroSkills[i].PassiveSkill,
				Cskill:       heros.HeroSkills[i].CounterSkill,
				Tskill:       heros.HeroSkills[i].TriggerSkill,
				SwingStarLvl: heros.HeroSwings[i].StarLv,
				SwingLvl:     heros.HeroSwings[i].Lv,
			})
		}
	}

	// custom info
	// gs
	sumgs := data.CorpCurrGS
	bestHero := make(BestHeroInfos, 0, 3)
	best9Heros := make(BestHeroInfos, 0, 9)

	for _, h := range data.BestHeroAvatar {
		bestHero = append(bestHero, BestHeroInfo{
			Id: h,
			Gs: data.HeroGs[h],
		})
	}
	_l := make(bestHeroList, 0, len(data.HeroBaseGs))
	for id, gs := range data.HeroBaseGs {
		_l = append(_l, BestHeroInfo{
			Id: id,
			Gs: gs,
		})
	}
	sort.Sort(_l)
	for i := 0; i < 9; i++ {
		best9Heros = append(best9Heros, BestHeroInfo{
			Id: _l[i].Id,
			Gs: _l[i].Gs,
		})
	}

	// guild
	guild_key := "pguild:" + acid
	guildDatb, b := rh(guild_key)
	if !b {
		return fmt.Errorf("BiScanRedis get guild data fail %s", guild_key)
	}
	var guildMap map[string]string
	if err := json.Unmarshal(guildDatb, &guildMap); err != nil {
		return fmt.Errorf("BiScanRedis unmarshal guilddata %v %s", err, guild_key)
	}
	guild := guildMap["guid"]
	ActivityAddonTodayCurrGuild, _ := strconv.ParseInt(guildMap["actaddoncurr"], 10, 64)
	ActivityAddonUUID := guildMap["actaddonguid"]
	ActivityAddonTime, _ := strconv.ParseInt(guildMap["actt"], 10, 64)
	guildAct := getAct(t_now.Unix(), guild, ActivityAddonTodayCurrGuild, ActivityAddonUUID, ActivityAddonTime)

	customInfo := CustomInfo{
		Phone: str_phone,
		//		RegTime:            createTime,
		//		LastLoginTime:      lastLoginTime,
		LoginDay: loginDayNum,
		//		Vip:                vip,
		//		CorpLvl:            lvl,
		SumGs:              sumgs,
		ArousalLvs:         arousalLvs,
		FarthestStage:      data.FarthestStageIndex,
		FarthestEliteStage: data.FarthestEliteStageIndex,
		FarthestHellStage:  data.FarthestHellStageIndex,
		//		Hc:                 hc.GetHC(),
		//		Money:              sc.GetSC(helper.SC_Money),
		FineIron:     sc.GetSC(helper.SC_FineIron),
		BossCoin:     sc.GetSC(helper.SC_BossCoin),
		PvpCoin:      sc.GetSC(helper.SC_PvpCoin),
		DestinyCoin:  sc.GetSC(helper.SC_DestinyCoin),
		EquipCoin:    sc.GetSC(helper.SC_EquipCoin),
		GuildCoin:    sc.GetSC(helper.SC_GuildCoin),
		CorpEquips:   eqRes,
		Role1Equips:  aequipRes[0],
		Role2Equips:  aequipRes[1],
		Role3Equips:  aequipRes[2],
		DGJades:      dg_jades_info,
		DestGen:      dg_info,
		GuildId:      guild,
		GuildAct:     guildAct,
		TrialMostLvl: trial.MostLevelId,
		Heros:        hasHeros,
		HeroFashion:  heroFashion.GetFashion2String(),
		VISB:         sc.GetSC(helper.SC_StarBlessCoin),
	}

	//now_t := time.Now().Unix()
	//cumStr 改动内容格式，老样式保留
	//cumStr, err := json.Marshal(customInfo)
	//if err != nil {
	//	return fmt.Errorf("BiScanRedis marshal CustomInfo %v", err)
	//}
	cumStr := mergeHeroInfo(heros, eqRes, dg_jades_info)
	accountInfo := fmt.Sprintf("%s$$%s$$%s$$%d$$%s$$%s$$%d$$%d$$%d$$%d$$%s$$%s$$%s$$%s$$%s",
		deviceId, accountNameId, accountName,
		hc.GetHC(), acid, name, lvl, vip, energy, sc.GetSC(helper.SC_Money),
		createTime, lastLoginTime, dat["channel"], cumStr, sid)
	//if now_t-lastLoginTimeO < util.DaySec*2 {
	if _, err := cs.writer.WriteString(accountInfo + "\n"); err != nil {
		return err
	}
	//}
	newhand_str := ""
	if dat["newhand_b"] != "" && len(dat["newhand_b"]) > 0 {
		bb, err := base64.StdEncoding.DecodeString(dat["newhand_b"])
		if err != nil {
			logs.Error("new hand decode b64 err %v", err)
			return err
		}
		var _b bytes.Buffer
		if _, err := _b.Write(bb); err != nil {
			logs.Error("bytes.buffer write err %v", err)
			return err
		}

		r, err := gzip.NewReader(&_b)
		if err != nil {
			logs.Error("gzip newreader %v", err)
			return err
		}
		defer r.Close()
		unzipbb, _ := ioutil.ReadAll(r)
		newhand_str = string(unzipbb)
	}

	csi := customInfo
	csvStr := fmt.Sprintf("%s,%s,%s,%s,"+ // accountId,name,渠道,手机号,
		"%s,%s,"+ // 注册时间,最后登陆日
		"%d,%d,"+ // 累计登陆日数,VIP等级
		"%s,"+ // 英雄
		"%v,"+ // 时装
		"%d,%d,%s,"+ // 战队等级,战队总战力,最强主将
		"%d,%d,%d,"+ // 最远关卡进度,最远精英关卡进度
		"%d,%d,%d,%d,%d,%d,%d,%d,"+ // 硬通数,软通数,精铁数,Boss代币数,PvP代币数,神将代币,装备代币,公会代币
		"%s,"+ // 神将宝石
		"%s,"+ // 神将
		"%s,"+ // 称号
		"%d,"+ // 武魂
		"%v,"+ // 战队装备
		"%s,%d,"+ // "公会,功勋值"
		"%d,"+ // "爬塔最高层,"
		"%d,"+ // "累计充值,"
		"%v,"+ // "羁绊,"
		"%d,"+ // "VI_SB"
		"%s,%s\r\n", // 新手引导, 引导Event
		acid, name, dat["channel"], csi.Phone,
		createTime, lastLoginTime,
		csi.LoginDay, vip,
		csi.Heros,
		csi.HeroFashion,
		lvl, csi.SumGs, best9Heros,
		csi.FarthestStage, csi.FarthestEliteStage, csi.FarthestHellStage,
		hc.GetHC(), sc.GetSC(helper.SC_Money), csi.FineIron, csi.BossCoin, csi.PvpCoin, csi.DestinyCoin, csi.EquipCoin, csi.GuildCoin,
		fmt.Sprintf("\"%v\"", dg_jades_info),
		dg_info,
		fmt.Sprintf("\"%v\"", title_info),
		hsoul.HeroSoulLevel,
		fmt.Sprintf("\"%v\"", eqRes),
		guild, guildAct,
		csi.TrialMostLvl,
		payInfo.MoneySum,
		fmt.Sprintf("\"%v\"", heroDestinyId),
		csi.VISB,
		newhand_str, dat["client_time_event"],
	)
	if _, err := cs.csv_profile_writer.WriteString(csvStr); err != nil {
		return err
	}

	logs.Info("RedisBiLog end key %s %s", key, accountNameId)
	return nil
}

type CustomInfo struct {
	Phone string
	//	RegTime            string
	//	LastLoginTime      string
	LoginDay int
	//	Vip                uint32
	//	CorpLvl            uint32
	SumGs              int
	ModuleGs           []int
	ArousalLvs         []uint32
	FarthestStage      int32
	FarthestEliteStage int32
	FarthestHellStage  int32
	//	Hc                 int64
	//	Money              int64
	FineIron     int64
	BossCoin     int64
	PvpCoin      int64
	DestinyCoin  int64
	EquipCoin    int64
	GuildCoin    int64
	CorpEquips   Equips
	Role1Equips  AvatarEquips
	Role2Equips  AvatarEquips
	Role3Equips  AvatarEquips
	DGJades      []string
	DestGen      DestinyGenerals
	GuildId      string
	GuildAct     int64
	TrialMostLvl int32
	Heros        HeroInfos
	HeroFashion  []string
	VISB         int64
}

type Equip struct {
	Slot       string
	Item       string
	Lv_upgrade uint32
	Lv_star    uint32
	Lv_matenh  uint32
	TrickGroup []string
	Jades      []string
}

func (e Equip) String() string {
	return fmt.Sprintf("[%s,%s,%d,%d,%d,%v,%v]",
		e.Slot, e.Item, e.Lv_upgrade, e.Lv_star, e.Lv_matenh, e.TrickGroup, e.Jades)
}

type AvatarEquip struct {
	Slot string
	Item string
}

func (ae AvatarEquip) String() string {
	return fmt.Sprintf("[%s,%s]", ae.Slot, ae.Item)
}

type Equips []Equip

func (es Equips) String() string {
	res := "\""
	for _, e := range es {
		res += e.String()
	}
	res += "\""
	return res
}

type AvatarEquips []AvatarEquip

func (aes AvatarEquips) String() string {
	res := "\""
	for _, e := range aes {
		res += e.String()
	}
	res += "\""
	return res
}

type DestinyGeneral struct {
	Id    int `json:"id"`
	Level int `json:"lv"`
}

func (ae DestinyGeneral) String() string {
	return fmt.Sprintf("[%d,%d]", ae.Id, ae.Level)
}

type DestinyGenerals []DestinyGeneral

func (aes DestinyGenerals) String() string {
	res := "\""
	for _, e := range aes {
		res += e.String()
	}
	res += "\""
	return res
}

type HeroInfo struct {
	Id           int
	Lvl          uint32
	StarLvl      uint32
	Talent       []uint32
	Pskill       []string
	Cskill       []string
	Tskill       []string
	SwingStarLvl int
	SwingLvl     int
}

func (h HeroInfo) String() string {
	return fmt.Sprintf("[%d,%d,%d,%v,%v,%v,%v,%d,%d]",
		h.Id, h.Lvl, h.StarLvl, h.Talent, h.Pskill, h.Cskill, h.Tskill, h.SwingStarLvl, h.SwingLvl)
}

type HeroInfos []HeroInfo

func (aes HeroInfos) String() string {
	res := "\""
	for _, e := range aes {
		res += e.String()
	}
	res += "\""
	return res
}

type BestHeroInfo struct {
	Id int
	Gs int
}

func (h BestHeroInfo) String() string {
	return fmt.Sprintf("[%d,%d]", h.Id, h.Gs)
}

type BestHeroInfos []BestHeroInfo

func (aes BestHeroInfos) String() string {
	res := "\""
	for _, e := range aes {
		res += e.String()
	}
	res += "\""
	return res
}

func getAct(nowTime int64, GuildUUID string, ActivityAddonTodayCurrGuild int64,
	ActivityAddonUUID string, ActivityAddonTime int64) int64 {
	if !util.IsSameUnixByStartTime(nowTime, ActivityAddonTime, util.TimeToBalance{
		WeekDay:   0,
		DailyTime: util.DailyTimeFromString(commonDailyStartTime),
	}) {
		ActivityAddonTime = nowTime
		ActivityAddonTodayCurrGuild = 0
	}
	if ActivityAddonUUID != GuildUUID {
		ActivityAddonTodayCurrGuild = 0
	}
	return ActivityAddonTodayCurrGuild
}

type bestHeroList []BestHeroInfo

func (pq bestHeroList) Len() int      { return len(pq) }
func (pq bestHeroList) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }
func (pq bestHeroList) Less(i, j int) bool {
	return pq[i].Gs > pq[j].Gs
}

//武将：     武将1id-强化等级-武将升星-品质-是否装备:数量;武将2id-强化等级-武将升星-品质-是否装备:数量
//宝石：     宝石1id-是否装备:数量;宝石2id-是否装备:数量
//武器：     武器1id-强化等级-品阶-星级-是否装备:数量;武器2id-强化等级-品阶-星级-是否装备:数量

func mergeHeroInfo(heroinfo *account.PlayerHero, eq []Equip, jade []string) string {
	var result string
	for i := 0; i < len(heroinfo.HeroLevel); i++ {
		if heroinfo.HeroStarLevel[i] <= 0 {
			continue
		}
		res := fmt.Sprintf("%d", i) + "-" + fmt.Sprintf("%d", heroinfo.HeroLevel[i]) + "-" + fmt.Sprintf("%d", heroinfo.HeroStarLevel[i]) + "-" +
			"1" + ":" + "1" + ";"
		result += res
	}

	for _, info := range eq {
		res := info.Item + "-" + fmt.Sprintf("%d", info.Lv_upgrade) + "-" +
			fmt.Sprintf("%d", info.Lv_matenh) + "-" + fmt.Sprintf("%d", info.Lv_star) + "-1" +
			":1;"
		result += res
	}
	var m map[string]int
	m = make(map[string]int)
	for _, info := range jade {
		if key, ok := m[info]; !ok {
			m[info] = 1
		} else {
			m[info] = key + 1
		}
	}

	for key, value := range m {
		res := key + "-" + "1" + ":" + fmt.Sprintf("%d", value) + ";"
		result += res
	}
	return result
}
