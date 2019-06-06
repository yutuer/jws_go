package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdEatBaoziData []*ProtobufGen.EATBAOZI
)

func loadEatBaoziData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.EATBAOZI_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	gdEatBaoziData = ar.GetItems()
}

func GetEatBaoziData() (itemID string, num, total, levelOpen uint32) {
	return gdEatBaoziData[0].GetItemID(),
		gdEatBaoziData[0].GetNumber(),
		gdEatBaoziData[0].GetTotal(),
		gdEatBaoziData[0].GetLevelOpen()
}

func mkEatBaoziDatas(loadFunc func(dfilepath string, loadfunc func(string))) {
	loadFunc("eatbaozi.data", loadEatBaoziData)

	logs.Trace("eatBaozi %v", *(gdEatBaoziData[0]))
}
