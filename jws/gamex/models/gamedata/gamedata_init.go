package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdAvatarInitData    []*ProtobufGen.INITIALIZESAVE
	gdAvatarInitBag     []*ProtobufGen.INITIALIZEITEM
	gdAvatarInitFashion map[int]*ProtobufGen.INITIALIZEFASHION
)

func loadAvatarInitConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.INITIALIZESAVE_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	// 出错就直接崩溃吧
	gdAvatarInitData = lv_data

	//logs.Trace("gdAvatarInitData %v", gdAvatarInitData)

}

func loadAvatarInitBagConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.INITIALIZEITEM_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()
	gdAvatarInitBag = lv_data
}

func loadAvatarInitFashion(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.INITIALIZEFASHION_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	gdAvatarInitFashion = make(map[int]*ProtobufGen.INITIALIZEFASHION, len(lv_ar.GetItems()))
	for _, item := range lv_ar.GetItems() {
		gdAvatarInitFashion[int(item.GetRole())] = item
	}
}

func GetAvatarInitData() []*ProtobufGen.INITIALIZESAVE {
	return gdAvatarInitData[:]
}

func GetAvatarInitBagData() []*ProtobufGen.INITIALIZEITEM {
	return gdAvatarInitBag[:]
}

func GetAvatarInitFashionData(avatar int) *ProtobufGen.INITIALIZEFASHION {
	return gdAvatarInitFashion[avatar]
}
