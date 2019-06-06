package client

import (
	"fmt"
	"strings"

	"vcs.taiyouxi.net/tools/comic_protogen/util"
)

func GenReqRspFile(infos []*util.ProtoInfo) {
	for _, info := range infos {
		genOneClientFile(info)
	}
}

func genOneClientFile(info *util.ProtoInfo) {
	fileUtil := util.NewGenFileUtil(fmt.Sprintf("%s/%s.cs", clientOutRootDir, info.Name))
	fileUtil.WriteString(clientHeader)
	fileUtil.WriteString(genClientReqRespCode(info.Name))
	fileUtil.WriteString(genClientHandlerCode(info.Name, info.ClientReqParams))
	fileUtil.Flush()
}

func genClientReqRespCode(protoName string) string {
	params := make([]interface{}, 7)
	for i := 0; i < 7; i++ {
		params[i] = protoName
	}
	return fmt.Sprintf(clientReqResp, params...)
}

func genClientHandlerCode(protoName string, reqParams []util.ClientReqParam) string {
	params := make([]interface{}, 11)
	for i := 0; i < 11; i++ {
		params[i] = protoName
	}

	funcParmas := ""
	funcBody := ""
	for _, param := range reqParams {
		if param.ParamType != "protogen.Req" {
			funcParmas = fmt.Sprintf("%s,%s %s", funcParmas, param.ParamType, param.ParamVar)
			funcBody += "\n"
			if strings.Contains(param.ParamType, "[]") {
				funcBody += fmt.Sprintf("		msg.req.%s.AddRange(%s);", param.ParamVar, param.ParamVar)
			} else {
				funcBody += fmt.Sprintf("		msg.req.%s=%s;", param.ParamVar, param.ParamVar)
			}
		}
	}
	params[6] = funcParmas
	params[10] = funcBody
	return fmt.Sprintf(clientHanlder, params...)
}
