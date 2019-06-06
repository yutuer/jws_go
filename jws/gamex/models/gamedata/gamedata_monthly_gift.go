package gamedata

import (
	"time"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const Mouth_Gift_Num_Max = 32
const Base_Gift_Idx = 0
const VIP_Addon_Gift_Idx = 0

type giftData struct {
	Base     CostData //现在就两种，一种是基础的，一种是VIP
	Vip      CostData
	VipAddon CostData //用于VIp补领
	VipNeed  uint32
}

func (g *giftData) setItem(
	item_id string,
	base uint32,
	vip_count uint32,
	vip_need uint32) {

	g.VipNeed = vip_need

	g.Base = CostData{}
	g.Vip = CostData{}
	g.VipAddon = CostData{}

	g.Base.AddItem(item_id, base)
	g.Vip.AddItem(item_id, vip_count)
	if vip_count < base {
		logs.Error("giftData Err vip_count <= base_count")
		return
	}
	g.VipAddon.AddItem(item_id, vip_count-base)
}

type monthlyGiftData struct {
	Id    uint32
	Typ   uint32
	Year  uint32
	Mouth uint32
	Gift  [Mouth_Gift_Num_Max]giftData
}

func (m *monthlyGiftData) setGift(data *ProtobufGen.MONTHLYGIFT) {
	if m.Id != data.GetActivityID() {
		return
	}

	day := int(data.GetActivityIndex())
	if day <= 0 || day >= Mouth_Gift_Num_Max {
		logs.Error("GetActivityIndex is err by %v", data)
		return
	}

	m.Gift[day-1].setItem(
		data.GetAWardID(),
		data.GetCount(),
		data.GetVIPCount(),
		data.GetVIPBase(),
	)
}

var (
	gdMonthlyGiftData map[uint32]*monthlyGiftData
)

func GetNowMonthlyGiftData(now_t int64) *monthlyGiftData {
	t := GetCommonDayBeginSec(now_t)
	now_time := time.Unix(t, 0).In(util.ServerTimeLocal)
	year, mouth, _ := now_time.Date()
	return getMonthlyGiftData(uint32(year), uint32(mouth))
}

// 获取每月签到中 今天是这个月的第几天
func GetNowMonthlyGiftDayth(now_t int64) int {
	t := GetCommonDayBeginSec(now_t)
	now_time := time.Unix(t, 0).In(util.ServerTimeLocal)
	_, _, day := now_time.Date()
	return day
}

// 这里的索引是 年份值*100 + 月份值
func getMonthlyGiftData(year, mouth uint32) *monthlyGiftData {
	idx := year*100 + mouth
	res, ok := gdMonthlyGiftData[idx]
	if !ok {
		return nil
	} else {
		return res
	}
}

func loadMonthlyActivityData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ac_ar := &ProtobufGen.MONTHLYACTIVITY_ARRAY{}
	err = proto.Unmarshal(buffer, ac_ar)
	errcheck(err)

	ac_data := ac_ar.GetItems()
	gdMonthlyGiftData = make(
		map[uint32]*monthlyGiftData,
		len(ac_data))

	for _, c := range ac_data {
		d := &monthlyGiftData{}
		d.Id = c.GetActivityID()
		d.Typ = c.GetGiftAcceptType()
		d.Year = c.GetYear()
		d.Mouth = c.GetMonth()

		gdMonthlyGiftData[d.Year*100+d.Mouth] = d

	}
}

func loadMonthlyGiftData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ac_ar := &ProtobufGen.MONTHLYGIFT_ARRAY{}
	err = proto.Unmarshal(buffer, ac_ar)
	errcheck(err)

	datas := ac_ar.GetItems()

	for _, c := range datas {
		for _, ac := range gdMonthlyGiftData {
			ac.setGift(c)
		}
	}

	for k, ac := range gdMonthlyGiftData {
		logs.Trace("MonthlyGiftData %d  ->  %v", k, ac)
	}

}
