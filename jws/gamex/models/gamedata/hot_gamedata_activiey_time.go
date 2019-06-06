package gamedata

import (
	"time"

	"strconv"

	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// activity type
const (
	ActHitEgg = iota + 50
	ActPayPreDay
	ActLogin
	ActPay
	ActHcCost
	ActGameMode
	ActBuy
	ActHeroStar
	ActDayPay
	ActDayHcCost
	ActMoneyCat // 60
	ActGvgDailySign
	ActOnlyPay
	ActLevelMode
	ActRedPacket    // 64
	ActWhiteGacha   //65
	ActActivityRank // 66
	ActDestinyActivityRank
	ActJadeActivityRank
	ActEquipStarLvActivityRank
	ActHeroStarActivityRank // 70
	ActCorpLvActivityRank
	ActCorpGsActivityRank
	ActStageExchangePropLoot
	ActExchangeShop
	ActBlackGachaHero          // 武将巡礼 75
	ActBlackGachaWeapon        // 神兵再临 76
	ActRankByAstrology         //星图榜
	ActRankByHersoDestiny      //羁绊榜
	ActRankByWingStar          //幻甲星级榜
	ActRankByExclusiveWeapon   //神兵榜
	ActrankId_RankByWuShuangGs //无双战力榜
	ActSevenDaysRankTwo        //第二周排行榜入口
	ActRankByHeroJadeTwo       //第二周宝石积分排行榜
	ActH5Pray                  //h5祈愿
	ActLuckyWheel              //幸运转盘
	// ...
)

const (
	ActHeroGachaRace_Begin = 1001
	ActHeroGachaRace_End   = 9999

	// 投资英雄活动类型范围
	ActHeroFund_Begin = 10001
	ActHeroFund_End   = 19999

	// 节日Boss活动类型范围
	ActFestivalBoss_Begin = 20001
	ActFestivalBoss_End   = 20099
)

const (
	HotTime_Absolute = 1
	HotTime_Relative = 2

	HotTime_BeginHour = 5
)

const (
	AllOpen_shadId           = 0
	LimitOpen_shadId         = 1
	ByServerStartDays_shadId = 2
)

type HotActivityInfo2Client struct {
	ActivityId   uint32 `codec:"aid"`
	ActivityType uint32 `codec:"atyp"`
	StartTime    int64  `codec:"st"`
	EndTime      int64  `codec:"et"`
	SubIdCount   int    `codec:"subc"`
	TeleType     uint32 `codec:"teletyp"`
	TeleID       string `codec:"teleid"`
	ConditionID  uint32 `codec:"condid"`
	HotActivity  uint32 `codec:"hot"`
	TabIDs       string `codec:"tabids"`
}

type HotActivityInfo struct {
	ActivityId       uint32
	ActivityType     uint32
	ActivityParentID uint32
	StartTime        int64
	EndTime          int64
	Cfg              *ProtobufGen.HotActivityTime
}

type hotActivityData struct {
	serverGroup           map[uint32]*ProtobufGen.SERVERGROUP
	activityTypeInfoValid map[uint32][]*HotActivityInfo                        // 本服可用并有效的活动配置
	activityIdInfoValid   map[uint32]*HotActivityInfo                          // 本服可用并有效的活动配置activityID
	marketActivity        map[uint32]map[uint32]*ProtobufGen.HOTACTIVITYDETAIL // activityId
	// 限时神将相关
	gdServerGroup         map[uint32]uint32   // shard->group
	gdServerGroupActivity map[uint32][]uint32 // group->activityid
	gdChannelGroup        map[uint32][]string
	gdMoneyCatData        []*ProtobufGen.MONEYGOD
	gdMoneyWeight         []int32
	gdMoneyCatSubData     [][]*ProtobufGen.MONEYGOD_Num1
	gdServerGroupSbatch   map[uint32]*ProtobufGen.SEVERGROUP //  本服充值批次
	gdWuShuangGroup       map[uint32][]uint32                // groupID -> ServersId
	gdCSRobGroup          map[uint32][]uint32                // groupID -> ServersId
	gdWBGroup             map[uint32][]uint32                // groupID -> ServersId
	// 白盒宝箱相关
	gdWhiteGachaSetings map[uint32]*ProtobufGen.HOTGACHASETTINGS             //白盒宝箱setings
	gdWhiteNormalGacha  map[uint32]NormalGachaConfig                         // 白盒宝箱
	gdWhiteGachaShow    map[uint32][]*ProtobufGen.HOTGACHASHOW_ItemCondition // 白盒宝箱奖励展示信息
	gdWhiteGachaSpecil  map[uint32][]*ProtobufGen.HOTGACHASPECIAL            // 暗控奖励
	gdWhiteGachaLowest  map[uint32][]*ProtobufGen.HOTGACHALOWEST             // 保底奖励
	//幸运轮盘
	gdWheelSetings map[uint32]*ProtobufGen.WHEELSETTINGS
	gdWheelGacha   map[uint32]WheelGachaConfig
	gdWheelCost    map[uint32][]*ProtobufGen.WHEELCOST
	gdWheelShow    map[uint32][]*ProtobufGen.WHEELSHOW_ItemCondition
}

type NormalGachaConfig []*ProtobufGen.HOTNORMALGACHA
type WheelGachaConfig []*ProtobufGen.WHEELGACHA

func (cfg WheelGachaConfig) GetWeight(index int) int {
	return int(cfg[index].GetWeight())
}

func (cfg WheelGachaConfig) Len() int {
	return len(cfg)
}

func (cfg WheelGachaConfig) RandomConfigWheel(Items []string) *ProtobufGen.WHEELGACHA {
	tCfg := make(WheelGachaConfig, 0)
	tVis := make(map[string]bool)
	for _, value := range Items {
		tVis[value] = true
	}

	for _, value := range cfg {
		if !tVis[value.GetItemID()] {
			tCfg = append(tCfg, value)
		}
	}
	return tCfg[util.RandomItem(tCfg)]
}

func (cfg WheelGachaConfig) GetName(index int) string {
	return cfg[index].GetItemID()
}

func (cfg NormalGachaConfig) GetWeight(index int) int {
	return int(cfg[index].GetWeight())
}

func (cfg NormalGachaConfig) Len() int {
	return len(cfg)
}

func (cfg NormalGachaConfig) GetName(index int) string {
	return cfg[index].GetItemID()
}

func (cfg NormalGachaConfig) RandomConfig() *ProtobufGen.HOTNORMALGACHA {
	return cfg[util.RandomItem(cfg)]
}

func (cfg NormalGachaConfig) RandomConfigByCount(count int) []*ProtobufGen.HOTNORMALGACHA {
	rets := util.RandomItemByCount(cfg, count)
	retData := make([]*ProtobufGen.HOTNORMALGACHA, len(rets))
	for i, v := range rets {
		retData[i] = cfg[v]
	}
	return retData
}

type hotActivityMng struct {
}

func (act *hotActivityMng) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.HotActivityTime_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.activityTypeInfoValid = make(map[uint32][]*HotActivityInfo, len(dataList.GetItems()))
	datas.Activity.activityIdInfoValid = make(map[uint32]*HotActivityInfo, len(dataList.GetItems()))
	for _, data := range dataList.GetItems() {
		if data.GetActivityValid() <= 0 {
			continue
		}
		var ts, te time.Time
		if data.GetTimeType() == HotTime_Absolute {
			_ts, err := time.ParseInLocation("20060102_15:04", data.GetStartTime(), util.ServerTimeLocal)
			if err != nil {
				return err
			}
			_te, err := time.ParseInLocation("20060102_15:04", data.GetEndTime(), util.ServerTimeLocal)
			if err != nil {
				return err
			}
			ts, te = _ts, _te
		} else if data.GetTimeType() == HotTime_Relative {
			// 转换为绝对时间
			if len(game.Cfg.ShardId) <= 0 { // multiplayer 不用加载
				continue
			}
			// 取index=0的shardId, 行不行!
			serverStartTime := game.ServerStartTime(game.Cfg.ShardId[0])

			beginTime := util.GetCurDayTimeAtHour(serverStartTime, HotTime_BeginHour)

			sdays, err := strconv.ParseFloat(data.GetStartTime(), 32)
			if err != nil {
				return err
			}
			edays, err := strconv.ParseFloat(data.GetEndTime(), 32)
			if err != nil {
				return err
			}
			ts = time.Unix(int64(sdays)*util.DaySec+beginTime, 0).In(util.ServerTimeLocal)
			te = time.Unix(int64(edays)*util.DaySec+beginTime+int64(data.GetDuration())*util.MinSec, 0).In(util.ServerTimeLocal)
			logs.Debug("Convert Absolute time begin %d-%d-%d %d:%d:%d  %v, %d to %d-%d-%d %d:%d:%d %v, %d",
				ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), ts.Location(), ts.Unix(),
				te.Year(), te.Month(), te.Day(), te.Hour(), te.Minute(), te.Second(), te.Location(), te.Unix())
		} else {
			logs.Error("Illegale hot time type: %d", data.GetTimeType())
			continue
		}

		serverCfg, ok := datas.Activity.serverGroup[data.GetServerGroupID()]
		if !ok || !_checkServerShardValid(serverCfg) {
			continue
		}
		if data.GetActivityType() >= uint32(ActHeroGachaRace_Begin) &&
			data.GetActivityType() <= uint32(ActHeroGachaRace_End) {
			has := false
			if len(game.Cfg.ShardId) <= 0 {
				continue
			}
			// 单服怎么办 ？
			if data.GetActivityID() == datas.Activity.GetHGRHotID(uint32(game.Cfg.ShardId[0])) {
				has = true
			} else {
				// 多服怎么办?
				for _, id := range datas.Activity.GetShardActivities(uint32(game.Cfg.ShardId[0])) {
					if data.GetActivityID() == id {
						has = true
						break
					}
				}
			}
			if !has {
				continue
			}
		}
		ha := &HotActivityInfo{
			ActivityId:       data.GetActivityID(),
			ActivityType:     data.GetActivityType(),
			ActivityParentID: data.GetActivityPID(),
			StartTime:        ts.Unix(),
			EndTime:          te.Unix(),
			Cfg:              data,
		}
		o, ok := datas.Activity.activityTypeInfoValid[data.GetActivityType()]
		if ok {
			o = append(o, ha)
			datas.Activity.activityTypeInfoValid[data.GetActivityType()] = o
		} else {
			datas.Activity.activityTypeInfoValid[data.GetActivityType()] = []*HotActivityInfo{ha}
		}
		datas.Activity.activityIdInfoValid[data.GetActivityID()] = ha
	}
	return nil
}

