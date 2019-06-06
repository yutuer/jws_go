package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//
// 配方数据
//

var (
	gdComposeNeedData []CostData
	gdComposeGiveData []CostData
)

func GetFormulaData(fid int) (ok bool, need *CostData, give *CostData) {
	ok = false
	need = nil
	give = nil

	// 注意 索引id从1开始，0位是空值
	if fid > 0 &&
		fid < len(gdComposeNeedData) &&
		fid < len(gdComposeGiveData) {
		ok = true
		need = &gdComposeNeedData[fid]
		give = &gdComposeGiveData[fid]
	}

	return
}

func loadComposeCofig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.FORMULA_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	gdComposeNeedData = make([]CostData, 1, len(dataList.GetItems())+1) // 表中索引从1开始
	gdComposeGiveData = make([]CostData, 1, len(dataList.GetItems())+1) // 表中索引从1开始

	for _, a := range dataList.GetItems() {
		fid := a.GetFormulaIndex()
		//logs.Trace("formula %s  ->  %v", fid, a.GetTargetID())

		//TODO 优化内存分配
		need := CostData{}
		give := CostData{}

		need_data := a.GetFormulaDetail()
		for i := 0; i < len(need_data); i++ {
			need_one_data := need_data[i]
			if need_one_data == nil {
				logs.Error("need_one_data nil %d : %d", fid, i)
				continue
			}
			need.AddItem(
				need_one_data.GetMaterial(),
				need_one_data.GetMcount())
		}

		give.AddItem(a.GetTargetID(), a.GetTargetCount())

		gdComposeNeedData = append(gdComposeNeedData, need)
		gdComposeGiveData = append(gdComposeGiveData, give)
		//logs.Trace("AllNeedData %d : %v", fid, need)
		//logs.Trace("AllGiveData %d : %v", fid, give)
	}

}
