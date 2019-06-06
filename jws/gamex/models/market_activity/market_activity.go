package market_activity

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	MA_ST_INIT = iota
	MA_ST_ACT
	MA_ST_GOT
)

const (
	MA_ST_OK = iota
	MA_ST_NONE
	MA_ST_GM_CLOSE
)

var (
	act2gmtools = map[int]int{
		gamedata.ActHitEgg:           uutil.Hot_Value_HitEeg,
		gamedata.ActPayPreDay:        uutil.Hot_Value_Total_Query_day,
		gamedata.ActLogin:            uutil.Hot_Value_Total_Enter_day,
		gamedata.ActPay:              uutil.Hot_Value_Total_Query,
		gamedata.ActHcCost:           uutil.Hot_Value_Total_Buy,
		gamedata.ActGameMode:         uutil.Hot_Value_Total_Play,
		gamedata.ActBuy:              uutil.Hot_Value_Total_Buy_Resource,
		gamedata.ActHeroStar:         uutil.Hot_Value_Star_Hero,
		gamedata.ActDayPay:           uutil.Hot_Value_Day_Total_Query,
		gamedata.ActDayHcCost:        uutil.Hot_Value_Day_Total_Buy,
		gamedata.ActMoneyCat:         uutil.Hot_Value_Money_Cat,
		gamedata.ActOnlyPay:          uutil.Hot_Value_Only_Pay,
		gamedata.ActRedPacket:        uutil.Hot_Value_Red_Packet,
		gamedata.ActActivityRank:     uutil.Hot_Value_ActivityRank,
		gamedata.ActExchangeShop:     uutil.Hot_value_ExchangeShop,
		gamedata.ActSevenDaysRankTwo: uutil.Hot_value_SevenRankTwo,
	}
)

type MarketActivityInfo2Client struct {
	ActivityId   uint32  `codec:"aid"`
	Value        int64   `codec:"v"`
	StateCount   int     `codec:"st_c"`
	Param        int64   `codec:"param"`
	OnlyPayCount []int64 `codec:"only_pay_count"`
}

type PlayerMarketActivity struct {
	ActivityId       uint32  `json:"id"`
	ActivityType     uint32  `json:"type"`
	ParentActivityID uint32  `json:"p_type"`
	Value            int64   `json:"v"`
	State            []int   `json:"state"` // 0未激活，1激活未领取，2已领取
	LastUpdateTime   int64   `json:"lt"`
	TmpValue         []int64 `json:"tmp"`      // 用来记录计算中间值，如每日充值金额
	Balanced         bool    `json:"balanced"` // 是否已经结算过了
	handler          func(uint32)
}

type PlayerMarketActivitys struct {
	Activitys   []PlayerMarketActivity `json:"activitys"`
	DataBuild   int                    `json:"data_build"`
	ChannelID   string                 `json:"channel_id"`
	handleProto func(uint32)
	helper.SyncObj
}

func (ma *PlayerMarketActivitys) RegHandler(f func(uint32)) {
	ma.handleProto = f
	for i, _ := range ma.Activitys {
		ma.Activitys[i].handler = f
	}
}

func (ma *PlayerMarketActivitys) OnGenMarketAcvititys(acid string, now_t int64, channelID string) {
	if ma.ChannelID == "" {
		ma.ChannelID = channelID
	}
	if ma.Activitys == nil {
		ma.initMarketActivitys(channelID)
	} else {
		ma.UpdateMarketActivity(acid, now_t)
	}
	// 登录活动
	ma.OnLogin(acid, now_t)
	ma.UpdateGVGDailySignOnLogin(acid, now_t)
}

func (ma *PlayerMarketActivitys) initMarketActivitys(channelID string) {
	activityCfg := gamedata.GetHotDatas().Activity
	allActs := activityCfg.GetAllActivitySimpleInfoByChannel(channelID)
	ma.Activitys = make([]PlayerMarketActivity, 0, len(allActs))
	for _, v := range allActs {
		subCfg := activityCfg.GetMarketActivitySubConfig(v.ActivityId)
		ma.Activitys = append(ma.Activitys, ma._genPlayerMarketActivity(v, subCfg))
	}
	ma.DataBuild = gamedata.GetHotDataVerCfg().Build
}

func (ma *PlayerMarketActivitys) GetMarketActivityById(id uint32) *PlayerMarketActivity {
	for i, a := range ma.Activitys {
		if a.ActivityId == id {
			return &ma.Activitys[i]
		}
	}
	return nil
}

