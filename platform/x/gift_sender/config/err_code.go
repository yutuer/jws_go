package config

const (
	RetCode_Success  = 0 //成功
	RetCode_ErrorArg = 1 // 参数不正确
	RetCode_LimitIP  = 2 // IP受限访问

	RetCode_InvalidSign = 3    // 签名不合法
	RetCode_NoRole      = 1001 // 用户不存在
	RetCode_DupPhoneNum = 1002 // 重复绑定手机号
)

var ErrMap map[int]string

func InitErrCode() {
	ErrMap = make(map[int]string, 6)
	ErrMap[RetCode_Success] = "成功"
	ErrMap[RetCode_ErrorArg] = "参数不正确"
	ErrMap[RetCode_LimitIP] = "IP受限访问"
	ErrMap[RetCode_InvalidSign] = "签名不合法"
	ErrMap[RetCode_NoRole] = "用户不存在"
	ErrMap[RetCode_DupPhoneNum] = "重复绑定手机号"
}
