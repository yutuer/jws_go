package hmt_gift

const (
	RetCode_Success  = 0 //成功
	RetCode_ErrorArg = 1 // 参数不正确
	RetCode_LimitIP  = 2 // IP受限访问

	RetCode_InvalidSign = 3    // 签名不合法
	RetCode_NoRole      = 1001 // 用户不存在
	RetCode_DupPhoneNum = 1002 // 重复绑定手机号

	RetCode_NoItem        = 4 // 物品不合法
	RetCode_UserInfoError = 5
	RetCode_GetInfoError  = 6
	RetCode_TimeError     = 7
	RetCode_Default       = 1000
)

var errMap map[int]string

func initErrorCode() {
	errMap = make(map[int]string, 6)
	errMap[RetCode_Success] = "成功"
	errMap[RetCode_ErrorArg] = "参数不正确"
	errMap[RetCode_LimitIP] = "IP受限访问"
	errMap[RetCode_InvalidSign] = "签名不合法"
	errMap[RetCode_NoRole] = "用户不存在"
	errMap[RetCode_DupPhoneNum] = "重复绑定手机号"
	errMap[RetCode_NoItem] = "reward info error"
	errMap[RetCode_UserInfoError] = "userinfo error"
	errMap[RetCode_Default] = "unknow server error"
	errMap[RetCode_GetInfoError] = "infoname error"
	errMap[RetCode_TimeError] = "time Error"
}

func GetErrorCode() map[int]string {
	return errMap
}