func (ma *PlayerMarketActivitys) GetMarketActivityForClient(acid string, now_t int64) (
	[]MarketActivityInfo2Client, []int) {
	r := make([]MarketActivityInfo2Client, 0, len(ma.Activitys))
	rs := make([]int, 0, len(ma.Activitys)*8)
	for _, a := range ma.Activitys {
		_info := MarketActivityInfo2Client{
			ActivityId: a.ActivityId,
			Value:      a.Value,
			StateCount: len(a.State),
		}
		if a.ActivityType == gamedata.ActPayPreDay {
			_info.Param = ma.GetCurDayHcPay(acid, now_t)
		}
		if a.ActivityType == gamedata.ActOnlyPay {
			for i := 1; i < len(a.TmpValue); i += 2 {
				_info.OnlyPayCount = append(_info.OnlyPayCount, a.TmpValue[i])
			}
		}
		r = append(r, _info)
		rs = append(rs, a.State...)
	}

	return r, rs
}

func (ma *PlayerMarketActivitys) ForceUpdateMarketActivity(acid string, now_t int64) bool {
	return ma.updateMarketActivity(acid, now_t, true)
}

func (ma *PlayerMarketActivitys) updateMarketActivity(acid string, now_t int64, force bool) bool {
	// 检查是否有活动过期
	activityCfg := gamedata.GetHotDatas().Activity
	for i, _ := range ma.Activitys {
		if ma.Activitys[i].Balanced {
			continue
		}
		cfg := activityCfg.GetActivitySimpleInfoById(ma.Activitys[i].ActivityId)
		if cfg != nil {
			if now_t >= cfg.EndTime {
				ma.Activitys[i]._balanceMarketActivity(acid)
			}
		}
	}
	// 检查数据是否变了
	hotBuild := gamedata.GetHotDataVerCfg().Build
	if ma.DataBuild == hotBuild && !force {
		return false
	}
	logs.Debug("UpdateMarketActivity data change %v %v %s", ma.DataBuild, gamedata.HotDataVerCfg.Build, ma.ChannelID)
	ma.DataBuild = hotBuild
	// 删除的活动和变更的活动
	delId := make(map[int]struct{}, len(ma.Activitys))
	allActs := activityCfg.GetAllActivitySimpleInfoByChannel(ma.ChannelID)
	logs.Debug("got all activites channel,%s, %v", ma.ChannelID, allActs)
	for id, ov := range ma.Activitys {
		nv, ok := allActs[ov.ActivityId]
		if !ok || !activityCfg.IsInChannelAct(ma.ChannelID, nv) {
			delId[id] = struct{}{}
		} else {
			subCfg := activityCfg.GetMarketActivitySubConfig(nv.ActivityId)
			if len(subCfg) > len(ov.State) {
				pov := &ma.Activitys[id]
				pov._chgPlayerMarketActivity(nv, subCfg)
			}
		}
	}
	logs.Debug("del activities %v", delId)
	oldId := make(map[uint32]struct{}, len(ma.Activitys))
	nact := make([]PlayerMarketActivity, 0, len(ma.Activitys))
	for id, ov := range ma.Activitys {
		if _, ok := delId[id]; ok {
			//ov._balanceMarketActivity(acid)
		} else {
			nact = append(nact, ov)
			oldId[ov.ActivityId] = struct{}{}
		}
	}
	logs.Debug("old activities %v", oldId)
	ma.Activitys = nact

	// 新加的活动
	for id, nv := range allActs {
		if _, ok := oldId[id]; !ok {
			nvSubCfg := activityCfg.GetMarketActivitySubConfig(nv.ActivityId)
			ma.Activitys = append(ma.Activitys, ma._genPlayerMarketActivity(nv, nvSubCfg))
		}
	}
	logs.Debug("ma activies %v", ma.Activitys)
	// 检查并更新各个活动的state, 无需再调用UpdateMarketActivity
	ma.UpdateOnHeroStar(acid, now_t)
	ma.UpdateOnGameMode(acid, now_t)
	ma.UpdateOnBy(acid, now_t)
	ma.UpdateOnHcCost(acid, now_t)
	ma.UpdateOnPay(acid, now_t)
	ma.UpdateonPayPreDay(acid, now_t)
	ma.UpdateOnLogin(acid, now_t)
	ma.UpdateOnLevel(acid, now_t)
	ma.UpdateOnlyPay(acid, now_t)
	ma.UpdateOnHeroFund(acid, now_t)
	return true
}

