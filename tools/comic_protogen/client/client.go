package client

import "vcs.taiyouxi.net/tools/comic_protogen/util"

var clientOutRootDir string = "./gen_client_temp"

type GenClientImpl struct {
}

func (gc *GenClientImpl) GenCode(infos []*util.ProtoInfo) {
	util.MkDir(clientOutRootDir)

	reqProtos := make([]*util.ProtoInfo, 0)
	pushProtos := make([]*util.ProtoInfo, 0)
	for _, info := range infos {
		if info.Type == util.MESSAGE_TYPE_REQ {
			reqProtos = append(reqProtos, info)
		} else if info.Type == util.MESSAGE_TYPE_PUSH {
			pushProtos = append(pushProtos, info)
		}
	}
	GenReqRspFile(reqProtos)
	GenPushFile(pushProtos)
}
