package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//默认的每个区间的长度
const (
	aptitudelength = 10
	DefaultProp    = 0
	DefaultType    = 0
)

var (
	magicPetConfig *ProtobufGen.MAGICPETCONFIG
	//第一个special／normal 表示是否使用特殊道具，
	//第二个special／normal表示是否是特殊区间.
	PetAptitudeSpecialSpecial []*ProtobufGen.PETAPTITUDE
	PetAptitudeSpecialNormal  []*ProtobufGen.PETAPTITUDE
	SpecialNormalWeight       uint32
	PetAptitudeNormalSpecial  []*ProtobufGen.PETAPTITUDE
	PetAptitudeNormalNormal   []*ProtobufGen.PETAPTITUDE
	NormalNormalWeight        uint32
	petLevel                  map[uint32]*ProtobufGen.PETLEVEL
	petStar                   map[uint32]*ProtobufGen.PETSTAR
	typeAptitude              map[uint32]*ProtobufGen.TYPEAPTITUDE
	maxStar                   uint32
)

var errCheck = func(err error) {
	if err != nil {
		panic(err)
	}
}

func loadMagicPetConfig(filepath string) {
	buffer, err := loadBin(filepath)
	errCheck(err)
	dataList := &ProtobufGen.MAGICPETCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)
	magicPetConfig = dataList.GetItems()[0]
}

func loadPetAptitude(filepath string) {
	//加载时，将使用道具的区间和不使用道具的区间分开存储
	//普通区间和特殊区间分开存储
	PetAptitudeSpecialSpecial = make([]*ProtobufGen.PETAPTITUDE, 0, aptitudelength)
	PetAptitudeSpecialNormal = make([]*ProtobufGen.PETAPTITUDE, 0, aptitudelength)
	PetAptitudeNormalSpecial = make([]*ProtobufGen.PETAPTITUDE, 0, aptitudelength)
	PetAptitudeNormalNormal = make([]*ProtobufGen.PETAPTITUDE, 0, aptitudelength)
	buffer, err := loadBin(filepath)
	errCheck(err)

	dataList := &ProtobufGen.PETAPTITUDE_ARRAY{}

	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	for i, data := range dataList.GetItems() {
		if data.GetNForSpecial() >= data.GetMForSpecial() && data.GetNForSpecial() != 0 {
			logs.Error("magicpetAptitude the %d data error,N>=M", i)
			panic("magicpetAptitude N/M,N>=M")
		}
		//data.AptitudeProp为0表示不使用道具，为1表示使用道具
		//data.AptitudeType为0表示普通区间，为1表示特殊区间
		if data.GetAptitudeProp() == DefaultProp {
			if data.GetAptitudeType() == DefaultType {
				NormalNormalWeight += data.GetIntervalWeight()
				PetAptitudeNormalNormal = append(PetAptitudeNormalNormal, data)
			} else {
				PetAptitudeNormalSpecial = append(PetAptitudeNormalSpecial, data)
			}
		} else {
			if data.GetAptitudeType() == DefaultType {
				SpecialNormalWeight += data.GetIntervalWeight()
				PetAptitudeSpecialNormal = append(PetAptitudeSpecialNormal, data)
			} else {
				PetAptitudeSpecialSpecial = append(PetAptitudeSpecialSpecial, data)
			}
		}
	}
}

func loadPetLevel(filepath string) {
	petLevel = make(map[uint32]*ProtobufGen.PETLEVEL)
	buffer, err := loadBin(filepath)
	errCheck(err)

	dataList := &ProtobufGen.PETLEVEL_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	for _, data := range dataList.GetItems() {
		petLevel[data.GetPetlevel()] = data

	}
}

func loadPetStar(filepath string) {
	petStar = make(map[uint32]*ProtobufGen.PETSTAR)

	buffer, err := loadBin(filepath)
	errCheck(err)

	dataList := &ProtobufGen.PETSTAR_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	for _, data := range dataList.GetItems() {
		petStar[data.GetStarID()] = data
		if maxStar < data.GetStarID() {
			maxStar = data.GetStarID()
		}
	}
}

func loadTypeAptitude(filepath string) {
	typeAptitude = make(map[uint32]*ProtobufGen.TYPEAPTITUDE)

	buffer, err := loadBin(filepath)
	errCheck(err)

	dataList := &ProtobufGen.TYPEAPTITUDE_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errCheck(err)

	for _, data := range dataList.GetItems() {
		typeAptitude[data.GetExistTimes()] = data
	}
}
func GetMagicPetLvInfo(lv uint32) *ProtobufGen.PETLEVEL {
	return petLevel[lv]
}
func GetMagicPetStarInfo(star uint32) *ProtobufGen.PETSTAR {
	return petStar[star]
}
func GetMagicPetConfig() *ProtobufGen.MAGICPETCONFIG {
	return magicPetConfig
}
func GetStar(starID uint32) *ProtobufGen.PETSTAR {
	if _, ok := petStar[starID]; ok {
		return petStar[starID]
	}
	return nil
}
func GetTypeAptitude(ExistTimes uint32) *ProtobufGen.TYPEAPTITUDE {
	if _, ok := typeAptitude[ExistTimes]; ok {
		return typeAptitude[ExistTimes]
	}
	return nil
}
func GetMaxStar() uint32 {
	return maxStar
}