func (d hotActivityData) GetActivitySimpleInfoById(activityId uint32) *HotActivityInfo {
	return d.activityIdInfoValid[activityId]
}

func (d hotActivityData) GetActActivity(nt int64, activityType int, channelId string) uint32 {
	info := d.GetActivityInfoValid(activityType, channelId, nt)
	if len(info) > 0 {
		return info[0].ActivityId
	}
	return 0
}

func (d hotActivityData) GetAllActivitySimpleInfo() map[uint32]*HotActivityInfo {
	return d.activityIdInfoValid
}

func (d hotActivityData) GetAllActivitySimpleInfoByChannel(channelID string) map[uint32]*HotActivityInfo {
	ret := make(map[uint32]*HotActivityInfo, 0)
	for k, v := range d.activityIdInfoValid {
		if d.IsInChannelAct(channelID, v) {
			ret[k] = v
		}
	}
	return ret
}

func (d hotActivityData) GetActivitySimpleInfo(activityType uint32) []*HotActivityInfo {
	return d.activityTypeInfoValid[activityType]
}

func (d hotActivityData) GetActivityInfoFilterTime(activityType uint32, now int64) []*HotActivityInfo {
	list := d.activityTypeInfoValid[activityType]
	ret := make([]*HotActivityInfo, 0)
	for _, ac := range list {
		if ac.StartTime <= now && ac.EndTime >= now {
			ret = append(ret, ac)
		}
	}

	return ret
}

