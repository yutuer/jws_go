package server

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"vcs.taiyouxi.net/tools/comic_protogen/util"
)

func GenReqRspMessage(protoInfos []*util.ProtoInfo) {
	// 生成消息处理文件
	genReqRspCode(protoInfos)
	// 生成handler文件
	genImplementHandlerCode(protoInfos)
	// 生成注册handler文件
	genRegisterHandlerCode(protoInfos)
}

func genReqRspCode(protoInfos []*util.ProtoInfo) {
	buffer := bytes.NewBuffer([]byte{})

	// 头文件
	buffer.WriteString(genReqRspHeader(protoInfos))

	// 内容
	for _, info := range protoInfos {
		buffer.WriteString("\n")
		buffer.WriteString(genOneReqRspCode(info.Name, info.Dir))
	}

	// 写文件
	ioutil.WriteFile(serverOutRootDir+"/gen_req_rsp_func.go", buffer.Bytes(), 0666)
}

func genOneReqRspCode(protoName, dirName string) string {
	names := make([]interface{}, 6)
	for i := 0; i < len(names); i++ {
		names[i] = protoName
	}
	names[4] = dirName
	return fmt.Sprintf(reqRspFunc, names...)
}

func genReqRspHeader(protoInfos []*util.ProtoInfo) string {
	retString := reqRspHeader
	hasWriteMap := make(map[string]struct{})
	for _, info := range protoInfos {
		if info.Dir != "" {
			if _, ok1 := hasWriteMap[info.Dir]; !ok1 {
				retString = retString + fmt.Sprintf(`	"vcs.taiyouxi.net/comic/gamex/logics/handlers/%s"`, info.Dir)
				retString += "\n"
				hasWriteMap[info.Dir] = struct{}{}
			}
		}
	}
	retString += ")"
	return retString
}

func genImplementHandlerCode(protoInfos []*util.ProtoInfo) {
	for _, info := range protoInfos {
		dirName := info.Dir
		genOneImpHandlerCode(info.Name, dirName)
	}
}

func genOneImpHandlerCode(protoName, dirName string) {
	fileDir := fmt.Sprintf("%s/%s", handlerDir, dirName)
	util.MkDir(fileDir)
	buffer := bytes.NewBuffer([]byte{})
	content := fmt.Sprintf(implementCode, dirName, protoName, protoName, protoName)
	buffer.WriteString(content)
	fileName := fmt.Sprintf("%s/%s.go", fileDir, protoName)
	ioutil.WriteFile(fileName, buffer.Bytes(), 0666)
}