func (ma *PlayerMarketActivitys) UpdateMarketActivity(acid string, now_t int64) bool {
	return ma.updateMarketActivity(acid, now_t, false)
}

func (pma *PlayerMarketActivity) _balanceMarketActivity(acid string) {
	if pma.Balanced {
		return
	}
	pma.Balanced = true
	activityCfg := gamedata.GetHotDatas().Activity
	subCfg := activityCfg.GetMarketActivitySubConfig(pma.ActivityId)

	awards := make(map[string]uint32, 4)
	pAct := activityCfg.GetActivitySimpleInfoById(pma.ParentActivityID)
	if pma.ActivityType == gamedata.ActOnlyPay {
		if subCfg == nil {
			return
		}
		for i, s := range pma.State {
			if s == MA_ST_ACT {

				cfg := subCfg[uint32(i+1)]
				if cfg != nil {
					for _, award := range cfg.Item_Table {
						c, ok := awards[award.GetItemID()]
						if ok {
							awards[award.GetItemID()] = c + (award.GetItemCount() * uint32(pma.TmpValue[i*2]-pma.TmpValue[i*2+1]))
						} else {
							awards[award.GetItemID()] = award.GetItemCount() * uint32(pma.TmpValue[i*2]-pma.TmpValue[i*2+1])
						}
					}
				}
				pma.State[i] = MA_ST_GOT
			}
		}
	} else if pma.ActivityType == gamedata.ActExchangeShop ||
		pma.ActivityType == gamedata.ActStageExchangePropLoot {
		for i, _ := range pma.State {
			pma.State[i] = MA_ST_GOT
		}
		if pma.ActivityType == gamedata.ActExchangeShop {
			pma.handler(pma.ActivityId)
		}
	} else if pAct != nil && pAct.ActivityType == gamedata.ActActivityRank {
		for i, _ := range pma.State {
			pma.State[i] = MA_ST_GOT
		}
	} else {
		if subCfg == nil {
			return
		}
		for i, s := range pma.State {
			if s == MA_ST_ACT {
				cfg := subCfg[uint32(i+1)]
				if cfg != nil {
					for _, award := range cfg.Item_Table {
						c, ok := awards[award.GetItemID()]
						if ok {
							awards[award.GetItemID()] = c + award.GetItemCount()
						} else {
							awards[award.GetItemID()] = award.GetItemCount()
						}
					}
				}
				pma.State[i] = MA_ST_GOT
			}
		}
	}

	if len(awards) <= 0 {
		return
	}
	account, _ := db.ParseAccount(acid)
	if !pma.GetActValueByActTyp(account.ShardId, int(pma.ActivityType)) {
		logs.Info("%d Not active", pma.ActivityType)
		return
	}
	// 发邮件
	titleIDS := gamedata.GetMarketActivityMailType(int(pma.ActivityType))
	if titleIDS == 0 {
		logs.Error("fail to find titleIds Activityid :%d , typ: %d", pma.ActivityId, pma.ActivityType)
		return
	}
	param := []string{}
	if subCfg != nil {
		switch pma.ActivityType {
		case gamedata.ActGameMode:
			param = []string{fmt.Sprintf("%d", subCfg[1].GetFCValue2())}
		case gamedata.ActBuy:
			param = []string{subCfg[1].GetSFCValue1()}
		}
	}
	mail_sender.SendMarketActivityMail(acid, titleIDS, param, awards)
	item_id := make([]string, 0, len(awards))
	count := make([]uint32, 0, len(awards))
	for k, v := range awards {
		item_id = append(item_id, k)
		count = append(count, v)
	}
}

func (ma *PlayerMarketActivity) isNeedBalance() bool {
	typ := ma.ActivityType
	return !(typ == gamedata.ActExchangeShop ||
		typ == gamedata.ActStageExchangePropLoot)
}

func (ma *PlayerMarketActivitys) getActByTyp(activityType uint32) []*PlayerMarketActivity {
	res := make([]*PlayerMarketActivity, 0, 4)
	for i, a := range ma.Activitys {
		if a.ActivityType == activityType {
			res = append(res, &ma.Activitys[i])
		}
	}
	return res
}