func (d hotActivityData) GetAllActivityInfo2Client(channelID string) ([]HotActivityInfo2Client,
	[]MarketSubActivityConfig2Client, []MarketSubActivityConfigReward2Client) {
	res := make([]HotActivityInfo2Client, 0, len(d.activityTypeInfoValid))
	resSub := make([]MarketSubActivityConfig2Client, 0, len(d.activityTypeInfoValid)*8)
	resRew := make([]MarketSubActivityConfigReward2Client, 0, len(d.activityTypeInfoValid)*16)

	for _, _v := range d.activityTypeInfoValid {
		for _, v := range _v {
			if !d.IsInChannelAct(channelID, v) {
				continue
			}
			info := HotActivityInfo2Client{
				ActivityId:   v.ActivityId,
				ActivityType: v.ActivityType,
				StartTime:    v.StartTime,
				EndTime:      v.EndTime,
				SubIdCount:   len(d.marketActivity[v.ActivityId]),
				TeleType:     v.Cfg.GetTeleType(),
				TeleID:       v.Cfg.GetTeleID(),
				ConditionID:  v.Cfg.GetConditionID(),
				HotActivity:  v.Cfg.GetHotActivity(),
				TabIDs:       v.Cfg.GetTabIDS(),
			}
			res = append(res, info)
			subActCfg := d.marketActivity[v.ActivityId]

			for i := 1; i <= info.SubIdCount; i++ {
				cfg := subActCfg[uint32(i)]
				//Kora 特别更改
				if isHeroFoundActivity(int(v.ActivityType)) && channelID == util.Android_Enjoy_Korea_GP_Channel {
					resSub = append(resSub, MarketSubActivityConfig2Client{
						FCCondTyp:   cfg.GetFCType(),
						FCParam1:    cfg.GetFCValue1(),
						FCParam2:    cfg.GetFCValue2(),
						FCParam3:    cfg.GetSFCValue1(),
						FCParam4:    Android_Enjoy_Korea_GP_HeroFound_Iap,
						RewardCount: len(cfg.GetItem_Table()),
					})
				} else if isHeroFoundActivity(int(v.ActivityType)) && channelID == util.Android_Enjoy_Korea_OneStore_Channel {
					resSub = append(resSub, MarketSubActivityConfig2Client{
						FCCondTyp:   cfg.GetFCType(),
						FCParam1:    cfg.GetFCValue1(),
						FCParam2:    cfg.GetFCValue2(),
						FCParam3:    cfg.GetSFCValue1(),
						FCParam4:    Android_Enjoy_Korea_OneStore_HeroFoud_Iap,
						RewardCount: len(cfg.GetItem_Table()),
					})
				} else {
					resSub = append(resSub, MarketSubActivityConfig2Client{
						FCCondTyp:   cfg.GetFCType(),
						FCParam1:    cfg.GetFCValue1(),
						FCParam2:    cfg.GetFCValue2(),
						FCParam3:    cfg.GetSFCValue1(),
						FCParam4:    cfg.GetSFCValue2(),
						RewardCount: len(cfg.GetItem_Table()),
					})
				}
				for _, rew := range cfg.GetItem_Table() {
					resRew = append(resRew, MarketSubActivityConfigReward2Client{
						ItemId:    rew.GetItemID(),
						ItemCount: rew.GetItemCount(),
					})
				}
			}
		}
	}
	return res, resSub, resRew
}

