package server

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/tools/comic_protogen/util"
)

var serverOutRootDir string = "./gen_server_temp"
var handlerDir string = serverOutRootDir + "/handlers"

/**

 */
type GenServerImpl struct {
}

func (gs *GenServerImpl) GenCode(protoInfos []*util.ProtoInfo) {
	mkOutDirs()
	reqProtos := make([]*util.ProtoInfo, 0)
	pushProtos := make([]*util.ProtoInfo, 0)
	for _, info := range protoInfos {
		if info.Type == util.MESSAGE_TYPE_REQ {
			reqProtos = append(reqProtos, info)
		} else if info.Type == util.MESSAGE_TYPE_PUSH {
			pushProtos = append(pushProtos, info)
		}
	}
	GenReqRspMessage(reqProtos)
	logs.Debug("pushProtos %d", len(pushProtos))
	GenPushMessage(pushProtos)
	genRegisterHandlerCode(protoInfos)
	genPlayerMsgCode(pushProtos)
}

func mkOutDirs() {
	util.MkDir(serverOutRootDir)
	util.MkDir(handlerDir)
}

func genRegisterHandlerCode(protoInfos []*util.ProtoInfo) {
	buffer := bytes.NewBuffer([]byte{})

	reqFuncs := ""
	pushFuncs := ""
	for _, info := range protoInfos {
		if info.Type == util.MESSAGE_TYPE_REQ {
			reqFuncs += fmt.Sprintf(registerReqFunc, info.Name, info.Name)
		} else {
			pushFuncs += fmt.Sprintf(registerPushFunc, info.Name, info.Name)
		}
	}

	buffer.WriteString(fmt.Sprintf(registerReqCode, reqFuncs))
	buffer.WriteString(fmt.Sprintf(registerPushCode, pushFuncs))

	writeFile := serverOutRootDir + "/gen_reg_func.go"
	ioutil.WriteFile(writeFile, buffer.Bytes(), 0666)
}

func genPlayerMsgCode(protoInfos []*util.ProtoInfo) {
	buffer := bytes.NewBuffer([]byte{})

	codes := ""
	for _, info := range protoInfos {
		codes += fmt.Sprintf(playerMsgOne, info.Name, info.Name)
	}

	buffer.WriteString(fmt.Sprintf(playerMsgCode, codes))

	writeFile := serverOutRootDir + "/msg.go"
	ioutil.WriteFile(writeFile, buffer.Bytes(), 0666)
}
