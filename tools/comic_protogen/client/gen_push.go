package client

import (
	"fmt"

	"vcs.taiyouxi.net/tools/comic_protogen/util"
)

func GenPushFile(infos []*util.ProtoInfo) {
	for _, info := range infos {
		genOnePushFile(info)
	}
}

func genOnePushFile(info *util.ProtoInfo) {
	fileUtil := util.NewGenFileUtil(fmt.Sprintf("%s/%s.cs", clientOutRootDir, info.Name))
	params := make([]interface{}, 16)
	for i := range params {
		params[i] = info.Name
	}
	fileUtil.WriteString(fmt.Sprintf(clientPushCode, params...))
	fileUtil.Flush()
}