func (d hotActivityData) GetActivityInfo(activityType int, channelID string) []*HotActivityInfo {
	act := d.activityTypeInfoValid[uint32(activityType)]
	ret := make([]*HotActivityInfo, 0)
	for _, a := range act {
		if d.IsInChannelAct(channelID, a) {
			ret = append(ret, a)
		}
	}
	return ret
}

func (d hotActivityData) GetActivityInfoValid(activityType int, channelID string, nowT int64) []*HotActivityInfo {
	act := d.activityTypeInfoValid[uint32(activityType)]
	ret := make([]*HotActivityInfo, 0)
	for _, a := range act {
		if d.IsInChannelAct(channelID, a) && nowT > a.StartTime && nowT < a.EndTime {
			ret = append(ret, a)
		}
	}
	return ret
}

func (d hotActivityData) GetAllActivityInfoValid(channelID string, nowT int64) []*HotActivityInfo {
	ret := make([]*HotActivityInfo, 0)
	for _, a := range d.activityIdInfoValid {
		if d.IsInChannelAct(channelID, a) && nowT >= a.StartTime && nowT < a.EndTime {
			ret = append(ret, a)
		}
	}
	return ret
}

func (d hotActivityData) IsInChannelAct(channelID string, hotInfo *HotActivityInfo) bool {
	actChannel := d.gdChannelGroup[hotInfo.Cfg.GetChannelGroupID()]
	// 检查活动服务器分组类型是否2，是否在开服活动时间范围内
	// nil slice 代表所有渠道均开放
	actServerGroup := d.IsInServerStartTimeAct(hotInfo.Cfg.GetServerGroupID(), hotInfo.StartTime)
	if actChannel == nil && actServerGroup {
		return true
	}
	for _, c := range actChannel {
		if c == channelID && d.IsInServerStartTimeAct(hotInfo.Cfg.GetServerGroupID(), hotInfo.StartTime) {
			return true
		}
	}
	return false
}

