package server

import (
	"bytes"
	"io/ioutil"

	"fmt"

	"vcs.taiyouxi.net/tools/comic_protogen/util"
)

func GenPushMessage(protoInfos []*util.ProtoInfo) {
	genPushCode(protoInfos)
	genPushBuildCode(protoInfos)
}

func genPushCode(protoInfos []*util.ProtoInfo) {
	buffer := bytes.NewBuffer([]byte{})

	// 头文件
	buffer.WriteString(pushCodeHeader)

	// 内容
	for _, info := range protoInfos {
		buffer.WriteString("\n")
		buffer.WriteString(genCodeFunc(info))
	}

	// 写文件
	ioutil.WriteFile(serverOutRootDir+"/gen_push_func.go", buffer.Bytes(), 0666)
}

func genCodeFunc(info *util.ProtoInfo) string {
	params := make([]interface{}, 6)
	for i := range params {
		params[i] = info.Name
	}
	return fmt.Sprintf(pushCodeFunc, params...)
}

func genPushBuildCode(protoInfos []*util.ProtoInfo) {
	fileDir := fmt.Sprintf("%s/push", handlerDir)
	util.MkDir(fileDir)
	for _, info := range protoInfos {
		genOneBuildPushCode(info.Name, fileDir)
	}
}

func genOneBuildPushCode(protoName, fileDir string) {
	buffer := bytes.NewBuffer([]byte{})
	content := fmt.Sprintf(pushCodeBuild, protoName, protoName, protoName)
	buffer.WriteString(content)
	fileName := fmt.Sprintf("%s/%s.go", fileDir, protoName)
	ioutil.WriteFile(fileName, buffer.Bytes(), 0666)
}
