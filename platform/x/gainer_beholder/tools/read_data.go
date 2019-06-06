package tools

import (
	"path/filepath"

	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/gamedata_load"
)

var Hot_package map[IDS]*ProtobufGen.HOTPACKAGE

type IDS struct {
	ID    uint32
	SubID uint32
}

func init() {
	//加载data文件
	dataRelPath := "data"
	dataAbsPath := filepath.Join(gamedata_load.GetMyDataPath(), dataRelPath)

	hotpackage := &ProtobufGen.HOTPACKAGE_ARRAY{}
	gamedata_load.Common_load(filepath.Join("", dataAbsPath, "hotpackage.data"), hotpackage)

	Hot_package = make(map[IDS]*ProtobufGen.HOTPACKAGE, len(hotpackage.Items))
	for _, value := range hotpackage.Items {
		key := IDS{ID: value.GetHotPackageID(), SubID: value.GetHotPackageSubID()}
		Hot_package[key] = value
	}
}