func (d hotActivityData) IsInServerStartTimeAct(groupId uint32, startTime int64) bool {
	actServerCfg := d.serverGroup[groupId]
	if actServerCfg.GetServerGroupType() == ByServerStartDays_shadId {
		if actServerCfg.GetServerGroupSubType() == 0 || actServerCfg.GetServerGroupSubType() == uint32(game.Cfg.Gid) {
			theDay := d.getDay(game.ServerStartTime(game.Cfg.ShardId[0]), startTime)
			if theDay < 0 {
				return false
			}
			if theDay >= actServerCfg.GetAccCon_Table()[0].GetServerGroupValue1() &&
				theDay <= actServerCfg.GetAccCon_Table()[0].GetServerGroupValue2() {
				actServerCfg.GetAccCon_Table()[0].GetServerGroupValue2()
				return true
			}
			return false
		}
		return false
	}
	return true

}

func (d hotActivityData) getDay(sTime, nTime int64) uint32 {
	s1 := time.Unix(sTime, 0)
	s2 := time.Unix(nTime, 0)

	sumD := d.timeSubDays(s2, s1)
	return uint32(sumD)
}

func (d hotActivityData) timeSubDays(t1, t2 time.Time) int {
	if t1.Location().String() != t2.Location().String() {
		return -1
	}
	hours := t1.Sub(t2).Hours()

	if hours <= 0 {
		return -1
	}
	// sub hours less than 24
	if hours < 24 {
		// may same day
		t1y, t1m, t1d := t1.Date()
		t2y, t2m, t2d := t2.Date()
		isSameDay := (t1y == t2y && t1m == t2m && t1d == t2d)

		if isSameDay {

			return 0
		} else {
			return 1
		}

	} else { // equal or more than 24

		if (hours/24)-float64(int(hours/24)) == 0 { // just 24's times
			return int(hours / 24)
		} else { // more than 24 hours
			return int(hours/24) + 1
		}
	}

}

func (d hotActivityData) debugSetActivityTime(activityId int, s, e int64) {
	id_info, ok := d.activityIdInfoValid[uint32(activityId)]
	if ok {
		id_info.StartTime = s
		id_info.EndTime = e
	}
}

type hotActivityServerGroupData struct {
}

func (sg *hotActivityServerGroupData) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.SERVERGROUP_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}
	datas.Activity.serverGroup = make(map[uint32]*ProtobufGen.SERVERGROUP, len(dataList.GetItems()))
	for _, data := range dataList.GetItems() {
		switch data.GetServerGroupType() {
		case AllOpen_shadId, LimitOpen_shadId, ByServerStartDays_shadId:
			datas.Activity.serverGroup[data.GetServerGroupID()] = data
		default:
			return fmt.Errorf("hotActivityServerGroup ServerGroupType not define  %d", data.GetServerGroupType())
		}
	}
	return nil
}

func _checkServerShardValid(cfg *ProtobufGen.SERVERGROUP) bool {
	if cfg.GetServerGroupType() == AllOpen_shadId || cfg.GetServerGroupType() == ByServerStartDays_shadId {
		return true
	}
	if len(game.Cfg.ShardId) <= 0 {
		return false
	}
	shard := uint32(game.Cfg.ShardId[0])
	if cfg.GetServerGroupType() == LimitOpen_shadId {
		for _, c := range cfg.AccCon_Table {
			if shard >= c.GetServerGroupValue1() && shard <= c.GetServerGroupValue2() {
				return true
			}
		}
	}
	return false
}
