package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdPrivilegeData map[int32]*ProtobufGen.PRIVILEGE
)

func loadPrivilegeData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.PRIVILEGE_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdPrivilegeData = make(map[int32]*ProtobufGen.PRIVILEGE, len(data))

	for _, c := range data {
		gdPrivilegeData[c.GetPrivilegeID()] = c
	}
}

func GetPrivilegeInfo(id int32) (times uint32, price uint32, found bool) {
	if prvl, ok := gdPrivilegeData[id]; ok {
		return prvl.GetPrivilegeNumber(), prvl.GetNewPrice(), true
	}
	return 0, 0, false
}

type AwardInfo struct {
	ItemId string
	Count  uint32
}

func GetPrivilegeAward(id int32) []AwardInfo {
	res := []AwardInfo{}
	if prvl, ok := gdPrivilegeData[id]; ok {
		// 1
		if len(prvl.GetItem1()) > 0 && prvl.GetCount1() > 0 {
			res = append(res, AwardInfo{
				prvl.GetItem1(),
				prvl.GetCount1(),
			})
		}
		// 2
		if len(prvl.GetItem2()) > 0 && prvl.GetCount2() > 0 {
			res = append(res, AwardInfo{
				prvl.GetItem2(),
				prvl.GetCount2(),
			})
		}
		// 3
		if len(prvl.GetItem3()) > 0 && prvl.GetCount3() > 0 {
			res = append(res, AwardInfo{
				prvl.GetItem3(),
				prvl.GetCount3(),
			})
		}
		// 4
		if len(prvl.GetItem4()) > 0 && prvl.GetCount4() > 0 {
			res = append(res, AwardInfo{
				prvl.GetItem4(),
				prvl.GetCount4(),
			})
		}
		// 5
		if len(prvl.GetItem5()) > 0 && prvl.GetCount5() > 0 {
			res = append(res, AwardInfo{
				prvl.GetItem5(),
				prvl.GetCount5(),
			})
		}
		// 6
		if len(prvl.GetItem6()) > 0 && prvl.GetCount6() > 0 {
			res = append(res, AwardInfo{
				prvl.GetItem6(),
				prvl.GetCount6(),
			})
		}
	}
	return res
}

func GetPrivilegeCfgCount() int {
	return len(gdPrivilegeData)
}
