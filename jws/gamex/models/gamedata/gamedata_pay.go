package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

type payData struct {
	ID                   uint32
	SID                  string
	RMB                  uint32
	RMBPoint             uint32
	HC                   uint32
	HCGive               uint32
	CommonDES            string
	CardPeriod           uint32
	CardDes              string
	FirstGiveID          uint32
	Icon                 string
	FirstGiveHC          uint32
	FirstGiveDescription string
}

var (
	gdPayData      map[string]*payData
	gdPayFirstGive []*ProtobufGen.FIRSTGIVE
)

func GetPayData(good_id string) *payData {
	res, ok := gdPayData[good_id]
	if !ok {
		return nil
	} else {
		return res
	}
}

func GetPayDatas() map[string]*payData {
	return gdPayData
}

func loadPayFirstGiveCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.FIRSTGIVE_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	data_len := len(dataList.GetItems()) + 1

	gdPayFirstGive = make([]*ProtobufGen.FIRSTGIVE, data_len, data_len)

	for _, a := range dataList.GetItems() {
		gdPayFirstGive[int(a.GetFirstGiveID())] = a
	}

}

func loadPayCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.PAY_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdPayData = make(map[string]*payData, len(dataList.GetItems()))

	for _, a := range dataList.GetItems() {
		d := payData{
			ID:          a.GetIID(),
			SID:         a.GetSID(),
			RMB:         a.GetRMB(),
			RMBPoint:    a.GetRMBPoint(),
			HC:          a.GetHC(),
			HCGive:      a.GetHCGive(),
			CommonDES:   a.GetCommonDES(),
			CardPeriod:  a.GetCardPeriod(),
			CardDes:     a.GetCardDes(),
			FirstGiveID: a.GetFirstGiveID(),
			Icon:        a.GetIcon(),
		}
		firid := int(a.GetFirstGiveID())
		if firid != 0 {
			first_data := gdPayFirstGive[firid]
			if first_data != nil {
				d.FirstGiveHC = first_data.GetFirstGiveHC()
				d.FirstGiveDescription = first_data.GetDescription()
			}
		}
		//logs.Trace("par data one : %v", d)
		gdPayData[d.SID] = &d
	}

	//logs.Trace("pay data %v", gdPayData)
}

func LoadPayConfigForOther(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.PAY_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdPayData = make(map[string]*payData, len(dataList.GetItems()))

	for _, a := range dataList.GetItems() {
		d := payData{
			ID:          a.GetIID(),
			SID:         a.GetSID(),
			RMB:         a.GetRMB(),
			RMBPoint:    a.GetRMBPoint(),
			HC:          a.GetHC(),
			HCGive:      a.GetHCGive(),
			CommonDES:   a.GetCommonDES(),
			CardPeriod:  a.GetCardPeriod(),
			CardDes:     a.GetCardDes(),
			FirstGiveID: a.GetFirstGiveID(),
			Icon:        a.GetIcon(),
		}
		gdPayData[d.SID] = &d
	}
}
