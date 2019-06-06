package gamedata

import (
	"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	MarketActivityMailTitle = map[int]int{
		ActPayPreDay: mail_sender.IDS_MAIL_ACTIVITY_ACCPAYDAY_TITLE,
		ActLogin:     mail_sender.IDS_MAIL_ACTIVITY_ACCLOGIN_TITLE,
		ActPay:       mail_sender.IDS_MAIL_ACTIVITY_ACCPAYSUM_TITLE,
		ActHcCost:    mail_sender.IDS_MAIL_ACTIVITY_ACCCONSUME_TITLE,
		ActGameMode:  mail_sender.IDS_MAIL_ACTIVITY_ACCLEVEL_TITLE,
		ActBuy:       mail_sender.IDS_MAIL_ACTIVITY_ACCRESOURCE_TITLE,
		ActHeroStar:  mail_sender.IDS_MAIL_ACTIVITY_HERO_STAR_TITLE,
		ActDayPay:    mail_sender.IDS_MAIL_ACTIVITY_ACCPAYDAYSUM_TITLE,
		ActDayHcCost: mail_sender.IDS_MAIL_ACTIVITY_ACCDAYCONSUME_TITLE,
		ActOnlyPay:   mail_sender.IDS_MAIL_SINGLEPAY_CONTENT,
	}
)

type MarketSubActivityConfig2Client struct {
	FCCondTyp   uint32 `codec:"cond"`
	FCParam1    uint32 `codec:"p1"`
	FCParam2    uint32 `codec:"p2"`
	FCParam3    string `codec:"p3"`
	FCParam4    string `codec:"p4"`
	RewardCount int    `codec:"rc"`
}
type MarketSubActivityConfigReward2Client struct {
	ItemId    string `codec:"item"`
	ItemCount uint32 `codec:"c"`
}

type hotMarketActivity struct {
}

func (act *hotMarketActivity) loadData(buffer []byte, datas *HotDatas) error {
	dataList := &ProtobufGen.HOTACTIVITYDETAIL_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}

	datas.Activity.marketActivity = make(map[uint32]map[uint32]*ProtobufGen.HOTACTIVITYDETAIL, len(dataList.GetItems()))
	datas.HeroFoundConfig.ActivityIap = make(map[int]bool, max_hero_found_iap)
	for _, data := range dataList.GetItems() {
		v, ok := datas.Activity.marketActivity[data.GetActivityID()]
		if !ok {
			v = make(map[uint32]*ProtobufGen.HOTACTIVITYDETAIL, 8)
			datas.Activity.marketActivity[data.GetActivityID()] = v	
		}
		v[data.GetActivitySubID()] = data

		// 设置投资英雄活动的充值ID
		if isHeroFoundActivity(int(data.GetFCType())) && data.GetActivitySubID() == 1 {
			if data.GetSFCValue1() != "" {
				temp := strings.Split(data.GetSFCValue1(), ",")
				for _, temp_IapID := range temp {
					if iapId, err := strconv.ParseFloat(temp_IapID, 32); err == nil {
						datas.HeroFoundConfig.ActivityIap[int(iapId)] = true
					} else {
						return err
					}
				}
			}
			if data.GetSFCValue2() != "" {
				temp := strings.Split(data.GetSFCValue2(), ",")
				for _, temp_IapID := range temp {
					if iapId, err := strconv.ParseFloat(temp_IapID, 32); err == nil {
						datas.HeroFoundConfig.ActivityIap[int(iapId)] = true
					} else {
						return err
					}
				}
			}
		}
	}
	logs.Debug("init hero found iap: %v", datas.HeroFoundConfig)
	return nil
}

func (d hotActivityData) GetMarketActivitySubConfig(activityId uint32) map[uint32]*ProtobufGen.HOTACTIVITYDETAIL {
	if _, ok := d.marketActivity[activityId]; ok {
		//存在
		return d.marketActivity[activityId]
	}
	return nil
}

func GetMarketActivityMailType(activityType int) int {
	if activityType >= ActHeroFund_Begin && activityType <= ActHeroFund_End {
		return mail_sender.IDS_MAIL_HERO_FUND_ON_BANLANCE
	}
	return MarketActivityMailTitle[activityType]
}
