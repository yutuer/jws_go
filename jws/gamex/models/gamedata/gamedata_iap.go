package gamedata

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
)

const (
	IAP_Monthly = iota
	IAP_Life
	IAP_Week
)

type IAPInfo struct {
	Android_Rmb_Price uint32
	IOS_Rmb_Price     uint32
	Info              *ProtobufGen.IAPMAIN
}

var (
	gdAndroidIAPId2Rmb map[string]*ProtobufGen.IAPBASE
	gdIOSIAPId2Rmb     map[string]*ProtobufGen.IAPBASE
	gdIAPIndex2Info    map[uint32]*IAPInfo
	gdIAPId2Index      map[string]uint32
	IAPMonth           *ProtobufGen.DUBBLERAWARD
	IAPLife            *ProtobufGen.DUBBLERAWARD
	IAPWeek            *ProtobufGen.DUBBLERAWARD
	gdIAPCardIndexs    map[uint32]struct{}
	gdIAPConfig        *ProtobufGen.IAPCONFIG
)

func loadIAPConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.IAPCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdIAPConfig = ar.Items[0]
}

func loadIAPBaseData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.IAPBASE_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdAndroidIAPId2Rmb = make(map[string]*ProtobufGen.IAPBASE, len(data))
	gdIOSIAPId2Rmb = make(map[string]*ProtobufGen.IAPBASE, len(data))
	for _, v := range data {
		if v.GetPlatform() == 0 {
			gdIOSIAPId2Rmb[v.GetIapID()] = v
		} else {
			gdAndroidIAPId2Rmb[v.GetIapID()] = v
		}
	}
}

func loadIAPMainData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.IAPMAIN_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdIAPIndex2Info = make(map[uint32]*IAPInfo, len(data))
	gdIAPId2Index = make(map[string]uint32, len(data))
	for _, v := range data {
		ios_rmb, ios_ok := gdIOSIAPId2Rmb[v.GetIapID()]
		android_rmb, android_ok := gdAndroidIAPId2Rmb[v.GetIapID()]
		if !ios_ok && !android_ok {
			panic(fmt.Errorf("iap %s not found in iapmain table", v.GetIapID()))
		}
		info := IAPInfo{
			Android_Rmb_Price: android_rmb.GetPrice(),
			IOS_Rmb_Price:     ios_rmb.GetPrice(),
			Info:              v}
		gdIAPIndex2Info[v.GetIndex()] = &info
		gdIAPId2Index[v.GetIapID()] = v.GetIndex()
	}
}

func loadIAPCard(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.DUBBLERAWARD_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	gdIAPCardIndexs = make(map[uint32]struct{}, len(ar.GetItems()))
	for _, item := range ar.GetItems() {
		switch item.GetActivityID() {
		case IAP_Monthly:
			IAPMonth = item
		case IAP_Life:
			IAPLife = item
		case IAP_Week:
			IAPWeek = item
		default:
			panic(fmt.Errorf("loadIAPCard not know id %d", item.GetActivityID()))
		}
		idx := item.GetConditionValue1ForIOS()
		if idx > 0 {
			gdIAPCardIndexs[idx] = struct{}{}
		}
		idx = item.GetConditionValue2ForIOS()
		if idx > 0 {
			gdIAPCardIndexs[idx] = struct{}{}
		}
		idx = item.GetConditionValue1ForAndroid()
		if idx > 0 {
			gdIAPCardIndexs[idx] = struct{}{}
		}
		idx = item.GetConditionValue2ForAndroid()
		if idx > 0 {
			gdIAPCardIndexs[idx] = struct{}{}
		}
	}
}

func GetIAPInfo(idx uint32) *IAPInfo {
	return gdIAPIndex2Info[idx]
}

func GetIAPIdxByID(id string) uint32 {
	idx, ok := gdIAPId2Index[id]
	if !ok {
		return 0
	} else {
		return idx
	}
}

func GetIAPIOSPrice(iapId string) uint32 {
	price, ok := gdIOSIAPId2Rmb[iapId]
	if !ok {
		return 0
	} else {
		return price.GetPrice()
	}
}

func IsIAPCardIdx(idx uint32) bool {
	_, ok := gdIAPCardIndexs[idx]
	return ok
}

func GetIAPConfig() *ProtobufGen.IAPCONFIG {
	return gdIAPConfig
}

func GetIAPBaseConfig(iapID string) *ProtobufGen.IAPBASE {
	return gdAndroidIAPId2Rmb[iapID]
}

func GetIAPBaseConfigAndroid(iapID string) *ProtobufGen.IAPBASE {
	return gdAndroidIAPId2Rmb[iapID]
}

func GetIAPBaseConfigIOS(iapID string) *ProtobufGen.IAPBASE {
	return gdIOSIAPId2Rmb[iapID]
}

func GetPlatformByIdx(idx uint32) string {
	info, ok := gdIAPIndex2Info[idx]
	if !ok {
		return ""
	}
	_, ok = gdAndroidIAPId2Rmb[info.Info.GetIapID()]
	if ok {
		return uutil.Android_Platform
	}
	return uutil.IOS_Platform
}

func LoadIapMainConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.IAPMAIN_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdIAPId2Index = make(map[string]uint32, len(data))
	for _, v := range data {
		gdIAPId2Index[v.GetIapID()] = v.GetIndex()
	}
}

func GetIAPIdxMap() map[string]uint32 {
	return gdIAPId2Index
}
