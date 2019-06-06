package util

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func ParseAllProtos(path string) []*ProtoInfo {
	protoInfos := make([]*ProtoInfo, 0)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		logs.Debug("<walk dir> %s", path)
		infos := parseProtosFromFile(path)
		protoInfos = append(protoInfos, infos...)
		return nil
	})
	if err != nil {
		logs.Error("walk dir err ", err)
	}

	return mergeProtos(protoInfos)
}

func parseProtosFromFile(path string) []*ProtoInfo {
	pathsByDot := strings.Split(path, ".")
	if strings.ToLower(pathsByDot[len(pathsByDot)-1]) == "proto" {
		tempInfos := parseProto(path)
		dirName := ""
		dirs := strings.Split(path, "/")
		if len(dirs) > 1 {
			dirName = dirs[len(dirs)-2]
		}
		if dirName != "" {
			for _, item := range tempInfos {
				item.Dir = dirName
			}
		}
		return tempInfos
	}
	return nil
}

func mergeProtos(infos []*ProtoInfo) []*ProtoInfo {
	protoMap := make(map[string]*ProtoInfo)
	for _, info := range infos {
		if info.Type != MESSAGE_TYPE_PUSH_REQ {
			protoMap[info.Name] = info
		}
	}

	for _, info := range infos {
		if info.Type == MESSAGE_TYPE_PUSH_REQ {
			if val, ok := protoMap[info.Name]; ok {
				val.HasReqIfPush = true
			}
		}
	}

	retProtos := make([]*ProtoInfo, 0)
	for _, value := range protoMap {
		retProtos = append(retProtos, value)
	}
	return retProtos
}

// TODO message 嵌套定义的情况
func parseProto(path string) []*ProtoInfo {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		logs.Error("read file err", err)
		return nil
	}
	protoReader := bufio.NewReader(bytes.NewReader(file))
	protoInfos := make([]*ProtoInfo, 0)

	reqBodyLineIndex := -1 // 如果大于0表示改行是字段行
	for {
		bytes, err := protoReader.ReadBytes('\n')
		if err != nil {
			break
		}
		lineStr := string(bytes)

		if protoName, protoType := parseProtoName(lineStr); protoName != "" {
			protoInfos = append(protoInfos, &ProtoInfo{
				Name:            protoName,
				Type:            protoType,
				ClientReqParams: []ClientReqParam{},
			})
			reqBodyLineIndex = len(protoInfos) - 1
			continue
		}

		if strings.Contains(lineStr, "}") {
			reqBodyLineIndex = -1
			continue
		}

		if reqBodyLineIndex > -1 {
			paramType, paramVar := parseReq(lineStr)
			logs.Debug("<parse proto> type=%s， var=%s", paramType, paramVar)
			if paramType != "" && paramVar != "" {
				protoInfos[reqBodyLineIndex].ClientReqParams = append(protoInfos[reqBodyLineIndex].ClientReqParams, ClientReqParam{
					ParamVar:  paramVar,
					ParamType: paramType,
				})
			}
		}
	}

	logs.Debug("<parse proto> result %v", protoInfos)
	return protoInfos
}

// (name type)
func parseProtoName(lineStr string) (string, string) {
	if strings.Contains(lineStr, "message") {
		lineStr = strings.Replace(lineStr, "message", "", -1)
		lineStr = strings.Replace(lineStr, "{", "", -1)
		lineStr = strings.TrimSpace(lineStr)
		return parseMessage(lineStr)
	}
	return "", ""
}

// (name type)
func parseMessage(message string) (string, string) {
	// 优先判断PushReq, 避免与req重复
	if endsWith(message, MESSAGE_TYPE_PUSH_REQ) {
		return message[:len(message)-7], MESSAGE_TYPE_PUSH_REQ
	}
	if endsWith(message, MESSAGE_TYPE_REQ) {
		return message[:len(message)-3], MESSAGE_TYPE_REQ
	}
	if endsWith(message, MESSAGE_TYPE_PUSH) {
		return message[:len(message)-4], MESSAGE_TYPE_PUSH
	}
	return "", ""
}

// type, var
func parseReq(line string) (string, string) {
	line = strings.TrimSpace(line)
	logs.Debug("<parse req>, %s", line)
	lineArray := strings.Split(line, " ")
	logs.Debug("<parse req> %d %v", len(lineArray), lineArray)
	if len(lineArray) < 3 {
		return "", ""
	}
	paramType := strings.TrimSpace(lineArray[1])
	paramType = getCsType(paramType)
	if strings.TrimSpace(lineArray[0]) == "repeated" {
		paramType += "[]"
	}
	line3 := strings.TrimSpace(lineArray[2])
	paramVar := ""
	if strings.Contains(line3, "=") {
		paramVar = strings.Split(line3, "=")[0]
	} else {
		paramVar = line3
	}
	return paramType, paramVar
}

func getCsType(paramType string) string {
	if retType, ok := Proto2CSTypeMap[paramType]; ok {
		return retType
	} else {
		return "protogen." + paramType
	}
}
