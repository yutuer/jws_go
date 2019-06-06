package gamedata

import (
	"fmt"

	"time"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	MaxShopNum             = 4
	LimitCountRefTyp_Daily = "Daily"
)

var (
	gdShop2Goods map[uint32]map[string]string
	gdGoods      map[string]*ProtobufGen.SHOPGOODS
)

func loadShop(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	s := &ProtobufGen.SHOPDISPLAY_ARRAY{}
	err = proto.Unmarshal(buffer, s)
	errcheck(err)

	gdShop2Goods = make(map[uint32]map[string]string, 10)
	for _, shop := range s.GetItems() {
		gs, ok := gdShop2Goods[shop.GetShopType()]
		if !ok {
			gdShop2Goods[shop.GetShopType()] = make(map[string]string, 10)
		}
		gs = gdShop2Goods[shop.GetShopType()]
		gs[shop.GetGoodsID()] = ""
		if _, ok := gdGoods[shop.GetGoodsID()]; !ok {
			panic(fmt.Errorf("shop %d good %s not define", shop.GetShopType(), shop.GetGoodsID()))
		}
	}
}

func loadShopGood(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	sg := &ProtobufGen.SHOPGOODS_ARRAY{}
	err = proto.Unmarshal(buffer, sg)
	errcheck(err)

	gdGoods = make(map[string]*ProtobufGen.SHOPGOODS, len(sg.GetItems()))
	for _, good := range sg.GetItems() {
		if _, ok := gdGoods[good.GetGoodsID()]; ok {
			panic(fmt.Errorf("shop good repeat (SHOPGOODS): %s", good.GetGoodsID()))
		}
		gdGoods[good.GetGoodsID()] = good
	}
}

func GetShopGoodCfg(shop uint32, good string) *ProtobufGen.SHOPGOODS {
	goods, ok := gdShop2Goods[shop]
	if ok {
		_, ok := goods[good]
		if ok {
			goodCfg, ok := gdGoods[good]
			if !ok {
				logs.Error("[GetShopGoodCfg] shop good in SHOPDISPLAY but not in SHOPGOODS, shop[%d] good[%s]", shop, good)
				return nil
			}
			return goodCfg
		}
	}
	return nil
}

func GetGoodCfg(good string) *ProtobufGen.SHOPGOODS {
	goodCfg, ok := gdGoods[good]
	if !ok {
		logs.Error("[GetGoodCfg] good not exit in SHOPGOODS, good[%s]", good)
		return nil
	}
	return goodCfg
}

func IsGoodInShop(shop uint32, good string) bool {
	goods, ok := gdShop2Goods[shop]
	if ok {
		_, ok := goods[good]
		return ok
	}
	return false
}

func IsGoodDailyRefresh(goodCfg *ProtobufGen.SHOPGOODS) bool {
	return goodCfg.GetLimitType() == "Daily"
}

func IsGoodTimeValid(goodCfg *ProtobufGen.SHOPGOODS, now_time int64) (bool, error) {
	if goodCfg.GetStartTime() != "" {
		start_time, err := time.ParseInLocation("2006/1/2 15:04", goodCfg.GetStartTime()+" 00:00", util.ServerTimeLocal)
		if err != nil {
			return false, err
		}
		if now_time < start_time.Unix() {
			return false, nil
		}
	}
	if goodCfg.GetEndTime() != "" {
		end_time, err := time.ParseInLocation("2006/1/2 15:04", goodCfg.GetEndTime()+" 00:00", util.ServerTimeLocal)
		if err != nil {
			return false, err
		}
		if now_time > end_time.Unix() {
			return false, nil
		}
	}
	return true, nil
}
