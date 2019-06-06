package sdk_shop

import (
	"encoding/json"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	RetCode_Success  = 0 //成功
	RetCode_ErrorArg = 1 // 参数不正确
	RetCode_LimitIP  = 2 // IP受限访问

	RetCode_InvalidSign = 3    // 签名不合法
	RetCode_NoRole      = 1001 // 用户不存在
	RetCode_DupPhoneNum = 1002 // 重复绑定手机号
)

var errMap map[int]string

func initErrorCode() {
	logs.Debug("ini err code")
	errMap = make(map[int]string, 6)
	errMap[RetCode_Success] = "成功"
	errMap[RetCode_ErrorArg] = "参数不正确"
	errMap[RetCode_LimitIP] = "IP受限访问"
	errMap[RetCode_InvalidSign] = "签名不合法"
	errMap[RetCode_NoRole] = "用户不存在"
	errMap[RetCode_DupPhoneNum] = "重复绑定手机号"
}

func GetErrorCode() map[int]string {
	return errMap
}

type Rsp struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    int    `json:"data"`
	Diamond int    `json:"diamond"`
}

func GetMsg(code int) string {
	logs.Debug("get Err code %s", errMap[code])
	return errMap[code]
}

func genRet(code int) string {
	ret := Rsp{
		Code: code,
		Msg:  GetMsg(code),
	}
	jsonStr, err := json.Marshal(ret)
	if err == nil {
		return string(jsonStr)
	} else {
		return ""
	}
}
