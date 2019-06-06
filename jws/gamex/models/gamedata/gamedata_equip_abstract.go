package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdEquipTrickData map[string]*ProtobufGen.EUIPTRICKDETAIL
)

func loadEquipTrickConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.EUIPTRICKDETAIL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	item_data := ar.GetItems()

	gdEquipTrickData = make(map[string]*ProtobufGen.EUIPTRICKDETAIL, len(item_data))
	gdTrickDetailAttrAddon = make(map[string]*avatarAttrAddon, len(gdEquipTrickData))

	for _, a := range item_data {
		gdEquipTrickData[a.GetTrickID()] = a
		attr := &avatarAttrAddon{}
		attr.AddTrickAddon(a)
		gdTrickDetailAttrAddon[a.GetTrickID()] = attr
		//logs.Trace("gdEquipTrickData %v", a)
		//logs.Trace("gdTrickDetailAttrAddon %v", attr)
	}

	//logs.Trace("gdEquipTrickData %v", gdEquipTrickData)

}

func GetEquipTrickData(tid string) *ProtobufGen.EUIPTRICKDETAIL {
	res, ok := gdEquipTrickData[tid]
	if !ok {
		return nil
	} else {
		return res
	}
}

var (
	gdEquipTrickSettingData []*ProtobufGen.EUIPTRICKSETTINGS
)

func loadEquipTrickSettingConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.EUIPTRICKSETTINGS_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	item_data := ar.GetItems()

	gdEquipTrickSettingData = make([]*ProtobufGen.EUIPTRICKSETTINGS,
		0,
		len(item_data)+1)

	for _, a := range item_data {
		tier := int(a.GetTier())
		for len(gdEquipTrickSettingData) <= tier {
			gdEquipTrickSettingData = append(gdEquipTrickSettingData, nil)
		}
		gdEquipTrickSettingData[tier] = a
	}

	//logs.Trace("gdEquipTrickSettingData %v", gdEquipTrickSettingData)

}

func GetEquipTrickSettingData(tier int) *ProtobufGen.EUIPTRICKSETTINGS {
	if tier < 0 || tier >= len(gdEquipTrickSettingData) {
		return nil
	} else {
		return gdEquipTrickSettingData[tier]
	}

}
