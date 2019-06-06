package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

var (
	gdPackageGroup map[string]*ProtobufGen.PACKAGEGROUP
)

func loadPackageGroup(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.PACKAGEGROUP_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))
	gdPackageGroup = make(map[string]*ProtobufGen.PACKAGEGROUP, len(ar.Items))
	for _, e := range ar.GetItems() {
		gdPackageGroup[e.GetPackageGroupID()] = e
	}
}

func GetPackageGroup(itemId string) (bool, *ProtobufGen.PACKAGEGROUP) {
	if pkg, ok := gdPackageGroup[itemId]; ok {
		return true, pkg
	}
	return false, nil
}
