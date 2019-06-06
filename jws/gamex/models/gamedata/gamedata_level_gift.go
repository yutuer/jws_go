package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdLevel2Gift      map[uint32]*ProtobufGen.LEVELGIFTPURCHASE
	gdIap2GiftIOS     map[uint32]*ProtobufGen.LEVELGIFTPURCHASE
	gdIap2GiftAndroid map[uint32]*ProtobufGen.LEVELGIFTPURCHASE
	gdId2LevelGift    map[string]*ProtobufGen.LEVELGIFTPURCHASE
)

func loadLevelGift(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.LEVELGIFTPURCHASE_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	item_data := ar.GetItems()
	gdLevel2Gift = make(map[uint32]*ProtobufGen.LEVELGIFTPURCHASE, len(item_data))
	gdIap2GiftIOS = make(map[uint32]*ProtobufGen.LEVELGIFTPURCHASE, len(item_data))
	gdIap2GiftAndroid = make(map[uint32]*ProtobufGen.LEVELGIFTPURCHASE, len(item_data))
	gdId2LevelGift = make(map[string]*ProtobufGen.LEVELGIFTPURCHASE, len(item_data))
	for _, item := range item_data {
		gdId2LevelGift[item.GetLevelGiftID()] = item

		if _, ok := gdLevel2Gift[item.GetLevelRequired()]; ok {
			panic(fmt.Errorf("loadLevelGift gdLevel2Gift LevelRequired duplicate %s", item.GetLevelGiftID()))
		}
		gdLevel2Gift[item.GetLevelRequired()] = item

		if _, ok := gdIap2GiftIOS[item.GetIapIDforIOS()]; ok {
			panic(fmt.Errorf("loadLevelGift gdLevel2Gift IapIDforIOS duplicate %s %d", item.GetLevelGiftID(), item.GetIapIDforIOS()))
		}
		gdIap2GiftIOS[item.GetIapIDforIOS()] = item

		if _, ok := gdIap2GiftAndroid[item.GetIapIDforAndroid()]; ok {
			panic(fmt.Errorf("loadLevelGift gdLevel2Gift IapIDforAndroid duplicate %s %d", item.GetLevelGiftID(), item.GetIapIDforAndroid()))
		}
		gdIap2GiftAndroid[item.GetIapIDforAndroid()] = item
	}
}

func LevelGiftOnLvlUp(lvl uint32) (id string) {
	var maxLvl uint32
	for l, cfg := range gdLevel2Gift {
		if lvl >= l && maxLvl < l {
			maxLvl = l
			id = cfg.GetLevelGiftID()
		}
	}
	return
}

func GetLevelGiftCfg(id string) *ProtobufGen.LEVELGIFTPURCHASE {
	return gdId2LevelGift[id]
}

func IapLevelGiftIOS(goodIdx uint32) *ProtobufGen.LEVELGIFTPURCHASE {
	return gdIap2GiftIOS[goodIdx]
}

func IapLevelGiftAndroid(goodIdx uint32) *ProtobufGen.LEVELGIFTPURCHASE {
	return gdIap2GiftAndroid[goodIdx]
}