func (ma *PlayerMarketActivitys) getActByTypeRange(beginType, endType uint32) []*PlayerMarketActivity {
	res := make([]*PlayerMarketActivity, 0, 4)
	for i, a := range ma.Activitys {
		if beginType <= a.ActivityType && a.ActivityType <= endType {
			res = append(res, &ma.Activitys[i])
		}
	}
	return res
}

func (ma *PlayerMarketActivitys) getAllActivityByType(acid string, now_t int64, activityType uint32) ([]*PlayerMarketActivity, int) {
	_pa := ma.getActByTyp(activityType)
	if _pa == nil || len(_pa) <= 0 {
		return nil, MA_ST_NONE
	}
	account, _ := db.ParseAccount(acid)
	hotType := GetHotTypeByActType(activityType)
	//活动是否开启
	if !game.Cfg.GetHotActValidData(account.ShardId, hotType) {
		return nil, MA_ST_GM_CLOSE
	}
	return _pa, MA_ST_OK
}

func (ma *PlayerMarketActivitys) _genPlayerMarketActivity(v *gamedata.HotActivityInfo,
	subCfg map[uint32]*ProtobufGen.HOTACTIVITYDETAIL) PlayerMarketActivity {
	r := PlayerMarketActivity{
		ActivityId:       v.ActivityId,
		ActivityType:     v.ActivityType,
		ParentActivityID: v.ActivityParentID,
		State:            make([]int, len(subCfg)),
	}

	if v.ActivityType == gamedata.ActPayPreDay {
		days := gamedata.GetCommonDayDiffC(v.StartTime, v.EndTime)
		r.TmpValue = make([]int64, days)
	} else if v.ActivityType == gamedata.ActExchangeShop {
		r.TmpValue = make([]int64, len(gamedata.GetHotDatas().HotExchangeShopData.GetExchangePropShowData(v.ActivityId)))
	}
	r.handler = ma.handleProto
	return r
}

func (pma *PlayerMarketActivity) _chgPlayerMarketActivity(v *gamedata.HotActivityInfo,
	subCfg map[uint32]*ProtobufGen.HOTACTIVITYDETAIL) {

	if len(subCfg) > len(pma.State) {
		stat_tmp := make([]int, len(subCfg))
		copy(stat_tmp, pma.State)
		pma.State = stat_tmp

		if v.ActivityType == gamedata.ActPayPreDay {
			days := gamedata.GetCommonDayDiffC(v.StartTime, v.EndTime)
			pma._TmpValueCheck(int(days))
		} else if v.ActivityType == gamedata.ActOnlyPay {
			pma._OnlypayTmpValueCheck(len(subCfg))
		}
	}
}

func (pma *PlayerMarketActivity) _TmpValueCheck(days int) {
	if days >= len(pma.TmpValue) {
		tmp := make([]int64, days+1)
		copy(tmp, pma.TmpValue)
		pma.TmpValue = tmp
	}
}

func (pma *PlayerMarketActivity) _OnlypayTmpValueCheck(length int) {
	if length*2 >= len(pma.TmpValue) {
		tmp := make([]int64, length*2)
		copy(tmp, pma.TmpValue)
		pma.TmpValue = tmp
	}
}

func (pma *PlayerMarketActivity) _getTmpValue(day int) int64 {
	pma._TmpValueCheck(day)
	return pma.TmpValue[day]
}

func (pma *PlayerMarketActivity) isActAvailable(acid string, now_t int64) bool {
	if pma.Balanced {
		return false
	}
	activityCfg := gamedata.GetHotDatas().Activity
	simpleCfg := activityCfg.GetActivitySimpleInfoById(pma.ActivityId)
	if simpleCfg == nil {
		return false
	}
	if now_t >= simpleCfg.EndTime || now_t < simpleCfg.StartTime {
		return false
	}
	return true
}

func (ma *PlayerMarketActivitys) DebugReset(channelID string) {
	if ma.ChannelID == "" {
		ma.ChannelID = channelID
	}
	ma.initMarketActivitys(channelID)
}

func (ma *PlayerMarketActivity) GetActValueByActTyp(sid uint, actType int) bool {
	gmTyp := GetHotTypeByActType(uint32(actType))
	return game.Cfg.GetHotActValidData(sid, gmTyp)
}

func GetHotTypeByActType(actType uint32) int {
	if actType >= gamedata.ActHeroFund_Begin && actType <= gamedata.ActHeroFund_End {
		return uutil.Hot_Value_HERO_FUND
	}
	return act2gmtools[int(actType)]
}
