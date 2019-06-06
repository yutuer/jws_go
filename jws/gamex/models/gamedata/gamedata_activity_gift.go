package gamedata

import (
	"math/rand"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	ActivityGift_Get_Typ_Pass = iota
	ActivityGift_Get_Typ_Total
)

type giftDailyData struct {
	Base     PriceDatas //现在就两种，一种是基础的，一种是VIP
	Vip      PriceDatas
	VipAddon PriceDatas //用于VIp补领
	VipNeed  uint32
}

func (g *giftDailyData) setDailyItem(data *ProtobufGen.DAILYGIFT) {

	g.VipNeed = data.GetVIPBase()

	g.Base = PriceDatas{}
	g.Vip = PriceDatas{}
	g.VipAddon = PriceDatas{}

	// 1
	g.setADailyItem(data.GetAWardID(), data.GetCount(), data.GetVIPBonus())
	// 2
	g.setADailyItem(data.GetAWardID2(), data.GetCount2(), data.GetVIPBonus())
	// 3
	g.setADailyItem(data.GetAWardID3(), data.GetCount3(), data.GetVIPBonus())
	// 4
	g.setADailyItem(data.GetAWardID4(), data.GetCount4(), data.GetVIPBonus())
}

var randerGift = rand.New(rand.NewSource(4677))

func (g *giftDailyData) setADailyItem(itemId string, count uint32, vipBonus float32) {
	if IsFixedIDItemID(itemId) {
		g.Base.AddItem(itemId, count)
		g.Vip.AddItem(itemId, uint32(vipBonus*float32(count)))
		if vipBonus < 1.0 {
			logs.Error("giftDailyData Err vip_bonus < 1")
			return
		}
		g.VipAddon.AddItem(itemId, uint32((vipBonus-1)*float32(count)))
	} else {
		rewardData := MakeItemData("all", randerGift, itemId)
		if rewardData == nil {
			rewardData = &BagItemData{}
		}
		g.Base.AddItemWithData(itemId, *rewardData, count)
		g.Vip.AddItemWithData(itemId, *rewardData, uint32(vipBonus*float32(count)))
		if vipBonus < 1.0 {
			logs.Error("giftDailyData Err vip_bonus < 1")
			return
		}
		g.VipAddon.AddItemWithData(itemId, *rewardData, uint32((vipBonus-1)*float32(count)))
	}
}

type activityGiftData struct {
	Id        uint32
	TimeTyp   uint32
	GetTyp    uint32
	StartTime int64
	EndTime   int64
	Gift      []giftDailyData // 根据职业区分
}

func (a *activityGiftData) setGift(data *ProtobufGen.DAILYGIFT) {
	if a.Id != data.GetActivityID() {
		return
	}

	idx := int(data.GetActivityIndex())
	if idx <= 0 {
		logs.Error("GetActivityIndex is err by %v", data)
		return
	}

	// 只用第一个角色的奖励
	if data.GetRoleLimit() > 0 {
		return
	}

	for idx > len(a.Gift) {
		// 注意，这里要求a.Gift中存的是真正的奖励，也就是说len(a.Gift)代表奖励数
		a.Gift = append(a.Gift, giftDailyData{})
	}

	a.Gift[idx-1].setDailyItem(data)
}

var (
	gdActivityGiftData    map[uint32]*activityGiftData
	gdActivityGiftDataArr []*activityGiftData
)

func GetActivityGiftMapData(key uint32) (*activityGiftData, bool) {
	re, ok := gdActivityGiftData[key]
	return re, ok
}

func GetActivityGiftData() []*activityGiftData {
	return gdActivityGiftDataArr[:]
}

func loadGiftActivityData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ac_ar := &ProtobufGen.GIFTACTIVITYLIST_ARRAY{}
	err = proto.Unmarshal(buffer, ac_ar)
	errcheck(err)

	ac_data := ac_ar.GetItems()
	gdActivityGiftData = make(
		map[uint32]*activityGiftData,
		len(ac_data))
	gdActivityGiftDataArr = make(
		[]*activityGiftData, 0, len(ac_data))

	for _, c := range ac_data {
		d := &activityGiftData{}
		d.Id = c.GetActivityID()
		d.TimeTyp = c.GetTimeType()
		d.GetTyp = c.GetGiftAcceptType()
		d.StartTime = getDataTimeInData(c.GetStartTime())
		d.EndTime = getDataTimeInData(c.GetEndTime())
		// 注意，这里要求a.Gift中存的是真正的奖励，也就是说len(a.Gift)代表奖励数
		d.Gift = make([]giftDailyData, 0, 6)

		gdActivityGiftData[d.Id] = d
		gdActivityGiftDataArr = append(gdActivityGiftDataArr, d)
	}
}

func loadDailyGiftData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ac_ar := &ProtobufGen.DAILYGIFT_ARRAY{}
	err = proto.Unmarshal(buffer, ac_ar)
	errcheck(err)

	datas := ac_ar.GetItems()

	for _, c := range datas {
		for _, ac := range gdActivityGiftData {
			ac.setGift(c)
		}
	}

	//for k, ac := range gdActivityGiftData {
	//logs.Trace("gdActivityGiftData %d  ->  %v", k, ac)
	//}

}
