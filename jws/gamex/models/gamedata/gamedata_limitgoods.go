package gamedata

import (
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Load_Limit_Good_Ok = iota
	Load_Limit_Good_Not_Valid
	Load_Limit_Good_Server_Group
)

// 限时商城
type LimitGoods2Client struct {
	GoodId         int         `codec:"gid"`
	GoodName       string      `codec:"gname"`
	GoodType       int         `codec:"gtype"`
	StartTime      int64       `codec:"stime"`
	Duration       int         `codec:"duration"`
	CoinItemId     string      `codec:"cid"`
	CurrentPrice   int         `codec:"cuprice"`
	OriginalCost   int         `codec:"orprice"`
	GoodItems      [][]byte    `codec:"gooditems"`
	GoodIcon       string      `codec:"gicon"`
	Discount       int         `codec:"disc"`
	VipLimit       int         `codec:"vpl"`
	LimitCount     int         `codec:"limit_count"`
	GoodItemsArray []LimitItem `codec:"-"`
}

type LimitItem struct {
	ItemId    string `codec:"itemid"`
	ItemCount int    `codec:"itemcount"`
}

// server配置
type hotLimitGood struct {
	Items []HotLimitGoodCfg
}

type HotLimitGoodCfg struct {
	Item      *ProtobufGen.LIMITGOODS
	StartTime int64
	Duration  int // 单位S
}

// implements IHotDataMng
type hotLimitGoodsMng struct {
}

func (act *hotLimitGoodsMng) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.LIMITGOODS_ARRAY{}

	if err := proto.Unmarshal(buffer, dataList); err != nil {
		logs.Error("load hot limit good error, %v", err)
		return err
	}
	datas.LimitGoodConfig.Items = make([]HotLimitGoodCfg, 0, len(dataList.GetItems()))
	if len(game.Cfg.ShardId) <= 0 { // multiplayer 不用加载
		logs.Warn("load limit good, shardId is nil")
		return nil
	}
	shardId := uint32(game.Cfg.ShardId[0])
	for _, item := range dataList.GetItems() {
		if retCode := act.isAvailable(item, datas, shardId); retCode == 0 {
			startTimeUnix := getRealStartTimeUnix(shardId, item)
			limitConfig := HotLimitGoodCfg{Item: item, StartTime: startTimeUnix, Duration: int(item.GetDuration() * 3600)}
			datas.LimitGoodConfig.Items = append(datas.LimitGoodConfig.Items, limitConfig)

		} else {
			logs.Debug("load limit good, not avaliable, %d", retCode)
		}
	}
	return nil
}

func getRealStartTimeUnix(shardId uint32, item *ProtobufGen.LIMITGOODS) int64 {
	if item.GetTimeType() == 1 {
		startTemp, err := time.ParseInLocation("20060102_15:04", item.GetStartTime(), util.ServerTimeLocal)
		if err != nil {
			logs.Error("parse limit good start time error, %v", err)
		}
		return startTemp.Unix()
	} else {
		sst := game.ServerStartTime(uint(shardId))
		i64, err := strconv.ParseInt(item.GetStartTime(), 10, 64)
		if err != nil {
			logs.Error("parse limit good start time error, %v", err)
		}
		beginTime := util.GetCurDayTimeAtHour(sst, HotTime_BeginHour)
		realUnix := beginTime + i64*86400
		logs.Debug("server start time %d, limit good start time %d", sst, realUnix)
		return realUnix
	}
}

func (act *hotLimitGoodsMng) isAvailable(item *ProtobufGen.LIMITGOODS, datas *HotDatas, shardId uint32) int {
	if item.GetActivityValid() == 0 {
		return Load_Limit_Good_Not_Valid
	}
	serverGroupCfg := datas.Activity.serverGroup[item.GetServerGroupID()]
	if serverGroupCfg.GetServerGroupType() == AllOpen_shadId || serverGroupCfg.GetServerGroupType() == ByServerStartDays_shadId {
		return Load_Limit_Good_Ok
	}
	for _, cfg := range serverGroupCfg.GetAccCon_Table() {
		if cfg.GetServerGroupValue1() <= shardId && shardId <= cfg.GetServerGroupValue2() {
			return Load_Limit_Good_Ok
		} else {
			logs.Debug("load limit good, %d, %d", cfg.GetServerGroupValue1(), cfg.GetServerGroupValue2())
		}
	}
	return Load_Limit_Good_Server_Group
}

func (h hotLimitGood) GetAllLimitGoodForClient() []LimitGoods2Client {
	ret := make([]LimitGoods2Client, len(h.Items))
	for i, item := range h.Items {
		serverGroupCfg := GetHotDatas().Activity.serverGroup[item.Item.GetServerGroupID()]
		if serverGroupCfg.GetServerGroupType() == ByServerStartDays_shadId {
			is2Client := GetHotDatas().Activity.IsInServerStartTimeAct(item.Item.GetServerGroupID(), item.StartTime)
			if !is2Client {
				continue
			}
		}
		ret[i] = LimitGoods2Client{
			GoodId:         int(item.Item.GetLimitGoodsID()),
			GoodName:       item.Item.GetGoodsName(),
			GoodType:       int(item.Item.GetGoodsType()),
			StartTime:      item.StartTime,
			Duration:       int(item.Duration),
			CoinItemId:     item.Item.GetCoinItemID(),
			CurrentPrice:   int(item.Item.GetCurrentPrice()),
			OriginalCost:   int(item.Item.GetOriginalCost()),
			GoodItemsArray: convertLimitItem(item.Item.GetFixed_Loot()),
			GoodIcon:       item.Item.GetGoodsIcon(),
			Discount:       int(item.Item.GetDiscount()),
			VipLimit:       int(item.Item.GetVIPLevel()),
			LimitCount:     int(item.Item.GetTimesLimit()),
		}
	}
	logs.Debug("send limit goods to client, %v", ret)
	return ret
}

func convertLimitItem(loots []*ProtobufGen.LIMITGOODS_Loot) []LimitItem {
	ret := make([]LimitItem, len(loots))
	for i, loot := range loots {
		ret[i].ItemId = loot.GetItemID()
		ret[i].ItemCount = int(loot.GetGoodsCount())
	}
	return ret
}

func (h hotLimitGood) GetLimitGoodConfig(goodId int64) (*HotLimitGoodCfg, bool) {
	for _, config := range h.Items {
		if config.Item.GetLimitGoodsID() == uint32(goodId) {
			return &config, true
		}
	}
	return nil, false
}
