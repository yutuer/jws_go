package gamedata

//import (
//	"github.com/golang/protobuf/proto"
//	"vcs.taiyouxi.net/jws/gamex/protogen"
//)
//
//var (
//	gdMoneyCatData []*ProtobufGen.MONEYGOD
//)
//
//func loadMoneyCatData(filepath string) {
//	buffer, err := loadBin(filepath)
//	panicIfErr(err)
//
//	ar := new(ProtobufGen.MONEYGOD_ARRAY)
//	panicIfErr(proto.Unmarshal(buffer, ar))
//
//	gdMoneyCatData = ar.GetItems()
//}
//
//func GetMoneyCatNum(step int64) (minnum, maxnum int64) {
//	return int64(gdMoneyCatData[step].GetMinNum()),
//		int64(gdMoneyCatData[step].GetMaxNum())
//
//}
//
//func GetMoneyCatCost(step int64) uint32  {
//	return uint32(gdMoneyCatData[step].GetCostHC())
//
//}
